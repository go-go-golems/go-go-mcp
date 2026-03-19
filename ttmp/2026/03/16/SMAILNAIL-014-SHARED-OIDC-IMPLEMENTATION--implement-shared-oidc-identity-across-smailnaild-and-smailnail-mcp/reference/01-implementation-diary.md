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

## Implementation Step 6: Prove the full local Keycloak + Dovecot + stored-account path

Goal of this step:

- stop relying only on fake providers and unit-level account resolution tests
- prove that one real local access token can drive the stored-account MCP path end to end
- capture the exact fixture assumptions so this flow can be rerun later without rediscovery

What was changed in `smailnail`:

- added `pkg/mcp/imapjs/local_stack_test.go`
  - gated by `SMAILNAIL_LOCAL_STACK_TEST=1`
  - talks to the real local Keycloak container on `http://127.0.0.1:18080`
  - ensures the `smailnail-mcp` client permits direct access grants for local automation
  - ensures the local user exists and has a usable password
  - fetches a real access token from Keycloak
  - provisions the same local user in a temp app DB
  - creates a stored IMAP account pointing at the local Dovecot fixture
  - runs `executeIMAPJS` with `accountId`
  - verifies the JavaScript service connects to real Dovecot and returns the mailbox name

Validation commands run:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/mcp/imapjs
SMAILNAIL_LOCAL_STACK_TEST=1 go test ./pkg/mcp/imapjs -run TestExecuteIMAPJSAgainstLocalKeycloakAndDovecot -v
go run ./cmd/smailnail fetch-mail --server 127.0.0.1 --port 993 --username a --password pass --mailbox INBOX --insecure --output json
```

Observed issues and fixes:

- the first run failed because the local Keycloak database contained stale state and the bootstrap admin user no longer existed
- fixed by resetting the local Keycloak Postgres dev volume and reimporting the dev realm
- the next run failed with `Account is not fully set up`
- fixed by making the test user updater explicitly clear `requiredActions`, enable the user, and mark email as verified
- the next run failed with `use of closed network connection` while opening `INBOX`
- reran the direct CLI IMAP path against the same local Dovecot fixture to confirm the fixture itself was healthy
- after the local Keycloak reset and rerun, the full end-to-end test passed consistently

Key result from this step:

- the shared identity design now has one live local proof path from Keycloak bearer token to stored account lookup to real IMAP connectivity

## Implementation Step 7: Document the operator path for shared OIDC in local and hosted environments

Goal of this step:

- close the remaining deployment and operations tasks in the ticket
- leave behind exact Keycloak, `smailnaild`, and MCP configuration notes instead of only test code
- make the containerized MCP deployment match the newly documented shared app DB requirements

What was changed in `smailnail`:

- added `docs/deployments/smailnaild-oidc-keycloak.md`
  - documents the current `smailnail-web` client shape
  - explains why the current implementation wants a confidential client
  - lists the exact OIDC/auth Glazed flags used by `smailnaild`
- added `docs/shared-oidc-playbook.md`
  - local stack startup
  - browser login
  - stored account creation
  - MCP startup against the same shared app DB
  - token fetch
  - stored-account MCP invocation
  - one-command local regression test
- updated `docs/deployments/smailnail-imap-mcp-coolify.md`
  - documents the shared app DB and encryption env vars needed for hosted stored-account resolution
- updated `scripts/docker-entrypoint.smailnail-imap-mcp.sh`
  - now forwards:
    - `SMAILNAIL_MCP_APP_DB_DRIVER`
    - `SMAILNAIL_MCP_APP_DB_DSN`
    - `SMAILNAIL_MCP_APP_ENCRYPTION_KEY_ID`
    - `SMAILNAIL_MCP_APP_ENCRYPTION_KEY_BASE64`
- updated `README.md`
  - links the shared-identity docs
  - documents the new local stack regression command

Validation commands run:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
sh -n scripts/docker-entrypoint.smailnail-imap-mcp.sh
go test ./pkg/mcp/imapjs
SMAILNAIL_LOCAL_STACK_TEST=1 go test ./pkg/mcp/imapjs -run TestExecuteIMAPJSAgainstLocalKeycloakAndDovecot -v
```

Key result from this step:

- the shared OIDC implementation is now backed by both executable coverage and a concrete operator playbook for local and hosted environments

## Implementation Step 8: Make the frontend respect authenticated ownership

Goal of this step:

- stop rendering the hosted account and rules UI as if anonymous users were already valid application users
- make the browser boot sequence explicitly depend on `/api/me`
- give the UI a minimal logged-out shell and a visible authenticated-user shell without redesigning the account workflows

What was changed in `smailnail`:

- extended `ui/src/api/client.ts`
  - adds `getCurrentUser()`
  - always includes credentials on API requests
- extended `ui/src/api/types.ts`
  - adds the `CurrentUser` type for `/api/me`
- added a new auth feature:
  - `ui/src/features/auth/authSlice.ts`
  - `ui/src/features/auth/LoggedOutShell.tsx`
  - `ui/src/features/auth/index.ts`
- updated `ui/src/store/index.ts`
  - registers the new `auth` reducer
- updated `ui/src/App.tsx`
  - boot-time `/api/me` fetch
  - loading shell while auth bootstrap is pending
  - logged-out shell when `/api/me` returns `401`
  - retry path for failed auth bootstrap
  - authenticated header badge showing the local user profile
  - logout link to `/auth/logout`
- updated `ui/src/styles/theme.css`
  - adds lightweight styling for the new auth shell

Behavior added by this step:

- the UI no longer tries to load account and rule screens before the current user is known
- anonymous users now see a focused login CTA instead of a misleading account-management shell
- authenticated users now see who the browser session resolved to before they start adding IMAP accounts
- existing account setup flows remain intact once auth bootstrap succeeds

Validation commands run:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui
pnpm run check
pnpm run build
```

Scope boundary for this step:

- this is still a server-side login model; the React app does not hold OIDC tokens directly
- the UI does not yet auto-redirect to `/auth/login`; it shows a logged-out shell with an explicit CTA instead

## Implementation Step 9: Debug the first real browser OIDC loop and fix the Vite auth boundary

This step was the first time the new shared-identity work was exercised like a real user would use it: browser on the Vite dev server, `smailnaild` on a separate port, local Keycloak as the OIDC provider, and the hosted app trying to bootstrap through `/api/me`. That immediately exposed the gap between “the backend supports OIDC” and “the frontend dev environment actually lets a browser complete the OIDC loop.”

The main lesson from this step was that the browser path had two separate failure classes that looked similar at first glance. One was expected authentication behavior, where `/api/me` correctly returned `401` before login. The other was infrastructure drift in the Vite dev proxy and backend redirect configuration, where `/auth/login` was not routed through Vite and `smailnaild` was sending Keycloak a stale callback URL.

### Prompt Context

**User prompt (verbatim):** "I get 401 on /api/me but no further info. Is keycloak running ?"

**Assistant interpretation:** The user was trying to validate the newly added browser-auth flow and was unsure whether the `401` meant “not logged in yet” or “the local Keycloak stack is broken.”

**Inferred user intent:** Confirm that the local OIDC system is alive end to end and make the browser login path work from the Vite development server instead of only from the backend port.

**Commit (code):** `310b137` — `chore(ui): proxy auth routes in vite dev server`

### What I did
- Verified the local Docker stack with:
  - `docker compose -f smailnail/docker-compose.local.yml ps`
  - `curl -i http://127.0.0.1:18080/realms/smailnail-dev/.well-known/openid-configuration`
- Confirmed Keycloak was healthy and that the initial `401` on `/api/me` was expected before login.
- Read the Vite config and confirmed only `/api` was proxied, not `/auth`.
- Read the live `smailnaild` process command line and confirmed the backend was running on port `8081` but still sending Keycloak `redirect_uri=http://127.0.0.1:8080/auth/callback`.
- Used `curl -D - http://127.0.0.1:8081/auth/login` and parsed the `Location` header to inspect the exact authorization URL generated by `smailnaild`.
- Patched `ui/vite.config.ts` so `/auth` now proxies to the backend just like `/api`.

### Why
- Without `/auth` proxying, `http://127.0.0.1:5050/auth/login` only exercised the Vite dev server, not the hosted backend.
- Without inspecting the actual authorization redirect, it was easy to blame Keycloak for what was really a stale `--oidc-redirect-url` on the backend process.
- The debugging target was not just “can login work somehow,” but “can a normal browser flow work from the same origin the frontend developer is using.”

### What worked
- The local stack itself was healthy.
- OIDC discovery from Keycloak responded correctly.
- Direct backend login endpoints on `8081` worked well enough to expose the real redirect mismatch.
- Inspecting the live backend process arguments made the redirect problem obvious immediately.
- Adding `/auth` proxy support closed the Vite-side gap cleanly.

### What didn't work
- Before the proxy fix, `http://127.0.0.1:5050/auth/login` did not work because Vite had no `/auth` proxy rule.
- The running backend process was started with:

```bash
go run ./cmd/smailnaild serve \
  --listen-port 8081 \
  ... \
  --oidc-redirect-url http://127.0.0.1:8080/auth/callback
```

- The resulting live redirect proved the mismatch:

```text
Location: http://127.0.0.1:18080/realms/smailnail-dev/protocol/openid-connect/auth?...&redirect_uri=http%3A%2F%2F127.0.0.1%3A8080%2Fauth%2Fcallback&...
```

- That failed against the local Keycloak client, which only allowed `8081` redirects at that point.

### What I learned
- A `401` on `/api/me` is a correct pre-login result and should not be treated as evidence that Keycloak is down.
- For browser-based OIDC, the dev server origin matters just as much as the backend origin; missing proxy coverage makes the entire auth loop look broken.
- When debugging redirect issues, the fastest signal is the actual `Location` header emitted by `/auth/login`, not the docs and not memory.

### What was tricky to build
- The failure looked like one problem from the browser, but it was really two stacked problems:
  - a missing `/auth` proxy in Vite
  - a stale backend redirect URL
- The backend had already been restarted multiple times during the session, so “what I think I launched” diverged from “what the running process is actually doing.”
- The symptom surface was noisy:
  - `/api/me` returned `401`
  - direct login sometimes reached Keycloak
  - login from Vite failed differently
- The reliable way through it was to check the running process and the exact redirect header rather than infer behavior from symptoms.

### What warrants a second pair of eyes
- The frontend still uses an explicit logged-out shell and CTA rather than auto-redirecting immediately into `/auth/login`.
- The dev flow now works, but the exact host/port consistency rules are still sharp enough that a new contributor could reproduce the same confusion if they mix `localhost` and `127.0.0.1`.

### What should be done in the future
- Add a tiny frontend debug panel or dev hint showing:
  - backend proxy target
  - current auth bootstrap status
  - whether `/api/me` failed because of `401` versus network failure

### Code review instructions
- Start in:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui/vite.config.ts`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/auth/oidc.go`
- Validate with:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui
pnpm run build
```

- Then manually:
  - start `smailnaild` on `8081`
  - start Vite on `5050`
  - visit `http://127.0.0.1:5050/auth/login`

### Technical details
- Diagnostic commands used:

```bash
curl -i -s http://127.0.0.1:18080/realms/smailnail-dev/.well-known/openid-configuration | sed -n '1,20p'
curl -s -o /tmp/smailnail-login.headers -D - http://127.0.0.1:8081/auth/login
ps -ef | rg 'smailnaild serve|cmd/smailnaild'
```

- Live finding that drove the fix:
  - backend port was `8081`
  - callback URI was still `127.0.0.1:8080/auth/callback`

## Implementation Step 10: Reset and stabilize local Keycloak defaults for repeatable testing

Once the redirect mismatch was understood, the next problem was local state drift inside Keycloak itself. The local realm import was no longer sufficient to guarantee a predictable test environment because the persisted Postgres volume contained changed client state, an unstable web-client secret, and user credentials that no longer matched the documented local defaults.

This step was about turning the local Keycloak stack back into an intentionally reproducible development fixture. The goal was not just to get one login working, but to make future resets recreate the same working client, user, and redirect setup without manual admin-console repair.

### Prompt Context

**User prompt (verbatim):** "i can't log in as admin / user it seems, maybe i changed the login , can we reset it?"

**Assistant interpretation:** The local Keycloak dev realm had drifted enough that the documented login credentials were no longer trustworthy, and the user wanted a clean reset instead of more guesswork.

**Inferred user intent:** Restore a deterministic local auth environment so browser login and MCP testing both have known credentials and known client settings.

**Commit (code):** `ae0f011` — `chore(smailnail): stabilize local keycloak defaults`

### What I did
- Inspected the local compose stack and volumes:
  - `smailnail_keycloak-postgres-data`
  - `smailnail_dovecot-home`
  - `smailnail_dovecot-ssl`
- Stopped and removed only the local Keycloak service containers.
- Removed the local Keycloak Postgres volume and restarted `keycloak-postgres` and `keycloak`.
- Waited for a fresh realm import and confirmed logs showed:
  - `Realm 'smailnail-dev' imported`
  - `Created temporary admin user with username admin`
- Used the admin API to normalize the local state immediately after reset:
  - `smailnail-web` client secret set to `smailnail-web-secret`
  - redirect URIs expanded to include `5050`, `8080`, and `8081` for both `localhost` and `127.0.0.1`
  - web origins expanded to match
  - `alice` test user created or updated with password `secret`
- Patched the checked-in local realm import to match those normalized dev defaults.
- Updated the README so the expected local auth credentials are explicit after a fresh reset.

### Why
- The previous local state was no longer aligned with the checked-in realm import or the docs.
- Fixing the live Keycloak instance only by API would have left the repo in a misleading state.
- The right outcome was: after deleting the local Keycloak DB volume, the imported realm should already be usable without manual admin surgery.

### What worked
- Resetting only the Keycloak Postgres dev volume left Dovecot intact and was enough to restore the admin bootstrap flow.
- The reimported realm came up cleanly.
- After normalization, the local defaults became predictable again:
  - `admin/admin`
  - `alice/secret`
  - `smailnail-web-secret`

### What didn't work
- The first attempt to remove the Postgres volume failed with:

```text
Error response from daemon: remove smailnail_keycloak-postgres-data: volume is in use
```

- That happened because stopped containers still held the volume attachment.
- The fix was to explicitly remove the stopped `keycloak` and `keycloak-postgres` containers before removing the volume.

### What I learned
- The local Keycloak fixture is only reproducible if the imported realm carries the actual dev defaults, not just “something close.”
- Client redirect URI coverage for browser testing should include both:
  - Vite origin (`5050`)
  - direct backend origins (`8080`, `8081`)
- Even after a successful reset, the checked-in import matters more than the live temporary admin changes because future debugging starts from the repo state.

### What was tricky to build
- This step mixed ephemeral container state and checked-in configuration, which creates a trap: it is easy to fix the current live stack and forget to encode the fix in the realm import.
- The debug loop included both:
  - “make the running environment usable now”
  - “make the next reset automatically reproduce this”
- Doing only the first would have made the current session work while guaranteeing the same confusion later.

### What warrants a second pair of eyes
- The local realm import now includes a fixed confidential client secret and a seeded local user. That is acceptable for a local dev stack, but it should stay clearly dev-only.
- If the team wants stricter local OIDC semantics later, this import may need a dedicated “browser-only dev realm” versus “connector test realm” split.

### What should be done in the future
- Consider a tiny `make reset-local-keycloak` target that:
  - removes the Keycloak dev volume
  - restarts the services
  - waits for health

### Code review instructions
- Start in:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/README.md`
- Validate with:

```bash
docker compose -f /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml stop keycloak keycloak-postgres
docker compose -f /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml rm -f -s keycloak keycloak-postgres
docker volume rm smailnail_keycloak-postgres-data
docker compose -f /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml up -d keycloak-postgres keycloak
```

- Then verify:
  - Keycloak admin login with `admin/admin`
  - browser user login with `alice/secret`

### Technical details
- Reset sequence that worked:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker compose -f docker-compose.local.yml rm -f -s keycloak keycloak-postgres
docker volume rm smailnail_keycloak-postgres-data
docker compose -f docker-compose.local.yml up -d keycloak-postgres keycloak
```

- Key live log lines after reset:

```text
Realm 'smailnail-dev' imported
Created temporary admin user with username admin
```

## Implementation Step 11: Add debug auth logging and fix the duplicate-account edit regression

After the environment was stable enough to log in, the next failures moved up the stack into application behavior. One was an auth-debugging problem: when the browser came back from Keycloak, the callback sometimes failed with `invalid oauth state`, but the backend emitted almost no request-level detail. The other was a pure UI regression: once an account had been created, using the result-screen “Edit” path created a second new account instead of updating the existing one.

These were two different bugs, but they were discovered in one continuous manual validation session and both blocked confidence in the hosted onboarding flow. The first needed observability. The second needed a state-machine fix in the account slice.

### Prompt Context

**User prompt (verbatim):** "can you add more logging across the board on smailnaild"

**Assistant interpretation:** The user wanted the hosted backend to emit enough structured debug information to diagnose the OIDC callback failure instead of treating it as a black box.

**Inferred user intent:** Make browser-auth failures explain themselves and remove the account-edit bug discovered during that same manual login/account-setup test.

**Commit (code):** `37a8691` — `feat(smailnaild): add debug auth logging`; `496f055` — `fix(smailnail-ui): keep new account in edit mode`

### What I did
- Added debug request logging around all hosted HTTP traffic in `pkg/smailnaild/http.go`.
- Added debug logs in `pkg/smailnaild/user.go` for:
  - dev fallback resolution
  - missing session cookie
  - empty session cookie
  - session lookup failure
  - expired session
  - successful session resolution
- Added detailed OIDC logs in `pkg/smailnaild/auth/oidc.go` for:
  - discovery fetch
  - login redirect construction
  - callback parameter presence
  - state-cookie mismatch
  - nonce-cookie missing
  - token exchange success/failure
  - ID token verification
  - local user resolution
  - hosted session creation
  - logout cleanup
- Reproduced the account edit behavior and inspected the account slice state transitions.
- Fixed `ui/src/features/accounts/accountsSlice.ts` so a successful create now sets `editingAccountId = account.id`, which keeps the form in update mode after the first create.

### Why
- Without debug logs, the auth callback failure only produced “invalid oauth state,” which is too little information to tell whether the cookie was missing, mismatched, or stale.
- Without fixing the slice, the manual validation of the hosted account flow would continue to create duplicate accounts whenever the user used “Edit” immediately after creating a first account.

### What worked
- The logging changes compiled cleanly and passed the hosted backend tests.
- The new logs now expose the exact callback mismatch fields:
  - `query_state`
  - `cookie_state`
  - `cookie_name`
- The account-slice fix is tiny but sufficient; it keeps the just-created account in edit mode and prevents the duplicate create on the next submit.
- The UI typecheck and build both passed after the slice fix.

### What didn't work
- Before the logging patch, debugging the callback meant guessing from browser behavior instead of reading the backend’s understanding of the state cookie.
- Before the slice fix, the result-screen edit path behaved like:
  - create account
  - show result
  - click edit
  - submit
  - create a second account

- The root cause was that `editingAccountId` remained `null` after a successful create, so the form submission still called `createAccountAndTest` instead of `updateAccountAndTest`.

### What I learned
- Browser-auth debugging without structured logs is too slow once cookies and cross-origin proxying get involved.
- The most useful auth logs are not generic request logs alone; they need auth-specific fields such as state, nonce, session ID, and resolved user ID.
- On the UI side, result-screen transitions are part of the form state machine. If the slice does not explicitly enter edit mode after create, the UI can look correct while still dispatching the wrong thunk.

### What was tricky to build
- The logging change touched three layers:
  - top-level request handling
  - user/session resolution
  - OIDC callback internals
- That breadth was necessary because the failure could have been:
  - request never reaching auth routes
  - session cookie not being sent
  - state cookie mismatch
  - token exchange failure
  - session write failure
- The duplicate-account fix was easy once found, but the symptom appeared only after a successful create-and-test cycle, which made it look like backend behavior at first glance.

### What warrants a second pair of eyes
- Request-level debug logging is intentionally verbose and may want redaction review later if additional sensitive fields are ever added.
- The auth logs currently include state and nonce values at debug level, which is acceptable for local debugging but should be reconsidered if the same logging style is used in shared environments.

### What should be done in the future
- Consider adding a focused frontend test around the account result-screen edit path so this regression does not come back.
- Consider a small auth-debug mode document that shows exactly which `smailnaild` log lines matter during `/auth/login` and `/auth/callback`.

### Code review instructions
- Start in:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/user.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/auth/oidc.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui/src/features/accounts/accountsSlice.ts`
- Validate with:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/smailnaild/... ./cmd/smailnaild/...

cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui
pnpm run check
pnpm run build
```

- Manual validation:
  - run `smailnaild` with `--log-level debug`
  - complete one login flow
  - create one account
  - use result-screen `Edit`
  - resubmit and verify the same account is updated

### Technical details
- Representative auth debug signals now emitted:
  - `Starting hosted OIDC login flow`
  - `OIDC callback state validation failed`
  - `OIDC token exchange succeeded`
  - `OIDC callback created hosted session`
  - `Hosted request resolved authenticated user`

- Validation commands used during this step:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
gofmt -w pkg/smailnaild/user.go pkg/smailnaild/http.go pkg/smailnaild/auth/oidc.go
go test ./pkg/smailnaild/... ./cmd/smailnaild/...

cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui
pnpm run check
pnpm run build
```
