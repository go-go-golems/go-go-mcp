# Design: Embedded OIDC Integration into go-go-mcp (No Proxy)

**Date:** January 9, 2025  
**Author:** Embedded OIDC design for `go-go-mcp`  
**Goal:** Merge the original OIDC server code directly into the go-go-mcp framework to let any app using the embeddable API expose an OIDC-protected MCP server without external proxies.

## Objectives

- Provide first-class, in-process OIDC support within `go-go-mcp`.
- Keep the embeddable API simple; minimal new options for basic use.
- Reuse the proven `idsrv` (Fosite-based) implementation, adapted into framework packages.
- Support HTTP transports (SSE and streamable HTTP) with bearer-protected `/mcp` out of the box.
- Remain optional: stdio remains unauthenticated; HTTP auth is opt-in.

In short, our aim is to make OIDC a first-class capability of the framework without forcing application authors to think about the details of OAuth 2.1, token lifecycles, discovery, keys, or persistence. By embedding this functionality, any Go application that already uses the embeddable API can gain secure, standards-compliant HTTP access to MCP tools with minimal code, enabling production-ready deployments and interoperability with modern clients.

## High-Level Architecture

```
App (uses embeddable API)
  └─ go-go-mcp embeddable
       ├─ Tool registry, session, middleware
       ├─ Backend (mcp-go): stdio | sse | streamable-http
       └─ OIDC Integration (new)
            ├─ pkg/auth/oidc: adapted from idsrv (Fosite)
            │    ├─ Discovery, AS metadata, JWKS, login, auth, token, DCR
            │    ├─ SQLite persistence for clients, keys, tokens, logs
            │    └─ Public API: Init/Routes/Introspect/Token helpers
            └─ HTTP backends (sse/streamable-http) gain auth hooks
                 ├─ Bearer middleware around /mcp
                 └─ RFC 9728 protected resource metadata
```

The architecture keeps clear boundaries: the embeddable layer remains the single entry point for application authors, while the new OIDC package encapsulates identity concerns. When an HTTP transport is chosen, the backend mounts both the MCP endpoint and the OIDC endpoints on the same server. A lightweight middleware enforces Bearer authentication for `/mcp`, delegating all protocol details (introspection, key management, dynamic client registration) to the OIDC package. This preserves the simplicity of the existing `mcp-go` integration and avoids leaking identity-specific complexity into tool implementations.

## Package Layout (new/changed)

- `pkg/auth/oidc/`
  - `server.go` (adapted from `go-go-labs/.../idsrv/idsrv.go`): Fosite provider, routes, SQLite, token helpers, tool-call logging
  - `exports.go`: small wrappers to expose key handlers (discovery, metadata, jwks, login, register)
  - `config.go`: configuration struct + validation
  - `persistence.go`: SQLite helpers

- `pkg/embeddable/`
  - `command.go` (unchanged public surface)
  - `server.go` (add minimal auth config plumbing)
  - `mcpgo_backend.go` (extend HTTP backends to mount OIDC routes and protect `/mcp` when enabled)
  - `README.md` (examples for OIDC-enabled apps)

This layout intentionally minimizes churn. The OIDC server is introduced as a cohesive package with a narrow, well-documented API. The embeddable package only gains a small configuration struct and a single option to toggle OIDC support. The HTTP backends receive the smallest possible adjustment: mounting routes and wrapping `/mcp` with a middleware when enabled. This balance reduces risk while making the feature broadly accessible.

## Public API Additions (Embeddable)

We add small, focused options—no large interface explosion:

```go
// OIDC options
type OIDCOptions struct {
    Issuer         string // e.g. https://myapp.example.com or http://localhost:3001
    DBPath         string // optional SQLite DB path for clients/keys/tokens
    EnableDevTokens bool  // optional; allow DB tokens for dev
}

func WithOIDC(opts OIDCOptions) ServerOption
```

These options are designed to be comprehensible at a glance. The `Issuer` defines the external base URL from which discovery documents and token endpoints are addressed. `DBPath` activates SQLite-backed persistence for registered clients, signing keys, and (optionally) development tokens, ensuring stable JWKS across restarts. `EnableDevTokens` provides a convenient, explicit escape hatch during local development and is discouraged in production. We deliberately avoid introducing a complex matrix of options in the MVP; the intent is to make the secure path the easy path.

Behavior:
- If `WithOIDC` is provided and transport is HTTP (`sse` or `streamable_http`), embeddable will:
  - Initialize the embedded OIDC server
  - Mount OIDC routes on the same HTTP server
  - Protect the `/mcp` endpoint with Bearer validation (Fosite introspection; optional dev tokens)
  - Serve `/.well-known/oauth-protected-resource`
- If stdio transport is used, OIDC is ignored (documented).

## OIDC Server (Adapted idsrv) – Key APIs

```go
package oidc

type Server struct {
    // fields: private key, issuer, fosite provider, memory store, sqlite path...
}

type Config struct {
    Issuer          string
    DBPath          string
    EnableDevTokens bool
    // Optional: default login user/pass for demo; overridable later
}

func New(cfg Config) (*Server, error)

// Attach standard OIDC/OAuth2 routes to mux
func (s *Server) Routes(mux *http.ServeMux)

// Token introspection and dev token support
func (s *Server) IntrospectAccessToken(ctx context.Context, token string) (subject, clientID string, ok bool, err error)

// Optional helpers for token persistence and tool-call logging (unchanged APIs)
func (s *Server) PersistToken(tr TokenRecord) error
func (s *Server) GetToken(token string) (TokenRecord, bool, error)
func (s *Server) LogMCPCall(entry MCPCallLog) error
```

Important: We keep the current behavior and table schemas intact to maximize reuse.

By retaining the original Fosite composition and SQLite schema, we preserve the robustness and operational knowledge already present in the `mcp-oidc-server`. The server exposes only the minimal operations the backend needs: attaching routes and validating access tokens. All other endpoints—discovery, JWKS, authorization, token—follow the standards and remain testable in isolation. This means operators can use their familiar OAuth/OIDC tooling for registration, flows, and diagnostics while developers focus on MCP tools.

## HTTP Backend Changes (sse, streamable-http)

Add an optional OIDC wiring path inside the existing backends, guarded by `WithOIDC` being present:

```go
// Pseudocode inside backend Start(...)
if cfg.oidcEnabled {
    // 1) Mount OIDC endpoints
    cfg.oidcServer.Routes(mux)

    // 2) Advertise protected resource metadata
    mux.HandleFunc("/.well-known/oauth-protected-resource", protectedResourceHandler(cfg))

    // 3) Protect /mcp with bearer validation (Fosite introspection + optional dev tokens)
    mcpHandler := mux.Handler("/mcp")
    mux.Handle("/mcp", oidcAuthMiddleware(cfg, mcpHandler))
}
```

The backend mounts identity and MCP on a shared HTTP listener, which simplifies deployment and avoids cross-process wiring. Importantly, MCP over stdio remains unchanged; OIDC enforcement only activates for HTTP transports. This duality gives application authors flexibility: local tools and CLI-driven flows can keep using stdio, while production-facing deployments can use the authenticated HTTP path without changing tool code.

### Auth Middleware (inline, minimal)

```go
func oidcAuthMiddleware(cfg *ServerConfig, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authz := r.Header.Get("Authorization")
        if !strings.HasPrefix(authz, "Bearer ") {
            advertiseWWWAuthenticate(w, cfg)
            http.Error(w, "missing bearer", http.StatusUnauthorized)
            return
        }
        tok := strings.TrimPrefix(authz, "Bearer ")
        subj, cid, ok, err := cfg.oidcServer.IntrospectAccessToken(r.Context(), tok)
        if err != nil || !ok {
            advertiseWWWAuthenticate(w, cfg)
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), ctxSubjectKey, subj)
        ctx = context.WithValue(ctx, ctxClientIDKey, cid)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

The middleware implements a clear, standards-friendly control point. It expects `Authorization: Bearer <token>`, validates it via the embedded OIDC server, and injects the authenticated subject and client ID into the request context. MCP tool handlers can retrieve these values to implement auditing or authorization checks as needed. Error responses include `WWW-Authenticate` with sufficient metadata for well-behaved clients to discover the correct authorization server and retry with a valid token.

### Protected Resource Metadata (RFC 9728)

```go
func protectedResourceHandler(cfg *ServerConfig) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        j := map[string]any{
            "authorization_servers": []string{cfg.OIDCOptions.Issuer},
            "resource": cfg.OIDCOptions.Issuer + "/mcp",
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(j)
    }
}

func advertiseWWWAuthenticate(w http.ResponseWriter, cfg *ServerConfig) {
    asMeta := cfg.OIDCOptions.Issuer + "/.well-known/oauth-authorization-server"
    prm := cfg.OIDCOptions.Issuer + "/.well-known/oauth-protected-resource"
    hdr := "Bearer realm=\"mcp\", resource=\"" + cfg.OIDCOptions.Issuer + "/mcp\"" + ", authorization_uri=\"" + asMeta + "\", resource_metadata=\"" + prm + "\""
    w.Header().Set("WWW-Authenticate", hdr)
}
```

Advertising protected resource metadata lets clients reliably discover the authorization server and understand the authentication requirements of the MCP endpoint. This pattern aligns with recent client expectations (including MCP-capable tools that support OAuth/OIDC flows) and reduces configuration friction. By emitting a standards-compliant `WWW-Authenticate` header on 401 responses, the server guides clients toward self-service remediation rather than opaque failures.

## Minimal Embeddable API Example

```go
root := &cobra.Command{Use: "myapp"}

err := embeddable.AddMCPCommand(root,
    embeddable.WithName("MyApp MCP Server"),
    embeddable.WithDefaultTransport("sse"),
    embeddable.WithDefaultPort(3001),
    embeddable.WithOIDC(embeddable.OIDCOptions{
        Issuer:          "http://localhost:3001",
        DBPath:          "/tmp/mcp-oidc.db",
        EnableDevTokens: false,
    }),
    embeddable.WithTool("search", searchHandler,
        embeddable.WithDescription("Search corpus"),
        embeddable.WithStringArg("query", "Query", true),
    ),
)
if err != nil { log.Fatal().Err(err).Msg("failed to add mcp") }

if err := root.Execute(); err != nil { log.Fatal().Err(err).Msg("cmd error") }
```

In this example, the application opts into the SSE transport, registers a simple `search` tool, and enables OIDC with a local issuer and SQLite persistence for keys and clients. Running the binary will expose both the OIDC endpoints (discovery, JWKS, authorization, token, registration) and the `/mcp` endpoint on the same port. A user can complete the authorization code with PKCE flow against the issuer, obtain an access token, and invoke MCP methods over HTTP using that token. No additional processes or proxies are needed.

## Migration of idsrv Code

- Copy `idsrv` logic into `pkg/auth/oidc/` with minimal changes:
  - Keep Fosite composition (`ComposeAllEnabled`) and global secret.
  - Keep login form and dev credentials (documented defaults; allow override later).
  - Keep SQLite schemas and helpers unchanged.
  - Rename package imports to new paths; remove standalone `main` dependencies.
- Add small shims in `exports.go` to expose the same handler surface for mux mounting.

Wherever possible, we keep the original implementation intact to reduce the risk of regressions. That includes preserving the demo login form and default credentials for local testing, while making them configurable for production. The persistence model remains straightforward: a single SQLite database manages registered clients and the signing key, ensuring JWKS stability across restarts and environments. The resulting package becomes a drop-in capability that is easy to reason about and test.

## Reference: Original Files and Symbols to Port or Reuse

This section maps the exact files and key symbols from the existing OIDC server to their new home in the framework so an implementer can copy them verbatim as a starting point and then make small, mechanical edits.

- Original files (source):
  - `go-go-labs/cmd/apps/mcp-oidc-server/pkg/idsrv/idsrv.go`
    - Symbols to keep (unchanged names where possible):
      - `type Server struct { ... }`
      - `func New(issuer string) (*Server, error)`
      - `func (s *Server) Routes(mux *http.ServeMux)`
      - `func (s *Server) InitSQLite(path string) error`
      - `func (s *Server) PersistToken(tr TokenRecord) error`
      - `func (s *Server) GetToken(token string) (TokenRecord, bool, error)`
      - `func (s *Server) ListTokens() ([]TokenRecord, error)`
      - `func (s *Server) LogMCPCall(entry MCPCallLog) error`
      - Discovery/login/token/register handlers: `oidcDiscovery`, `asMetadata`, `jwks`, `login`, `authorize`, `token`, `register`
      - Data types: `TokenRecord`, `MCPCallLog`
      - Internal helpers: `pemEncodeRSAPrivateKey`, `pemDecodeRSAPrivateKey`, `sqlOpen`, CSV helpers
    - Minimal behavior edits when embedding:
      - Rename package to `oidc` and move file to `go-go-mcp/pkg/auth/oidc/server.go`.
      - Replace `log` imports with framework logger if needed (current code already uses `github.com/rs/zerolog/log`).
      - Keep Fosite composition (`compose.ComposeAllEnabled`) and `fosite.Config` fields (`IDTokenIssuer`, `EnforcePKCEForPublicClients`, `GlobalSecret`).
      - Keep `loginTpl` and `cookieName` as-is; optionally make credentials configurable later.
  - `go-go-labs/cmd/apps/mcp-oidc-server/pkg/idsrv/exports.go`
    - Small exported wrappers: `RoutesDiscovery`, `RoutesASMetadata`, `Authorize`, `Token`, `Register`, `Login`.
    - When embedding, you may omit these if `Routes` is mounted wholesale; otherwise preserve for granular mounting. New path: `go-go-mcp/pkg/auth/oidc/exports.go`.
  - `go-go-labs/cmd/apps/mcp-oidc-server/pkg/server/server.go`
    - DO NOT copy the JSON-RPC MCP implementation (`handleMCP`, `rpcRequest/Response/Error`, `sampleDocs`). The framework already provides MCP via `mcp-go`.
    - DO copy or adapt the Bearer auth logic under `mcpAuthMiddleware` and the auth-context helpers:
      - `type ctxKey string`, `const ctxSubjectKey`, `const ctxClientIDKey`, `setAuthCtx`, `getAuthCtx`.
      - These should move into the HTTP backend integration in `go-go-mcp/pkg/embeddable/mcpgo_backend.go` (or a small `authctx` helper) so that `/mcp` requests carry `subject` and `client_id` in context.
    - Useful utilities that can be retained conceptually:
      - `writeJSONWithPreview` (optional), `LoggingMiddleware` (the framework may already log; avoid duplication).

Literal copy is viable and recommended for a first pass. After copying, perform the following mechanical edits:

1) Change the package name from `idsrv` to `oidc` and update imports to the new path `github.com/go-go-golems/go-go-mcp/pkg/auth/oidc` where referenced.
2) Remove any references to the standalone `main.go` or CLI commands (`tokens`, `list-clients`)—those belong to the old app; the framework won’t expose them directly. If desired, we can reintroduce an embeddable `mcp oidc` subcommand later.
3) Ensure the SQLite driver import remains (`_ "github.com/mattn/go-sqlite3"`) in `persistence.go` or the top of `server.go` to preserve runtime behavior when `DBPath` is set.
4) Keep the SQL table schemas as-is: `oauth_clients`, `oauth_keys`, `oauth_tokens`, and `mcp_tool_calls`.
5) Verify that `GlobalSecret` in `fosite.Config` is set to a 32-byte value (unchanged from the source) and that the RSA key is persisted when SQLite is enabled.

## Symbol Mapping: Old to New (Embedding)

- `idsrv.New(issuer string)` → `oidc.New(oidc.Config{Issuer: ...})`
  - New variant takes a config struct; internally calls the same Fosite composition.
- `idsrv.Server.Routes(mux)` → `oidc.Server.Routes(mux)` (unchanged behavior)
- `idsrv.Server.InitSQLite(path)` → `oidc.Server.InitSQLite(path)` (unchanged signature)
- `idsrv.ProviderRef().IntrospectToken(...)` usage in `mcpAuthMiddleware` → `oidc.Server.IntrospectAccessToken(ctx, token)` helper, implemented as a thin wrapper around Fosite introspection with `openid.DefaultSession`.

Example implementation of the new helper:

```go
// inside pkg/auth/oidc/server.go
func (s *Server) IntrospectAccessToken(ctx context.Context, token string) (string, string, bool, error) {
    sess := new(openid.DefaultSession)
    tt, ar, err := s.Provider.IntrospectToken(ctx, token, fosite.AccessToken, sess)
    if err != nil {
        // Optional dev-token fallback if DB enabled and feature flag on
        if s.dbPath != "" && s.enableDevTokens {
            rec, ok, derr := s.GetToken(token)
            if derr == nil && ok && time.Now().Before(rec.ExpiresAt) {
                return rec.Subject, rec.ClientID, true, nil
            }
        }
        return "", "", false, err
    }
    _ = tt
    subject := ""
    if sess.Claims != nil { subject = sess.Claims.Subject }
    clientID := ""
    if ar != nil && ar.GetClient() != nil { clientID = ar.GetClient().GetID() }
    return subject, clientID, true, nil
}
```

## What Not to Port (and Why)

Do not port the HTTP JSON-RPC MCP endpoint from `pkg/server/server.go` (`handleMCP`, request/response structs, and the `sampleDocs` data). The framework already integrates MCP using `mcp-go` with multiple transports and a full tool registry. Reintroducing a second, parallel MCP implementation would create confusion and duplication. Instead, reuse only the auth-context and bearer validation logic to wrap the framework’s `/mcp` route.

## Copy-First Strategy: Pros and Cons

Copying the `idsrv` code nearly verbatim into `pkg/auth/oidc/` is a pragmatic starting point.

- Pros:
  - Fast path to a working embedded OIDC server with minimal risk.
  - Preserves well-tested flows (discovery, JWKS, auth code with PKCE, token issuance).
  - Maintains SQLite schemas and key persistence, ensuring stability across restarts.
  - Eases operational migration: docs and troubleshooting steps remain applicable.

- Cons:
  - Some duplicated scaffolding (e.g., HTML template and cookie names) may need polish later.
  - The API surface initially mirrors the standalone server; a later refactor might slim or reshape it.
  - Additional testing is needed to validate coexistence with the framework’s logging and HTTP server lifecycle.

On balance, the copy-first approach aligns with our goals: get an embedded, standards-compliant OIDC layer into developers’ hands quickly, then iterate toward deeper integration (per-tool authorization, customizable login UI, and broader persistence options) based on real-world feedback.

## MVP Scope vs. Later Enhancements

MVP (embed and protect `/mcp`):
- Implement `pkg/auth/oidc` with the existing features (discovery, jwks, login, auth, token, DCR, SQLite, token helpers).
- Add `WithOIDC` and wire into HTTP backends only.
- Add RFC 9728 handler and `WWW-Authenticate`.
- Expose subject/client_id in request context for tools.

Later:
- Add per-tool scope checks and audience enforcement.
- Add configurable login UI hooks.
- Extract persistence interfaces (if needed beyond SQLite).
- Add optional stdio bearer check (if there’s a standardized pattern).

This phasing reflects a pragmatic approach. The MVP secures `/mcp` for HTTP transports with minimal API footprint and excellent compatibility. As usage matures, we can introduce per-tool authorization (mapping scopes to tool names or operations), audience enforcement for tighter token constraints, and configurable UI hooks for teams that need to integrate with existing identity experiences.

## Testing Plan

- Unit tests:
  - OIDC server configuration and route mounting.
  - Introspection path and 401 behavior.
  - Context propagation of subject/client_id.
- Integration tests:
  - Full authorization code + PKCE flow using the embedded server.
  - `tools/list` and `tools/call` via `/mcp` with valid/invalid tokens.
  - DCR (`/register`) with subsequent login + tool invocation.

We will maintain test fixtures that spin up the embedded OIDC server and HTTP backend in-process, exercising happy-path and error-path flows. Tests will assert token validation behavior, correct `WWW-Authenticate` headers, and the propagation of subject/client ID into tool contexts. Because the OIDC package preserves its original table layouts and flows, many existing operational playbooks and scripts remain applicable and can be adapted into automated tests.

## Security Notes

- Production should disable dev tokens; defaults off.
- Encourage HTTPS for Issuer in production; document TLS termination strategy.
- Consider adding `aud`/scope checks in follow-up.

Security posture improves meaningfully even in the MVP. Tokens are verified using the same Fosite primitives as in the original server, signing keys are persisted and rotated deliberately, and sensitive endpoints are co-located behind a single listener that can sit behind standard TLS termination. Documentation will clearly call out production guidance: disable development tokens, prefer HTTPS issuers, and consider additional controls such as CORS, rate limiting, and WAF/CDN integration depending on exposure.

## Rollout Plan

1) Land `pkg/auth/oidc` (adapted idsrv) + `WithOIDC` plumbing.  
2) Update examples and docs.  
3) Invite feedback; iterate on per-tool auth and configurability.

We will start by gating the feature behind the `WithOIDC` option to avoid surprising existing users. The first release will prioritize reliability and ergonomics over breadth. Feedback from early adopters—particularly around configuration ergonomics and per-tool authorization needs—will inform the next iteration. We will provide a migration note for users of the standalone OIDC server, outlining how to achieve equivalent functionality with fewer moving parts.


