package embeddable

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	embeddedoidc "github.com/go-go-golems/go-go-mcp/pkg/auth/oidc"
)

var errUnauthorizedToken = errors.New("unauthorized token")

type AuthPrincipal struct {
	Subject  string
	ClientID string
	Issuer   string
	Scopes   []string
}

type HTTPAuthProvider interface {
	MountRoutes(mux *http.ServeMux)
	ValidateBearerToken(ctx context.Context, token string) (AuthPrincipal, error)
	ProtectedResourceMetadata() map[string]any
	WWWAuthenticateHeader() string
}

func newHTTPAuthProvider(cfg *ServerConfig) (HTTPAuthProvider, error) {
	if cfg == nil || !cfg.authEnabled || !cfg.authOptions.Enabled() {
		return nil, nil
	}

	switch cfg.authOptions.Mode {
	case AuthModeNone:
		return nil, nil
	case AuthModeEmbeddedDev:
		return newEmbeddedDevAuthProvider(cfg.authOptions)
	case AuthModeExternalOIDC:
		return nil, fmt.Errorf("external_oidc auth mode not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported auth mode: %s", cfg.authOptions.Mode)
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

func (p *embeddedDevAuthProvider) MountRoutes(mux *http.ServeMux) {
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
	issuer := strings.TrimRight(p.opts.Issuer, "/")
	asMeta := issuer + "/.well-known/oauth-authorization-server"
	prm := issuer + "/.well-known/oauth-protected-resource"

	return "Bearer realm=\"mcp\", resource=\"" + p.resourceURL + "\"" +
		", authorization_uri=\"" + asMeta + "\", resource_metadata=\"" + prm + "\""
}
