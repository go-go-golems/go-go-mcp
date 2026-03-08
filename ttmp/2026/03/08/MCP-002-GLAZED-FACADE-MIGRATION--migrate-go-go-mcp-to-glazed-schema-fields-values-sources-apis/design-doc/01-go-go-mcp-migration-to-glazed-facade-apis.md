---
Title: go-go-mcp migration to glazed facade APIs
Ticket: MCP-002-GLAZED-FACADE-MIGRATION
Status: active
Topics:
    - mcp
    - go
    - glazed
    - refactor
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/go-go-mcp/cmds/client/layers/client.go
      Note: Client settings section migration
    - Path: cmd/go-go-mcp/cmds/server/layers/server.go
      Note: Server settings section migration
    - Path: pkg/cmds/cmd.go
      Note: Shell command schema/value migration
    - Path: pkg/tools/providers/config-provider/tool-provider.go
      Note: Facade middleware migration and Parka removal
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-08T18:30:22.401477264-04:00
WhatFor: ""
WhenToUse: ""
---


# go-go-mcp migration to glazed facade APIs

## Executive Summary

`go-go-mcp` currently compiles only when it ignores the workspace and resolves Glazed from the module cache. The local `glazed/` checkout has already removed the legacy `layers`, `parameters`, and `middlewares` packages, so the workspace build fails immediately. This ticket migrates `go-go-mcp` to Glazed's current facade packages: `schema`, `fields`, `values`, and `sources`.

The migration is intentionally direct. The goal is not to redesign command behavior, but to remove the broken dependency surface, preserve existing command semantics, and make `go test ./...` and `go build ./...` succeed with the workspace active.

## Problem Statement

The repository root uses a `go.work` file that points at the local `glazed/` checkout. `go-go-mcp` still imports removed Glazed packages and uses removed API names, so any workspace-mode build fails before runtime validation can begin.

The divergence shows up in three main areas:

- Shared command loading and shell-command execution in `pkg/cmds` and the config-provider path in `pkg/tools/providers/config-provider`
- `go-go-mcp` client/server Cobra commands and their reusable settings sections
- The scholarly example app commands, which repeat the old layer/parameter/value patterns

The scope of this ticket is limited to restoring compatibility with the current local Glazed APIs and documenting the migration clearly enough that a new engineer can continue work without re-deriving the mapping.

## Proposed Solution

The migration follows the Glazed facade refactor guide in the local `glazed` repo and applies it systematically:

- Replace legacy imports:
  - `layers` -> `schema`
  - `parameters` -> `fields`
  - `middlewares` -> `sources`
  - parsed-layer/value usage -> `values.Values`
- Replace deprecated command-description usage:
  - `WithLayersList(...)` -> `WithSections(...)`
  - `Description().Layers` -> `Description().Schema`
- Replace struct decoding:
  - `parsedLayers.InitializeStruct(...)` -> `parsedValues.DecodeSectionInto(...)`
- Replace tag syntax:
  - `glazed.parameter:"foo"` -> `glazed:"foo"`
- Replace config execution:
  - `middlewares.ExecuteMiddlewares(...)` -> `sources.Execute(...)`
  - `layers.NewParsedLayers()` -> `values.New()`

The migration should preserve CLI flag names, defaults, argument semantics, and user-visible behavior unless the newer Glazed APIs force a correctness fix.

## Design Decisions

- Use Glazed's current facade APIs directly instead of trying to restore compatibility shims inside `go-go-mcp`.
- Keep command-level behavior stable. This is a compatibility refactor, not a UX redesign.
- Migrate the shared command/config-provider layer first so downstream compile errors become smaller and easier to reason about.
- Validate in workspace mode, because that is the actual broken integration path this ticket exists to fix.

## Alternatives Considered

- Pin `go-go-mcp` back to a released Glazed version and continue using `GOWORK=off`
  - Rejected because it preserves the divergence instead of fixing it.
- Reintroduce legacy aliases in `glazed/`
  - Rejected because the local Glazed repo already has an explicit facade migration guide and the user asked to fix `go-go-mcp`.
- Rewrite command definitions more aggressively around a new app structure
  - Rejected because it increases risk and mixes compatibility work with architectural change.

## Implementation Plan

1. Update ticket bookkeeping and record the migration inventory.
2. Migrate the shared command-description and config-provider code.
3. Migrate `cmd/go-go-mcp/...` client/server commands and reusable sections.
4. Migrate scholarly commands.
5. Run `gofmt`, `go test ./...`, and targeted runtime/build validation in workspace mode.
6. Record exact fixes, residual risks, and review instructions in the diary and changelog.

## Open Questions

- Whether any legacy helper names outside the obvious package renames still appear after the main pass.
- Whether Parka-derived parameter-filter middleware maps cleanly onto the current `sources` interfaces without additional adaptation.
- Whether help-text grouping helpers in Glazed have changed names and need follow-on cleanup after compile validation.

## References

- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/glazed/pkg/doc/tutorials/migrating-to-facade-packages.md`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go.work`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/cmds/cmd.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/tools/providers/config-provider/tool-provider.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/cmd/go-go-mcp/cmds/client/layers/client.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/cmd/go-go-mcp/cmds/server/layers/server.go`
