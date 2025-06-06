package web

import (
	"net/http"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/gorilla/mux"
)

// SetupRoutes sets up the web routes
func SetupRoutes(jsEngine *engine.Engine) *mux.Router {
	r := mux.NewRouter()

	// Static files - highest priority
	r.PathPrefix("/static/").Handler(StaticHandler())

	// API endpoints - these need to be registered early
	r.HandleFunc("/api/repl/execute", ExecuteREPLHandler(jsEngine)).Methods("POST")
	r.HandleFunc("/api/reset-vm", ResetVMHandler(jsEngine)).Methods("POST")
	r.HandleFunc("/api/preset", PresetHandler()).Methods("GET")

	// Main application pages
	r.HandleFunc("/", HomeHandler()).Methods("GET")
	r.HandleFunc("/playground", PlaygroundHandler()).Methods("GET")
	r.HandleFunc("/repl", REPLHandler()).Methods("GET")
	r.HandleFunc("/history", HistoryHandler(jsEngine)).Methods("GET")
	r.HandleFunc("/docs", DocsHandler()).Methods("GET")

	// Admin interface (new templ-based)
	r.HandleFunc("/admin/logs", AdminLogsHandler(jsEngine.GetRequestLogger())).Methods("GET")

	// Legacy scripts interface (keep for now)
	r.HandleFunc("/scripts", ScriptsHandler(jsEngine))

	// Dynamic routes (registered by JavaScript) - must be last
	r.PathPrefix("/").HandlerFunc(DynamicRouteHandler(jsEngine))

	return r
}

// SetupRoutesWithAPI sets up routes including the execute API handler
func SetupRoutesWithAPI(jsEngine *engine.Engine, executeHandler http.HandlerFunc) *mux.Router {
	r := mux.NewRouter()

	// Static files - highest priority
	r.PathPrefix("/static/").Handler(StaticHandler())

	// API endpoints - must be before dynamic routes
	r.HandleFunc("/v1/execute", executeHandler).Methods("POST")
	r.HandleFunc("/api/repl/execute", ExecuteREPLHandler(jsEngine)).Methods("POST")
	r.HandleFunc("/api/reset-vm", ResetVMHandler(jsEngine)).Methods("POST")
	r.HandleFunc("/api/preset", PresetHandler()).Methods("GET")

	// Main application pages
	r.HandleFunc("/", HomeHandler()).Methods("GET")
	r.HandleFunc("/playground", PlaygroundHandler()).Methods("GET")
	r.HandleFunc("/repl", REPLHandler()).Methods("GET")
	r.HandleFunc("/history", HistoryHandler(jsEngine)).Methods("GET")
	r.HandleFunc("/docs", DocsHandler()).Methods("GET")

	// Admin interface (new templ-based)
	r.HandleFunc("/admin/logs", AdminLogsHandler(jsEngine.GetRequestLogger())).Methods("GET")

	// Legacy scripts interface (keep for now)
	r.HandleFunc("/scripts", ScriptsHandler(jsEngine))

	// Dynamic routes (registered by JavaScript) - must be last
	r.PathPrefix("/").HandlerFunc(DynamicRouteHandler(jsEngine))

	return r
}

// DynamicRouteHandler wraps the existing HandleDynamicRoute function
func DynamicRouteHandler(jsEngine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleDynamicRoute(jsEngine, w, r)
	}
}
