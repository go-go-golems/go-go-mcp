// Test Express.js functionality
console.log("Setting up Express.js test routes");

// Test basic JSON response
app.get("/test/json", (req, res) => {
    console.log("Handler called for /test/json");
    console.log("req type:", typeof req);
    console.log("res type:", typeof res);
    console.log("res.json type:", typeof res.json);
    
    try {
        res.json({ message: "Hello from Express.js!", timestamp: new Date().toISOString() });
    } catch (error) {
        console.error("Error in res.json:", error);
    }
});

// Test status chaining
app.get("/test/status", (req, res) => {
    console.log("Handler called for /test/status");
    console.log("res.status type:", typeof res.status);
    
    try {
        res.status(201).json({ created: true });
    } catch (error) {
        console.error("Error in status chaining:", error);
    }
});

// Test send method
app.get("/test/send", (req, res) => {
    console.log("Handler called for /test/send");
    console.log("res.send type:", typeof res.send);
    
    try {
        res.send("Hello from Express.js send method!");
    } catch (error) {
        console.error("Error in res.send:", error);
    }
});

// Test request object
app.get("/test/request", (req, res) => {
    console.log("Handler called for /test/request");
    console.log("Request object:", {
        method: req.method,
        path: req.path,
        url: req.url,
        query: req.query,
        headers: req.headers,
        ip: req.ip
    });
    
    res.json({
        method: req.method,
        path: req.path,
        url: req.url,
        query: req.query,
        userAgent: req.headers["user-agent"],
        ip: req.ip
    });
});

console.log("Express.js test routes registered"); 