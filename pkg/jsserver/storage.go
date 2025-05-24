package jsserver

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

func (s *JSWebServer) initDatabase() error {
	var err error
	s.db, err = sql.Open("sqlite3", s.config.DatabasePath)
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}

	if err := s.db.Ping(); err != nil {
		return errors.Wrap(err, "failed to ping database")
	}

	return s.createTables()
}

func (s *JSWebServer) createTables() error {
	schema := `
	-- Global state storage
	CREATE TABLE IF NOT EXISTS global_state (
		key TEXT PRIMARY KEY,
		value TEXT,
		type TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Execution history
	CREATE TABLE IF NOT EXISTS executions (
		id TEXT PRIMARY KEY,
		code TEXT,
		result TEXT,
		success BOOLEAN,
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		archived_file TEXT
	);

	-- Registered routes
	CREATE TABLE IF NOT EXISTS routes (
		path TEXT,
		method TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (path, method)
	);

	-- Registered files
	CREATE TABLE IF NOT EXISTS files (
		path TEXT PRIMARY KEY,
		mime_type TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes for performance
	CREATE INDEX IF NOT EXISTS idx_executions_executed_at ON executions(executed_at);
	CREATE INDEX IF NOT EXISTS idx_routes_method ON routes(method);
	CREATE INDEX IF NOT EXISTS idx_global_state_updated_at ON global_state(updated_at);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return errors.Wrap(err, "failed to create database schema")
	}

	return nil
}

func (s *JSWebServer) storeExecution(id, code, result string, success bool, archivedFile string) error {
	query := `
		INSERT INTO executions (id, code, result, success, archived_file)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, id, code, result, success, archivedFile)
	if err != nil {
		return errors.Wrap(err, "failed to store execution")
	}

	return nil
}

func (s *JSWebServer) saveState(key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "failed to marshal state value")
	}

	valueType := "unknown"
	switch value.(type) {
	case string:
		valueType = "string"
	case int, int8, int16, int32, int64:
		valueType = "integer"
	case float32, float64:
		valueType = "float"
	case bool:
		valueType = "boolean"
	case map[string]interface{}:
		valueType = "object"
	case []interface{}:
		valueType = "array"
	}

	query := `
		INSERT OR REPLACE INTO global_state (key, value, type, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err = s.db.Exec(query, key, string(valueBytes), valueType)
	if err != nil {
		return errors.Wrap(err, "failed to save state")
	}

	return nil
}

func (s *JSWebServer) loadState() error {
	query := `SELECT key, value, type FROM global_state`

	rows, err := s.db.Query(query)
	if err != nil {
		return errors.Wrap(err, "failed to query state")
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		var key, valueStr, valueType string
		if err := rows.Scan(&key, &valueStr, &valueType); err != nil {
			return errors.Wrap(err, "failed to scan state row")
		}

		var value interface{}
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			// If JSON unmarshal fails, store as string
			value = valueStr
		}

		s.globalState[key] = value
	}

	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "error iterating state rows")
	}

	return nil
}

func (s *JSWebServer) saveRoute(method, path string) error {
	query := `
		INSERT OR REPLACE INTO routes (method, path)
		VALUES (?, ?)
	`

	_, err := s.db.Exec(query, method, path)
	if err != nil {
		return errors.Wrap(err, "failed to save route")
	}

	return nil
}

func (s *JSWebServer) saveFile(path, mimeType string) error {
	query := `
		INSERT OR REPLACE INTO files (path, mime_type)
		VALUES (?, ?)
	`

	_, err := s.db.Exec(query, path, mimeType)
	if err != nil {
		return errors.Wrap(err, "failed to save file")
	}

	return nil
}

func (s *JSWebServer) getExecutionHistory(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT id, code, result, success, executed_at, archived_file
		FROM executions
		ORDER BY executed_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query execution history")
	}
	defer rows.Close()

	var executions []map[string]interface{}
	for rows.Next() {
		var id, code, result, archivedFile string
		var success bool
		var executedAt time.Time

		if err := rows.Scan(&id, &code, &result, &success, &executedAt, &archivedFile); err != nil {
			return nil, errors.Wrap(err, "failed to scan execution row")
		}

		executions = append(executions, map[string]interface{}{
			"id":            id,
			"code":          code,
			"result":        result,
			"success":       success,
			"executed_at":   executedAt,
			"archived_file": archivedFile,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating execution rows")
	}

	return executions, nil
}

func (s *JSWebServer) cleanupOldExecutions(keepDays int) error {
	query := `
		DELETE FROM executions
		WHERE executed_at < datetime('now', '-' || ? || ' days')
	`

	result, err := s.db.Exec(query, keepDays)
	if err != nil {
		return errors.Wrap(err, "failed to cleanup old executions")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		// Log cleanup activity
		s.storeExecution(
			"cleanup-"+time.Now().Format("2006-01-02T15:04:05Z"),
			"/* automatic cleanup */",
			"Cleaned up "+string(rune(rowsAffected))+" old executions",
			true,
			"",
		)
	}

	return nil
}
