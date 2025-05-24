package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/engine"
)

// ExecuteHandler returns an HTTP handler for the /v1/execute endpoint
func ExecuteHandler(jsEngine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read JavaScript code from request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		if len(body) == 0 {
			http.Error(w, "Empty request body", http.StatusBadRequest)
			return
		}

		code := string(body)

		// Persist source code with timestamp
		timestamp := time.Now().UTC().Format("20060102-150405")
		filename := fmt.Sprintf("scripts/%s.js", timestamp)
		
		if err := os.WriteFile(filename, body, 0644); err != nil {
			http.Error(w, "Failed to persist script", http.StatusInternalServerError)
			return
		}

		// Submit evaluation job (non-blocking)
		done := make(chan error, 1)
		job := engine.EvalJob{
			Handler: nil, // nil means execute raw code
			Code:    code,
			W:       w,
			R:       r,
			Done:    done,
		}

		jsEngine.SubmitJob(job)

		// Wait for completion
		if err := <-done; err != nil {
			http.Error(w, fmt.Sprintf("JavaScript execution failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Response is written by the dispatcher
	}
}