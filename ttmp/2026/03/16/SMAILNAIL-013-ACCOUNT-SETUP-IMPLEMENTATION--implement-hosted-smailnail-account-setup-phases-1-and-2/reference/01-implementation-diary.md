---
Title: Implementation diary
Ticket: SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - mcp
    - documentation
    - architecture
    - authentication
    - sql
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-012-WEB-UI-UX-RESEARCH--research-web-ui-for-smailnail-mcp-account-inbox-and-filter-management/design-doc/01-intern-guide-to-smailnail-web-ui-ux-architecture-and-implementation-for-hosted-mcp-mail-management.md
      Note: Source research ticket from which the implementation phases were derived
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go
      Note: Reviewed to capture the current hosted baseline before expanding to account and rule APIs
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go
      Note: Reviewed to identify the current schema bootstrap gap
ExternalSources: []
Summary: Chronological notes for turning the hosted smailnail UI research into a concrete Phase 1 and Phase 2 implementation ticket.
LastUpdated: 2026-03-16T10:00:00-04:00
WhatFor: Preserve the reasoning behind the task ordering and the chosen implementation slices.
WhenToUse: Use when starting implementation or revisiting why the work was broken down in this order.
---

# Implementation diary

## Goal

Create a concrete execution ticket for the first two hosted `smailnail` UI phases so the team can start building account setup, mailbox previews, rule CRUD, and rule dry-runs without having to reinterpret the earlier research document.

## 2026-03-16

### Step 1: created the execution ticket

I created `SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION` as a follow-on to the UX research ticket. The intent is to separate discovery from implementation planning.

### Step 2: mapped the research phases to code

I revisited the `SMAILNAIL-012` guide and extracted:

- Phase 1: hosted backend primitives
- Phase 2: rule CRUD and dry-run

I also checked the current code in:

- [http.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go)
- [db.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go)

That made the initial implementation gap explicit:

- the hosted server exists
- the app DB exists
- but there are no domain packages, no CRUD APIs, no rule endpoints, and no UI

### Step 3: chose vertical slices instead of layer-first tasks

I structured the plan around milestones that can be committed independently:

- schema
- secrets
- accounts
- testing and preview
- APIs
- frontend shell
- rules
- dry-runs

That ordering should make it possible to keep the app runnable and reviewable throughout the work.

## Quick reference

### First delivery target

- account CRUD
- account test
- mailbox list
- message preview

### Second delivery target

- saved rules
- rule validation
- dry-run previews

### Key dependency

- app-side encrypted secret storage from the earlier identity design ticket

## Related

- [01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md](../design-doc/01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md)

