# Tasks

## Research Deliverables

- [x] Create the ticket workspace, design doc, diary, and tracked scripts directory
- [x] Inventory the current `smailnail` architecture and confirm the absence of web, auth, and persistence layers
- [x] Inspect reusable MCP/OIDC patterns in `go-go-mcp`
- [x] Research current Coolify, GitHub OAuth, Keycloak, and MCP authorization guidance
- [x] Write the detailed intern-facing analysis, design, and implementation guide
- [x] Record the investigation chronologically in the diary
- [x] Add and run a ticket-local capability scan script
- [x] Run `docmgr doctor` cleanly
- [x] Upload the bundle to reMarkable and verify the remote listing

## Recommended Implementation Backlog

- [ ] Add a new hosted binary such as `cmd/smailnaild`
- [ ] Introduce a database-backed persistence layer for users, sessions, and encrypted IMAP accounts
- [ ] Add browser login with Keycloak OIDC and GitHub as a social provider
- [ ] Add a basic post-login settings UI for IMAP connection management
- [ ] Add JSON API endpoints for IMAP connection CRUD and connection testing
- [ ] Add remote MCP endpoints using `streamable_http`
- [ ] Implement bearer-token validation, audience checks, and MCP protected-resource metadata
- [ ] Replace plaintext IMAP tool args in hosted flows with stored `connection_id` references
- [ ] Add Docker/Coolify deployment artifacts for the app, Postgres, and Keycloak
- [ ] Add integration tests for login, IMAP settings, and authenticated MCP calls
