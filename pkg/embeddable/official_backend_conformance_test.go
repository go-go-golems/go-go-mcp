package embeddable

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	mcppsession "github.com/go-go-golems/go-go-mcp/pkg/session"
)

func TestMCPGoBackendConformance(t *testing.T) {
	var handlerCalls atomic.Int32
	cfg := NewServerConfig()
	options := []ServerOption{
		WithName("conformance-server"),
		WithVersion("1.2.3"),
		WithDefaultTransport("streamable_http"),
		WithTool("echo", func(ctx context.Context, args map[string]any) (*protocol.ToolResult, error) {
			handlerCalls.Add(1)
			if _, ok := mcppsession.SessionIDFromContext(ctx); ok {
				t.Fatal("stateless Streamable HTTP unexpectedly injected a session ID")
			}
			result := protocol.NewToolResult(protocol.WithJSON(map[string]any{"echo": args["message"]}))
			result.Meta = map[string]any{"test/result-meta": "preserved"}
			return result, nil
		},
			WithDescription("Echo one message"),
			WithSchema(json.RawMessage(`{"type":"object","properties":{"message":{"type":"string"}},"required":["message"],"additionalProperties":false}`)),
		),
	}
	for _, option := range options {
		if err := option(cfg); err != nil {
			t.Fatalf("configure server: %v", err)
		}
	}

	mux := http.NewServeMux()
	if err := MountHTTPHandlers(mux, cfg); err != nil {
		t.Fatalf("MountHTTPHandlers() error = %v", err)
	}

	sessionID, initialized := mcpRequest(t, mux, "", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"baseline-test","version":"1"}}}`)
	if got := initialized["jsonrpc"]; got != "2.0" {
		t.Fatalf("initialize jsonrpc = %#v", got)
	}
	initResult := objectField(t, initialized, "result")
	if got := initResult["protocolVersion"]; got != "2025-06-18" {
		t.Fatalf("protocolVersion = %#v, want 2025-06-18", got)
	}
	serverInfo := objectField(t, initResult, "serverInfo")
	if serverInfo["name"] != "conformance-server" || serverInfo["version"] != "1.2.3" {
		t.Fatalf("serverInfo = %#v", serverInfo)
	}

	_, listed := mcpRequest(t, mux, sessionID, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	listResult := objectField(t, listed, "result")
	tools, ok := listResult["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("tools = %#v, want one tool", listResult["tools"])
	}
	golden := `{"description":"Echo one message","inputSchema":{"additionalProperties":false,"properties":{"message":{"type":"string"}},"required":["message"],"type":"object"},"name":"echo"}`
	encodedTool, err := json.Marshal(tools[0])
	if err != nil {
		t.Fatal(err)
	}
	if got := string(encodedTool); got != golden {
		t.Fatalf("tools/list descriptor changed\nwant: %s\n got: %s", golden, got)
	}

	_, called := mcpRequest(t, mux, sessionID, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello"}}}`)
	callResult := objectField(t, called, "result")
	if got := callResult["isError"]; got != nil && got != false {
		t.Fatalf("isError = %#v, want absent or false", got)
	}
	meta := objectField(t, callResult, "_meta")
	if got := meta["test/result-meta"]; got != "preserved" {
		t.Fatalf("result metadata = %#v", meta)
	}
	contents, ok := callResult["content"].([]any)
	if !ok || len(contents) != 1 {
		t.Fatalf("content = %#v", callResult["content"])
	}
	content, ok := contents[0].(map[string]any)
	if !ok || content["type"] != "text" || content["text"] != `{"echo":"hello"}` {
		t.Fatalf("content = %#v", contents[0])
	}
	if got := handlerCalls.Load(); got != 1 {
		t.Fatalf("handler calls = %d, want 1", got)
	}
	if sessionID != "" {
		t.Fatalf("stateless Streamable HTTP issued session ID %q", sessionID)
	}
}

func TestStreamableHTTPDefaultsToModernStatelessDiscovery(t *testing.T) {
	cfg := NewServerConfig()
	for _, option := range []ServerOption{
		WithName("modern-server"),
		WithVersion("2.0.0"),
		WithDefaultTransport("streamable_http"),
		WithTool("echo", func(context.Context, map[string]any) (*protocol.ToolResult, error) {
			return protocol.NewToolResult(protocol.WithText("ok")), nil
		}),
	} {
		if err := option(cfg); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	if err := MountHTTPHandlers(mux, cfg); err != nil {
		t.Fatal(err)
	}
	body := `{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{},"io.modelcontextprotocol/clientInfo":{"name":"modern-test","version":"1"}}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2026-07-28")
	req.Header.Set("Mcp-Method", "server/discover")
	req.Header.Set("Mcp-Name", "modern-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("server/discover status = %d body=%s", rec.Code, rec.Body.String())
	}
	payload := rec.Body.Bytes()
	if bytes.HasPrefix(payload, []byte("event:")) {
		for _, line := range bytes.Split(payload, []byte("\n")) {
			if bytes.HasPrefix(line, []byte("data: ")) {
				payload = bytes.TrimPrefix(line, []byte("data: "))
				break
			}
		}
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode server/discover %q: %v", payload, err)
	}
	result := objectField(t, decoded, "result")
	versions, ok := result["supportedVersions"].([]any)
	if !ok || len(versions) == 0 || versions[0] != "2026-07-28" {
		t.Fatalf("supportedVersions = %#v", result["supportedVersions"])
	}
	capabilities := objectField(t, result, "capabilities")
	if _, ok := capabilities["tools"]; !ok {
		t.Fatalf("discovered capabilities = %#v", capabilities)
	}
	if sessionID := rec.Header().Get("Mcp-Session-Id"); sessionID != "" {
		t.Fatalf("server/discover issued session ID %q", sessionID)
	}
}

func TestMCPGoBackendMiddlewareAndHooksRunOnce(t *testing.T) {
	var middlewareCalls, handlerCalls, beforeCalls, afterCalls atomic.Int32
	cfg := NewServerConfig()
	options := []ServerOption{
		WithDefaultTransport("streamable_http"),
		WithMiddleware(func(next ToolHandler) ToolHandler {
			return func(ctx context.Context, args map[string]any) (*protocol.ToolResult, error) {
				middlewareCalls.Add(1)
				return next(ctx, args)
			}
		}),
		WithHooks(&Hooks{
			BeforeToolCall: func(context.Context, string, map[string]any) error {
				beforeCalls.Add(1)
				return nil
			},
			AfterToolCall: func(context.Context, string, *protocol.ToolResult, error) {
				afterCalls.Add(1)
			},
		}),
		WithTool("count", func(context.Context, map[string]any) (*protocol.ToolResult, error) {
			handlerCalls.Add(1)
			return protocol.NewToolResult(protocol.WithText("ok")), nil
		}),
	}
	for _, option := range options {
		if err := option(cfg); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	if err := MountHTTPHandlers(mux, cfg); err != nil {
		t.Fatal(err)
	}
	sessionID, _ := mcpRequest(t, mux, "", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`)
	mcpRequest(t, mux, sessionID, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"count","arguments":{}}}`)

	if got := middlewareCalls.Load(); got != 1 {
		t.Errorf("middleware calls = %d, want 1", got)
	}
	if got := handlerCalls.Load(); got != 1 {
		t.Errorf("handler calls = %d, want 1", got)
	}
	if got := beforeCalls.Load(); got != 1 {
		t.Errorf("before hook calls = %d, want 1", got)
	}
	if got := afterCalls.Load(); got != 1 {
		t.Errorf("after hook calls = %d, want 1", got)
	}
}

func mcpRequest(t *testing.T, handler http.Handler, sessionID, body string) (string, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("MCP status = %d body=%s", rec.Code, rec.Body.String())
	}
	payload := rec.Body.Bytes()
	if bytes.HasPrefix(payload, []byte("event:")) {
		for _, line := range bytes.Split(payload, []byte("\n")) {
			if bytes.HasPrefix(line, []byte("data: ")) {
				payload = bytes.TrimPrefix(line, []byte("data: "))
				break
			}
		}
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode MCP body %q: %v", payload, err)
	}
	return rec.Header().Get("Mcp-Session-Id"), decoded
}

func objectField(t *testing.T, object map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := object[key].(map[string]any)
	if !ok {
		t.Fatalf("%s = %#v, want object", key, object[key])
	}
	return value
}

// TestSimpleToolOutputSchemaAndAnnotations verifies that the simple WithTool
// path carries outputSchema and annotations through tools/list exactly.
func TestSimpleToolOutputSchemaAndAnnotations(t *testing.T) {
	cfg := NewServerConfig()
	annotations := protocol.ToolAnnotations{
		Title:           "Echo",
		ReadOnlyHint:    true,
		DestructiveHint: boolPtr(false),
		IdempotentHint:  true,
		OpenWorldHint:   boolPtr(false),
	}
	options := []ServerOption{
		WithDefaultTransport("streamable_http"),
		WithTool("echo", func(ctx context.Context, args map[string]any) (*protocol.ToolResult, error) {
			return protocol.NewToolResult(protocol.WithJSON(map[string]any{"echo": args["message"]})), nil
		},
			WithDescription("Echo one message"),
			WithSchema(json.RawMessage(`{"type":"object","properties":{"message":{"type":"string"}},"required":["message"],"additionalProperties":false}`)),
			WithOutputSchema(json.RawMessage(`{"type":"object","properties":{"echo":{"type":"string"}},"required":["echo"],"additionalProperties":false}`)),
			WithToolAnnotations(annotations),
		),
	}
	for _, option := range options {
		if err := option(cfg); err != nil {
			t.Fatalf("configure server: %v", err)
		}
	}

	mux := http.NewServeMux()
	if err := MountHTTPHandlers(mux, cfg); err != nil {
		t.Fatalf("MountHTTPHandlers() error = %v", err)
	}

	sessionID, _ := mcpRequest(t, mux, "", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"descriptor-test","version":"1"}}}`)
	_, listed := mcpRequest(t, mux, sessionID, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	listResult := objectField(t, listed, "result")
	tools, ok := listResult["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("tools = %#v, want one tool", listResult["tools"])
	}
	tool, ok := tools[0].(map[string]any)
	if !ok {
		t.Fatalf("tool = %#v", tools[0])
	}
	outputSchema, ok := tool["outputSchema"].(map[string]any)
	if !ok {
		t.Fatalf("outputSchema = %#v", tool["outputSchema"])
	}
	if outputSchema["type"] != "object" {
		t.Fatalf("outputSchema.type = %#v", outputSchema["type"])
	}
	props, ok := outputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("outputSchema.properties = %#v", outputSchema["properties"])
	}
	if _, ok := props["echo"]; !ok {
		t.Fatalf("outputSchema.properties missing echo: %#v", props)
	}

	gotAnnotations, ok := tool["annotations"].(map[string]any)
	if !ok {
		t.Fatalf("annotations = %#v", tool["annotations"])
	}
	if gotAnnotations["title"] != "Echo" {
		t.Fatalf("annotations.title = %#v", gotAnnotations["title"])
	}
	if gotAnnotations["readOnlyHint"] != true {
		t.Fatalf("annotations.readOnlyHint = %#v", gotAnnotations["readOnlyHint"])
	}
	if gotAnnotations["destructiveHint"] != false {
		t.Fatalf("annotations.destructiveHint = %#v", gotAnnotations["destructiveHint"])
	}
	if gotAnnotations["idempotentHint"] != true {
		t.Fatalf("annotations.idempotentHint = %#v", gotAnnotations["idempotentHint"])
	}
	if gotAnnotations["openWorldHint"] != false {
		t.Fatalf("annotations.openWorldHint = %#v", gotAnnotations["openWorldHint"])
	}
}
