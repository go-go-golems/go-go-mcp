# Tasks

## Implementation Backlog

- [ ] Write the concrete deployment plan and initial diary for the `smailnail-mcp` hosted slice
- [ ] Inspect the current MCP runtime, Coolify host state, and Keycloak baseline to confirm the exact production command and hostnames
- [ ] Add production packaging for `smailnail-imap-mcp` in the `smailnail` repo
- [ ] Improve the MCP binary defaults and docs so the deployment command is short, stable, and reviewable
- [ ] Validate the packaged container locally
- [ ] Add Coolify deployment documentation, env var reference, and verification steps for `smailnail.mcp.scapegoat.dev`
- [ ] Configure or document the Keycloak realm/client settings required for remote MCP OIDC
- [ ] Deploy `smailnail-mcp` to the Hetzner/Coolify host and verify the protected-resource metadata plus unauthenticated `401` behavior
- [ ] Set up a separate hosted Dovecot test target on the Coolify machine and document how to use it for remote testing
- [ ] Update the diary, changelog, related files, and task state as each implementation step lands
