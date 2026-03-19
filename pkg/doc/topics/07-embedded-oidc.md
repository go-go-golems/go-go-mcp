---
Title: Embedded OIDC in go-go-mcp
Slug: embedded-oidc
Short: Built‑in OpenID Connect for securing HTTP MCP endpoints with minimal configuration.
Topics:
- mcp
- oidc
- security
- http
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Embedded and External OIDC in go-go-mcp

## Overview

This document explains how HTTP auth works in `go-go-mcp` for SSE and streamable HTTP transports. There are now two supported auth modes:

- `embedded_dev`: the built-in OIDC server runs in-process and shares a mux with the MCP resource
- `external_oidc`: `go-go-mcp` trusts an external issuer such as Keycloak and validates bearer JWTs against discovery and JWKS metadata

The embedded integration is still first-class and runs in-process: the OIDC Authorization Server (AS) and the MCP resource live on the same port and share a single HTTP mux. The external mode keeps the same protected-resource behavior but removes the requirement that `go-go-mcp` itself be the issuer.

The goal is to make the secure path the easy path: enable standards‑compliant OAuth 2.1/OIDC flows (authorization code with PKCE), serve discovery/JWKS, protect `/mcp` with Bearer tokens, and provide pragmatic developer conveniences for local testing.

If you need the operator-facing deployment guide that starts with OAuth/OIDC fundamentals, then explains Keycloak, and then walks through a real Coolify-hosted case for OpenAI and Claude-compatible MCP clients, read [OIDC, Keycloak, and Coolify for Hosted MCP](./08-oidc-keycloak-coolify-hosted-mcp.md).

## Architecture

The embedded OIDC implementation is composed of two pieces:

- `pkg/auth/oidc`: A self‑contained OIDC/OAuth 2.1 server based on Fosite.
  - Routes: discovery, AS metadata, JWKS, login, authorize, token, dynamic client registration.
  - Optional SQLite persistence for registered clients, signing keys (JWKS), dev tokens, and tool‑call logs.
- `pkg/embeddable` HTTP backends: SSE and streamable HTTP.
  - Mount OIDC routes and the MCP endpoints on the same `http.ServeMux`.
  - Wrap `/mcp` with a Bearer middleware that validates tokens via OIDC introspection (plus optional dev helpers).

At runtime, a single `http.Server` is started for the chosen transport on the configured port. The mux contains:

- OIDC endpoints:
  - `/.well-known/openid-configuration`
  - `/.well-known/oauth-authorization-server`
  - `/jwks.json`
  - `/login`, `/oauth2/auth`, `/oauth2/token`, `/register`
  - `/.well-known/oauth-protected-resource` (RFC 9728)
- MCP endpoints (depending on transport):
  - SSE: `/mcp/sse` and `/mcp/message` (exposed by the SSE handler mounted at `/mcp/`)
  - Streamable HTTP: `/mcp`

The Bearer middleware enforces `Authorization: Bearer <access_token>` on `/mcp` (and subpaths). On `401`, the response includes a standards‑compliant `WWW-Authenticate` header pointing clients to discovery metadata and the protected resource metadata.

## Runtime Behavior

### Token Verification

- The Bearer middleware calls the embedded OIDC server to introspect opaque access tokens (Fosite `IntrospectToken`).
- On success, it injects the authenticated `subject` and `client_id` into request headers (`X-MCP-Subject`, `X-MCP-Client-ID`) for downstream processing.
- On failure, a `401 Unauthorized` is returned with a populated `WWW-Authenticate` header. The OIDC endpoints remain public.

### Developer Conveniences (Opt‑in)

- Static Auth Key: If configured, requests with `Authorization: Bearer <AuthKey>` are accepted and treated as an authenticated call (`subject=static-key-user`, `client_id=static-key-client`). Useful for quick end‑to‑end tests.
- Dev Tokens Fallback: If enabled together with SQLite, the middleware will accept tokens stored in the `oauth_tokens` table when introspection fails and the token is not expired. This is intended for local development.

Both features are disabled by default.

### Persistence

If `DBPath` is set, the OIDC server persists:

- `oauth_clients`: dynamically registered clients (ids and redirect URIs)
- `oauth_keys`: the RSA private key used for signing (ensures stable JWKS across restarts)
- `oauth_tokens`: optional dev tokens
- `mcp_tool_calls`: optional tool call logs

## Enabling Auth

The embeddable API now exposes an explicit auth model. `WithOIDC(...)` is still supported as a compatibility wrapper for the embedded path, but new code should prefer `WithAuth(...)`.

```go
type AuthMode string

const (
    AuthModeNone         AuthMode = "none"
    AuthModeEmbeddedDev  AuthMode = "embedded_dev"
    AuthModeExternalOIDC AuthMode = "external_oidc"
)

type AuthOptions struct {
    Mode        AuthMode
    ResourceURL string
    Embedded    EmbeddedOIDCOptions
    External    ExternalOIDCOptions
}

type EmbeddedOIDCOptions struct {
    Issuer          string
    DBPath          string
    EnableDevTokens bool
    AuthKey         string
    User            string
    Pass            string
}

type ExternalOIDCOptions struct {
    IssuerURL      string
    DiscoveryURL   string
    Audience       string
    RequiredScopes []string
}

func WithAuth(opts AuthOptions) ServerOption
func WithOIDC(opts OIDCOptions) ServerOption // compatibility wrapper to embedded_dev
```

## Enabling Embedded OIDC

The compatibility API is still available and keeps the old embedded behavior:

```go
// OIDC options
type OIDCOptions struct {
    Issuer          string // e.g. https://myapp.example.com
    DBPath          string // SQLite path for OIDC persistence (optional)
    EnableDevTokens bool   // Accept DB tokens for dev (optional; default: false)
    AuthKey         string // Static bearer token for dev (optional; default: empty)
    User            string // Static login user (dev only; ignored if custom authenticator)
    Pass            string // Static login password (dev only; ignored if custom authenticator)
}

func WithOIDC(opts OIDCOptions) ServerOption
```

You can enable embedded auth both programmatically and via CLI flags on the generated `mcp` command.

### Programmatic Configuration

```go
err := embeddable.AddMCPCommand(rootCmd,
    embeddable.WithDefaultTransport("sse"),
    embeddable.WithDefaultPort(3001),
    embeddable.WithAuth(embeddable.AuthOptions{
        Mode: embeddable.AuthModeEmbeddedDev,
        Embedded: embeddable.EmbeddedOIDCOptions{
            Issuer:          "https://your.domain",
            DBPath:          "/var/lib/mcp/oidc.db",
            EnableDevTokens: false,
            AuthKey:         "", // disabled by default
        },
    }),
)
```

### CLI Flags

When you add the embeddable server to your Cobra app, the generated `mcp start` command supports:

- `--auth-mode none|embedded_dev|external_oidc`
- `--auth-resource-url` (string): public MCP resource URL advertised in protected-resource metadata
- `--oidc-issuer-url` (string): external issuer URL for `external_oidc`
- `--oidc-discovery-url` (string): optional discovery override for `external_oidc`
- `--oidc-audience` (string): required audience for `external_oidc`
- `--oidc-required-scope` (repeatable/string-slice): required scopes for `external_oidc`
- `--embedded-issuer` (string): issuer base URL for `embedded_dev`
- `--embedded-db` (string): SQLite DB path for `embedded_dev`
- `--embedded-dev-tokens` (bool): accept DB tokens for embedded dev mode
- `--embedded-auth-key` (string): static bearer token for embedded dev mode
- `--embedded-user` (string): static login user for embedded dev mode
- `--embedded-pass` (string): static login password for embedded dev mode
- `--transport` (stdio | sse | streamable_http)
- `--port` (int)

Legacy embedded flags are still accepted for compatibility:

- `--oidc`
- `--issuer`
- `--oidc-db`
- `--oidc-dev-tokens`
- `--oidc-auth-key`
- `--oidc-user`
- `--oidc-pass`

Example:

```bash
go run . mcp start --transport sse --port 3001 \
  --auth-mode embedded_dev \
  --embedded-issuer https://your.ngrok.app \
  --embedded-db /tmp/mcp-oidc.db \
  --embedded-dev-tokens=false \
  --embedded-auth-key ''
```

## Example: `pkg/embeddable/examples/oidc/main.go`

This example demonstrates an OIDC‑protected MCP server exposing two tools tailored for deep research agents: `search` and `fetch`.

### Logging & Config

The example initializes logging via Glazed (`logging.InitLoggerFromViper()`) and uses Viper (`clay.InitViper`) to source configuration. This ensures structured logs and consistent CLI behavior.

### OIDC Setup

The server opts into SSE on port 3001 and enables OIDC with a local issuer and SQLite DB. For quick testing, the example also sets a static `AuthKey` and a static `User`/`Pass` for the embedded login page (you should remove these for real deployments).

### Tools

- `search(query string)`
  - Returns exactly one content item of type `text` whose `text` is a JSON‑encoded object matching:
    - `{ "results": [ { "id", "title", "url" }, ... ] }`
- `fetch(id string)`
  - Returns exactly one content item of type `text` whose `text` is a JSON‑encoded object matching:
    - `{ "id", "title", "text", "url", "metadata": { ... } }`

These shapes are designed to integrate cleanly with deep research workflows and conform to MCP’s content contract.

### Running the Example

```bash
go run ./pkg/embeddable/examples/oidc \
  mcp start --transport sse --port 3001 \
  --auth-mode embedded_dev \
  --embedded-issuer http://localhost:3001 \
  --embedded-db /tmp/mcp-oidc.db \
  --embedded-auth-key TEST_AUTH_KEY_123
```

Test `search` with a static key:

```bash
curl -s -H "Authorization: Bearer TEST_AUTH_KEY_123" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/call","params":{"name":"search","arguments":{"query":"oidc"}}}' \
  http://localhost:3001/mcp
```

## Authentication Flows

### Full OAuth 2.1 Authorization Code with PKCE

1. Discover endpoints at `/.well-known/openid-configuration`.
2. Start auth at `/oauth2/auth` with `response_type=code`, `code_challenge` and `S256`.
3. Login via `/login` (the example includes a simple form with demo credentials).
4. Exchange the `code` at `/oauth2/token` with `code_verifier` to obtain an access token.
5. Call `/mcp` with `Authorization: Bearer <access_token>`.

The server returns `401` with `WWW-Authenticate` advertising the AS and resource metadata when a valid token is not provided.

### Developer Modes (Optional)

- Static Auth Key: set `--embedded-auth-key <key>` and use it as the bearer token.
- Dev Tokens: set `--embedded-dev-tokens=true` and `--embedded-db <path>`. You can insert rows into `oauth_tokens` for quick testing; the middleware accepts them if not expired when introspection fails.
- Static User/Pass: set `--embedded-user` and `--embedded-pass` to enable a quick login for the Authorization Code flow without provisioning users in the DB.

## External OIDC Mode

`external_oidc` turns `go-go-mcp` into a protected resource that trusts an external issuer such as Keycloak. In this mode:

- `go-go-mcp` fetches the OIDC discovery document
- it fetches and caches the JWKS
- it validates bearer JWTs locally
- it checks issuer, time-based claims, optional audience, and optional scopes
- it still serves `/.well-known/oauth-protected-resource` and still emits `WWW-Authenticate`

Minimal example:

```go
err := embeddable.AddMCPCommand(rootCmd,
    embeddable.WithAuth(embeddable.AuthOptions{
        Mode:        embeddable.AuthModeExternalOIDC,
        ResourceURL: "https://mcp.example.com/mcp",
        External: embeddable.ExternalOIDCOptions{
            IssuerURL:      "https://keycloak.example.com/realms/myrealm",
            Audience:       "mcp-resource",
            RequiredScopes: []string{"mcp:invoke"},
        },
    }),
)
```

## Pluggable Authentication

The embedded OIDC login is now pluggable via an `Authenticator` interface used by the login handler:

- Default behavior: if `DBPath` is set, credentials are validated against `oauth_users` (bcrypt).
- Otherwise, a static user/pass is used (from flags or defaults).
- Advanced deployments can inject a custom authenticator when embedding the server programmatically.

### User Management (SQLite)

When `DBPath` is set, users are stored in `oauth_users` with bcrypt-hashed passwords. The CLI provides helpers:

```bash
mcp oidc users --db /tmp/mcp-oidc.db add --username alice --password secret
mcp oidc users --db /tmp/mcp-oidc.db passwd --username alice --password newsecret
mcp oidc users --db /tmp/mcp-oidc.db list
mcp oidc users --db /tmp/mcp-oidc.db del --username alice
```

### Tokens & Clients Admin

```bash
# tokens
mcp oidc tokens --db /tmp/mcp-oidc.db list
mcp oidc tokens --db /tmp/mcp-oidc.db del --token ABC123

# clients
mcp oidc clients --db /tmp/mcp-oidc.db list
mcp oidc clients --db /tmp/mcp-oidc.db upsert --id my-app --redirect-uri http://localhost/callback --redirect-uri http://localhost/return
```

## Protected Resource Metadata

The server exposes:

- `/.well-known/oauth-protected-resource` with:
  - `authorization_servers`: array containing the configured issuer
  - `resource`: the canonical resource (`<issuer>/mcp`)

This helps sophisticated clients discover the AS and the resource semantics per RFC 9728.

## Cancellation & Shutdown

Backends run with a shared context tied to process signals. On cancellation (e.g. Ctrl‑C), the HTTP server shuts down gracefully via `server.Shutdown`, closing connections and handlers cleanly.

## Logging & Troubleshooting

- OIDC endpoints log user‑agent and remote address for observability.
- The Bearer middleware logs missing/invalid tokens, static key usage, and successes (including `subject` and `client_id`).
- `/mcp` requests are logged (headers summarized, sensitive data censored).

If you see no logs when pointing a client to the server, check:

- Issuer correctness (must match the public base URL when using a tunnel like ngrok)
- That the client is actually hitting `/mcp` and not a different path
- That `Authorization` is present on protected endpoints

## Security Notes

- Prefer HTTPS issuers in production.
- Keep `EnableDevTokens=false` and `AuthKey` unset in production.
- Persist keys in SQLite to ensure stable JWKS across restarts.
- Consider adding audience/scope enforcement and rate limiting depending on exposure.

## Minimal Checklist

- Configure `WithOIDC` (Issuer must be the public base URL for discovery).
- Mount HTTP transport (SSE or streamable HTTP).
- Confirm OIDC endpoints respond and JWKS is available.
- Obtain a token (auth code with PKCE) or set a temporary `AuthKey` for smoke tests.
- Verify `/mcp` responds with `401` + `WWW-Authenticate` when unauthenticated.
- Verify `/mcp` accepts authenticated calls.
