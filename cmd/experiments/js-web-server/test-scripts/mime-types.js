// Test different MIME types

// JSON endpoint with explicit content type
registerHandler("GET", "/api/json", () => ({
    message: "This is JSON",
    timestamp: new Date().toISOString()
}), "application/json");

// XML endpoint
registerHandler("GET", "/api/xml", () => {
    return `<?xml version="1.0" encoding="UTF-8"?>
<response>
    <message>This is XML</message>
    <timestamp>${new Date().toISOString()}</timestamp>
</response>`;
}, "application/xml");

// Plain text endpoint
registerHandler("GET", "/api/text", () => {
    return "This is plain text content";
}, "text/plain");

// CSV endpoint
registerHandler("GET", "/api/csv", () => {
    return `name,email,age
Alice,alice@example.com,25
Bob,bob@example.com,30
Charlie,charlie@example.com,35`;
}, "text/csv");

// CSS endpoint
registerHandler("GET", "/styles.css", () => {
    return `body {
    font-family: Arial, sans-serif;
    background-color: #f0f0f0;
    margin: 0;
    padding: 20px;
}

.container {
    max-width: 800px;
    margin: 0 auto;
    background: white;
    padding: 20px;
    border-radius: 5px;
    box-shadow: 0 2px 5px rgba(0,0,0,0.1);
}

h1 {
    color: #333;
    border-bottom: 2px solid #007acc;
    padding-bottom: 10px;
}

.endpoint {
    background: #f8f8f8;
    border-left: 4px solid #007acc;
    padding: 10px;
    margin: 10px 0;
}`;
}, "text/css");

// JavaScript endpoint
registerHandler("GET", "/script.js", () => {
    return `// Dynamic JavaScript
console.log("This JavaScript was served by the playground!");

function testFunction() {
    fetch('/api/json')
        .then(response => response.json())
        .then(data => {
            console.log('API Response:', data);
            document.body.innerHTML += '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
        });
}

// Auto-run the test
if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', testFunction);
}`;
}, "application/javascript");

// SVG endpoint
registerHandler("GET", "/logo.svg", () => {
    return `<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg">
    <circle cx="50" cy="50" r="40" stroke="#007acc" stroke-width="3" fill="#e8f4f8"/>
    <text x="50" y="55" font-family="Arial" font-size="14" text-anchor="middle" fill="#333">JS</text>
</svg>`;
}, "image/svg+xml");

// HTML page that uses all the above endpoints
registerHandler("GET", "/mime-test", () => {
    return `<!DOCTYPE html>
<html>
<head>
    <title>MIME Type Test Page</title>
    <link rel="stylesheet" href="/styles.css">
</head>
<body>
    <div class="container">
        <h1>MIME Type Test Page</h1>
        <img src="/logo.svg" alt="JS Logo">
        
        <h2>Available Endpoints with Different MIME Types:</h2>
        
        <div class="endpoint">
            <strong>JSON:</strong> <a href="/api/json">/api/json</a> (application/json)
        </div>
        
        <div class="endpoint">
            <strong>XML:</strong> <a href="/api/xml">/api/xml</a> (application/xml)
        </div>
        
        <div class="endpoint">
            <strong>Plain Text:</strong> <a href="/api/text">/api/text</a> (text/plain)
        </div>
        
        <div class="endpoint">
            <strong>CSV:</strong> <a href="/api/csv">/api/csv</a> (text/csv)
        </div>
        
        <div class="endpoint">
            <strong>CSS:</strong> <a href="/styles.css">/styles.css</a> (text/css)
        </div>
        
        <div class="endpoint">
            <strong>JavaScript:</strong> <a href="/script.js">/script.js</a> (application/javascript)
        </div>
        
        <div class="endpoint">
            <strong>SVG:</strong> <a href="/logo.svg">/logo.svg</a> (image/svg+xml)
        </div>
        
        <h2>Test Results:</h2>
        <div id="results"></div>
    </div>
    
    <script src="/script.js"></script>
</body>
</html>`;
}, "text/html");

console.log("MIME type test endpoints registered");