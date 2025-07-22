package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

func TestCrushEditor_AllTransportTypes(t *testing.T) {
	// Create temporary file for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-crush.json")

	editor, err := NewCrushEditor(configPath)
	if err != nil {
		t.Fatalf("Failed to create CrushEditor: %v", err)
	}

	// Test 1: Add stdio server
	stdioServer := types.CommonServer{
		Name:    "test-stdio",
		Command: "node",
		Args:    []string{"server.js", "--port", "3000"},
		Env:     map[string]string{"NODE_ENV": "development", "DEBUG": "1"},
	}

	err = editor.AddMCPServer(stdioServer, false)
	if err != nil {
		t.Fatalf("Failed to add stdio server: %v", err)
	}

	// Test 2: Add HTTP server
	httpServer := types.CommonServer{
		Name:  "test-http",
		URL:   "https://api.example.com/mcp",
		IsSSE: false,
		Env:   map[string]string{"Authorization": "Bearer token123", "Content-Type": "application/json"},
	}

	err = editor.AddMCPServer(httpServer, false)
	if err != nil {
		t.Fatalf("Failed to add HTTP server: %v", err)
	}

	// Test 3: Add SSE server
	sseServer := types.CommonServer{
		Name:  "test-sse",
		URL:   "https://events.example.com/mcp/sse",
		IsSSE: true,
		Env:   map[string]string{"Authorization": "Bearer token456", "X-Client-ID": "client123"},
	}

	err = editor.AddMCPServer(sseServer, false)
	if err != nil {
		t.Fatalf("Failed to add SSE server: %v", err)
	}

	// Save configuration
	err = editor.Save()
	if err != nil {
		t.Fatalf("Failed to save configuration: %v", err)
	}

	// Verify JSON structure by reading the file directly
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config CrushMCPConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test 4: Verify stdio server in JSON
	stdioEntry, exists := config.MCP["test-stdio"]
	if !exists {
		t.Fatal("Stdio server not found in config")
	}
	if stdioEntry.Type != "stdio" {
		t.Errorf("Expected stdio type, got %s", stdioEntry.Type)
	}
	if stdioEntry.Command != "node" {
		t.Errorf("Expected command 'node', got %s", stdioEntry.Command)
	}
	if len(stdioEntry.Args) != 3 || stdioEntry.Args[0] != "server.js" {
		t.Errorf("Unexpected args: %v", stdioEntry.Args)
	}
	if stdioEntry.Env["NODE_ENV"] != "development" {
		t.Errorf("Expected env NODE_ENV=development, got %s", stdioEntry.Env["NODE_ENV"])
	}

	// Test 5: Verify HTTP server in JSON
	httpEntry, exists := config.MCP["test-http"]
	if !exists {
		t.Fatal("HTTP server not found in config")
	}
	if httpEntry.Type != "http" {
		t.Errorf("Expected http type, got %s", httpEntry.Type)
	}
	if httpEntry.URL != "https://api.example.com/mcp" {
		t.Errorf("Expected URL 'https://api.example.com/mcp', got %s", httpEntry.URL)
	}
	if httpEntry.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Expected Authorization header, got %s", httpEntry.Headers["Authorization"])
	}

	// Test 6: Verify SSE server in JSON
	sseEntry, exists := config.MCP["test-sse"]
	if !exists {
		t.Fatal("SSE server not found in config")
	}
	if sseEntry.Type != "sse" {
		t.Errorf("Expected sse type, got %s", sseEntry.Type)
	}
	if sseEntry.URL != "https://events.example.com/mcp/sse" {
		t.Errorf("Expected URL 'https://events.example.com/mcp/sse', got %s", sseEntry.URL)
	}
	if sseEntry.Headers["X-Client-ID"] != "client123" {
		t.Errorf("Expected X-Client-ID header, got %s", sseEntry.Headers["X-Client-ID"])
	}

	// Test 7: Test ListServers mapping back to CommonServer
	servers, err := editor.ListServers()
	if err != nil {
		t.Fatalf("Failed to list servers: %v", err)
	}

	if len(servers) != 3 {
		t.Fatalf("Expected 3 servers, got %d", len(servers))
	}

	// Verify stdio server mapping
	retrievedStdio := servers["test-stdio"]
	if retrievedStdio.Command != "node" {
		t.Errorf("Stdio server command mismatch: expected 'node', got %s", retrievedStdio.Command)
	}
	if retrievedStdio.IsSSE {
		t.Error("Stdio server should not be SSE")
	}
	if retrievedStdio.Env["NODE_ENV"] != "development" {
		t.Errorf("Stdio server env mismatch: expected development, got %s", retrievedStdio.Env["NODE_ENV"])
	}

	// Verify HTTP server mapping
	retrievedHTTP := servers["test-http"]
	if retrievedHTTP.URL != "https://api.example.com/mcp" {
		t.Errorf("HTTP server URL mismatch: expected 'https://api.example.com/mcp', got %s", retrievedHTTP.URL)
	}
	if retrievedHTTP.IsSSE {
		t.Error("HTTP server should not be SSE")
	}
	if retrievedHTTP.Env["Authorization"] != "Bearer token123" {
		t.Errorf("HTTP server env mismatch: expected Bearer token123, got %s", retrievedHTTP.Env["Authorization"])
	}

	// Verify SSE server mapping
	retrievedSSE := servers["test-sse"]
	if retrievedSSE.URL != "https://events.example.com/mcp/sse" {
		t.Errorf("SSE server URL mismatch: expected 'https://events.example.com/mcp/sse', got %s", retrievedSSE.URL)
	}
	if !retrievedSSE.IsSSE {
		t.Error("SSE server should be SSE")
	}
	if retrievedSSE.Env["X-Client-ID"] != "client123" {
		t.Errorf("SSE server env mismatch: expected client123, got %s", retrievedSSE.Env["X-Client-ID"])
	}

	// Test 8: Test GetServer for each type
	stdioRetrieved, found, err := editor.GetServer("test-stdio")
	if err != nil {
		t.Fatalf("Failed to get stdio server: %v", err)
	}
	if !found {
		t.Fatal("Stdio server not found")
	}
	if stdioRetrieved.Command != "node" {
		t.Error("GetServer failed for stdio server")
	}

	httpRetrieved, found, err := editor.GetServer("test-http")
	if err != nil {
		t.Fatalf("Failed to get HTTP server: %v", err)
	}
	if !found {
		t.Fatal("HTTP server not found")
	}
	if httpRetrieved.IsSSE {
		t.Error("GetServer failed for HTTP server - should not be SSE")
	}

	sseRetrieved, found, err := editor.GetServer("test-sse")
	if err != nil {
		t.Fatalf("Failed to get SSE server: %v", err)
	}
	if !found {
		t.Fatal("SSE server not found")
	}
	if !sseRetrieved.IsSSE {
		t.Error("GetServer failed for SSE server - should be SSE")
	}

	t.Logf("All transport type tests passed!")
}

func TestCrushEditor_InvalidServerConfigurations(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-invalid.json")

	editor, err := NewCrushEditor(configPath)
	if err != nil {
		t.Fatalf("Failed to create CrushEditor: %v", err)
	}

	// Test invalid configuration: neither command nor URL
	invalidServer := types.CommonServer{
		Name: "invalid-server",
	}

	err = editor.AddMCPServer(invalidServer, false)
	if err == nil {
		t.Error("Expected error for invalid server configuration, but got none")
	}

	// Test invalid configuration: both command and URL
	bothServer := types.CommonServer{
		Name:    "both-server",
		Command: "node",
		URL:     "https://example.com",
	}

	// This should work since we prioritize Command when both are present
	err = editor.AddMCPServer(bothServer, false)
	if err != nil {
		t.Errorf("Unexpected error for server with both command and URL: %v", err)
	}
}
