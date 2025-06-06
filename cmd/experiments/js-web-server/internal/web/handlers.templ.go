package web

import (
	"context"
	"embed"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/api"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/web/templates"
	"github.com/rs/zerolog/log"
)

//go:embed static
var staticFiles embed.FS

// GetStaticFS returns the embedded static filesystem for debugging
func GetStaticFS() embed.FS {
	return staticFiles
}

// PlaygroundHandler serves the JavaScript playground page
func PlaygroundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		component := templates.PlaygroundPage()
		err := component.Render(context.Background(), w)
		if err != nil {
			log.Error().Err(err).Msg("Failed to render playground page")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// REPLHandler serves the REPL page
func REPLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		component := templates.REPLPage()
		err := component.Render(context.Background(), w)
		if err != nil {
			log.Error().Err(err).Msg("Failed to render REPL page")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// HistoryHandler serves the execution history page
func HistoryHandler(jsEngine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		search := r.URL.Query().Get("search")
		sessionID := r.URL.Query().Get("sessionId")
		source := r.URL.Query().Get("source")

		limit := 20
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}

		offset := 0
		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		// Build filter
		filter := repository.ExecutionFilter{
			Search:    search,
			SessionID: sessionID,
			Source:    source,
		}

		pagination := repository.PaginationOptions{
			Limit:  limit,
			Offset: offset,
		}

		// Query executions
		result, err := jsEngine.GetRepositoryManager().Executions().ListExecutions(r.Context(), filter, pagination)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get execution history")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Render template
		component := templates.HistoryPage(result, filter, pagination)
		err = component.Render(context.Background(), w)
		if err != nil {
			log.Error().Err(err).Msg("Failed to render history page")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// AdminLogsHandler serves the admin logs page using templ
func AdminLogsHandler(logger *engine.RequestLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination parameters
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
				limit = parsed
			}
		}

		offset := 0
		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		// Get logs from the logger
		allLogs := logger.GetRecentRequests(limit + offset) // Get extra to account for offset
		total := len(allLogs)

		// Convert to the expected type
		logs := make([]engine.RequestLog, len(allLogs))
		for i, logPtr := range allLogs {
			logs[i] = *logPtr
		}

		// Apply offset and limit
		start := offset
		if start > len(logs) {
			start = len(logs)
		}

		end := start + limit
		if end > len(logs) {
			end = len(logs)
		}

		paginatedLogs := logs[start:end]

		// Render template
		component := templates.AdminPage(paginatedLogs, total, limit, offset)
		err := component.Render(context.Background(), w)
		if err != nil {
			log.Error().Err(err).Msg("Failed to render admin page")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// StaticHandler serves static files with correct MIME types
func StaticHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Remove /static/ prefix manually
		path := strings.TrimPrefix(r.URL.Path, "/static/")
		if path == r.URL.Path {
			// Prefix wasn't there
			http.NotFound(w, r)
			return
		}

		// Prevent directory traversal
		if strings.Contains(path, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Full path in embedded FS
		fullPath := "static/" + path
		log.Debug().Str("requestPath", r.URL.Path).Str("strippedPath", path).Str("fullPath", fullPath).Msg("Static file request")

		// Check if file exists in embedded FS
		file, err := staticFiles.Open(fullPath)
		if err != nil {
			log.Debug().Str("path", fullPath).Err(err).Msg("Static file not found")
			http.NotFound(w, r)
			return
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Error().Err(err).Str("path", fullPath).Msg("Failed to close static file")
			}
		}()

		// Set correct MIME type based on file extension
		ext := filepath.Ext(path)
		var contentType string
		switch ext {
		case ".js":
			contentType = "application/javascript"
		case ".css":
			contentType = "text/css"
		case ".html":
			contentType = "text/html"
		case ".json":
			contentType = "application/json"
		case ".svg":
			contentType = "image/svg+xml"
		case ".png":
			contentType = "image/png"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".gif":
			contentType = "image/gif"
		case ".ico":
			contentType = "image/x-icon"
		default:
			contentType = mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "application/octet-stream"
			}
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=3600")

		// Copy file content to response
		http.ServeContent(w, r, filepath.Base(path), time.Time{}, file.(io.ReadSeeker))
	})
}

// HomeHandler redirects to playground
func HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/playground", http.StatusFound)
	}
}

// ExecuteREPLHandler handles REPL execution (non-persistent)
func ExecuteREPLHandler(jsEngine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// This would execute JavaScript without storing to database
		// For now, we'll reuse the existing execute endpoint
		// In the future, we could add a separate REPL execution path

		// For now, redirect to the main execute endpoint
		// but we could implement a separate non-persistent execution here
		api.ExecuteHandler(jsEngine)(w, r)
	}
}

// ResetVMHandler resets the JavaScript VM state
func ResetVMHandler(jsEngine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// This would reset the VM state
		// For now, we'll just return success
		// In the future, we could implement actual VM reset

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"success": true, "message": "VM reset (not implemented)"}`)); err != nil {
			log.Error().Err(err).Msg("Failed to write response")
		}
	}
}
