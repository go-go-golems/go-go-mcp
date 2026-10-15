// Package embeddeddev links go-go-mcp's embedded development OIDC server into
// an embeddable MCP server, so that --auth-mode embedded_dev works.
//
// The embedded server depends on fosite and SQLite (and, through them, on
// gRPC and OpenTelemetry exporters). It lives in this package so that programs
// which never use embedded_dev do not link any of it:
//
//	embeddable.AddMCPCommand(root,
//		embeddeddev.Provide(),            // --auth-mode embedded_dev available
//		...)
//
//	embeddable.AddMCPCommand(root,
//		embeddeddev.WithOIDC(embeddable.EmbeddedOIDCOptions{Issuer: "http://localhost:3001"}),
//		...)
package embeddeddev

import (
	"context"
	"errors"
	"net/http"
	"time"

	embeddedoidc "github.com/go-go-golems/go-go-mcp/pkg/auth/oidc"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
)

// staticKeyValidity is the expiration reported for the static development
// key. The key has no lifetime of its own and is checked again on every
// request, so the value only states that it was valid when this request was
// authenticated.
const staticKeyValidity = time.Minute

var errUnauthorizedToken = errors.New("unauthorized token")

// Provide links the embedded development OIDC server, making
// --auth-mode embedded_dev (and the legacy --oidc flag) usable.
func Provide() embeddable.ServerOption {
	return embeddable.WithAuthModeProvider(embeddable.AuthModeEmbeddedDev, newAuthMode)
}

// WithOIDC links the embedded development OIDC server and selects it with
// the given options.
func WithOIDC(opts embeddable.EmbeddedOIDCOptions) embeddable.ServerOption {
	provide := Provide()
	auth := embeddable.WithAuth(embeddable.AuthOptions{
		Mode:     embeddable.AuthModeEmbeddedDev,
		Embedded: opts,
	})
	return func(cfg *embeddable.ServerConfig) error {
		if err := provide(cfg); err != nil {
			return err
		}
		return auth(cfg)
	}
}

func newAuthMode(opts embeddable.AuthOptions) (embeddable.HTTPAuthVerifier, func(*http.ServeMux), error) {
	p, err := newProvider(opts)
	if err != nil {
		return nil, nil, err
	}
	return p, p.server.Routes, nil
}

type provider struct {
	server      *embeddedoidc.Server
	opts        embeddable.EmbeddedOIDCOptions
	resourceURL string
}

var _ embeddable.HTTPAuthVerifier = (*provider)(nil)

func newProvider(opts embeddable.AuthOptions) (*provider, error) {
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
	return &provider{
		server:      srv,
		opts:        opts.Embedded,
		resourceURL: opts.EffectiveResourceURL(),
	}, nil
}

func (p *provider) ValidateBearerToken(ctx context.Context, token string) (embeddable.AuthPrincipal, error) {
	if p.opts.AuthKey != "" && token == p.opts.AuthKey {
		return embeddable.AuthPrincipal{
			Subject:    "static-key-user",
			ClientID:   "static-key-client",
			Issuer:     p.opts.Issuer,
			Expiration: time.Now().Add(staticKeyValidity),
		}, nil
	}

	info, ok, err := p.server.IntrospectAccessToken(ctx, token)
	if err != nil {
		return embeddable.AuthPrincipal{}, err
	}
	if !ok {
		return embeddable.AuthPrincipal{}, errUnauthorizedToken
	}

	return embeddable.AuthPrincipal{
		Subject:    info.Subject,
		ClientID:   info.ClientID,
		Issuer:     p.opts.Issuer,
		Expiration: info.ExpiresAt,
	}, nil
}

func (p *provider) ProtectedResourceMetadata() map[string]any {
	return map[string]any{
		"authorization_servers": []string{p.opts.Issuer},
		"resource":              p.resourceURL,
	}
}

func (p *provider) WWWAuthenticateHeader() string {
	return embeddable.BearerChallenge(embeddable.ProtectedResourceMetadataURL(p.resourceURL))
}
