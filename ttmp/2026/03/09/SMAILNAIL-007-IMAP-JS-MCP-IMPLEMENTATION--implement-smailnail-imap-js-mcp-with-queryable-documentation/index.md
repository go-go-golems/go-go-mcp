---
Title: Implement smailnail IMAP JS MCP with queryable documentation
Ticket: SMAILNAIL-007-IMAP-JS-MCP-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - mcp
    - javascript
    - documentation
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Execution ticket for implementing the dedicated smailnail IMAP JavaScript MCP with executeIMAPJS and getIMAPJSDocumentation.
LastUpdated: 2026-03-09T16:03:04.709171591-04:00
WhatFor: Track the concrete implementation of the smailnail IMAP JS MCP and its documentation-query stack.
WhenToUse: Use when reviewing the implementation work, commits, and validation for the dedicated IMAP JS MCP.
---

# Implement smailnail IMAP JS MCP with queryable documentation

## Overview

This ticket implements the architecture defined in `SMAILNAIL-006`. The target runtime is intentionally narrow: one MCP server hosted in `smailnail` that exposes only `executeIMAPJS` and `getIMAPJSDocumentation`.

The implementation is being done in small slices with commits after each completed unit of work. Code lives in `smailnail`; the ticket artifacts, task tracking, and diary live here under `go-go-mcp/ttmp`.

## Key Links

- Implementation guide: `design-doc/01-implement-the-smailnail-imap-js-mcp-and-queryable-docs.md`
- Diary: `reference/01-implementation-diary.md`
- Task list: `tasks.md`

## Status

Current status: **active**

## Topics

- smailnail
- mcp
- javascript
- documentation

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
