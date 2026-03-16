# Tasks

## Backend implementation

- [x] A1. Refactor the MCP server package so it can build mounted HTTP handlers instead of assuming a standalone server process.
- [x] A2. Define the merged config surface for `smailnaild serve`, including explicit MCP auth/resource settings alongside the existing web OIDC settings.
- [x] A3. Update `smailnaild` bootstrap to construct shared DB, encryption, identity, and account dependencies once and pass them into both the hosted API and the MCP layer.
- [x] A4. Mount `/.well-known/oauth-protected-resource` and `/mcp` into the hosted router while preserving the existing SPA and `/api/*` behavior.
- [x] A5. Add tests proving route separation: browser auth still protects `/api/*`, bearer auth still protects `/mcp`, and public metadata remains accessible.
- [x] A6. Update the MCP execution path to use the merged dependency graph rather than any standalone-only bootstrap.
- [x] A7. Decide the future of `cmd/smailnail-imap-mcp` and implement the chosen compatibility shape.

## Validation

- [x] B1. Update the local development workflow so the merged server can be started with the local Keycloak and Dovecot fixtures.
- [x] B2. Add or update integration tests that validate browser-session auth, stored account setup, and bearer-authenticated MCP requests against the same local database.
- [x] B3. Run manual smokes for the merged server:
  - [x] browser login
  - [x] `/api/me`
  - [x] account create and lightweight hosted IMAP access
  - [x] unauthenticated `/mcp` returns `401`
  - [x] authenticated MCP call works with stored `accountId`
  Note: the dedicated hosted `/api/accounts/{id}/test` endpoint returned a false-negative `use of closed network connection` against the remote Dovecot fixture, even though mailbox listing, message preview, and bearer-authenticated MCP access all succeeded against the same stored account.

## Deployment packaging

- [x] C1. Replace or generalize the existing MCP-only Docker/entrypoint setup so the merged server is the primary hosted image.
- [x] C2. Add clear environment-variable documentation for the merged hosted runtime.
- [x] C3. Produce a merged deployment guide that supersedes the split MCP-only guide.

## Coolify rollout

- [x] D1. Create the merged Coolify app or staging app.
- [x] D2. Configure the merged app with production DB, encryption, browser OIDC, and MCP OIDC settings.
- [x] D3. Verify hosted browser login against Keycloak.
- [x] D4. Verify hosted account creation and lightweight IMAP access against the hosted Dovecot fixture.
- [x] D5. Verify hosted MCP metadata and unauthenticated `401`.
- [x] D6. Verify hosted authenticated MCP execution using a stored account.
- [x] D7. Cut over from the split deployment to the merged deployment and record any compatibility notes.

## Documentation and handoff

- [x] E1. Keep a detailed implementation diary with debugging notes, exact commands, and validation outcomes.
- [x] E2. Update changelog entries as the work progresses.
- [x] E3. Write a final operator playbook for the merged hosted server.
- [x] E4. Document rollback and compatibility guidance for the old standalone MCP deployment.

## Granular sequencing

### Slice 1: code-structure refactor

- [x] S1. Extract reusable MCP mountable handler construction.
- [x] S2. Thread merged config into `smailnaild serve`.
- [x] S3. Mount the MCP routes into the hosted router.
- [x] S4. Keep the old standalone binary working via wrapper or adapter.

### Slice 2: local proof

- [x] S5. Run the merged server locally with Keycloak and Dovecot.
- [x] S6. Verify browser login and `/api/me`.
- [x] S7. Verify account creation and lightweight test.
- [x] S8. Verify bearer-authenticated MCP on the same server.

### Slice 3: deployment

- [x] S9. Build the merged image and update the runtime entrypoint.
- [x] S10. Deploy the merged app on Coolify.
- [x] S11. Run hosted browser and MCP smokes.
- [x] S12. Update docs and complete the ticket.
