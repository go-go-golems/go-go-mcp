package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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
			var executionErr error
			select {
			case err := <-done:
				executionErr = err
			case <-time.After(5 * time.Second):
				// Continue even if done signal is delayed
			}

			// Prepare data for database storage
			var resultStr, consoleLogStr, errorStr string
			
			if result.Value != nil {
				if resultBytes, err := json.Marshal(result.Value); err == nil {
					resultStr = string(resultBytes)
				}
			}
			
			if len(result.ConsoleLog) > 0 {
				consoleLogStr = strings.Join(result.ConsoleLog, "\n")
			}
			
			if executionErr != nil {
				errorStr = executionErr.Error()
			}

			// Store script execution in database
			if err := jsEngine.StoreScriptExecution(sessionID, code, resultStr, consoleLogStr, errorStr, "api"); err != nil {
				log.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to store script execution in database")
			} else {
				log.Debug().Str("sessionID", sessionID).Msg("Script execution stored in database")
			}

			// Handle execution error
			if executionErr != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
					"success":   false,
					"error":     fmt.Sprintf("JavaScript execution failed: %v", executionErr),
					"sessionID": sessionID,
				}); encodeErr != nil {
					log.Error().Err(encodeErr).Msg("Failed to encode error response")
				}
				return
			}

			// Create response with result and console output
			responseData := map[string]interface{}{
				"success":    true,
				"result":     result.Value,
				"consoleLog": result.ConsoleLog,
				"sessionID":  sessionID,
				"message":    "JavaScript code executed and stored in database",
			}

			// Return JSON response
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(responseData); err != nil {
				log.Error().Err(err).Msg("Failed to encode success response")
			}

		case <-time.After(30 * time.Second):
			// Store timeout error in database
			if err := jsEngine.StoreScriptExecution(sessionID, code, "", "", "Timeout waiting for JavaScript execution", "api"); err != nil {
				log.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to store timeout execution in database")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusRequestTimeout)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"success":   false,
				"error":     "Timeout waiting for JavaScript execution",
				"sessionID": sessionID,
			}); err != nil {
				log.Error().Err(err).Msg("Failed to encode timeout response")
			}
		}
	}
}
