# JavaScript Application Developer Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Sandboxed Environment Overview](#sandboxed-environment-overview)
4. [Express.js API Reference](#expressjs-api-reference)
5. [Database Integration](#database-integration)
6. [Global State Management](#global-state-management)
7. [Application Architecture Patterns](#application-architecture-patterns)
8. [Complete Examples](#complete-examples)
9. [Best Practices](#best-practices)
10. [Troubleshooting](#troubleshooting)

## Introduction

The JavaScript Playground Server provides a unique development environment that combines the familiar Express.js web framework with a secure, sandboxed JavaScript runtime powered by Go. This environment enables rapid prototyping and deployment of web applications without the complexity of traditional Node.js setups while maintaining security through isolated execution.

### What Makes This Environment Special

This sandboxed JavaScript environment offers several compelling advantages for web application development:

**Runtime Flexibility**: Unlike traditional web servers that require restarts for code changes, this environment allows you to modify your application logic, add new endpoints, and update database schemas entirely at runtime. Your JavaScript code is executed in a secure Goja runtime that provides V8-compatible JavaScript execution without the overhead of a full Node.js environment.

**Express.js Familiarity**: The environment provides a complete Express.js compatible API, meaning developers can leverage existing Express.js knowledge and patterns. Route handlers, middleware, request/response objects, and HTTP method routing all work exactly as expected in Express.js applications.

**Integrated Database**: SQLite database access is built directly into the JavaScript runtime, eliminating the need for external database drivers or connection pools. Database operations are automatically parameterized to prevent SQL injection, and the connection is managed transparently by the runtime.

**Security Through Isolation**: JavaScript code runs in a completely sandboxed environment with no access to the file system, network, or other system resources beyond the provided APIs. This makes it safe to execute dynamic code without compromising the host system.

## Express.js API

### Routes

```javascript
// Basic routes
app.get("/users", (req, res) => res.json(users));
app.post("/users", (req, res) => res.status(201).json(newUser));
app.put("/users/:id", (req, res) => res.json(updatedUser));
app.delete("/users/:id", (req, res) => res.status(204).end());

// Path parameters
app.get("/users/:id", (req, res) => {
  const userId = req.params.id;
  res.json({ userId });
});

// Multiple parameters
app.get("/api/:version/users/:id", (req, res) => {
  const { version, id } = req.params;
  res.json({ version, id });
});
```

### Request Object

```javascript
app.post("/data", (req, res) => {
  // Essential request properties
  const method = req.method; // 'POST'
  const path = req.path; // '/data'
  const query = req.query; // ?key=value -> { key: 'value' }
  const params = req.params; // Path parameters
  const body = req.body; // Request body (auto-parsed JSON)
  const headers = req.headers; // All headers
  const cookies = req.cookies; // Parsed cookies
  const ip = req.ip; // Client IP

  res.json({ received: "ok" });
});
```

### Response Methods

```javascript
app.get("/examples", (req, res) => {
  // JSON response
  res.json({ data: "value" });

  // Status codes
  res.status(404).json({ error: "Not found" });
  res.status(201).json({ created: true });

  // Headers
  res.set("X-Custom", "value");

  // Cookies
  res.cookie("session", "abc123", { maxAge: 3600000 });

  // Redirects
  res.redirect("/new-location");
  res.redirect(301, "/moved-permanently");

  // Text/HTML
  res.send("<h1>HTML Response</h1>");
  res.send("Plain text");

  // Empty response
  res.end();
});
```

## Database Integration

### Quick Setup

```javascript
// Create table
db.query(`
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
`);
```

### CRUD Operations

```javascript
// Create
app.post("/users", (req, res) => {
  const { name, email } = req.body;
  db.query("INSERT INTO users (name, email) VALUES (?, ?)", [name, email]);

  const users = db.query("SELECT * FROM users WHERE email = ?", [email]);
  res.status(201).json(users[0]);
});

// Read
app.get("/users", (req, res) => {
  const users = db.query("SELECT * FROM users ORDER BY created_at DESC");
  res.json(users);
});

app.get("/users/:id", (req, res) => {
  const users = db.query("SELECT * FROM users WHERE id = ?", [req.params.id]);
  if (users.length === 0)
    return res.status(404).json({ error: "User not found" });
  res.json(users[0]);
});

// Update
app.put("/users/:id", (req, res) => {
  const { name, email } = req.body;
  db.query("UPDATE users SET name = ?, email = ? WHERE id = ?", [
    name,
    email,
    req.params.id,
  ]);

  const users = db.query("SELECT * FROM users WHERE id = ?", [req.params.id]);
  res.json(users[0]);
});

// Delete
app.delete("/users/:id", (req, res) => {
  db.query("DELETE FROM users WHERE id = ?", [req.params.id]);
  res.status(204).end();
});
```

### Advanced Queries

```javascript
// Search with filters
app.get("/users/search", (req, res) => {
  const { q, limit = 10 } = req.query;

  if (q) {
    const users = db.query(
      "SELECT * FROM users WHERE name LIKE ? OR email LIKE ? LIMIT ?",
      [`%${q}%`, `%${q}%`, limit]
    );
    res.json(users);
  } else {
    const users = db.query("SELECT * FROM users LIMIT ?", [limit]);
    res.json(users);
  }
});

// Aggregations
app.get("/stats", (req, res) => {
  const stats = db.query(`
        SELECT 
            COUNT(*) as total_users,
            COUNT(CASE WHEN created_at >= date('now', '-30 days') THEN 1 END) as new_users_30d
        FROM users
    `)[0];

  res.json(stats);
});
```

## Express.js API Reference

### Application Object (`app`)

The `app` object provides the core routing functionality, mirroring the Express.js application interface:

#### Route Methods

All HTTP methods are supported with identical syntax to Express.js:

```javascript
// GET route
app.get("/users", (req, res) => {
  res.json({ message: "Get all users" });
});

// POST route
app.post("/users", (req, res) => {
  const { name, email } = req.body;
  res.status(201).json({ id: 1, name, email });
});

// PUT route
app.put("/users/:id", (req, res) => {
  const userId = req.params.id;
  res.json({ message: `Updated user ${userId}` });
});

// DELETE route
app.delete("/users/:id", (req, res) => {
  const userId = req.params.id;
  res.status(204).end();
});

// PATCH route
app.patch("/users/:id", (req, res) => {
  const userId = req.params.id;
  res.json({ message: `Partially updated user ${userId}` });
});
```

#### Route Parameters

Route parameters work exactly like Express.js, with support for dynamic segments:

```javascript
// Single parameter
app.get("/users/:id", (req, res) => {
  const userId = req.params.id;
  res.json({ userId });
});

// Multiple parameters
app.get("/users/:userId/posts/:postId", (req, res) => {
  const { userId, postId } = req.params;
  res.json({ userId, postId });
});

// Optional parameters (basic implementation)
app.get("/posts/:id/:slug?", (req, res) => {
  const { id, slug } = req.params;
  res.json({ id, slug: slug || "no-slug" });
});
```

#### Middleware (Basic Implementation)

Basic middleware support is available through `app.use()`:

```javascript
// Path-specific middleware
app.use("/api", (req, res) => {
  console.log(`API request: ${req.method} ${req.path}`);
  // Note: This implementation registers handlers for all HTTP methods
});

// This will create handlers for GET, POST, PUT, DELETE, PATCH on /api/*
```

### Request Object (`req`) - Comprehensive HTTP Request Interface

The request object provides complete access to all aspects of the incoming HTTP request, from basic metadata like method and URL to complex data like headers, cookies, and request bodies. Understanding the request object's capabilities is essential for building sophisticated web applications that can respond appropriately to different types of client requests.

The request object follows Express.js conventions exactly, ensuring that existing Express.js code patterns work without modification. All request properties are parsed and normalized by the runtime, providing consistent access to request data regardless of the client or HTTP version used.

#### Core Request Properties and Metadata

The core properties provide essential information about the request that's typically needed for routing decisions, logging, and basic request handling. These properties are automatically parsed from the HTTP request and made available as simple JavaScript values.

This example demonstrates accessing and using the fundamental request properties:

```javascript
app.get("/request-info", (req, res) => {
  // Comprehensive request information extraction
  const requestInfo = {
    // HTTP method (normalized to lowercase for consistency)
    method: req.method, // 'get', 'post', 'put', etc.

    // Complete URL including query string
    url: req.url, // '/request-info?debug=true&format=json'

    // Path portion only (no query string)
    path: req.path, // '/request-info'

    // Protocol information (http vs https)
    protocol: req.protocol, // 'http' or 'https'

    // Hostname from Host header
    hostname: req.hostname, // 'localhost', 'api.example.com'

    // Client IP address (with proxy support)
    ip: req.ip, // '127.0.0.1' or forwarded IP

    // Additional metadata
    timestamp: new Date().toISOString(),
    userAgent: req.headers["user-agent"] || "Unknown",
  };

  // Log request for debugging
  console.log(`${req.method} ${req.path} from ${req.ip}`);

  res.json({
    message: "Request information extracted successfully",
    request: requestInfo,
  });
});
```

#### Query Parameters

Query parameters are automatically parsed and available as an object:

```javascript
app.get("/search", (req, res) => {
  // URL: /search?q=javascript&limit=10&sort=date
  const query = req.query;
  /*
    query = {
        q: 'javascript',
        limit: '10',
        sort: 'date'
    }
    */

  // Individual access
  const searchTerm = req.query.q;
  const limit = parseInt(req.query.limit) || 20;

  res.json({ searchTerm, limit, query });
});
```

#### Headers

All HTTP headers are available in the headers object:

```javascript
app.get("/headers", (req, res) => {
  // Access specific headers
  const userAgent = req.headers["user-agent"];
  const contentType = req.headers["content-type"];
  const authorization = req.headers["authorization"];

  // All headers
  const allHeaders = req.headers;

  res.json({
    userAgent,
    contentType,
    authorization,
    allHeaders,
  });
});
```

#### Request Body

Request bodies are automatically parsed based on Content-Type:

```javascript
app.post("/data", (req, res) => {
  // JSON bodies are automatically parsed
  if (req.headers["content-type"]?.includes("application/json")) {
    const jsonData = req.body; // Already parsed as JavaScript object
    res.json({ received: jsonData });
  } else {
    const textData = req.body; // Raw string for other content types
    res.json({ received: textData });
  }
});
```

#### Cookies

Cookies are parsed and available as an object:

```javascript
app.get("/profile", (req, res) => {
  const sessionId = req.cookies.sessionId;
  const preferences = req.cookies.preferences;

  if (!sessionId) {
    return res.status(401).json({ error: "No session cookie" });
  }

  res.json({ sessionId, preferences });
});
```

### Response Object (`res`)

The response object provides all standard Express.js response methods:

#### Content Methods

```javascript
// Send JSON response
app.get("/json", (req, res) => {
  res.json({ message: "JSON response", data: [1, 2, 3] });
});

// Send HTML response
app.get("/html", (req, res) => {
  res.send("<h1>HTML Response</h1><p>This is HTML content</p>");
});

// Send plain text
app.get("/text", (req, res) => {
  res.send("Plain text response");
});

// Send empty response
app.get("/empty", (req, res) => {
  res.end();
});
```

#### Status Codes

```javascript
// Set status code
app.get("/not-found", (req, res) => {
  res.status(404).json({ error: "Resource not found" });
});

// Method chaining
app.post("/created", (req, res) => {
  res.status(201).json({ message: "Resource created" });
});

// Various status codes
app.get("/status-examples", (req, res) => {
  const examples = {
    success: 200,
    created: 201,
    noContent: 204,
    badRequest: 400,
    unauthorized: 401,
    forbidden: 403,
    notFound: 404,
    serverError: 500,
  };

  res.json(examples);
});
```

#### Headers

```javascript
// Set individual headers
app.get("/custom-headers", (req, res) => {
  res.set("X-Custom-Header", "MyValue");
  res.set("Cache-Control", "max-age=3600");
  res.json({ message: "Response with custom headers" });
});

// Set multiple headers
app.get("/multiple-headers", (req, res) => {
  res.set("X-API-Version", "1.0");
  res.set("X-Rate-Limit", "1000");
  res.json({ message: "Multiple headers set" });
});
```

#### Cookies

```javascript
// Set simple cookie
app.get("/login", (req, res) => {
  res.cookie("sessionId", "abc123");
  res.json({ message: "Logged in" });
});

// Set cookie with options
app.get("/secure-login", (req, res) => {
  res.cookie("sessionId", "abc123", {
    maxAge: 3600000, // 1 hour in milliseconds
    httpOnly: true, // Prevent JavaScript access
    secure: true, // HTTPS only
    path: "/", // Cookie path
  });
  res.json({ message: "Secure login" });
});
```

#### Redirects

```javascript
// Temporary redirect (302)
app.get("/old-page", (req, res) => {
  res.redirect("/new-page");
});

// Permanent redirect (301)
app.get("/moved", (req, res) => {
  res.redirect(301, "/new-location");
});

// External redirect
app.get("/external", (req, res) => {
  res.redirect("https://example.com");
});
```

## Database Integration

### Overview

The sandboxed environment provides direct SQLite database access through the `db` object. This integration offers several advantages over traditional database drivers:

- **Automatic Parameter Binding**: All queries are automatically parameterized to prevent SQL injection
- **Type Conversion**: JavaScript values are automatically converted to appropriate SQL types
- **Transaction Support**: Built-in transaction handling for data consistency
- **Connection Management**: Database connections are managed transparently by the runtime

### Basic Database Operations

#### Creating Tables and Schema Design

Database schema design in this environment follows standard SQLite practices while benefiting from automatic type handling and constraint enforcement. The schema creation process demonstrates how to build robust data models that support complex application requirements while maintaining referential integrity.

Understanding SQLite's type system and constraint capabilities is crucial for building reliable applications. SQLite provides flexible typing with automatic type conversion, foreign key support, and comprehensive constraint enforcement that helps maintain data quality.

This example demonstrates comprehensive schema design with proper constraints, relationships, and indexing:

```javascript
// Complete database schema setup with comprehensive table design
app.post("/setup-database", (req, res) => {
  try {
    // Users table with proper constraints and defaults
    db.query(`
            CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL CHECK(length(name) >= 2),
                email TEXT UNIQUE NOT NULL CHECK(email LIKE '%@%.%'),
                password_hash TEXT NOT NULL,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                is_active BOOLEAN DEFAULT 1,
                last_login DATETIME,
                profile_data TEXT -- JSON data stored as text
            )
        `);

    // Posts table with foreign key relationship
    db.query(`
            CREATE TABLE IF NOT EXISTS posts (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER NOT NULL,
                title TEXT NOT NULL CHECK(length(title) >= 5),
                content TEXT NOT NULL,
                excerpt TEXT,
                status TEXT DEFAULT 'draft' CHECK(status IN ('draft', 'published', 'archived')),
                view_count INTEGER DEFAULT 0,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                published_at DATETIME,
                FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
            )
        `);

    // Categories table for many-to-many relationships
    db.query(`
            CREATE TABLE IF NOT EXISTS categories (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT UNIQUE NOT NULL,
                description TEXT,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        `);

    // Junction table for post-category relationships
    db.query(`
            CREATE TABLE IF NOT EXISTS post_categories (
                post_id INTEGER,
                category_id INTEGER,
                PRIMARY KEY (post_id, category_id),
                FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
                FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
            )
        `);

    // Create indexes for performance
    db.query(`CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)`);
    db.query(`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id)`);
    db.query(`CREATE INDEX IF NOT EXISTS idx_posts_status ON posts (status)`);
    db.query(
      `CREATE INDEX IF NOT EXISTS idx_posts_published_at ON posts (published_at DESC)`
    );

    console.log("Database schema created successfully");
    res.json({
      message: "Database tables and indexes created successfully",
      tables: ["users", "posts", "categories", "post_categories"],
      indexes: 4,
    });
  } catch (error) {
    console.error("Database setup error:", error);
    res.status(500).json({
      error: "Failed to setup database",
      details: error.message,
    });
  }
});
```

#### Inserting Data

```javascript
// Insert single user
app.post("/users", (req, res) => {
  const { name, email } = req.body;

  if (!name || !email) {
    return res.status(400).json({ error: "Name and email are required" });
  }

  try {
    db.query("INSERT INTO users (name, email) VALUES (?, ?)", [name, email]);

    // Get the inserted user
    const users = db.query("SELECT * FROM users WHERE email = ?", [email]);

    res.status(201).json(users[0]);
  } catch (error) {
    console.error("Insert error:", error);
    res.status(500).json({ error: "Failed to create user" });
  }
});

// Bulk insert
app.post("/users/bulk", (req, res) => {
  const users = req.body.users;

  if (!Array.isArray(users)) {
    return res.status(400).json({ error: "Users must be an array" });
  }

  try {
    let insertedCount = 0;

    users.forEach((user) => {
      db.query("INSERT INTO users (name, email) VALUES (?, ?)", [
        user.name,
        user.email,
      ]);
      insertedCount++;
    });

    res.status(201).json({
      message: `${insertedCount} users created successfully`,
    });
  } catch (error) {
    console.error("Bulk insert error:", error);
    res.status(500).json({ error: "Failed to create users" });
  }
});
```

#### Querying Data

```javascript
// Get all users
app.get("/users", (req, res) => {
  const page = parseInt(req.query.page) || 1;
  const limit = parseInt(req.query.limit) || 10;
  const offset = (page - 1) * limit;

  try {
    // Get paginated users
    const users = db.query(
      "SELECT id, name, email, created_at, is_active FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?",
      [limit, offset]
    );

    // Get total count
    const countResult = db.query("SELECT COUNT(*) as total FROM users");
    const total = countResult[0].total;

    res.json({
      users,
      pagination: {
        page,
        limit,
        total,
        pages: Math.ceil(total / limit),
      },
    });
  } catch (error) {
    console.error("Query error:", error);
    res.status(500).json({ error: "Failed to fetch users" });
  }
});

// Get user by ID
app.get("/users/:id", (req, res) => {
  const userId = parseInt(req.params.id);

  if (isNaN(userId)) {
    return res.status(400).json({ error: "Invalid user ID" });
  }

  try {
    const users = db.query(
      "SELECT id, name, email, created_at, is_active FROM users WHERE id = ?",
      [userId]
    );

    if (users.length === 0) {
      return res.status(404).json({ error: "User not found" });
    }

    // Get user's posts
    const posts = db.query(
      "SELECT id, title, content, created_at FROM posts WHERE user_id = ? ORDER BY created_at DESC",
      [userId]
    );

    res.json({
      user: users[0],
      posts,
    });
  } catch (error) {
    console.error("Query error:", error);
    res.status(500).json({ error: "Failed to fetch user" });
  }
});
```

#### Updating Data

```javascript
// Update user
app.put("/users/:id", (req, res) => {
  const userId = parseInt(req.params.id);
  const { name, email, is_active } = req.body;

  if (isNaN(userId)) {
    return res.status(400).json({ error: "Invalid user ID" });
  }

  try {
    // Check if user exists
    const existingUsers = db.query("SELECT id FROM users WHERE id = ?", [
      userId,
    ]);

    if (existingUsers.length === 0) {
      return res.status(404).json({ error: "User not found" });
    }

    // Update user
    db.query(
      "UPDATE users SET name = ?, email = ?, is_active = ? WHERE id = ?",
      [name, email, is_active, userId]
    );

    // Return updated user
    const updatedUsers = db.query(
      "SELECT id, name, email, created_at, is_active FROM users WHERE id = ?",
      [userId]
    );

    res.json(updatedUsers[0]);
  } catch (error) {
    console.error("Update error:", error);
    res.status(500).json({ error: "Failed to update user" });
  }
});
```

#### Deleting Data

```javascript
// Delete user
app.delete("/users/:id", (req, res) => {
  const userId = parseInt(req.params.id);

  if (isNaN(userId)) {
    return res.status(400).json({ error: "Invalid user ID" });
  }

  try {
    // Check if user exists
    const existingUsers = db.query("SELECT id FROM users WHERE id = ?", [
      userId,
    ]);

    if (existingUsers.length === 0) {
      return res.status(404).json({ error: "User not found" });
    }

    // Delete user's posts first (foreign key constraint)
    db.query("DELETE FROM posts WHERE user_id = ?", [userId]);

    // Delete user
    db.query("DELETE FROM users WHERE id = ?", [userId]);

    res.status(204).end();
  } catch (error) {
    console.error("Delete error:", error);
    res.status(500).json({ error: "Failed to delete user" });
  }
});
```

### Advanced Database Patterns

#### Search and Filtering

```javascript
// Advanced user search
app.get("/users/search", (req, res) => {
  const { q, active, created_after, created_before } = req.query;

  let sql =
    "SELECT id, name, email, created_at, is_active FROM users WHERE 1=1";
  let params = [];

  // Text search
  if (q) {
    sql += " AND (name LIKE ? OR email LIKE ?)";
    params.push(`%${q}%`, `%${q}%`);
  }

  // Active filter
  if (active !== undefined) {
    sql += " AND is_active = ?";
    params.push(active === "true" ? 1 : 0);
  }

  // Date range filters
  if (created_after) {
    sql += " AND created_at >= ?";
    params.push(created_after);
  }

  if (created_before) {
    sql += " AND created_at <= ?";
    params.push(created_before);
  }

  sql += " ORDER BY created_at DESC";

  try {
    const users = db.query(sql, params);
    res.json({ users, query: req.query });
  } catch (error) {
    console.error("Search error:", error);
    res.status(500).json({ error: "Search failed" });
  }
});
```

#### Aggregations and Statistics

```javascript
// Dashboard statistics
app.get("/dashboard/stats", (req, res) => {
  try {
    // User statistics
    const userStats = db.query(`
            SELECT 
                COUNT(*) as total_users,
                COUNT(CASE WHEN is_active = 1 THEN 1 END) as active_users,
                COUNT(CASE WHEN created_at >= date('now', '-30 days') THEN 1 END) as new_users_30d
            FROM users
        `)[0];

    // Post statistics
    const postStats = db.query(`
            SELECT 
                COUNT(*) as total_posts,
                COUNT(CASE WHEN created_at >= date('now', '-7 days') THEN 1 END) as posts_last_week
            FROM posts
        `)[0];

    // Top active users
    const topUsers = db.query(`
            SELECT 
                u.name, 
                u.email, 
                COUNT(p.id) as post_count
            FROM users u
            LEFT JOIN posts p ON u.id = p.user_id
            WHERE u.is_active = 1
            GROUP BY u.id, u.name, u.email
            ORDER BY post_count DESC
            LIMIT 5
        `);

    res.json({
      users: userStats,
      posts: postStats,
      topUsers,
    });
  } catch (error) {
    console.error("Stats error:", error);
    res.status(500).json({ error: "Failed to fetch statistics" });
  }
});
```

## Global State Management

### Overview

The global state object (`globalState`) provides persistent data storage that survives across code executions. This is essential for maintaining application state, configuration, and shared data between different parts of your application.

### Basic Global State Usage

```javascript
// Initialize application state
if (!globalState.app) {
  globalState.app = {
    version: "1.0.0",
    startTime: new Date(),
    requestCount: 0,
    config: {
      maxPageSize: 100,
      defaultPageSize: 20,
    },
  };
  console.log("Application state initialized");
}

// Increment request counter
app.use("/", (req, res) => {
  globalState.app.requestCount++;
});

// Application info endpoint
app.get("/app/info", (req, res) => {
  const uptime = new Date() - globalState.app.startTime;

  res.json({
    version: globalState.app.version,
    uptime: Math.floor(uptime / 1000), // seconds
    requestCount: globalState.app.requestCount,
    config: globalState.app.config,
  });
});
```

### Session Management

```javascript
// Initialize session storage
if (!globalState.sessions) {
  globalState.sessions = new Map();
}

// Session middleware
function requireSession(req, res, next) {
  const sessionId = req.cookies.sessionId;

  if (!sessionId || !globalState.sessions.has(sessionId)) {
    return res.status(401).json({ error: "Valid session required" });
  }

  req.session = globalState.sessions.get(sessionId);
  next();
}

// Login endpoint
app.post("/auth/login", (req, res) => {
  const { username, password } = req.body;

  // Simple authentication (use proper hashing in production)
  const users = db.query(
    "SELECT id, name, email FROM users WHERE email = ? AND password = ?",
    [username, password]
  );

  if (users.length === 0) {
    return res.status(401).json({ error: "Invalid credentials" });
  }

  // Create session
  const sessionId = Math.random().toString(36).substring(2, 15);
  const session = {
    id: sessionId,
    userId: users[0].id,
    user: users[0],
    createdAt: new Date(),
    lastActivity: new Date(),
  };

  globalState.sessions.set(sessionId, session);

  res.cookie("sessionId", sessionId, { maxAge: 3600000 }); // 1 hour
  res.json({ message: "Logged in successfully", user: users[0] });
});

// Protected endpoint
app.get("/profile", requireSession, (req, res) => {
  req.session.lastActivity = new Date();
  res.json({ user: req.session.user });
});
```

### Caching

```javascript
// Initialize cache
if (!globalState.cache) {
  globalState.cache = {
    data: new Map(),
    stats: { hits: 0, misses: 0 },
  };
}

// Cache helper functions
function getCached(key) {
  if (globalState.cache.data.has(key)) {
    const item = globalState.cache.data.get(key);
    if (item.expires > Date.now()) {
      globalState.cache.stats.hits++;
      return item.value;
    } else {
      globalState.cache.data.delete(key);
    }
  }
  globalState.cache.stats.misses++;
  return null;
}

function setCache(key, value, ttlSeconds = 300) {
  globalState.cache.data.set(key, {
    value,
    expires: Date.now() + ttlSeconds * 1000,
  });
}

// Cached user endpoint
app.get("/users/:id/cached", (req, res) => {
  const userId = req.params.id;
  const cacheKey = `user:${userId}`;

  // Try cache first
  let user = getCached(cacheKey);

  if (!user) {
    // Cache miss - fetch from database
    const users = db.query(
      "SELECT id, name, email, created_at FROM users WHERE id = ?",
      [userId]
    );

    if (users.length === 0) {
      return res.status(404).json({ error: "User not found" });
    }

    user = users[0];
    setCache(cacheKey, user, 600); // Cache for 10 minutes
  }

  res.json({
    user,
    cached: globalState.cache.stats.hits > 0,
    cacheStats: globalState.cache.stats,
  });
});
```

## Application Architecture Patterns

### MVC Pattern Implementation

```javascript
// Initialize MVC structure in global state
if (!globalState.mvc) {
  globalState.mvc = {
    models: {},
    views: {},
    controllers: {},
  };
}

// Model layer
globalState.mvc.models.User = {
  findAll: (filters = {}) => {
    let sql = "SELECT * FROM users WHERE 1=1";
    let params = [];

    if (filters.active !== undefined) {
      sql += " AND is_active = ?";
      params.push(filters.active);
    }

    return db.query(sql, params);
  },

  findById: (id) => {
    const users = db.query("SELECT * FROM users WHERE id = ?", [id]);
    return users.length > 0 ? users[0] : null;
  },

  create: (userData) => {
    db.query("INSERT INTO users (name, email) VALUES (?, ?)", [
      userData.name,
      userData.email,
    ]);
    return globalState.mvc.models.User.findByEmail(userData.email);
  },

  findByEmail: (email) => {
    const users = db.query("SELECT * FROM users WHERE email = ?", [email]);
    return users.length > 0 ? users[0] : null;
  },
};

// View layer
globalState.mvc.views.User = {
  index: (users) => ({
    users: users.map((user) => ({
      id: user.id,
      name: user.name,
      email: user.email,
      createdAt: user.created_at,
    })),
  }),

  show: (user) => ({
    id: user.id,
    name: user.name,
    email: user.email,
    createdAt: user.created_at,
    isActive: user.is_active,
  }),

  error: (message, code = 500) => ({
    error: message,
    code,
  }),
};

// Controller layer
globalState.mvc.controllers.UserController = {
  index: (req, res) => {
    try {
      const users = globalState.mvc.models.User.findAll();
      const viewData = globalState.mvc.views.User.index(users);
      res.json(viewData);
    } catch (error) {
      console.error("UserController.index error:", error);
      const errorView = globalState.mvc.views.User.error(
        "Failed to fetch users"
      );
      res.status(500).json(errorView);
    }
  },

  show: (req, res) => {
    try {
      const user = globalState.mvc.models.User.findById(req.params.id);
      if (!user) {
        const errorView = globalState.mvc.views.User.error(
          "User not found",
          404
        );
        return res.status(404).json(errorView);
      }
      const viewData = globalState.mvc.views.User.show(user);
      res.json(viewData);
    } catch (error) {
      console.error("UserController.show error:", error);
      const errorView = globalState.mvc.views.User.error(
        "Failed to fetch user"
      );
      res.status(500).json(errorView);
    }
  },

  create: (req, res) => {
    try {
      const user = globalState.mvc.models.User.create(req.body);
      const viewData = globalState.mvc.views.User.show(user);
      res.status(201).json(viewData);
    } catch (error) {
      console.error("UserController.create error:", error);
      const errorView = globalState.mvc.views.User.error(
        "Failed to create user"
      );
      res.status(500).json(errorView);
    }
  },
};

// Route registration using controllers
app.get("/mvc/users", globalState.mvc.controllers.UserController.index);
app.get("/mvc/users/:id", globalState.mvc.controllers.UserController.show);
app.post("/mvc/users", globalState.mvc.controllers.UserController.create);
```

### Service Layer Pattern

```javascript
// Initialize services
if (!globalState.services) {
  globalState.services = {};
}

// User service
globalState.services.UserService = {
  async validateUser(userData) {
    const errors = [];

    if (!userData.name || userData.name.length < 2) {
      errors.push("Name must be at least 2 characters");
    }

    if (!userData.email || !userData.email.includes("@")) {
      errors.push("Valid email is required");
    }

    // Check if email already exists
    const existingUser = globalState.mvc.models.User.findByEmail(
      userData.email
    );
    if (existingUser) {
      errors.push("Email already exists");
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  },

  async createUser(userData) {
    const validation = await this.validateUser(userData);

    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(", ")}`);
    }

    return globalState.mvc.models.User.create(userData);
  },

  async getUserWithStats(userId) {
    const user = globalState.mvc.models.User.findById(userId);
    if (!user) {
      throw new Error("User not found");
    }

    // Get additional stats
    const posts = db.query(
      "SELECT COUNT(*) as count FROM posts WHERE user_id = ?",
      [userId]
    );
    const postCount = posts[0].count;

    return {
      ...user,
      stats: {
        postCount,
      },
    };
  },
};

// Service-based endpoints
app.post("/services/users", async (req, res) => {
  try {
    const user = await globalState.services.UserService.createUser(req.body);
    res.status(201).json(user);
  } catch (error) {
    console.error("Service error:", error);
    res.status(400).json({ error: error.message });
  }
});

app.get("/services/users/:id/stats", async (req, res) => {
  try {
    const userWithStats =
      await globalState.services.UserService.getUserWithStats(req.params.id);
    res.json(userWithStats);
  } catch (error) {
    console.error("Service error:", error);
    res.status(404).json({ error: error.message });
  }
});
```

## Complete Examples

### Blog Application

```javascript
// Blog application setup
if (!globalState.blog) {
  globalState.blog = {
    initialized: false,
    config: {
      postsPerPage: 10,
      allowComments: true,
    },
  };
}

// Initialize blog database
if (!globalState.blog.initialized) {
  // Create blog tables
  db.query(`
        CREATE TABLE IF NOT EXISTS blog_posts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            slug TEXT UNIQUE NOT NULL,
            content TEXT NOT NULL,
            excerpt TEXT,
            author_id INTEGER NOT NULL,
            status TEXT DEFAULT 'draft',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            published_at DATETIME
        )
    `);

  db.query(`
        CREATE TABLE IF NOT EXISTS blog_comments (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id INTEGER NOT NULL,
            author_name TEXT NOT NULL,
            author_email TEXT NOT NULL,
            content TEXT NOT NULL,
            status TEXT DEFAULT 'pending',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (post_id) REFERENCES blog_posts (id)
        )
    `);

  globalState.blog.initialized = true;
  console.log("Blog database initialized");
}

// Blog post endpoints
app.get("/blog/posts", (req, res) => {
  const page = parseInt(req.query.page) || 1;
  const limit = globalState.blog.config.postsPerPage;
  const offset = (page - 1) * limit;

  try {
    const posts = db.query(
      `
            SELECT 
                p.id, p.title, p.slug, p.excerpt, p.status,
                p.created_at, p.published_at,
                u.name as author_name
            FROM blog_posts p
            JOIN users u ON p.author_id = u.id
            WHERE p.status = 'published'
            ORDER BY p.published_at DESC
            LIMIT ? OFFSET ?
        `,
      [limit, offset]
    );

    const countResult = db.query(`
            SELECT COUNT(*) as total 
            FROM blog_posts 
            WHERE status = 'published'
        `);

    res.json({
      posts,
      pagination: {
        page,
        limit,
        total: countResult[0].total,
        hasNext: posts.length === limit,
      },
    });
  } catch (error) {
    console.error("Blog posts error:", error);
    res.status(500).json({ error: "Failed to fetch blog posts" });
  }
});

app.get("/blog/posts/:slug", (req, res) => {
  try {
    const posts = db.query(
      `
            SELECT 
                p.id, p.title, p.slug, p.content, p.excerpt,
                p.created_at, p.published_at,
                u.name as author_name, u.email as author_email
            FROM blog_posts p
            JOIN users u ON p.author_id = u.id
            WHERE p.slug = ? AND p.status = 'published'
        `,
      [req.params.slug]
    );

    if (posts.length === 0) {
      return res.status(404).json({ error: "Post not found" });
    }

    const post = posts[0];

    // Get comments if enabled
    let comments = [];
    if (globalState.blog.config.allowComments) {
      comments = db.query(
        `
                SELECT id, author_name, content, created_at
                FROM blog_comments
                WHERE post_id = ? AND status = 'approved'
                ORDER BY created_at ASC
            `,
        [post.id]
      );
    }

    res.json({
      post,
      comments,
    });
  } catch (error) {
    console.error("Blog post error:", error);
    res.status(500).json({ error: "Failed to fetch blog post" });
  }
});

// Create new blog post
app.post("/blog/posts", (req, res) => {
  const { title, content, excerpt, author_id } = req.body;

  if (!title || !content || !author_id) {
    return res.status(400).json({
      error: "Title, content, and author_id are required",
    });
  }

  // Generate slug from title
  const slug = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");

  try {
    db.query(
      `
            INSERT INTO blog_posts (title, slug, content, excerpt, author_id, status)
            VALUES (?, ?, ?, ?, ?, 'draft')
        `,
      [title, slug, content, excerpt || "", author_id]
    );

    const newPosts = db.query("SELECT * FROM blog_posts WHERE slug = ?", [
      slug,
    ]);

    res.status(201).json(newPosts[0]);
  } catch (error) {
    console.error("Create post error:", error);
    if (error.message.includes("UNIQUE constraint failed")) {
      res.status(400).json({ error: "A post with this title already exists" });
    } else {
      res.status(500).json({ error: "Failed to create blog post" });
    }
  }
});

// Publish blog post
app.put("/blog/posts/:id/publish", (req, res) => {
  const postId = req.params.id;

  try {
    db.query(
      `
            UPDATE blog_posts 
            SET status = 'published', published_at = CURRENT_TIMESTAMP
            WHERE id = ?
        `,
      [postId]
    );

    const updatedPosts = db.query("SELECT * FROM blog_posts WHERE id = ?", [
      postId,
    ]);

    if (updatedPosts.length === 0) {
      return res.status(404).json({ error: "Post not found" });
    }

    res.json(updatedPosts[0]);
  } catch (error) {
    console.error("Publish post error:", error);
    res.status(500).json({ error: "Failed to publish post" });
  }
});
```

### REST API with Authentication

```javascript
// API configuration
if (!globalState.api) {
  globalState.api = {
    version: "1.0",
    rateLimit: {
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 100, // limit each IP to 100 requests per windowMs
    },
    requests: new Map(), // IP -> { count, resetTime }
  };
}

// Rate limiting middleware
function rateLimit(req, res, next) {
  const ip = req.ip;
  const now = Date.now();
  const windowMs = globalState.api.rateLimit.windowMs;
  const max = globalState.api.rateLimit.max;

  if (!globalState.api.requests.has(ip)) {
    globalState.api.requests.set(ip, {
      count: 1,
      resetTime: now + windowMs,
    });
    return next();
  }

  const record = globalState.api.requests.get(ip);

  if (now > record.resetTime) {
    // Reset window
    record.count = 1;
    record.resetTime = now + windowMs;
    return next();
  }

  if (record.count >= max) {
    return res.status(429).json({
      error: "Too many requests",
      retryAfter: Math.ceil((record.resetTime - now) / 1000),
    });
  }

  record.count++;
  next();
}

// Authentication middleware
function authenticate(req, res, next) {
  const token = req.headers.authorization?.replace("Bearer ", "");

  if (!token) {
    return res.status(401).json({ error: "Authentication token required" });
  }

  // Simple token validation (use proper JWT in production)
  if (!globalState.sessions || !globalState.sessions.has(token)) {
    return res.status(401).json({ error: "Invalid or expired token" });
  }

  req.session = globalState.sessions.get(token);
  next();
}

// API versioning
const apiV1 = "/api/v1";

// Public endpoints
app.get(apiV1 + "/health", rateLimit, (req, res) => {
  res.json({
    status: "healthy",
    version: globalState.api.version,
    timestamp: new Date().toISOString(),
  });
});

// Authentication endpoint
app.post(apiV1 + "/auth/login", rateLimit, (req, res) => {
  const { email, password } = req.body;

  if (!email || !password) {
    return res.status(400).json({ error: "Email and password required" });
  }

  try {
    // Validate credentials (use proper password hashing)
    const users = db.query(
      "SELECT id, name, email FROM users WHERE email = ? AND password = ?",
      [email, password]
    );

    if (users.length === 0) {
      return res.status(401).json({ error: "Invalid credentials" });
    }

    // Generate token
    const token =
      Math.random().toString(36).substring(2, 15) +
      Math.random().toString(36).substring(2, 15);

    // Store session
    if (!globalState.sessions) {
      globalState.sessions = new Map();
    }

    globalState.sessions.set(token, {
      userId: users[0].id,
      user: users[0],
      createdAt: new Date(),
    });

    res.json({
      token,
      user: users[0],
      expiresIn: 3600, // 1 hour
    });
  } catch (error) {
    console.error("Login error:", error);
    res.status(500).json({ error: "Authentication failed" });
  }
});

// Protected endpoints
app.get(apiV1 + "/profile", rateLimit, authenticate, (req, res) => {
  res.json({
    user: req.session.user,
    sessionInfo: {
      createdAt: req.session.createdAt,
    },
  });
});

app.get(apiV1 + "/users", rateLimit, authenticate, (req, res) => {
  const page = parseInt(req.query.page) || 1;
  const limit = Math.min(parseInt(req.query.limit) || 20, 100);
  const offset = (page - 1) * limit;

  try {
    const users = db.query(
      `
            SELECT id, name, email, created_at, is_active
            FROM users
            ORDER BY created_at DESC
            LIMIT ? OFFSET ?
        `,
      [limit, offset]
    );

    const countResult = db.query("SELECT COUNT(*) as total FROM users");

    res.json({
      data: users,
      meta: {
        page,
        limit,
        total: countResult[0].total,
        pages: Math.ceil(countResult[0].total / limit),
      },
    });
  } catch (error) {
    console.error("API users error:", error);
    res.status(500).json({ error: "Failed to fetch users" });
  }
});
```

## Best Practices

### Code Organization

- **Modular Structure**: Organize your code into logical modules using the global state object
- **Error Handling**: Always wrap database operations and external calls in try-catch blocks
- **Input Validation**: Validate all user inputs before processing
- **Response Consistency**: Use consistent response formats across your API endpoints

### Security Considerations

- **Parameter Binding**: Always use parameterized queries for database operations
- **Input Sanitization**: Validate and sanitize all user inputs
- **Rate Limiting**: Implement rate limiting for public endpoints
- **Authentication**: Use proper authentication mechanisms for protected resources
- **HTTPS**: Always use HTTPS in production environments

### Performance Optimization

- **Database Indexing**: Create appropriate database indexes for frequently queried columns
- **Caching**: Use the global state object to cache frequently accessed data
- **Pagination**: Implement pagination for endpoints that return large datasets
- **Query Optimization**: Write efficient SQL queries and avoid N+1 query problems

### Development Workflow

- **Incremental Development**: Build and test your application incrementally
- **Logging**: Use console logging extensively for debugging
- **Error Monitoring**: Monitor error logs and application metrics
- **Testing**: Test your endpoints thoroughly with various inputs and edge cases

## Troubleshooting

### Common Issues

**Database Errors**: Most database errors are due to SQL syntax issues or constraint violations. Check your SQL queries and ensure proper parameter binding.

**Memory Leaks**: Avoid storing large objects in the global state without proper cleanup. Use Map objects for dynamic collections that need cleanup.

**Performance Issues**: Monitor your application's performance and optimize database queries and caching strategies as needed.

**Authentication Problems**: Ensure your authentication tokens are properly generated and validated. Check session storage and expiration logic.

### Debugging Techniques

- Use `console.log()` extensively to track request flow and data transformations
- Monitor the global state object to understand application state changes
- Check database query results and parameters to ensure correct data handling
- Test endpoints with various input combinations to identify edge cases

This comprehensive guide provides everything needed to build sophisticated JavaScript applications in the sandboxed environment. The Express.js compatible API, integrated database access, and persistent global state enable rapid development of production-ready web applications while maintaining security through isolation.
