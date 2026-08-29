package embeddable

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
)

func TestNewHTTPAuthProviderSelectsExternalOIDC(t *testing.T) {
	privateKey, publicJWK := generateTestJWK(t)
	server, issuer, discoveryURL := newTestOIDCServer(t, publicJWK)
	defer server.Close()

	cfg := NewServerConfig()
	cfg.authEnabled = true
	cfg.authOptions = AuthOptions{
		Mode:        AuthModeExternalOIDC,
		ResourceURL: "https://mcp.example.com/mcp",
		External: ExternalOIDCOptions{
			IssuerURL:    issuer,
			DiscoveryURL: discoveryURL,
		},
	}

	provider, err := newHTTPAuthProvider(cfg)
	if err != nil {
		t.Fatalf("newHTTPAuthProvider() error = %v", err)
	}
	if _, ok := provider.(*externalOIDCAuthProvider); !ok {
		t.Fatalf("expected externalOIDCAuthProvider, got %T", provider)
	}

	token := signExternalTestToken(t, privateKey, issuer, "mcp-resource", "client-1", "openid")
	if _, err := provider.ValidateBearerToken(context.Background(), token); err != nil {
		t.Fatalf("ValidateBearerToken() error = %v", err)
	}
}

func TestExternalOIDCProviderValidatesJWTAndAdvertisesResourceMetadata(t *testing.T) {
	privateKey, publicJWK := generateTestJWK(t)
	server, issuer, discoveryURL := newTestOIDCServer(t, publicJWK)
	defer server.Close()

	provider, err := newExternalOIDCAuthProvider(AuthOptions{
		Mode:        AuthModeExternalOIDC,
		ResourceURL: "https://mcp.example.com/mcp",
		External: ExternalOIDCOptions{
			IssuerURL:      issuer,
			DiscoveryURL:   discoveryURL,
			Audience:       "mcp-resource",
			RequiredScopes: []string{"mcp:invoke"},
		},
	})
	if err != nil {
		t.Fatalf("newExternalOIDCAuthProvider() error = %v", err)
	}

	token := signExternalTestToken(t, privateKey, issuer, "mcp-resource", "client-1", "openid profile mcp:invoke")

	principal, err := provider.ValidateBearerToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateBearerToken() error = %v", err)
	}

	if principal.Subject != "alice" {
		t.Fatalf("unexpected subject: %q", principal.Subject)
	}
	if principal.ClientID != "client-1" {
		t.Fatalf("unexpected client id: %q", principal.ClientID)
	}
	if principal.Issuer != issuer {
		t.Fatalf("unexpected issuer: %q", principal.Issuer)
	}
	if principal.Email != "alice@example.com" {
		t.Fatalf("unexpected email: %q", principal.Email)
	}
	if principal.PreferredUsername != "alice" {
		t.Fatalf("unexpected preferred username: %q", principal.PreferredUsername)
	}
	if principal.DisplayName != "Alice Example" {
		t.Fatalf("unexpected display name: %q", principal.DisplayName)
	}

	header := provider.WWWAuthenticateHeader()
	if want := `Bearer realm="mcp", resource_metadata="https://mcp.example.com/.well-known/oauth-protected-resource"`; header != want {
		t.Fatalf("unexpected WWW-Authenticate header:\nwant: %q\n got: %q", want, header)
	}
	if strings.Contains(header, "authorization_uri=") {
		t.Fatalf("unexpected authorization_uri in WWW-Authenticate: %q", header)
	}

	metadata := provider.ProtectedResourceMetadata()
	if got := metadata["resource"]; got != "https://mcp.example.com/mcp" {
		t.Fatalf("unexpected protected resource metadata: %#v", metadata)
	}
}

func TestExternalOIDCProviderRejectsInvalidTokens(t *testing.T) {
	privateKey, publicJWK := generateTestJWK(t)
	server, issuer, discoveryURL := newTestOIDCServer(t, publicJWK)
	defer server.Close()

	provider, err := newExternalOIDCAuthProvider(AuthOptions{
		Mode:        AuthModeExternalOIDC,
		ResourceURL: "https://mcp.example.com/mcp",
		External: ExternalOIDCOptions{
			IssuerURL:      issuer,
			DiscoveryURL:   discoveryURL,
			Audience:       "mcp-resource",
			RequiredScopes: []string{"mcp:invoke"},
		},
	})
	if err != nil {
		t.Fatalf("newExternalOIDCAuthProvider() error = %v", err)
	}

	now := time.Now()
	tests := []struct {
		name  string
		token string
	}{
		{name: "malformed", token: "not-a-jwt"},
		{name: "wrong audience", token: signExternalTestTokenWithTimes(t, privateKey, issuer, "another-resource", "client-1", "mcp:invoke", now.Add(-time.Minute), now.Add(time.Hour))},
		{name: "expired", token: signExternalTestTokenWithTimes(t, privateKey, issuer, "mcp-resource", "client-1", "mcp:invoke", now.Add(-2*time.Hour), now.Add(-time.Hour))},
		{name: "missing scope", token: signExternalTestTokenWithTimes(t, privateKey, issuer, "mcp-resource", "client-1", "openid profile", now.Add(-time.Minute), now.Add(time.Hour))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := provider.ValidateBearerToken(context.Background(), tt.token); err == nil {
				t.Fatalf("expected %s validation error", tt.name)
			}
		})
	}
}

func TestOfficialExternalOIDCMiddlewarePropagatesPrincipalAndChallenges(t *testing.T) {
	privateKey, publicJWK := generateTestJWK(t)
	server, issuer, discoveryURL := newTestOIDCServer(t, publicJWK)
	defer server.Close()

	cfg := NewServerConfig()
	cfg.authEnabled = true
	cfg.authOptions = AuthOptions{
		Mode:        AuthModeExternalOIDC,
		ResourceURL: "https://mcp.example.com/mcp",
		External: ExternalOIDCOptions{
			IssuerURL:      issuer,
			DiscoveryURL:   discoveryURL,
			Audience:       "mcp-resource",
			RequiredScopes: []string{"mcp:invoke"},
		},
	}
	mux := http.NewServeMux()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := GetAuthPrincipal(r.Context())
		if !ok || principal.Subject != "alice" || principal.ClientID != "client-1" {
			t.Fatalf("principal = %#v, ok=%v", principal, ok)
		}
		tokenInfo := mcpauth.TokenInfoFromContext(r.Context())
		if tokenInfo == nil || tokenInfo.UserID != "alice" || tokenInfo.Expiration.IsZero() {
			t.Fatalf("token info = %#v", tokenInfo)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	wrapped, err := wrapHTTPAuthentication(mux, cfg, next)
	if err != nil {
		t.Fatal(err)
	}
	mux.Handle("/mcp", wrapped)

	t.Run("valid", func(t *testing.T) {
		token := signExternalTestToken(t, privateKey, issuer, "mcp-resource", "client-1", "mcp:invoke")
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing bearer", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", rec.Code)
		}
		if challenge := rec.Header().Get("WWW-Authenticate"); !strings.Contains(challenge, `resource_metadata="https://mcp.example.com/.well-known/oauth-protected-resource"`) {
			t.Fatalf("challenge = %q", challenge)
		}
	})

	t.Run("insufficient scope", func(t *testing.T) {
		token := signExternalTestToken(t, privateKey, issuer, "mcp-resource", "client-1", "openid")
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if challenge := rec.Header().Get("WWW-Authenticate"); !strings.Contains(challenge, `scope="mcp:invoke"`) {
			t.Fatalf("challenge = %q", challenge)
		}
	})

	t.Run("protected resource metadata", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil))
		if rec.Code != http.StatusOK || rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("status=%d cors=%q body=%s", rec.Code, rec.Header().Get("Access-Control-Allow-Origin"), rec.Body.String())
		}
	})
}

func TestExternalOIDCProviderRequiresExplicitResourceURL(t *testing.T) {
	privateKey, publicJWK := generateTestJWK(t)
	server, issuer, discoveryURL := newTestOIDCServer(t, publicJWK)
	defer server.Close()

	_, _ = privateKey, issuer

	_, err := newExternalOIDCAuthProvider(AuthOptions{
		Mode: AuthModeExternalOIDC,
		External: ExternalOIDCOptions{
			IssuerURL:    issuer,
			DiscoveryURL: discoveryURL,
		},
	})
	if err == nil {
		t.Fatalf("expected missing resource url error")
	}
	if !strings.Contains(err.Error(), "resource url") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func generateTestJWK(t *testing.T) (*rsa.PrivateKey, jose.JSONWebKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey(): %v", err)
	}

	return privateKey, jose.JSONWebKey{
		Key:       &privateKey.PublicKey,
		KeyID:     "kid-1",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}
}

func signExternalTestToken(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience, clientID, scope string) string {
	t.Helper()
	now := time.Now()
	return signExternalTestTokenWithTimes(t, privateKey, issuer, audience, clientID, scope, now.Add(-time.Minute), now.Add(time.Hour))
}

func signExternalTestTokenWithTimes(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience, clientID, scope string, notBefore, expiry time.Time) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: privateKey},
		(&jose.SignerOptions{}).WithHeader(jose.HeaderKey("kid"), "kid-1"),
	)
	if err != nil {
		t.Fatalf("jose.NewSigner(): %v", err)
	}

	now := time.Now()
	token, err := jwt.Signed(signer).
		Claims(jwt.Claims{
			Issuer:    issuer,
			Subject:   "alice",
			Audience:  jwt.Audience{audience},
			Expiry:    jwt.NewNumericDate(expiry),
			NotBefore: jwt.NewNumericDate(notBefore),
			IssuedAt:  jwt.NewNumericDate(now),
		}).
		Claims(map[string]any{
			"azp":                clientID,
			"scope":              scope,
			"email":              "alice@example.com",
			"email_verified":     true,
			"preferred_username": "alice",
			"name":               "Alice Example",
			"picture":            "https://example.com/alice.png",
		}).
		CompactSerialize()
	if err != nil {
		t.Fatalf("CompactSerialize(): %v", err)
	}

	return token
}

func newTestOIDCServer(t *testing.T, publicJWK jose.JSONWebKey) (*httptest.Server, string, string) {
	t.Helper()

	var issuer string
	jwksBody, err := json.Marshal(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{publicJWK}})
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"issuer":                 issuer,
				"jwks_uri":               server.URL + "/jwks.json",
				"authorization_endpoint": issuer + "/protocol/openid-connect/auth",
				"token_endpoint":         issuer + "/protocol/openid-connect/token",
			})
		case "/jwks.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(jwksBody)
		default:
			http.NotFound(w, r)
		}
	}))

	issuer = server.URL + "/realms/test"
	return server, issuer, server.URL + "/.well-known/openid-configuration"
}
