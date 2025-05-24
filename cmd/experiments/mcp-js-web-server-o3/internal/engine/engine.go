package engine

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dop251/goja"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// Engine wraps the JavaScript runtime and SQLite connection
type Engine struct {
	rt       *goja.Runtime
	db       *sql.DB
	jobs     chan EvalJob
	handlers map[string]map[string]*HandlerInfo // [path][method] -> handler info
	files    map[string]goja.Callable           // [path] -> file handler
	mu       sync.RWMutex
}

// HandlerInfo contains handler function and metadata
type HandlerInfo struct {
	Fn          goja.Callable // JavaScript function
	ContentType string        // MIME type override
}

// EvalJob represents a JavaScript evaluation job
type EvalJob struct {
	Handler   *HandlerInfo        // pre-registered handler info (nil for direct code execution)
	Code      string              // JavaScript code to execute
	W         http.ResponseWriter // response writer
	R         *http.Request       // request
	Done      chan error          // completion signal
	Result    chan *EvalResult    // result channel for capturing execution results
	SessionID string              // session identifier for tracking
	Source    string              // source of execution ('api', 'mcp', 'file')
}

// EvalResult contains the result of JavaScript execution
type EvalResult struct {
	Value      interface{} `json:"value"`           // The actual result value
	ConsoleLog []string    `json:"consoleLog"`      // Captured console output
	Error      error       `json:"error,omitempty"` // Execution error if any
}

// NewEngine creates a new JavaScript engine with SQLite integration
func NewEngine(dbPath string) *Engine {
	log.Debug().Str("database", dbPath).Msg("Creating new JavaScript engine")

	rt := goja.New()

	// Open SQLite connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal().Err(err).Str("database", dbPath).Msg("Failed to open database")
	}
	log.Debug().Str("database", dbPath).Msg("Database connection established")

	e := &Engine{
		rt:       rt,
		db:       db,
		jobs:     make(chan EvalJob, 1024),
		handlers: make(map[string]map[string]*HandlerInfo),
		files:    make(map[string]goja.Callable),
	}

	// Setup JavaScript bindings
	log.Debug().Msg("Setting up JavaScript bindings")
	e.setupBindings()

	// Initialize database tables for script storage
	if err := e.initScriptsTables(); err != nil {
		log.Warn().Err(err).Msg("Failed to initialize scripts tables")
	}

	log.Debug().Msg("JavaScript engine initialized")
	return e
}

// Init loads and executes a bootstrap JavaScript file
func (e *Engine) Init(filename string) error {
	log.Debug().Str("file", filename).Msg("Initializing JavaScript engine with bootstrap file")

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Debug().Str("file", filename).Msg("Bootstrap file doesn't exist, creating default")

		// Create default bootstrap file
		bootstrap := `let globalCounter = 0;

registerHandler("GET", "/", () => "JS playground online");
registerHandler("GET", "/health", () => ({ok: true, counter: globalCounter}));
registerHandler("POST", "/counter", () => ({count: ++globalCounter}));

console.log("Bootstrap complete - server ready");`

		if err := os.WriteFile(filename, []byte(bootstrap), 0644); err == nil {
			log.Debug().Str("file", filename).Msg("Created default bootstrap file")
			return e.executeCode(bootstrap)
		}
		log.Error().Err(err).Str("file", filename).Msg("Failed to create bootstrap file")
		return err
	}

	log.Debug().Str("file", filename).Msg("Loading existing bootstrap file")
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Error().Err(err).Str("file", filename).Msg("Failed to read bootstrap file")
		return err
	}

	return e.executeCode(string(data))
}

// GetHandler returns a registered HTTP handler
func (e *Engine) GetHandler(method, path string) (*HandlerInfo, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if methods, exists := e.handlers[path]; exists {
		if handler, exists := methods[method]; exists {
			return handler, true
		}
	}
	return nil, false
}

// GetFileHandler returns a registered file handler
func (e *Engine) GetFileHandler(path string) (goja.Callable, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	handler, exists := e.files[path]
	return handler, exists
}

// SubmitJob submits a job to the dispatcher
func (e *Engine) SubmitJob(job EvalJob) {
	e.jobs <- job
}

// executeCode executes JavaScript code directly, wrapped in a function scope
func (e *Engine) executeCode(code string) error {
	// Wrap code in an IIFE to prevent variable conflicts on re-execution
	wrappedCode := `(function() {
		"use strict";
		` + code + `
	})();`

	log.Debug().Str("wrapped_code", wrappedCode).Msg("Executing wrapped JavaScript code")
	_, err := e.rt.RunString(wrappedCode)
	return err
}

// executeCodeWithResult executes JavaScript code and captures the result and console output
func (e *Engine) executeCodeWithResult(code string) (*EvalResult, error) {
	result := &EvalResult{
		ConsoleLog: []string{},
	}

	// Temporarily capture console output
	originalConsole := e.captureConsole(result)
	defer e.restoreConsole(originalConsole)

	// Wrap code in an IIFE to prevent variable conflicts on re-execution
	wrappedCode := `(function() {
		"use strict";
		return (function() {
			` + code + `
		})();
	})();`

	log.Debug().Str("wrapped_code", wrappedCode).Msg("Executing wrapped JavaScript code with result capture")

	value, err := e.rt.RunString(wrappedCode)
	if err != nil {
		result.Error = err
		return result, err
	}

	// Export the result to a Go-friendly format
	if value != nil && !goja.IsUndefined(value) {
		result.Value = value.Export()
	}

	return result, nil
}

// ScriptExecution represents a stored script execution record
type ScriptExecution struct {
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	Code       string    `json:"code"`
	Result     string    `json:"result"`
	ConsoleLog string    `json:"console_log"`
	Error      string    `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source"` // 'api', 'mcp', 'file'
}

// initScriptsTables initializes the database tables for script storage
func (e *Engine) initScriptsTables() error {
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
	
	_, err := e.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create scripts tables: %w", err)
	}
	
	log.Debug().Msg("Script storage tables initialized")
	return nil
}

// StoreScriptExecution stores a script execution in the database
func (e *Engine) StoreScriptExecution(sessionID, code, result, consoleLog, errorMsg, source string) error {
	query := `
	INSERT INTO script_executions (session_id, code, result, console_log, error, source)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	
	_, err := e.db.Exec(query, sessionID, code, result, consoleLog, errorMsg, source)
	if err != nil {
		return fmt.Errorf("failed to store script execution: %w", err)
	}
	
	return nil
}

// GetScriptExecutions retrieves script executions with pagination and search
func (e *Engine) GetScriptExecutions(search, sessionID string, limit, offset int) ([]ScriptExecution, int, error) {
	// Build WHERE clause
	var whereClause string
	var args []interface{}
	var conditions []string
	
	if search != "" {
		conditions = append(conditions, "(code LIKE ? OR result LIKE ? OR console_log LIKE ?)")
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}
	
	if sessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, sessionID)
	}
	
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM script_executions " + whereClause
	var total int
	err := e.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}
	
	// Get paginated results
	query := fmt.Sprintf(`
	SELECT id, session_id, code, result, console_log, error, timestamp, source 
	FROM script_executions %s
	ORDER BY timestamp DESC 
	LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query script executions: %w", err)
	}
	defer rows.Close()
	
	var executions []ScriptExecution
	for rows.Next() {
		var exec ScriptExecution
		var result, consoleLog, errorStr sql.NullString
		
		err := rows.Scan(&exec.ID, &exec.SessionID, &exec.Code, &result, &consoleLog, &errorStr, &exec.Timestamp, &exec.Source)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		
		exec.Result = result.String
		exec.ConsoleLog = consoleLog.String
		exec.Error = errorStr.String
		
		executions = append(executions, exec)
	}
	
	return executions, total, nil
}
