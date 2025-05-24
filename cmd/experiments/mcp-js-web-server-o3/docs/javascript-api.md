# JavaScript API Reference

## What is the JavaScript Playground Server?

The JavaScript Playground Server is a revolutionary approach to web development that combines the performance of Go with the flexibility of JavaScript. Instead of writing traditional Go handlers, you write JavaScript code that can:

- **Register HTTP endpoints in real-time** - Create REST APIs, web pages, and file servers on the fly
- **Access SQLite databases directly** - Query and modify data with simple JavaScript functions
- **Maintain persistent state** - Keep data alive across requests and script executions
- **Serve any content type** - HTML pages, JSON APIs, CSS files, SVG images, and more

### The Magic

What makes this special is that **everything happens at runtime**. You can:

```javascript
// Send JavaScript code to a running server
POST /v1/execute
{
  "registerHandler('GET', '/users', () => db.query('SELECT * FROM users'));"
}

// Immediately use the new endpoint
GET /users
// Returns: [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
```

The server runs a **single JavaScript runtime** backed by Go's performance, SQLite's reliability, and JavaScript's expressiveness. It's perfect for:

- **Rapid prototyping** - Build and test APIs in seconds
- **Dynamic content systems** - Pages that adapt based on data
- **Educational projects** - Learn web development with immediate feedback  
- **Microservices** - Lightweight, database-backed services
- **Live demos** - Show functionality that can be modified on the spot

### How It Works

1. **Go provides the foundation** - Fast HTTP server, SQLite integration, logging
2. **JavaScript provides the logic** - Your handlers, business logic, and responses  
3. **Single-threaded execution** - Thread-safe JavaScript with persistent state
4. **Real-time registration** - Add new endpoints without restarting the server

## API Overview

The JavaScript Playground Server provides a rich API for creating dynamic web applications entirely in JavaScript. This API includes HTTP endpoint registration, database access, file serving, and state management.

## Table of Contents

- [Global Objects](#global-objects)
- [HTTP Handler Registration](#http-handler-registration)
- [Database Operations](#database-operations)
- [File Serving](#file-serving)
- [Console Logging](#console-logging)
- [State Management](#state-management)
- [Utility Functions](#utility-functions)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Global Objects

### `globalState`

A persistent object that maintains state across script executions and HTTP requests.

```javascript
// Initialize persistent counters
if (!globalState.visitors) {
    globalState.visitors = 0;
}

// Use in handlers
registerHandler("GET", "/visitors", () => ({
    count: ++globalState.visitors
}));
```

**Properties:**
- Persists across script re-executions
- Shared between all HTTP handlers
- Survives server restarts if properly saved

---

## HTTP Handler Registration

### `registerHandler(method, path, handler [, contentType])`

Registers an HTTP endpoint that will be handled by a JavaScript function.

**Parameters:**
- `method` (string): HTTP method (`GET`, `POST`, `PUT`, `DELETE`, etc.)
- `path` (string): URL path (e.g., `/api/users`, `/health`)
- `handler` (function): JavaScript function to handle requests
- `contentType` (string, optional): MIME type for the response

**Handler Function Signature:**
```javascript
function handler(request) {
    // request object contains: method, url, path, query, headers
    return response; // Can be object, string, or bytes
}
```

### Basic Examples

```javascript
// Simple text response
registerHandler("GET", "/hello", () => "Hello, World!");

// JSON API endpoint
registerHandler("GET", "/api/status", () => ({
    status: "running",
    timestamp: new Date().toISOString()
}));

// Handler with request data
registerHandler("POST", "/api/echo", (req) => ({
    method: req.method,
    path: req.path,
    query: req.query,
    headers: req.headers
}));
```

### Important: Returning Values from Code Execution

When executing JavaScript code via the `/v1/execute` API endpoint or MCP tools, **you must use explicit `return` statements** to capture results:

```javascript
// ❌ This will not capture the result
const data = {name: "test", value: 42};
console.log("Created:", data);
data;  // This won't be returned

// ✅ This will capture the result
const data = {name: "test", value: 42};
console.log("Created:", data);
return data;  // Explicit return needed

// ✅ For simple expressions, wrap in return
return 5 + 3;  // Returns 8

// ✅ For complex logic
const users = db.query("SELECT * FROM users");
console.log("Found", users.length, "users");
return {count: users.length, users: users};
```

**Why?** The execution environment wraps your code in a function, so the last expression is not automatically returned. Use `return` to specify what should be captured as the execution result.

### MIME Type Examples

```javascript
// Explicit JSON response
registerHandler("GET", "/api/data", () => ({
    data: [1, 2, 3]
}), "application/json");

// CSS stylesheet
registerHandler("GET", "/styles.css", () => `
    body { 
        font-family: Arial, sans-serif; 
        margin: 0; 
        padding: 20px; 
    }
    .container { 
        max-width: 800px; 
        margin: 0 auto; 
    }
`, "text/css");

// XML response
registerHandler("GET", "/api/xml", () => `<?xml version="1.0"?>
<response>
    <status>success</status>
    <timestamp>${new Date().toISOString()}</timestamp>
</response>`, "application/xml");

// SVG image
registerHandler("GET", "/logo.svg", () => `
<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg">
    <circle cx="50" cy="50" r="40" stroke="blue" fill="lightblue"/>
    <text x="50" y="55" text-anchor="middle">API</text>
</svg>`, "image/svg+xml");
```

### Dynamic HTML Pages

```javascript
registerHandler("GET", "/admin", () => `
<!DOCTYPE html>
<html>
<head>
    <title>Admin Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; }
        .metric { padding: 10px; margin: 5px; background: #f0f0f0; }
    </style>
</head>
<body>
    <h1>Server Dashboard</h1>
    <div class="metric">
        Server Time: ${new Date().toISOString()}
    </div>
    <div class="metric">
        Active Sessions: ${globalState.sessions || 0}
    </div>
</body>
</html>`, "text/html");
```

---

## Database Operations

### `db.query(sql, ...parameters)`

Executes SQL queries against the SQLite database.

**Parameters:**
- `sql` (string): SQL query with `?` placeholders
- `parameters` (any[]): Parameters to bind to placeholders

**Returns:** Array of objects representing rows

### Query Examples

```javascript
// Simple SELECT
const users = db.query("SELECT * FROM users");
console.log("Found users:", users);

// Parameterized query
const activeUsers = db.query("SELECT * FROM users WHERE active = ?", [true]);

// INSERT with parameters
db.query("INSERT INTO users (name, email) VALUES (?, ?)", ["John Doe", "john@example.com"]);

// Complex query with multiple parameters
const recentUsers = db.query(`
    SELECT u.*, p.title as profile_title 
    FROM users u 
    LEFT JOIN profiles p ON u.id = p.user_id 
    WHERE u.created_at > ? AND u.active = ?
`, [new Date('2024-01-01'), true]);
```

### Database Schema Management

```javascript
// Create tables
db.query(`CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT,
    author_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(id)
)`);

// Add indexes
db.query("CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author_id)");

// Data migration example
const existingUsers = db.query("SELECT COUNT(*) as count FROM users");
if (existingUsers[0].count === 0) {
    db.query("INSERT INTO users (name, email) VALUES (?, ?)", ["Admin", "admin@example.com"]);
    console.log("Default admin user created");
}
```

### Complete CRUD API Example

```javascript
// Create
registerHandler("POST", "/api/users", (req) => {
    const { name, email } = req.query;
    if (!name || !email) {
        return { error: "Name and email required" };
    }
    
    db.query("INSERT INTO users (name, email) VALUES (?, ?)", [name, email]);
    return { success: true, message: "User created" };
});

// Read
registerHandler("GET", "/api/users", () => {
    const users = db.query("SELECT id, name, email, created_at FROM users ORDER BY created_at DESC");
    return { users, count: users.length };
});

// Update
registerHandler("PUT", "/api/users/:id", (req) => {
    const id = req.path.split('/').pop();
    const { name, email } = req.query;
    
    db.query("UPDATE users SET name = ?, email = ? WHERE id = ?", [name, email, id]);
    return { success: true, message: "User updated" };
});

// Delete
registerHandler("DELETE", "/api/users/:id", (req) => {
    const id = req.path.split('/').pop();
    db.query("DELETE FROM users WHERE id = ?", [id]);
    return { success: true, message: "User deleted" };
});
```

---

## File Serving

### `registerFile(path, handler)`

Registers a file handler for serving static or dynamic content.

**Parameters:**
- `path` (string): File path (e.g., `/images/logo.png`)
- `handler` (function): Function that returns file content as bytes or string

### File Serving Examples

```javascript
// Static text file
registerFile("/robots.txt", () => `User-agent: *
Disallow: /admin/
Allow: /api/
`);

// Dynamic CSV export
registerFile("/export/users.csv", () => {
    const users = db.query("SELECT name, email FROM users");
    let csv = "name,email\n";
    users.forEach(user => {
        csv += `"${user.name}","${user.email}"\n`;
    });
    return csv;
});

// Binary content (requires Uint8Array)
registerFile("/images/pixel.png", () => {
    // 1x1 transparent PNG
    return new Uint8Array([
        0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
        0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
        // ... PNG bytes
    ]);
});
```

---

## Console Logging

The console API integrates with the server's structured logging system.

### Available Methods

```javascript
console.log("Info message");       // General information
console.info("Information");       // Informational message  
console.warn("Warning message");   // Warning
console.error("Error occurred");   // Error
console.debug("Debug info");       // Debug information
```

### Logging Examples

```javascript
// Basic logging
console.log("Server started");
console.info("Configuration loaded");
console.warn("Deprecated API used");
console.error("Database connection failed");

// Structured logging with objects
console.log("User action", {
    userId: 123,
    action: "login",
    timestamp: new Date()
});

// Error handling with logging
registerHandler("POST", "/api/data", (req) => {
    try {
        const result = processData(req.body);
        console.log("Data processed successfully", { result });
        return { success: true, result };
    } catch (error) {
        console.error("Data processing failed", { error: error.message });
        return { error: "Processing failed" };
    }
});
```

---

## State Management

### Persistent State with `globalState`

```javascript
// Initialize application state
if (!globalState.app) {
    globalState.app = {
        startTime: new Date(),
        requestCount: 0,
        sessions: new Map(),
        config: {
            maxUsers: 100,
            timeout: 3600000 // 1 hour
        }
    };
}

// Request counter middleware
function trackRequest() {
    globalState.app.requestCount++;
    return globalState.app.requestCount;
}

// Session management
registerHandler("POST", "/api/login", (req) => {
    const sessionId = Math.random().toString(36).substr(2, 9);
    globalState.app.sessions.set(sessionId, {
        userId: req.query.userId,
        loginTime: new Date()
    });
    return { sessionId };
});

// Metrics endpoint
registerHandler("GET", "/metrics", () => ({
    uptime: Math.floor((new Date() - globalState.app.startTime) / 1000),
    requests: globalState.app.requestCount,
    activeSessions: globalState.app.sessions.size
}));
```

### Cache Implementation

```javascript
// Simple cache with TTL
if (!globalState.cache) {
    globalState.cache = new Map();
}

function setCache(key, value, ttlSeconds = 300) {
    globalState.cache.set(key, {
        value,
        expires: Date.now() + (ttlSeconds * 1000)
    });
}

function getCache(key) {
    const item = globalState.cache.get(key);
    if (!item) return null;
    
    if (Date.now() > item.expires) {
        globalState.cache.delete(key);
        return null;
    }
    
    return item.value;
}

// Cached database query
registerHandler("GET", "/api/expensive-query", () => {
    const cacheKey = "expensive-query";
    let result = getCache(cacheKey);
    
    if (!result) {
        console.log("Cache miss, executing query");
        result = db.query("SELECT COUNT(*) as total FROM large_table")[0];
        setCache(cacheKey, result, 60); // Cache for 1 minute
    }
    
    return result;
});
```

---

## Utility Functions

### `JSON.stringify(object)` and `JSON.parse(string)`

```javascript
// JSON utilities
const data = { name: "John", age: 30 };
const jsonString = JSON.stringify(data);
const parsed = JSON.parse(jsonString);

// API endpoint with JSON handling
registerHandler("POST", "/api/process", (req) => {
    try {
        const data = JSON.parse(req.body || "{}");
        const result = processData(data);
        return JSON.stringify(result);
    } catch (error) {
        return JSON.stringify({ error: "Invalid JSON" });
    }
});
```

### Custom Utility Functions

```javascript
// Add to globalThis for reusability
globalThis.formatDate = (date) => {
    return new Date(date).toISOString().split('T')[0];
};

globalThis.generateId = () => {
    return Math.random().toString(36).substr(2, 9);
};

globalThis.validateEmail = (email) => {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
};

// Use in handlers
registerHandler("POST", "/api/users", (req) => {
    const { name, email } = req.query;
    
    if (!validateEmail(email)) {
        return { error: "Invalid email format" };
    }
    
    const id = generateId();
    const created = formatDate(new Date());
    
    db.query("INSERT INTO users (id, name, email, created_date) VALUES (?, ?, ?, ?)", 
             [id, name, email, created]);
    
    return { success: true, id };
});
```

---

## Best Practices

### 1. Error Handling

```javascript
// Always wrap database operations in try-catch
registerHandler("GET", "/api/users/:id", (req) => {
    try {
        const id = req.path.split('/').pop();
        const users = db.query("SELECT * FROM users WHERE id = ?", [id]);
        
        if (users.length === 0) {
            return { error: "User not found", status: 404 };
        }
        
        return { user: users[0] };
    } catch (error) {
        console.error("Database error:", error);
        return { error: "Internal server error", status: 500 };
    }
});
```

### 2. Input Validation

```javascript
function validateUser(userData) {
    const errors = [];
    
    if (!userData.name || userData.name.length < 2) {
        errors.push("Name must be at least 2 characters");
    }
    
    if (!validateEmail(userData.email)) {
        errors.push("Invalid email format");
    }
    
    return errors;
}

registerHandler("POST", "/api/users", (req) => {
    const userData = {
        name: req.query.name,
        email: req.query.email
    };
    
    const errors = validateUser(userData);
    if (errors.length > 0) {
        return { error: "Validation failed", details: errors };
    }
    
    // Proceed with creation...
});
```

### 3. State Management

```javascript
// Initialize state properly
function initializeState() {
    if (!globalState.counters) {
        globalState.counters = {};
    }
    if (!globalState.config) {
        globalState.config = {
            apiVersion: "1.0.0",
            maxRequestsPerMinute: 60
        };
    }
}

initializeState();
```

### 4. Logging and Monitoring

```javascript
// Log important events
registerHandler("POST", "/api/login", (req) => {
    const { username } = req.query;
    
    console.info("Login attempt", { username, ip: req.headers['x-forwarded-for'] });
    
    // Authentication logic...
    
    if (authenticated) {
        console.info("Login successful", { username });
        return { success: true, token: generateToken() };
    } else {
        console.warn("Login failed", { username, reason: "invalid credentials" });
        return { error: "Invalid credentials" };
    }
});
```

---

## Examples

### Complete Blog API

```javascript
// Initialize blog state
if (!globalState.blog) {
    globalState.blog = {
        nextId: 1,
        posts: []
    };
    
    // Create posts table
    db.query(`CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        author TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`);
}

// List all posts
registerHandler("GET", "/api/posts", () => {
    const posts = db.query("SELECT * FROM posts ORDER BY created_at DESC");
    return { posts, count: posts.length };
});

// Get single post
registerHandler("GET", "/api/posts/:id", (req) => {
    const id = req.path.split('/').pop();
    const posts = db.query("SELECT * FROM posts WHERE id = ?", [id]);
    
    if (posts.length === 0) {
        return { error: "Post not found" };
    }
    
    return { post: posts[0] };
});

// Create new post
registerHandler("POST", "/api/posts", (req) => {
    const { title, content, author } = req.query;
    
    if (!title || !content || !author) {
        return { error: "Title, content, and author are required" };
    }
    
    db.query("INSERT INTO posts (title, content, author) VALUES (?, ?, ?)", 
             [title, content, author]);
    
    return { success: true, message: "Post created" };
});

// Blog homepage
registerHandler("GET", "/blog", () => {
    const posts = db.query("SELECT id, title, author, created_at FROM posts ORDER BY created_at DESC LIMIT 10");
    
    let html = `<!DOCTYPE html>
<html>
<head><title>My Blog</title></head>
<body>
    <h1>Latest Posts</h1>`;
    
    posts.forEach(post => {
        html += `
        <div style="margin: 20px 0; padding: 15px; border: 1px solid #ddd;">
            <h2><a href="/blog/posts/${post.id}">${post.title}</a></h2>
            <p>By ${post.author} on ${post.created_at}</p>
        </div>`;
    });
    
    html += `</body></html>`;
    return html;
}, "text/html");
```

### Real-time Metrics Dashboard

```javascript
// Initialize metrics
if (!globalState.metrics) {
    globalState.metrics = {
        requests: 0,
        errors: 0,
        startTime: new Date(),
        endpoints: {}
    };
}

// Middleware to track requests
function trackMetrics(endpoint, success = true) {
    globalState.metrics.requests++;
    if (!success) globalState.metrics.errors++;
    
    if (!globalState.metrics.endpoints[endpoint]) {
        globalState.metrics.endpoints[endpoint] = { calls: 0, errors: 0 };
    }
    globalState.metrics.endpoints[endpoint].calls++;
    if (!success) globalState.metrics.endpoints[endpoint].errors++;
}

// Metrics API
registerHandler("GET", "/api/metrics", () => {
    trackMetrics("/api/metrics");
    
    return {
        uptime: Math.floor((new Date() - globalState.metrics.startTime) / 1000),
        totalRequests: globalState.metrics.requests,
        totalErrors: globalState.metrics.errors,
        errorRate: globalState.metrics.requests > 0 ? 
                   (globalState.metrics.errors / globalState.metrics.requests * 100).toFixed(2) + '%' : '0%',
        endpoints: globalState.metrics.endpoints
    };
});

// Metrics dashboard
registerHandler("GET", "/dashboard", () => `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics Dashboard</title>
    <script>
        function updateMetrics() {
            fetch('/api/metrics')
                .then(r => r.json())
                .then(data => {
                    document.getElementById('metrics').innerHTML = 
                        '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
                });
        }
        setInterval(updateMetrics, 5000);
        updateMetrics();
    </script>
</head>
<body>
    <h1>Real-time Metrics</h1>
    <div id="metrics">Loading...</div>
</body>
</html>`, "text/html");
```

### File Upload Handler

```javascript
registerHandler("POST", "/api/upload", (req) => {
    try {
        // In a real implementation, you'd parse multipart form data
        const { filename, content } = req.query;
        
        if (!filename || !content) {
            return { error: "Filename and content required" };
        }
        
        // Store file metadata in database
        db.query("INSERT INTO files (filename, size, uploaded_at) VALUES (?, ?, ?)", 
                 [filename, content.length, new Date().toISOString()]);
        
        // In production, you'd save to disk or cloud storage
        if (!globalState.uploads) globalState.uploads = {};
        globalState.uploads[filename] = content;
        
        console.info("File uploaded", { filename, size: content.length });
        
        return { 
            success: true, 
            filename, 
            size: content.length,
            downloadUrl: `/api/download/${filename}`
        };
    } catch (error) {
        console.error("Upload failed", error);
        return { error: "Upload failed" };
    }
});

registerHandler("GET", "/api/download/:filename", (req) => {
    const filename = req.path.split('/').pop();
    
    if (!globalState.uploads || !globalState.uploads[filename]) {
        return { error: "File not found" };
    }
    
    return globalState.uploads[filename];
});
```

This JavaScript API provides a powerful foundation for building dynamic web applications with database integration, proper HTTP handling, and state management - all running on a fast, lightweight Go server.