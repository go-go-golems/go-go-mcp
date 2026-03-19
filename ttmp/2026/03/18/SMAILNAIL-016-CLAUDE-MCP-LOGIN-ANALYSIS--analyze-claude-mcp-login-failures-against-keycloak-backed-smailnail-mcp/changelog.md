# Changelog

## 2026-03-18

- Initial workspace created
- Recorded the production server-side fixes and redeploy that simplified the bearer challenge and removed `/readyz` debug log noise
- Confirmed from live smailnail logs that Claude still fails after the server-side fix
- Confirmed from live Keycloak logs that Claude reaches dynamic client registration and is rejected because requested scope `service_account` is not trusted by the anonymous registration policy
- Added a detailed design and implementation guide plus a full investigation diary for intern onboarding
- Applied the chosen Keycloak-side remediation by expanding the anonymous DCR `Allowed Client Scopes` policy to include `service_account`, validated DCR directly with a `201` registration response, and confirmed from the user that Claude now works end to end

## 2026-03-18

Documented the production auth challenge fix, the redeploy, and the Claude login failure sequence from live smailnail logs

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider.go — Bearer challenge fix deployed to production
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go — /readyz log suppression deployed to production


## 2026-03-18

Confirmed from Keycloak logs and client registration policy inspection that Claude reaches dynamic client registration and is rejected because anonymous registration does not trust the requested service_account scope

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/18/SMAILNAIL-016-CLAUDE-MCP-LOGIN-ANALYSIS--analyze-claude-mcp-login-failures-against-keycloak-backed-smailnail-mcp/design-doc/01-claude-mcp-oauth-and-keycloak-dynamic-client-registration-guide-for-smailnail.md — Records the production Keycloak evidence and design options
