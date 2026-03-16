# Changelog

## 2026-03-16

- Initial workspace created
# Changelog

- 2026-03-16: Created ticket `SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT`, added the merged-server implementation plan, the granular task breakdown, and the kickoff diary.
- 2026-03-16: Completed the first merge slice by extracting mountable MCP HTTP handlers in `go-go-mcp`, wiring hosted MCP settings into `smailnaild serve`, and mounting `/mcp` plus the protected-resource metadata route into the hosted app handler.
- 2026-03-16: Added route-separation coverage proving `/api/*` still uses browser-session auth, `/mcp` still uses bearer auth, and the protected-resource metadata endpoint remains public.
- 2026-03-16: Repackaged `smailnaild` as the primary hosted image with the embedded web UI, a merged Docker entrypoint, and deployment documentation that supersedes the standalone MCP image.
- 2026-03-16: Verified the standalone `smailnail-imap-mcp` binary still works as a compatibility wrapper after the merge.
- 2026-03-16: Reused the existing production host `smailnail.mcp.scapegoat.dev` for the merged deployment, updated the Coolify app config, and rolled the app forward to the merged image.
- 2026-03-16: Created the production Keycloak web client `smailnail-web`, provisioned browser-login redirect URIs, and added the production test user `alice`.
- 2026-03-16: Completed hosted browser-session validation for `/auth/login`, `/auth/callback`, `/api/me`, saved-account creation, mailbox listing, message preview, unauthenticated `/mcp` protection, and bearer-authenticated MCP execution using the same stored account.
- 2026-03-16: Added reusable hosted smoke scripts under `scripts/` for browser-login validation and merged `/mcp` bearer-auth validation.
- 2026-03-16: Hardened the hosted IMAP account-test path to retry transient network-closure failures, added unit coverage for retry vs non-retry behavior, redeployed the merged Coolify app, and revalidated hosted `/api/accounts/{id}/test` plus merged `/mcp` against a freshly recreated account.
- 2026-03-16: Added an operator playbook covering post-deploy hosted validation and the next GitHub SSO setup sequence for the production Keycloak realm.
