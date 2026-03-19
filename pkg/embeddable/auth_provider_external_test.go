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

func TestExternalOIDCProviderRejectsMissingScope(t *testing.T) {
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

	token := signExternalTestToken(t, privateKey, issuer, "mcp-resource", "client-1", "openid profile")
	if _, err := provider.ValidateBearerToken(context.Background(), token); err == nil {
		t.Fatalf("expected missing-scope validation error")
	}
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
			Expiry:    jwt.NewNumericDate(now.Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-time.Minute)),
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
