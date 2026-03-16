# Tasks

## Analysis and design

- [x] Define the problem statement and desired end state for shared OIDC identity across browser and MCP flows.
- [x] Document the provider-neutral identity key and explain why it is `(issuer, subject)`.
- [x] Describe the target interaction between Keycloak, `smailnaild`, `smailnail-mcp`, and the shared application database.
- [x] Produce a detailed implementation plan for the next execution phases.

## Backend workstream

### Identity schema and repositories

- [x] `A1` Bump `smailnaild` schema version and add `users` table.
- [x] `A2` Add `user_external_identities` table with unique `(issuer, subject)` constraint.
- [x] `A3` Add `web_sessions` table with expiry metadata.
- [x] `A4` Add `pkg/smailnaild/identity/types.go` for user, external identity, session, and principal models.
- [x] `A5` Add `pkg/smailnaild/identity/repository.go` with CRUD helpers for users, external identities, and sessions.
- [x] `A6` Add schema migration tests covering fresh bootstrap and migration from schema version `5`.

### Shared user identity service

- [x] `B1` Define the provider-neutral principal type used by both web auth and MCP auth.
- [x] `B2` Implement `ResolveOrProvisionUser(ctx, principal)` with idempotent upsert behavior.
- [x] `B3` Persist selected profile claims on both `users` and `user_external_identities` without using them as identity keys.
- [x] `B4` Add unit tests for first login, repeated login, and profile refresh behavior.
- [x] `B5` Add unit tests showing the same principal resolves to the same local user across repeated calls.

### smailnaild web auth

- [x] `C1` Add a Glazed OIDC settings section for issuer, client ID, client secret, redirect URL, scopes, and session cookie settings.
- [x] `C2` Add OIDC bootstrap wiring in `cmd/smailnaild/commands/serve.go`.
- [ ] `C3` Add `/auth/login` to start the authorization code flow.
- [ ] `C4` Add `/auth/callback` to exchange code for tokens and create a local session.
- [ ] `C5` Add `/auth/logout` to clear the session and optionally redirect through provider logout.
- [x] `C6` Add `/api/me` for the authenticated frontend bootstrap call.
- [x] `C7` Add session cookie issuance, lookup, refresh, and deletion middleware.
- [x] `C8` Replace implicit `local-user` fallback with authenticated session resolution in hosted mode.
- [x] `C9` Preserve an explicit development override mode only when a dedicated flag or env toggle is enabled.
- [ ] `C10` Add integration tests for anonymous access, login callback, me, logout, and expired session behavior.

### MCP shared identity

- [ ] `D1` Extend the `go-go-mcp` auth boundary to carry a richer verified principal than forwarded headers alone.
- [ ] `D2` Add a smailnail-side principal adapter from verified OIDC token claims to the shared identity service.
- [ ] `D3` Resolve or provision local users for bearer-authenticated MCP requests.
- [ ] `D4` Thread the resolved local user through MCP tool execution context.
- [ ] `D5` Add integration tests proving browser login and MCP bearer auth resolve to the same local user.

### Account ownership and MCP account usage

- [ ] `E1` Route account loading through the resolved local `user_id` rather than the dev fallback.
- [ ] `E2` Extend MCP execution APIs so stored accounts can be selected by account ID instead of requiring raw credentials.
- [ ] `E3` Add authorization checks for cross-user account access attempts.
- [ ] `E4` Ensure the same local user can use browser-created accounts through MCP.
- [ ] `E5` Add end-to-end tests against local Keycloak and local Dovecot.

## Frontend workstream

- [ ] `F1` Add boot-time `/api/me` fetch and unauthenticated redirect handling.
- [ ] `F2` Add logged-out shell and login CTA.
- [ ] `F3` Add authenticated user display in the top-level app shell.
- [ ] `F4` Update account setup flows to assume authenticated ownership instead of dev fallback behavior.

## Deployment and operations

- [ ] `G1` Add Keycloak client setup notes for `smailnail-web`.
- [ ] `G2` Add environment examples for local dev and Coolify production.
- [ ] `G3` Add a test playbook for local Keycloak plus local Dovecot plus hosted UI.
- [ ] `G4` Add a test playbook for remote Keycloak plus hosted MCP.
