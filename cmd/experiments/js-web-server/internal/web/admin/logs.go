package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
	"github.com/rs/zerolog/log"
)

// LogsHandler handles log-related admin endpoints
type LogsHandler struct {
	logger *engine.RequestLogger
	repos  repository.RepositoryManager
}

// NewLogsHandler creates a new logs handler
func NewLogsHandler(logger *engine.RequestLogger, repos repository.RepositoryManager) *LogsHandler {
	return &LogsHandler{
		logger: logger,
		repos:  repos,
	}
}

// HandleLogsAPI handles API endpoints for log data
func (lh *LogsHandler) HandleLogsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.URL.Path == "/admin/logs/api/stats":
		lh.handleStatsAPI(w, r)
	case r.URL.Path == "/admin/logs/api/requests":
		lh.handleRequestsAPI(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/logs/api/requests/"):
		requestID := strings.TrimPrefix(r.URL.Path, "/admin/logs/api/requests/")
		lh.handleRequestDetailsAPI(w, r, requestID)
	case r.URL.Path == "/admin/logs/api/executions":
		lh.handleExecutionsAPI(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/logs/api/executions/"):
		executionID := strings.TrimPrefix(r.URL.Path, "/admin/logs/api/executions/")
		lh.handleExecutionDetailsAPI(w, r, executionID)
	case r.URL.Path == "/admin/logs/api/clear":
		lh.handleClearLogsAPI(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleStatsAPI returns logging statistics
func (lh *LogsHandler) handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	stats := lh.logger.GetStats()
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Error().Err(err).Msg("Failed to encode stats response")
	}
}

// handleRequestsAPI returns request logs
func (lh *LogsHandler) handleRequestsAPI(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	requests := lh.logger.GetRecentRequests(limit)
	if err := json.NewEncoder(w).Encode(requests); err != nil {
		log.Error().Err(err).Msg("Failed to encode requests response")
	}
}

// handleRequestDetailsAPI returns details for a specific request
func (lh *LogsHandler) handleRequestDetailsAPI(w http.ResponseWriter, r *http.Request, requestID string) {
	if request, exists := lh.logger.GetRequestByID(requestID); exists {
		if err := json.NewEncoder(w).Encode(request); err != nil {
			log.Error().Err(err).Msg("Failed to encode request details response")
		}
	} else {
		http.NotFound(w, r)
	}
}

// handleClearLogsAPI clears all logs
func (lh *LogsHandler) handleClearLogsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lh.logger.ClearLogs()
	log.Info().Msg("Request logs cleared via admin interface")

	response := map[string]interface{}{
		"success": true,
		"message": "Logs cleared successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode clear logs response")
	}
}

// handleExecutionsAPI returns script execution history
func (lh *LogsHandler) handleExecutionsAPI(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	filter := repository.ExecutionFilter{
		Search: r.URL.Query().Get("search"),
	}

	pagination := repository.PaginationOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := lh.repos.Executions().ListExecutions(context.Background(), filter, pagination)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch script executions")
		http.Error(w, "Failed to fetch executions", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Error().Err(err).Msg("Failed to encode executions response")
	}
}

// handleExecutionDetailsAPI returns details for a specific script execution
func (lh *LogsHandler) handleExecutionDetailsAPI(w http.ResponseWriter, r *http.Request, executionIDStr string) {
	executionID, err := strconv.Atoi(executionIDStr)
	if err != nil {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}

	execution, err := lh.repos.Executions().GetExecution(context.Background(), executionID)
	if err != nil {
		log.Error().Err(err).Int("executionID", executionID).Msg("Failed to fetch script execution")
		http.NotFound(w, r)
		return
	}

	if err := json.NewEncoder(w).Encode(execution); err != nil {
		log.Error().Err(err).Msg("Failed to encode execution details response")
	}
}
