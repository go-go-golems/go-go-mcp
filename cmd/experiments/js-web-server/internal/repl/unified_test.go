package repl

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestUnifiedImplementation verifies that the REPL now uses the same engine as the MCP version
func TestUnifiedImplementation(t *testing.T) {
	// Create a new REPL model with the unified engine
	model := NewModel(false)

	// Test that the engine is properly initialized
	if model.shared.jsEngine == nil {
		t.Fatal("JavaScript engine not initialized")
	}

	// Test basic JavaScript execution
	result, err := model.shared.jsEngine.ExecuteScript("2 + 2")
	if err != nil {
		t.Fatalf("Failed to execute basic JavaScript: %v", err)
	}

	if result.Value != int64(4) {
		t.Errorf("Expected 4, got %v (type: %T)", result.Value, result.Value)
	}

	// Test console logging
	result, err = model.shared.jsEngine.ExecuteScript(`console.log("Hello World"); "done"`)
	if err != nil {
		t.Fatalf("Failed to execute console.log: %v", err)
	}

	if len(result.ConsoleLog) == 0 {
		t.Error("Console output not captured")
	}

	// Test app.get registration (Express.js API)
	_, err = model.shared.jsEngine.ExecuteScript(`
		app.get("/test", (req, res) => {
			res.json({message: "Hello from unified engine"});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// Test path parameter support
	_, err = model.shared.jsEngine.ExecuteScript(`
		app.get("/users/:id", (req, res) => {
			res.json({userId: req.params.id, path: req.path});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register handler with path params: %v", err)
	}

	// Test database functionality
	_, err = model.shared.jsEngine.ExecuteScript(`
		var users = db.query("SELECT 1 as id, 'test' as name");
		console.log("Query result:", users);
	`)
	if err != nil {
		t.Fatalf("Failed to execute database query: %v", err)
	}

	// Test geppetto integration
	_, err = model.shared.jsEngine.ExecuteScript(`
		var conv = new Conversation();
		var msgId = conv.addMessage("user", "Hello Geppetto!");
		console.log("Message ID:", msgId);
	`)
	if err != nil {
		t.Fatalf("Failed to use Geppetto API: %v", err)
	}

	// Test that handlers are properly registered in the engine
	handler, exists := model.shared.jsEngine.GetHandler("GET", "/test")
	if !exists {
		t.Error("Handler not found in engine registry")
	}
	if handler == nil {
		t.Error("Handler is nil")
	}

	// Test path parameter matching
	handler, exists = model.shared.jsEngine.GetHandler("GET", "/users/123")
	if !exists {
		t.Error("Path parameter handler not found")
	}
	if handler == nil {
		t.Error("Path parameter handler is nil")
	}
}

// TestHTTPIntegration tests the HTTP server integration
func TestHTTPIntegration(t *testing.T) {
	model := NewModel(false)

	// Register a test handler
	_, err := model.shared.jsEngine.ExecuteScript(`
		app.get("/api/test", (req, res) => {
			res.json({
				method: req.method,
				path: req.path,
				success: true
			});
		});
		
		app.get("/users/:id", (req, res) => {
			res.json({
				userId: req.params.id,
				method: req.method,
				path: req.path
			});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register handlers: %v", err)
	}

	// Test the dynamic route handler
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	handleDynamicRoute(model.shared.jsEngine, w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"success":true`) {
		t.Errorf("Response doesn't contain expected content: %s", body)
	}

	// Test path parameter extraction
	req = httptest.NewRequest("GET", "/users/456", nil)
	w = httptest.NewRecorder()

	handleDynamicRoute(model.shared.jsEngine, w, req)

	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body = w.Body.String()
	if !strings.Contains(body, `"userId":"456"`) {
		t.Errorf("Response doesn't contain expected userId: %s", body)
	}

	// Test 404 for unregistered routes
	req = httptest.NewRequest("GET", "/nonexistent", nil)
	w = httptest.NewRecorder()

	handleDynamicRoute(model.shared.jsEngine, w, req)

	resp = w.Result()
	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

// TestProcessInput tests the REPL input processing
func TestProcessInput(t *testing.T) {
	model := NewModel(false)

	// Test basic JavaScript execution
	model = model.processInput("2 + 2")

	if len(model.history) == 0 {
		t.Fatal("No history entry created")
	}

	lastEntry := model.history[len(model.history)-1]
	if lastEntry.isErr {
		t.Errorf("Execution marked as error: %s", lastEntry.output)
	}

	if !strings.Contains(lastEntry.output, "4") {
		t.Errorf("Expected output to contain '4', got: %s", lastEntry.output)
	}

	// Test console output capture
	model = model.processInput(`console.log("Test message"); "result"`)

	if len(model.history) < 2 {
		t.Fatal("Not enough history entries")
	}

	lastEntry = model.history[len(model.history)-1]
	if !strings.Contains(lastEntry.output, "Test message") {
		t.Errorf("Console output not captured: %s", lastEntry.output)
	}

	// Test error handling
	model = model.processInput("undefined.property")

	if len(model.history) < 3 {
		t.Fatal("Error case not added to history")
	}

	lastEntry = model.history[len(model.history)-1]
	if !lastEntry.isErr {
		t.Error("Error not properly marked")
	}
}
