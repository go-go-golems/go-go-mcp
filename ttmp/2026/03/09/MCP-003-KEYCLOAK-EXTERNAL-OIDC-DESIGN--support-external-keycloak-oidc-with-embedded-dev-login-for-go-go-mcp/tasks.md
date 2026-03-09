# Tasks

- [x] Create the research ticket and primary design/diary documents.
- [x] Inspect the current embedded OIDC server, HTTP auth middleware, CLI flags, and admin commands in `go-go-mcp`.
- [x] Capture the current capability surface in a ticket-local scan script.
- [x] Review prior `SMAILNAIL-003` design assumptions about external issuers and MCP auth.
- [x] Check current official guidance for MCP authorization, Keycloak identity brokering, and Coolify Keycloak deployment.
- [x] Analyze architectural options for supporting an external Keycloak issuer without losing a simple local password login flow.
- [x] Write the detailed architecture / implementation guide for a new intern.
- [x] Write a chronological investigation diary with commands, failures, and design rationale.
- [x] Relate the key implementation files and update the ticket changelog.
- [x] Run `docmgr doctor` for ticket hygiene.
- [x] Upload the finished ticket bundle to reMarkable and verify the remote listing.

## Implementation

- [x] Turn the ticket into an execution tracker with implementation tasks and diary entries.
- [x] Introduce an auth-mode / auth-options model that can represent `embedded_dev` and `external_oidc`.
- [x] Add an auth provider abstraction for HTTP MCP auth and protected-resource metadata.
- [x] Migrate the current embedded issuer flow behind the new provider abstraction without changing behavior.
- [x] Implement external OIDC discovery and Keycloak-compatible JWT validation.
- [x] Support separate resource URL / audience configuration for external issuer mode.
- [x] Refactor CLI flags so external and embedded settings are explicit, while keeping legacy `WithOIDC` usable.
- [x] Add focused tests for provider selection, protected-resource metadata, `WWW-Authenticate`, embedded dev auth, and external JWT validation.
- [x] Add or update smoke coverage for both auth modes where practical.
- [x] Update docs and the ticket diary/changelog as each slice lands.
- [x] Run validation and upload the updated implementation ticket bundle to reMarkable.
- [x] Add and validate a ticket-local Docker Compose Keycloak playbook for local external OIDC testing.
