---
Title: Investigation diary
Ticket: 2026-08-31-GO-GO-MCP-CLEANUP
Status: active
Topics:
    - mcp
    - go
    - cleanup
    - dependencies
    - refactoring
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/cmd/go-go-mcp/cmds/client/helpers/client.go
      Note: Evidence of official SDK client usage
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/client/client.go
      Note: Evidence for orphaned legacy client cleanup
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/client/sse.go
      Note: Evidence for the legacy dependency cleanup candidate
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/embeddable/official_backend.go
      Note: Evidence that the custom facade remains active
    - Path: /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/protocol
      Note: Evidence that the compatibility facade remains active and is out of Phase 1 scope
    - Path: repo://pkg/client/stdio.go
      Note: Legacy transport included in the Phase 1 deletion candidate
ExternalSources: []
Summary: Initial investigation created a staged cleanup design and identified the orphaned legacy client as the safest first deletion.
LastUpdated: 2026-08-31T15:10:00-04:00
WhatFor: Chronological record of evidence and decisions for the future cleanup ticket.
WhenToUse: Read before resuming implementation work on this ticket.
---




# Diary

## Goal

Capture the evidence, scope boundaries, and implementation plan for removing pre-official-SDK MCP client code and deprecated dependencies from `go-go-mcp` without accidentally deleting the still-used compatibility facade.

## Step 1: Create cleanup ticket and establish scope

This step created a dedicated future-work ticket in the `go-go-mcp` repository and recorded the first evidence-based cleanup boundary. The investigation distinguishes the orphaned legacy client from the custom protocol and registry facade, which remains connected to the current official SDK backend.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket in go-go-mcp to cleanup old dependencies / deprecated code, add a design doc, but then let's move back to coinvault, that's for the future."

**Assistant interpretation:** Create and document a future cleanup ticket in the actual `go-go-mcp` repository, without implementing the cleanup now.

**Inferred user intent:** Preserve the cleanup decision and scope so the current CoinVault work can continue without losing the opportunity to remove obsolete MCP code later.

### What I did

- Initialized `go-go-mcp/.ttmp.yaml` with a repository-local `ttmp` root.
- Initialized `go-go-mcp/ttmp` vocabulary and docmgr scaffolding.
- Created ticket `2026-08-31-GO-GO-MCP-CLEANUP`.
- Added the design document and this investigation diary.
- Inspected repository imports and implementation paths.
- Identified that `pkg/client` is not imported by repository code and hardcodes MCP `2024-11-05`.
- Identified that the current CLI client uses `github.com/modelcontextprotocol/go-sdk`.
- Identified that `pkg/protocol` and the embeddable registry facade are still used by the official backend adapter.

### Why

The old client is misleading and appears removable, but deleting the entire custom protocol layer would conflate a safe dependency cleanup with a broad API migration. The ticket records a smaller first phase and preserves the larger decision for explicit future review.

### What worked

- The initial accidental ticket creation under the parent CoinVault docmgr root was detected immediately.
- The accidental untracked CoinVault ticket directory was removed.
- A repository-local docmgr configuration was created under `go-go-mcp`.
- The ticket, design document, and diary now live under the intended repository.
- Import inspection provided a concrete deletion candidate and protected active code from premature removal.

### What didn't work

- The first command was run from `go-go-mcp`, but docmgr followed the parent `/home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/.ttmp.yaml` and created the ticket under `coinvault/ttmp`.
- The exact command that exposed the issue was:

  `cd go-go-mcp && docmgr status --summary-only`

  It reported:

  `root=/home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/coinvault/ttmp`

- The accidental directory was removed before the correct ticket was created; no CoinVault code or tracked documentation was modified by that mistake.

### What I learned

- Docmgr discovers the nearest available configuration; nested repositories need their own `.ttmp.yaml` when the parent has a configuration.
- `go-go-mcp/cmd/go-go-mcp/cmds/client/helpers/client.go` already uses the official SDK client, so the legacy `pkg/client` is not needed for the current CLI client.
- `pkg/client/client.go` hardcodes `2024-11-05`, while the official SDK is the intended wire implementation.
- `pkg/embeddable/official_backend.go` still maps active custom registry/protocol types into the official SDK, so `pkg/protocol` is not currently dead code.

### What was tricky to build

The main sharp edge was distinguishing “old MCP implementation” from “old-looking internal API.” The legacy client package is isolated and unreferenced, but the custom protocol types are still the input/output contract for active registries and providers. The design therefore separates deletion of `pkg/client` from any future protocol-facade migration.

### What warrants a second pair of eyes

- Confirm whether external consumers rely on the exported `pkg/client` package despite no in-repository imports.
- Confirm `github.com/r3labs/sse/v2` has no use outside `pkg/client/sse.go` before removing it.
- Confirm `official_sdk_spike_test.go` duplicates existing official backend coverage before deleting it.
- Review whether the project wants source compatibility or a major-version boundary for deleting the legacy package.

### What should be done in the future

- Implement Phase 1: delete `pkg/client`, remove its exclusive SSE dependency, and validate the full module.
- Audit and retire or rename the official SDK spike test.
- Decide separately whether to migrate `pkg/protocol` and the registry facade to official SDK types.

### Code review instructions

- Start with `pkg/client/client.go`, `pkg/client/sse.go`, and `pkg/client/stdio.go`.
- Compare them with `cmd/go-go-mcp/cmds/client/helpers/client.go` and `pkg/embeddable/official_backend.go`.
- Verify imports with `rg 'go-go-mcp/pkg/client' .`.
- Validate dependency removal with `go list -deps ./...`, `go test ./...`, `go vet ./...`, and the repository lint command.

### Technical details

- Ticket: `2026-08-31-GO-GO-MCP-CLEANUP`
- Repository: `/home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp`
- Documentation root: `/home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/ttmp`
- Safe first deletion candidate: `pkg/client/`
- Candidate exclusive dependency: `github.com/r3labs/sse/v2`
- Active official SDK client path: `cmd/go-go-mcp/cmds/client/helpers/client.go`
