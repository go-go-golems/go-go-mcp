package embeddable

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubAuthProvider struct {
	principal      AuthPrincipal
	err            error
	metadata       map[string]any
	authHeader     string
	receivedTokens []string
}

func (p *stubAuthProvider) MountRoutes(mux *http.ServeMux) {}

func (p *stubAuthProvider) ValidateBearerToken(ctx context.Context, token string) (AuthPrincipal, error) {
	_ = ctx
	p.receivedTokens = append(p.receivedTokens, token)
	if p.err != nil {
		return AuthPrincipal{}, p.err
	}
	return p.principal, nil
}

func (p *stubAuthProvider) ProtectedResourceMetadata() map[string]any {
	return p.metadata
}

func (p *stubAuthProvider) WWWAuthenticateHeader() string {
	return p.authHeader
}

func TestWithOIDCUsesEmbeddedDevMode(t *testing.T) {
	cfg := NewServerConfig()
	if err := WithOIDC(OIDCOptions{Issuer: "http://localhost:3001"})(cfg); err != nil {
		t.Fatalf("WithOIDC() error = %v", err)
	}

	if !cfg.authEnabled {
		t.Fatalf("expected auth to be enabled")
	}
	if cfg.authOptions.Mode != AuthModeEmbeddedDev {
		t.Fatalf("expected embedded dev mode, got %q", cfg.authOptions.Mode)
	}
}

func TestEffectiveResourceURLDoesNotDefaultExternalOIDCToIssuer(t *testing.T) {
	got := (AuthOptions{
		Mode: AuthModeExternalOIDC,
		External: ExternalOIDCOptions{
			IssuerURL: "https://issuer.example.com/realms/test",
		},
	}).EffectiveResourceURL()

	if got != "" {
		t.Fatalf("expected empty resource url fallback for external oidc, got %q", got)
	}
}

func TestAuthMiddlewareRejectsMissingBearer(t *testing.T) {
	provider := &stubAuthProvider{
		authHeader: `Bearer realm="mcp", resource="https://mcp.example.com/mcp"`,
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rec := httptest.NewRecorder()

	authMiddleware(provider, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected next handler call")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != provider.authHeader {
		t.Fatalf("unexpected WWW-Authenticate header: %q", got)
	}
}

func TestAuthMiddlewareInjectsPrincipalHeaders(t *testing.T) {
	provider := &stubAuthProvider{
		principal: AuthPrincipal{
			Subject:           "alice",
			ClientID:          "client-123",
			Issuer:            "https://auth.example.com/realms/smailnail",
			Email:             "alice@example.com",
			PreferredUsername: "alice",
		},
		authHeader: `Bearer realm="mcp"`,
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer token-123")
	rec := httptest.NewRecorder()

	authMiddleware(provider, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-MCP-Subject"); got != "alice" {
			t.Fatalf("unexpected subject header: %q", got)
		}
		if got := r.Header.Get("X-MCP-Client-ID"); got != "client-123" {
			t.Fatalf("unexpected client id header: %q", got)
		}
		principal, ok := GetAuthPrincipal(r.Context())
		if !ok {
			t.Fatalf("expected auth principal in request context")
		}
		if principal.Email != "alice@example.com" || principal.PreferredUsername != "alice" {
			t.Fatalf("unexpected principal in context: %#v", principal)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if len(provider.receivedTokens) != 1 || provider.receivedTokens[0] != "token-123" {
		t.Fatalf("unexpected received tokens: %#v", provider.receivedTokens)
	}
}

func TestEmbeddedDevProviderStaticAuthKey(t *testing.T) {
	provider, err := newEmbeddedDevAuthProvider(AuthOptions{
		Mode: AuthModeEmbeddedDev,
		Embedded: EmbeddedOIDCOptions{
			Issuer:  "http://localhost:3001",
			AuthKey: "STATIC_TOKEN",
		},
	})
	if err != nil {
		t.Fatalf("newEmbeddedDevAuthProvider() error = %v", err)
	}

	principal, err := provider.ValidateBearerToken(context.Background(), "STATIC_TOKEN")
	if err != nil {
		t.Fatalf("ValidateBearerToken() error = %v", err)
	}

	if principal.Subject != "static-key-user" || principal.ClientID != "static-key-client" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
}

func TestProtectedResourceHandlerUsesProviderMetadata(t *testing.T) {
	provider := &stubAuthProvider{
		metadata: map[string]any{
			"authorization_servers": []string{"https://issuer.example.com/realms/test"},
			"resource":              "https://mcp.example.com/mcp",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
	rec := httptest.NewRecorder()

	protectedResourceHandler(provider).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if got := body["resource"]; got != "https://mcp.example.com/mcp" {
		t.Fatalf("unexpected resource: %#v", got)
	}
}

func TestBuildBearerChallengeOnlyAdvertisesResourceMetadata(t *testing.T) {
	got := buildBearerChallenge("https://mcp.example.com/.well-known/oauth-protected-resource")
	want := `Bearer realm="mcp", resource_metadata="https://mcp.example.com/.well-known/oauth-protected-resource"`

	if got != want {
		t.Fatalf("unexpected challenge header:\nwant: %q\n got: %q", want, got)
	}
}
