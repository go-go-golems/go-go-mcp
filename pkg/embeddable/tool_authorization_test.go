package embeddable

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	official "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolAuthorizationPolicyDrivesDescriptorAndDispatch(t *testing.T) {
	const metadataURL = "https://mcp.example.com/.well-known/oauth-protected-resource"
	var handlerCalls atomic.Int32
	cfg := NewServerConfig()
	cfg.authOptions.ResourceURL = "https://mcp.example.com/mcp"
	for _, option := range []ServerOption{
		WithTool("protected", func(context.Context, map[string]any) (*protocol.ToolResult, error) {
			handlerCalls.Add(1)
			return protocol.NewToolResult(protocol.WithJSON(map[string]any{"ok": true})), nil
		}),
		WithToolAuthorization("protected", ToolAuthorizationPolicy{RequiredScopes: []string{"scope:b", "scope:a", "scope:a"}}),
	} {
		if err := option(cfg); err != nil {
			t.Fatal(err)
		}
	}

	server := official.NewServer(&official.Implementation{Name: "policy-test", Version: "1"}, nil)
	if err := registerToolsFromRegistry(context.Background(), server, cfg.toolRegistry, cfg); err != nil {
		t.Fatal(err)
	}
	streamable := official.NewStreamableHTTPHandler(func(*http.Request) *official.Server { return server }, &official.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	verifier := func(_ context.Context, token string, _ *http.Request) (*mcpauth.TokenInfo, error) {
		scopes := map[string][]string{
			"insufficient": {"scope:a"},
			"allowed":      {"scope:a", "scope:b"},
		}[token]
		if scopes == nil {
			return nil, mcpauth.ErrInvalidToken
		}
		return &mcpauth.TokenInfo{UserID: "alice", Scopes: scopes, Expiration: time.Now().Add(time.Hour)}, nil
	}
	handler := mcpauth.RequireBearerToken(verifier, &mcpauth.RequireBearerTokenOptions{ResourceMetadataURL: metadataURL})(streamable)

	listed := policyMCPRequest(t, handler, "allowed", `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`)
	tools := officialObjectField(t, listed, "result")["tools"].([]any)
	tool := tools[0].(map[string]any)
	meta := officialObjectField(t, tool, "_meta")
	schemes := meta["securitySchemes"].([]any)
	scheme := schemes[0].(map[string]any)
	scopes := scheme["scopes"].([]any)
	if len(scopes) != 2 || scopes[0] != "scope:a" || scopes[1] != "scope:b" {
		t.Fatalf("advertised scopes = %#v", scopes)
	}

	denied := policyMCPRequest(t, handler, "insufficient", `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"protected","arguments":{}}}`)
	deniedResult := officialObjectField(t, denied, "result")
	if deniedResult["isError"] != true {
		t.Fatalf("denied result = %#v", deniedResult)
	}
	challengeValues := officialObjectField(t, deniedResult, "_meta")["mcp/www_authenticate"].([]any)
	wantChallenge := `Bearer resource_metadata="https://mcp.example.com/.well-known/oauth-protected-resource", error="insufficient_scope", scope="scope:a scope:b"`
	if len(challengeValues) != 1 || challengeValues[0] != wantChallenge {
		t.Fatalf("challenge = %#v", challengeValues)
	}
	if got := handlerCalls.Load(); got != 0 {
		t.Fatalf("handler calls after denial = %d", got)
	}

	allowed := policyMCPRequest(t, handler, "allowed", `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"protected","arguments":{}}}`)
	if result := officialObjectField(t, allowed, "result"); result["isError"] == true {
		t.Fatalf("allowed result = %#v", result)
	}
	if got := handlerCalls.Load(); got != 1 {
		t.Fatalf("handler calls = %d, want 1", got)
	}
}

func TestToolAuthorizationPolicyValidation(t *testing.T) {
	tests := []ToolAuthorizationPolicy{
		{},
		{Public: true, RequiredScopes: []string{"scope:a"}},
		{RequiredScopes: []string{`bad scope`}},
	}
	for _, policy := range tests {
		if _, err := normalizeToolAuthorizationPolicy(policy); err == nil {
			t.Fatalf("expected validation error for %#v", policy)
		}
	}
}

func TestToolAuthorizationRejectsUnknownTool(t *testing.T) {
	cfg := NewServerConfig()
	if err := WithToolAuthorization("missing", ToolAuthorizationPolicy{RequiredScopes: []string{"scope:a"}})(cfg); err != nil {
		t.Fatal(err)
	}
	server := official.NewServer(&official.Implementation{Name: "test", Version: "1"}, nil)
	if err := registerToolsFromRegistry(context.Background(), server, cfg.toolRegistry, cfg); err == nil || !strings.Contains(err.Error(), "unregistered tool") {
		t.Fatalf("error = %v", err)
	}
}

func policyMCPRequest(t *testing.T, handler http.Handler, token, body string) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}
