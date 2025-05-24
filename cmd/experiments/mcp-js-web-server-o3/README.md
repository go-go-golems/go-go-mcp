# JavaScript Playground Server

A dynamic, JavaScript-powered web server built in Go that allows you to create, modify, and serve web applications entirely through JavaScript code - with built-in SQLite database integration and real-time endpoint registration.

## 🚀 Quick Start

```bash
# Start the server
go run . serve -p 8080

# Execute JavaScript code
go run . execute "registerHandler('GET', '/hello', () => 'Hello World!')"

# Test the server
go run . test
```

Then visit `http://localhost:8080/hello` to see your endpoint in action!

## ✨ Features

- **Dynamic JavaScript Runtime**: Execute JavaScript code that can register HTTP endpoints in real-time
- **SQLite Integration**: Direct database access from JavaScript with automatic parameter binding
- **MIME Type Support**: Serve HTML, JSON, XML, CSS, JavaScript, SVG, and more with proper content types
- **Persistent State**: `globalState` object maintains data across script executions
- **Hot Reloading**: Modify endpoints without server restart
- **Script Isolation**: Safe execution with function scope wrapping
- **Structured Logging**: Comprehensive logging with configurable levels
- **Built-in Examples**: Ready-to-use API endpoints and web pages

## 📖 Documentation

- **[Server Architecture & Internals](docs/server-architecture.md)** - Deep dive into how the server works
- **[JavaScript API Reference](docs/javascript-api.md)** - Complete API documentation with examples

## 🎯 Use Cases

- **Rapid Prototyping**: Build APIs and web interfaces quickly
- **Dynamic Content Management**: Create content that changes based on database queries
- **Educational Projects**: Learn web development with immediate feedback
- **Microservices**: Lightweight services with minimal overhead
- **API Mocking**: Create mock APIs for testing and development

## 📋 Examples

### Simple API Endpoint

```javascript
registerHandler("GET", "/api/users", () => {
    const users = db.query("SELECT * FROM users");
    return { users, count: users.length };
});
```

### Dynamic HTML Page

```javascript
registerHandler("GET", "/dashboard", () => `
<!DOCTYPE html>
<html>
<head><title>Dashboard</title></head>
<body>
    <h1>Server Status</h1>
    <p>Time: ${new Date().toISOString()}</p>
    <p>Requests: ${globalState.requestCount || 0}</p>
</body>
</html>`, "text/html");
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

## 🏗️ Project Structure

```
cmd/experiments/mcp-js-web-server-o3/
├── main.go                    # CLI interface and server bootstrap
├── internal/
│   ├── engine/
│   │   ├── engine.go         # Core JavaScript engine
│   │   ├── dispatcher.go     # Single-threaded job processor
│   │   └── bindings.go       # JavaScript API bindings
│   ├── api/
│   │   └── execute.go        # /v1/execute endpoint
│   └── web/
│       └── router.go         # Dynamic route handling
├── test-scripts/             # Example JavaScript files
├── docs/                     # Documentation
└── README.md
```

## 🚦 Getting Started

### 1. Start the Server

```bash
cd cmd/experiments/mcp-js-web-server-o3
go run . serve
```

### 2. Create Your First Endpoint

```bash
# Create a simple greeting endpoint
go run . execute "
registerHandler('GET', '/greet', (req) => ({
    message: 'Hello, ' + (req.query.name || 'World') + '!',
    timestamp: new Date().toISOString()
}));
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

// Create API endpoint
registerHandler('GET', '/api/users', () => {
    const users = db.query('SELECT * FROM users');
    return { users, total: users.length };
});

console.log('Users API created!');
"
```

Test it: `curl http://localhost:8080/api/users`

### 5. Build a Complete Web Page

```bash
go run . execute "
registerHandler('GET', '/users', () => {
    const users = db.query('SELECT * FROM users');
    
    return \`<!DOCTYPE html>
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
}, 'text/html');

console.log('User directory page created!');
"
```

Visit `http://localhost:8080/users` to see your web page!

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

// Track requests
registerHandler("GET", "/stats", () => ({
    version: globalState.app.version,
    uptime: Math.floor((new Date() - globalState.app.startTime) / 1000),
    requests: ++globalState.app.requestCount
}));
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
RUN go build -o js-playground ./cmd/experiments/mcp-js-web-server-o3

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