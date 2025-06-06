package web

import (
	"net/http"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/gorilla/mux"
)

// SetupJSRoutes sets up routes for the JavaScript web server (user-facing)
func SetupJSRoutes(jsEngine *engine.Engine) *mux.Router {
	r := mux.NewRouter()

	// Dynamic routes (registered by JavaScript) - catch all for JS server
	r.PathPrefix("/").HandlerFunc(DynamicRouteHandler(jsEngine))

	return r
}

// SetupAdminServerRoutes sets up routes for the admin/system interface
func SetupAdminServerRoutes(jsEngine *engine.Engine) *mux.Router {
	r := mux.NewRouter()

	// Static files - highest priority
	r.PathPrefix("/static/").Handler(StaticHandler())

	// API endpoints - these need to be registered early
	r.HandleFunc("/api/repl/execute", ExecuteREPLHandler(jsEngine)).Methods("POST")
	r.HandleFunc("/api/reset-vm", ResetVMHandler(jsEngine)).Methods("POST")
	r.HandleFunc("/api/preset", PresetHandler()).Methods("GET")
	r.HandleFunc("/api/docs", DocsAPIHandler()).Methods("GET")

	// Main application pages
	r.HandleFunc("/", PlaygroundHandler()).Methods("GET") // Default to playground
	r.HandleFunc("/playground", PlaygroundHandler()).Methods("GET")
	r.HandleFunc("/repl", REPLHandler()).Methods("GET")
	r.HandleFunc("/history", HistoryHandler(jsEngine)).Methods("GET")
	r.HandleFunc("/docs", DocsHandler()).Methods("GET")

	// Setup admin routes using existing function
	SetupAdminRoutes(r, jsEngine)

	// Legacy scripts interface (keep for now)
	r.HandleFunc("/scripts", ScriptsHandler(jsEngine))

	return r
}

// SetupRoutes sets up the web routes (legacy compatibility)
func SetupRoutes(jsEngine *engine.Engine) *mux.Router {
	return SetupAdminServerRoutes(jsEngine)
}

// SetupRoutesWithAPI sets up admin routes including the execute API handler
func SetupRoutesWithAPI(jsEngine *engine.Engine, executeHandler http.HandlerFunc) *mux.Router {
	r := SetupAdminServerRoutes(jsEngine)

	// Add the execute API handler
	r.HandleFunc("/v1/execute", executeHandler).Methods("POST")

	return r
}

// DynamicRouteHandler wraps the existing HandleDynamicRoute function
func DynamicRouteHandler(jsEngine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleDynamicRoute(jsEngine, w, r)
	}
}
