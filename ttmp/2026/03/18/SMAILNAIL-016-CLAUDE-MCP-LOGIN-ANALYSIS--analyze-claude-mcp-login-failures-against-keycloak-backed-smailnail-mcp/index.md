---
Title: Analyze Claude MCP login failures against Keycloak-backed smailnail MCP
Ticket: SMAILNAIL-016-CLAUDE-MCP-LOGIN-ANALYSIS
Status: complete
Topics:
    - authentication
    - keycloak
    - mcp
    - oidc
    - smailnail
    - claude
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnaild/commands/serve.go
      Note: Hosted smailnaild entrypoint that mounts the MCP handler
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: Core embeddable MCP HTTP auth and protected resource metadata handling
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md
      Note: Deployment contract for the merged hosted server
ExternalSources: []
Summary: Ticket for the Claude.ai login failure against the hosted smailnail MCP server, including the completed server-side challenge fix, the production redeploy, the Keycloak DCR root-cause evidence, and the intern-facing design guide.
LastUpdated: 2026-03-18T22:01:33.674539828-04:00
WhatFor: Track analysis and remediation planning for Claude's failure to complete OAuth against the hosted smailnail MCP deployment.
WhenToUse: Use when investigating Claude connector auth failures, explaining the hosted smailnail MCP auth stack, or continuing the Keycloak client registration remediation.
---


# Analyze Claude MCP login failures against Keycloak-backed smailnail MCP

## Overview

This ticket captures the production investigation into why Claude.ai failed to log into the hosted smailnail MCP server even though the server served correct protected resource metadata and a minimal bearer challenge. The investigation showed the failure occurred in Keycloak during dynamic client registration, not in the smailnail MCP transport itself, and the chosen remediation has now been applied and verified.

The ticket now contains:

- a detailed design and implementation guide for a new intern,
- a diary of the debugging and deployment work completed so far,
- concrete Keycloak and smailnail evidence collected from production.

## Key Links

- **Design doc**: `design-doc/01-claude-mcp-oauth-and-keycloak-dynamic-client-registration-guide-for-smailnail.md`
- **Diary**: `reference/01-diary.md`
- **Prior deployment reference**: `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/reference/02-recreate-and-verify-hosted-smailnail-mcp.md`
- **Related Files**: See frontmatter `RelatedFiles`

## Status

Current status: **complete**

## Topics

- authentication
- keycloak
- mcp
- oidc
- smailnail
- claude

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
