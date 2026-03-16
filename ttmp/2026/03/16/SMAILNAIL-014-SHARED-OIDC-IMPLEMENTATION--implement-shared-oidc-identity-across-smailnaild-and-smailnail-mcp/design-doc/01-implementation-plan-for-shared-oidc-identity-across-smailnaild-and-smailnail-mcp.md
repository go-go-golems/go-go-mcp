---
Title: Implementation plan for shared OIDC identity across smailnaild and smailnail MCP
Ticket: SMAILNAIL-014-SHARED-OIDC-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - mcp
    - oidc
    - authentication
    - keycloak
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/user.go
      Note: Current fallback identity resolver that should be replaced by a real authenticated principal resolver
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go
      Note: Hosted app HTTP routes where auth session, me, and protected API middleware will be introduced
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go
      Note: Existing schema bootstrap to extend with identity and session tables
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external.go
      Note: Existing verified OIDC principal extraction path for the MCP server
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: Existing MCP auth middleware boundary where richer principal propagation should be introduced
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go
      Note: MCP server construction point that must be taught to resolve local user context
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go
      Note: Tool path that should stop taking only raw credentials and gain account ownership resolution
ExternalSources:
    - https://openid.net/specs/openid-connect-core-1_0.html
    - https://openid.net/specs/openid-connect-discovery-1_0.html
    - https://datatracker.ietf.org/doc/html/rfc7591
    - https://www.keycloak.org/docs/latest/server_admin/
Summary: Detailed implementation plan for one shared, provider-neutral OIDC identity model spanning browser login in smailnaild and bearer-authenticated requests to smailnail MCP.
LastUpdated: 2026-03-16T11:28:46.457011378-04:00
WhatFor: Explain the target shared identity architecture and break it into concrete execution phases that can be implemented against the current codebase.
WhenToUse: Use when implementing real authentication and identity mapping for hosted smailnail and its MCP service.
---

# Implementation plan for shared OIDC identity across smailnaild and smailnail MCP

## Executive Summary

`smailnail` needs one local identity model that both the hosted web app and the hosted MCP can trust. The right model is:

- use Keycloak only as the current OpenID Provider
- keep the application identity key provider-neutral as `(issuer, subject)`
- use separate OIDC clients for browser and MCP use cases
- create one local `users` row per external identity
- attach IMAP accounts, rules, tests, and future mailbox artifacts to that local user

This design deliberately does not make the web UI "reuse the MCP auth flow" directly. Instead, it reuses:

- the same OIDC provider
- the same realm
- the same end-user identity
- the same local-user mapping logic

The browser flow and the MCP flow remain different at the transport layer:

- browser: Authorization Code + PKCE + secure session cookie
- MCP: bearer token validation using the existing external OIDC path

But after authentication, both flows converge into the same local user lookup and the same account ownership model.

## Problem Statement

The current system has an architectural split:

- `smailnail-mcp` can validate OIDC tokens from Keycloak through `go-go-mcp`
- `smailnaild` has no real login and falls back to `local-user`
- account ownership in the hosted database is therefore local-dev only
- authenticated MCP requests do not yet resolve the stored IMAP accounts belonging to the user

That produces three concrete problems:

1. The browser and the MCP do not share a user identity model.
2. The hosted web UI cannot safely manage real user-owned IMAP credentials.
3. The MCP cannot consume the hosted configuration that the browser creates.

If left as-is, the product becomes two disconnected systems:

- one browser app that stores account data under a fake user
- one MCP server that authenticates correctly but has no path from token to account ownership

That is the wrong foundation for a hosted product.

## Proposed Solution

### Identity Model

Use the OIDC issuer URL and the OIDC subject claim as the canonical external identity key.

```text
external_identity_key = (issuer, subject)
```

Where:

- `issuer` is the exact `iss` URL from the token or session metadata
- `subject` is the OIDC `sub` claim

Important semantics:

- `sub` is stable only within one issuer
- `client_id` identifies the application client, not the person
- `aud` identifies the intended audience of the token
- `azp` may identify the authorized party for multi-audience cases
- `tenant_id` is provider-specific and optional; it can be stored as metadata but must not replace `(issuer, subject)` as the main key

### Local User Model

Add local user and external identity storage to the hosted application database.

Suggested schema shape:

```text
users
  id                  TEXT PRIMARY KEY
  primary_email       TEXT NULL
  display_name        TEXT NULL
  avatar_url          TEXT NULL
  created_at          TIMESTAMP NOT NULL
  updated_at          TIMESTAMP NOT NULL

user_external_identities
  id                  TEXT PRIMARY KEY
  user_id             TEXT NOT NULL REFERENCES users(id)
  issuer              TEXT NOT NULL
  subject             TEXT NOT NULL
  provider_kind       TEXT NOT NULL
  email               TEXT NULL
  email_verified      BOOLEAN NOT NULL
  preferred_username  TEXT NULL
  raw_claims_json     TEXT NULL
  created_at          TIMESTAMP NOT NULL
  updated_at          TIMESTAMP NOT NULL
  UNIQUE (issuer, subject)

web_sessions
  id                  TEXT PRIMARY KEY
  user_id             TEXT NOT NULL REFERENCES users(id)
  issuer              TEXT NOT NULL
  subject             TEXT NOT NULL
  expires_at          TIMESTAMP NOT NULL
  created_at          TIMESTAMP NOT NULL
  last_seen_at        TIMESTAMP NOT NULL
```

Then update the existing tables so ownership points to the local user:

- `imap_accounts.user_id`
- `rules.user_id`
- future dry-run history, saved inbox state, and audit tables

### Web Flow

For `smailnaild`, implement a normal browser OIDC login flow:

```text
Browser
  -> GET /auth/login
  -> redirect to Keycloak authorize endpoint
  -> user authenticates
  -> GET /auth/callback?code=...
  -> backend exchanges code for tokens
  -> backend validates ID token
  -> backend upserts (issuer, subject) into local user storage
  -> backend creates secure session cookie
  -> browser calls GET /api/me
  -> protected APIs use session-backed principal
```

The React app should not be the source of truth for identity. The backend should own:

- the authorization code exchange
- token verification
- user provisioning
- session storage
- secure cookie issuance

### MCP Flow

For the hosted MCP, keep using external OIDC bearer validation, but extend the principal handling:

```text
Remote MCP client
  -> presents bearer token
  -> go-go-mcp validates issuer/signature/scope/audience
  -> smailnail principal mapper extracts (issuer, subject, claims)
  -> smailnail local-user service resolves or provisions the local user
  -> MCP tools run with local user context
  -> tool can access user's stored IMAP accounts by ownership
```

This requires two code-level changes:

1. Carry a richer principal structure than only forwarded headers.
2. Teach `executeIMAPJS` and future tools how to resolve the local user before accessing stored accounts.

### Shared User Resolution Layer

Create one shared service in `smailnail`, conceptually:

```go
type ExternalPrincipal struct {
    Issuer            string
    Subject           string
    ClientID          string
    Audience          []string
    Scopes            []string
    Email             string
    EmailVerified     bool
    PreferredUsername string
    Claims            map[string]any
}

type UserIdentityService interface {
    ResolveOrProvisionUser(ctx context.Context, principal ExternalPrincipal) (*User, error)
}
```

Both the web auth callback and the MCP auth boundary should call this same service.

### Why This Is Not Keycloak-Specific

The system must not hard-code Keycloak semantics into the user model. Keycloak is only the current issuer.

Provider-neutral rules:

- canonical external key is `(issuer, subject)`
- email is profile data, not the primary identity key
- `tenant_id` is optional metadata only
- the provider can be Keycloak today and something else later
- protocol mappers may enrich tokens, but local ownership rules must not depend on provider-specific claims

The only Keycloak-specific pieces should live in configuration and deployment:

- realm URL
- client IDs
- protocol mapper choices
- login/logout endpoint details

## Design Decisions

### Decision 1: Use separate OIDC clients for browser and MCP

Rationale:

- browser and MCP have different redirect and security requirements
- browser sessions want Authorization Code + PKCE
- MCP connectors may rely on dynamic registration or remote bearer flows
- rotating one client does not break the other

Expected clients:

- `smailnail-web`
- `smailnail-mcp`

### Decision 2: Use `(issuer, subject)` as the stable key

Rationale:

- this matches OIDC semantics
- it survives email changes and username changes
- it is provider-neutral

Rejected identifiers:

- `email`: mutable and sometimes absent
- `preferred_username`: display-friendly but mutable
- `client_id`: identifies the app, not the user

### Decision 3: Keep browser auth server-side

Rationale:

- the backend already owns the protected APIs
- session cookies are simpler for the frontend than token storage
- it avoids leaking auth complexity into every React API call

### Decision 4: Auto-provision the local user on first successful login or bearer-authenticated request

Rationale:

- lowest-friction hosted onboarding
- avoids pre-seeding user rows
- keeps browser and MCP flows symmetric

### Decision 5: Store IMAP credentials in the app DB, not in Keycloak

Rationale:

- Keycloak is the IdP, not the application secret vault
- the app owns the IMAP account lifecycle
- multiple IMAP accounts per user are an application concern
- this aligns with the existing encrypted account storage in `smailnaild`

## Alternatives Considered

### Alternative A: Put all auth in the frontend

Rejected because:

- the UI would need to manage tokens directly
- API calls would become more complex
- session revocation and logout become harder to centralize
- it does not help the MCP side at all

### Alternative B: Reuse one client for web and MCP

Rejected because:

- redirect URIs and consent behaviors differ
- operational blast radius is larger
- it couples browser and MCP rollout unnecessarily

### Alternative C: Key off email address

Rejected because:

- email is mutable
- some providers do not guarantee a stable, verified email claim
- account collisions become easier during provider migration

### Alternative D: Store IMAP credentials in Keycloak user attributes

Rejected because:

- wrong separation of concerns
- secrets management becomes awkward
- multiple IMAP accounts and account metadata fit poorly there
- app-side encryption and rotation are easier in the application database

## Implementation Plan

### Phase 0: Preserve current account setup work

Do not block the existing account-setup ticket. This ticket layers authentication and shared identity on top of it.

Dependencies:

- `SMAILNAIL-013` backend account storage and APIs

### Phase 1: Shared identity schema and services

Build the provider-neutral local identity layer first.

Deliverables:

- `users`
- `user_external_identities`
- `web_sessions`
- repositories and service interfaces
- `ResolveOrProvisionUser` service

Acceptance criteria:

- given `(issuer, subject)` the app can consistently resolve one local user
- repeated logins do not create duplicate users

### Phase 2: Browser OIDC login for smailnaild

Add hosted web login and logout.

Deliverables:

- OIDC config section for `smailnaild`
- `/auth/login`
- `/auth/callback`
- `/auth/logout`
- `/api/me`
- session cookie middleware
- protected API middleware

Acceptance criteria:

- anonymous browser requests to protected APIs return `401`
- login creates a local user and a session
- the React app can call `/api/me` and render the logged-in identity

### Phase 3: Replace dev user resolution in hosted APIs

Remove the implicit `local-user` behavior for normal hosted mode.

Deliverables:

- auth-aware user resolver
- development override mode only when explicitly enabled
- tests for anonymous, logged-in, and logged-out states

Acceptance criteria:

- production-style mode has no implicit user fallback
- local development still has an explicit escape hatch if needed

### Phase 4: MCP principal to local-user mapping

Extend the MCP auth boundary to use the same shared user identity layer.

Deliverables:

- richer principal propagation from `go-go-mcp`
- smailnail principal adapter
- local user provisioning/lookup from bearer-authenticated requests

Acceptance criteria:

- the same Keycloak person logging into the browser and calling the MCP resolves to the same local `user_id`

### Phase 5: Account ownership and tool resolution

Teach the MCP to consume stored user-owned IMAP accounts instead of only raw credentials.

Deliverables:

- tool changes so JS execution can reference stored account IDs
- ownership checks on account reads
- execution context that includes the local user

Acceptance criteria:

- MCP tools can access only the caller's accounts
- the browser-created account can be used by the same authenticated user through MCP

### Suggested Commit Slicing

1. schema + repositories for users/external identities/sessions
2. shared user identity service
3. web auth config and callback/login/logout endpoints
4. session middleware and `/api/me`
5. replacement of `HeaderUserResolver` default behavior
6. richer MCP principal propagation
7. account resolution inside MCP tools

### Pseudocode

```go
func HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {
    tokens := oidcClient.ExchangeCode(r.Context(), r.URL.Query().Get("code"))
    idToken := verifier.Verify(tokens.IDToken)

    principal := ExternalPrincipal{
        Issuer: idToken.Issuer,
        Subject: idToken.Subject,
        Email: idToken.Email,
        EmailVerified: idToken.EmailVerified,
        PreferredUsername: idToken.PreferredUsername,
        Claims: idToken.Claims,
    }

    user := userIdentityService.ResolveOrProvisionUser(r.Context(), principal)
    session := sessionService.Create(r.Context(), user.ID, principal)
    setSessionCookie(w, session)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

```go
func ResolveMCPUser(ctx context.Context, principal ExternalPrincipal) (*User, error) {
    user, err := userIdentityService.ResolveOrProvisionUser(ctx, principal)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

### Flow Diagram

```text
             +--------------------+
             |      Keycloak      |
             +--------------------+
                ^             ^
                |             |
     browser OIDC |             | bearer token / discovery / JWKS
                |             |
        +---------------+   +------------------+
        |  smailnaild   |   | smailnail-mcp    |
        | web backend   |   | hosted MCP       |
        +---------------+   +------------------+
                \             /
                 \           /
                  v         v
             +--------------------+
             | user identity svc  |
             | (issuer, subject)  |
             +--------------------+
                       |
                       v
             +--------------------+
             | app SQL database   |
             | users, identities, |
             | sessions, accounts |
             +--------------------+
```

## Open Questions

Open questions to answer during implementation:

- Should first-login provisioning be allowed from MCP requests, or only from browser login?
- Do we want long-lived DB-backed sessions or signed stateless cookies for the web app?
- Should the browser and MCP share the same local-role model from day one, or postpone authorization roles until later?
- Which claims should be persisted beyond `issuer`, `subject`, `email`, and `preferred_username`?
- Do we want a dedicated `/api/me/accounts` bootstrap endpoint for the frontend home screen?

## References

- OIDC Core: https://openid.net/specs/openid-connect-core-1_0.html
- OIDC Discovery: https://openid.net/specs/openid-connect-discovery-1_0.html
- OAuth Dynamic Client Registration: https://datatracker.ietf.org/doc/html/rfc7591
- Keycloak Server Admin Guide: https://www.keycloak.org/docs/latest/server_admin/
- Related ticket: `SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION`
