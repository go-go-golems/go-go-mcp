# Changelog

## 2026-03-09

- Initial workspace created


## 2026-03-09

Created the external-issuer research ticket, captured the current embedded OIDC/auth capability surface, and wrote the architecture guide for Keycloak production mode plus embedded dev login mode.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/auth/oidc/server.go — Current embedded issuer baseline for the refactor
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go — Current auth middleware coupling that the design recommends splitting
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/current-capability-scan.sh — Ticket-local scan script used to capture the current auth/admin surface

## 2026-03-09

Converted the research ticket into an implementation tracker and landed the first refactor slice: explicit auth modes in `pkg/embeddable`, a provider contract for HTTP auth, and an embedded-dev provider that keeps the existing OIDC server behavior while removing the backend's direct dependency on `*oidc.Server`.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/server.go — New `AuthMode` / `AuthOptions` model and legacy `WithOIDC` compatibility mapping
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider.go — Shared HTTP auth provider contract and embedded-dev provider implementation
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go — Backend migration from direct embedded-server coupling to provider-based middleware
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_test.go — Regression tests for `WWW-Authenticate`, principal propagation, static auth key handling, and protected-resource metadata

## 2026-03-09

Implemented external OIDC support for embeddable MCP HTTP backends, including discovery/JWKS/JWT validation, separate resource URL handling, explicit external-versus-embedded CLI flags, updated repo docs, and a ticket-local smoke script that covers external-provider tests plus embedded runtime auth behavior.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external.go — External OIDC discovery, JWKS cache, JWT verification, and scope/audience enforcement
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external_test.go — In-process external OIDC provider tests with local discovery and JWKS endpoints
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/command.go — Explicit `auth-mode`, external issuer, audience, scope, and embedded-dev flags
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/doc/topics/07-embedded-oidc.md — Updated user-facing auth documentation for embedded and external modes
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/auth-mode-smoke.sh — Ticket-local smoke harness used during the implementation closeout

## 2026-03-09

Added a ticket-local Docker Compose Keycloak playbook, an imported test realm with both basic and strict service-account clients, and a fully validated local smoke harness that proves `go-go-mcp` works against a real external Keycloak issuer in both lenient and audience-plus-scope-enforced modes.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/playbooks/01-local-keycloak-external-oidc-testing-playbook.md — Human-readable local setup and verification instructions
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/keycloak-local/docker-compose.yml — Ticket-local Keycloak dev stack
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/keycloak-local/import/mcp-local-realm.json — Imported realm with `mcp-cli-basic` and `mcp-cli-strict`
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/local-keycloak-smoke.sh — End-to-end local Keycloak smoke harness validated in this workspace
