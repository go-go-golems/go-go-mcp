// Utility endpoints and functions

// Global utilities
globalThis.formatDate = (date) => {
    return new Date(date).toISOString().split('T')[0];
};

globalThis.randomId = () => {
    return Math.random().toString(36).substr(2, 9);
};

// Metrics endpoint
let requestCount = 0;
const startTime = new Date();

registerHandler("GET", "/metrics", () => {
    return {
        uptime: Math.floor((new Date() - startTime) / 1000),
        requests: ++requestCount,
        memory: process.memoryUsage ? process.memoryUsage() : "Not available",
        timestamp: new Date().toISOString()
    };
});

// Random data generator
registerHandler("GET", "/random/user", () => {
    const names = ["Alice", "Bob", "Charlie", "Diana", "Eve", "Frank"];
    const domains = ["example.com", "test.org", "demo.net"];
    const name = names[Math.floor(Math.random() * names.length)];
    const domain = domains[Math.floor(Math.random() * domains.length)];
    
    return {
        id: randomId(),
        name: name,
        email: `${name.toLowerCase()}@${domain}`,
        created: formatDate(new Date())
    };
});

// Calculator endpoint
registerHandler("POST", "/calc", (req) => {
    try {
        const {operation, a, b} = req.query;
        let result;
        
        switch(operation) {
            case 'add': result = a + b; break;
            case 'sub': result = a - b; break;
            case 'mul': result = a * b; break;
            case 'div': result = b !== 0 ? a / b : "Division by zero"; break;
            default: result = "Unknown operation";
        }
        
        return {
            operation,
            operands: [a, b],
            result,
            timestamp: new Date().toISOString()
        };
    } catch (e) {
        return {error: e.message};
    }
});

console.log("Utility endpoints registered");