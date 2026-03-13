---
Title: Diary
Ticket: JESUS-001-RUNTIME-MCP-ANALYSIS
Status: active
Topics:
    - go
    - javascript
    - mcp
    - review
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-goja/engine/factory.go
      Note: Runtime factory evidence inspected during analysis
    - Path: go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: Embeddable backend evidence inspected during analysis
    - Path: jesus/pkg/engine/engine.go
      Note: Primary runtime evidence inspected during analysis
    - Path: jesus/pkg/mcp/server.go
      Note: Primary MCP evidence inspected during analysis
ExternalSources: []
Summary: Chronological investigation diary for the jesus runtime and MCP opportunity assessment, including commands run and the resulting documentation deliverables.
LastUpdated: 2026-03-12T23:37:08.872565949-04:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Capture how the analysis was produced, which files and commands informed it, and how to review or continue the work later.

## Step 1: Analyze jesus runtime and MCP evolution opportunities

This step captured an evidence-backed architecture assessment rather than code changes. The goal was to answer whether `jesus` should adopt newer `go-go-goja` runtime features and newer `go-go-mcp` capabilities, then package the result into a reusable ticket deliverable.

The work focused on identifying leverage, not proving every possible migration path. The strongest conclusion was that `go-go-goja` runtime ownership and composition are immediately valuable, while most `go-go-mcp` improvements are incremental and should follow the engine refactor.

### Prompt Context

**User prompt (verbatim):** "Ok, let's analyze if jesus could benefit from some of the new features in go-go-goja (runtime engine, etc...) and from potential updates in go-go-mcp ?"

**Assistant interpretation:** Review the current `jesus` integration points against the current local `go-go-goja` and `go-go-mcp` codebases, and determine which new features are actually worth adopting.

**Inferred user intent:** Build a migration strategy that improves the `jesus` architecture and feature set without doing speculative or low-value modernization work.

**Commit (code):** N/A

### What I did
- Mapped `jesus` runtime, REPL, API, and MCP entrypoints with `rg` and `sed`.
- Read the current `go-go-goja` factory, runtime-owner, module-spec, and module-root code.
- Read the current `go-go-mcp` embeddable server, backend, enhanced tool, prompt, and resource code.
- Compared the current `jesus` usage model against newer runtime and protocol capabilities.
- Wrote the primary design doc with phased recommendations and file-backed references.

### Why
- The question only matters if the recommendations are tied to current code, not memory or generic ideas.
- `jesus` has both runtime concerns and MCP concerns, so the analysis had to separate architectural bottlenecks from optional protocol improvements.

### What worked
- The sibling repositories are present locally, so the assessment could stay entirely evidence-based.
- The newer `go-go-goja` API is concrete enough to map directly onto `jesus/pkg/engine`.
- The newer `go-go-mcp` embeddable layer already exposes enough enhanced tool functionality to recommend practical MCP surface improvements.

### What didn't work
- `sed -n '1,260p' go-go-mcp/pkg/embeddable/config.go`

  Exact error:

  ```text
  sed: can't read go-go-mcp/pkg/embeddable/config.go: No such file or directory
  ```

- Resolution: the relevant configuration lived in `go-go-mcp/pkg/embeddable/server.go` instead.

### What I learned
- `jesus` still treats the runtime as an owned raw `*goja.Runtime` even after starting to adopt `go-go-goja` in the REPL.
- The strongest `go-go-goja` benefit is not “more modules”; it is the explicit ownership model around runtime access.
- `go-go-mcp` has good enhanced tool ergonomics today, but prompts/resources are only worth adding after deciding whether `jesus` wants those surfaces as product features.

### What was tricky to build
- The hard part was distinguishing “library capability exists” from “this is worth integrating in jesus now”.
- `go-go-mcp` exposes more protocol features than `jesus` currently uses, but not all of them are first-order needs. The analysis had to avoid recommending prompts/resources/auth just because they exist.
- `jesus` already has a custom dispatcher, so it would be easy to overstate the value of the new runtime owner unless the concurrency and lifecycle overlap were made explicit.

### What warrants a second pair of eyes
- Whether `jesus` truly requires one long-lived shared JS runtime with persistent state across every execution path.
- Whether the custom dispatcher should be removed entirely or retained as a higher-level execution-log/persistence wrapper around runtime-owner calls.
- Whether remote/authenticated MCP is actually in scope for `jesus`, since that changes the priority of streamable HTTP and OIDC features.

### What should be done in the future
- Phase 1 implementation plan for migrating `jesus/pkg/engine` onto `go-go-goja` runtime ownership.
- Follow-up decision on whether `jesus` should expose resources for docs, scripts, and execution history.

### Code review instructions
- Start with the primary design doc: `design-doc/01-jesus-runtime-and-mcp-evolution-analysis.md`.
- Then review the current runtime evidence in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/engine.go` and `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/dispatcher.go`.
- Compare that with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go` and `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/runtimeowner/runner.go`.
- For MCP follow-ups, review `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go` against `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/enhanced_tools.go` and `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go`.

### Technical details
- Key commands run:

```bash
rg -n "go-go-goja|go-go-mcp|mcp|NewEngine\\(|ExecuteScript|StartDispatcher|setupBindings|Require|goja|tool" jesus -g '*.go'
rg -n "type Runtime|FactoryBuilder|NewBuilder|RuntimeInitializer|DefaultRegistryModules|ModuleSpec|runtimeowner" go-go-goja -g '*.go'
rg -n "type Server|Tool|resources|prompts|sampling|stdio|sse|streamable|AddMCPCommand|CallTool" go-go-mcp jesus/pkg/mcp -g '*.go'
nl -ba jesus/pkg/engine/engine.go | sed -n '1,180p'
nl -ba go-go-goja/engine/factory.go | sed -n '1,220p'
nl -ba go-go-mcp/pkg/embeddable/mcpgo_backend.go | sed -n '1,260p'
```

- Primary deliverable:
  - `design-doc/01-jesus-runtime-and-mcp-evolution-analysis.md`

## Related

- `../design-doc/01-jesus-runtime-and-mcp-evolution-analysis.md`
