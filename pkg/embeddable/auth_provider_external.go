package embeddable

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

const (
	defaultOIDCHTTPTimeout   = 10 * time.Second
	defaultJWKSRefresh       = 5 * time.Minute
	maxIntrospectionResponse = 1 << 20
)

// errAuthorizationServerUnavailable marks failures to obtain an answer from
// the identity provider (network errors, non-200 responses, malformed bodies),
// as opposed to an answer that the token is not valid. The HTTP middleware
// reports these as server errors instead of asking the client to sign in again.
var errAuthorizationServerUnavailable = errors.New("authorization server unavailable")

type oidcDiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	JWKSURI               string `json:"jwks_uri"`
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string `json:"token_endpoint,omitempty"`
	IntrospectionEndpoint string `json:"introspection_endpoint,omitempty"`
}

type externalBearerClaims struct {
	jwt.Claims
	AuthorizedParty   string   `json:"azp,omitempty"`
	ClientID          string   `json:"client_id,omitempty"`
	Scope             string   `json:"scope,omitempty"`
	SCP               []string `json:"scp,omitempty"`
	Email             string   `json:"email,omitempty"`
	EmailVerified     bool     `json:"email_verified,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Name              string   `json:"name,omitempty"`
	Picture           string   `json:"picture,omitempty"`
}

type externalOIDCAuthProvider struct {
	opts             ExternalOIDCOptions
	resourceURL      string
	discovery        oidcDiscoveryDocument
	httpClient       *http.Client
	jwks             *jwksCache
	requiredScopes   map[string]struct{}
	validation       TokenValidation
	introspectionURL string
}

// introspectionResponse is the RFC 7662 §2.2 response, plus the profile claims
// identity providers commonly add.
type introspectionResponse struct {
	Active            bool             `json:"active"`
	Scope             string           `json:"scope,omitempty"`
	ClientID          string           `json:"client_id,omitempty"`
	Username          string           `json:"username,omitempty"`
	TokenType         string           `json:"token_type,omitempty"`
	Typ               string           `json:"typ,omitempty"` // Keycloak: Bearer, Refresh, Offline, ID
	Expiry            *jwt.NumericDate `json:"exp,omitempty"`
	NotBefore         *jwt.NumericDate `json:"nbf,omitempty"`
	Subject           string           `json:"sub,omitempty"`
	Audience          jwt.Audience     `json:"aud,omitempty"`
	Issuer            string           `json:"iss,omitempty"`
	Email             string           `json:"email,omitempty"`
	EmailVerified     bool             `json:"email_verified,omitempty"`
	PreferredUsername string           `json:"preferred_username,omitempty"`
	Name              string           `json:"name,omitempty"`
	Picture           string           `json:"picture,omitempty"`
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
	if _, err := validateExternalOIDCEndpointURL(issuerURL); err != nil {
		return nil, fmt.Errorf("invalid external oidc issuer url: %w", err)
	}

	discoveryURL := strings.TrimSpace(opts.External.DiscoveryURL)
	if discoveryURL == "" {
		discoveryURL = strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"
	}
	if _, err := validateExternalOIDCEndpointURL(discoveryURL); err != nil {
		return nil, fmt.Errorf("invalid external oidc discovery url: %w", err)
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
	if _, err := validateExternalOIDCEndpointURL(discovery.Issuer); err != nil {
		return nil, fmt.Errorf("invalid oidc discovery issuer: %w", err)
	}
	if normalizeURLForCompare(discovery.Issuer) != normalizeURLForCompare(issuerURL) {
		return nil, fmt.Errorf("oidc discovery issuer %q does not match configured issuer %q", discovery.Issuer, issuerURL)
	}
	if _, err := validateExternalOIDCEndpointURL(discovery.JWKSURI); err != nil {
		return nil, fmt.Errorf("invalid jwks uri: %w", err)
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
		resourceURL:    strings.TrimSpace(opts.EffectiveResourceURL()),
		discovery:      discovery,
		httpClient:     httpClient,
		requiredScopes: requiredScopes,
		jwks: &jwksCache{
			client:          httpClient,
			jwksURI:         discovery.JWKSURI,
			refreshInterval: defaultJWKSRefresh,
		},
	}

	if provider.resourceURL == "" {
		return nil, fmt.Errorf("external oidc auth requires resource url")
	}
	if err := provider.configureTokenValidation(); err != nil {
		return nil, err
	}

	return provider, nil
}

// configureTokenValidation checks the introspection settings at startup, so a
// misconfiguration fails the server instead of every request.
func (p *externalOIDCAuthProvider) configureTokenValidation() error {
	p.validation = p.opts.TokenValidation
	switch p.validation {
	case "":
		p.validation = TokenValidationJWT
		return nil
	case TokenValidationJWT:
		return nil
	case TokenValidationIntrospection, TokenValidationAuto:
	default:
		return fmt.Errorf("unsupported token validation %q (want jwt, introspection or auto)", p.validation)
	}

	endpoint := strings.TrimSpace(p.opts.IntrospectionURL)
	if endpoint == "" {
		endpoint = strings.TrimSpace(p.discovery.IntrospectionEndpoint)
	}
	if endpoint == "" {
		return fmt.Errorf("token validation %q needs an introspection endpoint: the discovery document has none and no introspection url is set", p.validation)
	}
	if _, err := validateExternalOIDCEndpointURL(endpoint); err != nil {
		return fmt.Errorf("invalid introspection url: %w", err)
	}
	if strings.TrimSpace(p.opts.IntrospectionClientID) == "" || p.opts.IntrospectionClientSecret == "" {
		return fmt.Errorf("token validation %q needs an introspection client id and secret", p.validation)
	}
	p.introspectionURL = endpoint
	return nil
}

func (p *externalOIDCAuthProvider) ValidateBearerToken(ctx context.Context, token string) (AuthPrincipal, error) {
	return p.validateBearerToken(ctx, token, true)
}

func (p *externalOIDCAuthProvider) validateBearerToken(ctx context.Context, token string, enforceRequiredScopes bool) (AuthPrincipal, error) {
	switch p.validation {
	case TokenValidationIntrospection:
		return p.introspect(ctx, token, enforceRequiredScopes)
	case TokenValidationAuto:
		if !looksLikeJWS(token) {
			return p.introspect(ctx, token, enforceRequiredScopes)
		}
	case TokenValidationJWT:
	}
	return p.validateJWT(ctx, token, enforceRequiredScopes)
}

func (p *externalOIDCAuthProvider) validateJWT(ctx context.Context, token string, enforceRequiredScopes bool) (AuthPrincipal, error) {
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
	if err := claims.Validate(expected); err != nil {
		return AuthPrincipal{}, err
	}

	scopeSet := parseScopeSet(claims.Scope, claims.SCP)
	if enforceRequiredScopes {
		for scope := range p.requiredScopes {
			if _, ok := scopeSet[scope]; !ok {
				return AuthPrincipal{}, fmt.Errorf("%w: missing required scope %q", errUnauthorizedToken, scope)
			}
		}
	}

	clientID := claims.AuthorizedParty
	if clientID == "" {
		clientID = claims.ClientID
	}
	if clientID == "" && len(claims.Audience) > 0 {
		clientID = claims.Audience[0]
	}

	typedScopes, err := ParseScopeSet(sortedScopeKeys(scopeSet))
	if err != nil {
		return AuthPrincipal{}, fmt.Errorf("parse verified scopes: %w", err)
	}
	return AuthPrincipal{
		Subject:           claims.Subject,
		ClientID:          clientID,
		Issuer:            claims.Issuer,
		Scopes:            typedScopes,
		Email:             claims.Email,
		EmailVerified:     claims.EmailVerified,
		PreferredUsername: claims.PreferredUsername,
		DisplayName:       claims.Name,
		AvatarURL:         claims.Picture,
		Expiration:        claims.Expiry.Time(),
		Claims: map[string]any{
			"email":              claims.Email,
			"email_verified":     claims.EmailVerified,
			"preferred_username": claims.PreferredUsername,
			"name":               claims.Name,
			"picture":            claims.Picture,
		},
	}, nil
}

// introspect asks the identity provider about an access token (RFC 7662) and
// applies the same issuer, audience, lifetime and scope rules as for JWTs.
func (p *externalOIDCAuthProvider) introspect(ctx context.Context, token string, enforceRequiredScopes bool) (AuthPrincipal, error) {
	form := url.Values{"token": {token}, "token_type_hint": {"access_token"}}
	// #nosec G704 -- the introspection URL is validated at startup.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.introspectionURL, strings.NewReader(form.Encode()))
	if err != nil {
		return AuthPrincipal{}, err
	}
	// RFC 6749 §2.3.1: form-encode the credentials before Basic encoding.
	req.SetBasicAuth(url.QueryEscape(p.opts.IntrospectionClientID), url.QueryEscape(p.opts.IntrospectionClientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// #nosec G704 -- outbound call limited to the validated introspection URL.
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return AuthPrincipal{}, fmt.Errorf("%w: introspection request: %v", errAuthorizationServerUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return AuthPrincipal{}, fmt.Errorf("%w: introspection returned status %d", errAuthorizationServerUnavailable, resp.StatusCode)
	}
	var r introspectionResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxIntrospectionResponse)).Decode(&r); err != nil {
		return AuthPrincipal{}, fmt.Errorf("%w: introspection response: %v", errAuthorizationServerUnavailable, err)
	}

	if !r.Active {
		return AuthPrincipal{}, fmt.Errorf("%w: token is not active", errUnauthorizedToken)
	}
	if r.Issuer != "" && normalizeURLForCompare(r.Issuer) != normalizeURLForCompare(p.discovery.Issuer) {
		return AuthPrincipal{}, fmt.Errorf("%w: token issuer %q does not match %q", errUnauthorizedToken, r.Issuer, p.discovery.Issuer)
	}
	now := time.Now()
	if r.Expiry == nil {
		return AuthPrincipal{}, fmt.Errorf("%w: introspection response has no exp", errUnauthorizedToken)
	}
	if !r.Expiry.Time().After(now) {
		return AuthPrincipal{}, fmt.Errorf("%w: token expired", errUnauthorizedToken)
	}
	if r.NotBefore != nil && r.NotBefore.Time().After(now) {
		return AuthPrincipal{}, fmt.Errorf("%w: token not valid yet", errUnauthorizedToken)
	}
	if audience := strings.TrimSpace(p.opts.Audience); audience != "" && !r.Audience.Contains(audience) {
		return AuthPrincipal{}, fmt.Errorf("%w: token audience does not include %q", errUnauthorizedToken, audience)
	}
	if err := p.checkIntrospectedTokenType(r); err != nil {
		return AuthPrincipal{}, err
	}

	scopeSet := parseScopeSet(r.Scope, nil)
	if enforceRequiredScopes {
		for scope := range p.requiredScopes {
			if _, ok := scopeSet[scope]; !ok {
				return AuthPrincipal{}, fmt.Errorf("%w: missing required scope %q", errUnauthorizedToken, scope)
			}
		}
	}
	typedScopes, err := ParseScopeSet(sortedScopeKeys(scopeSet))
	if err != nil {
		return AuthPrincipal{}, fmt.Errorf("parse introspected scopes: %w", err)
	}

	issuer := r.Issuer
	if issuer == "" {
		issuer = p.discovery.Issuer
	}
	username := r.PreferredUsername
	if username == "" {
		username = r.Username
	}
	return AuthPrincipal{
		Subject:           r.Subject,
		ClientID:          r.ClientID,
		Issuer:            issuer,
		Scopes:            typedScopes,
		Email:             r.Email,
		EmailVerified:     r.EmailVerified,
		PreferredUsername: username,
		DisplayName:       r.Name,
		AvatarURL:         r.Picture,
		Expiration:        r.Expiry.Time(),
		Claims: map[string]any{
			"email":              r.Email,
			"email_verified":     r.EmailVerified,
			"preferred_username": username,
			"name":               r.Name,
			"picture":            r.Picture,
		},
	}, nil
}

// checkIntrospectedTokenType accepts only bearer access tokens. An
// introspection endpoint may report refresh tokens as active, and the
// token_type_hint in the request is only a lookup hint, so the response must
// positively identify an access token. RFC 6749 §5.1 token types are
// case-insensitive. DPoP-bound tokens are rejected because this server does
// not verify DPoP proofs.
func (p *externalOIDCAuthProvider) checkIntrospectedTokenType(r introspectionResponse) error {
	if r.Typ != "" && !strings.EqualFold(r.Typ, "Bearer") {
		return fmt.Errorf("%w: introspected typ %q is not an access token", errUnauthorizedToken, r.Typ)
	}
	switch {
	case r.TokenType == "":
		if !p.opts.IntrospectionAllowMissingTokenType {
			return fmt.Errorf("%w: introspection response has no token_type, so an access token cannot be told from a refresh token", errUnauthorizedToken)
		}
	case !strings.EqualFold(r.TokenType, "Bearer"):
		return fmt.Errorf("%w: token_type %q is not a bearer access token", errUnauthorizedToken, r.TokenType)
	}
	return nil
}

// looksLikeJWS reports whether token has the shape of a compact JWS: three
// base64url segments separated by dots.
func looksLikeJWS(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" {
		return false
	}
	for _, part := range parts {
		for _, c := range part {
			if !isBase64URLRune(c) {
				return false
			}
		}
	}
	return true
}

func isBase64URLRune(c rune) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	default:
		return c == '-' || c == '_'
	}
}

func (p *externalOIDCAuthProvider) ProtectedResourceMetadata() map[string]any {
	return map[string]any{
		"authorization_servers": []string{p.discovery.Issuer},
		"resource":              p.resourceURL,
	}
}

func (p *externalOIDCAuthProvider) WWWAuthenticateHeader() string {
	return BearerChallenge(ProtectedResourceMetadataURL(p.resourceURL))
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
	if _, err := validateExternalOIDCEndpointURL(c.jwksURI); err != nil {
		return err
	}
	// #nosec G704 -- jwks_uri is validated against allowed OIDC endpoint URL rules.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.jwksURI, nil)
	if err != nil {
		return err
	}

	// #nosec G704 -- outbound fetch is limited to a validated jwks_uri.
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
	if _, err := validateExternalOIDCEndpointURL(discoveryURL); err != nil {
		return oidcDiscoveryDocument{}, err
	}
	// #nosec G704 -- discovery URL is validated before request construction.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return oidcDiscoveryDocument{}, err
	}

	// #nosec G704 -- outbound fetch is limited to a validated OIDC discovery URL.
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

func validateExternalOIDCEndpointURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, err
	}
	if !parsed.IsAbs() {
		return nil, fmt.Errorf("url must be absolute")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("url must include a host")
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("userinfo is not allowed in oidc endpoint urls")
	}
	if parsed.Fragment != "" {
		return nil, fmt.Errorf("fragments are not allowed in oidc endpoint urls")
	}

	switch parsed.Scheme {
	case "https":
		return parsed, nil
	case "http":
		host := parsed.Hostname()
		if host == "localhost" {
			return parsed, nil
		}
		ip := net.ParseIP(host)
		if ip != nil && ip.IsLoopback() {
			return parsed, nil
		}
		return nil, fmt.Errorf("http is allowed only for localhost or loopback addresses")
	default:
		return nil, fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
}

func normalizeURLForCompare(rawURL string) string {
	return strings.TrimRight(strings.TrimSpace(rawURL), "/")
}

// ProtectedResourceMetadataURL returns the RFC 9728 metadata URL for a resource URL.
func ProtectedResourceMetadataURL(resourceURL string) string {
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
