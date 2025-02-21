package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Server struct {
	dir   string
	pages map[string]UIDefinition
}

type UIDefinition struct {
	Components map[string]interface{} `yaml:",inline"`
}

func NewServer(dir string) *Server {
	return &Server{
		dir:   dir,
		pages: make(map[string]UIDefinition),
	}
}

func (s *Server) Start(ctx context.Context, port int) error {
	if err := s.loadPages(); err != nil {
		return fmt.Errorf("failed to load pages: %w", err)
	}

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

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) loadPages() error {
	return filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".yaml") {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			var def UIDefinition
			if err := yaml.Unmarshal(data, &def); err != nil {
				return fmt.Errorf("failed to parse YAML in %s: %w", path, err)
			}

			relPath, err := filepath.Rel(s.dir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path for %s: %w", path, err)
			}

			s.pages[relPath] = def
		}
		return nil
	})
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
