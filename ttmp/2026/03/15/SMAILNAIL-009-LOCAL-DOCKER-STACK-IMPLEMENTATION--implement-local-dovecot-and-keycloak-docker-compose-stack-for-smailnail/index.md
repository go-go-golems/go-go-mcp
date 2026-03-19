---
Title: Implement local Dovecot and Keycloak docker compose stack for smailnail
Ticket: SMAILNAIL-009-LOCAL-DOCKER-STACK-IMPLEMENTATION
Status: complete
Topics:
    - smailnail
    - go
    - sql
    - keycloak
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml
      Note: Repo-local Docker Compose stack for Dovecot, Keycloak, and Postgres
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json
      Note: Imported development realm and starter OIDC clients
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/README.md
      Note: Developer-facing startup, port, and credential instructions for the local stack
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/.gitignore
      Note: Ignore local SQLite runtime artifacts produced during hosted-app work
ExternalSources: []
Summary: Local Docker Compose stack added for Dovecot plus Keycloak-with-Postgres, with an imported development realm and documented local defaults.
LastUpdated: 2026-03-15T19:28:00-04:00
WhatFor: Stand up a repeatable local environment for smailnail IMAP and OIDC work without depending on external infrastructure.
WhenToUse: Use when developing smailnaild, Keycloak integration, or MCP OIDC flows locally.
---

# Implement local Dovecot and Keycloak docker compose stack for smailnail

## Overview

This ticket adds a repo-local Docker Compose stack for the two external systems the next smailnail slices depend on:

- a Dovecot test fixture exposed on the usual local IMAP ports
- a Keycloak instance backed by PostgreSQL, with a pre-imported `smailnail-dev` realm

The implementation is complete. The stack was brought up locally, the Keycloak discovery endpoint was verified, and Dovecot IMAPS was confirmed reachable on `127.0.0.1:993`.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **complete**

## Topics

- smailnail
- go
- sql
- keycloak

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
