---
Title: Legacy MCP dependency and deprecated code cleanup design
Ticket: 2026-08-31-GO-GO-MCP-CLEANUP
Status: active
Topics:
    - mcp
    - go
    - cleanup
    - dependencies
    - refactoring
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/cmd/go-go-mcp/cmds/client/helpers/client.go
      Note: Current official SDK client path replacing pkg/client
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/go.mod
      Note: Dependency declarations including official SDK and legacy SSE package
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/client/client.go
      Note: Legacy client with hardcoded 2024-11-05 protocol version; first deletion candidate
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/client/sse.go
      Note: |-
        Legacy SSE transport and source of the removable r3labs/sse dependency
        Legacy SSE transport and exclusive r3labs/sse dependency candidate
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/client/stdio.go
      Note: Legacy stdio transport paired with the obsolete client abstraction
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/embeddable/official_backend.go
      Note: Active official SDK server adapter that still consumes the custom facade
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/protocol
      Note: |-
        Still-used internal protocol/domain types; not a Phase 1 deletion target
        Active compatibility types used by registries and the official backend adapter
ExternalSources: []
Summary: Remove orphaned pre-official-SDK client code without prematurely deleting the still-used compatibility facade.
LastUpdated: 2026-08-31T15:10:00-04:00
WhatFor: Future cleanup of go-go-mcp after the official MCP Go SDK migration.
WhenToUse: Use this design before deleting legacy client, protocol, or facade packages.
---



# Legacy MCP dependency and deprecated code cleanup design

## Executive Summary

`go-go-mcp` contains two different generations of MCP code. The repository's current client command and embeddable HTTP backend use `github.com/modelcontextprotocol/go-sdk`, while `pkg/client` is an older standalone client implementation with its own protocol types and an SSE dependency. Repository-wide import inspection found no production import of `pkg/client`; its client initialization also hardcodes MCP `2024-11-05`.

The first cleanup phase should therefore delete the orphaned `pkg/client` implementation and remove its exclusive `github.com/r3labs/sse/v2` dependency. A second phase may migrate the internal `pkg/protocol`/tool-registry facade to official SDK types, but that is not a safe deletion-only change: the official backend, registries, tool providers, examples, and public embeddable API still depend on those types. The phases must remain separate so a small dependency cleanup does not become an unreviewable framework rewrite.

## Problem Statement

The official SDK is now the authoritative MCP wire implementation, but the repository still contains legacy code that suggests the old client is supported. This creates three risks:

1. **Conflicting protocol signals.** `pkg/client/client.go` sends `2024-11-05`, while the official SDK path is the maintained implementation.
2. **Dead dependency surface.** `pkg/client/sse.go` is the only source use of `github.com/r3labs/sse/v2` found in the repository scan.
3. **Migration ambiguity.** `pkg/protocol` is not dead code. It is currently the internal domain contract connecting tool providers and registries to the official SDK adapter in `pkg/embeddable/official_backend.go`.

The cleanup must remove misleading legacy code while preserving current public behavior, the official backend, tool authorization, examples that remain supported, and the standalone `go-go-mcp` CLI.

## Scope

### In scope

- Confirm and remove the unused `pkg/client` package.
- Remove dependencies used only by that package, especially `github.com/r3labs/sse/v2`.
- Audit and remove one-off migration spike tests when equivalent official-backend coverage exists.
- Update documentation and package comments so the official SDK is the only documented client/server wire implementation.
- Record a separate future design for migrating the internal compatibility facade if desired.

### Out of scope for the first phase

- Deleting `pkg/protocol`.
- Rewriting all tool registries and providers around official SDK request/result types.
- Removing the `pkg/embeddable` public API.
- Removing the scholarly application or its domain clients.
- Reintroducing Mark3 Labs or `mcp-go` as a compatibility path.

## Current-State Architecture

### Official SDK paths

The command-line client uses `mcp.NewClient` and the official transport in `cmd/go-go-mcp/cmds/client/helpers/client.go`. The embeddable backend constructs an official SDK server in `pkg/embeddable/official_backend.go`, registers mapped tools, and mounts the official Streamable HTTP and SSE handlers.

The official backend is intentionally an adapter boundary. It accepts the repository's registry/provider model, maps `pkg/protocol.Tool` to official SDK tools, invokes the existing provider, and maps `pkg/protocol.ToolResult` back to official SDK results. This is active code, not cleanup residue.

### Legacy client path

`pkg/client/client.go` defines a separate `Client` abstraction over a custom `Transport`. `pkg/client/stdio.go` and `pkg/client/sse.go` implement transports using the repository's old `pkg/protocol` types. `Client.Initialize` sends protocol version `2024-11-05` directly. Repository-wide import inspection found no import of `github.com/go-go-golems/go-go-mcp/pkg/client`; the current CLI client imports the official SDK instead.

### Compatibility facade

`pkg/protocol`, `pkg/tools`, `pkg/resources`, `pkg/prompts`, `pkg/session`, and `pkg/embeddable/server.go` remain connected. The registry and provider layers use the custom protocol/domain types, and `pkg/embeddable/official_mapping.go` translates them into official SDK values. Deleting these packages requires an explicit API migration plan and broad test updates; they should not be removed as part of the orphaned-client cleanup.

## Proposed Solution

### Phase 1: delete the orphaned client

Delete:

- `pkg/client/client.go`
- `pkg/client/stdio.go`
- `pkg/client/sse.go`
- `pkg/client/logcopter.go`

Then remove `github.com/r3labs/sse/v2` from `go.mod` and `go.sum` if `go mod tidy` confirms no remaining use. Run the full test suite and package/build enumeration to ensure no hidden consumer exists.

Acceptance criteria:

- `rg 'go-go-mcp/pkg/client' .` returns no imports.
- `go list -deps ./...` no longer includes the legacy client package or `r3labs/sse`.
- `go test ./...` passes.
- The official CLI client still supports its tested transports and operations.
- No public command or documented example refers to the deleted package.

### Phase 2: retire migration-only spike coverage

Inspect `pkg/embeddable/official_sdk_spike_test.go`. If its assertions are fully covered by the official backend conformance, mapping, authorization, and transport tests, delete the spike test. If it still covers a unique regression, rename it to describe the behavior under test rather than retaining “spike” terminology.

### Phase 3: decide whether to remove the custom protocol facade

This phase is a separate architectural project. Inventory every exported type and package consumer, then choose one of two deliberate outcomes:

1. Keep `pkg/protocol` as an internal/public compatibility model and document that the official SDK owns wire behavior; or
2. Migrate registries, providers, examples, and embeddable APIs to official SDK types, then remove the custom protocol model and mapping layer in a coordinated breaking-change or major-version release.

No Phase 3 deletion should begin from import-count evidence alone.

## Design Decisions

### DR-1: Delete orphaned client code first

- **Context:** `pkg/client` is unused and hardcodes an obsolete protocol version.
- **Options:** retain it, mark it deprecated, or delete it.
- **Decision:** delete it after a final import/build audit.
- **Rationale:** retaining an unused client advertises an unsupported protocol implementation and preserves an unnecessary dependency.
- **Consequences:** callers of the old package must migrate to the official SDK; repository-internal callers have already migrated.
- **Status:** proposed.

### DR-2: Do not delete the compatibility facade in the same change

- **Context:** official backend code still consumes custom registry and protocol types.
- **Options:** delete all custom types now, keep them indefinitely, or separate the migration.
- **Decision:** separate the changes and preserve the facade for now.
- **Rationale:** this keeps the cleanup reviewable and avoids coupling dependency removal to a broad public API change.
- **Consequences:** a small amount of adapter code remains until a future decision is made.
- **Status:** accepted.

### DR-3: Official SDK owns protocol negotiation

- **Context:** the repository now uses the official SDK for MCP wire behavior.
- **Options:** maintain a custom client/protocol implementation, use the legacy client, or use the official SDK.
- **Decision:** use the official SDK and its current supported protocol version for client/server acceptance tests.
- **Rationale:** one implementation avoids contradictory protocol behavior and duplicate maintenance.
- **Consequences:** compatibility tests should target the official SDK's supported version rather than preserving a legacy client's hardcoded version.
- **Status:** accepted.

## Alternatives Considered

### Delete the entire repository's custom MCP layer

Rejected for this ticket. The layer is still used by the official backend adapter and by multiple tool/provider packages. This would require a coordinated API migration, not cleanup.

### Keep `pkg/client` for external users

Rejected unless a concrete downstream consumer is identified. If consumers exist, publish a migration note and deprecation window rather than silently deleting it.

### Replace the official SDK with `go-go-mcp`'s old client

Rejected. The old client is precisely the obsolete implementation being removed and does not provide the current official Streamable HTTP behavior.

## Implementation Plan

1. Run `rg`, `go list -deps`, and `go test ./...` from the repository root; record any external consumers.
2. Delete `pkg/client` and remove its exclusive dependency.
3. Run `gofmt`, `go mod tidy`, unit tests, integration tests, lint, and vulnerability checks.
4. Audit `official_sdk_spike_test.go` against existing official-backend tests.
5. Delete or rename the spike test in a separate focused commit.
6. Update package documentation, changelog, and migration notes.
7. Revisit Phase 3 only with an inventory of exported protocol/facade APIs and a dedicated design review.

## Validation Strategy

```text
go test ./...
go list -deps ./... | grep -E 'pkg/client|r3labs/sse'  # expect no output
go vet ./...
golangci-lint run ./...
rg -n 'pkg/client|r3labs/sse|2024-11-05' .
```

The final search should distinguish historical changelog references from active implementation or user-facing documentation. Build the CLI and exercise the official client command separately if those commands are retained.

## Risks and Rollback

- **Hidden external users:** deleting an exported package can break downstream consumers. Mitigate with a release note, migration guidance, and a versioning decision before merge.
- **Indirect dependency retention:** `go mod tidy` may retain a dependency through another package. Remove it only after dependency graph verification.
- **False-positive spike deletion:** preserve or rename the test if it covers unique official SDK behavior.
- **Facade entanglement:** avoid broad edits in Phase 1; if deleting the client unexpectedly requires facade changes, stop and split the work.

Rollback is a focused revert of the deletion commit. No data migration or deployment change is required.

## Open Questions

1. Does the project promise source-level compatibility for the unused `pkg/client` package to external module consumers?
2. Should the custom `pkg/protocol` model remain a stable public API, or should a future major release make the official SDK types canonical throughout the repository?
3. Which official SDK protocol version should be pinned in acceptance evidence as the SDK evolves?
4. Should the standalone `go-go-mcp` CLI remain a supported product, or become an example/application around the official SDK?

## References

- `pkg/client/client.go` — legacy client and hardcoded protocol version.
- `pkg/client/sse.go` — legacy SSE transport and exclusive `r3labs/sse` use.
- `pkg/client/stdio.go` — legacy stdio transport.
- `cmd/go-go-mcp/cmds/client/helpers/client.go` — current official SDK client construction.
- `pkg/embeddable/official_backend.go` — current official SDK server and transport boundary.
- `pkg/embeddable/official_mapping.go` — custom protocol to official SDK mapping.
- `pkg/embeddable/official_sdk_spike_test.go` — migration spike candidate for later cleanup.
- `go.mod` — module dependencies and official SDK version.
