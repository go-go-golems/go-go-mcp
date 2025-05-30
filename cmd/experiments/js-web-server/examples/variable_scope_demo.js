// Demonstration of variable scope and re-execution behavior

console.log("=== Variable Scope Demo ===");

// ❌ This would fail on second execution:
// let counter = 0;  // SyntaxError: Identifier 'counter' has already been declared

// ✅ These approaches work for re-execution:

// Approach 1: Use var (can be redeclared)
var counter = 0;
counter++;

// Approach 2: Direct global assignment
requestCount = (requestCount || 0) + 1;

// Approach 3: Use globalState (recommended for app data)
if (!globalState.sessionCount) {
    globalState.sessionCount = 0;
}
globalState.sessionCount++;

// Functions can be safely redefined
function getCurrentStats() {
    return {
        counter: counter,
        requestCount: requestCount,
        sessionCount: globalState.sessionCount,
        timestamp: new Date().toISOString()
    };
}

// Register handlers that use our variables
registerHandler("GET", "/demo/stats", () => {
    return Response.json(getCurrentStats());
});

registerHandler("GET", "/demo/increment", () => {
    counter++;
    requestCount++;
    globalState.sessionCount++;
    
    return Response.json({
        message: "Counters incremented",
        stats: getCurrentStats()
    });
});

registerHandler("GET", "/demo/reset", () => {
    counter = 0;
    requestCount = 0;
    globalState.sessionCount = 0;
    
    return Response.json({
        message: "Counters reset",
        stats: getCurrentStats()
    });
});

console.log("Variable scope demo handlers registered!");
console.log("Try these endpoints:");
console.log("  GET /demo/stats     - View current counters");
console.log("  GET /demo/increment - Increment all counters");
console.log("  GET /demo/reset     - Reset all counters");

// Return current state to show execution result
getCurrentStats();