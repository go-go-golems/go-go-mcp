package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// StartDispatcher starts the single-threaded JavaScript dispatcher goroutine
func (e *Engine) StartDispatcher() {
	log.Info().Msg("JavaScript dispatcher started")
	
	for job := range e.jobs {
		e.processJob(job)
	}
	
	log.Info().Msg("JavaScript dispatcher stopped")
}

// processJob handles a single evaluation job
func (e *Engine) processJob(job EvalJob) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("panic", r).Msg("JavaScript panic during job execution")
			if job.W != nil {
				job.W.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(job.W, "JavaScript error: %v", r)
			}
			if job.Done != nil {
				job.Done <- fmt.Errorf("panic: %v", r)
			}
		}
	}()

	var err error
	
	if job.Handler != nil {
		log.Debug().Msg("Executing registered handler function")
		err = e.executeHandler(job)
	} else if job.Code != "" {
		log.Debug().Str("code", job.Code).Msg("Executing raw JavaScript code")
		err = e.executeCode(job.Code)
		if job.W != nil {
			job.W.WriteHeader(http.StatusAccepted)
			job.W.Write([]byte("JavaScript executed"))
		}
	}

	if err != nil {
		log.Error().Err(err).Msg("Job execution failed")
	} else {
		log.Debug().Msg("Job execution completed successfully")
	}

	if job.Done != nil {
		job.Done <- err
	}
}

// executeHandler executes a registered JavaScript handler
func (e *Engine) executeHandler(job EvalJob) error {
	log.Debug().Str("method", job.R.Method).Str("path", job.R.URL.Path).Msg("Executing JavaScript handler")
	
	// Create request object for JavaScript
	reqObj := e.createRequestObject(job.R)
	
	// Call the JavaScript function
	result, err := job.Handler.Fn(goja.Undefined(), e.rt.ToValue(reqObj))
	if err != nil {
		log.Error().Err(err).Str("method", job.R.Method).Str("path", job.R.URL.Path).Msg("Handler execution error")
		if job.W != nil {
			job.W.WriteHeader(http.StatusInternalServerError)
			job.W.Write([]byte(err.Error()))
		}
		return err
	}

	log.Debug().Str("method", job.R.Method).Str("path", job.R.URL.Path).Msg("Handler executed successfully")
	
	// Process the result
	return e.writeResponse(job.W, result, job.Handler.ContentType)
}

// createRequestObject creates a JavaScript-compatible request object
func (e *Engine) createRequestObject(r *http.Request) map[string]interface{} {
	// Parse query parameters
	query := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			query[k] = v[0]
		} else {
			query[k] = v
		}
	}

	// Parse headers
	headers := make(map[string]interface{})
	for k, v := range r.Header {
		if len(v) == 1 {
			headers[k] = v[0]
		} else {
			headers[k] = v
		}
	}

	return map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL.String(),
		"path":    r.URL.Path,
		"query":   query,
		"headers": headers,
	}
}

// writeResponse writes the JavaScript result to the HTTP response
func (e *Engine) writeResponse(w http.ResponseWriter, result goja.Value, contentTypeOverride ...string) error {
	if result == nil || goja.IsUndefined(result) {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	exported := result.Export()

	switch v := exported.(type) {
	case []byte:
		// Raw bytes - write directly
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err := w.Write(v)
		return err
	case string:
		// String response - use override or detect content type
		var contentType string
		if len(contentTypeOverride) > 0 && contentTypeOverride[0] != "" {
			contentType = contentTypeOverride[0]
		} else {
			contentType = "text/plain; charset=utf-8"
			if isHTML(v) {
				contentType = "text/html; charset=utf-8"
			} else if isJSON(v) {
				contentType = "application/json"
			}
		}
		w.Header().Set("Content-Type", contentType)
		_, err := w.Write([]byte(v))
		return err
	default:
		// JSON response
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(v)
	}
}

// isHTML checks if a string appears to be HTML content
func isHTML(s string) bool {
	trimmed := strings.TrimSpace(s)
	return strings.HasPrefix(strings.ToLower(trimmed), "<!doctype html") ||
		strings.HasPrefix(strings.ToLower(trimmed), "<html") ||
		strings.HasPrefix(trimmed, "<!")
}

// isJSON checks if a string appears to be JSON content
func isJSON(s string) bool {
	trimmed := strings.TrimSpace(s)
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}