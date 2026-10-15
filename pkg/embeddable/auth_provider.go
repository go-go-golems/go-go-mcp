package embeddable

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var errUnauthorizedToken = errors.New("unauthorized token")

// errEmbeddedDevNotLinked is returned when embedded_dev is selected but the
// program did not link the embedded authorization server. Keeping that server
// out of this package keeps fosite, SQLite and their dependencies out of every
// program that imports embeddable.
var errEmbeddedDevNotLinked = errors.New("auth mode embedded_dev is not linked into this program; add embeddeddev.Provide() from github.com/go-go-golems/go-go-mcp/pkg/embeddable/embeddeddev to the server options")

type AuthPrincipal struct {
	Subject           string
	ClientID          string
	Issuer            string
	Scopes            ScopeSet
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

// AuthModeProvider builds the resource-server verifier for an auth mode. When
// the mode runs its own authorization server, it also returns a function that
// mounts that server's routes next to the MCP endpoint; otherwise nil.
type AuthModeProvider func(opts AuthOptions) (HTTPAuthVerifier, func(*http.ServeMux), error)

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

	if linked, ok := cfg.authModeProviders[cfg.authOptions.Mode]; ok {
		provider, mount, err := linked(cfg.authOptions)
		if err != nil {
			return httpAuthRuntime{}, err
		}
		return httpAuthRuntime{provider: provider, mountAuthorizationServer: mount}, nil
	}

	switch cfg.authOptions.Mode {
	case AuthModeNone:
		return httpAuthRuntime{}, nil
	case AuthModeEmbeddedDev:
		return httpAuthRuntime{}, errEmbeddedDevNotLinked
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

// BearerChallenge returns the WWW-Authenticate value that points MCP clients
// to the protected resource metadata (RFC 9728) at resourceMetadataURL.
func BearerChallenge(resourceMetadataURL string) string {
	return "Bearer realm=\"mcp\", resource_metadata=\"" + resourceMetadataURL + "\""
}
