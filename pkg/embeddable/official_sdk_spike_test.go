package embeddable

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	official "github.com/modelcontextprotocol/go-sdk/mcp"
	oauthex "github.com/modelcontextprotocol/go-sdk/oauthex"
)

// TestOfficialSDKSpike proves the official SDK capabilities needed before the
// production embeddable backend is switched away from Mark3 Labs.
func TestOfficialSDKSpike(t *testing.T) {
	const (
		resourceURL = "https://mcp.example.com/mcp"
		metadataURL = "https://mcp.example.com/.well-known/oauth-protected-resource"
	)

	server := official.NewServer(&official.Implementation{Name: "official-spike", Version: "0.1.0"}, nil)
	rawSchema := json.RawMessage(`{"type":"object","properties":{"message":{"type":"string"}},"required":["message"],"additionalProperties":false}`)
	server.AddTool(&official.Tool{
		Name:        "echo",
		Description: "Echo a message through the official SDK",
		InputSchema: rawSchema,
		Meta: official.Meta{
			"securitySchemes": []any{map[string]any{"type": "oauth2", "scopes": []string{"mcp:invoke"}}},
		},
	}, func(_ context.Context, req *official.CallToolRequest) (*official.CallToolResult, error) {
		var input struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, err
		}
		structured := map[string]any{"echo": input.Message}
		return &official.CallToolResult{
			Meta:              official.Meta{"test/result-meta": "preserved"},
			Content:           []official.Content{&official.TextContent{Text: input.Message}},
			StructuredContent: structured,
		}, nil
	})

	streamable := official.NewStreamableHTTPHandler(
		func(*http.Request) *official.Server { return server },
		&official.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
	verifier := func(_ context.Context, token string, _ *http.Request) (*mcpauth.TokenInfo, error) {
		if token != "valid-token" {
			return nil, mcpauth.ErrInvalidToken
		}
		return &mcpauth.TokenInfo{
			UserID:     "alice",
			Scopes:     []string{"mcp:invoke"},
			Expiration: time.Now().Add(time.Hour),
		}, nil
	}
	protected := mcpauth.RequireBearerToken(verifier, &mcpauth.RequireBearerTokenOptions{
		Scopes:              []string{"mcp:invoke"},
		ResourceMetadataURL: metadataURL,
	})(streamable)

	mux := http.NewServeMux()
	mux.Handle("/mcp", protected)
	mux.Handle("/.well-known/oauth-protected-resource", mcpauth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:             resourceURL,
		AuthorizationServers: []string{"https://auth.example.com"},
		ScopesSupported:      []string{"mcp:invoke"},
	}))

	t.Run("RFC 9728 challenge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"spike","version":"1"}}}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
		challenge := rec.Header().Get("WWW-Authenticate")
		if !strings.Contains(challenge, `resource_metadata="`+metadataURL+`"`) {
			t.Fatalf("WWW-Authenticate = %q", challenge)
		}
	})

	t.Run("protected resource metadata", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("CORS allow origin = %q, want *", got)
		}
		var metadata oauthex.ProtectedResourceMetadata
		if err := json.Unmarshal(rec.Body.Bytes(), &metadata); err != nil {
			t.Fatal(err)
		}
		if metadata.Resource != resourceURL || len(metadata.ScopesSupported) != 1 || metadata.ScopesSupported[0] != "mcp:invoke" {
			t.Fatalf("metadata = %#v", metadata)
		}
	})

	initialized := officialSDKRequest(t, mux, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"spike","version":"1"}}}`)
	if got := officialObjectField(t, initialized, "result")["protocolVersion"]; got != "2025-06-18" {
		t.Fatalf("protocolVersion = %#v", got)
	}

	listed := officialSDKRequest(t, mux, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	tools := officialObjectField(t, listed, "result")["tools"].([]any)
	tool := tools[0].(map[string]any)
	if tool["name"] != "echo" {
		t.Fatalf("tool = %#v", tool)
	}
	meta := officialObjectField(t, tool, "_meta")
	if meta["securitySchemes"] == nil {
		t.Fatalf("tool metadata = %#v", meta)
	}
	if schema := tool["inputSchema"].(map[string]any); schema["additionalProperties"] != false {
		t.Fatalf("raw input schema = %#v", schema)
	}

	called := officialSDKRequest(t, mux, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello"}}}`)
	result := officialObjectField(t, called, "result")
	if got := officialObjectField(t, result, "structuredContent")["echo"]; got != "hello" {
		t.Fatalf("structuredContent.echo = %#v", got)
	}
	if got := officialObjectField(t, result, "_meta")["test/result-meta"]; got != "preserved" {
		t.Fatalf("result metadata = %#v", result["_meta"])
	}
}

func officialSDKRequest(t *testing.T, handler http.Handler, body string) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return response
}

func officialObjectField(t *testing.T, object map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := object[key].(map[string]any)
	if !ok {
		t.Fatalf("%s = %#v, want object", key, object[key])
	}
	return value
}
