---
Title: Implement smailnaild hosted app skeleton with Clay SQL bootstrap
Ticket: SMAILNAIL-008-SMAILNAILD-SQL-BOOTSTRAP-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - glazed
    - sql
    - go
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../code/wesen/corporate-headquarters/clay/pkg/sql/config.go
      Note: Driver normalization and sqlx connection bootstrap for SQLite and Postgres
    - Path: ../../../../../../../../../code/wesen/corporate-headquarters/clay/pkg/sql/settings.go
      Note: Clay SQL section and config helpers that the hosted binary should reuse
    - Path: smailnail/README.md
      Note: Current repo shape and command surface for the hosted binary work
ExternalSources: []
Summary: 'Implementation ticket for the first hosted smailnaild slice: a Glazed-based root command, a serve command, Clay SQL-backed application database bootstrap, and a minimal HTTP server surface suitable for later login, settings, and MCP work.'
LastUpdated: 2026-03-15T18:03:01.252598537-04:00
WhatFor: Use this ticket to implement the initial hosted smailnaild binary with a database connection model that is SQLite-first but Postgres-ready through Clay SQL sections.
WhenToUse: Use when building or reviewing the first deployable hosted smailnail application skeleton and its DB bootstrap path.
---


# Implement smailnaild hosted app skeleton with Clay SQL bootstrap

## Overview

This ticket implements the first hosted `smailnaild` slice rather than the full product described in `SMAILNAIL-003`. The focus is narrow and actionable:

- add a new `smailnaild` binary
- use Glazed/Cobra for the command surface
- use Clay SQL sections for app DB configuration
- bootstrap a real SQL connection in a way that works for SQLite today and Postgres later
- expose a small HTTP surface that proves the hosted process can start, report health, and advertise DB/runtime metadata

The point of this slice is to establish the hosted process boundary correctly before layering in Keycloak login, per-user IMAP settings, or user-aware MCP execution.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

Implementation status:

- ticket created
- implementation plan being written
- granular tasks being tracked
- code implementation pending

## Topics

- smailnail
- glazed
- sql
- go

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
