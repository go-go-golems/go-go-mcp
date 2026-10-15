package embeddable

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
)

const (
	testIntrospectionClientID     = "mcp resource"
	testIntrospectionClientSecret = "s3cr:et&"
)

// fakeIntrospectionServer is an identity provider with discovery, a JWKS and
// an RFC 7662 introspection endpoint. Responses are looked up by token.
type fakeIntrospectionServer struct {
	*httptest.Server
	issuer       string
	discoveryURL string

	mu         sync.Mutex
	responses  map[string]any
	status     int
	rawBody    string
	calls      []introspectionCall
	noEndpoint bool
}

type introspectionCall struct {
	user, pass string
	form       map[string]string
}

func newFakeIntrospectionServer(t *testing.T, publicJWK jose.JSONWebKey) *fakeIntrospectionServer {
	t.Helper()
	jwksBody, err := json.Marshal(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{publicJWK}})
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeIntrospectionServer{responses: map[string]any{}, status: http.StatusOK}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			doc := map[string]any{"issuer": f.issuer, "jwks_uri": f.URL + "/jwks.json"}
			f.mu.Lock()
			if !f.noEndpoint {
				doc["introspection_endpoint"] = f.URL + "/oauth/introspect"
			}
			f.mu.Unlock()
			_ = json.NewEncoder(w).Encode(doc)
		case "/jwks.json":
			_, _ = w.Write(jwksBody)
		case "/oauth/introspect":
			if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			user, pass, _ := r.BasicAuth()
			if err := r.ParseForm(); err != nil {
				http.Error(w, "bad form", http.StatusBadRequest)
				return
			}
			f.mu.Lock()
			defer f.mu.Unlock()
			f.calls = append(f.calls, introspectionCall{user: user, pass: pass, form: map[string]string{
				"token":           r.PostForm.Get("token"),
				"token_type_hint": r.PostForm.Get("token_type_hint"),
			}})
			if f.status != http.StatusOK {
				http.Error(w, "unavailable", f.status)
				return
			}
			if f.rawBody != "" {
				_, _ = w.Write([]byte(f.rawBody))
				return
			}
			resp, ok := f.responses[r.PostForm.Get("token")]
			if !ok {
				resp = map[string]any{"active": false}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	f.issuer = f.URL + "/realms/test"
	f.discoveryURL = f.URL + "/.well-known/openid-configuration"
	t.Cleanup(f.Close)
	return f
}

func (f *fakeIntrospectionServer) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func (f *fakeIntrospectionServer) activeResponse(extra map[string]any) map[string]any {
	resp := map[string]any{
		"active":             true,
		"iss":                f.issuer,
		"sub":                "alice",
		"client_id":          "claude-connector",
		"scope":              "openid mcp:invoke",
		"aud":                []string{"mcp-resource", "other"},
		"exp":                time.Now().Add(time.Hour).Unix(),
		"token_type":         "Bearer",
		"email":              "alice@example.com",
		"email_verified":     true,
		"preferred_username": "alice",
		"name":               "Alice Example",
	}
	for k, v := range extra {
		if v == nil {
			delete(resp, k)
		} else {
			resp[k] = v
		}
	}
	return resp
}

func introspectionOptions(f *fakeIntrospectionServer, validation TokenValidation) AuthOptions {
	return AuthOptions{
		Mode:        AuthModeExternalOIDC,
		ResourceURL: "https://mcp.example.com/mcp",
		External: ExternalOIDCOptions{
			IssuerURL:                 f.issuer,
			DiscoveryURL:              f.discoveryURL,
			Audience:                  "mcp-resource",
			RequiredScopes:            []string{"mcp:invoke"},
			TokenValidation:           validation,
			IntrospectionClientID:     testIntrospectionClientID,
			IntrospectionClientSecret: testIntrospectionClientSecret,
		},
	}
}

func TestIntrospectionAcceptsActiveOpaqueToken(t *testing.T) {
	_, publicJWK := generateTestJWK(t)
	f := newFakeIntrospectionServer(t, publicJWK)
	f.responses["opaque-1"] = f.activeResponse(nil)

	provider, err := newExternalOIDCAuthProvider(introspectionOptions(f, TokenValidationIntrospection))
	if err != nil {
		t.Fatalf("newExternalOIDCAuthProvider() error = %v", err)
	}
	principal, err := provider.ValidateBearerToken(context.Background(), "opaque-1")
	if err != nil {
		t.Fatalf("ValidateBearerToken() error = %v", err)
	}

	if principal.Subject != "alice" || principal.ClientID != "claude-connector" || principal.Issuer != f.issuer {
		t.Fatalf("unexpected principal identity: %#v", principal)
	}
	if principal.Email != "alice@example.com" || !principal.EmailVerified || principal.PreferredUsername != "alice" || principal.DisplayName != "Alice Example" {
		t.Fatalf("unexpected principal profile: %#v", principal)
	}
	if got := strings.Join(principal.Scopes.Strings(), " "); got != "mcp:invoke openid" {
		t.Fatalf("scopes = %q", got)
	}
	if time.Until(principal.Expiration) < 59*time.Minute {
		t.Fatalf("expiration = %v, want about an hour from now", principal.Expiration)
	}

	if f.callCount() != 1 {
		t.Fatalf("introspection calls = %d, want 1", f.callCount())
	}
	call := f.calls[0]
	// RFC 6749 §2.3.1: credentials are form-encoded before Basic encoding.
	if call.user != "mcp+resource" || call.pass != "s3cr%3Aet%26" {
		t.Fatalf("basic credentials = %q / %q", call.user, call.pass)
	}
	if call.form["token"] != "opaque-1" || call.form["token_type_hint"] != "access_token" {
		t.Fatalf("form = %#v", call.form)
	}
}

func TestIntrospectionRejectsTokensItMustNotAccept(t *testing.T) {
	_, publicJWK := generateTestJWK(t)
	f := newFakeIntrospectionServer(t, publicJWK)
	provider, err := newExternalOIDCAuthProvider(introspectionOptions(f, TokenValidationIntrospection))
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]map[string]any{
		"inactive":        {"active": false},
		"wrong issuer":    f.activeResponse(map[string]any{"iss": "https://evil.example.com"}),
		"expired":         f.activeResponse(map[string]any{"exp": time.Now().Add(-time.Minute).Unix()}),
		"missing exp":     f.activeResponse(map[string]any{"exp": nil}),
		"not yet valid":   f.activeResponse(map[string]any{"nbf": time.Now().Add(time.Hour).Unix()}),
		"wrong audience":  f.activeResponse(map[string]any{"aud": "someone-else"}),
		"refresh token":   f.activeResponse(map[string]any{"token_type": "refresh_token"}),
		"missing scope":   f.activeResponse(map[string]any{"scope": "openid"}),
		"unknown token":   nil,
		"string aud fail": f.activeResponse(map[string]any{"aud": "other"}),
	}
	for name, resp := range cases {
		t.Run(name, func(t *testing.T) {
			token := "tok-" + strings.ReplaceAll(name, " ", "-")
			if resp != nil {
				f.responses[token] = resp
			}
			_, err := provider.ValidateBearerToken(context.Background(), token)
			if !errors.Is(err, errUnauthorizedToken) {
				t.Fatalf("ValidateBearerToken() error = %v, want errUnauthorizedToken", err)
			}
		})
	}

	t.Run("single string audience accepted", func(t *testing.T) {
		f.responses["tok-string-aud"] = f.activeResponse(map[string]any{"aud": "mcp-resource"})
		if _, err := provider.ValidateBearerToken(context.Background(), "tok-string-aud"); err != nil {
			t.Fatalf("ValidateBearerToken() error = %v", err)
		}
	})
}

func TestIntrospectionReportsAuthorizationServerFailures(t *testing.T) {
	_, publicJWK := generateTestJWK(t)
	f := newFakeIntrospectionServer(t, publicJWK)
	provider, err := newExternalOIDCAuthProvider(introspectionOptions(f, TokenValidationIntrospection))
	if err != nil {
		t.Fatal(err)
	}

	f.status = http.StatusInternalServerError
	if _, err := provider.ValidateBearerToken(context.Background(), "any"); !errors.Is(err, errAuthorizationServerUnavailable) {
		t.Fatalf("status 500: error = %v, want errAuthorizationServerUnavailable", err)
	}

	f.status = http.StatusOK
	f.rawBody = "not json"
	if _, err := provider.ValidateBearerToken(context.Background(), "any"); !errors.Is(err, errAuthorizationServerUnavailable) {
		t.Fatalf("invalid JSON: error = %v, want errAuthorizationServerUnavailable", err)
	}
}

func TestAutoValidationRoutesByTokenShape(t *testing.T) {
	privateKey, publicJWK := generateTestJWK(t)
	f := newFakeIntrospectionServer(t, publicJWK)
	provider, err := newExternalOIDCAuthProvider(introspectionOptions(f, TokenValidationAuto))
	if err != nil {
		t.Fatal(err)
	}

	jwtToken := signExternalTestToken(t, privateKey, f.issuer, "mcp-resource", "client-1", "mcp:invoke")
	if _, err := provider.ValidateBearerToken(context.Background(), jwtToken); err != nil {
		t.Fatalf("JWT: error = %v", err)
	}
	if f.callCount() != 0 {
		t.Fatalf("a valid JWT was sent to introspection (%d calls)", f.callCount())
	}

	otherKey, _ := generateTestJWK(t)
	forged := signExternalTestToken(t, otherKey, f.issuer, "mcp-resource", "client-1", "mcp:invoke")
	if _, err := provider.ValidateBearerToken(context.Background(), forged); err == nil {
		t.Fatal("forged JWT accepted")
	}
	if f.callCount() != 0 {
		t.Fatalf("a forged JWT was sent to introspection (%d calls)", f.callCount())
	}

	f.responses["opaque-2"] = f.activeResponse(nil)
	if _, err := provider.ValidateBearerToken(context.Background(), "opaque-2"); err != nil {
		t.Fatalf("opaque: error = %v", err)
	}
	if f.callCount() != 1 {
		t.Fatalf("opaque token introspection calls = %d, want 1", f.callCount())
	}
}

func TestLooksLikeJWS(t *testing.T) {
	for token, want := range map[string]bool{
		"eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhIn0.c2ln": true,
		"eyJhbGciOiJub25lIn0.eyJzdWIiOiJhIn0.":      true,
		"opaque-token-without-dots":                 false,
		"a.b":                                       false,
		"a.b.c.d":                                   false,
		".b.c":                                      false,
		"a b.c.d":                                   false,
		"V1StGXR8_Z5jdHi6B-myT":                     false,
	} {
		if got := looksLikeJWS(token); got != want {
			t.Errorf("looksLikeJWS(%q) = %v, want %v", token, got, want)
		}
	}
}

func TestIntrospectionConfigurationIsCheckedAtStartup(t *testing.T) {
	_, publicJWK := generateTestJWK(t)
	f := newFakeIntrospectionServer(t, publicJWK)

	t.Run("unknown validation", func(t *testing.T) {
		opts := introspectionOptions(f, "magic")
		if _, err := newExternalOIDCAuthProvider(opts); err == nil || !strings.Contains(err.Error(), "unsupported token validation") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("missing credentials", func(t *testing.T) {
		opts := introspectionOptions(f, TokenValidationIntrospection)
		opts.External.IntrospectionClientSecret = ""
		if _, err := newExternalOIDCAuthProvider(opts); err == nil || !strings.Contains(err.Error(), "client id and secret") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("insecure endpoint", func(t *testing.T) {
		opts := introspectionOptions(f, TokenValidationIntrospection)
		opts.External.IntrospectionURL = "http://idp.example.com/introspect"
		if _, err := newExternalOIDCAuthProvider(opts); err == nil || !strings.Contains(err.Error(), "invalid introspection url") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("no endpoint anywhere", func(t *testing.T) {
		f.mu.Lock()
		f.noEndpoint = true
		f.mu.Unlock()
		defer func() { f.mu.Lock(); f.noEndpoint = false; f.mu.Unlock() }()
		if _, err := newExternalOIDCAuthProvider(introspectionOptions(f, TokenValidationAuto)); err == nil || !strings.Contains(err.Error(), "introspection endpoint") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("jwt needs nothing", func(t *testing.T) {
		opts := introspectionOptions(f, TokenValidationJWT)
		opts.External.IntrospectionClientID, opts.External.IntrospectionClientSecret = "", ""
		if _, err := newExternalOIDCAuthProvider(opts); err != nil {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestIntrospectionThroughHTTPMiddleware(t *testing.T) {
	_, publicJWK := generateTestJWK(t)
	f := newFakeIntrospectionServer(t, publicJWK)
	f.responses["opaque-ok"] = f.activeResponse(nil)

	cfg := NewServerConfig()
	if err := WithAuth(introspectionOptions(f, TokenValidationIntrospection))(cfg); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	wrapped, err := wrapHTTPAuthentication(mux, cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := GetAuthPrincipal(r.Context())
		if !ok || principal.Subject != "alice" {
			t.Fatalf("principal = %#v, ok=%v", principal, ok)
		}
		if info := mcpauth.TokenInfoFromContext(r.Context()); info == nil || info.Expiration.IsZero() {
			t.Fatalf("token info = %#v", info)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	if err != nil {
		t.Fatal(err)
	}
	mux.Handle("/mcp", wrapped)

	send := func(token string) int {
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Code
	}
	if code := send("opaque-ok"); code != http.StatusNoContent {
		t.Fatalf("active token: status %d", code)
	}
	if code := send("opaque-unknown"); code != http.StatusUnauthorized {
		t.Fatalf("inactive token: status %d, want 401", code)
	}
	f.status = http.StatusServiceUnavailable
	if code := send("opaque-ok"); code != http.StatusInternalServerError {
		t.Fatalf("identity provider down: status %d, want 500", code)
	}
}

func TestMCPStartIntrospectionFlags(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(secretFile, []byte("from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stop := errors.New("stop after CLI auth configuration")
	var cfg *ServerConfig
	cmd := NewMCPCommand(
		func(config *ServerConfig) error { cfg = config; return nil },
		WithHooks(&Hooks{OnServerStart: func(context.Context) error { return stop }}),
	)
	cmd.SetArgs([]string{"start",
		"--auth-mode", "external_oidc",
		"--auth-resource-url", "https://mcp.example.com/mcp",
		"--oidc-issuer-url", "https://idp.example.com",
		"--oidc-token-validation", "auto",
		"--oidc-introspection-url", "https://idp.example.com/oauth/v2/introspect",
		"--oidc-introspection-client-id", "mcp-resource",
		"--oidc-introspection-client-secret-file", secretFile,
	})
	if err := cmd.Execute(); !errors.Is(err, stop) {
		t.Fatalf("Execute() error = %v, want %v", err, stop)
	}
	ext := cfg.authOptions.External
	if ext.TokenValidation != TokenValidationAuto || ext.IntrospectionURL != "https://idp.example.com/oauth/v2/introspect" ||
		ext.IntrospectionClientID != "mcp-resource" || ext.IntrospectionClientSecret != "from-file" {
		t.Fatalf("introspection flags not applied: %+v", ext)
	}
}
