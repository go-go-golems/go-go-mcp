package embeddable

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

const (
	defaultOIDCHTTPTimeout = 10 * time.Second
	defaultJWKSRefresh     = 5 * time.Minute
)

type oidcDiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	JWKSURI               string `json:"jwks_uri"`
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string `json:"token_endpoint,omitempty"`
}

type externalBearerClaims struct {
	jwt.Claims
	AuthorizedParty string   `json:"azp,omitempty"`
	ClientID        string   `json:"client_id,omitempty"`
	Scope           string   `json:"scope,omitempty"`
	SCP             []string `json:"scp,omitempty"`
}

type externalOIDCAuthProvider struct {
	opts           ExternalOIDCOptions
	resourceURL    string
	discoveryURL   string
	discovery      oidcDiscoveryDocument
	httpClient     *http.Client
	jwks           *jwksCache
	requiredScopes map[string]struct{}
}

type jwksCache struct {
	mu              sync.RWMutex
	client          *http.Client
	jwksURI         string
	refreshInterval time.Duration
	lastFetched     time.Time
	keySet          jose.JSONWebKeySet
}

func newExternalOIDCAuthProvider(opts AuthOptions) (*externalOIDCAuthProvider, error) {
	issuerURL := strings.TrimSpace(opts.External.IssuerURL)
	if issuerURL == "" {
		return nil, fmt.Errorf("external oidc issuer url is required")
	}

	discoveryURL := strings.TrimSpace(opts.External.DiscoveryURL)
	if discoveryURL == "" {
		discoveryURL = strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"
	}

	httpClient := &http.Client{Timeout: defaultOIDCHTTPTimeout}
	discovery, err := fetchOIDCDiscovery(context.Background(), httpClient, discoveryURL)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(discovery.Issuer) == "" {
		return nil, fmt.Errorf("oidc discovery document missing issuer")
	}
	if strings.TrimSpace(discovery.JWKSURI) == "" {
		return nil, fmt.Errorf("oidc discovery document missing jwks_uri")
	}

	requiredScopes := make(map[string]struct{}, len(opts.External.RequiredScopes))
	for _, scope := range opts.External.RequiredScopes {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			requiredScopes[scope] = struct{}{}
		}
	}

	provider := &externalOIDCAuthProvider{
		opts:           opts.External,
		resourceURL:    opts.EffectiveResourceURL(),
		discoveryURL:   discoveryURL,
		discovery:      discovery,
		httpClient:     httpClient,
		requiredScopes: requiredScopes,
		jwks: &jwksCache{
			client:          httpClient,
			jwksURI:         discovery.JWKSURI,
			refreshInterval: defaultJWKSRefresh,
		},
	}

	return provider, nil
}

func (p *externalOIDCAuthProvider) MountRoutes(mux *http.ServeMux) {}

func (p *externalOIDCAuthProvider) ValidateBearerToken(ctx context.Context, token string) (AuthPrincipal, error) {
	claims, err := p.verifyToken(ctx, token, false)
	if err != nil {
		claims, err = p.verifyToken(ctx, token, true)
	}
	if err != nil {
		return AuthPrincipal{}, err
	}

	expected := jwt.Expected{
		Issuer: p.discovery.Issuer,
		Time:   time.Now(),
	}
	if audience := strings.TrimSpace(p.opts.Audience); audience != "" {
		expected.Audience = jwt.Audience{audience}
	}
	if err := claims.Claims.Validate(expected); err != nil {
		return AuthPrincipal{}, err
	}

	scopeSet := parseScopeSet(claims.Scope, claims.SCP)
	for scope := range p.requiredScopes {
		if _, ok := scopeSet[scope]; !ok {
			return AuthPrincipal{}, fmt.Errorf("%w: missing required scope %q", errUnauthorizedToken, scope)
		}
	}

	clientID := claims.AuthorizedParty
	if clientID == "" {
		clientID = claims.ClientID
	}
	if clientID == "" && len(claims.Audience) > 0 {
		clientID = claims.Audience[0]
	}

	return AuthPrincipal{
		Subject:  claims.Subject,
		ClientID: clientID,
		Issuer:   claims.Issuer,
		Scopes:   sortedScopeKeys(scopeSet),
	}, nil
}

func (p *externalOIDCAuthProvider) ProtectedResourceMetadata() map[string]any {
	return map[string]any{
		"authorization_servers": []string{p.discovery.Issuer},
		"resource":              p.resourceURL,
	}
}

func (p *externalOIDCAuthProvider) WWWAuthenticateHeader() string {
	return "Bearer realm=\"mcp\", resource=\"" + p.resourceURL + "\"" +
		", authorization_uri=\"" + p.discoveryURL + "\", resource_metadata=\"" + protectedResourceMetadataURL(p.resourceURL) + "\""
}

func (p *externalOIDCAuthProvider) verifyToken(ctx context.Context, token string, forceRefresh bool) (*externalBearerClaims, error) {
	parsed, err := jwt.ParseSigned(token)
	if err != nil {
		return nil, err
	}

	kid := ""
	if len(parsed.Headers) > 0 {
		kid = parsed.Headers[0].KeyID
	}

	keys, err := p.jwks.Keys(ctx, kid, forceRefresh)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, key := range keys {
		var claims externalBearerClaims
		if err := parsed.Claims(key.Key, &claims); err == nil {
			return &claims, nil
		} else {
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = errUnauthorizedToken
	}
	return nil, lastErr
}

func (c *jwksCache) Keys(ctx context.Context, kid string, forceRefresh bool) ([]jose.JSONWebKey, error) {
	if forceRefresh || c.needsRefresh() {
		if err := c.refresh(ctx); err != nil {
			return nil, err
		}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if kid != "" {
		if keys := c.keySet.Key(kid); len(keys) > 0 {
			return keys, nil
		}
	}

	if len(c.keySet.Keys) == 0 {
		return nil, fmt.Errorf("jwks key set is empty")
	}

	return append([]jose.JSONWebKey(nil), c.keySet.Keys...), nil
}

func (c *jwksCache) needsRefresh() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.keySet.Keys) == 0 || time.Since(c.lastFetched) >= c.refreshInterval
}

func (c *jwksCache) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.jwksURI, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks request failed with status %d", resp.StatusCode)
	}

	var keySet jose.JSONWebKeySet
	if err := json.NewDecoder(resp.Body).Decode(&keySet); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.keySet = keySet
	c.lastFetched = time.Now()
	return nil
}

func fetchOIDCDiscovery(ctx context.Context, client *http.Client, discoveryURL string) (oidcDiscoveryDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return oidcDiscoveryDocument{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return oidcDiscoveryDocument{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return oidcDiscoveryDocument{}, fmt.Errorf("oidc discovery request failed with status %d", resp.StatusCode)
	}

	var doc oidcDiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return oidcDiscoveryDocument{}, err
	}
	return doc, nil
}

func protectedResourceMetadataURL(resourceURL string) string {
	u, err := url.Parse(resourceURL)
	if err != nil {
		return strings.TrimRight(resourceURL, "/") + "/.well-known/oauth-protected-resource"
	}

	u.Path = "/.well-known/oauth-protected-resource"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func parseScopeSet(scope string, scp []string) map[string]struct{} {
	ret := make(map[string]struct{})
	for _, part := range strings.Fields(scope) {
		ret[part] = struct{}{}
	}
	for _, part := range scp {
		part = strings.TrimSpace(part)
		if part != "" {
			ret[part] = struct{}{}
		}
	}
	return ret
}

func sortedScopeKeys(scopeSet map[string]struct{}) []string {
	ret := make([]string, 0, len(scopeSet))
	for scope := range scopeSet {
		ret = append(ret, scope)
	}
	if len(ret) > 1 {
		// Sort for deterministic tests and stable logs.
		for i := 0; i < len(ret)-1; i++ {
			for j := i + 1; j < len(ret); j++ {
				if ret[j] < ret[i] {
					ret[i], ret[j] = ret[j], ret[i]
				}
			}
		}
	}
	return ret
}
