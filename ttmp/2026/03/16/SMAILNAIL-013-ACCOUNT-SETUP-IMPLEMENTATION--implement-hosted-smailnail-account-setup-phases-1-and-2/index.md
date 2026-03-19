---
Title: Implement hosted smailnail account setup phases 1 and 2
Ticket: SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION
Status: active
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
      Note: Current hosted HTTP baseline to expand into account and rule APIs
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go
      Note: Current schema bootstrap entrypoint that the implementation plan begins by extending
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-012-WEB-UI-UX-RESEARCH--research-web-ui-for-smailnail-mcp-account-inbox-and-filter-management/design-doc/01-intern-guide-to-smailnail-web-ui-ux-architecture-and-implementation-for-hosted-mcp-mail-management.md
      Note: Source research ticket from which these implementation phases were derived
ExternalSources: []
Summary: Execution ticket for implementing the first two hosted smailnail UI phases: account setup, mailbox previews, rule CRUD, and rule dry-runs.
LastUpdated: 2026-03-16T10:00:00-04:00
WhatFor: Organize the concrete work required to make hosted smailnail usable for account setup and safe rule testing.
WhenToUse: Use when starting implementation of the hosted account-management and rule-preview product slice.
---

# Implement hosted smailnail account setup phases 1 and 2

## Overview

This ticket turns the earlier hosted UI research into an implementation plan and task list. It is focused on the first two execution phases needed to make hosted `smailnail` meaningfully usable:

- Phase 1: account setup, account testing, mailbox listing, and message previews
- Phase 2: rule CRUD and rule dry-runs

## Key Links

- Main implementation plan: [01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md](./design-doc/01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md)
- Diary: [01-implementation-diary.md](./reference/01-implementation-diary.md)
- API reference: [02-api-reference-for-hosted-smailnail-account-setup-and-rule-dry-run-backend.md](./reference/02-api-reference-for-hosted-smailnail-account-setup-and-rule-dry-run-backend.md)

## Status

Current status: **active**

## Tasks

See [tasks.md](./tasks.md) for the detailed milestone and execution breakdown.

## Changelog

See [changelog.md](./changelog.md) for ticket updates.
