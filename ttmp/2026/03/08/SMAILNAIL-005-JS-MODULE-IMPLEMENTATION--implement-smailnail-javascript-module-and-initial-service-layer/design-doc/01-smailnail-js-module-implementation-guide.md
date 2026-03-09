---
Title: smailnail JS module implementation guide
Ticket: SMAILNAIL-005-JS-MODULE-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - go
    - email
    - mcp
    - javascript
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Implementation guide for the first smailnail JavaScript module milestone, covering the service layer, native module registration, tests, and smoke validation.
LastUpdated: 2026-03-08T23:06:02.688769124-04:00
WhatFor: Use this guide to implement and review the first smailnail native JavaScript module.
WhenToUse: Read this before modifying the new service package or JS module code.
---

# smailnail JS module implementation guide

## Executive Summary

This ticket implements the first production-quality JavaScript surface for `smailnail`. The immediate goal is not the full sandboxed MCP server yet. Instead, this ticket focuses on the lower layers that the future MCP server will depend on: a pure Go service package and a native `go-go-goja` module called `smailnail`.

## Problem Statement

`smailnail` currently exposes its useful behavior through package code plus CLI adapters. There is no stable API boundary for JavaScript code to call, and there is no registered native module that a Goja runtime can load via `require("smailnail")`.

## Proposed Solution

Implement the feature in three stages:

1. Add a small Go service package that owns rule parsing from YAML strings, rule construction from JS-friendly option structs, and shaping `dsl.EmailMessage` values into plain result objects.
2. Add an injectable service/session boundary so JS-facing tests can validate behavior without a live IMAP server.
3. Register a native `smailnail` module and add integration tests that boot a `go-go-goja` runtime and call module exports.

## Design Decisions

- Keep the first exported API synchronous where possible.
- Prefer plain Go/JS values over exposing raw IMAP client structs.
- Keep the initial module surface small: rule utilities and service construction first, remote IMAP execution second.
- Build on `engine.NewBuilder().WithModules(...).Build().NewRuntime(...)`, not removed legacy helper patterns.

## Alternatives Considered

- Directly exposing CLI command structs was rejected because it would couple the JS API to flag parsing and Glazed sections.
- Building the full MCP server in the same first commit was rejected because it would mix runtime-host concerns with the lower-level module work that needs tests first.

## Implementation Plan

1. Wire the new dependency and confirm the workspace still builds.
2. Create `pkg/services/smailnailjs`.
3. Port the CLI rule-building logic into the service package in a JS-friendly form.
4. Add a module package under `pkg/js/modules/smailnail`.
5. Add unit tests and goja integration tests.
6. Add a smoke/demo entrypoint.
7. Update ticket docs and validate with `docmgr doctor`.

## Open Questions

- Whether the first service constructor should expose connection methods immediately or keep the first version fully rule-oriented.
- Whether the maintained smoke path belongs in `cmd/` or under `scripts/`.

## References

- `SMAILNAIL-004-JS-IMAP-EVAL-MCP-DESIGN`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/modules/common.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go`
