# Tasks

## Analysis and design

- [x] Define the problem statement and desired end state for shared OIDC identity across browser and MCP flows.
- [x] Document the provider-neutral identity key and explain why it is `(issuer, subject)`.
- [x] Describe the target interaction between Keycloak, `smailnaild`, `smailnail-mcp`, and the shared application database.
- [x] Produce a detailed implementation plan for the next execution phases.

## Backend workstream

### Identity schema and repositories

- [ ] Add `users` table and repository.
- [ ] Add `user_external_identities` table with unique `(issuer, subject)` constraint.
- [ ] Add `web_sessions` table and repository.
- [ ] Add schema migration tests for identity and session tables.

### Shared user identity service

- [ ] Define provider-neutral principal types under `pkg/smailnaild` or a shared hosted auth package.
- [ ] Implement `ResolveOrProvisionUser(ctx, principal)` with idempotent upsert behavior.
- [ ] Persist selected claims as profile metadata without making them primary keys.
- [ ] Add unit tests for repeated login and repeated bearer-authenticated resolution.

### smailnaild web auth

- [ ] Add Glazed configuration section for web OIDC settings.
- [ ] Add OIDC client/bootstrap wiring in `cmd/smailnaild/commands/serve.go`.
- [ ] Add `/auth/login`.
- [ ] Add `/auth/callback`.
- [ ] Add `/auth/logout`.
- [ ] Add `/api/me`.
- [ ] Add session cookie issuance and validation middleware.
- [ ] Replace implicit `local-user` fallback with authenticated session resolution in hosted mode.
- [ ] Preserve an explicit development override mode only if intentionally enabled.
- [ ] Add integration tests for anonymous, login, me, and logout flows.

### MCP shared identity

- [ ] Extend the MCP auth boundary to carry a richer principal structure than only headers.
- [ ] Add a smailnail-side principal adapter from the verified OIDC token to the shared user identity service.
- [ ] Resolve or provision local users for bearer-authenticated MCP requests.
- [ ] Add integration tests proving browser login and MCP bearer auth resolve to the same local user.

### Account ownership and MCP account usage

- [ ] Add ownership checks to account loading paths.
- [ ] Extend MCP execution APIs so stored accounts can be selected by account ID instead of requiring raw credentials.
- [ ] Ensure the same local user can use browser-created accounts through MCP.
- [ ] Add end-to-end tests against local Keycloak and local Dovecot.

## Frontend workstream

- [ ] Add boot-time `/api/me` fetch and unauthenticated redirect handling.
- [ ] Add logged-out shell and login CTA.
- [ ] Add authenticated user display in the top-level app shell.
- [ ] Update account setup flows to assume authenticated ownership instead of dev fallback behavior.

## Deployment and operations

- [ ] Add Keycloak client setup notes for `smailnail-web`.
- [ ] Add environment examples for local dev and Coolify production.
- [ ] Add a test playbook for local Keycloak plus local Dovecot plus hosted UI.
- [ ] Add a test playbook for remote Keycloak plus hosted MCP.
