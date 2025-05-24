package jsserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Config struct {
	DatabasePath string
	ArchiveDir   string
	StaticDir    string
}

type JSWebServer struct {
	vm          *goja.Runtime
	db          *sql.DB
	config      *Config
	router      *mux.Router
	routes      map[string]JSRoute
	files       map[string]JSFile
	globalState map[string]interface{}
	mu          sync.RWMutex
}

type JSRoute struct {
	Method   string    `json:"method"`
	Path     string    `json:"path"`
	Handler  goja.Value `json:"-"`
	Created  time.Time `json:"created"`
}

type JSFile struct {
	Path      string    `json:"path"`
	Generator goja.Value `json:"-"`
	MimeType  string    `json:"mime_type"`
	Created   time.Time `json:"created"`
}

type ExecuteRequest struct {
	Code    string `json:"code"`
	Persist bool   `json:"persist"`
	Name    string `json:"name,omitempty"`
}

type ExecuteResponse struct {
	Success      bool   `json:"success"`
	Result       string `json:"result,omitempty"`
	Error        string `json:"error,omitempty"`
	ExecutionID  string `json:"execution_id,omitempty"`
	ArchivedFile string `json:"archived_file,omitempty"`
}

func New(config *Config) (*JSWebServer, error) {
	server := &JSWebServer{
		config:      config,
		routes:      make(map[string]JSRoute),
		files:       make(map[string]JSFile),
		globalState: make(map[string]interface{}),
		router:      mux.NewRouter(),
	}

	// Initialize database
	if err := server.initDatabase(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize database")
	}

	// Initialize JavaScript runtime
	if err := server.initJavaScript(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize JavaScript runtime")
	}

	// Setup API routes
	server.setupAPIRoutes()
	server.setupExtendedAPIRoutes()

	return server, nil
}

func (s *JSWebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *JSWebServer) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *JSWebServer) setupAPIRoutes() {
	api := s.router.PathPrefix("/api").Subrouter()
	
	// Code execution
	api.HandleFunc("/execute", s.handleExecute).Methods("POST")
	
	// Route management
	api.HandleFunc("/routes", s.handleListRoutes).Methods("GET")
	api.HandleFunc("/routes/{path:.*}", s.handleDeleteRoute).Methods("DELETE")
	
	// File management
	api.HandleFunc("/files", s.handleListFiles).Methods("GET")
	api.HandleFunc("/files/{path:.*}", s.handleDeleteFile).Methods("DELETE")
	
	// State management
	api.HandleFunc("/state", s.handleGetAllState).Methods("GET")
	api.HandleFunc("/state/{key}", s.handleGetState).Methods("GET")
	api.HandleFunc("/state/{key}", s.handleSetState).Methods("PUT")
	api.HandleFunc("/state/{key}", s.handleDeleteState).Methods("DELETE")

	// Catch-all for dynamic routes and files
	s.router.PathPrefix("/").HandlerFunc(s.handleDynamic)
}

func (s *JSWebServer) handleExecute(w http.ResponseWriter, r *http.Request) {
	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := s.executeJavaScript(req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *JSWebServer) handleListRoutes(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	routes := make([]JSRoute, 0, len(s.routes))
	for _, route := range s.routes {
		routes = append(routes, JSRoute{
			Method:  route.Method,
			Path:    route.Path,
			Created: route.Created,
		})
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

func (s *JSWebServer) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	routePath := "/" + vars["path"]

	s.mu.Lock()
	deleted := false
	for key, route := range s.routes {
		if route.Path == routePath {
			delete(s.routes, key)
			deleted = true
		}
	}
	s.mu.Unlock()

	if !deleted {
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *JSWebServer) handleListFiles(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	files := make([]JSFile, 0, len(s.files))
	for _, file := range s.files {
		files = append(files, JSFile{
			Path:     file.Path,
			MimeType: file.MimeType,
			Created:  file.Created,
		})
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (s *JSWebServer) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := "/" + vars["path"]

	s.mu.Lock()
	delete(s.files, filePath)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *JSWebServer) handleGetAllState(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	state := make(map[string]interface{})
	for k, v := range s.globalState {
		state[k] = v
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (s *JSWebServer) handleGetState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.mu.RLock()
	value, exists := s.globalState[key]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"value": value})
}

func (s *JSWebServer) handleSetState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	value, exists := body["value"]
	if !exists {
		http.Error(w, "Missing 'value' field", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.globalState[key] = value
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *JSWebServer) handleDeleteState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.mu.Lock()
	delete(s.globalState, key)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *JSWebServer) handleDynamic(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	method := r.Method

	// Check for registered routes first
	s.mu.RLock()
	routeKey := fmt.Sprintf("%s %s", method, requestPath)
	if route, exists := s.routes[routeKey]; exists {
		s.mu.RUnlock()
		s.executeJSHandler(route.Handler, w, r)
		return
	}

	// Check for registered files
	if file, exists := s.files[requestPath]; exists {
		s.mu.RUnlock()
		s.executeJSFileGenerator(file, w, r)
		return
	}
	s.mu.RUnlock()

	// Check for static files
	if _, err := http.Dir(s.config.StaticDir).Open(strings.TrimPrefix(requestPath, "/")); err == nil {
		http.FileServer(http.Dir(s.config.StaticDir)).ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}