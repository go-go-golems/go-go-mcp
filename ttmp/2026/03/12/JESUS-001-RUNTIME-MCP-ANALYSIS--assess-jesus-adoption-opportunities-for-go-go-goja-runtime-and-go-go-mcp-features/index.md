---
Title: Assess jesus adoption opportunities for go-go-goja runtime and go-go-mcp features
Ticket: JESUS-001-RUNTIME-MCP-ANALYSIS
Status: complete
Topics:
    - go
    - javascript
    - mcp
    - review
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-goja/engine/factory.go
      Note: Factory-owned runtime construction model
    - Path: go-go-goja/pkg/runtimeowner/runner.go
      Note: Owner-mediated runtime access semantics
    - Path: go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: Current mcp-go backend and transport capabilities
    - Path: jesus/pkg/engine/dispatcher.go
      Note: Custom dispatcher and execution queue design
    - Path: jesus/pkg/engine/engine.go
      Note: Current runtime lifecycle and direct eval paths
    - Path: jesus/pkg/mcp/server.go
      Note: Current MCP tool registration and server startup flow
ExternalSources: []
Summary: Ticket workspace for the assessment of which newer go-go-goja and go-go-mcp capabilities should be adopted by jesus, with an evidence-backed design doc and diary.
LastUpdated: 2026-03-12T23:42:15.787477064-04:00
WhatFor: ""
WhenToUse: ""
---



# Assess jesus adoption opportunities for go-go-goja runtime and go-go-mcp features

## Overview

This ticket stores the architectural assessment of whether `jesus` should adopt newer `go-go-goja` runtime-engine features and newer `go-go-mcp` capabilities.

The primary conclusion is that `jesus` should prioritize a runtime-layer refactor onto the `go-go-goja` factory/owner model first, then selectively improve its MCP surface with enhanced tools and, later, resources.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary design doc**: `design-doc/01-jesus-runtime-and-mcp-evolution-analysis.md`
- **Investigation diary**: `reference/01-diary.md`

## Status

Current status: **active**

## Topics

- go
- javascript
- mcp
- review

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
