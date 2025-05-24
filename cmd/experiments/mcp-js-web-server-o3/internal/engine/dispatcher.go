package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dop251/goja"
)

// StartDispatcher starts the single-threaded JavaScript dispatcher goroutine
func (e *Engine) StartDispatcher() {
	log.Println("Starting JavaScript dispatcher...")
	
	for job := range e.jobs {
		e.processJob(job)
	}
}

// processJob handles a single evaluation job
func (e *Engine) processJob(job EvalJob) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("JavaScript panic: %v", r)
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
	
	if job.Fn != nil {
		// Execute pre-registered handler function
		err = e.executeHandler(job)
	} else if job.Code != "" {
		// Execute raw JavaScript code
		err = e.executeCode(job.Code)
		if job.W != nil {
			job.W.WriteHeader(http.StatusAccepted)
			job.W.Write([]byte("JavaScript executed"))
		}
	}

	if job.Done != nil {
		job.Done <- err
	}
}

// executeHandler executes a registered JavaScript handler
func (e *Engine) executeHandler(job EvalJob) error {
	// Create request object for JavaScript
	reqObj := e.createRequestObject(job.R)
	
	// Call the JavaScript function
	result, err := job.Fn(goja.Undefined(), e.rt.ToValue(reqObj))
	if err != nil {
		log.Printf("Handler execution error: %v", err)
		if job.W != nil {
			job.W.WriteHeader(http.StatusInternalServerError)
			job.W.Write([]byte(err.Error()))
		}
		return err
	}

	// Process the result
	return e.writeResponse(job.W, result)
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
func (e *Engine) writeResponse(w http.ResponseWriter, result goja.Value) error {
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
		// String response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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