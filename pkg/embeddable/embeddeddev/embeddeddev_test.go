package embeddeddev

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	embeddedoidc "github.com/go-go-golems/go-go-mcp/pkg/auth/oidc"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
)

const initializeRequest = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`

func TestProviderStaticAuthKeyReportsExpiration(t *testing.T) {
	p, err := newProvider(embeddable.AuthOptions{
		Mode:     embeddable.AuthModeEmbeddedDev,
		Embedded: embeddable.EmbeddedOIDCOptions{Issuer: "http://localhost:3001", AuthKey: "STATIC_TOKEN"},
	})
	if err != nil {
		t.Fatalf("newProvider() error = %v", err)
	}

	principal, err := p.ValidateBearerToken(context.Background(), "STATIC_TOKEN")
	if err != nil {
		t.Fatalf("ValidateBearerToken() error = %v", err)
	}
	if principal.Subject != "static-key-user" || principal.ClientID != "static-key-client" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
	if !principal.Expiration.After(time.Now()) {
		t.Fatalf("static key principal expiration = %v, want a time in the future", principal.Expiration)
	}

	if _, err := p.ValidateBearerToken(context.Background(), "WRONG"); err == nil {
		t.Fatal("expected an unknown token to be rejected")
	}
}

func TestProviderDevTokenReportsStoredExpiration(t *testing.T) {
	p, err := newProvider(embeddable.AuthOptions{
		Mode: embeddable.AuthModeEmbeddedDev,
		Embedded: embeddable.EmbeddedOIDCOptions{
			Issuer:          "http://localhost:3001",
			DBPath:          filepath.Join(t.TempDir(), "oidc.db"),
			EnableDevTokens: true,
		},
	})
	if err != nil {
		t.Fatalf("newProvider() error = %v", err)
	}
	expires := time.Now().Add(time.Hour).Truncate(time.Second)
	if err := p.server.PersistToken(embeddedoidc.TokenRecord{Token: "dev-token", Subject: "alice", ClientID: "dev-client", ExpiresAt: expires}); err != nil {
		t.Fatalf("PersistToken() error = %v", err)
	}

	principal, err := p.ValidateBearerToken(context.Background(), "dev-token")
	if err != nil {
		t.Fatalf("ValidateBearerToken() error = %v", err)
	}
	if principal.Subject != "alice" || principal.ClientID != "dev-client" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
	if !principal.Expiration.Equal(expires) {
		t.Fatalf("principal expiration = %v, want %v", principal.Expiration, expires)
	}
}

// TestEmbeddedDevAcceptsStaticKeyOverHTTP drives the real streamable HTTP
// handler. Before expirations were reported, go-sdk's bearer middleware
// answered every embedded_dev token with "401 token missing expiration".
func TestEmbeddedDevAcceptsStaticKeyOverHTTP(t *testing.T) {
	cfg := embeddable.NewServerConfig()
	for _, opt := range []embeddable.ServerOption{
		embeddable.WithDefaultTransport("streamable_http"),
		WithOIDC(embeddable.EmbeddedOIDCOptions{Issuer: "http://localhost:3001", AuthKey: "STATIC_TOKEN"}),
	} {
		if err := opt(cfg); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	if err := embeddable.MountHTTPHandlers(mux, cfg); err != nil {
		t.Fatalf("MountHTTPHandlers() error = %v", err)
	}

	post := func(token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(initializeRequest))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	if rec := post("STATIC_TOKEN"); rec.Code != http.StatusOK {
		t.Fatalf("static key: status %d, body %q", rec.Code, rec.Body.String())
	}
	if rec := post("WRONG"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token: status %d, want 401", rec.Code)
	}

	// The embedded authorization server's routes are mounted next to /mcp.
	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"issuer"`) {
		t.Fatalf("discovery: status %d, body %q", rec.Code, rec.Body.String())
	}
}

func TestWithOIDCSelectsEmbeddedDevMode(t *testing.T) {
	cfg := embeddable.NewServerConfig()
	if err := WithOIDC(embeddable.EmbeddedOIDCOptions{Issuer: "http://localhost:3001"})(cfg); err != nil {
		t.Fatalf("WithOIDC() error = %v", err)
	}
	mux := http.NewServeMux()
	if err := embeddable.WithDefaultTransport("streamable_http")(cfg); err != nil {
		t.Fatal(err)
	}
	if err := embeddable.MountHTTPHandlers(mux, cfg); err != nil {
		t.Fatalf("MountHTTPHandlers() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(initializeRequest))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("request without a token: status %d, want 401", rec.Code)
	}
}
