---
Title: Migrate go-go-mcp to glazed schema fields values sources APIs
Ticket: MCP-002-GLAZED-FACADE-MIGRATION
Status: active
Topics:
    - mcp
    - go
    - glazed
    - refactor
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources:
    - local:01-mcp-testing.md
Summary: ""
LastUpdated: 2026-03-08T19:10:45.748661167-04:00
WhatFor: ""
WhenToUse: ""
---


# Migrate go-go-mcp to glazed schema fields values sources APIs

## Overview

This ticket migrated `go-go-mcp` off Glazed's removed legacy `layers`, `parameters`, and `middlewares` APIs and onto the current `schema`, `fields`, `values`, and `sources` facade. The migration also removed a stale Parka-based helper path that still depended on the deleted Glazed packages.

Current state: the repository now passes `go test ./...` and `go build ./cmd/go-go-mcp` with the workspace active, and the diary/design doc capture both the direct code migration and the transitive dependency break that had to be resolved.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- mcp
- go
- glazed
- refactor

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
