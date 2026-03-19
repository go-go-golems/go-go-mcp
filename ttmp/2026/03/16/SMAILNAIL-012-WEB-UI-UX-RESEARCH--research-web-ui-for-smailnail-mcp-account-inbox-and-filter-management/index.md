---
Title: Research web UI for smailnail MCP account, inbox, and filter management
Ticket: SMAILNAIL-012-WEB-UI-UX-RESEARCH
Status: complete
Topics:
    - smailnail
    - mcp
    - documentation
    - architecture
    - authentication
    - sql
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go
      Note: Current hosted skeleton that this research proposes to extend into a real product UI
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go
      Note: Current account connection primitive that drives the proposed onboarding and test UX
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go
      Note: Current rule schema that should become editable through the hosted UI
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go
      Note: Current fetch engine that can power mailbox previews and rule dry-runs
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go
      Note: Current MCP surface that the hosted UI should configure and explain
ExternalSources:
    - https://support.google.com/accounts/answer/185833
    - https://support.google.com/mail/answer/7126229
    - https://learn.microsoft.com/en-us/exchange/clients-and-mobile-in-exchange-online/deprecation-of-basic-authentication-exchange-online
    - https://support.apple.com/en-us/102578
    - https://support.apple.com/en-il/guide/mail/mlhlp1040/mac
    - https://support.apple.com/en-il/guide/mail/mlhlp1190/mac
    - https://datatracker.ietf.org/doc/html/rfc3501
Summary: Ticket workspace for researching and designing a hosted smailnail web UI that manages multiple IMAP accounts, mailbox previews, rule creation, and MCP-facing account policy.
LastUpdated: 2026-03-16T09:48:00-04:00
WhatFor: Organize the intern-facing UX and implementation research for turning smailnail into a hosted mail operations console rather than a collection of CLI and MCP entrypoints.
WhenToUse: Use when onboarding engineers to the planned hosted UI, deciding MVP scope, or mapping user needs to backend APIs and screen designs.
---

# Research web UI for smailnail MCP account, inbox, and filter management

## Overview

This ticket contains an intern-facing design and implementation guide for a hosted `smailnail` web UI. The guide starts from user needs, derives concrete product features, and then maps those features to:

- data model changes
- backend API shape
- service/package boundaries
- implementation phases
- ASCII screen mockups

The goal is to make `smailnail` usable as a hosted product for configuring multiple IMAP accounts, previewing inboxes, building and testing filters, and controlling what the MCP layer can access.

## Key Links

- Main guide: [01-intern-guide-to-smailnail-web-ui-ux-architecture-and-implementation-for-hosted-mcp-mail-management.md](./design-doc/01-intern-guide-to-smailnail-web-ui-ux-architecture-and-implementation-for-hosted-mcp-mail-management.md)
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

- `design-doc/` contains the main intern-facing research and design guide
- `reference/` contains the investigation diary
- `scripts/` is reserved for any future ticket-local scripts
