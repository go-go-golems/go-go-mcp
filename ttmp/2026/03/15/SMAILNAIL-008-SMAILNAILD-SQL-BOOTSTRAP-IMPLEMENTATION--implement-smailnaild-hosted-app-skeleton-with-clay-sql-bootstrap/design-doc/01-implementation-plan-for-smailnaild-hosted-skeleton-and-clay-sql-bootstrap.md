---
Title: Implementation plan for smailnaild hosted skeleton and Clay SQL bootstrap
Ticket: SMAILNAIL-008-SMAILNAILD-SQL-BOOTSTRAP-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - glazed
    - sql
    - go
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/clay/pkg/sql/config.go
      Note: Clay SQL config and connection logic to build on
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/clay/pkg/sql/settings.go
      Note: Clay SQL section wiring to reuse in serve command
    - Path: smailnail/cmd/smailnail/commands/fetch_mail.go
      Note: Reference command pattern for Glazed command wiring in smailnail
ExternalSources: []
Summary: Concrete implementation plan for introducing a hosted smailnaild binary that uses Glazed for its CLI surface and Clay SQL sections for application database connectivity, with a minimal HTTP server proving the hosted process boundary.
LastUpdated: 2026-03-15T18:03:11.892125737-04:00
WhatFor: Use this document to guide the first implementation slice of hosted smailnail infrastructure and to review whether the code establishes the right foundation for later auth and per-user IMAP work.
WhenToUse: Use when implementing or reviewing the initial smailnaild root/serve commands, DB bootstrap, and hosted HTTP skeleton.
---


# Implementation plan for smailnaild hosted skeleton and Clay SQL bootstrap

## Executive Summary

The right first implementation slice is not Keycloak, sessions, or per-user IMAP storage. It is a new hosted process boundary: `smailnaild`.

This slice should:

- add a dedicated hosted binary
- use Glazed/Cobra for a stable command surface
- reuse Clay SQL sections so the application DB can be configured consistently for SQLite or Postgres
- default to a local SQLite file for developer convenience
- start a small HTTP server with health and readiness endpoints

This keeps the initial scope small while locking in the one architectural decision that matters most right now: app state should live behind a SQL-backed repository boundary rather than being hard-wired to SQLite-only flags or ad hoc DSN parsing.

## Problem Statement

`smailnail` currently ships as CLI tools and an MCP binary, but it has no hosted application binary and no shared application database bootstrap. The earlier hosted-product design ticket assumed a future `smailnaild`, but that process does not exist yet.

If we skip directly to login or settings work, we will end up spreading configuration, DB opening, and server lifecycle concerns across ad hoc packages. The hosted app needs one place where:

- command-line/config/env parsing happens
- SQL connection details are resolved
- DB defaults are applied
- readiness can reflect database state
- later features can plug in without changing the process boundary again

## Proposed Solution

Add a new `cmd/smailnaild` binary with a single `serve` subcommand.

The `serve` command should include:

- a default section for host/port and small hosted-server flags
- Clay SQL connection and dbt sections

At runtime, the command should:

1. build a Clay `DatabaseConfig` from parsed Glazed values
2. apply SQLite-first defaults when the user did not supply DB configuration
3. open a real `sqlx.DB`
4. initialize a minimal metadata table
5. start an HTTP server exposing:
   - `GET /healthz`
   - `GET /readyz`
   - `GET /api/info`

The app DB bootstrap should stay generic enough for:

- SQLite via `database=smailnaild.sqlite`, `db-type=sqlite`
- Postgres via `dsn=postgres://...` or structured host/port/database/user flags

The first slice should not yet include:

- Keycloak client callbacks
- browser sessions
- per-user IMAP account CRUD
- authenticated MCP routes

## Design Decisions

### Decision 1: Start with `smailnaild`, not by mutating existing CLIs

The existing `smailnail` CLI solves operator-driven IMAP tasks. A hosted app has different lifecycle concerns and should not overload those roots.

### Decision 2: Use Clay SQL sections now

Clay already provides Glazed sections and config normalization for SQLite, Postgres/pgx, and DSN-based configuration. Reusing that avoids inventing another connection model.

### Decision 3: Default to SQLite for the hosted MVP

SQLite keeps local development and single-instance deployment easy. The command surface, config model, and storage package should still remain compatible with Postgres so the app can move later without reworking the CLI or repository contracts.

### Decision 4: Use stdlib HTTP for the first server slice

This step only needs health/readiness/info endpoints. Standard library HTTP keeps the bootstrap small and avoids framework churn before the real product surface exists.

## Alternatives Considered

### Reuse embedded OIDC SQLite storage as the app DB

Rejected. The OIDC package is framework-specific and SQLite-specific. The hosted app should own its own application data model.

### Jump straight to Postgres-only

Rejected for the first slice. It raises local setup cost without buying much yet.

### Add raw DB flags directly to `smailnaild`

Rejected because Clay SQL already solves this and gives us a migration path to multiple backends.

## Implementation Plan

1. Add `cmd/smailnaild/main.go` and a root command package.
2. Add a `serve` Glazed command with a hosted default section plus Clay SQL/dbt sections.
3. Add a small hosted app package that:
   - loads DB config from parsed values
   - applies SQLite-first defaults
   - opens the database
   - creates a metadata table
4. Add HTTP handlers and server bootstrap.
5. Add tests for DB config defaulting and HTTP endpoints.
6. Update `README.md` with hosted binary usage.
7. Keep diary/changelog/task tracking current as each code slice lands.

## Open Questions

- Whether the hosted app DB should use file-based SQLite by default in the repo root or under a dedicated data directory.
- Whether readiness should do a fresh ping on every request or cache DB state from startup.
- Whether to add a migration mechanism now or keep startup bootstrap to a single table for the first slice.

## References

- Clay SQL settings and connection helpers:
  - `/home/manuel/code/wesen/corporate-headquarters/clay/pkg/sql/settings.go`
  - `/home/manuel/code/wesen/corporate-headquarters/clay/pkg/sql/config.go`
  - `/home/manuel/code/wesen/corporate-headquarters/clay/pkg/sql/sources.go`
- Current hosted-product design:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md`
