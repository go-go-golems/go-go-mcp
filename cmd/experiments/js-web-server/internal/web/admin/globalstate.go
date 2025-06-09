package admin

import (
	"net/http"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
)

// GlobalStateHandler handles global state management
type GlobalStateHandler struct {
	jsEngine *engine.Engine
}

// NewGlobalStateHandler creates a new global state handler
func NewGlobalStateHandler(jsEngine *engine.Engine) *GlobalStateHandler {
	return &GlobalStateHandler{
		jsEngine: jsEngine,
	}
}

// HandleGlobalState serves the globalState interface and API
func (gsh *GlobalStateHandler) HandleGlobalState(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.Header.Get("Accept") == "application/json" {
			// API request - return JSON
			globalState := gsh.jsEngine.GetGlobalState()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(globalState))
		} else {
			// Regular request - serve HTML interface (this will be handled by static file server)
			http.Error(w, "Use static file server for HTML", http.StatusNotImplemented)
		}
	case "POST":
		// Update globalState
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		jsonData := r.FormValue("globalState")
		if jsonData == "" {
			http.Error(w, "Missing globalState data", http.StatusBadRequest)
			return
		}

		if err := gsh.jsEngine.SetGlobalState(jsonData); err != nil {
			http.Error(w, "Failed to update globalState: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Return success response
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success": true}`))
		} else {
			// Redirect back to the interface
			http.Redirect(w, r, "/admin/globalstate", http.StatusSeeOther)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
