---
Title: Implementation plan for local Dovecot and Keycloak stack
Ticket: SMAILNAIL-009-LOCAL-DOCKER-STACK-IMPLEMENTATION
Status: complete
Topics:
    - smailnail
    - go
    - sql
    - keycloak
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml
    - /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json
    - /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/README.md
ExternalSources: []
Summary: Implementation plan and outcome for a local Docker stack that provides Dovecot plus Keycloak-backed OIDC for smailnail development.
LastUpdated: 2026-03-15T19:28:00-04:00
WhatFor: Capture the local infrastructure choices and why they fit the next hosted smailnail work.
WhenToUse: Use when extending or troubleshooting the local dev stack for IMAP and OIDC work.
---

# Implementation plan for local Dovecot and Keycloak stack

## Executive Summary

The local development gap was simple: `smailnail` already had good IMAP-focused smoke paths, but the next hosted/OIDC work also needs an identity provider that behaves like the planned production shape. The implemented solution adds a single repo-local compose file that starts:

- Dovecot for IMAP testing
- PostgreSQL for Keycloak persistence
- Keycloak with an imported `smailnail-dev` realm

This keeps the local setup explicit, reproducible, and aligned with the likely production split of external Keycloak plus application-owned SQL storage.

## Problem Statement

The hosted `smailnaild` and MCP OIDC work cannot be validated against the existing IMAP-only fixture alone. We need a local stack that provides:

- a stable IMAP server for mailbox interactions
- a stable OIDC issuer for auth and discovery
- persistence for the issuer so realms and clients survive restarts

Without that stack, every auth experiment depends on ad hoc local setup and there is no documented baseline for the team.

## Proposed Solution

Add a `smailnail/docker-compose.local.yml` file and keep it intentionally narrow:

- `dovecot`
  - use the existing Dovecot fixture image
  - publish the common IMAP/POP/ManageSieve ports on `127.0.0.1`
- `keycloak-postgres`
  - use Postgres 16
  - keep persistence in a named volume
- `keycloak`
  - run Keycloak 26.5.5 in dev mode
  - store persistent state in the Postgres service
  - import a checked-in `smailnail-dev` realm on startup
  - publish Keycloak on `127.0.0.1:18080`

The realm import should provide starter clients:

- `smailnail-web`
- `smailnail-mcp`

The README should document the commands, ports, default credentials, and issuer URL so the stack is usable without reverse-engineering the compose file.

## Design Decisions

`docker-compose.local.yml` instead of changing the existing fixture setup

This keeps the local hosted/OIDC work isolated from the existing IMAP smoke path and avoids breaking older workflows that already rely on the separate Dovecot fixture.

Keycloak backed by PostgreSQL instead of embedded H2

The project discussion already leaned toward SQL portability and external OIDC. Using Postgres here makes the local issuer behave more like the intended deployed shape and avoids building new local assumptions around H2.

Realm import checked into the repo

The imported realm gives the team a repeatable starting point and removes manual console setup from the critical path.

Bind to `127.0.0.1`

These services are for local development. Loopback binding reduces accidental exposure and keeps the default stance conservative.

## Alternatives Considered

Keep using only the external `docker-test-dovecot` checkout

Rejected because it solves only IMAP, not OIDC, and leaves the hosted/app-auth path without a standard local environment.

Use Keycloak with embedded H2

Rejected for this stack because Postgres is closer to the expected deployed shape and is still simple enough locally.

Run Keycloak manually outside compose

Rejected because the main goal is a one-command local baseline that is easy to document and verify.

## Implementation Plan

1. Add the compose file with Dovecot, Keycloak, and Postgres.
2. Add a repo-local Keycloak realm import for `smailnail-dev`.
3. Update the README with startup and verification guidance.
4. Bring the stack up locally and verify:
   - `docker compose ps`
   - Keycloak discovery endpoint
   - Dovecot IMAPS reachability
5. Record the implementation and verification in the ticket docs.

Outcome: all five steps were completed in this ticket.

## Open Questions

Whether we also want seeded test users in the imported realm for automated browser or MCP auth flows. That is intentionally left out of this first slice.

Whether `smailnaild` should eventually join this compose file or remain a separately launched process during development. For now, keeping it separate is cleaner while the app is still moving quickly.

## References

- Ticket index: `../index.md`
- Diary: `../reference/01-diary.md`
- Stack file: `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml`
