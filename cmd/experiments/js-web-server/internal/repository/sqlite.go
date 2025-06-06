package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// sqliteRepositoryManager implements RepositoryManager for SQLite
type sqliteRepositoryManager struct {
	db            *sql.DB
	executionRepo ExecutionRepository
}

// NewSQLiteRepositoryManager creates a new SQLite repository manager
func NewSQLiteRepositoryManager(dbPath string) (RepositoryManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	manager := &sqliteRepositoryManager{
		db: db,
	}

	// Initialize execution repository
	manager.executionRepo = &sqliteExecutionRepository{db: db}

	// Initialize database schema
	if err := manager.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return manager, nil
}

// Executions returns the execution repository
func (m *sqliteRepositoryManager) Executions() ExecutionRepository {
	return m.executionRepo
}

// Close closes the database connection
func (m *sqliteRepositoryManager) Close() error {
	return m.db.Close()
}

// initSchema initializes the database schema
func (m *sqliteRepositoryManager) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS script_executions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		code TEXT NOT NULL,
		result TEXT,
		console_log TEXT,
		error TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		source TEXT DEFAULT 'api'
	);
	
	CREATE INDEX IF NOT EXISTS idx_script_executions_session_id ON script_executions(session_id);
	CREATE INDEX IF NOT EXISTS idx_script_executions_timestamp ON script_executions(timestamp);
	CREATE INDEX IF NOT EXISTS idx_script_executions_source ON script_executions(source);
	`

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Debug().Msg("Database schema initialized")
	return nil
}

// sqliteExecutionRepository implements ExecutionRepository for SQLite
type sqliteExecutionRepository struct {
	db *sql.DB
}

// CreateExecution stores a new script execution
func (r *sqliteExecutionRepository) CreateExecution(ctx context.Context, req CreateExecutionRequest) (*ScriptExecution, error) {
	query := `
	INSERT INTO script_executions (session_id, code, result, console_log, error, source)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING id, session_id, code, result, console_log, error, timestamp, source
	`

	var execution ScriptExecution
	err := r.db.QueryRowContext(ctx, query, req.SessionID, req.Code, req.Result, req.ConsoleLog, req.Error, req.Source).Scan(
		&execution.ID,
		&execution.SessionID,
		&execution.Code,
		&execution.Result,
		&execution.ConsoleLog,
		&execution.Error,
		&execution.Timestamp,
		&execution.Source,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	log.Debug().
		Str("sessionID", execution.SessionID).
		Int("id", execution.ID).
		Str("source", execution.Source).
		Msg("Script execution stored")

	return &execution, nil
}

// GetExecution retrieves a script execution by ID
func (r *sqliteExecutionRepository) GetExecution(ctx context.Context, id int) (*ScriptExecution, error) {
	query := `
	SELECT id, session_id, code, result, console_log, error, timestamp, source
	FROM script_executions 
	WHERE id = ?
	`

	var execution ScriptExecution
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&execution.ID,
		&execution.SessionID,
		&execution.Code,
		&execution.Result,
		&execution.ConsoleLog,
		&execution.Error,
		&execution.Timestamp,
		&execution.Source,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("execution with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return &execution, nil
}

// GetExecutionBySessionID retrieves a script execution by session ID
func (r *sqliteExecutionRepository) GetExecutionBySessionID(ctx context.Context, sessionID string) (*ScriptExecution, error) {
	query := `
	SELECT id, session_id, code, result, console_log, error, timestamp, source
	FROM script_executions 
	WHERE session_id = ?
	ORDER BY timestamp DESC
	LIMIT 1
	`

	var execution ScriptExecution
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&execution.ID,
		&execution.SessionID,
		&execution.Code,
		&execution.Result,
		&execution.ConsoleLog,
		&execution.Error,
		&execution.Timestamp,
		&execution.Source,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("execution with session_id %s not found", sessionID)
		}
		return nil, fmt.Errorf("failed to get execution by session ID: %w", err)
	}

	return &execution, nil
}

// ListExecutions retrieves script executions with filtering and pagination
func (r *sqliteExecutionRepository) ListExecutions(ctx context.Context, filter ExecutionFilter, pagination PaginationOptions) (*ExecutionQueryResult, error) {
	// Build WHERE clause
	var whereClause string
	var args []interface{}
	var conditions []string

	if filter.Search != "" {
		conditions = append(conditions, "(code LIKE ? OR result LIKE ? OR console_log LIKE ?)")
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if filter.SessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, filter.SessionID)
	}

	if filter.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, filter.Source)
	}

	if filter.FromDate != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, filter.FromDate)
	}

	if filter.ToDate != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, filter.ToDate)
	}

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM script_executions " + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
	SELECT id, session_id, code, result, console_log, error, timestamp, source 
	FROM script_executions %s
	ORDER BY timestamp DESC 
	LIMIT ? OFFSET ?
	`, whereClause)

	paginationArgs := append(args, pagination.Limit, pagination.Offset)
	rows, err := r.db.QueryContext(ctx, query, paginationArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query script executions: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database rows")
		}
	}()

	var executions []ScriptExecution
	for rows.Next() {
		var exec ScriptExecution
		err := rows.Scan(
			&exec.ID,
			&exec.SessionID,
			&exec.Code,
			&exec.Result,
			&exec.ConsoleLog,
			&exec.Error,
			&exec.Timestamp,
			&exec.Source,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		executions = append(executions, exec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &ExecutionQueryResult{
		Executions: executions,
		Total:      total,
		Limit:      pagination.Limit,
		Offset:     pagination.Offset,
	}, nil
}

// DeleteExecution removes a script execution by ID
func (r *sqliteExecutionRepository) DeleteExecution(ctx context.Context, id int) error {
	query := "DELETE FROM script_executions WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete execution: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("execution with id %d not found", id)
	}

	return nil
}

// DeleteExecutionsBySessionID removes all executions for a session
func (r *sqliteExecutionRepository) DeleteExecutionsBySessionID(ctx context.Context, sessionID string) error {
	query := "DELETE FROM script_executions WHERE session_id = ?"
	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete executions by session ID: %w", err)
	}

	return nil
}

// GetExecutionStats returns statistics about script executions
func (r *sqliteExecutionRepository) GetExecutionStats(ctx context.Context) (*ExecutionStats, error) {
	stats := &ExecutionStats{
		ExecutionsBySource: make(map[string]int),
	}

	// Get total executions
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM script_executions").Scan(&stats.TotalExecutions)
	if err != nil {
		return nil, fmt.Errorf("failed to get total executions: %w", err)
	}

	// Get successful executions (no error)
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM script_executions WHERE error IS NULL OR error = ''").Scan(&stats.SuccessfulExecutions)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful executions: %w", err)
	}

	// Calculate failed executions
	stats.FailedExecutions = stats.TotalExecutions - stats.SuccessfulExecutions

	// Get executions by source
	rows, err := r.db.QueryContext(ctx, "SELECT source, COUNT(*) FROM script_executions GROUP BY source")
	if err != nil {
		return nil, fmt.Errorf("failed to get executions by source: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database rows")
		}
	}()

	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, fmt.Errorf("failed to scan source stats: %w", err)
		}
		stats.ExecutionsBySource[source] = count
	}

	return stats, nil
}
