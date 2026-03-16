---
Title: Implementation diary
Ticket: SMAILNAIL-014-SHARED-OIDC-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - mcp
    - oidc
    - authentication
    - keycloak
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/user.go
      Note: Starting point for the current dev-only identity behavior that motivated this ticket
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external.go
      Note: Existing verified-token path reviewed while scoping the shared identity work
ExternalSources:
    - https://openid.net/specs/openid-connect-core-1_0.html
    - https://openid.net/specs/openid-connect-discovery-1_0.html
    - https://datatracker.ietf.org/doc/html/rfc7591
    - https://www.keycloak.org/docs/latest/server_admin/
Summary: Diary capturing the creation of the shared OIDC identity implementation ticket and the reasoning that connects the current browser stub to the existing MCP OIDC flow.
LastUpdated: 2026-03-16T13:29:00-04:00
WhatFor: Record the concrete reasoning, references, and scoping decisions behind the shared identity implementation ticket.
WhenToUse: Use when reviewing why the ticket exists, what assumptions it makes, and which code paths were inspected at kickoff.
---

# Implementation diary

## Goal

Capture the starting context for the shared OIDC identity implementation ticket so future implementation work can see what was already established.

## Context

Existing state at ticket creation:

- `smailnaild` still resolves users through `HeaderUserResolver` and falls back to `local-user`
- `smailnail-mcp` already validates OIDC access tokens through `go-go-mcp`
- account storage and rule storage now exist in the hosted SQL backend
- the product needs the same human to be recognized identically in the browser and in MCP calls

## Quick Reference

### Notes from kickoff

- Browser auth and MCP auth should share identity semantics, not necessarily the same transport mechanism.
- The stable local key should be `(issuer, subject)`.
- The web app should use server-side OIDC code exchange and session cookies.
- The MCP should keep bearer-token validation and then call the same local-user mapping layer.
- IMAP accounts should stay in the application database and belong to the resolved local user.

### References consulted

- OIDC Core for claim semantics, especially `iss`, `sub`, and audience rules
- OIDC Discovery for issuer metadata conventions
- RFC 7591 because hosted MCP connectors may rely on dynamic client registration
- Keycloak server admin docs for client scopes, mappers, and client setup

## Usage Examples

Use this diary when:

- starting implementation on the backend auth layer
- onboarding a contributor to the shared identity problem
- justifying why email or `client_id` should not become the primary local user key

## Related

- Main plan: [../design-doc/01-implementation-plan-for-shared-oidc-identity-across-smailnaild-and-smailnail-mcp.md](../design-doc/01-implementation-plan-for-shared-oidc-identity-across-smailnaild-and-smailnail-mcp.md)
- Background guide: [../../SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE--explain-oidc-identity-user-mapping-and-imap-credential-storage-design-for-smailnail/index.md](../../SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE--explain-oidc-identity-user-mapping-and-imap-credential-storage-design-for-smailnail/index.md)
- Execution dependency: [../../SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION--implement-hosted-smailnail-account-setup-phases-1-and-2/index.md](../../SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION--implement-hosted-smailnail-account-setup-phases-1-and-2/index.md)

## History Analysis Addendum

To support wider repo-history searches, a reusable exporter was added under:

- [../scripts/export_git_history_to_sqlite.py](../scripts/export_git_history_to_sqlite.py)

That script exports full reachable git history into SQLite so future questions about timeline, path introduction, and file-level evolution can be answered with SQL instead of repeated ad hoc `git log` commands.

It was exercised against the `smailnail` repo and confirmed:

- the repo is not shallow
- the oldest reachable commit is `27c2460` (`Initial commit`)
- the first reachable MCP files appear in `ff584b4` (`Add smailnail IMAP JS MCP runtime slice`)
- there is no earlier non-JS MCP surface in the reachable history before that commit

## Implementation Step 1: Identity schema and shared user service foundation

Goal of this step:

- create the first executable slice of `SMAILNAIL-014`
- add local identity storage and provider-neutral principal resolution
- avoid touching the still-unrelated frontend embed drift in the `smailnail` worktree

What was changed in `smailnail`:

- bumped the hosted schema from version `5` to `6`
- added three new tables in `pkg/smailnaild/db.go`
  - `users`
  - `user_external_identities`
  - `web_sessions`
- added a new package:
  - `pkg/smailnaild/identity/types.go`
  - `pkg/smailnaild/identity/repository.go`
  - `pkg/smailnaild/identity/service.go`
  - `pkg/smailnaild/identity/service_test.go`
- extended `pkg/smailnaild/db_test.go` to assert the new schema version and tables

What the new slice does:

- defines a provider-neutral `ExternalPrincipal`
- defines local `User`, `ExternalIdentity`, and `WebSession` records
- provides repository helpers for users, external identities, and sessions
- implements `ResolveOrProvisionUser(ctx, principal)`
- refreshes stored profile fields on repeated resolution without changing the canonical identity key

Validation commands run:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/smailnaild/...
go test ./cmd/smailnaild/...
```

Observed issue and fix:

- first test run failed because `pkg/smailnaild/identity/repository.go` used `time.Time` in a helper but did not import `time`
- fixed by adding the missing import and rerunning the focused test suites

Worktree notes:

- unrelated existing drift was left untouched:
  - `go.mod`
  - `pkg/smailnaild/web/embed/public/...`
  - `smailnail-imap-mcp`
  - `smailnaild.sqlite`
  - `ui/tsconfig.tsbuildinfo`

## Implementation Step 2: Hosted auth mode config, session resolver, and `/api/me`

Goal of this step:

- stop treating every hosted request as implicitly authenticated
- add an explicit hosted auth mode surface
- make session-backed user resolution possible before the full OIDC callback/login flow exists

What was changed in `smailnail`:

- added `pkg/smailnaild/auth/config.go`
  - new Glazed `auth` section
  - `auth-mode` choices: `dev`, `session`, `oidc`
  - session cookie and future OIDC fields
- updated `pkg/smailnaild/user.go`
  - `UserResolver` now returns `(string, error)`
  - added `ErrUnauthenticated`
  - added `SessionUserResolver`
- updated `cmd/smailnaild/commands/serve.go`
  - loads auth settings
  - uses `HeaderUserResolver` only in `auth-mode=dev`
  - uses `SessionUserResolver` in `session` and `oidc` modes
- updated `pkg/smailnaild/http.go`
  - protected APIs now call `requireUserID(...)`
  - unauthenticated access returns `401`
  - added `GET /api/me`
- extended tests in `pkg/smailnaild/http_test.go`
  - unauthenticated `401` case
  - cookie-backed `/api/me` success case

Validation commands run:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/smailnaild/...
go test ./cmd/smailnaild/...
```

Observed issue and fix:

- moving `identity` tests away from the parent `smailnaild` bootstrap caused a package import cycle to disappear but also removed the transitive SQLite driver import
- fixed by:
  - self-bootstrapping the minimal identity tables inside `identity/service_test.go`
  - adding the explicit `_ "github.com/mattn/go-sqlite3"` import to that test file

Scope boundary for this step:

- this step does **not** yet add `/auth/login`, `/auth/callback`, or `/auth/logout`
- it lays the auth boundary and config surface needed for those routes in the next slice

## Implementation Step 3: OIDC login, callback, logout, and web-session round-trip

Goal of this step:

- make `auth-mode=oidc` actually usable for browser login
- connect OIDC discovery and JWT verification to the shared identity service
- prove the web flow with an executable fake-provider test instead of relying on manual login only

What was changed in `smailnail`:

- completed `pkg/smailnaild/auth/oidc.go`
  - OIDC discovery fetch
  - JWKS cache and `id_token` signature verification
  - authorization code exchange
  - local user provisioning through `identity.Service`
  - hosted session creation and cookie issuance
- updated `pkg/smailnaild/http.go`
  - added optional web auth handler wiring
  - registers:
    - `GET /auth/login`
    - `GET /auth/callback`
    - `GET /auth/logout`
- updated `cmd/smailnaild/commands/serve.go`
  - creates `identity.Service`
  - boots `OIDCAuthenticator` when `auth-mode=oidc`
  - passes the web auth handler into the hosted HTTP server
- added `pkg/smailnaild/oidc_test.go`
  - fake OIDC provider with discovery, JWKS, token exchange, and signed `id_token`
  - asserts login redirect
  - asserts callback session issuance
  - asserts `/api/me` returns the provisioned user
  - asserts logout invalidates the session
- extended `pkg/smailnaild/http_test.go`
  - added expired-session rejection coverage for `/api/me`

Validation commands run:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/smailnaild/... ./cmd/smailnaild/...
go test ./...
```

Observed issue and fix:

- the initial `oidc.go` draft had a stray `github.com/golang-jwt/jwt/v5/request` import kept alive by a dummy reference
- removed it before wiring the authenticator so the new slice stayed on one JWT stack

Behavior added by this step:

- in `auth-mode=oidc`, `/auth/login` now redirects to the provider authorization endpoint with state and nonce cookies
- `/auth/callback` now exchanges the code, verifies the `id_token`, provisions or refreshes the local user, and creates a `web_sessions` row
- `/auth/logout` now deletes the local session and clears the browser cookie
- authenticated browser requests can now bootstrap through `/api/me` using the same local identity tables that future MCP bearer auth will use

Scope boundary for this step:

- provider logout redirect is still local-only; there is not yet an upstream end-session flow
- MCP bearer-authenticated user mapping is still the next implementation slice

## Implementation Step 4: Carry verified MCP principals into the same local-user mapping path

Goal of this step:

- stop treating MCP bearer auth as a separate identity world
- carry a richer verified OIDC principal through `go-go-mcp`
- resolve the same local user for browser sessions and MCP bearer requests

What was changed in `go-go-mcp`:

- extended `pkg/embeddable/auth_provider.go`
  - `AuthPrincipal` now carries profile fields and raw claims, not just `subject` and `client_id`
- added `pkg/embeddable/auth_context.go`
  - `WithAuthPrincipal(...)`
  - `GetAuthPrincipal(...)`
- updated `pkg/embeddable/mcpgo_backend.go`
  - auth middleware now stores the verified principal in request context before handing off to the MCP HTTP transport
- updated `pkg/embeddable/auth_provider_external.go`
  - external OIDC validation now maps `email`, `email_verified`, `preferred_username`, `name`, and `picture` into the principal
- extended tests:
  - `pkg/embeddable/auth_test.go`
  - `pkg/embeddable/auth_provider_external_test.go`

What was changed in `smailnail`:

- added MCP-side resolved-user context support:
  - `pkg/mcp/imapjs/identity_context.go`
- added MCP startup/runtime wiring:
  - `pkg/mcp/imapjs/identity_middleware.go`
  - new `--app-db-driver`
  - new `--app-db-dsn`
- updated `pkg/mcp/imapjs/server.go`
  - boots a shared-identity runtime
  - initializes the shared application database on server start
  - applies MCP middleware that:
    - reads the verified principal from context
    - resolves or provisions the local user via `identity.Service`
    - stores the resolved local identity back into tool execution context
- added tests:
  - `pkg/mcp/imapjs/identity_middleware_test.go`
  - `pkg/mcp/imapjs/web_identity_integration_test.go`

Validation commands run:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp
go test ./pkg/embeddable/...

cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/mcp/imapjs ./cmd/smailnail-imap-mcp
go test ./pkg/mcp/imapjs
```

Key result from this step:

- a browser-style OIDC login flow and an MCP bearer-authenticated request now resolve to the same local `users.id` when they share the same `(issuer, subject)`

Scope boundary for this step:

- the resolved local user is now available to tool handlers, but stored IMAP account selection by account ID is still the next slice
- the JavaScript execution tool still accepts only raw code; it has not yet been taught to consume browser-created stored accounts

## Implementation Step 5: Use browser-owned stored IMAP accounts from MCP JavaScript execution

Goal of this step:

- make the shared identity work useful in practice by letting MCP execution use hosted stored accounts
- keep one JavaScript API that works in both local and hosted modes
- enforce account ownership through the existing hosted account repository instead of duplicating credential checks

What was changed in `smailnail`:

- extended `pkg/services/smailnailjs/service.go`
  - `ConnectOptions` now accepts `accountId`
  - added `StoredAccountResolver`
  - `Service.Connect(...)` now resolves stored account credentials when `accountId` is provided
- updated `pkg/js/modules/smailnail/module.go`
  - module instances now carry an execution context
  - JavaScript `connect(...)` now uses the per-call context instead of `context.Background()`
- added `pkg/mcp/imapjs/service_context.go`
  - stores a per-request stored-account resolver and optional test dialer in context
- extended `pkg/mcp/imapjs/identity_middleware.go`
  - shared runtime now accepts app encryption flags
  - initializes `accounts.Service` when the shared app database and encryption key are available
  - injects a stored-account resolver bound to the resolved local `user_id`
- updated `pkg/mcp/imapjs/execute_tool.go`
  - runtime service construction now consumes resolver/dialer values from context
- added tests:
  - `pkg/services/smailnailjs/service_test.go`
  - `pkg/js/modules/smailnail/module_test.go`
  - `pkg/mcp/imapjs/execute_tool_account_test.go`
  - extended `pkg/mcp/imapjs/web_identity_integration_test.go`

Validation commands run:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/services/smailnailjs ./pkg/js/modules/smailnail ./pkg/mcp/imapjs
```

Key result from this step:

- hosted browser-created IMAP accounts can now be selected from MCP JavaScript using:

```javascript
const smailnail = require("smailnail");
const svc = smailnail.newService();
const session = svc.connect({ accountId: "acc-123" });
```

Security behavior now covered:

- account lookup is scoped to the resolved local `user_id`
- cross-user account IDs fail instead of leaking credentials
- an account must be explicitly marked `mcpEnabled` before MCP can use it

Scope boundary for this step:

- the full end-to-end test against local Keycloak and local Dovecot is still pending
- documentation for the new `accountId` connect path should be expanded in a later polish pass so the docs tool reflects it more clearly
