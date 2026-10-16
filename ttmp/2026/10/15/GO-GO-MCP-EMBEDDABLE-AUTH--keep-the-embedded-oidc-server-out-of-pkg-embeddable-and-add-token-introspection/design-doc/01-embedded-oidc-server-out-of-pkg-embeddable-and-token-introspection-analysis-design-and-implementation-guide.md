---
Title: 'Embedded OIDC server out of pkg/embeddable and token introspection: analysis, design and implementation guide'
Ticket: GO-GO-MCP-EMBEDDABLE-AUTH
Status: active
Topics:
    - mcp
    - oidc
    - authentication
    - introspection
    - dependencies
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Make pkg/embeddable a light dependency by linking the embedded development OIDC server only into programs that ask for it (new package pkg/embeddable/embeddeddev), fix embedded_dev tokens being rejected for missing expiration, and add RFC 7662 token introspection to external_oidc so opaque access tokens (ZITADEL's default for dynamically registered clients) can be verified."
LastUpdated: 2026-10-15T19:30:00-04:00
WhatFor: "Implementing and reviewing the GO-GO-MCP-EMBEDDABLE-AUTH changes; understanding go-go-mcp's HTTP auth modes."
WhenToUse: "Before changing pkg/embeddable auth, pkg/auth/oidc, or adding an auth mode."
---

# Embedded OIDC server out of pkg/embeddable and token introspection: analysis, design and implementation guide

## 1. Executive summary

`pkg/embeddable` is the part of go-go-mcp that applications import to become MCP servers. Measured with `go list -deps`, importing it compiles 318 non-standard-library packages, including gRPC, protobuf, OpenTelemetry exporters for Jaeger, Zipkin and OTLP, `ory/fosite`, `ory/x` and `mattn/go-sqlite3`. Without one import, the same package compiles 61. That import is `pkg/auth/oidc`, the embedded development OIDC server, which `pkg/embeddable/auth_provider.go` links into every program whether or not it ever selects `--auth-mode embedded_dev`.

At the same time, the production auth mode, `external_oidc`, can verify only JSON Web Token (JWT) access tokens. Identity providers often issue opaque access tokens instead; ZITADEL does so by default for clients that register themselves at runtime, which is how hosted MCP clients such as claude.ai connectors sign in. Opaque tokens can be verified only by asking the identity provider, through OAuth 2.0 token introspection (RFC 7662).

This ticket makes three changes:

1. **Link the embedded server on request.** The `embedded_dev` implementation moves to a new package, `pkg/embeddable/embeddeddev`. Programs that want it add `embeddeddev.Provide()` (or `embeddeddev.WithOIDC(...)`) to their server options. `pkg/embeddable` no longer imports `pkg/auth/oidc`.
2. **Fix `embedded_dev` over HTTP.** Its tokens are rejected today with `401 token missing expiration`, because the provider returns no expiration and the go-sdk bearer middleware requires one. The embedded server now reports each token's expiry.
3. **Add introspection to `external_oidc`.** A new option selects how access tokens are verified: `jwt` (today's behaviour, the default), `introspection`, or `auto` (JWT when the token is a JWS, introspection otherwise).

## 2. Problem statement and scope

### 2.1 Why now

The trigger is francois (FRANCOIS-007), a Go web application that plans an MCP server. Its analysis measured go-go-mcp's footprint and found go-sdk directly to be the lighter choice (29 packages) unless go-go-mcp stopped linking its development OIDC server. The same analysis found that francois's identity provider, ZITADEL, issues opaque tokens to dynamically registered clients ([zitadel#12658](https://github.com/zitadel/zitadel/issues/12658)), which go-go-mcp's `external_oidc` mode cannot verify. Both changes make go-go-mcp usable for any application in that position, not only francois.

### 2.2 In scope

- The new package and the registration option; moving `WithOIDC` there; updating the example and the documentation.
- Reporting token expiry from `pkg/auth/oidc` and using it in the embedded provider.
- Introspection in `external_oidc`: options, `mcp start` flags, discovery of the endpoint, client authentication with `client_secret_basic`, response validation, tests.
- A test that keeps `pkg/auth/oidc` out of `pkg/embeddable`'s import graph.

### 2.3 Out of scope

- Removing `geppetto/pkg/helpers` from `pkg/tools` (22 of the remaining 61 packages, including watermill). It is used for two reflection helpers in `pkg/tools/reflect.go:27,66` and is recorded as a follow-up.
- Caching introspection results (decision in §5.4).
- `private_key_jwt` client authentication for introspection (ZITADEL's "JWT profile"); `client_secret_basic` is supported first.
- Client ID Metadata Documents and dynamic client registration; those are between the MCP client and the identity provider.

## 3. Orientation

### 3.1 Terms

- **MCP server / resource server.** The program that serves `/mcp`. In OAuth terms it is a *resource server*: it receives a bearer access token with every request and must decide whether the token is valid.
- **Authorization server / identity provider.** The service that signs users in and issues access tokens (Keycloak, ZITADEL, or go-go-mcp's embedded development server).
- **Bearer token.** `Authorization: Bearer <token>`. Whoever holds the token may use it.
- **JWT access token.** A token that is itself a signed JSON document (a JWS). The resource server verifies the signature with the authorization server's public keys, published as a **JWKS** (JSON Web Key Set), and reads the claims (`iss`, `sub`, `exp`, `aud`, `scope`) without contacting anyone.
- **Opaque access token.** A random string with no readable content. Only the authorization server knows what it means.
- **Token introspection (RFC 7662).** A POST from the resource server to the authorization server's `introspection_endpoint` with `token=<token>`, authenticated as the resource server. The answer is JSON with `active: true|false` and, when active, claims such as `sub`, `client_id`, `scope`, `exp`, `aud`, `iss`.
- **Discovery document.** `/.well-known/openid-configuration` on the issuer: JSON naming the endpoints (`jwks_uri`, `token_endpoint`, `introspection_endpoint`, …).
- **Protected resource metadata (RFC 9728).** `/.well-known/oauth-protected-resource` on the MCP server: which authorization server to use for this resource. MCP clients read it after a `401` whose `WWW-Authenticate` header points to it.
- **Auth mode.** go-go-mcp's choice of verifier: `none`, `embedded_dev` (go-go-mcp runs its own development authorization server) or `external_oidc` (an external identity provider).

### 3.2 How a request is authenticated today

```text
 MCP client ──POST /mcp, Authorization: Bearer T──▶ http.ServeMux (official_backend.go)
                                                     │
                                  wrapHTTPAuthentication (official_backend.go:304)
                                    newHTTPAuthRuntime (auth_provider.go:42)
                                      mode none          → no wrapping
                                      mode embedded_dev  → embeddedDevAuthProvider ── imports pkg/auth/oidc
                                      mode external_oidc → externalOIDCAuthProvider (JWT + JWKS)
                                      custom verifier    → application's HTTPAuthVerifier
                                                     │
                                   go-sdk auth.RequireBearerToken(verifier, opts)
                                     verifier(ctx, T) → TokenInfo{UserID, Scopes, Expiration}
                                     reject if Expiration is zero (unless AllowMissingExpiration)
                                                     │
                                   injectPrincipal: AuthPrincipal into ctx, X-MCP-Subject header
                                                     ▼
                                              MCP handler
```

The verifier interface every mode implements is `HTTPAuthVerifier` (`pkg/embeddable/auth_provider.go:31-35`):

```go
type HTTPAuthVerifier interface {
	ValidateBearerToken(ctx context.Context, token string) (AuthPrincipal, error)
	ProtectedResourceMetadata() map[string]any
	WWWAuthenticateHeader() string
}
```

## 4. Current state, with evidence

### 4.1 Where the dependency weight comes from

`pkg/embeddable/auth_provider.go:10` imports `embeddedoidc "github.com/go-go-golems/go-go-mcp/pkg/auth/oidc"`. `pkg/auth/oidc` imports `ory/fosite` (five packages) and `mattn/go-sqlite3`. `fosite` imports `ory/x`, which imports OpenTelemetry, its exporters and gRPC.

| Measurement (`GOWORK=off go list -deps`, non-standard packages) | Packages |
| --- | --- |
| `./pkg/embeddable` | 318 |
| the packages `embeddable` imports other than `pkg/auth/oidc` (`./pkg ./pkg/protocol ./pkg/session ./pkg/tools ./pkg/tools/providers/tool-registry`) | 61 |
| go-sdk `mcp` + `auth` alone | 29 |

`mattn/go-sqlite3` needs cgo. In a `CGO_ENABLED=0` build it compiles to a stub that fails at run time, so applications that build static binaries carry it without being able to use it.

The only code in `pkg/embeddable` that uses `pkg/auth/oidc` is the embedded provider (`auth_provider.go:73-135`). The rest of the embedded mode is plain data: `AuthModeEmbeddedDev` and `EmbeddedOIDCOptions` (`server.go:250-284`) and the `mcp start` flags that fill them (`command.go:57-78`, `181-240`). Other users of `pkg/auth/oidc` are the `go-go-mcp oidc` admin commands (`cmd/go-go-mcp/cmds/oidc.go:6`), which belong to the go-go-mcp binary, and the example `pkg/embeddable/examples/oidc/main.go:38`.

### 4.2 `embedded_dev` rejects every token over HTTP

`embeddedDevAuthProvider.ValidateBearerToken` (`auth_provider.go:102-124`) returns a principal without `Expiration`. `authMiddleware` (`official_backend.go:368-406`) passes it to go-sdk's `RequireBearerToken` without `AllowMissingExpiration`, and go-sdk rejects a zero expiration (`go-sdk@v1.7.0/auth/auth.go:162-166`). A probe test on `main` (86db8ed) sending the configured static key through `authMiddleware` returned:

```text
status=401 body="token missing expiration\n"
```

The existing tests did not catch it because `TestEmbeddedDevProviderStaticAuthKey` calls the provider directly, and the middleware tests use a stub provider that sets `Expiration`. The same path applies to tokens issued by the embedded server: `IntrospectAccessToken` (`pkg/auth/oidc/server.go:206-229`) returns only subject and client ID.

### 4.3 `external_oidc` verifies JWTs only

`externalOIDCAuthProvider.validateBearerToken` (`auth_provider_external.go:132-192`) calls `jwt.ParseSigned(token)`. An opaque token fails parsing, is retried once with a forced JWKS refresh, and is rejected. The discovery struct (`:23-28`) does not read `introspection_endpoint`. There is no option for client credentials, which introspection requires.

## 5. Design

### 5.1 Auth-mode providers

`pkg/embeddable` gains a way to receive an auth mode's implementation from the application:

```go
// AuthModeProvider builds the resource-server verifier for an auth mode and, when the
// mode runs its own authorization server, a function that mounts that server's routes.
type AuthModeProvider func(opts AuthOptions) (HTTPAuthVerifier, func(*http.ServeMux), error)

// WithAuthModeProvider links an auth mode's implementation into this server.
func WithAuthModeProvider(mode AuthMode, provider AuthModeProvider) ServerOption
```

`ServerConfig` holds a map from mode to provider. `newHTTPAuthRuntime` looks the mode up there before its built-in cases. `embedded_dev` has no built-in case any more: if it is selected and nothing provides it, startup fails with

```text
auth mode embedded_dev is not linked into this program; add embeddeddev.Provide() from github.com/go-go-golems/go-go-mcp/pkg/embeddable/embeddeddev to the server options
```

The new package holds the implementation:

```go
package embeddeddev // github.com/go-go-golems/go-go-mcp/pkg/embeddable/embeddeddev

// Provide links the embedded development OIDC server, so --auth-mode embedded_dev works.
func Provide() embeddable.ServerOption

// WithOIDC links the embedded server and selects embedded_dev with these options.
func WithOIDC(opts embeddable.EmbeddedOIDCOptions) embeddable.ServerOption
```

`embeddable.WithOIDC` is removed; its one in-repository caller (the example) switches to `embeddeddev.WithOIDC`. `EmbeddedOIDCOptions`, the mode constant and the `mcp start` flags stay in `pkg/embeddable`, because they are data and cost nothing to link.

#### Decision: explicit option rather than registration in `init`

- **Context:** The embedded server must stay out of the import graph unless the program asks for it, while `--auth-mode embedded_dev` must keep working in programs that do.
- **Options considered:** (a) a build tag; (b) a global registry filled by a blank import's `init` (the `database/sql` driver pattern); (c) an explicit `ServerOption`.
- **Decision:** (c).
- **Rationale:** A build tag changes behaviour invisibly and is easy to forget in CI. A blank import works but hides the dependency in an `init` side effect and makes tests order-dependent through global state. An option is visible at the call site, scoped to one server configuration, and follows the package's existing configuration style (`WithAuth`, `WithHTTPAuthVerifier`).
- **Consequences:** Programs that used `embeddable.WithOIDC` or `--auth-mode embedded_dev` must add one option. No compatibility shim is kept (AGENT.md: no backwards-compatibility layers unless asked).
- **Status:** accepted

#### Decision: `external_oidc` stays built in

- **Context:** The same mechanism could move `external_oidc` out too.
- **Decision:** keep it in `pkg/embeddable`.
- **Rationale:** It needs only `go-jose` (and, after this ticket, the standard library for introspection). It is the mode production deployments use, and moving it would make every hosted user add an option for no dependency gain.
- **Status:** accepted

### 5.2 Exported challenge helpers

The moved provider needs two functions that are unexported today. They are useful to any application writing its own `HTTPAuthVerifier`, so they become exported:

```go
// ProtectedResourceMetadataURL returns the RFC 9728 metadata URL for a resource URL.
func ProtectedResourceMetadataURL(resourceURL string) string
// BearerChallenge returns the WWW-Authenticate value pointing clients to that metadata.
func BearerChallenge(resourceMetadataURL string) string
```

### 5.3 Token expiry from the embedded server

`(*oidc.Server).IntrospectAccessToken` changes from `(subject, clientID string, ok bool, err error)` to:

```go
type AccessTokenInfo struct {
	Subject   string
	ClientID  string
	ExpiresAt time.Time
}

func (s *Server) IntrospectAccessToken(ctx context.Context, token string) (AccessTokenInfo, bool, error)
```

For tokens issued by fosite, `ExpiresAt` is `requester.GetSession().GetExpiresAt(fosite.AccessToken)`. For development tokens stored in SQLite it is the stored `ExpiresAt`. The embedded provider copies it into `AuthPrincipal.Expiration`.

The static development key (`EmbeddedOIDCOptions.AuthKey`) has no lifetime. Because the middleware re-validates the key on every request, the provider reports an expiration one minute after validation; the value only states that the key was valid at the time of the request.

#### Decision: report expiry instead of allowing missing expiry

- **Context:** The middleware could instead be configured with `AllowMissingExpiration` for embedded mode.
- **Decision:** report real expiry; synthesize only for the static key.
- **Rationale:** Allowing a missing expiration would weaken the check for every token the embedded server issues, and those tokens do have lifetimes.
- **Status:** accepted

### 5.4 Token introspection in `external_oidc`

**Options.**

```go
type TokenValidation string

const (
	TokenValidationJWT           TokenValidation = "jwt"           // default: verify locally with the JWKS
	TokenValidationIntrospection TokenValidation = "introspection" // ask the identity provider (RFC 7662)
	TokenValidationAuto          TokenValidation = "auto"          // JWT when the token is a JWS, otherwise introspection
)

type ExternalOIDCOptions struct {
	IssuerURL      string
	DiscoveryURL   string
	Audience       string
	RequiredScopes []string

	TokenValidation           TokenValidation // "" means jwt
	IntrospectionURL          string          // default: introspection_endpoint from discovery
	IntrospectionClientID     string
	IntrospectionClientSecret string
}
```

**`mcp start` flags.**

| Flag | Field |
| --- | --- |
| `--oidc-token-validation jwt\|introspection\|auto` | `TokenValidation` |
| `--oidc-introspection-url` | `IntrospectionURL` |
| `--oidc-introspection-client-id` | `IntrospectionClientID` |
| `--oidc-introspection-client-secret-file` | `IntrospectionClientSecret`, read from the file and trimmed |

The secret is read from a file so it does not appear in process listings or shell history. Programmatic users set the field directly.

**Startup checks.** When validation is `introspection` or `auto`: the endpoint (explicit or discovered) must exist and pass the same URL rules as the other OIDC endpoints (`validateExternalOIDCEndpointURL`: absolute, HTTPS or loopback HTTP, no user info, no fragment); the client ID and secret must be set. Failing any check fails startup, not the first request.

**The request.** RFC 7662 §2.1 with `client_secret_basic` (RFC 6749 §2.3.1: client ID and secret are form-URL-encoded, then sent as HTTP Basic credentials):

```text
POST <introspection_endpoint>
Authorization: Basic base64(urlencode(client_id) ":" urlencode(client_secret))
Content-Type: application/x-www-form-urlencoded
Accept: application/json

token=<token>&token_type_hint=access_token
```

**Validating the response.** The response body is limited to 1 MiB. A non-200 status is an error, as is a body that is not JSON. Then, in order:

1. `active` must be `true`; otherwise the token is rejected as unauthorized.
2. If `iss` is present it must equal the discovered issuer.
3. `exp` must be present and in the future. (The go-sdk middleware would reject a missing expiration anyway; checking here gives a precise error.)
4. If `nbf` is present it must not be in the future.
5. If an audience is configured, `aud` (a string or an array) must contain it.
6. The response must identify a bearer access token: `token_type` must equal `Bearer`, compared case-insensitively (RFC 6749 §5.1). A missing `token_type` is rejected unless `IntrospectionAllowMissingTokenType` is set, because the request's `token_type_hint=access_token` is only a lookup hint and an issuer may report an active refresh token without a type. A Keycloak-style `typ` other than `Bearer` (`Refresh`, `Offline`) is rejected even with the opt-out. DPoP-bound tokens are rejected because DPoP proofs are not verified. (Revised after review on PR #89; the first version only rejected the literal `refresh_token`.)
7. Required scopes are checked against `scope` (space-separated), exactly as for JWTs.

The principal is built from `sub`, `client_id`, `iss`, `scope` and `exp`, plus `email`, `email_verified`, `preferred_username`/`username`, `name` and `picture` when the identity provider includes them (ZITADEL can include profile claims in introspection responses).

**Identity provider failures.** A network error, a non-200 status or a malformed body is not an answer about the token. Such errors wrap `errAuthorizationServerUnavailable`, and the HTTP middleware returns them unwrapped, so go-sdk answers `500` instead of `401`. A `401` would tell the MCP client to sign in again, which cannot help while the identity provider is unreachable. (Added during implementation.)

**`auto`.** A JWS in compact form has three base64url segments separated by dots. If the token has that shape it is verified as a JWT and is *not* sent to introspection when verification fails, so a forged JWT never reaches the identity provider. Anything else is introspected.

#### Decision: no introspection cache in this change

- **Context:** Every MCP request is authenticated, and with stateless Streamable HTTP every JSON-RPC message is a request, so introspection adds one round trip to the identity provider per message.
- **Options considered:** a TTL cache keyed by the token's SHA-256; no cache.
- **Decision:** no cache for now.
- **Rationale:** Without a cache, revocation at the identity provider takes effect on the next request. A cache trades that for latency, and the right TTL depends on the deployment. It can be added later as an option without changing the interface.
- **Status:** accepted

## 6. Pseudocode

### 6.1 Choosing the runtime

```go
func newHTTPAuthRuntime(cfg *ServerConfig) (httpAuthRuntime, error) {
	if cfg == nil || !cfg.authEnabled { return httpAuthRuntime{}, nil }
	if cfg.customAuthVerifier != nil { return httpAuthRuntime{provider: cfg.customAuthVerifier}, nil }
	if !cfg.authOptions.Enabled() { return httpAuthRuntime{}, nil }
	if p, ok := cfg.authModeProviders[cfg.authOptions.Mode]; ok {
		verifier, mount, err := p(cfg.authOptions)
		if err != nil { return httpAuthRuntime{}, err }
		return httpAuthRuntime{provider: verifier, mountAuthorizationServer: mount}, nil
	}
	switch cfg.authOptions.Mode {
	case AuthModeExternalOIDC:
		provider, err := newExternalOIDCAuthProvider(cfg.authOptions)
		...
	case AuthModeEmbeddedDev:
		return httpAuthRuntime{}, errEmbeddedDevNotLinked
	default:
		return httpAuthRuntime{}, fmt.Errorf("unsupported auth mode: %s", cfg.authOptions.Mode)
	}
}
```

### 6.2 Validating an external token

```go
func (p *externalOIDCAuthProvider) validateBearerToken(ctx context.Context, token string, enforceScopes bool) (AuthPrincipal, error) {
	switch p.validation {
	case TokenValidationIntrospection:
		return p.introspect(ctx, token, enforceScopes)
	case TokenValidationAuto:
		if !looksLikeJWS(token) {
			return p.introspect(ctx, token, enforceScopes)
		}
	}
	return p.validateJWT(ctx, token, enforceScopes) // today's code, unchanged
}

func (p *externalOIDCAuthProvider) introspect(ctx context.Context, token string, enforceScopes bool) (AuthPrincipal, error) {
	resp := postForm(p.introspectionURL, basicAuth(id, secret), {"token": token, "token_type_hint": "access_token"})
	if resp.Status != 200 { return error }
	var r introspectionResponse // json, ≤1 MiB
	if !r.Active { return errUnauthorizedToken }
	check iss, exp, nbf, aud, token_type
	scopes := parseScopeSet(r.Scope, nil)
	if enforceScopes { check required scopes }
	return principalFrom(r, scopes), nil
}
```

## 7. API and flag changes

| Change | Kind |
| --- | --- |
| `embeddable.AuthModeProvider`, `embeddable.WithAuthModeProvider` | added |
| `embeddable.ProtectedResourceMetadataURL`, `embeddable.BearerChallenge` | added (were unexported) |
| `embeddable.WithOIDC` | removed; use `embeddeddev.WithOIDC` |
| package `pkg/embeddable/embeddeddev` with `Provide`, `WithOIDC` | added |
| `oidc.(*Server).IntrospectAccessToken` returns `AccessTokenInfo` | changed |
| `embeddable.TokenValidation` and four `ExternalOIDCOptions` fields | added |
| `mcp start --oidc-token-validation`, `--oidc-introspection-url`, `--oidc-introspection-client-id`, `--oidc-introspection-client-secret-file` | added |

## 8. Implementation plan

1. Ticket and this guide.
2. Move `embedded_dev`: the provider option, the new package, `AccessTokenInfo`, exported helpers, the example, the import-graph test, and the expiry fix with an HTTP-level test that reproduces §4.2.
3. Introspection: options, flags, discovery, request, validation, tests against an `httptest` identity provider.
4. Documentation: `pkg/doc/topics/07-embedded-oidc.md`, `08-oidc-keycloak-coolify-hosted-mcp.md`, `pkg/embeddable/README.md`.

Each step ends with `GOWORK=off go test ./...` and `golangci-lint run` passing, and is one commit.

## 9. Tests

- **Import graph.** `go list -deps ./pkg/embeddable` must not contain `pkg/auth/oidc`, `ory/fosite` or `mattn/go-sqlite3`. The test runs the `go` tool and is skipped when it is unavailable.
- **Not linked.** Selecting `embedded_dev` without `Provide()` fails with the message in §5.1.
- **Embedded over HTTP.** The static key and a development token pass through `authMiddleware` with status 204 and a non-zero expiration (this fails on `main`).
- **Introspection.** Against an `httptest` server that serves discovery, a JWKS and an introspection endpoint:
  - active token accepted; principal fields mapped; Basic credentials and form body as specified;
  - inactive token, wrong issuer, expired, missing `exp`, future `nbf`, wrong audience, `refresh_token` type, missing required scope: each rejected;
  - endpoint returns 500 or invalid JSON: rejected;
  - `auto`: a JWT is verified locally without an introspection call; an opaque token is introspected; a forged JWT is rejected without an introspection call;
  - startup fails when introspection is selected without an endpoint or credentials;
  - through `officialExternalOIDCMiddleware`: an introspected token reaches the handler with the principal in context.

## 10. Risks and open questions

- **Downstream breakage.** Programs calling `embeddable.WithOIDC` or using `embedded_dev` need one more option when they upgrade. Known users in the workspace (smailnail, jesus) pin v0.0.19 and are unaffected until they upgrade.
- **Introspection latency.** One round trip per MCP request (§5.4).
- **Client authentication.** ZITADEL API applications can use Basic or private-key JWT. Only Basic is implemented; private-key JWT is a follow-up if needed.
- **Follow-up:** remove `geppetto/pkg/helpers` from `pkg/tools` to bring `pkg/embeddable` from 61 packages towards go-sdk's 29.

## 11. References

- `pkg/embeddable/auth_provider.go:10-139` — mode selection and the embedded provider.
- `pkg/embeddable/auth_provider_external.go:23-410` — external OIDC provider.
- `pkg/embeddable/official_backend.go:304-429` — HTTP auth wrapping and middleware.
- `pkg/embeddable/server.go:250-336` — auth options, `WithAuth`, `WithHTTPAuthVerifier`, `WithOIDC`.
- `pkg/embeddable/command.go:57-240` — `mcp start` auth flags.
- `pkg/auth/oidc/server.go:32-41,116,206-229` — embedded server configuration and token introspection.
- `~/go/pkg/mod/github.com/modelcontextprotocol/go-sdk@v1.7.0/auth/auth.go:97-174` — `RequireBearerToken` and the expiration rule.
- RFC 7662 (OAuth 2.0 Token Introspection), RFC 9728 (Protected Resource Metadata), RFC 6749 §2.3.1 (client password authentication).
- [zitadel#12658](https://github.com/zitadel/zitadel/issues/12658) — opaque tokens for dynamically registered clients.
- francois FRANCOIS-007 design doc §6.7–6.8 — the measurement and the use case.
