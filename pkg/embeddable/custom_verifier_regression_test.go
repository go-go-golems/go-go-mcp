package embeddable

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestCustomVerifierToolChallengeAndPrivateArguments(t *testing.T) {
	var captured bytes.Buffer
	old := log.Logger
	log.Logger = zerolog.New(&captured).Level(zerolog.TraceLevel)
	t.Cleanup(func() { log.Logger = old })
	scopes, _ := ParseScopeSet([]string{"other"})
	provider := &stubAuthProvider{principal: AuthPrincipal{Subject: "alice", Issuer: "https://ttc.test", Expiration: time.Now().Add(time.Hour), Scopes: scopes}, metadata: map[string]any{"resource": "https://ttc.test/mcp"}}
	cfg := NewServerConfig()
	called := 0
	for _, o := range []ServerOption{WithDefaultTransport("streamable_http"), WithStreamableHTTPStateless(true), WithStreamableHTTPJSONResponse(true), WithHTTPAuthVerifier(provider), WithTool("query", func(context.Context, map[string]any) (*protocol.ToolResult, error) {
		called++
		return protocol.NewToolResult(protocol.WithJSON(map[string]any{"row": "secret-result-canary-926"})), nil
	}, WithSchema(json.RawMessage(`{"type":"object","properties":{"sql":{"type":"string"}}}`))), WithToolAuthorization("query", ToolAuthorizationPolicy{RequiredScopes: []string{"read"}})} {
		if err := o(cfg); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	if err := MountHTTPHandlers(mux, cfg); err != nil {
		t.Fatal(err)
	}
	denied := policyMCPRequest(t, mux, "valid", `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query","arguments":{"sql":"secret-canary-925"}}}`)
	data, _ := json.Marshal(denied)
	if !bytes.Contains(data, []byte("oauth-protected-resource/mcp")) || called != 0 {
		t.Fatalf("missing metadata or unauthorized invocation: %s calls=%d", data, called)
	}
	provider.principal.Scopes, _ = ParseScopeSet([]string{"read"})
	policyMCPRequest(t, mux, "valid", `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"query","arguments":{"sql":"secret-canary-925"}}}`)
	if bytes.Contains(captured.Bytes(), []byte("secret-result-canary-926")) {
		t.Fatal("private result logged")
	}
	if called != 1 || bytes.Contains(captured.Bytes(), []byte("secret-canary-925")) {
		t.Fatalf("calls=%d, private argument logged=%v", called, bytes.Contains(captured.Bytes(), []byte("secret-canary-925")))
	}
}
