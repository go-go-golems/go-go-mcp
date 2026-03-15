---
Title: Diary
Ticket: SMAILNAIL-008-SMAILNAILD-SQL-BOOTSTRAP-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - glazed
    - sql
    - go
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Implementation diary for creating the first hosted smailnaild slice with Glazed command wiring, Clay SQL-backed DB bootstrap, and a minimal HTTP server.
LastUpdated: 2026-03-15T18:03:01.273047273-04:00
WhatFor: Use this diary to follow the implementation sequence, understand why the hosted slice was scoped this way, and review exact commands, failures, and commits.
WhenToUse: Use when continuing the smailnaild hosted-app bootstrap work or reviewing how the initial command and DB bootstrap were built.
---

# Diary

## Goal

Record the implementation steps for the first hosted `smailnaild` slice, including ticket setup, code changes, tests, commits, and any dead ends encountered while wiring Glazed and Clay SQL together.

## Step 1: Create the implementation ticket and lock the first hosted slice

The hosted-product design already existed, but it was too broad to execute cleanly in one pass. I created a dedicated implementation ticket for a smaller, buildable slice: introduce `smailnaild`, use Glazed for the CLI surface, use Clay SQL sections for database configuration, and stand up a minimal hosted HTTP process that proves the new boundary works.

This step is intentionally narrow. It avoids pretending we are implementing Keycloak or per-user IMAP settings before the hosted process, SQL bootstrap, and readiness model exist. That keeps the first code changes defensible and makes later tickets easier to stage.

### Prompt Context

**User prompt (verbatim):** "use glazed and also use clay's pkg/sql section to setup a SQL connection, that way we are ready for both sqlite and postgres: /home/manuel/code/wesen/corporate-headquarters/clay"

**Assistant interpretation:** Create a dedicated implementation ticket for a hosted `smailnaild` slice that explicitly uses Glazed and Clay SQL, then implement it incrementally with commits and a maintained diary.

**Inferred user intent:** Establish the hosted app foundation correctly now so later login, per-user IMAP settings, and MCP auth work can build on a reusable SQL-backed base instead of ad hoc configuration.

### What I did
- Created the ticket workspace `SMAILNAIL-008-SMAILNAILD-SQL-BOOTSTRAP-IMPLEMENTATION`.
- Added a design/implementation document and this diary.
- Wrote a granular task list oriented around a real first vertical slice rather than the full hosted-product roadmap.
- Reviewed Clay SQL helpers and existing `smailnail` command patterns before touching code.

### Why
- The broad hosted-product ticket (`SMAILNAIL-003`) is useful architecture context but too large to execute directly.
- Clay already provides the right SQL configuration abstraction, so the implementation ticket needed to anchor on that instead of inventing new flags.

### What worked
- `docmgr` ticket creation and document creation worked once the ticket workspace existed.
- The Clay SQL package inspection confirmed that it already normalizes SQLite and Postgres/pgx configuration and can open `sqlx.DB` connections.

### What didn't work
- Running `docmgr doc add --ticket SMAILNAIL-008-SMAILNAILD-SQL-BOOTSTRAP-IMPLEMENTATION --doc-type design-doc --title "Implementation plan for smailnaild hosted skeleton and Clay SQL bootstrap"` immediately after ticket creation failed once with:
  `Error: failed to find ticket directory: ticket not found: SMAILNAIL-008-SMAILNAILD-SQL-BOOTSTRAP-IMPLEMENTATION`
- Re-running the same command after the workspace was fully visible succeeded.

### What I learned
- The docs root is the shared `go-go-mcp/ttmp` workspace, so ticket documentation still lives outside the `smailnail` repo.
- `smailnail` already has Clay available transitively, but not yet as an explicit, intentional part of a hosted-app command surface.

### What was tricky to build
- The scope boundary was the main tricky part. The user's direction touched a large hosted-app roadmap, but the first defensible implementation slice is much smaller. The challenge was to make the ticket narrow enough to build end-to-end now while still aligning with the larger hosted-product design.

### What warrants a second pair of eyes
- Whether the chosen first slice is narrow enough without being too throwaway.
- Whether SQLite-first defaults are the right ergonomic choice for `smailnaild`.

### What should be done in the future
- Add later diary steps for code commits, tests, and follow-up decisions as each implementation task lands.

### Code review instructions
- Start with the implementation plan doc and tasks list in this ticket.
- Confirm the planned code changes stay limited to `smailnaild` command scaffolding, Clay SQL bootstrap, and minimal hosted HTTP handlers.

### Technical details
- Ticket path:
  `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/15/SMAILNAIL-008-SMAILNAILD-SQL-BOOTSTRAP-IMPLEMENTATION--implement-smailnaild-hosted-app-skeleton-with-clay-sql-bootstrap`
- Clay SQL reference files:
  - `/home/manuel/code/wesen/corporate-headquarters/clay/pkg/sql/settings.go`
  - `/home/manuel/code/wesen/corporate-headquarters/clay/pkg/sql/config.go`
  - `/home/manuel/code/wesen/corporate-headquarters/clay/pkg/sql/sources.go`
