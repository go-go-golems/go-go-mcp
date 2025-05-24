package web

import (
	"net/http"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/engine"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// SetupAdminRoutes configures all admin routes on the given router
func SetupAdminRoutes(r *mux.Router, jsEngine *engine.Engine) {
	// Scripts management admin route
	r.HandleFunc("/admin/scripts", ScriptsHandler(jsEngine)).Methods("GET", "POST")
	log.Debug().Msg("Registered admin endpoint: GET/POST /admin/scripts")

	// Admin log routes
	adminHandler := NewAdminHandler(jsEngine.GetRequestLogger())
	r.PathPrefix("/admin/logs").HandlerFunc(adminHandler.HandleAdminLogs)
	log.Debug().Msg("Registered admin endpoint: /admin/logs")
}

// SetupDynamicRoutes configures the dynamic JavaScript-handled routes with request logging
func SetupDynamicRoutes(r *mux.Router, jsEngine *engine.Engine) {
	// Dynamic routes (handled by JS) - wrapped with request logging middleware
	dynamicHandler := jsEngine.GetRequestLogger().RequestLoggerMiddleware(func(w http.ResponseWriter, r *http.Request) {
		HandleDynamicRoute(jsEngine, w, r)
	})
	r.PathPrefix("/").HandlerFunc(dynamicHandler)
	log.Debug().Msg("Registered dynamic route handler with request logging")
}

// SetupFullServer configures a complete server with all routes
func SetupFullServer(jsEngine *engine.Engine) *mux.Router {
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/v1/execute", func(w http.ResponseWriter, r *http.Request) {
		// Import here to avoid circular imports if needed
		// For now, we'll assume api.ExecuteHandler is accessible
		// This might need to be passed as a parameter if there are import issues
		log.Debug().Msg("API execute endpoint called via dynamic setup")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error": "Execute handler not configured in dynamic setup"}`))
	}).Methods("POST")
	log.Debug().Msg("Registered API endpoint: POST /v1/execute")

	// Setup admin routes
	SetupAdminRoutes(r, jsEngine)

	// Setup dynamic routes (this should be last as it's a catch-all)
	SetupDynamicRoutes(r, jsEngine)

	return r
}
