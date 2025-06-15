# JavaScript Web Server Implementation Analysis

## Overview

The `js-web-server` functionality in go-go-mcp provides two different approaches for running JavaScript-based web servers:

1. **MCP Version**: Full-featured engine with advanced routing, path parameters, and production capabilities
2. **REPL Version**: Simplified implementation for interactive development and testing

## MCP Version Architecture

### Core Components

#### 1. Engine (`internal/engine/engine.go`)
- **JavaScript Runtime**: Uses Goja with Node.js event loop support
- **Handler Registry**: `map[string]map[string]*HandlerInfo` stores handlers by `[path][method]`
- **Path Pattern Matching**: Supports Express.js-style parameters like `/users/:id`
- **Database Integration**: Direct SQLite access with connection pooling
- **Event Loop**: Async operation support via `goja_nodejs/eventloop`
- **Request Logging**: Built-in admin interface with request/response tracking

#### 2. Handler System (`internal/engine/handlers.go`)
- **Express.js Compatibility**: Full request/response object API
- **Content-Type Detection**: Automatic HTML/JSON/plain text detection
- **HTTP Methods**: GET, POST, PUT, DELETE, PATCH support
- **Middleware Support**: Via `app.use()` functionality
- **Cookie/Header Management**: Full HTTP feature support

#### 3. Route Dispatching (`internal/web/router.go`, `internal/engine/dispatcher.go`)
- **Single-threaded JavaScript Execution**: Via job queue for thread safety
- **Pattern Matching**: Reuses paths with parameter extraction
- **Error Recovery**: Panic recovery and proper error responses
- **Admin Interface**: Request monitoring and debugging tools

#### 4. JavaScript Bindings (`internal/engine/bindings.go`)
- **Database Access**: `db.query()` and `db.exec()` methods
- **Console Functions**: Full console API with request-scoped logging
- **Global State**: Persistent state across requests via `globalState`
- **Geppetto APIs**: Conversation, ChatStepFactory, and AI integration

### Path Reuse and Pattern Matching

The MCP version supports sophisticated path reuse through:

```javascript
// Register handlers with parameters
app.get('/users/:id', (req, res) => {
    const userId = req.params.id;
    // Handle /users/123, /users/456, etc.
});

app.get('/posts/:postId/comments/:commentId', (req, res) => {
    const { postId, commentId } = req.params;
    // Handle /posts/1/comments/5, etc.
});
```

**Implementation Details:**
- **Pattern Storage**: Handlers stored with original pattern as key
- **Runtime Matching**: `pathMatches()` function compares patterns to incoming paths
- **Parameter Extraction**: `parsePathParams()` extracts variables from URLs
- **Exact Match Priority**: Exact paths checked before pattern matching

### Request/Response Objects

**Request Object Features:**
```javascript
{
    method: "get",           // lowercase HTTP method
    url: "/users/123?sort=name",
    path: "/users/123",
    query: { sort: "name" }, // parsed query parameters
    headers: { ... },        // lowercase header names
    body: { ... },          // parsed JSON or raw string
    cookies: { ... },       // cookie name-value pairs
    params: { id: "123" },  // extracted path parameters
    ip: "192.168.1.1",      // client IP
    protocol: "https",
    hostname: "example.com"
}
```

**Response Object Features:**
```javascript
res.status(200)
   .set('Content-Type', 'application/json')
   .cookie('session', 'abc123', { httpOnly: true })
   .json({ success: true });

res.redirect(302, '/login');
res.send('Hello World');
res.end();
```

## REPL Version Implementation

### Current State
The REPL version provides a basic implementation with:

- **Basic Routing**: Simple path-to-handler mapping via Gorilla mux
- **Limited Request/Response**: Basic `send()` and `json()` methods only
- **No Path Parameters**: No support for `:id` style parameters
- **No Pattern Matching**: Exact path matching only
- **Simplified Database**: Placeholder functions, no real implementation

### Missing Features
1. **Path Parameter Support**: Cannot handle `/users/:id` patterns
2. **Express.js Request Object**: Missing query, headers, body parsing
3. **Advanced Response Methods**: No status(), cookie(), redirect() support
4. **Pattern Matching**: No reuse of path patterns
5. **Event Loop Integration**: No async operation support
6. **Admin Interface**: No request logging or debugging tools

## Unification Plan

### Phase 1: Extract Core Engine Components

Create a shared engine package that both MCP and REPL can use:

```go
// pkg/jsengine/engine.go
type JSEngine interface {
    RegisterHandler(method, path string, handler goja.Value)
    GetHandler(method, path string) (*HandlerInfo, bool)
    ExecuteScript(code string) (*EvalResult, error)
    SetupBindings()
}

type SharedEngine struct {
    rt           *goja.Runtime
    handlers     map[string]map[string]*HandlerInfo
    // ... other shared components
}
```

### Phase 2: Unified Handler System

Extract the Express.js compatibility layer:

```go
// pkg/jsengine/express.go
type ExpressHandler struct {
    engine *SharedEngine
}

func (h *ExpressHandler) SetupExpressAPI(rt *goja.Runtime) {
    // Setup app.get, app.post, etc.
    // Setup request/response object creation
    // Setup path parameter parsing
}
```

### Phase 3: Configurable Features

Allow different feature sets for different use cases:

```go
type EngineConfig struct {
    EnableEventLoop    bool
    EnableRequestLog   bool
    EnableAdminAPI     bool
    EnableDatabase     bool
    DatabasePath       string
}

func NewEngine(config EngineConfig) *SharedEngine {
    // Create engine with selected features
}
```

### Phase 4: REPL Integration

Update the REPL to use the shared engine:

```go
// cmd/experiments/js-web-server/internal/repl/model.go
func NewModel(startMultiline bool) Model {
    config := jsengine.EngineConfig{
        EnableEventLoop:  false, // Keep simple for REPL
        EnableRequestLog: false,
        EnableAdminAPI:   false,
        EnableDatabase:   true,
        DatabasePath:     ":memory:",
    }
    
    engine := jsengine.NewEngine(config)
    // ... rest of setup
}
```

## Benefits of Unification

1. **Code Reuse**: Eliminate duplication between MCP and REPL versions
2. **Feature Parity**: REPL gains path parameters and advanced routing
3. **Consistency**: Same JavaScript API across both environments
4. **Maintainability**: Single codebase for core functionality
5. **Testing**: Shared test suite for JavaScript engine behavior

## Migration Strategy

### Step 1: Create Shared Package
- Extract core engine to `pkg/jsengine`
- Move Express.js handlers to shared location
- Create configuration system

### Step 2: Update MCP Version
- Refactor to use shared engine
- Maintain all existing functionality
- Ensure backward compatibility

### Step 3: Update REPL Version
- Replace custom implementation with shared engine
- Add missing features (path parameters, etc.)
- Test interactive functionality

### Step 4: Documentation and Examples
- Update documentation to reflect unified approach
- Create examples showing both MCP and REPL usage
- Add migration guide for existing users

## Implementation Differences Summary

| Feature | MCP Version | REPL Version | Unified Target |
|---------|-------------|--------------|----------------|
| Path Parameters | ✅ `/users/:id` | ❌ Exact match only | ✅ Full support |
| Request Object | ✅ Full Express.js API | ❌ Basic only | ✅ Full API |
| Response Methods | ✅ status(), cookie(), etc. | ❌ send(), json() only | ✅ Full API |
| Database Access | ✅ db.query(), db.exec() | ❌ Placeholder | ✅ Full integration |
| Event Loop | ✅ Async support | ❌ None | 🔧 Configurable |
| Admin Interface | ✅ Request logging | ❌ None | 🔧 Configurable |
| Pattern Matching | ✅ Advanced routing | ❌ None | ✅ Full support |
| Middleware | ✅ app.use() | ❌ None | ✅ Full support |

**Legend:**
- ✅ Implemented
- ❌ Missing
- 🔧 Configurable (optional feature)

## Conclusion

The current implementation shows a clear split between a production-ready MCP version and a simplified REPL version. Unifying these implementations will provide the best of both worlds: the full feature set of the MCP version with the interactive development experience of the REPL version.

The key insight is that path reuse and pattern matching are core features that make the JavaScript web server truly useful for building web applications, not just simple request handlers. The unified approach will ensure both environments support the same powerful routing capabilities.
