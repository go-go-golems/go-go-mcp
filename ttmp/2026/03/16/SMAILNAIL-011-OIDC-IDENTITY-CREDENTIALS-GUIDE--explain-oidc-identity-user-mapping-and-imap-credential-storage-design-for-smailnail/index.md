---
Title: Explain OIDC identity, user mapping, and IMAP credential storage design for smailnail
Ticket: SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE
Status: complete
Topics:
    - smailnail
    - mcp
    - keycloak
    - oidc
    - authentication
    - sql
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external.go
      Note: External OIDC verification path analyzed in the guide
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: MCP auth middleware and protected-resource path analyzed in the guide
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go
      Note: Current tool execution path that still ignores authenticated identity
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go
      Note: Current raw-credential IMAP execution path discussed in the guide
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go
      Note: App-side storage bootstrap that should host local user and secret storage
ExternalSources:
    - https://openid.net/specs/openid-connect-core-1_0.html
    - https://openid.net/specs/openid-connect-discovery-1_0.html
    - https://datatracker.ietf.org/doc/html/rfc6749
    - https://datatracker.ietf.org/doc/html/rfc7519
    - https://www.keycloak.org/docs/latest/server_admin/#assembly-managing-clients_server_administration_guide
Summary: Ticket workspace for a detailed onboarding guide that explains hosted smailnail OIDC authentication, provider-neutral identity normalization, and app-side IMAP credential storage design.
LastUpdated: 2026-03-16T09:18:00-04:00
WhatFor: Organize the intern-facing analysis and implementation guidance for turning OIDC-authenticated MCP requests into local user and IMAP account resolution in hosted smailnail.
WhenToUse: Use when onboarding new engineers to the hosted auth architecture or planning the next implementation slice for user-aware IMAP access.
---

# Explain OIDC identity, user mapping, and IMAP credential storage design for smailnail

## Overview

This ticket contains an intern-facing explanation of the hosted `smailnail` OIDC architecture and the missing application layer between token validation and per-user IMAP access. The main output is a detailed guide that explains:

- how the current hosted OIDC flow works
- what claims such as `sub`, `client_id`, and `tenant_id` actually mean
- why the local user key should be `(issuer, subject)`
- why IMAP credentials belong in the app database rather than in Keycloak
- how to implement a provider-neutral identity model that can survive future IdP changes

## Key Links

- Main guide: [01-intern-guide-to-oidc-identity-user-mapping-and-imap-credential-storage-in-smailnail.md](./design-doc/01-intern-guide-to-oidc-identity-user-mapping-and-imap-credential-storage-in-smailnail.md)
- Diary: [01-investigation-diary.md](./reference/01-investigation-diary.md)
- Related files: see the frontmatter `RelatedFiles` field
- External references: see the frontmatter `ExternalSources` field

## Status

Current status: **complete**

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- `design-doc/` contains the detailed intern-facing guide
- `reference/` contains the investigation diary and supporting context
- `scripts/` is reserved for any future ticket-local scripts
