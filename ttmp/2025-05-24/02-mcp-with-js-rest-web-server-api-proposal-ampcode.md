# MCP JavaScript REST Web Server - API Design Proposal

## Overview

A Go web server that embeds a JavaScript runtime (Goja) allowing dynamic REST endpoint registration and file serving through JavaScript code execution. The system provides persistent storage, state management, and seamless integration between Go and JavaScript.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP Client   │ ── │   Go Web Server  │ ── │  JavaScript VM  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │  Route Registry  │    │   SQLite DB     │
                       └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │  File Registry   │    │  Code Archive   │
                       └──────────────────┘    └─────────────────┘
```

## Core Components

### 1. Go Web Server (`pkg/jsserver`)

```go
type JSWebServer struct {
    vm           *goja.Runtime
    db           *sql.DB
    routes       map[string]JSRoute
    files        map[string]JSFile
    globalState  map[string]interface{}
    codeArchive  string // directory to store executed code
}

type JSRoute struct {
    Method   string
    Path     string
    Handler  goja.Value // JavaScript function
    Created  time.Time
}

type JSFile struct {
    Path     string
    Generator goja.Value // JavaScript function
    MimeType string
    Created  time.Time
}
```

### 2. JavaScript Runtime Integration

#### Core APIs Exposed to JavaScript

```javascript
// SQLite Access
const db = {
    exec(sql, ...params),           // Execute SQL with params
    query(sql, ...params),          // Query with results
    prepare(sql),                   // Prepare statement
    transaction(fn)                 // Execute in transaction
};

// Route Registration
const server = {
    get(path, handler),             // Register GET route
    post(path, handler),            // Register POST route
    put(path, handler),             // Register PUT route
    delete(path, handler),          // Register DELETE route
    any(path, handler),             // Register any method route
    
    // File serving
    file(path, generator, mimeType), // Register file generator
    static(path, content, mimeType)  // Register static content
};

// Global State Management
const state = {
    get(key),                       // Get global state
    set(key, value),               // Set global state
    delete(key),                   // Delete state key
    clear(),                       // Clear all state
    keys()                         // List all keys
};

// Utilities
const utils = {
    log(...args),                  // Logging
    sleep(ms),                     // Async sleep
    fetch(url, options),           // HTTP client
    uuid(),                        // Generate UUID
    now(),                         // Current timestamp
    env(key)                       // Environment variables
};
```

### 3. External API Endpoints

#### Code Execution API
```
POST /api/execute
Content-Type: application/json

{
    "code": "server.get('/hello', (req, res) => { res.json({message: 'Hello World'}) })",
    "persist": true,        // Save to code archive
    "name": "hello-endpoint" // Optional name for the code
}

Response:
{
    "success": true,
    "result": "Route registered: GET /hello",
    "execution_id": "2025-05-24T10:30:15Z-abc123",
    "archived_file": "/archive/2025-05-24T10:30:15Z-abc123.js"
}
```

#### Route Management API
```
GET /api/routes              // List all registered routes
DELETE /api/routes/{path}    // Remove route
GET /api/files              // List all registered files
DELETE /api/files/{path}    // Remove file
```

#### State Management API
```
GET /api/state              // Get all global state
GET /api/state/{key}        // Get specific state
PUT /api/state/{key}        // Set state value
DELETE /api/state/{key}     // Delete state
```

## Implementation Details

### 1. JavaScript Context Setup

```go
func (s *JSWebServer) initializeJSContext() error {
    // Set up SQLite bindings
    s.vm.Set("db", s.createDBObject())
    
    // Set up server bindings
    s.vm.Set("server", s.createServerObject())
    
    // Set up state management
    s.vm.Set("state", s.createStateObject())
    
    // Set up utilities
    s.vm.Set("utils", s.createUtilsObject())
    
    return nil
}
```

### 2. Route Handler Wrapper

```go
func (s *JSWebServer) wrapJSHandler(jsFunc goja.Value) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Create request/response objects for JS
        req := s.createJSRequest(r)
        res := s.createJSResponse(w)
        
        // Call JavaScript function
        _, err := jsFunc.Call(goja.Undefined(), req, res)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}
```

### 3. Code Persistence

```go
func (s *JSWebServer) archiveCode(code string) (string, error) {
    timestamp := time.Now().Format("2006-01-02T15:04:05Z")
    uuid := generateUUID()
    filename := fmt.Sprintf("%s-%s.js", timestamp, uuid)
    filepath := path.Join(s.codeArchive, filename)
    
    return filepath, os.WriteFile(filepath, []byte(code), 0644)
}
```

### 4. Database Schema

```sql
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
```

## Security Considerations

### 1. JavaScript Sandbox
- Disable dangerous APIs (file system, process, network outside of provided utils)
- Implement timeout for code execution
- Memory limits for JavaScript runtime
- Whitelist allowed packages/modules

### 2. API Security
- Authentication/authorization for `/api/execute`
- Rate limiting on code execution
- Input validation and sanitization
- CORS configuration

### 3. Database Access
- Prepared statements only
- Query timeout limits
- Connection pooling
- Transaction isolation

## Usage Examples

### 1. Simple REST API
```javascript
// Register a simple GET endpoint
server.get('/api/users', (req, res) => {
    const users = db.query('SELECT * FROM users');
    res.json(users);
});

// Register POST endpoint with validation
server.post('/api/users', (req, res) => {
    const { name, email } = req.body;
    
    if (!name || !email) {
        return res.status(400).json({ error: 'Name and email required' });
    }
    
    const result = db.exec(
        'INSERT INTO users (name, email) VALUES (?, ?)',
        name, email
    );
    
    res.json({ id: result.lastInsertId, name, email });
});
```

### 2. Dynamic File Generation
```javascript
// Generate CSV reports
server.file('/reports/users.csv', (req, res) => {
    const users = db.query('SELECT * FROM users');
    const csv = users.map(u => `${u.name},${u.email}`).join('\n');
    return csv;
}, 'text/csv');

// Generate JSON config
server.file('/config.json', () => {
    return JSON.stringify({
        version: state.get('app_version') || '1.0.0',
        timestamp: utils.now(),
        users_count: db.query('SELECT COUNT(*) as count FROM users')[0].count
    });
}, 'application/json');
```

### 3. Stateful Applications
```javascript
// Initialize counter
if (!state.get('counter')) {
    state.set('counter', 0);
}

server.get('/counter', (req, res) => {
    res.json({ counter: state.get('counter') });
});

server.post('/counter/increment', (req, res) => {
    const current = state.get('counter') || 0;
    const newValue = current + 1;
    state.set('counter', newValue);
    res.json({ counter: newValue });
});
```

## Project Structure

```
pkg/jsserver/
├── server.go           // Main server implementation
├── vm.go              // JavaScript VM setup and management
├── bindings.go        // JavaScript API bindings
├── routes.go          // Route management
├── storage.go         // Database and state management
├── archive.go         // Code archiving
└── handlers.go        // HTTP handlers

cmd/js-web-server/
└── main.go            // CLI entry point

examples/js-web-server/
├── basic-api.js       // Example REST API
├── file-server.js     // Example file serving
└── stateful-app.js    // Example stateful application
```

## Development Phases

### Phase 1: Core Infrastructure
- Go web server with Goja integration
- Basic JavaScript API bindings
- SQLite integration
- Code execution endpoint

### Phase 2: Route Management
- Dynamic route registration
- Request/response objects
- Route persistence and management

### Phase 3: File Serving
- Dynamic file generation
- Static content serving
- MIME type handling

### Phase 4: State Management
- Global state API
- Database persistence
- State management endpoints
