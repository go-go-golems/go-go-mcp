---
Title: Deploy smailnail on Coolify with GitHub SSO, per-user IMAP config, and MCP OIDC
Ticket: SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN
Status: active
Topics:
    - smailnail
    - glazed
    - go
    - email
    - review
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
      Note: Repository being analyzed for hosted deployment and auth design
ExternalSources: []
Summary: Research ticket describing how to turn smailnail from a CLI-only IMAP toolkit into a Coolify-hosted web application and remote MCP server with GitHub-backed SSO, per-user IMAP credentials, and standards-compliant OIDC authorization.
LastUpdated: 2026-03-08T22:32:49.824107113-04:00
WhatFor: Use this ticket to understand the recommended deployment architecture, auth model, data model, API design, and implementation sequence for shipping smailnail as a hosted application and MCP service.
WhenToUse: Use when planning the hosted smailnail product, onboarding an intern to the work, or implementing Coolify deployment, GitHub login, per-user IMAP configuration, and MCP authorization.
---


# Deploy smailnail on Coolify with GitHub SSO, per-user IMAP config, and MCP OIDC

## Overview

This ticket maps how to evolve `smailnail` from a CLI-only IMAP toolkit into a hosted application on Coolify with GitHub-backed sign-in, per-user IMAP connection storage, and remote MCP access protected by OAuth 2.1 / OIDC.

The main conclusion is that `smailnail` does not currently have any of the hosted-app substrate required for this feature set. It has strong IMAP domain logic, but no web server, no database, no user/session model, and no MCP server. The recommended target design is therefore:

1. Add a new Go binary, tentatively `smailnaild`, that serves web UI, JSON API, and MCP.
2. Use Keycloak as the product OIDC provider.
3. Configure GitHub as a social login provider inside Keycloak.
4. Store each user’s IMAP settings in Postgres with app-managed encryption.
5. Expose remote MCP over `streamable_http` with protected-resource metadata and bearer-token validation.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary guide**: [design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md](./design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md)
- **Diary**: [reference/01-investigation-diary.md](./reference/01-investigation-diary.md)
- **Capability scan**: [scripts/current-capability-scan.sh](./scripts/current-capability-scan.sh)

## Status

Current status: **active**

Research deliverable status:

- current-state architecture mapped
- external deployment and auth constraints checked against current docs
- target architecture documented
- phased implementation plan documented
- ticket-local capability scan added and executed
- ticket validation and reMarkable upload pending

## Topics

- smailnail
- glazed
- go
- email
- review

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
