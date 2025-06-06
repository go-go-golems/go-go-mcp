# JavaScript API Reference

## Overview

The JavaScript Playground Server provides a runtime environment where JavaScript code executes in a **persistent global context**. Code runs at the top level without function wrapping, enabling dynamic web application creation with database integration.

**Key Features:**
- **Express.js-compatible API** - Familiar routing and middleware patterns
- **SQLite database access** - Direct database operations via `db` object
- **Persistent state** - `globalState` object survives across executions
- **Real-time endpoint registration** - Add routes dynamically without restarts
- **Admin console** - Monitor requests and logs at `/admin/logs`

**Execution Context:**
- Code runs in **global scope** (no function wrapping)
- **No `return` statements** - Last expression is automatically returned
- **Avoid `let`/`const`** - Use `var` or direct assignment for reloadable code
- **Persistent runtime** - Variables and functions remain between executions

## Quick Start

```javascript
// Create a simple API endpoint
app.get('/hello', (req, res) => {
  res.json({ message: 'Hello World!' });
});

// Database setup
db.query(`CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL
)`);

// CRUD endpoint
app.post('/users', (req, res) => {
  const { name, email } = req.body;
  db.query('INSERT INTO users (name, email) VALUES (?, ?)', [name, email]);
  res.status(201).json({ success: true });
});
```

## Express.js API

### Route Registration
```javascript
app.get('/path', handler)     // GET requests
app.post('/path', handler)    // POST requests  
app.put('/path', handler)     // PUT requests
app.delete('/path', handler)  // DELETE requests
app.patch('/path', handler)   // PATCH requests

// Path parameters
app.get('/users/:id', (req, res) => {
  const userId = req.params.id;
  res.json({ userId });
});
```

### Request Object
```javascript
app.post('/data', (req, res) => {
  const method = req.method;        // HTTP method
  const path = req.path;            // URL path
  const query = req.query;          // Query parameters
  const params = req.params;        // Path parameters
  const body = req.body;            // Request body (auto-parsed JSON)
  const headers = req.headers;      // Request headers
  const cookies = req.cookies;      // Parsed cookies
  const ip = req.ip;                // Client IP
});
```

### Response Methods
```javascript
res.json(data)                    // JSON response
res.send(text)                    // Text/HTML response
res.status(code)                  // Set status code
res.set(header, value)            // Set header
res.cookie(name, value, options)  // Set cookie
res.redirect(url)                 // Redirect
res.end()                         // Empty response
```

## Database Operations

### Query Execution
```javascript
// SELECT queries - returns array of objects
const users = db.query('SELECT * FROM users WHERE active = ?', [true]);

// INSERT/UPDATE/DELETE - returns result object
const result = db.query('INSERT INTO users (name, email) VALUES (?, ?)', [name, email]);
// result: { success: boolean, rowsAffected: number, lastInsertId: number }
```

### Common Patterns
```javascript
// Create table
db.query(`CREATE TABLE IF NOT EXISTS posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  content TEXT,
  user_id INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`);

// CRUD operations
const user = db.query('SELECT * FROM users WHERE id = ?', [userId])[0];
db.query('UPDATE users SET name = ? WHERE id = ?', [newName, userId]);
db.query('DELETE FROM users WHERE id = ?', [userId]);
```

## State Management

### Global State
```javascript
// Initialize persistent state
if (!globalState.app) {
  globalState.app = {
    requestCount: 0,
    config: { maxUsers: 100 }
  };
}

// Use across requests
app.get('/stats', (req, res) => {
  globalState.app.requestCount++;
  res.json({ requests: globalState.app.requestCount });
});
```

### Session Management
```javascript
// Initialize sessions
if (!globalState.sessions) {
  globalState.sessions = new Map();
}

// Create session
app.post('/login', (req, res) => {
  const sessionId = Math.random().toString(36).substr(2, 9);
  globalState.sessions.set(sessionId, { userId: 123, createdAt: new Date() });
  res.cookie('sessionId', sessionId);
  res.json({ success: true });
});
```

## Static File Serving

### Best Practice: Separate Endpoints
```javascript
// CSS endpoint
app.get('/static/app.css', (req, res) => {
  const css = `
    body { font-family: Arial, sans-serif; }
    .container { max-width: 800px; margin: 0 auto; }
  `;
  res.set('Content-Type', 'text/css');
  res.send(css);
});

// JavaScript endpoint  
app.get('/static/app.js', (req, res) => {
  const js = `
    document.addEventListener('DOMContentLoaded', function() {
      console.log('App loaded');
    });
  `;
  res.set('Content-Type', 'application/javascript');
  res.send(js);
});

// HTML page referencing separate assets
app.get('/', (req, res) => {
  res.send(`
    <!DOCTYPE html>
    <html>
    <head>
      <title>My App</title>
      <link rel="stylesheet" href="/static/app.css">
    </head>
    <body>
      <div class="container">Content</div>
      <script src="/static/app.js"></script>
    </body>
    </html>
  `);
});
```

## Complete Examples

### Simple Blog API
```javascript
// Setup
db.query(`CREATE TABLE IF NOT EXISTS posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`);

// List posts
app.get('/posts', (req, res) => {
  const posts = db.query('SELECT * FROM posts ORDER BY created_at DESC');
  res.json(posts);
});

// Create post
app.post('/posts', (req, res) => {
  const { title, content } = req.body;
  db.query('INSERT INTO posts (title, content) VALUES (?, ?)', [title, content]);
  res.status(201).json({ success: true });
});

// Get single post
app.get('/posts/:id', (req, res) => {
  const posts = db.query('SELECT * FROM posts WHERE id = ?', [req.params.id]);
  if (posts.length === 0) return res.status(404).json({ error: 'Not found' });
  res.json(posts[0]);
});
```

### Authentication System
```javascript
// Initialize sessions
if (!globalState.sessions) globalState.sessions = new Map();

// Login
app.post('/auth/login', (req, res) => {
  const { email, password } = req.body;
  const users = db.query('SELECT * FROM users WHERE email = ? AND password = ?', [email, password]);
  
  if (users.length === 0) {
    return res.status(401).json({ error: 'Invalid credentials' });
  }
  
  const sessionId = Math.random().toString(36).substr(2, 15);
  globalState.sessions.set(sessionId, { userId: users[0].id, user: users[0] });
  
  res.cookie('sessionId', sessionId, { maxAge: 3600000 });
  res.json({ success: true, user: users[0] });
});

// Protected route
app.get('/profile', (req, res) => {
  const sessionId = req.cookies.sessionId;
  if (!sessionId || !globalState.sessions.has(sessionId)) {
    return res.status(401).json({ error: 'Authentication required' });
  }
  
  const session = globalState.sessions.get(sessionId);
  res.json({ user: session.user });
});
```

## Error Handling

```javascript
app.get('/users/:id', (req, res) => {
  try {
    const users = db.query('SELECT * FROM users WHERE id = ?', [req.params.id]);
    if (users.length === 0) {
      return res.status(404).json({ error: 'User not found' });
    }
    res.json(users[0]);
  } catch (error) {
    console.error('Database error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});
```

## Best Practices

1. **Variable Declarations**: Use `var` or direct assignment instead of `let`/`const` for reloadable code
2. **State Initialization**: Always check if global state exists before initializing
3. **Error Handling**: Wrap database operations in try-catch blocks
4. **Static Files**: Create separate endpoints for CSS/JS instead of embedding in HTML
5. **Database Security**: Always use parameterized queries to prevent SQL injection
6. **Session Management**: Use `globalState` for persistent session storage
7. **Status Codes**: Use appropriate HTTP status codes (200, 201, 400, 401, 404, 500)

## Admin Console

Access the admin console at `/admin/logs` to:
- Monitor HTTP requests in real-time
- View console logs and database operations
- Debug application issues
- Track performance metrics

The console automatically captures all JavaScript console output and database operations during request processing. 