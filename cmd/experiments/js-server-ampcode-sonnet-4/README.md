# JavaScript Web Server with Embedded Runtime

A Go web server that embeds a JavaScript runtime (Goja) allowing dynamic REST endpoint registration and file serving through JavaScript code execution. This experimental implementation provides a powerful platform for building dynamic web APIs with persistent state management.

## Features

- **Embedded JavaScript Runtime**: Execute JavaScript code within a sandboxed Goja environment
- **Dynamic Route Registration**: Register REST endpoints on-the-fly using JavaScript
- **Dynamic File Generation**: Create files and content generators using JavaScript functions
- **Global State Management**: Persistent state across requests with database backing
- **SQLite Integration**: Full database access from JavaScript with prepared statements and transactions
- **Code Archiving**: Automatic archiving of executed JavaScript code with timestamps
- **Hot Loading**: Load JavaScript files from directories on startup
- **Comprehensive API**: Rich set of utilities including HTTP client, UUID generation, logging
- **CLI Management**: Full command-line interface for server and API management

## Quick Start

### Running the Server

```bash
# Start server on default port 8080
go run ./cmd/experiments/js-server-ampcode-sonnet-4 serve

# Start with custom configuration
go run ./cmd/experiments/js-server-ampcode-sonnet-4 serve \
  --port 3000 \
  --db my-app.db \
  --archive my-code-archive \
  --load-dir examples/js-web-server

# Start with automatic JavaScript file loading
go run ./cmd/experiments/js-server-ampcode-sonnet-4 serve \
  --load-dir examples/js-web-server
```

### Server Options

- `--port` - Server port (default: 8080)
- `--db` - SQLite database path (default: jsserver.db)
- `--archive` - Directory for code archives (default: code-archive)
- `--static` - Static files directory (default: static)  
- `--load-dir` - Directory to load JavaScript files from on startup

## JavaScript API

The embedded JavaScript runtime provides access to several powerful APIs:

### Database Operations (`db`)

```javascript
// Execute SQL with parameters
const result = db.exec('INSERT INTO users (name, email) VALUES (?, ?)', 'John', 'john@example.com');
console.log('Inserted ID:', result.lastInsertId);

// Query with results
const users = db.query('SELECT * FROM users WHERE active = ?', true);

// Prepared statements
const stmt = db.prepare('SELECT * FROM users WHERE id = ?');
const user = stmt.exec(123);
stmt.close();

// Transactions
db.transaction((tx) => {
    tx.exec('INSERT INTO users (name) VALUES (?)', 'Alice');
    tx.exec('INSERT INTO users (name) VALUES (?)', 'Bob');
});
```

### Server Registration (`server`)

```javascript
// Register GET endpoint
server.get('/api/users', (req, res) => {
    const users = db.query('SELECT * FROM users');
    res.json({ users: users });
});

// Register POST endpoint with validation
server.post('/api/users', (req, res) => {
    const { name, email } = req.body;
    
    if (!name || !email) {
        return res.status(400).json({ error: 'Name and email required' });
    }
    
    const result = db.exec('INSERT INTO users (name, email) VALUES (?, ?)', name, email);
    res.json({ id: result.lastInsertId, name, email });
});

// Register dynamic file generator
server.file('/reports/users.csv', (req) => {
    const users = db.query('SELECT * FROM users');
    let csv = 'Name,Email\n';
    users.forEach(user => {
        csv += `"${user.name}","${user.email}"\n`;
    });
    return csv;
}, 'text/csv');

// Register static content
server.static('/robots.txt', 'User-agent: *\nDisallow:', 'text/plain');
```

### Global State Management (`state`)

```javascript
// Set and get state
state.set('counter', 0);
const counter = state.get('counter');

// Increment counter
state.set('counter', counter + 1);

// Delete state
state.delete('old_key');

// Clear all state
state.clear();

// List all keys
const keys = state.keys();
```

### Utilities (`utils`)

```javascript
// Logging
utils.log('Hello from JavaScript!');

// Sleep/delay
utils.sleep(1000); // 1 second

// HTTP client
const response = utils.fetch('https://api.example.com/data', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: { key: 'value' }
});

// UUID generation
const id = utils.uuid();

// Current timestamp
const now = utils.now();

// Environment variables
const nodeEnv = utils.env('NODE_ENV');
```

### Request/Response Objects

```javascript
server.get('/api/test', (req, res) => {
    // Request object
    console.log('Method:', req.method);
    console.log('URL:', req.url);
    console.log('Headers:', req.headers);
    console.log('Body:', req.body);
    console.log('Query params:', req.query);
    console.log('Path params:', req.params);
    
    // Response methods
    res.json({ message: 'Hello' });           // JSON response
    res.text('Hello World');                  // Text response
    res.html('<h1>Hello</h1>');              // HTML response
    res.status(201).json({ created: true }); // Set status code
    res.header('X-Custom', 'value');         // Set headers
});
```

## CLI Commands

### API Management

```bash
# Execute JavaScript code from file
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api execute script.js --persist --name "my-script"

# Execute from stdin
echo 'server.get("/hello", (req, res) => res.json({msg: "hi"}))' | \
  go run ./cmd/experiments/js-server-ampcode-sonnet-4 api execute --persist

# List registered routes
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api routes list

# Delete a route
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api routes delete /api/users

# List registered files
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api files list

# Delete a file
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api files delete /reports/users.csv
```

### State Management

```bash
# Get all state
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api state get

# Get specific key
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api state get counter

# Set state value
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api state set counter 42
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api state set config '{"debug": true}'

# Delete state key
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api state delete old_key
```

### Archive Management

```bash
# List archived code files
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api archive list

# Get archived file content
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api archive get 2025-05-24T10-30-15Z-abc123.js
```

### Directory Loading

```bash
# Load all JavaScript files from a directory
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api load-dir ./my-scripts

# Get server status
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api status
```

### Custom Server URL

```bash
# Use different server URL
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api --server http://localhost:3000 routes list
```

## Example Applications

The `examples/js-web-server/` directory contains ready-to-use applications:

### Basic API (`basic-api.js`)
Complete REST API for user management with SQLite backend:
- `GET /api/users` - List users
- `POST /api/users` - Create user  
- `GET /api/users/:id` - Get user by ID
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

### File Server (`file-server.js`)
Dynamic file generation and static content:
- `/reports/users.csv` - Dynamic CSV report
- `/config.json` - Dynamic JSON configuration
- `/dashboard.html` - Interactive HTML dashboard
- `/robots.txt` - Static content

### Stateful App (`stateful-app.js`)
Demonstrates global state and session management:
- `/counter` - Simple counter with persistence
- `/stats/visitors` - Visitor tracking
- `/session/*` - Session management endpoints
- `/status` - Real-time server status
- `/state/*` - State import/export APIs

## Testing the Examples

1. Start the server with examples:
```bash
go run ./cmd/experiments/js-server-ampcode-sonnet-4 serve --load-dir examples/js-web-server
```

2. Test the endpoints:
```bash
# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'

# Get dashboard
curl http://localhost:8080/dashboard.html

# Increment counter
curl -X POST http://localhost:8080/counter/increment

# Get server status
curl http://localhost:8080/status
```

## Architecture

### Components
- **Go HTTP Server**: Handles HTTP requests and routing
- **Goja JavaScript Runtime**: Executes JavaScript code safely
- **SQLite Database**: Stores state, routes, and execution history
- **Code Archive**: Filesystem storage for executed code
- **Route Registry**: Dynamic route management
- **File Registry**: Dynamic file generation

### Security Features
- Sandboxed JavaScript execution
- Limited filesystem access
- Prepared SQL statements
- Input validation and sanitization
- Timeout protection for code execution

### Database Schema
- `global_state` - Key-value state storage
- `executions` - Code execution history
- `routes` - Registered route tracking  
- `files` - Registered file tracking

## Development

### Building
```bash
go build ./cmd/experiments/js-server-ampcode-sonnet-4
```

### Testing
```bash
# Test help system
go run ./cmd/experiments/js-server-ampcode-sonnet-4 --help
go run ./cmd/experiments/js-server-ampcode-sonnet-4 serve --help
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api --help

# Test server startup
go run ./cmd/experiments/js-server-ampcode-sonnet-4 serve --port 8081

# Test API commands (with server running)
go run ./cmd/experiments/js-server-ampcode-sonnet-4 api routes list
```

## Use Cases

- **Rapid Prototyping**: Quickly build REST APIs without compilation
- **Dynamic Microservices**: Services that can modify themselves at runtime
- **Administrative Dashboards**: Self-modifying admin interfaces
- **Configuration APIs**: APIs that can reconfigure themselves
- **Learning Platform**: Teach web development with immediate feedback
- **Integration Testing**: Dynamic test endpoint generation
- **Content Management**: Dynamic content generation and serving

## Limitations

- Single-threaded JavaScript execution
- No module system (ES6 imports/exports)
- Limited filesystem access
- Memory usage grows with state size
- JavaScript execution timeout (30 seconds)

## Future Enhancements

- WebSocket support
- JavaScript module system
- Plugin architecture
- Clustering support
- Metrics and monitoring
- Authentication/authorization
- Rate limiting
- Content compression