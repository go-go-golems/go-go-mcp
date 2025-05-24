package jsserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/gorilla/mux"
)

func (s *JSWebServer) executeJSHandler(handler goja.Value, w http.ResponseWriter, r *http.Request) {
	// Create request and response objects for JavaScript
	req := s.vm.ToValue(s.createJSRequest(r))
	res := s.vm.ToValue(s.createJSResponse(w))

	// Execute the JavaScript handler
	var err error
	if callable, ok := goja.AssertFunction(handler); ok {
		_, err = callable(goja.Undefined(), req, res)
	} else {
		err = fmt.Errorf("handler is not a function")
	}
	if err != nil {
		// Log the error
		fmt.Printf("JavaScript handler error: %v\n", err)
		
		// Return error response if headers haven't been written
		if !isResponseWritten(w) {
			http.Error(w, fmt.Sprintf("JavaScript handler error: %v", err), http.StatusInternalServerError)
		}
	}
}

func (s *JSWebServer) executeJSFileGenerator(file JSFile, w http.ResponseWriter, r *http.Request) {
	// Set content type
	if file.MimeType != "" {
		w.Header().Set("Content-Type", file.MimeType)
	}

	// Create request object for JavaScript
	req := s.vm.ToValue(s.createJSRequest(r))

	// Execute the JavaScript file generator
	var result goja.Value
	var err error
	if callable, ok := goja.AssertFunction(file.Generator); ok {
		result, err = callable(goja.Undefined(), req)
	} else {
		err = fmt.Errorf("file generator is not a function")
	}
	if err != nil {
		// Log the error
		fmt.Printf("JavaScript file generator error: %v\n", err)
		
		// Return error response
		http.Error(w, fmt.Sprintf("File generator error: %v", err), http.StatusInternalServerError)
		return
	}

	// Write the result to response
	content := result.String()
	w.Write([]byte(content))
}

// isResponseWritten checks if the response has already been written to
// This is a simple heuristic based on checking if any headers have been set
func isResponseWritten(w http.ResponseWriter) bool {
	// Check if Content-Length header exists, which is typically set after writing
	return w.Header().Get("Content-Length") != ""
}

// Additional handler methods for extended API

func (s *JSWebServer) handleGetArchivedFiles(w http.ResponseWriter, r *http.Request) {
	files, err := s.listArchivedFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (s *JSWebServer) handleGetArchivedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	content, err := s.getArchivedFile(filename)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(content))
}

func (s *JSWebServer) handleDeleteArchivedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	if err := s.deleteArchivedFile(filename); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *JSWebServer) handleGetExecutionHistory(w http.ResponseWriter, r *http.Request) {
	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	executions, err := s.getExecutionHistory(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executions)
}

func (s *JSWebServer) handleCleanup(w http.ResponseWriter, r *http.Request) {
	keepDays := 30 // Default: keep 30 days
	if daysStr := r.URL.Query().Get("keep_days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			keepDays = parsedDays
		}
	}

	// Cleanup old executions
	if err := s.cleanupOldExecutions(keepDays); err != nil {
		http.Error(w, fmt.Sprintf("Failed to cleanup executions: %v", err), http.StatusInternalServerError)
		return
	}

	// Cleanup old archived files
	if err := s.cleanupArchivedFiles(keepDays); err != nil {
		http.Error(w, fmt.Sprintf("Failed to cleanup archived files: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Cleaned up records older than %d days", keepDays),
	})
}

// Add these new endpoints to the server setup
func (s *JSWebServer) setupExtendedAPIRoutes() {
	api := s.router.PathPrefix("/api").Subrouter()
	
	// Archive management
	api.HandleFunc("/archive", s.handleGetArchivedFiles).Methods("GET")
	api.HandleFunc("/archive/{filename}", s.handleGetArchivedFile).Methods("GET")
	api.HandleFunc("/archive/{filename}", s.handleDeleteArchivedFile).Methods("DELETE")
	
	// Execution history
	api.HandleFunc("/executions", s.handleGetExecutionHistory).Methods("GET")
	
	// Cleanup
	api.HandleFunc("/cleanup", s.handleCleanup).Methods("POST")
}