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

### Execution Model

JavaScript code executes directly in the **global scope** of a persistent JavaScript runtime. This means:

- **Persistent State**: Variables and functions remain in memory between executions
- **No Function Wrapping**: Your code runs directly, not wrapped in functions
- **Global Scope**: All code shares the same global context
- **Re-execution Safe**: Code should be written to handle multiple executions

**Key Implications:**
- Use `var` or direct assignment instead of `let`/`const` for reloadable code
- Use `globalState` object for application data
- Functions can be redefined safely
- Handlers are replaced when re-registered

## Table of Contents

- [Global Objects](#global-objects)
- [HTTP Handler Registration](#http-handler-registration)
- [Database Operations](#database-operations)
- [File Serving](#file-serving)
- [HTTP Requests](#http-requests)
- [Console Logging](#console-logging)
- [State Management](#state-management)
- [Utility Functions](#utility-functions)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Quick Reference

### Database Operations
- `db.query(sql, ...args)` - Execute SELECT queries, returns array of objects
- `db.exec(sql, ...args)` - Execute INSERT/UPDATE/DELETE/CREATE, returns {success, rowsAffected, lastInsertId}

### HTTP Handlers
- `registerHandler(method, path, handler [, options])` - Register HTTP endpoint with enhanced features
- `registerFile(path, handler)` - Register file endpoint

### HTTP Requests
- `fetch(url [, options])` - Modern fetch API for HTTP requests
- `HTTP.get(url [, options])` - GET request shortcut
- `HTTP.post(url [, options])` - POST request shortcut  
- `HTTP.put(url [, options])` - PUT request shortcut
- `HTTP.delete(url [, options])` - DELETE request shortcut
- `HTTP.patch(url [, options])` - PATCH request shortcut
- `HTTP.head(url [, options])` - HEAD request shortcut

### HTTP Constants & Helpers
- `HTTP.OK`, `HTTP.CREATED`, `HTTP.NOT_FOUND`, `HTTP.INTERNAL_SERVER_ERROR`, etc. - Status code constants
- `Response.json(data [, status])` - Create JSON response
- `Response.text(text [, status])` - Create text response  
- `Response.html(html [, status])` - Create HTML response
- `Response.redirect(url [, status])` - Create redirect response
- `Response.error(message [, status])` - Create error response

### Console
- `console.log(...)`, `console.info(...)`, `console.warn(...)`, `console.error(...)`, `console.debug(...)`

### State & Utils
- `globalState` - Persistent object across executions
- `JSON.stringify(obj)`, `JSON.parse(str)` - JSON utilities

### Enhanced Handler Function
```javascript
function handler(request) {
  // Enhanced request: {method, url, path, query, headers, body, params, cookies, remoteIP}
  return response; // object, string, Uint8Array, or ResponseObject
}
```

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

### `registerHandler(method, path, handler [, options])`

Registers an HTTP endpoint that will be handled by a JavaScript function.

**Parameters:**
- `method` (string): HTTP method (`GET`, `POST`, `PUT`, `DELETE`, etc.)
- `path` (string): URL path (e.g., `/api/users`, `/health`, `/users/:id` for path parameters)
- `handler` (function): JavaScript function to handle requests
- `options` (object|string, optional): Handler options or content type string for backward compatibility

**Handler Function Signature:**
```javascript
function handler(request) {
    // Enhanced request object with rich properties
    return response; // Can be object, string, bytes, or ResponseObject
}
```

## Admin Console

The JavaScript Playground Server includes a powerful admin console for monitoring and debugging your applications.

### `/admin/logs` - Live Request Console

Access the live request console at `http://localhost:8080/admin/logs` to:

**Available in both server modes:**
- **Serve mode**: `./js-server serve` - Admin console at `http://localhost:8080/admin/logs`
- **MCP mode**: `./js-server mcp` - Admin console automatically available at `http://localhost:8080/admin/logs`

- **View all HTTP requests** in real-time with detailed information
- **See console logs** generated during request processing  
- **Monitor performance** with request timing and statistics
- **Debug issues** with full request/response data and error details
- **Filter and search** through request history

**Features:**
- Real-time request monitoring with auto-refresh
- Detailed request information (headers, query parameters, body)
- Console log capture per request (console.log, console.error, etc.)
- Response data and status codes
- Performance metrics and timing
- Error tracking and debugging
- Clean, responsive interface

The admin console automatically captures:
- All JavaScript console output during request processing
- **Database operations** with SQL queries, parameters, results, and timing
- Request timing and performance metrics  
- Full request and response data
- Error details and stack traces

**Example Usage:**
```javascript
// This JavaScript handler logs will appear in the admin console
registerHandler("GET", "/debug", (req) => {
    console.log("Debug endpoint called", { 
        method: req.method, 
        path: req.path,
        userAgent: req.headers["user-agent"]
    });
    
    // Database operations are automatically logged
    const userCount = db.query("SELECT COUNT(*) as count FROM users")[0].count;
    console.warn("Current user count:", userCount);
    
    // This insert operation will also be logged
    db.exec("INSERT INTO access_log (endpoint, timestamp) VALUES (?, ?)", 
            [req.path, new Date().toISOString()]);
    
    console.error("This is an error for testing");
    
    return Response.json({
        message: "Check the admin console for logs and database operations",
        userCount: userCount,
        timestamp: new Date().toISOString()
    });
});
```

When you call this endpoint, all the console messages AND database operations will be captured and displayed in the admin interface, associated with that specific request. You'll see:
- Console logs with timestamps and levels
- Database queries with SQL, parameters, results, and execution time
- Performance metrics and timing information

---

**Request Object Properties:**

**Note:** Use capitalized field names for direct property access in JavaScript:

```javascript
{
    Method: "GET",           // HTTP method (use request.Method)
    URL: "/api/users?page=1", // Full URL with query string (use request.URL)
    Path: "/api/users",      // URL path only (use request.Path)
    Query: {                 // Parsed query parameters (use request.Query)
        page: "1",           // String values for single params
        tags: ["a", "b"]     // Array for multiple values
    },
    Headers: {               // Request headers (use request.Headers)
        "content-type": "application/json",
        "authorization": "Bearer token123"
    },
    Body: "...",            // Request body as string (use request.Body)
    Params: {               // URL path parameters (use request.Params)
        id: "123"           // Automatically extracted from URL path
    },
    Cookies: {              // Request cookies (use request.Cookies)
        session: "abc123"
    },
    RemoteIP: "192.168.1.1" // Client IP address (use request.RemoteIP)
}
```

**Important:** While `JSON.stringify(request)` shows lowercase field names (JSON tags), direct property access requires **capitalized field names** (Go struct fields). Use `request.Params`, `request.Path`, `request.Method`, etc.

**Response Formats:**

**1. Simple Response (backward compatible):**
```javascript
// String response
return "Hello World";

// JSON response  
return { message: "success", data: [...] };

// Raw bytes
return new Uint8Array([...]);
```

**2. Enhanced Response Object:**
```javascript
return {
    status: 200,                    // HTTP status code (optional, defaults to 200)
    headers: {                      // Response headers (optional)
        "x-custom": "value",
        "cache-control": "no-cache"
    },
    body: "Response content",       // Response body (string, object, or bytes)
    contentType: "text/html",       // Content-Type override (optional)
    cookies: [{                     // Response cookies (optional)
        name: "session",
        value: "abc123",
        path: "/",
        domain: "example.com",
        maxAge: 3600,              // Seconds
        secure: true,              // HTTPS only
        httpOnly: true,            // No JavaScript access
        sameSite: "Strict"         // "Strict", "Lax", or "None"
    }],
    redirect: "https://example.com" // Redirect URL (sets 302 status if not specified)
};
```

### Path Parameter Routing

The server supports path parameters using the `:paramName` syntax. Parameters are automatically extracted and available in `request.params`.

```javascript
// Single path parameter
registerHandler("GET", "/users/:id", (request) => {
    const userId = request.Params.id;  // Use request.Params (capitalized)
    console.log("User ID:", userId);
    
    const user = db.query("SELECT * FROM users WHERE id = ?", [userId]);
    if (user.length === 0) {
        return Response.error("User not found", 404);
    }
    
    return Response.json(user[0]);
});

// Multiple path parameters
registerHandler("GET", "/api/:version/users/:userId/posts/:postId", (request) => {
    const { version, userId, postId } = request.Params;  // Use request.Params (capitalized)
    
    return Response.json({
        apiVersion: version,
        userId: userId,
        postId: postId,
        data: "Post content here..."
    });
});

// Trivia game answer endpoint example
registerHandler("GET", "/answer/:answerIndex", (request) => {
    const answerIndex = parseInt(request.Params.answerIndex);  // Use request.Params (capitalized)
    
    if (isNaN(answerIndex) || answerIndex < 0) {
        return Response.error("Invalid answer index", 400);
    }
    
    // Game logic here...
    return Response.json({
        selectedAnswer: answerIndex,
        correct: checkAnswer(answerIndex)
    });
});
```

**Path Parameter Rules:**
- Use `:paramName` syntax in the path pattern
- Parameters are automatically extracted to `request.Params` (capitalized!)
- Parameter names can contain letters, numbers, and underscores
- All path segments must match exactly except parameter segments
- Parameters are always strings - convert with `parseInt()`, etc. as needed

### Basic Examples

```javascript
// Simple text response (backward compatible)
registerHandler("GET", "/hello", () => "Hello, World!");

// JSON API endpoint (backward compatible)
registerHandler("GET", "/api/status", () => ({
    status: "running",
    timestamp: new Date().toISOString()
}));

// Enhanced response with custom status and headers
registerHandler("GET", "/api/info", () => ({
    status: 200,
    headers: {
        "x-api-version": "1.0",
        "cache-control": "max-age=3600"
    },
    body: {
        server: "JavaScript Playground",
        version: "1.0.0",
        timestamp: new Date().toISOString()
    }
}));

// Handler with enhanced request data
registerHandler("POST", "/api/echo", (req) => ({
    body: {
        receivedData: {
            method: req.method,
            path: req.path,
            query: req.query,
            headers: req.headers,
            body: req.body,
            cookies: req.cookies,
            remoteIP: req.remoteIP
        }
    },
    status: 200,
    contentType: "application/json"
}));

// Redirect example
registerHandler("GET", "/old-page", () => ({
    redirect: "/new-page",
    status: 301  // Permanent redirect
}));

// Cookie example
registerHandler("POST", "/login", (req) => {
    const { username, password } = JSON.parse(req.body || "{}");
    
    if (username === "admin" && password === "secret") {
        return {
            status: 200,
            body: { success: true, message: "Login successful" },
            cookies: [{
                name: "session",
                value: "abc123",
                path: "/",
                maxAge: 3600,
                httpOnly: true,
                secure: true
            }]
        };
    }
    
    return Response.error("Invalid credentials", 401);
});
```

### Important: Code Execution and Variable Scope

JavaScript code executes directly in the **global scope** of a persistent JavaScript runtime. This has important implications:

#### Variable Declaration Guidelines

**❌ Avoid `let` and `const` for reloadable code:**
```javascript
// This will fail on re-execution because variables can't be redeclared
let userCount = 0;
const API_VERSION = "1.0";

// Error on second execution: "userCount has already been declared"
```

**✅ Use `var` or direct assignment for global variables:**
```javascript
// These can be safely re-executed multiple times
var userCount = 0;
API_VERSION = "1.0";  // Creates/updates global property

// Or check before declaring
if (typeof userCount === 'undefined') {
    var userCount = 0;
}
```

**✅ Use `globalState` for persistent data:**
```javascript
// Recommended approach for application state
if (!globalState.userCount) {
    globalState.userCount = 0;
}
globalState.userCount++;
```

#### Returning Values from Code Execution

When executing JavaScript code via the `/v1/execute` API endpoint or MCP tools, the **last expression** is automatically returned:

```javascript
// ✅ Simple expressions return their value
5 + 3;  // Returns 8

// ✅ Object expressions
{message: "hello", value: 42};  // Returns the object

// ✅ Function call results
db.query("SELECT COUNT(*) as count FROM users");  // Returns query result

// ✅ Explicit returns also work
const users = db.query("SELECT * FROM users");
console.log("Found", users.length, "users");
return {count: users.length, users: users};
```

#### Best Practices for Reloadable Code

```javascript
// ✅ Safe function definitions (can be redefined)
function processUser(userData) {
    return {
        id: userData.id,
        name: userData.name.toUpperCase()
    };
}

// ✅ Safe handler registration (replaces existing)
registerHandler("GET", "/api/users", () => {
    return db.query("SELECT * FROM users");
});

// ✅ Safe state initialization
globalState.config = globalState.config || {
    maxUsers: 100,
    rateLimit: 1000
};

// ✅ Safe utility setup
if (typeof formatDate === 'undefined') {
    var formatDate = function(date) {
        return new Date(date).toISOString().split('T')[0];
    };
}
```

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

### `db.exec(sql, ...parameters)`

Executes SQL statements that don't return rows (INSERT, UPDATE, DELETE, CREATE, etc.).

**Parameters:**
- `sql` (string): SQL statement with `?` placeholders
- `parameters` (any[]): Parameters to bind to placeholders

**Returns:** Object with:
- `success` (boolean): Whether the operation succeeded
- `rowsAffected` (number): Number of rows affected
- `lastInsertId` (number): Last inserted row ID (for INSERT statements)
- `error` (string): Error message if operation failed

### Exec Examples

```javascript
// Create table
const result = db.exec(`CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE
)`);
console.log("Table created:", result.success);

// Insert data
const insertResult = db.exec("INSERT INTO users (name, email) VALUES (?, ?)", ["Alice", "alice@example.com"]);
console.log("Inserted user ID:", insertResult.lastInsertId);

// Update records
const updateResult = db.exec("UPDATE users SET name = ? WHERE id = ?", ["Alice Smith", 1]);
console.log("Updated rows:", updateResult.rowsAffected);

// Delete records
const deleteResult = db.exec("DELETE FROM users WHERE email LIKE ?", ["%@temp.com"]);
console.log("Deleted rows:", deleteResult.rowsAffected);

// Multiple statements in one exec
db.exec(`
    CREATE TABLE IF NOT EXISTS contacts (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      email TEXT NOT NULL,
      company TEXT,
      phone TEXT
    );
    DELETE FROM contacts;
    INSERT INTO contacts (name, email, company, phone) VALUES
      ('Alice Smith', 'alice@acme.com', 'Acme Corp', '555-1234'),
      ('Bob Johnson', 'bob@globex.com', 'Globex Inc', '555-5678');
`);
```

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
    const { name, email } = req.Query;  // Use req.Query (capitalized)
    if (!name || !email) {
        return Response.error("Name and email required", 400);
    }
    
    const result = db.exec("INSERT INTO users (name, email) VALUES (?, ?)", [name, email]);
    return Response.json({ 
        success: true, 
        message: "User created",
        id: result.lastInsertId
    });
});

// Read
registerHandler("GET", "/api/users", () => {
    const users = db.query("SELECT id, name, email, created_at FROM users ORDER BY created_at DESC");
    return { users, count: users.length };
});

// Read single user
registerHandler("GET", "/api/users/:id", (req) => {
    const id = req.Params.id;  // Use req.Params (capitalized)
    const users = db.query("SELECT * FROM users WHERE id = ?", [id]);
    
    if (users.length === 0) {
        return Response.error("User not found", 404);
    }
    
    return Response.json(users[0]);
});

// Update
registerHandler("PUT", "/api/users/:id", (req) => {
    const id = req.Params.id;  // Use req.Params (capitalized)
    const { name, email } = req.Query;  // Use req.Query (capitalized)
    
    if (!name || !email) {
        return Response.error("Name and email required", 400);
    }
    
    const result = db.exec("UPDATE users SET name = ?, email = ? WHERE id = ?", [name, email, id]);
    if (result.rowsAffected === 0) {
        return Response.error("User not found", 404);
    }
    
    return Response.json({ success: true, message: "User updated" });
});

// Delete
registerHandler("DELETE", "/api/users/:id", (req) => {
    const id = req.Params.id;  // Use req.Params (capitalized)
    const result = db.exec("DELETE FROM users WHERE id = ?", [id]);
    
    if (result.rowsAffected === 0) {
        return Response.error("User not found", 404);
    }
    
    return Response.json({ success: true, message: "User deleted" });
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

## HTTP Requests

The JavaScript Playground Server provides a comprehensive HTTP client API that allows your JavaScript code to make outbound HTTP requests to other services, APIs, and endpoints.

### `fetch(url [, options])`

Modern fetch API for making HTTP requests, similar to the browser's fetch API.

**Parameters:**
- `url` (string): The URL to request
- `options` (object, optional): Request configuration options

**Options Object:**
```javascript
{
    method: "GET",              // HTTP method (GET, POST, PUT, DELETE, etc.)
    headers: {                  // Request headers
        "Authorization": "Bearer token",
        "Content-Type": "application/json"
    },
    body: "request data",       // Request body (string, object, or array)
    query: {                    // Query parameters to append to URL
        page: 1,
        limit: 10,
        tags: ["javascript", "api"]  // Array values become multiple params
    },
    timeout: 30                 // Request timeout in seconds (default: 30)
}
```

**Returns:** Response object with:
```javascript
{
    status: 200,                    // HTTP status code
    statusText: "OK",               // HTTP status text
    headers: {                      // Response headers
        "content-type": "application/json"
    },
    body: "response text",          // Response body as string
    json: {...},                    // Parsed JSON (if response is JSON)
    ok: true,                       // true if status 200-299
    url: "https://api.example.com", // Final URL (after redirects)
    error: "error message"          // Error message if request failed
}
```

### HTTP Method Shortcuts

Convenient shortcuts for common HTTP methods:

#### `HTTP.get(url [, options])`
#### `HTTP.post(url [, options])`  
#### `HTTP.put(url [, options])`
#### `HTTP.delete(url [, options])`
#### `HTTP.patch(url [, options])`
#### `HTTP.head(url [, options])`

These methods automatically set the HTTP method and accept the same options as `fetch()`.

### Basic Examples

```javascript
// Simple GET request
const response = fetch("https://jsonplaceholder.typicode.com/posts/1");
console.log("Post:", response.json);

// GET with query parameters
const posts = fetch("https://jsonplaceholder.typicode.com/posts", {
    query: {
        userId: 1,
        _limit: 5
    }
});
console.log("User posts:", posts.json);

// POST request with JSON body
const newPost = fetch("https://jsonplaceholder.typicode.com/posts", {
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
    body: {
        title: "My New Post",
        body: "This is the content",
        userId: 1
    }
});
console.log("Created post:", newPost.json);

// Using method shortcuts
const user = HTTP.get("https://jsonplaceholder.typicode.com/users/1");
const deleteResult = HTTP.delete("https://api.example.com/items/123", {
    headers: {
        "Authorization": "Bearer " + apiToken
    }
});
```

### Advanced Examples

```javascript
// API integration with error handling
registerHandler("GET", "/api/external-data", () => {
    const response = fetch("https://api.external-service.com/data", {
        headers: {
            "Authorization": "Bearer " + globalState.apiToken,
            "User-Agent": "MyApp/1.0"
        },
        timeout: 10
    });
    
    if (!response.ok) {
        console.error("External API error:", response.status, response.body);
        return Response.error("External service unavailable", 503);
    }
    
    return Response.json({
        data: response.json,
        cached: false,
        timestamp: new Date().toISOString()
    });
});

// Webhook notification
registerHandler("POST", "/api/notify", (req) => {
    const { message, channel } = JSON.parse(req.body || "{}");
    
    // Send to Slack webhook
    const slackResponse = fetch("https://hooks.slack.com/services/YOUR/WEBHOOK/URL", {
        method: "POST",
        body: {
            text: message,
            channel: channel || "#general"
        }
    });
    
    if (slackResponse.ok) {
        console.log("Notification sent successfully");
        return Response.json({ success: true });
    } else {
        console.error("Slack notification failed:", slackResponse.body);
        return Response.error("Notification failed", 500);
    }
});

// Proxy endpoint with transformation
registerHandler("GET", "/api/weather/:city", (req) => {
    const city = req.params.city;
    const apiKey = globalState.weatherApiKey;
    
    const weather = fetch(`https://api.openweathermap.org/data/2.5/weather`, {
        query: {
            q: city,
            appid: apiKey,
            units: "metric"
        }
    });
    
    if (!weather.ok) {
        return Response.error("Weather data unavailable", 404);
    }
    
    // Transform the response
    return Response.json({
        city: weather.json.name,
        temperature: weather.json.main.temp,
        description: weather.json.weather[0].description,
        humidity: weather.json.main.humidity,
        timestamp: new Date().toISOString()
    });
});

// Multi-service aggregation
registerHandler("GET", "/api/dashboard-data", () => {
    // Fetch data from multiple APIs in parallel
    const userStats = fetch("https://api.internal.com/user-stats");
    const systemMetrics = fetch("https://monitoring.example.com/metrics", {
        headers: { "Authorization": "Bearer " + globalState.monitoringToken }
    });
    const externalData = fetch("https://api.third-party.com/summary");
    
    return Response.json({
        users: userStats.ok ? userStats.json : { error: "unavailable" },
        system: systemMetrics.ok ? systemMetrics.json : { error: "unavailable" },
        external: externalData.ok ? externalData.json : { error: "unavailable" },
        generated: new Date().toISOString()
    });
});
```

### Request Configuration Examples

```javascript
// Custom headers and authentication
const authResponse = fetch("https://api.secure-service.com/data", {
    headers: {
        "Authorization": "Bearer " + globalState.accessToken,
        "X-API-Version": "2.0",
        "Accept": "application/json",
        "User-Agent": "MyServer/1.0"
    }
});

// File upload simulation (form data)
const uploadResponse = fetch("https://upload.example.com/files", {
    method: "POST",
    headers: {
        "Content-Type": "application/octet-stream",
        "X-Filename": "data.json"
    },
    body: JSON.stringify({ data: "file content" })
});

// Query parameter variations
const searchResults = fetch("https://api.search.com/query", {
    query: {
        q: "javascript",
        type: ["code", "docs"],     // Multiple values: ?type=code&type=docs
        limit: 20,
        sort: "relevance"
    }
});

// Timeout and retry logic
function fetchWithRetry(url, options, maxRetries = 3) {
    for (let i = 0; i < maxRetries; i++) {
        const response = fetch(url, { ...options, timeout: 5 });
        if (response.ok) {
            return response;
        }
        console.warn(`Request failed, attempt ${i + 1}/${maxRetries}`);
    }
    throw new Error("Max retries exceeded");
}
```

### Integration with Database

```javascript
// Cache external API responses in database
registerHandler("GET", "/api/cached-data/:id", (req) => {
    const id = req.params.id;
    
    // Check cache first
    const cached = db.query("SELECT * FROM api_cache WHERE id = ? AND expires_at > datetime('now')", [id]);
    if (cached.length > 0) {
        console.log("Returning cached data for", id);
        return Response.json(JSON.parse(cached[0].data));
    }
    
    // Fetch fresh data
    const response = fetch(`https://api.external.com/items/${id}`);
    if (!response.ok) {
        return Response.error("External API error", response.status);
    }
    
    // Cache the response for 1 hour
    const expiresAt = new Date(Date.now() + 3600000).toISOString();
    db.exec("INSERT OR REPLACE INTO api_cache (id, data, expires_at) VALUES (?, ?, ?)", 
            [id, JSON.stringify(response.json), expiresAt]);
    
    return Response.json(response.json);
});

// Sync external data to local database
registerHandler("POST", "/api/sync-users", () => {
    const response = fetch("https://api.hr-system.com/employees", {
        headers: {
            "Authorization": "Bearer " + globalState.hrApiToken
        }
    });
    
    if (!response.ok) {
        return Response.error("Failed to fetch employee data", 500);
    }
    
    let syncCount = 0;
    response.json.employees.forEach(emp => {
        const result = db.exec(`
            INSERT OR REPLACE INTO users (external_id, name, email, department) 
            VALUES (?, ?, ?, ?)
        `, [emp.id, emp.full_name, emp.email, emp.department]);
        
        if (result.success) syncCount++;
    });
    
    console.log(`Synced ${syncCount} employees`);
    return Response.json({ synced: syncCount, total: response.json.employees.length });
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

## HTTP Utilities and Response Helpers

The JavaScript Playground Server provides convenient HTTP constants and response helper functions to make building APIs easier and more expressive.

### HTTP Status Code Constants

```javascript
// Success codes
HTTP.OK                    // 200
HTTP.CREATED               // 201
HTTP.ACCEPTED              // 202
HTTP.NO_CONTENT            // 204

// Redirect codes  
HTTP.MOVED_PERMANENTLY     // 301
HTTP.FOUND                 // 302
HTTP.NOT_MODIFIED          // 304

// Client error codes
HTTP.BAD_REQUEST           // 400
HTTP.UNAUTHORIZED          // 401
HTTP.FORBIDDEN             // 403
HTTP.NOT_FOUND             // 404
HTTP.METHOD_NOT_ALLOWED    // 405
HTTP.CONFLICT              // 409

// Server error codes
HTTP.INTERNAL_SERVER_ERROR // 500
HTTP.NOT_IMPLEMENTED       // 501
HTTP.BAD_GATEWAY          // 502
HTTP.SERVICE_UNAVAILABLE   // 503
```

### Response Helper Functions

#### `Response.json(data [, status])`

Creates a JSON response with proper content type.

```javascript
// Simple JSON response
registerHandler("GET", "/api/users", () => {
    const users = db.query("SELECT * FROM users");
    return Response.json(users);
});

// JSON response with custom status
registerHandler("POST", "/api/users", (req) => {
    const userData = JSON.parse(req.body);
    const result = db.exec("INSERT INTO users (name, email) VALUES (?, ?)", 
                          [userData.name, userData.email]);
    
    return Response.json({
        success: true,
        id: result.lastInsertId
    }, HTTP.CREATED);
});
```

#### `Response.text(text [, status])`

Creates a plain text response.

```javascript
registerHandler("GET", "/health", () => {
    return Response.text("OK");
});

registerHandler("GET", "/robots.txt", () => {
    return Response.text(`User-agent: *
Disallow: /admin/
Allow: /api/`);
});
```

#### `Response.html(html [, status])`

Creates an HTML response with proper content type.

```javascript
registerHandler("GET", "/dashboard", () => {
    const userCount = db.query("SELECT COUNT(*) as count FROM users")[0].count;
    
    return Response.html(`
        <!DOCTYPE html>
        <html>
        <head><title>Dashboard</title></head>
        <body>
            <h1>Dashboard</h1>
            <p>Total Users: ${userCount}</p>
        </body>
        </html>
    `);
});
```

#### `Response.redirect(url [, status])`

Creates a redirect response.

```javascript
// Temporary redirect (302)
registerHandler("GET", "/old-url", () => {
    return Response.redirect("/new-url");
});

// Permanent redirect (301)
registerHandler("GET", "/legacy", () => {
    return Response.redirect("/new-path", HTTP.MOVED_PERMANENTLY);
});

// External redirect
registerHandler("GET", "/external", () => {
    return Response.redirect("https://example.com");
});
```

#### `Response.error(message [, status])`

Creates a standardized error response.

```javascript
registerHandler("GET", "/api/protected", (req) => {
    if (!req.headers.authorization) {
        return Response.error("Authorization required", HTTP.UNAUTHORIZED);
    }
    
    // Process request...
    return Response.json({ data: "secret information" });
});

registerHandler("POST", "/api/users", (req) => {
    try {
        const userData = JSON.parse(req.body);
        if (!userData.email) {
            return Response.error("Email is required", HTTP.BAD_REQUEST);
        }
        
        // Create user...
        return Response.json({ success: true }, HTTP.CREATED);
    } catch (e) {
        return Response.error("Invalid JSON", HTTP.BAD_REQUEST);
    }
});
```

### Advanced Response Examples

```javascript
// Combining helpers with enhanced response objects
registerHandler("POST", "/api/login", (req) => {
    const { username, password } = JSON.parse(req.body || "{}");
    
    if (username === "admin" && password === "secret") {
        // Success with session cookie
        return {
            ...Response.json({ 
                success: true, 
                user: { username, role: "admin" } 
            }),
            cookies: [{
                name: "session",
                value: generateSessionToken(),
                path: "/",
                maxAge: 3600,
                httpOnly: true,
                secure: true
            }]
        };
    }
    
    return Response.error("Invalid credentials", HTTP.UNAUTHORIZED);
});

// Content negotiation example
registerHandler("GET", "/api/data", (req) => {
    const data = db.query("SELECT * FROM items");
    const accept = req.headers.accept || "";
    
    if (accept.includes("text/csv")) {
        let csv = "id,name,value\n";
        data.forEach(item => {
            csv += `${item.id},"${item.name}",${item.value}\n`;
        });
        return {
            body: csv,
            contentType: "text/csv",
            headers: {
                "content-disposition": "attachment; filename=data.csv"
            }
        };
    }
    
    return Response.json(data);
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
// Define reusable utility functions (safe for re-execution)
function formatDate(date) {
    return new Date(date).toISOString().split('T')[0];
}

function generateId() {
    return Math.random().toString(36).substr(2, 9);
}

function validateEmail(email) {
    var re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
}

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