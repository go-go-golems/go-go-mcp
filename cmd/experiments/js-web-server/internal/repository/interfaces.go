package repository

import "context"

// ExecutionRepository defines the interface for script execution storage
type ExecutionRepository interface {
	// CreateExecution stores a new script execution
	CreateExecution(ctx context.Context, req CreateExecutionRequest) (*ScriptExecution, error)

	// GetExecution retrieves a script execution by ID
	GetExecution(ctx context.Context, id int) (*ScriptExecution, error)

	// GetExecutionBySessionID retrieves a script execution by session ID
	GetExecutionBySessionID(ctx context.Context, sessionID string) (*ScriptExecution, error)

	// ListExecutions retrieves script executions with filtering and pagination
	ListExecutions(ctx context.Context, filter ExecutionFilter, pagination PaginationOptions) (*ExecutionQueryResult, error)

	// DeleteExecution removes a script execution by ID
	DeleteExecution(ctx context.Context, id int) error

	// DeleteExecutionsBySessionID removes all executions for a session
	DeleteExecutionsBySessionID(ctx context.Context, sessionID string) error

	// GetExecutionStats returns statistics about script executions
	GetExecutionStats(ctx context.Context) (*ExecutionStats, error)
}

// ExecutionStats contains statistics about script executions
type ExecutionStats struct {
	TotalExecutions      int            `json:"total_executions"`
	SuccessfulExecutions int            `json:"successful_executions"`
	FailedExecutions     int            `json:"failed_executions"`
	ExecutionsBySource   map[string]int `json:"executions_by_source"`
	AverageExecutionTime *float64       `json:"average_execution_time,omitempty"`
}

// RepositoryManager manages all repositories
type RepositoryManager interface {
	Executions() ExecutionRepository
	Close() error
}
