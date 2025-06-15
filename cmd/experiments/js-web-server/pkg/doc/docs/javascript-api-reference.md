# JavaScript API Reference

## Overview

The JavaScript Playground Server provides a runtime environment where JavaScript code executes in a **persistent global context**. Code runs at the top level without function wrapping, enabling dynamic web application creation with database integration and AI capabilities through Geppetto.

**Key Features:**
- **Express.js-compatible API** - Familiar routing and middleware patterns
- **SQLite database access** - Direct database operations via `db` object
- **Persistent state** - `globalState` object survives across executions
- **Real-time endpoint registration** - Add routes dynamically without restarts
- **AI Integration** - Full Geppetto JavaScript API for conversations, embeddings, steps, and chat

**Execution Context:**
- Code runs in **global scope** (no function wrapping)
- **No `return` statements** - Last expression is automatically returned
- **Function definitions** - Use `function name() { }` syntax for reusable functions
- **Variable scoping** - Wrap `const`/`let` in IIFE to avoid global pollution
- **Global variables** - Use `globalState` object for persistent data
- **Persistent runtime** - Functions and global state remain between executions
- **AI Operations** - Async/await support for AI interactions through event loop integration

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

---

# Geppetto AI API Reference

The js-web-server includes full integration with Geppetto's JavaScript API, providing powerful AI capabilities including conversations, embeddings, steps, and chat functionality. This section provides comprehensive documentation for using these AI features in your JavaScript applications.

## AI API Overview

Geppetto's JavaScript API exposes four main components:

1. **Conversation API** (`Conversation`): Create and manage conversations with messages, tool uses, and tool results
2. **Embeddings API** (`embeddings`): Generate vector embeddings from text using various embedding models  
3. **Steps API** (`steps`): Execute asynchronous computation steps with streaming, cancellation, and composition
4. **Chat Step Factory** (`ChatStepFactory`): Create and manage chat completion steps for AI interactions

These APIs provide both synchronous and asynchronous interfaces, with full support for streaming, error handling, and cancellation patterns.

## Quick AI Start

```javascript
// Create a simple AI chat endpoint
const factory = new ChatStepFactory();
const chatStep = factory.newStep();

app.post('/ai/chat', async (req, res) => {
  try {
    const { message } = req.body;
    
    // Create conversation
    const conv = new Conversation();
    conv.AddMessage("system", "You are a helpful assistant");
    conv.AddMessage("user", message);
    
    // Get AI response
    const response = await chatStep.startAsync(conv);
    res.json({ response });
  } catch (error) {
    console.error('AI chat error:', error);
    res.status(500).json({ error: 'Failed to process chat' });
  }
});
```

## Conversation API

The Conversation API provides a JavaScript interface for creating and managing conversations with messages, tool uses, and tool results.

### Creating and Managing Conversations

```javascript
// Create a new conversation
const conv = new Conversation();

// Add a simple chat message
const msgId = conv.AddMessage("user", "Hello, how can I help you?");

// Add a message with options
const msgWithOptions = conv.AddMessage("system", "System prompt", {
    metadata: { source: "config" },
    parentID: "parent-message-id",
    time: "2024-01-01T00:00:00Z",
    id: "custom-id"  // optional, will generate UUID if not provided
});

// Add a message with an image
const msgWithImage = conv.AddMessageWithImage(
    "user",
    "Here's an image",
    "/path/to/image.jpg"  // supports local files and URLs
);
```

### Message Options

The `MessageOptions` interface provides flexible configuration:

```typescript
interface MessageOptions {
    metadata?: Record<string, any>;  // Additional metadata
    parentID?: string;               // Parent message ID
    time?: string;                   // RFC3339 format timestamp
    id?: string;                     // Custom message ID
}
```

### Tool Integration

The Conversation API supports tool use and tool result messages for AI function calling:

```javascript
// Add a tool use
const toolUseId = conv.AddToolUse(
    "tool123",
    "searchCode",
    { query: "find main function" }
);

// Add a tool result
const resultId = conv.AddToolResult(
    "tool123",
    "Found main function in main.go"
);
```

### Working with Messages

```javascript
// Get all messages
const messages = conv.GetMessages();
// Returns an array of message objects

// Get formatted view of a specific message
const messageView = conv.GetMessageView(msgId);
// Returns formatted string based on message type:
// - Chat: "[role]: text"
// - Tool Use: "ToolUseContent{...}"
// - Tool Result: "ToolResultContent{...}"

// Update message metadata
conv.UpdateMetadata(msgId, { processed: true });

// Get conversation as a single prompt string
const prompt = conv.GetSinglePrompt();

// Convert back to Go conversation object
const goConv = conv.toGoConversation();
```

### Message Object Structure

Messages returned by `GetMessages()` have different structures based on their type:

#### Common Fields
All message objects include:
```javascript
{
    id: string,          // Unique message ID
    parentID: string,    // Parent message ID
    time: Date,          // Creation timestamp
    lastUpdate: Date,    // Last update timestamp
    metadata: object,    // Additional metadata
    type: string        // Message type: "chat-message", "tool-use", or "tool-result"
}
```

#### Chat Message (type: "chat-message")
```javascript
{
    ...commonFields,
    role: string,        // "system", "assistant", "user", or "tool"
    text: string,        // Message content
    images?: [{          // Optional array of images
        imageURL: string,
        imageName: string,
        mediaType: string,
        detail: string
    }]
}
```

#### Tool Use (type: "tool-use")
```javascript
{
    ...commonFields,
    toolID: string,      // Tool identifier
    name: string,        // Tool name
    input: object,       // Tool input parameters
    toolType: string     // Tool type (e.g., "function")
}
```

#### Tool Result (type: "tool-result")
```javascript
{
    ...commonFields,
    toolID: string,      // Tool identifier
    result: string       // Tool execution result
}
```

### Image Support

The conversation API supports adding images to messages:

```javascript
// Add message with image
const msgWithImage = conv.AddMessageWithImage(
    "user",
    "What's in this image?",
    "/path/to/image.jpg"
);
```

**Supported formats**: PNG, JPEG, WebP, and GIF  
**Maximum file size**: 20MB  
**Sources**: Local file paths and URLs  
**Constraints**: Images are automatically validated for format and size

## Embeddings API

The Embeddings API provides JavaScript bindings for generating vector embeddings from text using various embedding models.

### Core Concepts

Embeddings are vector representations of text that capture semantic meaning in a high-dimensional space. They're useful for:
- Semantic search and similarity comparison
- Document clustering and classification
- Information retrieval systems
- Machine learning features

### Model Information

Each embeddings provider exposes information about its model:

```javascript
const model = embeddings.getModel();
// Returns: { name: string, dimensions: number }
console.log("Using model:", model.name);
console.log("Vector dimensions:", model.dimensions);
```

### Synchronous API

For simple, blocking operations:

```javascript
const text = "Hello, world!";
try {
    const embedding = embeddings.generateEmbedding(text);
    // Returns: Float32Array of dimensions length
    console.log("Embedding dimensions:", embedding.length);
} catch (err) {
    console.error("Failed to generate embedding:", err);
}
```

### Asynchronous Promise API

Promise-based API for better error handling and non-blocking operations:

```javascript
async function generateEmbedding(text) {
    try {
        const embedding = await embeddings.generateEmbeddingAsync(text);
        console.log("Embedding dimensions:", embedding.length);
        return embedding;
    } catch (err) {
        console.error("Failed to generate embedding:", err);
        throw err;
    }
}

// Usage
const embedding = await generateEmbedding("Hello, world!");
```

### Callback API with Cancellation

For operations that need cancellation support:

```javascript
const text = "Hello, world!";
const cancel = embeddings.generateEmbeddingWithCallbacks(text, {
    onSuccess: (embedding) => {
        console.log("Embedding generated:", embedding);
    },
    onError: (err) => {
        console.error("Error:", err);
    }
});

// Cancel the operation if needed
setTimeout(() => {
    cancel();
}, 5000);
```

### Semantic Search Example

Implementing semantic search using embeddings:

```javascript
// Function to compute cosine similarity between vectors
function cosineSimilarity(a, b) {
    let dotProduct = 0;
    let normA = 0;
    let normB = 0;
    
    for (let i = 0; i < a.length; i++) {
        dotProduct += a[i] * b[i];
        normA += a[i] * a[i];
        normB += b[i] * b[i];
    }
    
    return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
}

// Create semantic search endpoint
app.post('/search/semantic', async (req, res) => {
    try {
        const { query, documents } = req.body;
        
        // Generate query embedding
        const queryEmbedding = await embeddings.generateEmbeddingAsync(query);
        
        // Generate document embeddings
        const documentEmbeddings = await Promise.all(
            documents.map(doc => embeddings.generateEmbeddingAsync(doc))
        );
        
        // Calculate similarities
        const similarities = documentEmbeddings.map(docEmb => 
            cosineSimilarity(queryEmbedding, docEmb)
        );
        
        // Find best match
        const bestMatchIndex = similarities.indexOf(Math.max(...similarities));
        res.json({
            document: documents[bestMatchIndex],
            similarity: similarities[bestMatchIndex],
            allSimilarities: similarities
        });
    } catch (err) {
        console.error("Semantic search failed:", err);
        res.status(500).json({ error: "Search failed" });
    }
});
```

## Steps API

The Steps API provides JavaScript access to Geppetto's step abstraction - a powerful system for asynchronous computation that combines features of async operations and list monads.

### Core Step Concepts

A Step represents a computation that:

1. **Takes a single input** and produces zero or more outputs asynchronously
2. **Can be cancelled** at any point during execution
3. **Can be composed** with other steps to create pipelines
4. **Supports streaming results** for real-time feedback
5. **Carries metadata** about its execution

### Step Execution APIs

Each registered step provides three execution methods:

#### Promise-based API
Best for single-result operations or when you want to wait for all results:

```javascript
// Async/await style
try {
    const promise = step.startAsync(input);
    console.log("Promise created");
    const results = await promise;
    console.log("Results:", results);
} catch (err) {
    console.error("Error:", err);
}
```

#### Synchronous API
Use when you need blocking behavior and have all results immediately:

```javascript
try {
    const results = step.startBlocking(input);
    console.log("Results:", results);
} catch (err) {
    console.error("Error:", err);
}
```

#### Callback-based Streaming API
Best for handling streaming results or long-running operations:

```javascript
const cancel = step.startWithCallbacks(input, {
    onResult: (result) => {
        console.log("Got result:", result);
    },
    onError: (err) => {
        console.error("Error occurred:", err);
    },
    onDone: () => {
        console.log("Processing complete");
    },
    onCancel: () => {
        console.log("Operation cancelled");
    }
});

// Cancel the operation when needed
setTimeout(() => {
    cancel();
}, 5000);
```

### Cancellation Support

All step operations support cancellation:

```javascript
// With callbacks
const cancel = step.startWithCallbacks(input, callbacks);
// Later...
cancel();

// With promises using AbortController
const controller = new AbortController();
const promise = step.startAsync(input, { signal: controller.signal });
// Later...
controller.abort();
```

## Chat Step Factory

The Chat Step Factory provides a specialized interface for creating chat completion steps that integrate with various LLM providers.

### Basic Usage

```javascript
// Create a factory instance
const factory = new ChatStepFactory();

// Create a new chat step
const step = factory.newStep();

// Use Promise-based API
step.startAsync({ 
    messages: [
        { role: "user", content: "Hello, how can I help you?" }
    ]
})
.then(result => {
    console.log("Response:", result);
})
.catch(err => {
    console.error("Error:", err);
});
```

### Streaming Chat Responses

Chat steps excel at streaming responses for real-time output:

```javascript
step.startWithCallbacks(
    { 
        messages: [
            { role: "user", content: "Explain quantum computing" }
        ]
    },
    {
        onResult: (result) => {
            console.log("Got chunk:", result);
            // Display streaming text in UI
        },
        onError: (err) => {
            console.error("Error occurred:", err);
        },
        onDone: () => {
            console.log("Chat complete");
        }
    }
);
```

### Conversation Integration

The Chat Step Factory supports two input formats:

#### Using Conversation Objects (Recommended)
```javascript
const conv = new Conversation();

// Add messages with full conversation management
conv.AddMessage("system", "You are a helpful assistant");
conv.AddMessage("user", "What is quantum computing?");

// Add messages with metadata
conv.AddMessage("user", "Hello", {
    metadata: { source: "user-input" },
    time: "2024-03-20T15:04:05Z"
});

// Add messages with images
conv.AddMessageWithImage("user", "What's in this image?", "path/to/image.jpg");

// Add tool usage
conv.AddToolUse("tool123", "search", { query: "quantum computing" });
conv.AddToolResult("tool123", "Found relevant articles...");

// Use with chat step
const response = await step.startAsync(conv);
```

#### Legacy Format (Backward Compatibility)
```javascript
const input = {
    messages: [
        { role: "system", content: "You are a helpful assistant" },
        { role: "user", content: "What is quantum computing?" }
    ]
};

const response = await step.startAsync(input);
```

### Complete Chat Application Example

```javascript
// Create factory and step
const factory = new ChatStepFactory();
const chatStep = factory.newStep();

// Create and setup conversation
const conversation = new Conversation();
conversation.AddMessage("system", 
    "You are a helpful AI assistant. Be concise and clear in your responses."
);

// Create a stateful chat endpoint
app.post('/ai/chat', async (req, res) => {
    try {
        const { message, sessionId } = req.body;
        
        // Get or create conversation for session
        if (!globalState.conversations) {
            globalState.conversations = new Map();
        }
        
        let conv = globalState.conversations.get(sessionId);
        if (!conv) {
            conv = new Conversation();
            conv.AddMessage("system", "You are a helpful assistant");
            globalState.conversations.set(sessionId, conv);
        }
        
        // Add user message
        conv.AddMessage("user", message);
        
        // Get AI response
        const response = await chatStep.startAsync(conv);
        
        // Add assistant response to conversation
        conv.AddMessage("assistant", response);
        
        res.json({ response, sessionId });
    } catch (error) {
        console.error('Chat error:', error);
        res.status(500).json({ error: 'Failed to process chat' });
    }
});

// Streaming chat endpoint
app.post('/ai/stream-chat', (req, res) => {
    const { message, sessionId } = req.body;
    
    // Set up Server-Sent Events
    res.writeHead(200, {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        'Connection': 'keep-alive'
    });
    
    // Get conversation
    let conv = globalState.conversations.get(sessionId) || new Conversation();
    conv.AddMessage("user", message);
    
    let fullResponse = "";
    
    // Stream response
    const cancel = chatStep.startWithCallbacks(conv, {
        onResult: (chunk) => {
            fullResponse += chunk;
            res.write(`data: ${JSON.stringify({ chunk })}\n\n`);
        },
        onError: (err) => {
            res.write(`data: ${JSON.stringify({ error: err })}\n\n`);
            res.end();
        },
        onDone: () => {
            conv.AddMessage("assistant", fullResponse);
            globalState.conversations.set(sessionId, conv);
            res.write(`data: ${JSON.stringify({ done: true })}\n\n`);
            res.end();
        }
    });
    
    // Handle client disconnect
    req.on('close', () => {
        cancel();
    });
});
```

### Error Handling Best Practices

```javascript
// With callbacks
chatStep.startWithCallbacks(conversation, {
    onResult: (chunk) => { /* ... */ },
    onError: (err) => {
        console.error("LLM error:", err);
        // Handle specific error cases
        if (err.includes("rate limit")) {
            // Handle rate limiting
        } else if (err.includes("context length")) {
            // Handle context length errors
        }
    }
});

// With promises
try {
    await chatStep.startAsync(conversation);
} catch (err) {
    if (err.includes("context length")) {
        // Handle context length errors
    } else if (err.includes("invalid api key")) {
        // Handle authentication errors
    }
}
```

## AI Integration Examples

### Smart Content Generation

```javascript
// Create an AI-powered content generation endpoint
app.post('/ai/generate-content', async (req, res) => {
    try {
        const { topic, style, length } = req.body;
        
        const conv = new Conversation();
        conv.AddMessage("system", `You are a professional content writer. Generate ${style} content about ${topic}. Keep it approximately ${length} words.`);
        conv.AddMessage("user", `Generate content about: ${topic}`);
        
        const factory = new ChatStepFactory();
        const step = factory.newStep();
        
        const content = await step.startAsync(conv);
        
        // Store in database
        db.query('INSERT INTO generated_content (topic, style, content, created_at) VALUES (?, ?, ?, ?)', 
            [topic, style, content, new Date().toISOString()]);
        
        res.json({ content, topic, style });
    } catch (error) {
        console.error('Content generation error:', error);
        res.status(500).json({ error: 'Failed to generate content' });
    }
});
```

### AI-Powered Document Analysis

```javascript
// Document analysis with embeddings and chat
app.post('/ai/analyze-document', async (req, res) => {
    try {
        const { document, question } = req.body;
        
        // Generate document embedding for future similarity searches
        const docEmbedding = await embeddings.generateEmbeddingAsync(document);
        
        // Store document and embedding
        const docId = db.query('INSERT INTO documents (content, embedding) VALUES (?, ?)', 
            [document, JSON.stringify(Array.from(docEmbedding))]).lastInsertId;
        
        // Use chat to analyze the document
        const conv = new Conversation();
        conv.AddMessage("system", "You are a document analysis expert. Analyze the provided document and answer questions about it.");
        conv.AddMessage("user", `Document: ${document}\n\nQuestion: ${question}`);
        
        const factory = new ChatStepFactory();
        const step = factory.newStep();
        
        const analysis = await step.startAsync(conv);
        
        res.json({ 
            analysis, 
            documentId: docId,
            question 
        });
    } catch (error) {
        console.error('Document analysis error:', error);
        res.status(500).json({ error: 'Failed to analyze document' });
    }
});
```

### Intelligent Search with RAG

```javascript
// RAG (Retrieval-Augmented Generation) search endpoint
app.post('/ai/rag-search', async (req, res) => {
    try {
        const { query } = req.body;
        
        // Generate query embedding
        const queryEmbedding = await embeddings.generateEmbeddingAsync(query);
        
        // Retrieve similar documents from database
        const documents = db.query('SELECT content FROM documents ORDER BY RANDOM() LIMIT 10');
        
        // Find most relevant documents using embeddings
        let bestDocs = [];
        for (const doc of documents) {
            const docEmbedding = await embeddings.generateEmbeddingAsync(doc.content);
            const similarity = cosineSimilarity(queryEmbedding, docEmbedding);
            bestDocs.push({ content: doc.content, similarity });
        }
        
        // Sort by similarity and take top 3
        bestDocs.sort((a, b) => b.similarity - a.similarity);
        const topDocs = bestDocs.slice(0, 3).map(d => d.content);
        
        // Use chat to generate answer based on retrieved documents
        const conv = new Conversation();
        conv.AddMessage("system", "You are a helpful assistant. Answer the user's question based on the provided context documents. If the documents don't contain relevant information, say so.");
        conv.AddMessage("user", `Context documents:\n${topDocs.join('\n---\n')}\n\nQuestion: ${query}`);
        
        const factory = new ChatStepFactory();
        const step = factory.newStep();
        
        const answer = await step.startAsync(conv);
        
        res.json({ 
            answer, 
            query,
            relevantDocs: topDocs.length,
            similarities: bestDocs.slice(0, 3).map(d => d.similarity)
        });
    } catch (error) {
        console.error('RAG search error:', error);
        res.status(500).json({ error: 'Failed to perform RAG search' });
    }
});

// Helper function for cosine similarity (same as defined earlier)
function cosineSimilarity(a, b) {
    let dotProduct = 0;
    let normA = 0;
    let normB = 0;
    
    for (let i = 0; i < a.length; i++) {
        dotProduct += a[i] * b[i];
        normA += a[i] * a[i];
        normB += b[i] * b[i];
    }
    
    return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
}
```

## AI API Best Practices

1. **Always handle errors** - AI operations can fail for various reasons (rate limits, context length, network issues)
2. **Use streaming for real-time UX** - Stream chat responses for better user experience
3. **Implement cancellation** - Allow users to cancel long-running AI operations
4. **Cache embeddings** - Store embeddings in database to avoid re-computation
5. **Manage conversation context** - Keep track of conversation history for multi-turn chats
6. **Rate limiting** - Implement rate limiting for AI endpoints to manage costs
7. **Error recovery** - Implement retry logic for transient failures
8. **Monitor costs** - Track API usage and costs for AI services
9. **Validate inputs** - Sanitize and validate user inputs before sending to AI
10. **Handle timeouts** - Set appropriate timeouts for AI operations

## Configuration

AI capabilities are configured through profile management. Common configurations include:

- **Model selection**: Choose between different LLM models (GPT-4, Claude, etc.)
- **API keys**: Configure credentials for different AI providers
- **Temperature settings**: Control randomness in AI responses
- **Token limits**: Set maximum token counts for responses
- **Timeout settings**: Configure request timeouts
- **Embedding models**: Select embedding providers and models

Use `go run ./cmd/experiments/js-web-server profiles list` to see available configurations. 