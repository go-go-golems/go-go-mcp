package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// JSPlayground represents the main application structure
type JSPlayground struct {
	vm           *goja.Runtime
	db           *sql.DB
	router       *mux.Router
	handlerMap   map[string]*JSHandler
	fileMap      map[string]*JSFileHandler
	mu           sync.RWMutex
	executionLog []ExecutionRecord
	globalState  map[string]interface{}
	stateMu      sync.RWMutex
}

// JSHandler represents a JavaScript function registered as HTTP handler
type JSHandler struct {
	Function goja.Value
	Method   string
	Path     string
}

// JSFileHandler represents a JavaScript function that serves files
type JSFileHandler struct {
	Function goja.Value
	Path     string
	MimeType string
}

// ExecutionRecord tracks executed JavaScript code
type ExecutionRecord struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// JSContext provides JavaScript APIs
type JSContext struct {
	playground *JSPlayground
}

// NewJSPlayground creates a new JavaScript playground instance
func NewJSPlayground() (*JSPlayground, error) {
	// Initialize SQLite database
	db, err := sql.Open("sqlite3", "./playground.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create tables
	if err := initDatabase(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	// Initialize goja VM
	vm := goja.New()

	playground := &JSPlayground{
		vm:           vm,
		db:           db,
		router:       mux.NewRouter(),
		handlerMap:   make(map[string]*JSHandler),
		fileMap:      make(map[string]*JSFileHandler),
		globalState:  make(map[string]interface{}),
		executionLog: make([]ExecutionRecord, 0),
	}

	// Setup JavaScript APIs
	if err := playground.setupJSAPIs(); err != nil {
		return nil, fmt.Errorf("failed to setup JS APIs: %v", err)
	}

	// Setup HTTP routes
	playground.setupRoutes()

	// Load JavaScript files from startup directory
	if err := playground.loadJSFolder("./js-apps"); err != nil {
		log.Warn().Err(err).Msg("Failed to load JS folder")
	}

	return playground, nil
}

// setupJSAPIs configures the JavaScript environment with APIs
func (p *JSPlayground) setupJSAPIs() error {
	ctx := &JSContext{playground: p}

	// Register global APIs
	p.vm.Set("console", map[string]interface{}{
		"log":   ctx.consoleLog,
		"error": ctx.consoleError,
	})

	// Database APIs
	p.vm.Set("db", map[string]interface{}{
		"query":  ctx.dbQuery,
		"exec":   ctx.dbExec,
		"get":    ctx.dbGet,
		"all":    ctx.dbAll,
		"begin":  ctx.dbBegin,
		"commit": ctx.dbCommit,
	})

	// HTTP Handler APIs
	p.vm.Set("registerHandler", ctx.registerHandler)
	p.vm.Set("registerFileHandler", ctx.registerFileHandler)

	// State management APIs
	p.vm.Set("setState", ctx.setState)
	p.vm.Set("getState", ctx.getState)
	p.vm.Set("clearState", ctx.clearState)

	// Utility APIs
	p.vm.Set("fetch", ctx.fetch)
	p.vm.Set("setTimeout", ctx.setTimeout)

	return nil
}

// setupRoutes configures HTTP routes
func (p *JSPlayground) setupRoutes() {
	// API routes
	p.router.HandleFunc("/api/execute", p.executeJSHandler).Methods("POST")
	p.router.HandleFunc("/api/executions", p.getExecutionsHandler).Methods("GET")
	p.router.HandleFunc("/api/handlers", p.getHandlersHandler).Methods("GET")
	p.router.HandleFunc("/api/state", p.getStateHandler).Methods("GET")
	p.router.HandleFunc("/api/state", p.clearStateHandler).Methods("DELETE")

	// Dynamic JavaScript handlers
	p.router.PathPrefix("/js/").HandlerFunc(p.dynamicHandler)
	p.router.PathPrefix("/files/").HandlerFunc(p.fileHandler)

	// Static file serving
	p.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
}

// executeJSHandler handles JavaScript code execution requests
func (p *JSPlayground) executeJSHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result := p.ExecuteJS(request.Code)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ExecuteJS executes JavaScript code and logs execution
func (p *JSPlayground) ExecuteJS(code string) ExecutionRecord {
	p.mu.Lock()
	defer p.mu.Unlock()

	record := ExecutionRecord{
		ID:        len(p.executionLog) + 1,
		Code:      code,
		Timestamp: time.Now(),
		Success:   true,
	}

	// Execute JavaScript
	_, err := p.vm.RunString(code)
	if err != nil {
		record.Success = false
		record.Error = err.Error()
	}

	// Log execution to file
	p.logExecution(record)

	// Add to execution log
	p.executionLog = append(p.executionLog, record)

	return record
}

// logExecution writes execution record to file
func (p *JSPlayground) logExecution(record ExecutionRecord) {
	timestamp := record.Timestamp.Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("executions/exec_%d_%s.js", record.ID, timestamp)

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(filename), 0755)

	// Write code to file
	content := fmt.Sprintf("// Execution ID: %d\n// Timestamp: %s\n// Success: %t\n",
		record.ID, record.Timestamp.Format(time.RFC3339), record.Success)
	if record.Error != "" {
		content += fmt.Sprintf("// Error: %s\n", record.Error)
	}
	content += "\n" + record.Code

	os.WriteFile(filename, []byte(content), 0644)
}

// dynamicHandler handles requests to JavaScript-registered endpoints
func (p *JSPlayground) dynamicHandler(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	handler, exists := p.handlerMap[r.URL.Path]
	p.mu.RUnlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	// Check HTTP method
	if handler.Method != r.Method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Call JavaScript function
	p.callJSHandler(w, r, handler.Function)
}

// fileHandler serves files from JavaScript-registered file handlers
func (p *JSPlayground) fileHandler(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	handler, exists := p.fileMap[r.URL.Path]
	p.mu.RUnlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	// Set content type
	if handler.MimeType != "" {
		w.Header().Set("Content-Type", handler.MimeType)
	}

	// Call JavaScript function to get file content
	p.callJSFileHandler(w, r, handler.Function)
}

// callJSHandler executes a JavaScript handler function
func (p *JSPlayground) callJSHandler(w http.ResponseWriter, r *http.Request, fn goja.Value) {
	// Create request object for JavaScript
	reqObj := p.vm.NewObject()
	reqObj.Set("method", r.Method)
	reqObj.Set("url", r.URL.String())
	reqObj.Set("path", r.URL.Path)

	// Parse body if present
	if r.Body != nil {
		body, _ := io.ReadAll(r.Body)
		reqObj.Set("body", string(body))
	}

	// Parse query parameters
	queryObj := p.vm.NewObject()
	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			queryObj.Set(key, values[0])
		} else {
			queryObj.Set(key, values)
		}
	}
	reqObj.Set("query", queryObj)

	// Parse headers
	headersObj := p.vm.NewObject()
	for key, values := range r.Header {
		headersObj.Set(key, values[0])
	}
	reqObj.Set("headers", headersObj)

	// Create response object
	resObj := p.vm.NewObject()
	resObj.Set("status", func(code int) { w.WriteHeader(code) })
	resObj.Set("header", func(key, value string) { w.Header().Set(key, value) })
	resObj.Set("json", func(data interface{}) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})
	resObj.Set("text", func(text string) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(text))
	})
	resObj.Set("html", func(html string) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// Call the JavaScript function
	if callable, ok := goja.AssertFunction(fn); ok {
		callable(nil, reqObj, resObj)
	}
}

// callJSFileHandler executes a JavaScript file handler function
func (p *JSPlayground) callJSFileHandler(w http.ResponseWriter, r *http.Request, fn goja.Value) {
	// Create request object for JavaScript
	reqObj := p.vm.NewObject()
	reqObj.Set("method", r.Method)
	reqObj.Set("url", r.URL.String())
	reqObj.Set("path", r.URL.Path)

	// Create response object
	resObj := p.vm.NewObject()
	resObj.Set("text", func(text string) {
		w.Write([]byte(text))
	})
	resObj.Set("html", func(html string) {
		w.Write([]byte(html))
	})

	// Call the JavaScript function
	if callable, ok := goja.AssertFunction(fn); ok {
		callable(nil, reqObj, resObj)
	}
}

// JavaScript API implementations

// consoleLog implements console.log for JavaScript
func (ctx *JSContext) consoleLog(args ...interface{}) {
	log.Info().Interface("args", args).Msg("JS console.log")
}

// consoleError implements console.error for JavaScript
func (ctx *JSContext) consoleError(args ...interface{}) {
	log.Error().Interface("args", args).Msg("JS console.error")
}

// registerHandler registers a JavaScript function as HTTP handler
func (ctx *JSContext) registerHandler(method, path string, handler goja.Value) {
	ctx.playground.mu.Lock()
	defer ctx.playground.mu.Unlock()

	fullPath := "/js" + path
	ctx.playground.handlerMap[fullPath] = &JSHandler{
		Function: handler,
		Method:   method,
		Path:     fullPath,
	}

	log.Info().Str("method", method).Str("path", fullPath).Msg("Registered HTTP handler")
}

// registerFileHandler registers a JavaScript function as file handler
func (ctx *JSContext) registerFileHandler(path string, handler goja.Value, mimeType string) {
	ctx.playground.mu.Lock()
	defer ctx.playground.mu.Unlock()

	fullPath := "/files" + path
	ctx.playground.fileMap[fullPath] = &JSFileHandler{
		Function: handler,
		Path:     fullPath,
		MimeType: mimeType,
	}

	log.Info().Str("path", fullPath).Str("mime_type", mimeType).Msg("Registered file handler")
}

// setState sets a value in global state
func (ctx *JSContext) setState(key string, value interface{}) {
	ctx.playground.stateMu.Lock()
	defer ctx.playground.stateMu.Unlock()

	ctx.playground.globalState[key] = value
}

// getState gets a value from global state
func (ctx *JSContext) getState(key string) interface{} {
	ctx.playground.stateMu.RLock()
	defer ctx.playground.stateMu.RUnlock()

	return ctx.playground.globalState[key]
}

// clearState clears global state
func (ctx *JSContext) clearState() {
	ctx.playground.stateMu.Lock()
	defer ctx.playground.stateMu.Unlock()

	ctx.playground.globalState = make(map[string]interface{})
}

// Database API implementations
func (ctx *JSContext) dbQuery(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := ctx.playground.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

func (ctx *JSContext) dbExec(query string, args ...interface{}) error {
	_, err := ctx.playground.db.Exec(query, args...)
	return err
}

func (ctx *JSContext) dbGet(query string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := ctx.playground.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, col := range columns {
		result[col] = values[i]
	}

	return result, nil
}

func (ctx *JSContext) dbAll(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return ctx.dbQuery(query, args...)
}

// Transaction represents a database transaction
type Transaction struct {
	tx *sql.Tx
}

func (ctx *JSContext) dbBegin() (*Transaction, error) {
	tx, err := ctx.playground.db.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

func (ctx *JSContext) dbCommit(tx *Transaction) error {
	return tx.tx.Commit()
}

// Additional JavaScript APIs
func (ctx *JSContext) fetch(url string, options ...map[string]interface{}) (map[string]interface{}, error) {
	// Simplified fetch implementation
	// In a real implementation, you'd use net/http to make requests
	return map[string]interface{}{
		"status": 200,
		"data":   "Mock response from " + url,
	}, nil
}

func (ctx *JSContext) setTimeout(callback goja.Value, delay int) {
	// Simplified setTimeout - in production you'd use a proper scheduler
	go func() {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		if callable, ok := goja.AssertFunction(callback); ok {
			callable(nil)
		}
	}()
}

// initDatabase creates necessary tables
func initDatabase(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS js_state (
		key TEXT PRIMARY KEY,
		value TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS js_executions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		success BOOLEAN,
		error TEXT,
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(schema)
	return err
}

// HTTP handlers for API endpoints
func (p *JSPlayground) getExecutionsHandler(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p.executionLog)
}

func (p *JSPlayground) getHandlersHandler(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	handlers := make([]map[string]string, 0)
	for path, handler := range p.handlerMap {
		handlers = append(handlers, map[string]string{
			"method": handler.Method,
			"path":   path,
		})
	}

	for path := range p.fileMap {
		handlers = append(handlers, map[string]string{
			"type": "file",
			"path": path,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(handlers)
}

func (p *JSPlayground) getStateHandler(w http.ResponseWriter, r *http.Request) {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p.globalState)
}

func (p *JSPlayground) clearStateHandler(w http.ResponseWriter, r *http.Request) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	p.globalState = make(map[string]interface{})
	w.WriteHeader(http.StatusNoContent)
}

// loadJSFolder loads all JavaScript files from a directory on startup
func (p *JSPlayground) loadJSFolder(folderPath string) error {
	log.Info().Str("folder", folderPath).Msg("Scanning for JavaScript files")

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		log.Warn().Str("folder", folderPath).Msg("JS folder does not exist, skipping")
		return fmt.Errorf("JS folder does not exist: %s", folderPath)
	}

	var loadedFiles []string
	var failedFiles []string

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().Str("path", path).Err(err).Msg("Error walking directory")
			return err
		}

		// Skip directories and non-JS files
		if info.IsDir() || filepath.Ext(path) != ".js" {
			return nil
		}

		// Read JavaScript file
		content, err := os.ReadFile(path)
		if err != nil {
			log.Error().Str("file", path).Err(err).Msg("Error reading JS file")
			failedFiles = append(failedFiles, path)
			return nil // Continue with other files
		}

		// Execute the JavaScript file
		log.Info().
			Str("file", path).
			Int("size_bytes", len(content)).
			Msg("Loading JavaScript file")

		record := p.ExecuteJS(string(content))
		if !record.Success {
			log.Error().
				Str("file", path).
				Str("error", record.Error).
				Int("execution_id", record.ID).
				Msg("Failed to execute JS file")
			failedFiles = append(failedFiles, path)
		} else {
			log.Debug().
				Str("file", path).
				Int("execution_id", record.ID).
				Msg("Successfully executed JS file")
			loadedFiles = append(loadedFiles, path)
		}

		return nil
	})

	// Log summary
	log.Info().
		Int("loaded", len(loadedFiles)).
		Int("failed", len(failedFiles)).
		Strs("loaded_files", loadedFiles).
		Msg("JavaScript file loading complete")

	if len(failedFiles) > 0 {
		log.Warn().Strs("failed_files", failedFiles).Msg("Some JavaScript files failed to load")
	}

	return err
}

// Start starts the web server
func (p *JSPlayground) Start(addr string) error {
	// Log detailed startup information
	p.mu.RLock()
	handlerCount := len(p.handlerMap)
	fileHandlerCount := len(p.fileMap)

	// Log all registered HTTP handlers
	if handlerCount > 0 {
		log.Info().Msg("Registered HTTP handlers:")
		for path, handler := range p.handlerMap {
			log.Info().
				Str("method", handler.Method).
				Str("path", path).
				Msg("  HTTP handler")
		}
	}

	// Log all registered file handlers
	if fileHandlerCount > 0 {
		log.Info().Msg("Registered file handlers:")
		for path, handler := range p.fileMap {
			log.Info().
				Str("path", path).
				Str("mime_type", handler.MimeType).
				Msg("  File handler")
		}
	}
	p.mu.RUnlock()

	// Database info
	var dbStats struct {
		Posts int `json:"posts"`
		Todos int `json:"todos"`
	}

	if posts, err := p.db.Query("SELECT COUNT(*) as count FROM posts"); err == nil {
		defer posts.Close()
		if posts.Next() {
			var count int64
			posts.Scan(&count)
			dbStats.Posts = int(count)
		}
	}

	if todos, err := p.db.Query("SELECT COUNT(*) as count FROM todos"); err == nil {
		defer todos.Close()
		if todos.Next() {
			var count int64
			todos.Scan(&count)
			dbStats.Todos = int(count)
		}
	}

	log.Info().
		Int("posts", dbStats.Posts).
		Int("todos", dbStats.Todos).
		Msg("Database statistics")

	// Server configuration
	host := "localhost"
	if addr[0] == ':' {
		host = "localhost" + addr
	} else {
		host = addr
	}

	log.Info().
		Str("address", addr).
		Str("host", host).
		Int("http_handlers", handlerCount).
		Int("file_handlers", fileHandlerCount).
		Msg("Starting JavaScript Playground server")

	// Log available web interfaces with full URLs
	baseURL := "http://" + host
	log.Info().Msg("Available web interfaces:")
	log.Info().Str("url", baseURL+"/files/dashboard.html").Msg("  📊 Dashboard - Main interface")
	log.Info().Str("url", baseURL+"/files/blog.html").Msg("  📝 Blog - Post management")
	log.Info().Str("url", baseURL+"/files/todos.html").Msg("  ✅ Todos - Task management")

	// Log API endpoints
	log.Info().Msg("Available API endpoints:")
	log.Info().Str("endpoint", "POST "+baseURL+"/api/execute").Msg("  Execute JavaScript code")
	log.Info().Str("endpoint", "GET "+baseURL+"/api/executions").Msg("  Get execution history")
	log.Info().Str("endpoint", "GET "+baseURL+"/api/handlers").Msg("  List registered handlers")
	log.Info().Str("endpoint", "GET "+baseURL+"/api/state").Msg("  Get global state")

	log.Info().
		Str("address", addr).
		Str("pid", fmt.Sprintf("%d", os.Getpid())).
		Msg("🚀 Server is ready and listening for connections")

	return http.ListenAndServe(addr, p.router)
}

func main() {
	// Parse command line flags
	var logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	var port = flag.String("port", "8080", "Server port")
	var jsFolder = flag.String("js-folder", "./js-apps", "Directory containing JavaScript applications")
	flag.Parse()

	// Setup zerolog
	setupLogging(*logLevel)

	log.Info().
		Str("version", "1.0.0").
		Str("port", *port).
		Str("log_level", *logLevel).
		Str("js_folder", *jsFolder).
		Msg("🎮 JavaScript Playground Server initializing")

	// Log system information
	log.Info().
		Str("go_version", fmt.Sprintf("go%s", os.Getenv("GOVERSION"))).
		Str("pid", fmt.Sprintf("%d", os.Getpid())).
		Str("working_dir", mustGetWorkingDir()).
		Msg("System information")

	playground, err := NewJSPlayground()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create playground")
	}

	// Override JS folder if specified
	if *jsFolder != "./js-apps" {
		log.Info().Str("folder", *jsFolder).Msg("Using custom JavaScript folder")
		if err := playground.loadJSFolder(*jsFolder); err != nil {
			log.Warn().Err(err).Msg("Failed to load custom JS folder")
		}
	}

	addr := ":" + *port
	if err := playground.Start(addr); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

// mustGetWorkingDir returns the current working directory or "unknown"
func mustGetWorkingDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "unknown"
}

// setupLogging configures zerolog with the specified level
func setupLogging(level string) {
	// Setup console writer with colors and timestamps
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	})

	// Set log level
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		log.Warn().Str("level", level).Msg("Unknown log level, defaulting to info")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Debug().Str("level", level).Msg("Logging configured")
}
