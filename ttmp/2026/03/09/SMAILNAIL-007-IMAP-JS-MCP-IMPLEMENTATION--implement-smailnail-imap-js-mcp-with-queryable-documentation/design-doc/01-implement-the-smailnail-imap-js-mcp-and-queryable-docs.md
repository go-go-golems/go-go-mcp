---
Title: Implement the smailnail IMAP JS MCP and queryable docs
Ticket: SMAILNAIL-007-IMAP-JS-MCP-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - mcp
    - javascript
    - documentation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: smailnail/cmd/smailnail-imap-mcp/main.go
      Note: New dedicated MCP binary entrypoint
    - Path: smailnail/pkg/mcp/imapjs/execute_tool.go
      Note: Implements the first executeIMAPJS runtime path
    - Path: smailnail/pkg/mcp/imapjs/execute_tool_test.go
      Note: Validates success and error execution paths
    - Path: smailnail/pkg/mcp/imapjs/server.go
      Note: Registers the two-tool MCP surface
ExternalSources: []
Summary: Detailed implementation guide for building the dedicated smailnail IMAP JavaScript MCP, its queryable documentation tool, and its validation/smoke coverage.
LastUpdated: 2026-03-09T16:05:07.176403861-04:00
WhatFor: Guide the implementation of the smailnail IMAP JS MCP in concrete, reviewable steps.
WhenToUse: Use while implementing, reviewing, or extending the smailnail IMAP JS MCP and its documentation stack.
---


# Implement the smailnail IMAP JS MCP and queryable docs

## Executive Summary

This implementation ticket turns the design from `SMAILNAIL-006` into code. The deliverable is a new `smailnail` binary that embeds `go-go-mcp`, boots a fresh Goja runtime per request, exposes the existing `smailnail` JavaScript module through an `executeIMAPJS` tool, and serves a structured documentation query surface through `getIMAPJSDocumentation`.

The safest implementation order is:

1. add the dedicated MCP binary and `executeIMAPJS`,
2. validate the runtime/tool surface,
3. add structured JS doc assets and a store loader,
4. add the documentation tool,
5. lock the result down with drift tests and smoke scripts.

## Problem Statement

`smailnail` already has a reusable JS-facing service layer in `pkg/services/smailnailjs` and a native module in `pkg/js/modules/smailnail`, but there is no MCP host exposing that surface yet. The project also lacks queryable symbol-level documentation suitable for an eval-style MCP tool.

The implementation needs to satisfy several constraints at once:

- keep the MCP surface intentionally small,
- avoid duplicating business logic in the handler layer,
- keep documentation close to the JS API surface,
- return structured, agent-friendly results,
- and provide a validation path that catches documentation drift and runtime regressions.

## Proposed Solution

### Build targets

Add a new CLI target:

- `cmd/smailnail-imap-mcp`

Add a new implementation package:

- `pkg/mcp/imapjs`

Add embedded documentation assets near the JS module:

- `pkg/js/modules/smailnail/docs/*.js`

### Runtime structure

The runtime should remain per-request and disposable.

```text
MCP tool request
   |
   v
executeIMAPJS handler
   |
   +--> bind request with embeddable.Arguments
   +--> build go-go-goja runtime
   +--> register native smailnail module
   +--> run JavaScript
   +--> return JSON result

getIMAPJSDocumentation handler
   |
   +--> bind query request
   +--> query in-memory DocStore
   +--> optionally render markdown via exportmd
   +--> return JSON result
```

### Concrete implementation phases

#### Phase 1: binary and base server wiring

Files to add:

- `smailnail/cmd/smailnail-imap-mcp/main.go`
- `smailnail/pkg/mcp/imapjs/server.go`
- `smailnail/pkg/mcp/imapjs/types.go`

Responsibilities:

- initialize logging consistently with the other CLIs,
- add the `mcp` command with two tools only,
- set a clear server name/description,
- keep stdio as the first validation target.

#### Phase 2: `executeIMAPJS`

Files to add:

- `smailnail/pkg/mcp/imapjs/execute_tool.go`
- `smailnail/pkg/mcp/imapjs/execute_tool_test.go`

Responsibilities:

- define request/response types,
- bind the request with `embeddable.Arguments.BindArguments`,
- boot a runtime using `go-go-goja/engine.NewBuilder()`,
- register the `smailnail` module explicitly,
- evaluate user code,
- shape JSON output with success, value, and structured error fields.

The first version can defer fancy console capture if the runtime path is otherwise sound, but the response should be designed to carry console output later without breaking schema.

#### Phase 3: embedded JS docs and loader

Files to add:

- `smailnail/pkg/js/modules/smailnail/docs/package.js`
- `smailnail/pkg/js/modules/smailnail/docs/service.js`
- `smailnail/pkg/js/modules/smailnail/docs/examples.js`
- `smailnail/pkg/mcp/imapjs/docs_registry.go`

Responsibilities:

- embed the docs FS,
- parse each JS doc file with `jsdoc/extract`,
- build a `model.DocStore`,
- cache the store for repeated tool calls.

#### Phase 4: `getIMAPJSDocumentation`

Files to add:

- `smailnail/pkg/mcp/imapjs/docs_tool.go`
- `smailnail/pkg/mcp/imapjs/docs_query.go`
- `smailnail/pkg/mcp/imapjs/docs_tool_test.go`

Responsibilities:

- support `overview`, `package`, `symbol`, `example`, `concept`, `search`, and `render`,
- return structured objects first,
- use `exportmd.Write(...)` only for render mode.

#### Phase 5: drift validation and smoke coverage

Files to add:

- `smailnail/pkg/mcp/imapjs/docs_validation_test.go`
- `smailnail/scripts/imap-js-mcp-smoke.sh`

Files to update:

- `smailnail/Makefile`
- `smailnail/README.md`

Responsibilities:

- confirm documented symbols exist,
- confirm examples refer to real symbols,
- smoke `list-tools`,
- smoke `executeIMAPJS`,
- smoke `getIMAPJSDocumentation`.

## Design Decisions

### Decision: implement in `smailnail`, not in `go-go-mcp`

`go-go-mcp` is the framework. The runtime behavior, doc assets, and tests belong in the application repo that owns the IMAP surface.

### Decision: keep the server to two tools

That is the user’s requested boundary and also the cleanest way to avoid turning the MCP adapter into another application framework.

### Decision: use JS sentinel docs as the canonical documentation source

The workspace already contains the extraction and store logic in `go-go-goja/pkg/jsdoc`. Reusing it keeps prose and examples close to the JS API and avoids inventing a parallel doc format.

### Decision: keep drift detection in tests, not in the runtime path

Export validation is valuable, but it should fail CI/tests rather than complicate every tool call.

## Alternatives Considered

### Alternative: implement docs as one embedded markdown file

Rejected because the second tool is explicitly intended to be queryable and symbol-aware.

### Alternative: build the documentation tool before `executeIMAPJS`

Rejected because the docs tool depends on knowing the actual exported API shape, and the execution tool gives the fastest runtime feedback.

### Alternative: move generic query helpers into `go-go-goja` first

Rejected for now because that would delay shipping the first concrete consumer. Generic extraction can happen after the `smailnail` implementation proves which abstractions are real.

## Implementation Plan

### Task order

1. Create binary and base MCP wiring.
2. Implement and test `executeIMAPJS`.
3. Commit the runtime slice.
4. Add embedded docs and registry.
5. Implement and test `getIMAPJSDocumentation`.
6. Commit the docs slice.
7. Add drift tests and maintained smoke script.
8. Update README/Makefile and final validation.
9. Commit final docs and ticket updates.

### Validation checkpoints

- `go test ./pkg/js/modules/smailnail -count=1`
- `go test ./pkg/mcp/imapjs -count=1`
- `go test ./... -count=1`
- `go build ./cmd/smailnail-imap-mcp`
- `go run ./cmd/smailnail-imap-mcp mcp list-tools`
- maintained smoke script

### Review notes for the intern

- Do not put IMAP business logic in the MCP handlers. Reuse `smailnailjs.Service`.
- Keep tool results deterministic JSON strings wrapped in `protocol.NewToolResult(protocol.WithText(...))`.
- Keep docs files short and composable. A few focused symbol/example files are easier to maintain than one giant JS reference file.

## Open Questions

The only live open question for implementation is how much console capture to support in v1. The runtime already provides a console, but reliable capture may require extra wiring. If that becomes expensive, keep the response field reserved and ship the execution path first.

## References

- `smailnail/pkg/js/modules/smailnail/module.go`
- `smailnail/pkg/js/modules/smailnail/module_test.go`
- `smailnail/pkg/services/smailnailjs/service.go`
- `go-go-goja/engine/factory.go`
- `go-go-goja/pkg/jsdoc/model/store.go`
- `go-go-goja/pkg/jsdoc/extract/extract.go`
- `go-go-goja/pkg/jsdoc/exportmd/exportmd.go`
- `go-go-mcp/pkg/embeddable/command.go`
- `go-go-mcp/pkg/embeddable/arguments.go`
