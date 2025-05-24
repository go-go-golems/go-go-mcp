// Test script for re-execution
console.log("Test script loaded at:", new Date().toISOString());

// Use globalState for persistent counter
if (!globalState.testCounter) {
    globalState.testCounter = 0;
    console.log("Initialized testCounter");
} else {
    console.log("testCounter already exists:", globalState.testCounter);
}

globalState.testCounter++;
console.log("testCounter is now:", globalState.testCounter);

// Register a simple endpoint that shows the counter
registerHandler("GET", "/test-reload", () => ({
    message: "Test reload endpoint",
    executions: globalState.testCounter,
    timestamp: new Date().toISOString()
}));

console.log("Test script execution complete");