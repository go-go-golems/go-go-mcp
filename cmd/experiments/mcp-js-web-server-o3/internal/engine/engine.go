package engine

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/dop251/goja"
	_ "github.com/mattn/go-sqlite3"
)

// Engine wraps the JavaScript runtime and SQLite connection
type Engine struct {
	rt       *goja.Runtime
	db       *sql.DB
	jobs     chan EvalJob
	handlers map[string]map[string]goja.Callable // [path][method] -> handler
	files    map[string]goja.Callable            // [path] -> file handler
	mu       sync.RWMutex
}

// EvalJob represents a JavaScript evaluation job
type EvalJob struct {
	Fn   goja.Callable        // pre-registered closure (nil for direct code execution)
	Code string               // JavaScript code to execute
	W    http.ResponseWriter  // response writer
	R    *http.Request        // request
	Done chan error           // completion signal
}

// NewEngine creates a new JavaScript engine with SQLite integration
func NewEngine(dbPath string) *Engine {
	rt := goja.New()
	
	// Open SQLite connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to open database: %v", err))
	}

	e := &Engine{
		rt:       rt,
		db:       db,
		jobs:     make(chan EvalJob, 1024),
		handlers: make(map[string]map[string]goja.Callable),
		files:    make(map[string]goja.Callable),
	}

	// Setup JavaScript bindings
	e.setupBindings()

	return e
}

// Init loads and executes a bootstrap JavaScript file
func (e *Engine) Init(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create default bootstrap file
		bootstrap := `let globalCounter = 0;

registerHandler("GET", "/", () => "JS playground online");
registerHandler("GET", "/health", () => ({ok: true, counter: globalCounter}));
registerHandler("POST", "/counter", () => ({count: ++globalCounter}));

console.log("Bootstrap complete - server ready");`
		
		if err := os.WriteFile(filename, []byte(bootstrap), 0644); err == nil {
			return e.executeCode(bootstrap)
		}
		return err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return e.executeCode(string(data))
}

// GetHandler returns a registered HTTP handler
func (e *Engine) GetHandler(method, path string) (goja.Callable, bool) {
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

// executeCode executes JavaScript code directly
func (e *Engine) executeCode(code string) error {
	_, err := e.rt.RunString(code)
	return err
}