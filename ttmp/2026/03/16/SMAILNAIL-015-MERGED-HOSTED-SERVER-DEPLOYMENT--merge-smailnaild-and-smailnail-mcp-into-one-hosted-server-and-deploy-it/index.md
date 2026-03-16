---
Title: Merge smailnaild and smailnail MCP into one hosted server and deploy it
Ticket: SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT
Status: complete
Topics:
    - smailnail
    - mcp
    - oidc
    - keycloak
    - coolify
    - deployments
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go
      Note: Current hosted router that should become the single HTTP surface for SPA, API, auth, and MCP
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go
      Note: Current standalone MCP server bootstrap that should be refactored into a reusable mounted handler
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnaild/commands/serve.go
      Note: Primary hosted command that should absorb MCP configuration and serve the merged deployment
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail-imap-mcp/main.go
      Note: Standalone MCP entrypoint that should become a thin compatibility wrapper or be retired
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnail-imap-mcp.sh
      Note: Existing MCP container entrypoint that should be replaced or generalized for the merged hosted server
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnail-imap-mcp-coolify.md
      Note: Existing production MCP deployment guide that must be superseded by a merged-host deployment guide
ExternalSources:
    - https://modelcontextprotocol.io/specification/2025-06-18/basic/transports
    - https://datatracker.ietf.org/doc/html/rfc9728
Summary: Merged the hosted web backend and the hosted MCP surface into one server process and one production deployment, validated browser-session and bearer-token auth paths on the same host, and then hardened the hosted account-test path against transient IMAP connection-closure failures.
LastUpdated: 2026-03-16T16:29:07-04:00
WhatFor: Organize the design, implementation, validation, and Coolify deployment work required to make smailnaild the single hosted server for the SPA, browser auth flow, application API, and MCP HTTP endpoint.
WhenToUse: Use when refactoring smailnail toward one hosted deployment surface and when preparing the next Coolify rollout that should replace the separate smailnail-imap-mcp app.
---

# Merge smailnaild and smailnail MCP into one hosted server and deploy it

## Overview

Today the hosted production shape is split:

- `smailnaild` owns the SPA, browser OIDC flow, and application API locally
- `smailnail-imap-mcp` owns the MCP HTTP surface in production
- both need the same Keycloak realm, the same local-user mapping model, the same app DB, and the same encryption keys

That split made early deployment easier, but it is now friction:

- separate container images and entrypoints
- duplicated environment wiring
- more chances for auth or DB config drift
- an unnecessarily fragmented product surface

The goal of this ticket is to converge on one hosted server process, served by one `http.Server`, where:

- `/` serves the SPA
- `/auth/*` handles browser login and logout
- `/api/*` serves the hosted account/rule API
- `/mcp` serves the MCP streamable HTTP transport
- `/.well-known/oauth-protected-resource` serves MCP auth metadata

The critical architectural constraint is that one server does not mean one auth mechanism. The merged server must continue to use:

- browser-session auth for the web app and API
- bearer-token OIDC auth for the MCP route

This ticket covers both the code refactor and the deployment cutover on Coolify.

## Key Links

- Design doc: [design/01-implementation-plan-for-merging-smailnaild-and-mcp-into-one-hosted-server.md](./design/01-implementation-plan-for-merging-smailnaild-and-mcp-into-one-hosted-server.md)
- Diary: [reference/01-implementation-diary.md](./reference/01-implementation-diary.md)
- Prior shared-identity ticket: [../SMAILNAIL-014-SHARED-OIDC-IMPLEMENTATION--implement-shared-oidc-identity-across-smailnaild-and-smailnail-mcp/index.md](../SMAILNAIL-014-SHARED-OIDC-IMPLEMENTATION--implement-shared-oidc-identity-across-smailnaild-and-smailnail-mcp/index.md)
- Prior hosted MCP deployment ticket: [../SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/index.md](../SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/index.md)
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **complete**

Residual note:

- the merged host is still using container-local SQLite, so redeploying the app clears saved hosted accounts until the app DB is moved to persistent storage

## Topics

- smailnail
- mcp
- oidc
- keycloak
- coolify
- deployment
- architecture

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
