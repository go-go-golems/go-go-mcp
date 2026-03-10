---
Title: Design IMAP JS MCP with queryable documentation
Ticket: SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN
Status: active
Topics:
    - smailnail
    - mcp
    - javascript
    - documentation
    - oidc
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Design ticket for a dedicated smailnail IMAP JavaScript MCP with a structured documentation query tool backed by go-go-goja jsdoc extraction.
LastUpdated: 2026-03-09T00:42:44.052515175-04:00
WhatFor: Assess and design a minimal IMAP JavaScript MCP with a rich documentation query surface.
WhenToUse: Use when planning or onboarding implementation of the dedicated smailnail IMAP JS MCP server.
---

# Design IMAP JS MCP with queryable documentation

## Overview

This ticket designs a dedicated MCP surface for the `smailnail` JavaScript API. The target is intentionally small: one execution tool, `executeIMAPJS`, and one documentation tool, `getIMAPJSDocumentation`.

The key design question is how documentation should be authored, kept in sync, and queried. The recommended answer is to reuse the existing `go-go-goja/pkg/jsdoc` extraction and store model, author canonical docs in JavaScript sentinel files, and add a thin Go-side query and validation layer.

## Key Links

- Primary design doc: `design-doc/01-imap-js-mcp-and-queryable-documentation-architecture-guide.md`
- Investigation diary: `reference/01-investigation-diary.md`
- Architecture scan script: `scripts/mcp-docs-architecture-scan.sh`

## Status

Current status: **active**

## Topics

- smailnail
- mcp
- javascript
- documentation
- oidc

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- design-doc/ - Primary ticket analysis and implementation guides
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
