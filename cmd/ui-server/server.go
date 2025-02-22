package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/go-go-golems/clay/pkg/watcher"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Server struct {
	dir        string
	pages      map[string]UIDefinition
	routes     map[string]http.HandlerFunc
	watcher    *watcher.Watcher
	mux        *http.ServeMux
	mu         sync.RWMutex
	publisher  message.Publisher
	subscriber message.Subscriber
}

type UIDefinition struct {
	Components []map[string]interface{} `yaml:"components"`
}

func NewServer(dir string) *Server {
	// Initialize watermill publisher/subscriber
	publisher := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)

	s := &Server{
		dir:        dir,
		pages:      make(map[string]UIDefinition),
		routes:     make(map[string]http.HandlerFunc),
		mux:        http.NewServeMux(),
		publisher:  publisher,
		subscriber: publisher,
	}

	// Create a watcher for the pages directory
	log.Debug().Str("directory", dir).Msg("Initializing watcher for directory")
	w := watcher.NewWatcher(
		watcher.WithPaths(dir),
		watcher.WithMask("**/*.yaml"),
		watcher.WithWriteCallback(s.handleFileChange),
		watcher.WithRemoveCallback(s.handleFileRemove),
	)

	s.watcher = w

	// Set up SSE endpoint
	s.mux.HandleFunc("/sse", s.handleSSE())

	// Set up static file handler
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Set up dynamic page handler - must come before index handler
	s.mux.Handle("/pages/", s.handleAllPages())

	// Set up index handler - must come last as it's the most general
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		s.handleIndex()(w, r)
	})

	return s
}

func (s *Server) Start(ctx context.Context, port int) error {
	log.Debug().Int("port", port).Msg("Starting server")
	if err := s.loadPages(); err != nil {
		return fmt.Errorf("failed to load pages: %w", err)
	}

	// Start the file watcher
	go func() {
		log.Debug().Msg("Starting file watcher")
		if err := s.watcher.Run(ctx); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Watcher error")
		}
	}()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.mux,
	}

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("server error: %w", err)
		}
		close(serverErr)
	}()

	// Wait for either context cancellation or server error
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Info().Msg("Server shutdown initiated")
		// Graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		log.Info().Msg("Server shutdown completed")
		return nil
	}
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) loadPages() error {
	log.Debug().Str("directory", s.dir).Msg("Loading pages recursively from directory")

	// First, clear existing pages and routes
	s.mu.Lock()
	s.pages = make(map[string]UIDefinition)
	s.routes = make(map[string]http.HandlerFunc)
	s.mu.Unlock()

	return filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error walking directory")
			return err
		}

		if d.IsDir() {
			log.Debug().Str("directory", path).Msg("Entering directory")
			return nil
		}

		if strings.HasSuffix(d.Name(), ".yaml") {
			log.Debug().Str("path", path).Msg("Found YAML page")
			if err := s.loadPage(path); err != nil {
				log.Error().Err(err).Str("path", path).Msg("Failed to load page")
				return err
			}
		}
		return nil
	})
}

func (s *Server) loadPage(path string) error {
	log.Debug().Str("path", path).Msg("Loading page")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to read file")
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var def UIDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to parse YAML")
		return fmt.Errorf("failed to parse YAML in %s: %w", path, err)
	}

	relPath, err := filepath.Rel(s.dir, path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to get relative path")
		return fmt.Errorf("failed to get relative path for %s: %w", path, err)
	}

	// Normalize the relative path to use as a URL path
	urlPath := "/pages/" + strings.TrimSuffix(relPath, ".yaml")
	urlPath = strings.ReplaceAll(urlPath, string(os.PathSeparator), "/")

	s.mu.Lock()
	s.pages[relPath] = def
	s.routes[urlPath] = s.handlePage(relPath)
	s.mu.Unlock()

	log.Info().Str("path", relPath).Str("url", urlPath).Msg("Loaded page")
	return nil
}

func (s *Server) handleFileChange(path string) error {
	if !strings.HasSuffix(path, ".yaml") {
		return nil
	}

	log.Debug().Str("path", path).Msg("File change detected")
	log.Info().Str("path", path).Msg("Reloading page")
	return s.loadPage(path)
}

func (s *Server) handleFileRemove(path string) error {
	if !strings.HasSuffix(path, ".yaml") {
		return nil
	}

	log.Debug().Str("path", path).Msg("File removal detected")
	relPath, err := filepath.Rel(s.dir, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path for %s: %w", path, err)
	}

	urlPath := "/pages/" + strings.TrimSuffix(relPath, ".yaml")
	urlPath = strings.ReplaceAll(urlPath, string(os.PathSeparator), "/")

	s.mu.Lock()
	delete(s.pages, relPath)
	delete(s.routes, urlPath)
	s.mu.Unlock()

	log.Info().Str("path", relPath).Str("url", urlPath).Msg("Removed page")
	return nil
}

func (s *Server) handleAllPages() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("path", r.URL.Path).Msg("Handling page request")

		s.mu.RLock()
		handler, ok := s.routes[r.URL.Path]
		s.mu.RUnlock()

		if !ok {
			log.Debug().Str("path", r.URL.Path).Msg("Page not found")
			http.NotFound(w, r)
			return
		}

		handler(w, r)
	})
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		s.mu.RLock()
		pages := s.pages
		s.mu.RUnlock()

		component := indexTemplate(pages)
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) handlePage(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		def, ok := s.pages[name]
		s.mu.RUnlock()

		if !ok {
			http.NotFound(w, r)
			return
		}

		component := pageTemplate(name, def)
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleSSE() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Get page ID from query
		pageID := r.URL.Query().Get("page")
		if pageID == "" {
			http.Error(w, "page parameter is required", http.StatusBadRequest)
			return
		}

		// Subscribe to page-specific topic
		messages, err := s.subscriber.Subscribe(r.Context(), "ui-updates."+pageID)
		if err != nil {
			http.Error(w, "Failed to subscribe to events", http.StatusInternalServerError)
			return
		}

		// Stream messages
		for msg := range messages {
			// Format SSE message
			fmt.Fprintf(w, "event: %s\n", msg.Metadata["event-type"])
			fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
			w.(http.Flusher).Flush()
		}
	}
}
