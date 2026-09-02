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
	var handlerSessionID atomic.Value
	cfg := NewServerConfig()
	options := []ServerOption{
		WithName("conformance-server"),
		WithVersion("1.2.3"),
		WithDefaultTransport("streamable_http"),
		WithTool("echo", func(ctx context.Context, args map[string]any) (*protocol.ToolResult, error) {
			handlerCalls.Add(1)
			if sessionID, ok := mcppsession.SessionIDFromContext(ctx); ok {
				handlerSessionID.Store(sessionID)
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
	if got := handlerSessionID.Load().(string); got != sessionID {
		t.Fatalf("handler session ID = %q, want %q", got, sessionID)
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
