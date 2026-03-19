---
Title: Implementation plan for merging smailnaild and MCP into one hosted server
Ticket: SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - oidc
    - keycloak
    - coolify
    - deployments
    - architecture
DocType: design-doc
Intent: long-term
Summary: Detailed design and rollout plan for serving the SPA, browser auth flow, application API, and MCP over one hosted HTTP server.
---

# Implementation plan for merging smailnaild and MCP into one hosted server

## Executive summary

The current split deployment was the right bootstrap strategy, but it is no longer the right product shape. The hosted web app and the hosted MCP are now coupled by the same identity model, the same local-user tables, the same encrypted IMAP credential storage, and the same operational server. Running them as separate deployed applications creates unnecessary divergence risk.

The target design is to make `smailnaild` the single hosted binary and server surface. The MCP HTTP transport should be mounted into the existing hosted router instead of booting a separate standalone server. The standalone `smailnail-imap-mcp` binary can remain as a thin compatibility wrapper for local testing and older workflows, but it should no longer define the primary hosted runtime shape.

## Desired runtime shape

```text
Internet
  -> Traefik / Coolify
  -> smailnaild :8080 (single container, single process)
       /                                     -> SPA
       /auth/login                          -> browser OIDC redirect
       /auth/callback                       -> browser OIDC callback
       /auth/logout                         -> browser logout
       /api/me                              -> session-backed identity
       /api/accounts/*                      -> session-backed app API
       /api/rules/*                         -> session-backed app API
       /.well-known/oauth-protected-resource -> public MCP resource metadata
       /mcp                                 -> bearer-authenticated MCP
```

This design keeps one HTTP server while preserving two different security contracts.

## Why this design makes sense

### Product reasons

- Users experience one application, not two unrelated endpoints.
- The web UI that stores IMAP accounts and the MCP that consumes them are conceptually one product.
- A single origin simplifies mental models for hosted usage and later onboarding.

### Engineering reasons

- One image and one deployment reduce config drift.
- The MCP route can reuse the same DB pool, encryption config, and identity service.
- Centralized request logging and health reporting become simpler.
- Debugging cross-surface issues becomes much easier because requests land in one process.

### Operational reasons

- One Coolify app instead of two custom apps.
- One set of secrets for DB and encryption.
- One domain to monitor, roll back, and smoke test.
- One deployment story to hand off.

## What must stay separate

The main implementation trap is to conflate browser auth and MCP auth. They are not the same thing, even when served by one server.

### Web app and API

- Uses browser redirects to the OIDC provider
- Uses state and nonce cookies
- Stores a local session cookie after successful login
- Reads the local user via the session store

### MCP

- Uses bearer access tokens presented on each request
- Validates tokens against the OIDC issuer metadata and JWKS
- Does not need a callback route of its own
- Resolves the local user from the validated `(issuer, subject)` principal

The merge should unify deployment, not weaken these contracts.

## Existing code shape

### Hosted app

- [http.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go)
- [serve.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnaild/commands/serve.go)
- [oidc.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/auth/oidc.go)

This side already owns the main Chi router, static asset serving, app API, and browser-session identity.

### Standalone MCP server

- [server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go)
- [execute_tool.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go)
- [main.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail-imap-mcp/main.go)

This side currently bootstraps its own command surface and its own HTTP server assumptions.

## Proposed refactor

### Step 1: Extract reusable MCP handler construction

Refactor the MCP package so it can produce a mounted HTTP handler instead of assuming it owns the entire standalone server process. The key output should be something close to:

```go
type MountedMCP struct {
    Handler http.Handler
    MetadataHandler http.Handler
}

func NewMountedMCP(cfg Config, deps Deps) (*MountedMCP, error)
```

The exact shape can differ, but the important part is:

- `smailnaild` can mount `/mcp`
- `smailnaild` can mount `/.well-known/oauth-protected-resource`
- the same code can still be used by the standalone compatibility binary

### Step 2: Promote shared dependencies upward

The following should be created once in `smailnaild serve` and then passed to both the app API and the MCP layer:

- SQL database handle
- encryption config/service
- identity service
- account repository/service
- shared logger

This reduces the chance that the web app and MCP each open their own incompatible dependency graph.

### Step 3: Merge config surfaces carefully

The merged `serve` command needs two groups of auth settings:

- web OIDC client settings
- MCP external OIDC resource settings

That should remain explicit in the CLI to avoid hidden coupling. A good command shape is:

```text
smailnaild serve
  --auth-mode oidc
  --oidc-issuer-url ...
  --oidc-client-id ...
  --oidc-client-secret ...
  --oidc-redirect-url ...
  --mcp-auth-mode external_oidc
  --mcp-auth-resource-url ...
  --mcp-oidc-issuer-url ...
  --mcp-oidc-audience ...
  --mcp-oidc-required-scope ...
```

If the implementation reuses the same issuer by default, that is fine, but the config model should still make the two surfaces readable.

### Step 4: Mount MCP routes into the hosted router

The hosted router should explicitly mount:

- `/.well-known/oauth-protected-resource`
- `/mcp`

The rest of the app stays unchanged.

### Step 5: Keep or retire the standalone MCP binary

Recommended short-term shape:

- keep `cmd/smailnail-imap-mcp`
- make it a thin wrapper around the shared mounted-MCP package
- mark it as compatibility-oriented rather than the primary hosted runtime

This avoids breaking existing scripts or smoke tests too early.

## Deployment target

The target hosted deployment should be one Coolify application for the merged server, likely on:

- `https://smailnail.scapegoat.dev/`

With:

- web app at `/`
- app API at `/api/*`
- browser login at `/auth/*`
- MCP metadata at `/.well-known/oauth-protected-resource`
- MCP endpoint at `/mcp`

This assumes the single public origin model is acceptable for the remote MCP consumers you care about. It should be.

## Browser and MCP auth flow together

```text
Browser user
  -> GET /auth/login
  -> redirect to Keycloak
  -> /auth/callback
  -> local session cookie
  -> /api/accounts

MCP client
  -> discover protected-resource metadata
  -> obtain bearer token from Keycloak
  -> POST /mcp with Authorization: Bearer ...
  -> local principal resolution
  -> account lookup by users.id
```

Both flows should resolve to the same local user tables:

```text
(issuer, subject)
   -> user_external_identities
   -> users.id
   -> imap_accounts.user_id
   -> rules.user_id
```

## Coolify rollout strategy

Do not cut directly from the split deployment to the merged deployment without an overlap window.

Recommended rollout:

1. Merge code locally and validate against local Keycloak + local Dovecot.
2. Build and run the merged container locally.
3. Create a new Coolify app for the merged server on a staging or temporary domain if available.
4. Validate:
   - browser login
   - `/api/me`
   - account create/test
   - `/.well-known/oauth-protected-resource`
   - `/mcp` unauthenticated `401`
   - bearer-authenticated `executeIMAPJS` using stored `accountId`
5. Only then cut traffic or repoint the main domain.

If a temporary domain is not available, use a temporary Coolify app with a generated subdomain first.

## Required docs and playbooks

This ticket should produce:

- a merged deployment guide
- an operator test playbook for local and hosted validation
- explicit migration notes from split deployment to merged deployment
- a handoff note describing whether `smailnail-imap-mcp` still exists and why

## Risks

### Risk: route auth confusion

If web middleware accidentally wraps `/mcp`, or MCP bearer auth accidentally wraps the browser API, the hosted system will become unusable or subtly insecure.

Mitigation:

- explicit route grouping
- dedicated tests for both auth boundaries

### Risk: config drift during migration

If the merged server uses different DB or encryption settings than the current MCP app, stored account resolution will fail even though auth works.

Mitigation:

- test the merged server against the existing production-like shared DB settings before cutover

### Risk: remote client assumptions

Some MCP clients may cache the old resource URL or expect the old hostname.

Mitigation:

- maintain the old domain during transition if needed
- document any hostname change clearly

## Pseudocode sketch

```go
func buildHostedServer(cfg Config) (*http.Server, error) {
    db := openAppDB(cfg.DB)
    crypto := newSecrets(cfg.Encryption)
    identitySvc := identity.NewService(...)
    appRouter := smailnaild.NewRouter(...)

    mcpHandler, err := imapjs.NewMountedMCP(imapjs.Config{
        AuthMode: cfg.MCP.AuthMode,
        ResourceURL: cfg.MCP.ResourceURL,
        IssuerURL: cfg.MCP.IssuerURL,
        AppDB: db,
        Encryption: crypto,
        IdentityService: identitySvc,
    })
    if err != nil {
        return nil, err
    }

    root := chi.NewRouter()
    root.Mount("/auth", appRouter.Auth)
    root.Mount("/api", appRouter.API)
    root.Handle("/.well-known/oauth-protected-resource", mcpHandler.MetadataHandler)
    root.Handle("/mcp", mcpHandler.Handler)
    root.Handle("/*", appRouter.SPA)

    return &http.Server{
        Addr: cfg.ListenAddr,
        Handler: root,
    }, nil
}
```

## Definition of done

This ticket is done when all of the following are true:

- `smailnaild` serves SPA, app API, browser auth, and MCP over one `http.Server`
- bearer-authenticated MCP calls can resolve stored account ownership from the same shared app DB used by the web app
- the merged server is deployed on Coolify
- the hosted browser login and the hosted MCP smoke both pass
- deployment and rollback instructions are documented
