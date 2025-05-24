package engine

import (
	"database/sql"
	"net/http"
	"os"
	"sync"

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
	Handler *HandlerInfo         // pre-registered handler info (nil for direct code execution)
	Code    string               // JavaScript code to execute
	W       http.ResponseWriter  // response writer
	R       *http.Request        // request
	Done    chan error           // completion signal
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

// executeCode executes JavaScript code directly
func (e *Engine) executeCode(code string) error {
	_, err := e.rt.RunString(code)
	return err
}