# Tasks

## Backend implementation

- [x] A1. Refactor the MCP server package so it can build mounted HTTP handlers instead of assuming a standalone server process.
- [x] A2. Define the merged config surface for `smailnaild serve`, including explicit MCP auth/resource settings alongside the existing web OIDC settings.
- [x] A3. Update `smailnaild` bootstrap to construct shared DB, encryption, identity, and account dependencies once and pass them into both the hosted API and the MCP layer.
- [x] A4. Mount `/.well-known/oauth-protected-resource` and `/mcp` into the hosted router while preserving the existing SPA and `/api/*` behavior.
- [ ] A5. Add tests proving route separation: browser auth still protects `/api/*`, bearer auth still protects `/mcp`, and public metadata remains accessible.
- [x] A6. Update the MCP execution path to use the merged dependency graph rather than any standalone-only bootstrap.
- [ ] A7. Decide the future of `cmd/smailnail-imap-mcp` and implement the chosen compatibility shape.

## Local validation

- [ ] B1. Update the local development workflow so the merged server can be started with the local Keycloak and Dovecot fixtures.
- [ ] B2. Add or update integration tests that validate browser-session auth, stored account setup, and bearer-authenticated MCP requests against the same local database.
- [ ] B3. Run local manual smokes for the merged server:
  - [ ] browser login
  - [ ] `/api/me`
  - [ ] account create and lightweight test
  - [ ] unauthenticated `/mcp` returns `401`
  - [ ] authenticated MCP call works with stored `accountId`

## Deployment packaging

- [ ] C1. Replace or generalize the existing MCP-only Docker/entrypoint setup so the merged server is the primary hosted image.
- [ ] C2. Add clear environment-variable documentation for the merged hosted runtime.
- [ ] C3. Produce a merged deployment guide that supersedes the split MCP-only guide.

## Coolify rollout

- [ ] D1. Create the merged Coolify app or staging app.
- [ ] D2. Configure the merged app with production DB, encryption, browser OIDC, and MCP OIDC settings.
- [ ] D3. Verify hosted browser login against Keycloak.
- [ ] D4. Verify hosted account creation and lightweight IMAP test against the hosted Dovecot fixture.
- [ ] D5. Verify hosted MCP metadata and unauthenticated `401`.
- [ ] D6. Verify hosted authenticated MCP execution using a stored account.
- [ ] D7. Cut over from the split deployment to the merged deployment and record any compatibility notes.

## Documentation and handoff

- [ ] E1. Keep a detailed implementation diary with debugging notes, exact commands, and validation outcomes.
- [ ] E2. Update changelog entries as the work progresses.
- [ ] E3. Write a final operator playbook for the merged hosted server.
- [ ] E4. Document rollback and compatibility guidance for the old standalone MCP deployment.

## Granular sequencing

### Slice 1: code-structure refactor

- [x] S1. Extract reusable MCP mountable handler construction.
- [x] S2. Thread merged config into `smailnaild serve`.
- [x] S3. Mount the MCP routes into the hosted router.
- [x] S4. Keep the old standalone binary working via wrapper or adapter.

### Slice 2: local proof

- [ ] S5. Run the merged server locally with Keycloak and Dovecot.
- [ ] S6. Verify browser login and `/api/me`.
- [ ] S7. Verify account creation and lightweight test.
- [ ] S8. Verify bearer-authenticated MCP on the same server.

### Slice 3: deployment

- [ ] S9. Build the merged image and update the runtime entrypoint.
- [ ] S10. Deploy the merged app on Coolify.
- [ ] S11. Run hosted browser and MCP smokes.
- [ ] S12. Update docs and complete the ticket.
