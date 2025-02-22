package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/clay/pkg/watcher"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Server struct {
	dir     string
	pages   map[string]UIDefinition
	watcher *watcher.Watcher
}

type UIDefinition struct {
	Components []map[string]interface{} `yaml:"components"`
}

func NewServer(dir string) *Server {
	s := &Server{
		dir:   dir,
		pages: make(map[string]UIDefinition),
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
		Handler: s.Handler(),
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
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Index page
	mux.HandleFunc("/", s.handleIndex())

	// Individual pages
	for name := range s.pages {
		pagePath := "/" + strings.TrimSuffix(name, ".yaml")
		mux.HandleFunc(pagePath, s.handlePage(name))
	}

	return mux
}

func (s *Server) loadPages() error {
	log.Debug().Str("directory", s.dir).Msg("Loading pages from directory")
	return filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error walking directory")
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".yaml") {
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
	urlPath := "/" + strings.TrimSuffix(relPath, ".yaml")
	urlPath = strings.ReplaceAll(urlPath, string(os.PathSeparator), "/")

	s.pages[relPath] = def
	log.Info().Str("path", relPath).Msg("Loaded page")

	// Register the page handler for the URL path
	http.HandleFunc(urlPath, s.handlePage(relPath))

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

	delete(s.pages, relPath)
	log.Info().Str("path", relPath).Msg("Removed page")
	return nil
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		component := indexTemplate(s.pages)
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) handlePage(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		def, ok := s.pages[name]
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
