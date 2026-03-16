# Changelog

## 2026-03-16

- Initial workspace created
- Scoped the ticket around production packaging for `smailnail-imap-mcp`, Coolify deployment on `smailnail.mcp.scapegoat.dev`, Keycloak external OIDC via `auth.scapegoat.dev`, and a separate hosted Dovecot test target
- Added production packaging for `smailnail-imap-mcp`, including a Dockerfile, Coolify-oriented entrypoint, deployment documentation, and a validated local OIDC container smoke (`ab5df7b`)
- Added a Coolify-friendly root `Dockerfile`, pushed the public deployment branch, created the hosted application in Coolify, created the `smailnail` realm and `smailnail-mcp` client in Keycloak, and documented the resulting hosted deployment flow (`f24629d`)
- Fixed the production image so Coolify health checks can run inside the container, removed duplicate app env rows created during API setup, and verified the live hosted MCP on `https://smailnail.mcp.scapegoat.dev` returns protected-resource metadata plus unauthenticated `401 missing bearer` (`6072f7c`)
- Added an explicit recreation/verification reference doc and an explicit `/mcp` routing explanation so the hosted MCP slice can be re-run without reconstructing steps from the diary
- Added a hosted Dovecot fixture definition, created the raw-port Coolify service `gh32795yh1av2dpi2j6lhn6h`, and validated remote IMAPS mailbox creation, message append, and message fetch against `89.167.52.236:993` (`04f2762`)
- Consolidated the ad hoc deployment helpers into the ticket `scripts/` directory so the Coolify, Keycloak, and hosted Dovecot steps can be replayed without shell-history archaeology
- Added a proper authenticated streamable-HTTP smoke client under the ticket `scripts/` directory and validated a live Keycloak-backed `executeIMAPJS` call that connected to the hosted Dovecot fixture and returned `{"mailbox":"INBOX"}`
