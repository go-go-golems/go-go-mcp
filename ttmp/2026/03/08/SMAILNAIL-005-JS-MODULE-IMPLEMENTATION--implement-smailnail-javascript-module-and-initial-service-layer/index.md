---
Title: Implement smailnail JavaScript module and initial service layer
Ticket: SMAILNAIL-005-JS-MODULE-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - go
    - email
    - mcp
    - javascript
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: smailnail/pkg/js/modules/smailnail/module.go
      Note: Native go-go-goja module registration and JS exports
    - Path: smailnail/pkg/js/modules/smailnail/module_test.go
      Note: Runtime integration coverage for require smailnail
    - Path: smailnail/pkg/services/smailnailjs/service.go
      Note: Primary JS-facing service layer and dialer-backed connection abstraction
    - Path: smailnail/pkg/services/smailnailjs/views.go
      Note: Conversion layer from internal dsl types to JS-friendly objects
    - Path: smailnail/scripts/js-module-smoke.sh
      Note: Maintained smoke path for the new JS module slice
ExternalSources: []
Summary: Initial implementation ticket for a reusable smailnail Go service layer and native go-go-goja module.
LastUpdated: 2026-03-08T23:06:02.627686024-04:00
WhatFor: Use this ticket to track the first end-to-end implementation of smailnail's JavaScript surface before the eval-style MCP host is added.
WhenToUse: Read this ticket when implementing or reviewing the service package, native module, tests, or smoke/demo path.
---


# Implement smailnail JavaScript module and initial service layer

## Overview

This ticket tracks the first implementation slice for the JavaScript work designed in `SMAILNAIL-004`. The scope here is intentionally narrower than the final hosted MCP system: add a reusable Go service layer, expose it as a native `go-go-goja` module, and prove that a runtime can load `require("smailnail")`.

The output of this ticket should be real code in `smailnail`, integration tests against a real Goja runtime, a small smoke/demo path, and detailed ticket bookkeeping that explains exactly what was implemented and why.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- smailnail
- go
- email
- mcp
- javascript

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
