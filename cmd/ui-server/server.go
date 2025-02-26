package main

import (
	"bytes"
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
	"github.com/go-go-golems/go-go-mcp/pkg/events"
	"github.com/go-go-golems/go-go-mcp/pkg/server/sse"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
	events     events.EventManager
	sseHandler *sse.SSEHandler
}

type UIDefinition struct {
	Components []map[string]interface{} `yaml:"components"`
}

func NewServer(dir string) (*Server, error) {
	// Initialize watermill publisher/subscriber
	publisher := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)

	// Initialize event manager with logger
	eventManager, err := events.NewWatermillEventManager(&log.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create event manager: %w", err)
	}

	// Create SSE handler
	sseHandler := sse.NewSSEHandler(eventManager, &log.Logger)

	s := &Server{
		dir:        dir,
		pages:      make(map[string]UIDefinition),
		routes:     make(map[string]http.HandlerFunc),
		mux:        http.NewServeMux(),
		publisher:  publisher,
		subscriber: publisher,
		events:     eventManager,
		sseHandler: sseHandler,
	}

	// Register component renderers
	s.registerComponentRenderers()

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
	s.mux.Handle("/sse", sseHandler)

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

	return s, nil
}

func (s *Server) Start(ctx context.Context, port int) error {
	log.Debug().Int("port", port).Msg("Starting server")
	if err := s.loadPages(); err != nil {
		return fmt.Errorf("failed to load pages: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	// Start the file watcher
	g.Go(func() error {
		log.Debug().Msg("Starting file watcher")
		if err := s.watcher.Run(ctx); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Watcher error")
			return err
		}
		return nil
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.mux,
	}

	// Start server
	g.Go(func() error {
		log.Info().Str("addr", srv.Addr).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	// Wait for shutdown
	g.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("Server shutdown initiated")

		// Graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// Close event manager
		if err := s.events.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close event manager")
		}

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}

		log.Info().Msg("Server shutdown completed")
		return nil
	})

	return g.Wait()
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

	// Publish page reload event
	event := events.NewPageReloadEvent(relPath, def)
	if err := s.events.Publish(relPath, event); err != nil {
		log.Error().Err(err).Str("path", relPath).Msg("Failed to publish page reload event")
	}

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

// registerComponentRenderers registers component renderers with the SSE handler
func (s *Server) registerComponentRenderers() {
	// Register component-update renderer
	s.sseHandler.RegisterRenderer("component-update", func(componentID string, data interface{}) (string, error) {
		log.Debug().Str("componentID", componentID).Interface("data", data).Msg("Rendering component")

		// Get the page definition for this component
		s.mu.RLock()
		defer s.mu.RUnlock()

		// Find the component in the pages
		for _, def := range s.pages {
			for _, component := range def.Components {
				for typ, props := range component {
					propsMap, ok := props.(map[string]interface{})
					if !ok {
						continue
					}

					id, ok := propsMap["id"].(string)
					if !ok {
						continue
					}

					if id == componentID {
						// Found the component, now update it with the new data
						dataMap, ok := data.(map[string]interface{})
						if ok {
							// Merge the new data with the existing props
							for k, v := range dataMap {
								propsMap[k] = v
							}
						}

						// Render the component
						var buf bytes.Buffer
						err := renderComponent(typ, propsMap).Render(context.Background(), &buf)
						if err != nil {
							return "", fmt.Errorf("failed to render component: %w", err)
						}

						return buf.String(), nil
					}
				}
			}
		}

		return "", fmt.Errorf("component not found: %s", componentID)
	})

	// Register page-template renderer
	s.sseHandler.RegisterRenderer("page-content-template", func(pageID string, data interface{}) (string, error) {
		log.Debug().Str("pageID", pageID).Msg("Rendering page content template")

		// Get the page definition
		var def UIDefinition

		// First check if we have data from the event
		if data != nil {
			// Try to use the data from the event
			if eventData, ok := data.(map[string]interface{}); ok {
				if pageData, ok := eventData["data"]; ok {
					// Try to convert the data to UIDefinition
					if pageDefMap, ok := pageData.(map[string]interface{}); ok {
						if components, ok := pageDefMap["components"].([]interface{}); ok {
							// Convert components to the expected format
							compList := make([]map[string]interface{}, 0, len(components))
							for _, comp := range components {
								if compMap, ok := comp.(map[string]interface{}); ok {
									compList = append(compList, compMap)
								}
							}
							def = UIDefinition{Components: compList}
						}
					}
				}
			}
		}

		// If we couldn't get the definition from the event data, use the stored one
		if len(def.Components) == 0 {
			s.mu.RLock()
			storedDef, ok := s.pages[pageID]
			s.mu.RUnlock()

			if !ok {
				return "", fmt.Errorf("page not found: %s", pageID)
			}

			def = storedDef
		}

		// Render just the page content template, not the full page with base
		var buf bytes.Buffer
		err := pageContentTemplate(pageID, def).Render(context.Background(), &buf)
		if err != nil {
			return "", fmt.Errorf("failed to render page content template: %w", err)
		}

		return buf.String(), nil
	})

	// Register yaml-update renderer
	s.sseHandler.RegisterRenderer("yaml-update", func(componentID string, data interface{}) (string, error) {
		log.Debug().Str("componentID", componentID).Interface("data", data).Msg("Rendering YAML")

		// Convert the data to YAML
		yamlBytes, err := yaml.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal YAML: %w", err)
		}

		// Create HTML with syntax highlighting
		html := fmt.Sprintf("<pre><code class=\"language-yaml\">%s</code></pre>", string(yamlBytes))
		return html, nil
	})
}
