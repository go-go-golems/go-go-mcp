---
Title: Investigation diary
Ticket: SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE
Status: complete
Topics:
    - smailnail
    - mcp
    - keycloak
    - oidc
    - authentication
    - sql
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external.go
      Note: Reviewed to document the exact external OIDC verification steps and the current limited principal model
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: Reviewed to confirm how the HTTP middleware propagates authentication into downstream requests today
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go
      Note: Reviewed to confirm that the smailnail MCP handler ignores authenticated identity today
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go
      Note: Reviewed to document the current raw-credential requirement in the IMAP connection path
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go
      Note: Reviewed as the starting point for provider-neutral app-side user and secret storage
ExternalSources:
    - https://openid.net/specs/openid-connect-core-1_0.html
    - https://openid.net/specs/openid-connect-discovery-1_0.html
    - https://datatracker.ietf.org/doc/html/rfc6749
    - https://datatracker.ietf.org/doc/html/rfc7519
Summary: Chronological investigation notes for the intern guide on OIDC identity mapping and IMAP credential storage in hosted smailnail.
LastUpdated: 2026-03-16T09:18:00-04:00
WhatFor: Record the evidence gathering and design reasoning behind the intern-facing OIDC and credential-storage guide.
WhenToUse: Use when reviewing how the guide was assembled, retracing claims back to code, or continuing the implementation work from this ticket.
---

# Investigation diary

## Goal

Create a detailed intern-facing explanation of how hosted `smailnail-mcp` currently authenticates requests with OIDC, why that is not yet enough for per-user IMAP access, and how to implement a provider-neutral identity-to-user-to-credential mapping layer.

## 2026-03-16

### Step 1: created a dedicated ticket workspace

I created `SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE` under `go-go-mcp/ttmp/2026/03/16` so the guide would be separate from the earlier deployment and DCR debugging tickets. The intent was to produce a durable onboarding document rather than an implementation diff.

### Step 2: inspected the current authentication code path

I reviewed [auth_provider_external.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external.go) and [auth_provider.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider.go).

The important findings were:

- discovery and JWKS handling are already generic OIDC, not Keycloak-specific
- issuer, token expiry, audience, and required scopes are already enforced
- the resulting `AuthPrincipal` only contains `Subject`, `ClientID`, `Issuer`, and `Scopes`

That made the main application gap obvious: the system validates bearer tokens correctly but does not carry enough typed identity information into the application layer.

### Step 3: inspected the MCP middleware propagation path

I reviewed [mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go).

The decisive finding was that successful authentication currently results in:

- request cloning
- `X-MCP-Subject` header injection
- `X-MCP-Client-ID` header injection

There is no typed principal attached to `context.Context`, and there is no downstream helper such as `GetAuthPrincipal(ctx)`.

That means the authentication layer is still operating more like infrastructure than application identity.

### Step 4: confirmed that `smailnail` does not consume identity today

I reviewed [execute_tool.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go).

The handler:

- binds raw arguments
- creates a Goja runtime
- registers the `smailnail` module
- runs arbitrary JavaScript

There is no identity lookup, no user repository, and no IMAP account resolution based on the authenticated caller.

### Step 5: confirmed that the IMAP layer still expects raw credentials

I reviewed [service.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go).

The key finding was the existing `ConnectOptions` structure:

- `Server`
- `Port`
- `Username`
- `Password`
- `Mailbox`
- `Insecure`

`RealDialer.Dial` refuses empty username or password. That confirms the hosted MCP path still depends on caller-provided secrets instead of stored user-bound credentials.

### Step 6: checked the current application DB bootstrap

I reviewed [db.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go) and Clay SQL settings in [settings.go](/home/manuel/code/wesen/corporate-headquarters/clay/pkg/sql/settings.go).

The useful conclusion was that the app-side storage direction is already established:

- Clay SQL opens either SQLite or Postgres-backed connections
- `smailnaild` already defaults to SQLite but is structurally compatible with Postgres
- the DB layer is therefore the right place for local users, external identities, and encrypted IMAP credentials

### Step 7: checked standards references

I reviewed:

- OpenID Connect Core
- OpenID Connect Discovery
- OAuth 2.0
- JWT
- Keycloak client administration docs

I used those references to keep the guide clear about what is standard versus provider-specific. The most important narrative constraint was this:

- `sub` is standard and central
- `client_id` is standard but identifies the OAuth client, not the human
- `tenant_id` is not a standard OIDC claim and must be treated as optional provider-specific context

### Step 8: design conclusion

The design conclusion is that hosted `smailnail` needs a three-layer identity model:

1. external proof from OIDC
2. normalized application principal
3. local application user plus stored IMAP accounts

The stable join between layer `1` and layer `3` should be `(issuer, subject)`, not email and not `client_id`.

### Step 9: validation and publication

I ran:

- `docmgr doctor --ticket SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE --stale-after 30`
- `remarquee upload bundle ... --name "SMAILNAIL-011 OIDC Identity and IMAP Credential Guide" --remote-dir "/ai/2026/03/16/SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE"`

The workspace validated cleanly, and the PDF bundle was uploaded successfully to reMarkable. I then verified the remote directory listing and the uploaded document name.

## Quick reference

### Current state

- OIDC token validation exists.
- Typed application identity propagation does not.
- Per-user stored IMAP credentials do not.

### Recommended state

- richer principal in `go-go-mcp`
- typed context propagation
- local `users` plus `external_identities` tables
- encrypted `imap_accounts` table
- hosted MCP execution bound to stored account resolution

## Related

- [01-intern-guide-to-oidc-identity-user-mapping-and-imap-credential-storage-in-smailnail.md](../design-doc/01-intern-guide-to-oidc-identity-user-mapping-and-imap-credential-storage-in-smailnail.md)
- [03-openai-keycloak-dcr-debug-guide.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/reference/03-openai-keycloak-dcr-debug-guide.md)
