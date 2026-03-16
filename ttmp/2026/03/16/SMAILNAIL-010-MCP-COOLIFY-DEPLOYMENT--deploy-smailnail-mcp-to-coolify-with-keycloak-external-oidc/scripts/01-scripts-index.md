---
Title: Ticket script index
Ticket: SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - coolify
    - keycloak
    - deployments
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Inventory of the ticket-scoped helper scripts used to create, configure, and validate the hosted smailnail MCP and Dovecot deployment.
LastUpdated: 2026-03-16T06:15:00-04:00
WhatFor: Provide a single place to discover the executable helpers captured under the ticket scripts directory.
WhenToUse: Use when replaying the deployment steps or reviewing which scripts correspond to which live operation.
---

# Ticket Scripts

These scripts capture the operator steps used during the hosted `smailnail-mcp` and hosted Dovecot rollout.

They are intended to be run from the workstation with SSH access to `root@89.167.52.236`.

## Scripts

- `create_coolify_mcp_app.sh`: create the public-repo Coolify app for `smailnail-imap-mcp`
- `set_coolify_mcp_envs.sh`: create the hosted MCP env vars through the Coolify API
- `dedupe_coolify_mcp_envs.sh`: remove duplicate env rows for the MCP app
- `create_keycloak_realm_and_mcp_client.sh`: create the `smailnail` realm and the Claude-facing `smailnail-mcp` client
- `create_keycloak_smoke_client.sh`: create a confidential service-account client for bearer smoke tests
- `create_coolify_dovecot_service.sh`: create the hosted raw-port Dovecot fixture service from the checked-in compose file
- `smoke_hosted_dovecot.sh`: validate mailbox create, append, and fetch against the hosted IMAPS endpoint
- `smoke_hosted_mcp_oidc.go`: proper streamable-HTTP MCP client smoke that authenticates with Keycloak and calls `executeIMAPJS`
- `smoke_hosted_mcp_oidc.sh`: wrapper that resolves the Keycloak smoke-client secret and runs the hosted MCP smoke

## Expected environment

- Coolify API token is stored on the server in `~/.apitoken`
- `coolify` CLI is installed on the server and configured with context `scapegoat`
- For Keycloak admin operations, export:
  - `KEYCLOAK_ADMIN_USER`
  - `KEYCLOAK_ADMIN_PASSWORD`
- For the hosted bearer-auth MCP smoke, either export:
  - `SMAILNAIL_MCP_SMOKE_CLIENT_SECRET`
  - or the Keycloak admin credentials above so the wrapper can fetch the secret automatically

## Notes

- These scripts intentionally target the current deployment UUIDs and hostnames from this ticket.
- The Dovecot fixture uses self-signed TLS; IMAPS validation uses `--insecure`.
