# JavaScript Playground Server

A dynamic, JavaScript-powered web server built in Go that provides an Express.js compatible API for creating web applications entirely through JavaScript code - with built-in SQLite database integration and real-time endpoint registration.

## 🚀 Quick Start

```bash
# Start the server
go run . serve -p 8080

# Execute JavaScript code (Express.js style)
go run . execute "app.get('/hello', (req, res) => res.send('Hello World!'))"

# Start interactive REPL for experimentation
go run . repl

# Test the server
go run . test
```

Then visit `http://localhost:8080/hello` to see your endpoint in action!

## ✨ Features

- **Express.js Compatible API**: Use familiar Express.js syntax (`app.get`, `app.post`, `req`, `res`)
- **Interactive REPL**: JavaScript Read-Eval-Print Loop for quick experimentation and debugging
- **Dynamic JavaScript Runtime**: Execute JavaScript code that can register HTTP endpoints in real-time
- **SQLite Integration**: Direct database access from JavaScript with automatic parameter binding
- **Express.js Response Methods**: `res.send()`, `res.json()`, `res.status()`, `res.redirect()`, etc.
- **Persistent State**: `globalState` object maintains data across script executions
- **Hot Reloading**: Modify endpoints without server restart
- **Script Isolation**: Safe execution with function scope wrapping
- **Structured Logging**: Comprehensive logging with configurable levels
- **Legacy Support**: Backward compatible with custom `registerHandler` API

## 📖 Documentation

- **[JavaScript Developer Guide](pkg/doc/docs/javascript-developer-guide.md)** - Complete guide to building applications in the sandboxed environment
- **[Server Architecture & Internals](pkg/doc/docs/server-architecture.md)** - Deep dive into how the server works internally
- **[Express.js API Reference](#expressjs-api)** - Familiar Express.js compatible API for web development

### Quick Reference

The server provides a complete Express.js compatible development environment:

```javascript
// Express.js style routing
app.get('/users/:id', (req, res) => {
    const user = db.query('SELECT * FROM users WHERE id = ?', [req.params.id])[0];
    if (!user) return res.status(404).json({ error: 'User not found' });
    res.json(user);
});

// Database integration
app.post('/users', (req, res) => {
    const { name, email } = req.body;
    db.query('INSERT INTO users (name, email) VALUES (?, ?)', [name, email]);
    res.status(201).json({ message: 'User created' });
});

// Global state management
if (!globalState.appConfig) {
    globalState.appConfig = { version: '1.0.0', requestCount: 0 };
}
```

## 📋 Examples

### Simple API Endpoint (Express.js style)

```javascript
app.get("/api/users", (req, res) => {
    const users = db.query("SELECT * FROM users");
    res.json({ users, count: users.length });
});
```

### Dynamic HTML Page

```javascript
app.get("/dashboard", (req, res) => {
    const html = `
<!DOCTYPE html>
<html>
<head><title>Dashboard</title></head>
<body>
    <h1>Server Status</h1>
    <p>Time: ${new Date().toISOString()}</p>
    <p>Requests: ${globalState.requestCount || 0}</p>
</body>
</html>`;
    res.send(html);
});
```

### Route Parameters

```javascript
app.get("/users/:id", (req, res) => {
    const userId = req.params.id;
    const user = db.query("SELECT * FROM users WHERE id = ?", [userId])[0];
    
    if (!user) {
        return res.status(404).json({ error: "User not found" });
    }
    
    res.json(user);
});
```

### POST Endpoint with JSON Body

```javascript
app.post("/api/users", (req, res) => {
    const { name, email } = req.body;
    
    if (!name || !email) {
        return res.status(400).json({ error: "Name and email required" });
    }
    
    db.query("INSERT INTO users (name, email) VALUES (?, ?)", [name, email]);
    res.status(201).json({ message: "User created successfully" });
});
```

### Database Operations

```javascript
// Create table
db.query(`CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY,
    title TEXT,
    content TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`);

// Insert data
db.query("INSERT INTO posts (title, content) VALUES (?, ?)", 
         ["Hello World", "This is my first post"]);

// Query data
const posts = db.query("SELECT * FROM posts ORDER BY created_at DESC");
```

## 🛠️ CLI Commands

### Server Commands

```bash
# Start server with custom configuration
go run . serve --port 8080 --db data.sqlite --log-level info

# Load JavaScript files on startup
go run . serve --scripts ./my-scripts/

# Production mode
go run . serve --port 80 --log-level warn --db /data/production.sqlite
```

### Client Commands

```bash
# Execute JavaScript from file
go run . execute script.js

# Execute JavaScript from command line
go run . execute "console.log('Hello from CLI')"

# Test server endpoints
go run . test --url http://localhost:8080
```

### Interactive REPL

```bash
# Start basic REPL
go run . repl

# Start in multiline mode
go run . repl --multiline

# Show REPL help
go run . repl --help
```

The REPL provides an interactive JavaScript environment with:
- **Real-time execution**: Test JavaScript expressions immediately
- **Multiline support**: Use Ctrl+J for multi-line input or start with `--multiline`
- **History navigation**: Use arrow keys (↑/↓) to navigate through command history
- **External editor**: Press Ctrl+E or use `/edit` to open code in your preferred editor
- **Built-in commands**: `/help`, `/clear`, `/multiline`, `/edit`, `/quit`
- **Error recovery**: Syntax and runtime errors don't crash the session
- **Console.log support**: Debug output directly in the REPL

#### REPL Demo Videos

Visual demonstrations of REPL features are available in the `demos/` directory:
- **Basic Usage**: Simple expressions and console.log
- **Multiline Mode**: Function definitions and complex code blocks  
- **Slash Commands**: Built-in REPL commands and help system
- **Error Handling**: How the REPL handles various error conditions
- **History Navigation**: Arrow key navigation through command history
- **External Editor**: Integration with external editors via Ctrl+E

Generate demo GIFs: `cd demos && ./generate-all.sh` (requires [VHS](https://github.com/charmbracelet/vhs))

## 🏗️ Project Structure

```
cmd/experiments/js-web-server/
├── main.go                           # CLI interface and server bootstrap
├── internal/
│   ├── engine/
│   │   ├── engine.go                # Core JavaScript runtime (Goja)
│   │   ├── dispatcher.go            # Single-threaded job processor
│   │   ├── bindings.go              # JavaScript API bindings
│   │   └── handlers.go              # Express.js compatible routing
│   ├── repl/                        # Interactive REPL implementation
│   │   ├── model.go                 # REPL UI model with Bubble Tea
│   │   └── styles.go                # Visual styling with Lipgloss
│   ├── api/
│   │   └── execute.go               # /v1/execute endpoint for code execution
│   ├── web/
│   │   └── router.go                # Dynamic route handling
│   └── mcp/
│       └── server.go                # MCP server integration
├── demos/                           # VHS demo tapes for REPL
│   ├── README.md                    # Demo documentation
│   ├── generate-all.sh              # Script to generate all demos
│   └── *.tape                       # VHS tape files
├── pkg/
│   └── doc/                         # Documentation package
│       ├── docs/                    # Embedded documentation files
│       └── doc.go                   # Documentation access functions
├── test-scripts/                    # Example JavaScript applications
└── scripts/                         # Runtime JavaScript storage
```

## 🚦 Getting Started

### 1. Start the Server

```bash
cd cmd/experiments/js-web-server
go run . serve
```

### 2. Create Your First Endpoint

```bash
# Create a simple greeting endpoint (Express.js style)
go run . execute "
app.get('/greet', (req, res) => {
    const name = req.query.name || 'World';
    res.json({
        message: 'Hello, ' + name + '!',
        timestamp: new Date().toISOString()
    });
});
console.log('Greeting endpoint created!');
"
```

### 3. Test Your Endpoint

Visit `http://localhost:8080/greet?name=Alice` or use curl:

```bash
curl "http://localhost:8080/greet?name=Alice"
# {"message":"Hello, Alice!","timestamp":"2024-01-15T10:30:00.000Z"}
```

### 4. Create a Database-Driven API

```bash
go run . execute "
// Create users table
db.query(\`CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    name TEXT,
    email TEXT
)\`);

// Add sample data
db.query('INSERT OR IGNORE INTO users (name, email) VALUES (?, ?)', ['Alice', 'alice@example.com']);
db.query('INSERT OR IGNORE INTO users (name, email) VALUES (?, ?)', ['Bob', 'bob@example.com']);

// Create API endpoint (Express.js style)
app.get('/api/users', (req, res) => {
    const users = db.query('SELECT * FROM users');
    res.json({ users, total: users.length });
});

console.log('Users API created!');
"
```

Test it: `curl http://localhost:8080/api/users`

### 5. Build a Complete Web Page

```bash
go run . execute "
app.get('/users', (req, res) => {
    const users = db.query('SELECT * FROM users');
    
    const html = \`<!DOCTYPE html>
<html>
<head>
    <title>User Directory</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .user { background: #f5f5f5; margin: 10px 0; padding: 15px; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>User Directory</h1>
    \${users.map(user => \`
        <div class=\"user\">
            <strong>\${user.name}</strong><br>
            <a href=\"mailto:\${user.email}\">\${user.email}</a>
        </div>
    \`).join('')}
</body>
</html>\`;
    
    res.send(html);
});

console.log('User directory page created!');
"
```

Visit `http://localhost:8080/users` to see your web page!

## 🚀 Express.js API

### Core Routing Methods

```javascript
// HTTP method routing
app.get('/path', (req, res) => { /* GET handler */ });
app.post('/path', (req, res) => { /* POST handler */ });
app.put('/path', (req, res) => { /* PUT handler */ });
app.delete('/path', (req, res) => { /* DELETE handler */ });
app.patch('/path', (req, res) => { /* PATCH handler */ });

// Route parameters
app.get('/users/:id/posts/:postId', (req, res) => {
    const { id, postId } = req.params;
    res.json({ userId: id, postId });
});
```

### Request Object (`req`)

```javascript
app.get('/info', (req, res) => {
    res.json({
        method: req.method,        // HTTP method
        path: req.path,            // URL path
        query: req.query,          // Query parameters
        params: req.params,        // Route parameters
        headers: req.headers,      // HTTP headers
        body: req.body,            // Request body (auto-parsed JSON)
        cookies: req.cookies,      // Cookies
        ip: req.ip                 // Client IP
    });
});
```

### Response Object (`res`)

```javascript
app.get('/response-examples', (req, res) => {
    // JSON response
    res.json({ message: 'Hello' });
    
    // Status codes
    res.status(404).json({ error: 'Not found' });
    
    // HTML response
    res.send('<h1>Hello World</h1>');
    
    // Redirects
    res.redirect('/new-location');
    
    // Headers
    res.set('X-Custom-Header', 'value');
    
    // Cookies
    res.cookie('sessionId', 'abc123', { maxAge: 3600000 });
});
```

### Database Integration

```javascript
// Create tables
db.query(`CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL
)`);

// Insert data
app.post('/users', (req, res) => {
    const { name, email } = req.body;
    db.query('INSERT INTO users (name, email) VALUES (?, ?)', [name, email]);
    res.status(201).json({ message: 'User created' });
});

// Query data
app.get('/users', (req, res) => {
    const users = db.query('SELECT * FROM users ORDER BY created_at DESC');
    res.json({ users, count: users.length });
});
```

### Global State

```javascript
// Initialize application state
if (!globalState.app) {
    globalState.app = {
        version: '1.0.0',
        startTime: new Date(),
        requestCount: 0
    };
}

// Use persistent state
app.get('/stats', (req, res) => {
    globalState.app.requestCount++;
    res.json({
        version: globalState.app.version,
        uptime: new Date() - globalState.app.startTime,
        requests: globalState.app.requestCount
    });
});
```

## 🔧 Advanced Features

### Load Scripts on Startup

Create JavaScript files in a directory and load them when the server starts:

```bash
# Create script directory
mkdir my-api
echo "registerHandler('GET', '/status', () => ({status: 'running'}));" > my-api/status.js

# Start server with scripts
go run . serve --scripts my-api/
```

### Persistent State Management

```javascript
// Initialize application state
if (!globalState.app) {
    globalState.app = {
        version: "1.0.0",
        startTime: new Date(),
        requestCount: 0
    };
}

// Track requests (Express.js style)
app.get("/stats", (req, res) => {
    res.json({
        version: globalState.app.version,
        uptime: Math.floor((new Date() - globalState.app.startTime) / 1000),
        requests: ++globalState.app.requestCount
    });
});
```

### File Serving

```javascript
// Serve CSS files
registerFile("/styles.css", () => `
    body { background: #f0f0f0; font-family: Arial; }
    .container { max-width: 800px; margin: 0 auto; }
`);

// Serve dynamic content
registerFile("/data.json", () => {
    const data = db.query("SELECT * FROM metrics");
    return JSON.stringify(data);
});
```

## 🔍 Monitoring and Debugging

### Built-in Endpoints

The server includes several built-in endpoints:

- `GET /health` - Health check
- `GET /` - Welcome message  
- `POST /counter` - Request counter

### Logging

Configure logging levels for development and production:

```bash
# Development - see everything
go run . serve --log-level debug

# Production - errors and warnings only
go run . serve --log-level warn
```

### JavaScript Console

Use console methods in your JavaScript code:

```javascript
console.log("Info message");
console.warn("Warning message"); 
console.error("Error message");
console.debug("Debug information");
```

## 🚀 Deployment

### Docker

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o js-playground ./cmd/experiments/js-web-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/js-playground .
EXPOSE 8080
CMD ["./js-playground", "serve"]
```

### Environment Variables

```bash
export PORT=8080
export DB_PATH=/data/production.sqlite
export LOG_LEVEL=info
```

## 📈 Performance

- **Throughput**: 100-1000 RPS depending on JavaScript complexity
- **Latency**: Sub-millisecond for simple handlers
- **Memory**: Efficient single JavaScript context
- **Concurrency**: Single-threaded JavaScript execution with Go-based HTTP handling

## 🤝 Contributing

This is an experimental project demonstrating the integration of JavaScript runtime with Go web servers. Feel free to explore, modify, and extend the functionality!

## 📄 License

This project is part of the go-go-mcp experimental suite.