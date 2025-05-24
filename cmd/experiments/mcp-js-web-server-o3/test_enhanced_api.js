// Test script for enhanced HTTP API features
// Note: Uses global scope execution - safe for re-execution

console.log("Setting up enhanced API test endpoints...");

// Test basic backward compatibility
registerHandler("GET", "/test/simple", () => "Simple response works!");

// Test enhanced response object
registerHandler("GET", "/test/enhanced", (req) => ({
    status: 200,
    headers: {
        "x-test": "enhanced-api",
        "x-request-ip": req.remoteIP
    },
    body: {
        message: "Enhanced API working!",
        requestInfo: {
            method: req.method,
            path: req.path,
            query: req.query,
            userAgent: req.headers["user-agent"] || "unknown",
            cookies: req.cookies
        }
    }
}));

// Test Response helper functions
registerHandler("GET", "/test/json", () => 
    Response.json({ 
        test: "JSON helper works", 
        timestamp: new Date().toISOString() 
    })
);

registerHandler("GET", "/test/text", () => 
    Response.text("Plain text helper works!")
);

registerHandler("GET", "/test/html", () => 
    Response.html(`
        <!DOCTYPE html>
        <html>
        <head><title>Test Page</title></head>
        <body>
            <h1>HTML Helper Works!</h1>
            <p>Server time: ${new Date().toISOString()}</p>
        </body>
        </html>
    `)
);

// Test redirect
registerHandler("GET", "/test/redirect", () => 
    Response.redirect("/test/enhanced", HTTP.MOVED_PERMANENTLY)
);

// Test error responses
registerHandler("GET", "/test/error", () => 
    Response.error("This is a test error", HTTP.BAD_REQUEST)
);

// Test cookie setting
registerHandler("GET", "/test/cookie", () => ({
    body: "Cookie set! Check your browser.",
    cookies: [{
        name: "test-cookie",
        value: "test-value-" + Date.now(),
        path: "/",
        maxAge: 3600
    }]
}));

// Test status codes
registerHandler("POST", "/test/created", () => 
    Response.json({ message: "Resource created" }, HTTP.CREATED)
);

// Test advanced request parsing
registerHandler("POST", "/test/echo", (req) => ({
    status: HTTP.OK,
    body: {
        echo: "Request received",
        method: req.method,
        contentType: req.headers["content-type"],
        body: req.body,
        queryParams: req.query,
        clientIP: req.remoteIP
    },
    headers: {
        "x-echo-response": "true"
    }
}));

console.log("Enhanced API test endpoints registered!");
console.log("Available test endpoints:");
console.log("  GET  /test/simple     - Basic response");
console.log("  GET  /test/enhanced   - Enhanced response object");
console.log("  GET  /test/json       - Response.json() helper");
console.log("  GET  /test/text       - Response.text() helper");
console.log("  GET  /test/html       - Response.html() helper");
console.log("  GET  /test/redirect   - Response.redirect() helper");
console.log("  GET  /test/error      - Response.error() helper");
console.log("  GET  /test/cookie     - Cookie setting");
console.log("  POST /test/created    - HTTP.CREATED status");
console.log("  POST /test/echo       - Request echo with enhanced parsing");

return "Enhanced API test setup complete!";