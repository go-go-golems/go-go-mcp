# JavaScript API Reference

## Overview

The JavaScript Playground Server provides a runtime environment where JavaScript code executes in a **persistent global context**. Code runs at the top level without function wrapping, enabling dynamic web application creation with database integration.

**Key Features:**
- **Express.js-compatible API** - Familiar routing and middleware patterns
- **SQLite database access** - Direct database operations via `db` object
- **Persistent state** - `globalState` object survives across executions
- **Real-time endpoint registration** - Add routes dynamically without restarts

**Execution Context:**
- Code runs in **global scope** (no function wrapping)
- **No `return` statements** - Last expression is automatically returned
- **Function definitions** - Use `function name() { }` syntax for reusable functions
- **Variable scoping** - Wrap `const`/`let` in IIFE to avoid global pollution
- **Global variables** - Use `globalState` object for persistent data
- **Persistent runtime** - Functions and global state remain between executions

## Quick Start

```javascript
// Create a simple API endpoint
app.get('/hello', (req, res) => {
  res.json({ message: 'Hello World!' });
});

// Database setup - runs at global scope
db.query(`CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL
)`);

// Define reusable functions
function validateUser(name, email) {
  return name && email && email.includes('@');
}

// CRUD endpoint
app.post('/users', (req, res) => {
  const { name, email } = req.body;
  
  if (!validateUser(name, email)) {
    return res.status(400).json({ error: 'Invalid user data' });
  }
  
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

### **CRITICAL: Inspect Schema First**

**ALWAYS inspect your database schema before performing any operations.** The database persists between code executions, so tables may already exist with data. Check what exists before creating or modifying anything.

#### ✅ CORRECT: Schema Inspection Pattern
```javascript
// ALWAYS start by inspecting existing schema
const tables = db.query(`SELECT name FROM sqlite_master WHERE type='table'`);
console.log('Existing tables:', tables.map(t => t.name));

// Check if specific table exists
const userTableExists = tables.some(t => t.name === 'users');
if (!userTableExists) {
  console.log('Creating users table...');
  db.query(`CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  )`);
} else {
  // Inspect existing schema
  const userSchema = db.query(`PRAGMA table_info(users)`);
  console.log('Users table schema:', userSchema);
}

// Check posts table
const postTableExists = tables.some(t => t.name === 'posts');
if (!postTableExists) {
  console.log('Creating posts table...');
  db.query(`CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT,
    user_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
  )`);
}

// Now safe to perform operations
const users = db.query('SELECT * FROM users LIMIT 5');
console.log('Sample users:', users);
```

#### ❌ WRONG: Operations Without Schema Inspection
```javascript
// DON'T DO THIS - May fail or overwrite existing data
const users = db.query('SELECT * FROM users');  // ERROR if table missing
db.query('CREATE TABLE users...');              // May lose existing data
```

### Query Execution
```javascript
// SELECT queries - returns array of objects
const users = db.query('SELECT * FROM users WHERE active = ?', [true]);

// INSERT/UPDATE/DELETE - returns result object
const result = db.query('INSERT INTO users (name, email) VALUES (?, ?)', [name, email]);
// result: { success: boolean, rowsAffected: number, lastInsertId: number }
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

### **CRITICAL: Always Separate HTML, CSS, and JavaScript**

**DO NOT embed CSS or JavaScript directly in HTML responses.** This creates maintenance nightmares and breaks caching. **ALWAYS** create separate endpoints for each file type.

#### ✅ CORRECT: Separate Endpoints with Proper MIME Types
```javascript
// CSS endpoint - MUST set text/css MIME type
app.get('/static/app.css', (req, res) => {
  const css = `
    body { font-family: Arial, sans-serif; }
    .container { max-width: 800px; margin: 0 auto; }
    .btn { padding: 10px 20px; background: #007bff; color: white; border: none; }
  `;
  res.set('Content-Type', 'text/css');  // REQUIRED for CSS
  res.send(css);
});

// JavaScript endpoint - MUST set application/javascript MIME type
app.get('/static/app.js', (req, res) => {
  const js = `
    document.addEventListener('DOMContentLoaded', function() {
      console.log('App loaded');
      // Your client-side logic here
    });
  `;
  res.set('Content-Type', 'application/javascript');  // REQUIRED for JS
  res.send(js);
});

// HTML page - MUST set text/html MIME type
app.get('/', (req, res) => {
  const html = `
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
  `;
  res.set('Content-Type', 'text/html; charset=utf-8');  // REQUIRED for HTML
  res.send(html);
});
```

#### ❌ WRONG: Embedded Styles/Scripts
```javascript
// DON'T DO THIS - Embedded CSS/JS is bad practice
app.get('/bad-example', (req, res) => {
  res.send(`
    <html>
    <head>
      <style>body { color: red; }</style>  <!-- BAD -->
    </head>
    <body>
      <script>alert('bad');</script>       <!-- BAD -->
    </body>
    </html>
  `);
});
```

**Why separate endpoints with proper MIME types matter:**
- **Browser caching** - Static assets cache independently
- **Development** - Edit CSS/JS without touching HTML
- **Performance** - Parallel loading of resources
- **Maintainability** - Clear separation of concerns
- **Browser compatibility** - Proper MIME types ensure correct parsing
- **Security** - Prevents MIME type sniffing vulnerabilities

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
// Define authentication helper functions
function generateSessionId() {
  return Math.random().toString(36).substr(2, 15);
}

function validateCredentials(email, password) {
  return db.query('SELECT * FROM users WHERE email = ? AND password = ?', [email, password]);
}

function requireAuth(req, res, next) {
  const sessionId = req.cookies.sessionId;
  if (!sessionId || !globalState.sessions.has(sessionId)) {
    return res.status(401).json({ error: 'Authentication required' });
  }
  req.session = globalState.sessions.get(sessionId);
  next();
}

// Initialize sessions in globalState
if (!globalState.sessions) {
  globalState.sessions = new Map();
}

// Login endpoint
app.post('/auth/login', (req, res) => {
  const { email, password } = req.body;
  const users = validateCredentials(email, password);
  
  if (users.length === 0) {
    return res.status(401).json({ error: 'Invalid credentials' });
  }
  
  const sessionId = generateSessionId();
  globalState.sessions.set(sessionId, { userId: users[0].id, user: users[0] });
  
  res.cookie('sessionId', sessionId, { maxAge: 3600000 });
  res.json({ success: true, user: users[0] });
});

// Protected route using helper function
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

## Variable Scoping and Function Definitions

### ✅ CORRECT: Function Definitions and Variable Scoping
```javascript
// Define reusable functions at global scope
function calculateTax(amount, rate) {
  return amount * rate;
}

function formatCurrency(amount) {
  return `$${amount.toFixed(2)}`;
}

// For complex initialization with const/let, use IIFE to avoid global pollution
(function() {
  const CONFIG = {
    taxRate: 0.08,
    currency: 'USD',
    maxItems: 100
  };
  
  const VALIDATION_RULES = {
    email: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
    phone: /^\d{10}$/
  };
  
  // Store in globalState for access across executions
  if (!globalState.appConfig) {
    globalState.appConfig = CONFIG;
    globalState.validationRules = VALIDATION_RULES;
  }
})();

// Use the functions and global state in endpoints
app.post('/calculate', (req, res) => {
  const { amount } = req.body;
  const tax = calculateTax(amount, globalState.appConfig.taxRate);
  const total = amount + tax;
  
  res.json({
    subtotal: formatCurrency(amount),
    tax: formatCurrency(tax),
    total: formatCurrency(total)
  });
});
```

### ❌ WRONG: Global const/let pollution
```javascript
// DON'T DO THIS - Pollutes global namespace
const CONFIG = { taxRate: 0.08 };  // BAD - global const
let currentUser = null;            // BAD - global let

// This creates global variables that can't be redefined on reload
```

### Function Definition Patterns
```javascript
// ✅ Named functions - preferred for reusability
function processOrder(order) {
  return order.items.reduce((total, item) => total + item.price, 0);
}

// ✅ Arrow functions in handlers - fine for inline use
app.get('/orders/:id', (req, res) => {
  const order = getOrder(req.params.id);
  res.json(order);
});

// ✅ Complex logic wrapped in IIFE
(function() {
  const helperFunctions = {
    validateEmail: (email) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email),
    sanitizeInput: (input) => input.trim().toLowerCase()
  };
  
  // Make available globally through globalState
  globalState.helpers = helperFunctions;
})();
```

## Best Practices

1. **Database Schema**: **ALWAYS inspect existing tables and schema before any operations** - the database persists between executions
2. **Static Files**: **NEVER embed CSS/JS in HTML** - always create separate endpoints for each file type
3. **Function Definitions**: Use `function name() { }` syntax for reusable functions at global scope
4. **Variable Scoping**: Wrap `const`/`let` declarations in IIFE `(function() { })()` to avoid global pollution
5. **Global State**: Use `globalState` object for persistent data across executions
6. **State Initialization**: Always check if global state exists before initializing
7. **Error Handling**: Wrap database operations in try-catch blocks
8. **Database Security**: Always use parameterized queries to prevent SQL injection
9. **Session Management**: Use `globalState` for persistent session storage
10. **Status Codes**: Use appropriate HTTP status codes (200, 201, 400, 401, 404, 500)

## Admin Console

Access the admin console at `/admin/logs` to:
- Monitor HTTP requests in real-time
- View console logs and database operations
- Debug application issues
- Track performance metrics

The console automatically captures all JavaScript console output and database operations during request processing. 