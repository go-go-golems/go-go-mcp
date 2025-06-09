package web

import (
	"embed"
	"net/http"
	"strings"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/web/admin"
)

//go:embed static/admin
var adminStaticFiles embed.FS

// AdminHandler provides admin endpoints for managing the server
type AdminHandler struct {
	logger           *engine.RequestLogger
	repos            repository.RepositoryManager
	jsEngine         *engine.Engine
	logsHandler      *admin.LogsHandler
	globalHandler    *admin.GlobalStateHandler
	sseHandler       *admin.SSEHandler
	staticFileServer http.Handler
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(logger *engine.RequestLogger, repos repository.RepositoryManager, jsEngine *engine.Engine) *AdminHandler {
	ah := &AdminHandler{
		logger:           logger,
		repos:            repos,
		jsEngine:         jsEngine,
		logsHandler:      admin.NewLogsHandler(logger, repos),
		globalHandler:    admin.NewGlobalStateHandler(jsEngine),
		sseHandler:       admin.NewSSEHandler(logger, repos),
		staticFileServer: http.FileServer(http.FS(adminStaticFiles)),
	}

	return ah
}

// HandleAdminLogs serves the admin logs interface
func (ah *AdminHandler) HandleAdminLogs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/logs" {
		// Serve the static HTML file directly
		content, err := adminStaticFiles.ReadFile("static/admin/logs.html")
		if err != nil {
			http.Error(w, "Failed to read logs.html: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content)
		return
	}

	if r.URL.Path == "/admin/logs/api" || strings.HasPrefix(r.URL.Path, "/admin/logs/api/") {
		ah.logsHandler.HandleLogsAPI(w, r)
		return
	}

	if r.URL.Path == "/admin/logs/events" {
		ah.sseHandler.ServeSSE(w, r)
		return
	}

	http.NotFound(w, r)
}

// HandleGlobalState serves the globalState interface and API
func (ah *AdminHandler) HandleGlobalState(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.Header.Get("Accept") != "application/json" {
		// Serve the static HTML file directly
		content, err := adminStaticFiles.ReadFile("static/admin/globalstate.html")
		if err != nil {
			http.Error(w, "Failed to read globalstate.html: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content)
		return
	}

	// Delegate to the global state handler for API requests
	ah.globalHandler.HandleGlobalState(w, r)
}

// HandleStaticFiles serves admin static files
func (ah *AdminHandler) HandleStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Strip /static prefix to match embedded filesystem structure
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")
	ah.staticFileServer.ServeHTTP(w, r)
}
