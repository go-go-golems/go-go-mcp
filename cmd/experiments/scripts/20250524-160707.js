// HTML page endpoints

registerHandler("GET", "/admin", () => {
    return html(`<!DOCTYPE html>
<html>
<head>
    <title>JS Playground Admin</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; }
        .endpoint { background: #e8f4f8; padding: 10px; margin: 5px 0; border-radius: 3px; }
        .method { font-weight: bold; color: #0066cc; }
    </style>
</head>
<body>
    <div class="header">
        <h1>JavaScript Playground Admin</h1>
        <p>Server running at: ${new Date().toISOString()}</p>
    </div>
    
    <div class="section">
        <h2>Available Endpoints</h2>
        <div class="endpoint"><span class="method">GET</span> /health - Health check</div>
        <div class="endpoint"><span class="method">GET</span> /api/status - API status</div>
        <div class="endpoint"><span class="method">GET</span> /api/users - List users</div>
        <div class="endpoint"><span class="method">POST</span> /api/echo - Echo request</div>
        <div class="endpoint"><span class="method">POST</span> /v1/execute - Execute JavaScript</div>
    </div>
    
    <div class="section">
        <h2>Quick Actions</h2>
        <button onclick="fetch('/health').then(r=>r.json()).then(d=>alert(JSON.stringify(d)))">Test Health</button>
        <button onclick="fetch('/api/status').then(r=>r.json()).then(d=>alert(JSON.stringify(d)))">API Status</button>
        <button onclick="fetch('/api/users').then(r=>r.json()).then(d=>alert(JSON.stringify(d)))">List Users</button>
    </div>
</body>
</html>`);
});

registerHandler("GET", "/docs", () => {
    return html(`<!DOCTYPE html>
<html>
<head>
    <title>JS Playground Documentation</title>
    <style>
        body { font-family: monospace; margin: 40px; line-height: 1.6; }
        .code { background: #f4f4f4; padding: 15px; margin: 10px 0; border-radius: 5px; }
        h1, h2 { color: #333; }
    </style>
</head>
<body>
    <h1>JavaScript Playground Documentation</h1>
    
    <h2>Available Functions</h2>
    
    <h3>registerHandler(method, path, function)</h3>
    <p>Register an HTTP endpoint:</p>
    <div class="code">
registerHandler("GET", "/myendpoint", () => ({message: "Hello!"}));
    </div>
    
    <h3>db.query(sql, ...params)</h3>
    <p>Execute SQL queries:</p>
    <div class="code">
const users = db.query("SELECT * FROM users WHERE id = ?", [1]);
    </div>
    
    <h3>console.log(...args)</h3>
    <p>Log to server console:</p>
    <div class="code">
console.log("Debug message", {data: value});
    </div>
    
    <h2>Example Code</h2>
    <div class="code">
// Create a simple API
registerHandler("POST", "/greet", (req) => {
    const name = req.query.name || "World";
    return {greeting: \`Hello, \${name}!\`};
});

// Database interaction
const results = db.query("SELECT COUNT(*) as total FROM users");
console.log("Total users:", results[0].total);
    </div>
</body>
</html>`);
});

console.log("Page endpoints registered");