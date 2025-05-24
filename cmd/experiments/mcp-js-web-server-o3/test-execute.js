// Test JavaScript code for the executeJS tool
console.log("Testing executeJS tool");
console.info("This is an info message");
console.warn("This is a warning");

// Create a simple calculation
const result = 2 + 3 * 4;
console.log("Calculation result:", result);

// Register a test endpoint
registerHandler("GET", "/test-api", () => ({
    message: "Hello from dynamically created endpoint!",
    timestamp: new Date().toISOString(),
    result: result
}));

console.log("Test endpoint registered at /test-api");

// Return a result
({
    success: true,
    calculatedValue: result,
    message: "Test completed successfully"
})