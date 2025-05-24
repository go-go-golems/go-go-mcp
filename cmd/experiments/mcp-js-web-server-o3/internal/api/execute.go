package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/engine"
	"github.com/google/uuid"
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

		// Generate session ID for tracking
		sessionID := uuid.New().String()

		// Submit evaluation job with result capture
		done := make(chan error, 1)
		resultChan := make(chan *engine.EvalResult, 1)
		job := engine.EvalJob{
			Handler:   nil, // nil means execute raw code
			Code:      code,
			W:         nil, // Don't let dispatcher write directly
			R:         r,
			Done:      done,
			Result:    resultChan,
			SessionID: sessionID,
			Source:    "api",
		}

		jsEngine.SubmitJob(job)

		// Wait for completion with timeout
		select {
		case result := <-resultChan:
			// Also wait for done signal to ensure completion
			select {
			case err := <-done:
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("JavaScript execution failed: %v", err),
					})
					return
				}
			case <-time.After(5 * time.Second):
				// Continue even if done signal is delayed
			}

			// Create response with result and console output
			responseData := map[string]interface{}{
				"success":    true,
				"result":     result.Value,
				"consoleLog": result.ConsoleLog,
				"savedAs":    filename,
				"sessionID":  sessionID,
				"message":    "JavaScript code executed successfully",
			}

			// Return JSON response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(responseData)

		case <-time.After(30 * time.Second):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusRequestTimeout)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Timeout waiting for JavaScript execution",
			})
		}
	}
}
