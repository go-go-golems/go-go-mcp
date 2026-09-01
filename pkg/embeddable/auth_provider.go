package embeddable

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	embeddedoidc "github.com/go-go-golems/go-go-mcp/pkg/auth/oidc"
)

var errUnauthorizedToken = errors.New("unauthorized token")

type AuthPrincipal struct {
	Subject           string
	ClientID          string
	Issuer            string
	Scopes            []string
	Email             string
	EmailVerified     bool
	PreferredUsername string
	DisplayName       string
	AvatarURL         string
	Expiration        time.Time
	Claims            map[string]any
}

// HTTPAuthVerifier is the MCP resource-server boundary. Authorization-server
// routes belong to the application composition root, not to bearer middleware.
type HTTPAuthVerifier interface {
	ValidateBearerToken(ctx context.Context, token string) (AuthPrincipal, error)
	ProtectedResourceMetadata() map[string]any
	WWWAuthenticateHeader() string
}

type httpAuthRuntime struct {
	provider                 HTTPAuthVerifier
	mountAuthorizationServer func(*http.ServeMux)
}

func newHTTPAuthRuntime(cfg *ServerConfig) (httpAuthRuntime, error) {
	if cfg == nil || !cfg.authEnabled {
		return httpAuthRuntime{}, nil
	}
	if cfg.customAuthVerifier != nil {
		return httpAuthRuntime{provider: cfg.customAuthVerifier}, nil
	}
	if !cfg.authOptions.Enabled() {
		return httpAuthRuntime{}, nil
	}

	switch cfg.authOptions.Mode {
	case AuthModeNone:
		return httpAuthRuntime{}, nil
	case AuthModeEmbeddedDev:
		provider, err := newEmbeddedDevAuthProvider(cfg.authOptions)
		if err != nil {
			return httpAuthRuntime{}, err
		}
		return httpAuthRuntime{provider: provider, mountAuthorizationServer: provider.MountAuthorizationServer}, nil
	case AuthModeExternalOIDC:
		provider, err := newExternalOIDCAuthProvider(cfg.authOptions)
		if err != nil {
			return httpAuthRuntime{}, err
		}
		return httpAuthRuntime{provider: provider}, nil
	default:
		return httpAuthRuntime{}, fmt.Errorf("unsupported auth mode: %s", cfg.authOptions.Mode)
	}
}

type embeddedDevAuthProvider struct {
	server      *embeddedoidc.Server
	opts        EmbeddedOIDCOptions
	resourceURL string
}

func newEmbeddedDevAuthProvider(opts AuthOptions) (*embeddedDevAuthProvider, error) {
	srv, err := embeddedoidc.New(embeddedoidc.Config{
		Issuer:          opts.Embedded.Issuer,
		DBPath:          opts.Embedded.DBPath,
		EnableDevTokens: opts.Embedded.EnableDevTokens,
		User:            opts.Embedded.User,
		Pass:            opts.Embedded.Pass,
	})
	if err != nil {
		return nil, err
	}

	return &embeddedDevAuthProvider{
		server:      srv,
		opts:        opts.Embedded,
		resourceURL: opts.EffectiveResourceURL(),
	}, nil
}

func (p *embeddedDevAuthProvider) MountAuthorizationServer(mux *http.ServeMux) {
	p.server.Routes(mux)
}

func (p *embeddedDevAuthProvider) ValidateBearerToken(ctx context.Context, token string) (AuthPrincipal, error) {
	if p.opts.AuthKey != "" && token == p.opts.AuthKey {
		return AuthPrincipal{
			Subject:  "static-key-user",
			ClientID: "static-key-client",
			Issuer:   p.opts.Issuer,
		}, nil
	}

	subj, cid, ok, err := p.server.IntrospectAccessToken(ctx, token)
	if err != nil {
		return AuthPrincipal{}, err
	}
	if !ok {
		return AuthPrincipal{}, errUnauthorizedToken
	}

	return AuthPrincipal{
		Subject:  subj,
		ClientID: cid,
		Issuer:   p.opts.Issuer,
	}, nil
}

func (p *embeddedDevAuthProvider) ProtectedResourceMetadata() map[string]any {
	return map[string]any{
		"authorization_servers": []string{p.opts.Issuer},
		"resource":              p.resourceURL,
	}
}

func (p *embeddedDevAuthProvider) WWWAuthenticateHeader() string {
	return buildBearerChallenge(protectedResourceMetadataURL(p.resourceURL))
}

func buildBearerChallenge(resourceMetadataURL string) string {
	return "Bearer realm=\"mcp\", resource_metadata=\"" + resourceMetadataURL + "\""
}
