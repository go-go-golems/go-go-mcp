# Tasks

## Implementation Backlog

- [x] Write the concrete deployment plan and initial diary for the `smailnail-mcp` hosted slice
- [x] Inspect the current MCP runtime, Coolify host state, and Keycloak baseline to confirm the exact production command and hostnames
- [x] Add production packaging for `smailnail-imap-mcp` in the `smailnail` repo
- [x] Improve the MCP binary defaults and docs so the deployment command is short, stable, and reviewable
- [x] Validate the packaged container locally
- [x] Add Coolify deployment documentation, env var reference, and verification steps for `smailnail.mcp.scapegoat.dev`
- [x] Make the MCP packaging compatible with Coolify public-repo builds using a standard Dockerfile path
- [x] Push the deployment branch so Coolify can build the current MCP deployment shape from the public repository
- [x] Configure and document the Keycloak realm/client settings required for remote MCP OIDC
- [x] Deploy `smailnail-mcp` to the Hetzner/Coolify host and verify the protected-resource metadata plus unauthenticated `401` behavior
- [x] Set up a separate hosted Dovecot test target on the Coolify machine and document how to use it for remote testing
- [ ] Update the diary, changelog, related files, and task state as each implementation step lands
