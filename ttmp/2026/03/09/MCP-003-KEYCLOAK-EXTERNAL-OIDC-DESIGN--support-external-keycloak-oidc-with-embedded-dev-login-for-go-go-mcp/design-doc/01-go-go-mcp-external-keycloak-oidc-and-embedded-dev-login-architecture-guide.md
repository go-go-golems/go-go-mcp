---
Title: go-go-mcp external Keycloak OIDC and embedded dev login architecture guide
Ticket: MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN
Status: active
Topics:
    - mcp
    - oidc
    - keycloak
    - authentication
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-mcp/cmd/go-go-mcp/cmds/oidc.go
      Note: |-
        Existing embedded OIDC admin CLI for users, clients, and dev tokens
        Existing embedded OIDC admin CLI proving dev auth is a real supported path
    - Path: go-go-mcp/pkg/auth/oidc/server.go
      Note: |-
        Current embedded issuer, login form, token issuance, and SQLite-backed user/client/token storage
        Current embedded issuer
    - Path: go-go-mcp/pkg/doc/topics/07-embedded-oidc.md
      Note: |-
        Current documented contract and user-facing claims about embedded OIDC
        Current user-facing contract that should be split into production and dev guidance
    - Path: go-go-mcp/pkg/embeddable/command.go
      Note: |-
        Current CLI flag surface for OIDC and dev auth helpers
        Current CLI flags that mix production and dev auth concerns
    - Path: go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: |-
        Current SSE/streamable HTTP auth wrapping and MCP protected-resource metadata
        Current MCP protected-resource middleware and issuer coupling point
    - Path: go-go-mcp/pkg/embeddable/server.go
      Note: |-
        Current OIDC options surface exposed to embeddable callers
        Current auth option surface and missing auth-mode abstraction
    - Path: go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md
      Note: |-
        Prior product-level design that recommended Keycloak as the external issuer
        Prior product-level recommendation to use Keycloak as the issuer
    - Path: go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/current-capability-scan.sh
      Note: |-
        Reproducible scan script for the current auth surface
        Ticket-local reproducible auth capability scan
ExternalSources:
    - https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization
    - https://modelcontextprotocol.io/docs/tutorials/security/authorization
    - https://www.keycloak.org/guides#server
    - https://www.keycloak.org/securing-apps/oidc-layers
    - https://coolify.io/docs/services/keycloak
Summary: Detailed intern-facing guide for evolving go-go-mcp from an embedded-issuer-only model to a dual-mode auth architecture that supports Keycloak as an external issuer while retaining an embedded local password login path for development.
LastUpdated: 2026-03-09T18:12:00-04:00
WhatFor: Explain the current auth architecture, identify the coupling that prevents external issuer support today, and provide a concrete implementation design for Keycloak-compatible production auth plus embedded dev auth.
WhenToUse: Use when designing or implementing external issuer support, refactoring go-go-mcp auth middleware, or onboarding an engineer to the current and future OIDC/MCP auth architecture.
---


# go-go-mcp external Keycloak OIDC and embedded dev login architecture guide

## Executive Summary

`go-go-mcp` already contains a real embedded OIDC authorization server and uses it to protect HTTP MCP endpoints. That is not hypothetical. The current code can serve discovery metadata, JWKS, login and token endpoints, dynamic client registration, and protected-resource metadata, all on the same port as the MCP server. The main technical limitation is architectural coupling: the HTTP backends currently instantiate an in-process issuer and the auth middleware only knows how to validate bearer tokens by asking that in-process issuer to introspect them.

That design is useful for demos and local development, but it is the wrong production shape for hosted deployments that already have a real identity provider such as Keycloak. In production, `go-go-mcp` should behave as an MCP protected resource server that trusts an external issuer, validates JWT access tokens locally, enforces issuer and audience constraints, and advertises the correct authorization metadata to clients. At the same time, the repository should keep an embedded developer path because a simple local username/password login flow is still valuable for demos, tests, and offline development.

The recommended design is a dual-mode auth subsystem behind a shared validation interface:

- `external_oidc` mode: trust Keycloak as the issuer, fetch discovery metadata and JWKS, verify JWTs, and advertise Keycloak in the MCP metadata surface.
- `embedded_dev` mode: keep the current in-process issuer and the lightweight password login flow, but explicitly classify it as development-only.

The crucial refactor is not “rewrite everything around Keycloak.” It is “split issuer responsibilities from resource-server responsibilities.” Today those responsibilities are fused together. After the refactor, the MCP HTTP layer should depend on an abstract auth provider contract instead of directly creating an embedded issuer.

## Problem Statement

The user wants to support two operational modes at the same time:

1. Hosted / production auth where Keycloak is the real OIDC issuer for MCP bearer tokens.
2. Local / development auth where a developer can still use a simple local password login without standing up Keycloak.

The current `go-go-mcp` code only solves the second problem directly. It already has:

- an embedded issuer in [pkg/auth/oidc/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/auth/oidc/server.go)
- embedded auth CLI flags in [pkg/embeddable/command.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/command.go)
- HTTP auth middleware in [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go)
- admin helpers for embedded users/clients/tokens in [cmd/go-go-mcp/cmds/oidc.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/cmd/go-go-mcp/cmds/oidc.go)

But it does not yet have:

- an external issuer mode
- remote discovery / JWKS fetching
- JWT verification for externally issued access tokens
- audience enforcement
- auth provider abstraction between “embedded issuer” and “external issuer”
- a CLI/config model that makes dev-only auth clearly separate from production auth

The design therefore needs to answer four questions:

1. What exact parts of the current system should remain?
2. What exact parts must be decoupled?
3. How should Keycloak-mode and local-dev-mode coexist without becoming a confusing flag soup?
4. How should the MCP protected-resource metadata and `WWW-Authenticate` headers behave in each mode?

## Scope

### In scope

- analysis of the current embedded OIDC implementation
- analysis of the current MCP auth wrapping
- design for external issuer support with Keycloak
- design for preserving a local password login flow
- CLI and config refactor recommendations
- testing strategy and phased implementation plan

### Out of scope

- implementing the refactor in this ticket
- replacing Keycloak itself
- designing a general-purpose identity management product
- adding browser UI beyond what is needed for local dev login
- changing stdio transport auth, which is not HTTP bearer-token based

## Current-State Architecture

### 1. The repository already has an embedded issuer

The embedded issuer lives in [pkg/auth/oidc/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/auth/oidc/server.go). The `Config` type includes:

- `Issuer`
- `DBPath`
- `EnableDevTokens`
- `User`
- `Pass`
- `Authenticator`

That is evidence that the auth layer already supports:

- discovery and AS metadata
- a login form
- static user/password auth
- SQLite-backed user/password auth
- optional pluggable authentication logic

The current `Server` object is not only a token validator. It is all of these at once:

- authorization server
- login application
- client registry
- key store
- dev token store
- user store
- token introspector

That is both the strength of the current implementation and the main reason external issuer support is awkward today.

### 2. HTTP transports instantiate the embedded issuer directly

Both the SSE backend and the streamable HTTP backend create the embedded issuer themselves in [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go#L213) and [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go#L264).

That means:

- the backend constructs `oidc.New(...)`
- the backend mounts OIDC routes on the same mux
- the backend adds `/.well-known/oauth-protected-resource`
- the backend wraps the MCP handler with `oidcAuthMiddleware(...)`

This is the key coupling point. The transport layer is currently saying:

```go
if oidcEnabled {
    issuer := oidc.New(...)
    issuer.Routes(mux)
    mux.HandleFunc("/.well-known/oauth-protected-resource", ...)
    handler = oidcAuthMiddleware(cfg, issuer, handler)
}
```

That is fine for embedded mode, but impossible to reuse for “trust Keycloak” without either:

- creating a fake local `oidc.Server` wrapper for external JWTs, or
- separating “issuer routes” from “bearer token validation” and “protected resource metadata”

The second option is cleaner.

### 3. The current middleware only understands in-process introspection

The current middleware in [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go#L301) does three things:

1. allows public auth/discovery routes through
2. accepts a static auth key for dev
3. validates all other bearer tokens via `oidcSrv.IntrospectAccessToken(...)`

That gives the following behavior:

```text
Authorization header missing
  -> 401 + WWW-Authenticate

Authorization header is static dev auth key
  -> accept immediately

Authorization header is bearer token
  -> ask embedded oidcSrv to introspect token
  -> if valid, set X-MCP-Subject and X-MCP-Client-ID
  -> otherwise 401
```

Important implication: there is no path here for:

- fetching external discovery metadata
- looking up external JWKS
- verifying JWT signatures
- checking `iss`
- checking `aud`
- checking scopes or roles

The current code is therefore not “almost external issuer support.” It is “resource protection hard-wired to the embedded issuer object.”

### 4. The current config model is issuer-centric, not provider-centric

The OIDC options in [pkg/embeddable/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/server.go#L223) are:

- `Issuer`
- `DBPath`
- `EnableDevTokens`
- `AuthKey`
- `User`
- `Pass`

And the CLI flags in [pkg/embeddable/command.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/command.go#L55) expose the same worldview:

- `--oidc`
- `--issuer`
- `--oidc-db`
- `--oidc-dev-tokens`
- `--oidc-auth-key`
- `--oidc-user`
- `--oidc-pass`

This is adequate for embedded mode, but it conflates:

- issuer identity
- token validation mode
- dev login mode
- dev shortcuts
- persistence backing store

The design gap is that there is no first-class “auth mode.”

### 5. The existing dev login path is real and worth preserving

The login form and cookie-backed dev session are implemented in [pkg/auth/oidc/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/auth/oidc/server.go#L281) through [pkg/auth/oidc/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/auth/oidc/server.go#L347).

Notable characteristics:

- HTML form at `/login`
- username/password POST
- pluggable `Authenticator`
- static fallback user/pass if no custom authenticator
- SQLite-backed bcrypt user store if `DBPath` is set
- cookie named `sid`

This means local password login is already present as a working pattern. The right question is not “can we still support local login?” The answer is yes. The right question is “how do we preserve it without making it look like a production identity strategy?”

### 6. The repo already exposes embedded OIDC administration

The CLI group in [cmd/go-go-mcp/cmds/oidc.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/cmd/go-go-mcp/cmds/oidc.go) provides:

- `oidc users`
- `oidc tokens`
- `oidc clients`

That is useful evidence because it implies:

- embedded mode is treated as a first-class feature today
- dev/local workflows already depend on SQLite-managed users and clients
- removing embedded mode entirely would be a regression for current users

### 7. Documentation currently frames embedded OIDC as the main secure path

The user-facing doc [07-embedded-oidc.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/doc/topics/07-embedded-oidc.md) describes the embedded setup as “first-class and runs in-process.” That is accurate today, but once external issuer support exists, the documentation should separate:

- production / hosted recommendation
- local / demo / development recommendation

Otherwise the docs will keep steering users toward the wrong architecture for real deployments.

## Gap Analysis

### What is already good

- The project already understands MCP protected-resource metadata and `WWW-Authenticate` advertising.
- The project already understands OIDC discovery and AS metadata.
- The project already has a pluggable username/password authenticator abstraction.
- The project already has dev-friendly SQLite and static-credential flows.
- The project already has a single-port “same mux as MCP” deployment model.

### What is missing

- explicit production-grade external issuer support
- clear auth-mode selection
- JWT verification for externally issued access tokens
- remote JWKS caching/refresh
- audience enforcement
- explicit dev-only guardrails for static auth key and local passwords
- documentation split between embedded dev and external production modes

### What is risky in the current design

- The current middleware only checks “can the embedded server introspect this?” That is too narrow.
- `AuthKey` is a useful shortcut but too dangerous if it leaks into production habits.
- `User` / `Pass` defaults (`admin` / `password123`) are acceptable for demo mode only; they should not be silently reachable in deployed systems.
- The current config surface does not force users to think in terms of auth mode, so misuse is easy.

## Recommended Solution

## Design Decision 1: Split auth into provider modes

Introduce an auth provider abstraction used by the HTTP backends. The backend should no longer create `oidc.New(...)` unconditionally when auth is enabled.

Instead, it should build one provider instance from config:

- `embedded_dev`
- `external_oidc`

Recommended conceptual interface:

```go
type HTTPAuthProvider interface {
    PublicRoutes(mux *http.ServeMux)
    ValidateBearerToken(ctx context.Context, token string) (Principal, error)
    ProtectedResourceMetadata() ProtectedResourceMetadata
    WWWAuthenticateHeader() string
    Mode() string
}

type Principal struct {
    Subject  string
    ClientID string
    Issuer   string
    Scopes   []string
    Claims   map[string]any
}
```

Why this matters:

- embedded mode can keep mounting login / token / register routes
- external mode can mount only resource metadata or nothing beyond MCP resource metadata
- the middleware becomes mode-agnostic

## Design Decision 2: Keep embedded login, but explicitly classify it as dev-only

Do not remove embedded auth. Reclassify it.

Recommended positioning:

- `embedded_dev` is for:
  - local development
  - demos
  - CI smoke setups
  - offline testing
- `external_oidc` is for:
  - hosted environments
  - multi-user deployments
  - any environment exposed beyond localhost

Recommended guardrails:

- require `auth.mode=embedded_dev` explicitly
- log a warning at startup when embedded mode is used
- optionally refuse embedded mode unless:
  - issuer is localhost, or
  - an explicit `--allow-insecure-embedded-auth` flag is set

## Design Decision 3: In production, validate Keycloak tokens locally using discovery + JWKS

Do not make `go-go-mcp` ask Keycloak to introspect every request unless there is a specific reason to prefer remote introspection.

Default production behavior should be:

1. fetch OIDC discovery metadata
2. fetch JWKS URI from discovery
3. cache keys
4. verify JWT access token locally
5. enforce:
   - `iss`
   - signature
   - expiry / not-before
   - audience or resource indicator
   - optional scopes

Why local JWT validation is preferable:

- fewer network round trips
- better resilience
- standard protected-resource server behavior
- cleaner operational separation between issuer and resource server

## Design Decision 4: Keep MCP protected-resource metadata independent from issuer implementation

The MCP HTTP resource should always be able to advertise:

- authorization server metadata location
- protected-resource metadata
- canonical resource identifier

That behavior should not depend on whether the issuer is embedded or external.

The current code already does this in [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go#L343). Preserve the behavior, but move the values behind the provider abstraction.

## Proposed API and Config Shape

### New config model

Recommended replacement for today’s flat OIDC options:

```go
type AuthMode string

const (
    AuthModeNone        AuthMode = "none"
    AuthModeEmbeddedDev AuthMode = "embedded_dev"
    AuthModeExternalOIDC AuthMode = "external_oidc"
)

type AuthOptions struct {
    Mode AuthMode

    ExternalOIDC ExternalOIDCOptions
    EmbeddedDev  EmbeddedDevOptions
}

type ExternalOIDCOptions struct {
    IssuerURL       string
    DiscoveryURL    string
    Audience        []string
    RequiredScopes  []string
    AllowedAlgs     []string
    JWKSRefresh     time.Duration
    HTTPTimeout     time.Duration
}

type EmbeddedDevOptions struct {
    Issuer          string
    DBPath          string
    EnableDevTokens bool
    AuthKey         string
    User            string
    Pass            string
}
```

### Recommended CLI flags

Production-friendly flags:

- `--auth-mode none|embedded_dev|external_oidc`
- `--oidc-issuer-url https://keycloak.example.com/realms/myrealm`
- `--oidc-discovery-url ...` (optional override)
- `--oidc-audience mcp-resource`
- `--oidc-required-scope mcp:invoke`

Embedded-dev-only flags:

- `--embedded-issuer http://localhost:3001`
- `--embedded-db /tmp/mcp-oidc.db`
- `--embedded-dev-tokens`
- `--embedded-auth-key ...`
- `--embedded-user ...`
- `--embedded-pass ...`

This avoids the current ambiguity where `--issuer` and `--oidc-user` are part of one blended mode.

## Proposed Runtime Architecture

```text
                           +------------------------------+
                           |      embeddable backend      |
                           |  SSE / Streamable HTTP mux   |
                           +---------------+--------------+
                                           |
                                           v
                           +------------------------------+
                           |      HTTPAuthProvider        |
                           +---------------+--------------+
                                           |
                    +----------------------+----------------------+
                    |                                             |
                    v                                             v
      +-------------------------------+          +--------------------------------+
      | EmbeddedDevAuthProvider       |          | ExternalOIDCAuthProvider       |
      | - wraps current oidc.Server   |          | - discovery fetch              |
      | - mounts /login,/oauth2/...   |          | - jwks cache                   |
      | - local sqlite/static auth    |          | - jwt verification             |
      +-------------------------------+          +--------------------------------+
```

### Middleware pseudocode

```go
func authMiddleware(provider HTTPAuthProvider, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if provider.IsPublicRoute(r.URL.Path) {
            next.ServeHTTP(w, r)
            return
        }

        token, ok := extractBearer(r.Header.Get("Authorization"))
        if !ok {
            w.Header().Set("WWW-Authenticate", provider.WWWAuthenticateHeader())
            http.Error(w, "missing bearer", http.StatusUnauthorized)
            return
        }

        principal, err := provider.ValidateBearerToken(r.Context(), token)
        if err != nil {
            w.Header().Set("WWW-Authenticate", provider.WWWAuthenticateHeader())
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        r2 := r.Clone(withPrincipal(r.Context(), principal))
        r2.Header.Set("X-MCP-Subject", principal.Subject)
        r2.Header.Set("X-MCP-Client-ID", principal.ClientID)
        next.ServeHTTP(w, r2)
    })
}
```

### Embedded-dev provider pseudocode

```go
func NewEmbeddedDevProvider(opts EmbeddedDevOptions) (*EmbeddedDevProvider, error) {
    srv, err := oidc.New(oidc.Config{
        Issuer:          opts.Issuer,
        DBPath:          opts.DBPath,
        EnableDevTokens: opts.EnableDevTokens,
        User:            opts.User,
        Pass:            opts.Pass,
    })
    if err != nil { return nil, err }

    return &EmbeddedDevProvider{server: srv, opts: opts}, nil
}

func (p *EmbeddedDevProvider) ValidateBearerToken(ctx context.Context, token string) (Principal, error) {
    if p.opts.AuthKey != "" && token == p.opts.AuthKey {
        return Principal{Subject: "static-key-user", ClientID: "static-key-client", Issuer: p.opts.Issuer}, nil
    }
    subj, cid, ok, err := p.server.IntrospectAccessToken(ctx, token)
    if err != nil || !ok { return Principal{}, errUnauthorized }
    return Principal{Subject: subj, ClientID: cid, Issuer: p.opts.Issuer}, nil
}
```

### External provider pseudocode

```go
func NewExternalOIDCProvider(opts ExternalOIDCOptions) (*ExternalOIDCProvider, error) {
    metadata := fetchDiscovery(opts)
    jwks := newJWKSCache(metadata.JWKSURI, opts.JWKSRefresh)
    return &ExternalOIDCProvider{
        issuer:   metadata.Issuer,
        jwks:     jwks,
        audience: opts.Audience,
        scopes:   opts.RequiredScopes,
    }, nil
}

func (p *ExternalOIDCProvider) ValidateBearerToken(ctx context.Context, token string) (Principal, error) {
    claims, err := verifyJWTWithJWKS(ctx, token, p.jwks)
    if err != nil { return Principal{}, errUnauthorized }
    if claims.Issuer != p.issuer { return Principal{}, errUnauthorized }
    if !audienceAllowed(claims.Audience, p.audience) { return Principal{}, errUnauthorized }
    if !scopesAllowed(claims.Scope, p.scopes) { return Principal{}, errUnauthorized }
    return Principal{
        Subject:  claims.Subject,
        ClientID: claims.AuthorizedParty,
        Issuer:   claims.Issuer,
        Scopes:   parseScopes(claims.Scope),
        Claims:   claims.Raw,
    }, nil
}
```

## Alternatives Considered

## Alternative A: Expand the embedded issuer until it behaves like Keycloak

This would mean making `go-go-mcp` itself handle:

- external/social login brokering
- richer user lifecycle
- more client administration
- more token policy configuration
- more issuer features

Why not:

- this turns `go-go-mcp` into an identity product
- it duplicates what Keycloak already does well
- it creates ongoing security and maintenance burden

Conclusion: reject.

## Alternative B: Keep the current code and add a “Keycloak compatibility” shim inside `pkg/auth/oidc`

This would mean forcing the current `oidc.Server` type to also represent “external issuer mode.”

Why not:

- `pkg/auth/oidc/server.go` is conceptually an embedded issuer implementation
- overloading it to mean “sometimes I am an issuer, sometimes I am just a JWT validator” makes the type harder to reason about
- the HTTP backend coupling remains muddy

Conclusion: reject.

## Alternative C: Introduce an auth-provider abstraction and keep embedded issuer code mostly intact

This is the recommended approach.

Why:

- least invasive path
- preserves existing embedded functionality
- allows clean external issuer support
- lets production and dev modes be explicit

Conclusion: accept.

## Relationship to Prior SMAILNAIL-003 Design

The earlier hosted `smailnail` design in [SMAILNAIL-003](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md) recommended:

- Keycloak as the product issuer
- GitHub as upstream social login
- MCP bearer-token auth via the same issuer

This ticket refines the `go-go-mcp` side of that story. The earlier ticket was a product architecture recommendation. This ticket is the supporting library/server refactor needed to make that recommendation practical.

## Testing and Validation Strategy

### Unit tests

- provider selection based on `auth.mode`
- `WWW-Authenticate` header construction
- protected-resource metadata construction
- external JWT validation:
  - valid issuer
  - wrong issuer
  - wrong audience
  - expired token
  - missing scope
- embedded dev validation:
  - static auth key accepted
  - sqlite-backed user login flow works
  - invalid password rejected

### Integration tests

- embedded dev mode:
  - start local server
  - fetch discovery
  - login
  - auth code with PKCE
  - token exchange
  - authenticated MCP request
- external mode:
  - use a test issuer or synthetic JWKS server
  - issue JWT
  - call protected MCP endpoint
  - verify subject/client propagation

### Smoke tests

Add a maintained smoke script later that can:

- boot a test server in embedded dev mode
- verify `401` + `WWW-Authenticate`
- verify `/.well-known/oauth-protected-resource`
- verify successful authenticated tool request

For external mode, a realistic smoke should either:

- use Keycloak in Docker, or
- use a minimal local fake issuer that serves discovery + JWKS + signed test JWTs

## Implementation Plan

### Phase 1: Refactor configuration and naming

Files to change first:

- [pkg/embeddable/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/server.go)
- [pkg/embeddable/command.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/command.go)

Goals:

- introduce explicit `auth.mode`
- separate external and embedded option groups
- preserve backwards compatibility temporarily if needed, but mark old flags deprecated quickly

### Phase 2: Introduce the auth provider abstraction

Create a new package, for example:

- `pkg/auth/providers`

Suggested files:

- `provider.go`
- `embedded_dev.go`
- `external_oidc.go`
- `metadata.go`
- `middleware.go`

Goals:

- move generic HTTP auth behavior out of `mcpgo_backend.go`
- keep the current embedded issuer implementation mostly intact

### Phase 3: Wire embedded mode through the abstraction

Files:

- [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go)
- `pkg/auth/providers/embedded_dev.go`

Goals:

- prove no behavior regression in embedded mode
- preserve current login form and admin CLI behavior

### Phase 4: Add external Keycloak mode

New dependencies will likely use packages already present transitively in `go.mod`, such as JOSE/JWT tooling.

Files:

- `pkg/auth/providers/external_oidc.go`
- tests for JWT validation and JWKS refresh

Goals:

- discovery fetch
- jwks cache
- JWT validation
- issuer / audience / scope checks

### Phase 5: Documentation and examples

Update:

- [pkg/doc/topics/07-embedded-oidc.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/doc/topics/07-embedded-oidc.md)

Add:

- a new doc like `08-external-oidc.md`
- an example server configured for external Keycloak mode

## Suggested File Layout

```text
pkg/auth/
  oidc/
    server.go               # keep embedded issuer implementation here
  providers/
    provider.go             # shared interface and principal type
    embedded_dev.go         # adapter around pkg/auth/oidc/server.go
    external_oidc.go        # Keycloak-compatible discovery/JWKS/JWT validation
    middleware.go           # shared HTTP auth middleware
    metadata.go             # protected-resource metadata helpers

pkg/embeddable/
  server.go                 # new auth config model
  command.go                # new flags / deprecations
  mcpgo_backend.go          # consume provider abstraction instead of raw oidc.Server
```

## API References and External Guidance

The design in this ticket is informed by:

- MCP authorization spec:
  - https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization
- MCP auth tutorial:
  - https://modelcontextprotocol.io/docs/tutorials/security/authorization
- Keycloak server guides:
  - https://www.keycloak.org/guides#server
- Keycloak OIDC layers / endpoints:
  - https://www.keycloak.org/securing-apps/oidc-layers
- Coolify Keycloak service docs:
  - https://coolify.io/docs/services/keycloak

Inference from those sources:

- the MCP side should behave as a protected resource server, not as a bespoke auth system
- Keycloak is a better production issuer than an expanded `go-go-mcp` embedded issuer
- local password login is still reasonable if explicitly scoped to development

## Risks and Review Points

### Risk 1: Backwards-compatibility pressure creates an unclear config API

If old `--oidc-*` flags are left in place forever, users will not know whether they are in embedded or external mode.

Mitigation:

- add explicit `auth.mode`
- deprecate ambiguous flags quickly

### Risk 2: External mode ships without audience enforcement

That would make JWT verification incomplete for a real protected resource.

Mitigation:

- require at least one configured audience/resource identifier in production examples

### Risk 3: Dev shortcuts leak into production habits

`AuthKey` and static passwords are useful, but dangerous if they become “normal.”

Mitigation:

- warnings
- explicit mode naming
- optional startup refusal on non-localhost embedded mode

### Risk 4: Keycloak assumptions become too Keycloak-specific

The implementation should be “external OIDC issuer” first and “Keycloak-compatible” second.

Mitigation:

- build on discovery/JWKS/standard claims
- keep Keycloak-specific assumptions minimal

## Open Questions

1. Should external mode support token introspection as an optional alternative to local JWT verification?
2. Should scope enforcement be optional or required whenever `external_oidc` mode is enabled?
3. Should embedded dev mode be allowed on non-localhost hosts with an explicit override, or be hard-disabled?
4. Should `X-MCP-Subject` and `X-MCP-Client-ID` remain header-based propagation, or move to typed context values only?

## Recommended Immediate Next Ticket

Implement:

`MCP-004 external issuer support in go-go-mcp`

Suggested acceptance criteria:

1. `auth.mode=external_oidc` validates Keycloak-issued bearer JWTs.
2. `auth.mode=embedded_dev` preserves the current login flow and admin CLI.
3. `/.well-known/oauth-protected-resource` is served in both modes.
4. `WWW-Authenticate` is correct in both modes.
5. Embedded shortcuts are clearly documented as development-only.

## Quick Onboarding Summary For An Intern

If you are new to this system, remember the mental model:

- `pkg/auth/oidc/server.go` is today’s embedded issuer and login app.
- `pkg/embeddable/mcpgo_backend.go` is today’s MCP HTTP resource wrapper.
- the main refactor is to make the wrapper depend on an abstract auth provider, not directly on the embedded issuer.
- Keycloak mode should make `go-go-mcp` act like a protected resource server.
- embedded mode should remain the fast local path, but it must stop pretending to be the default production answer.
