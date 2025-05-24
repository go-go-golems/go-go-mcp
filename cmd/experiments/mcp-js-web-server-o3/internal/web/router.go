package web

import (
	"net/http"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/engine"
)

// HandleDynamicRoute processes requests for JavaScript-registered handlers
func HandleDynamicRoute(jsEngine *engine.Engine, w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	// Check for registered HTTP handler
	if handler, exists := jsEngine.GetHandler(method, path); exists {
		done := make(chan error, 1)
		job := engine.EvalJob{
			Fn:   handler,
			W:    w,
			R:    r,
			Done: done,
		}

		jsEngine.SubmitJob(job)
		
		// Wait for completion
		<-done
		return
	}

	// Check for registered file handler
	if fileHandler, exists := jsEngine.GetFileHandler(path); exists {
		done := make(chan error, 1)
		job := engine.EvalJob{
			Fn:   fileHandler,
			W:    w,
			R:    r,
			Done: done,
		}

		jsEngine.SubmitJob(job)
		
		// Wait for completion
		<-done
		return
	}

	// No handler found
	http.NotFound(w, r)
}