---
Title: Initial assessment of go-go-mcp
Ticket: MCP-001-INITIAL-ASSESSMENT
Status: active
Topics:
    - mcp
    - go
    - assessment
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Assessment ticket for go-go-mcp covering current runtime architecture, working validation paths, stale or misleading surfaces, and recommended cleanup priorities."
LastUpdated: 2026-03-08T17:51:19.976886357-04:00
WhatFor: "Track the initial health assessment of go-go-mcp and collect the resulting investigation artifacts."
WhenToUse: "Use this ticket when you need the high-level summary and links for the go-go-mcp assessment work."
---

# Initial assessment of go-go-mcp

## Overview

This ticket evaluates whether `go-go-mcp` still works, how far behind current MCP expectations it is, what parts are healthy, what parts are stale, and what cleanup should happen first.

Headline result: the repo is currently broken in its checked-out multi-module workspace form, but the standalone `go-go-mcp` module is still functional and worth preserving. The live runtime path is centered on `pkg/embeddable` plus `mcp-go`, while documentation and some legacy packages still reflect older architectural assumptions.

## Key Links

- Design doc: `design-doc/01-go-go-mcp-initial-assessment-and-modernization-guide.md`
- Diary: `reference/01-investigation-diary.md`
- Smoke scripts: `scripts/standalone-smoke.sh`, `scripts/oidc-smoke.sh`

## Status

Current status: **active**

## Topics

- mcp
- go
- assessment
- architecture

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
