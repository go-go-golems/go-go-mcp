package repository

import "time"

// ScriptExecution represents a stored script execution record
type ScriptExecution struct {
	ID         int       `json:"id" db:"id"`
	SessionID  string    `json:"session_id" db:"session_id"`
	Code       string    `json:"code" db:"code"`
	Result     *string   `json:"result" db:"result"`           // Nullable
	ConsoleLog *string   `json:"console_log" db:"console_log"` // Nullable
	Error      *string   `json:"error" db:"error"`             // Nullable
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	Source     string    `json:"source" db:"source"` // 'api', 'mcp', 'file'
}

// ExecutionFilter provides filtering options for script execution queries
type ExecutionFilter struct {
	Search    string `json:"search,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Source    string `json:"source,omitempty"`
	FromDate  *time.Time `json:"from_date,omitempty"`
	ToDate    *time.Time `json:"to_date,omitempty"`
}

// PaginationOptions provides pagination parameters
type PaginationOptions struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ExecutionQueryResult contains paginated execution results
type ExecutionQueryResult struct {
	Executions []ScriptExecution `json:"executions"`
	Total      int               `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// CreateExecutionRequest contains data for creating a new script execution
type CreateExecutionRequest struct {
	SessionID  string  `json:"session_id"`
	Code       string  `json:"code"`
	Result     *string `json:"result,omitempty"`
	ConsoleLog *string `json:"console_log,omitempty"`
	Error      *string `json:"error,omitempty"`
	Source     string  `json:"source"`
}
