---
Title: Design a JS IMAP API and sandboxed eval-style MCP for smailnail
Ticket: SMAILNAIL-004-JS-IMAP-EVAL-MCP-DESIGN
Status: active
Topics:
    - smailnail
    - go
    - email
    - review
    - mcp
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-goja/engine/factory.go
      Note: Runtime factory pattern for sandbox construction
    - Path: go-go-goja/modules/common.go
      Note: Native module contract for the future require smailnail module
    - Path: jesus/pkg/mcp/server.go
      Note: Reference for executeJS style MCP registration
    - Path: smailnail/pkg/dsl/processor.go
      Note: Core fetch pipeline the future service layer must wrap
    - Path: smailnail/pkg/imap/layer.go
      Note: Current IMAP connection settings and login entrypoint
ExternalSources: []
Summary: Design ticket for turning smailnail's Go IMAP logic into a reusable JavaScript API and exposing it through a sandboxed eval-style MCP server.
LastUpdated: 2026-03-08T22:52:20.157036218-04:00
WhatFor: Use this ticket to plan a hosted or local JavaScript execution environment where users can script IMAP workflows against smailnail through go-go-goja and an MCP server.
WhenToUse: Read this ticket before implementing a smailnail JS runtime, native module, or executeJS-style MCP surface.
---


# Design a JS IMAP API and sandboxed eval-style MCP for smailnail

## Overview

This ticket captures the design work for a new product layer, not a source-code change. The goal is to take `smailnail`'s existing IMAP and mail-generation capabilities, expose them through a stable JavaScript API, and then make that JavaScript environment callable through an MCP server in the style of `jesus`.

The recommended design is intentionally layered. First, extract or formalize a pure Go service layer inside `smailnail`. Second, wrap that layer as a native `go-go-goja` module such as `require("smailnail")`. Third, host that module inside a sandboxed runtime that exposes a constrained `executeJS`-style MCP tool instead of arbitrary process access.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary design doc**: [design-doc/01-js-imap-api-and-sandbox-eval-mcp-architecture-guide.md](./design-doc/01-js-imap-api-and-sandbox-eval-mcp-architecture-guide.md)
- **Diary**: [reference/01-investigation-diary.md](./reference/01-investigation-diary.md)
- **Reproducible scan**: [scripts/architecture-scan.sh](./scripts/architecture-scan.sh)

## Status

Current status: **active**

## Topics

- smailnail
- go
- email
- review
- mcp

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

## Current Recommendation

Build this in four explicit phases:

1. Introduce a `pkg/service` layer in `smailnail` so JavaScript bindings call stable Go APIs instead of Cobra command code.
2. Add a `go-go-goja` native module package that exports low-level IMAP operations and higher-level rule helpers.
3. Build a sandbox host that boots a runtime with a minimal module allowlist and no accidental file-system escape hatch.
4. Expose the sandbox through MCP with one `executeJS`-style tool and a small set of structured helper tools for common use cases.
