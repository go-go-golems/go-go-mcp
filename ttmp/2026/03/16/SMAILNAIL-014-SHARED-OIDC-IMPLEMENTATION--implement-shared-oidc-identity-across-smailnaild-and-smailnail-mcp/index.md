---
Title: Implement shared OIDC identity across smailnaild and smailnail MCP
Ticket: SMAILNAIL-014-SHARED-OIDC-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - mcp
    - oidc
    - authentication
    - keycloak
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/user.go
      Note: Current hosted app user identity stub that must be replaced with real session-backed identity resolution
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnaild/commands/serve.go
      Note: Current hosted app bootstrap where OIDC settings and middleware wiring will be introduced
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external.go
      Note: Existing MCP OIDC token verification path that should feed the shared local-user mapping layer
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: Current MCP auth middleware path that carries only subject and client metadata forward
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go
      Note: Current MCP execution path that still ignores authenticated identity and stored IMAP accounts
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go
      Note: Existing application schema bootstrap that should grow user identity and session storage
ExternalSources:
    - https://openid.net/specs/openid-connect-core-1_0.html
    - https://openid.net/specs/openid-connect-discovery-1_0.html
    - https://datatracker.ietf.org/doc/html/rfc7591
    - https://www.keycloak.org/docs/latest/server_admin/
Summary: Implementation ticket for making the hosted web app and hosted MCP use the same OIDC-backed user identity model, so browser login and bearer-authenticated MCP requests resolve to the same local user and IMAP account storage.
LastUpdated: 2026-03-16T11:28:46.369552646-04:00
WhatFor: Organize the concrete design and execution work required to replace the current dev-only user stub with a real shared OIDC identity layer across smailnaild and smailnail MCP.
WhenToUse: Use when implementing browser login, local user provisioning, shared account ownership, and MCP-side user-aware IMAP resolution.
---

# Implement shared OIDC identity across smailnaild and smailnail MCP

## Overview

This ticket covers the missing identity layer between the current hosted web UI and the already OIDC-protected hosted MCP. Today:

- `smailnail-mcp` can validate bearer tokens from Keycloak
- `smailnaild` still treats every browser request as `local-user`
- neither side resolves identity into a shared local user record
- stored IMAP accounts are not yet tied to real OIDC-authenticated users

The goal of this ticket is to establish one provider-neutral application identity model based on `(issuer, subject)`, then use it consistently in both places:

- the browser login/session flow for `smailnaild`
- the bearer-token flow for `smailnail-mcp`
- the database rows that own IMAP accounts, tests, rules, and future mailbox state

## Key Links

- Main design document: [design-doc/01-implementation-plan-for-shared-oidc-identity-across-smailnaild-and-smailnail-mcp.md](./design-doc/01-implementation-plan-for-shared-oidc-identity-across-smailnaild-and-smailnail-mcp.md)
- Diary: [reference/01-implementation-diary.md](./reference/01-implementation-diary.md)
- Related account setup ticket: [../SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION--implement-hosted-smailnail-account-setup-phases-1-and-2/index.md](../SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION--implement-hosted-smailnail-account-setup-phases-1-and-2/index.md)
- Prior background guide: [../SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE--explain-oidc-identity-user-mapping-and-imap-credential-storage-design-for-smailnail/index.md](../SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE--explain-oidc-identity-user-mapping-and-imap-credential-storage-design-for-smailnail/index.md)
- Related Files: see frontmatter
- External Sources: see frontmatter

## Status

Current status: **active**

## Topics

- smailnail
- mcp
- oidc
- authentication
- keycloak
- architecture

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- `design-doc/` contains the detailed implementation plan
- `reference/` contains the diary and any follow-up API or auth references
- `scripts/` is reserved for ticket-local helper scripts if needed later
