---
Title: jesus Runtime and MCP Evolution Analysis
Ticket: JESUS-001-RUNTIME-MCP-ANALYSIS
Status: active
Topics:
    - go
    - javascript
    - mcp
    - review
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-goja/engine/factory.go
      Note: Target runtime factory model
    - Path: go-go-goja/engine/module_roots.go
      Note: Script-relative require roots for future adoption
    - Path: go-go-goja/pkg/runtimeowner/runner.go
      Note: Owner scheduling semantics
    - Path: go-go-mcp/pkg/embeddable/enhanced_tools.go
      Note: Typed MCP tool upgrade path
    - Path: go-go-mcp/pkg/prompts/registry.go
      Note: Prompt surface available for second-phase work
    - Path: go-go-mcp/pkg/resources/registry.go
      Note: Resource surface available for second-phase work
    - Path: jesus/pkg/api/execute.go
      Note: Current API execution entrypoint
    - Path: jesus/pkg/engine/dispatcher.go
      Note: Evidence for queue-based serialization
    - Path: jesus/pkg/engine/engine.go
      Note: Evidence for current runtime construction and direct runtime access
    - Path: jesus/pkg/mcp/server.go
      Note: Current MCP tool and server behavior
    - Path: jesus/pkg/repl/model.go
      Note: Partial adoption of go-go-goja runtime in REPL
ExternalSources: []
Summary: Evidence-backed assessment of where jesus should adopt newer go-go-goja runtime features and go-go-mcp server capabilities, ordered by leverage and migration risk.
LastUpdated: 2026-03-12T23:37:08.934957564-04:00
WhatFor: ""
WhenToUse: ""
---


# jesus Runtime and MCP Evolution Analysis

## Executive Summary

`jesus` would benefit materially from the newer `go-go-goja` runtime engine APIs, and only selectively from newer `go-go-mcp` features.

The highest-value change is to replace the hand-rolled `goja` runtime lifecycle in `jesus/pkg/engine` with the factory-owned runtime, explicit module registration, and runtime-owner scheduling now provided by `go-go-goja`. That change addresses a real architectural mismatch in the current code: `jesus` exposes multiple execution paths into a shared runtime, but its concurrency contract is implicit and enforced only by convention.

`go-go-mcp` is already modern enough that `jesus` can improve its MCP surface without changing much server wiring. The most useful near-term upgrades are enhanced typed tools and better schema/annotations for existing MCP handlers. Prompts, resources, and remote-authenticated transports are viable second-phase work, but they are not as immediately valuable as the runtime refactor.

## Problem Statement

The current `jesus` runtime layer predates the newer ownership and composition model in `go-go-goja`, and the current MCP integration uses only the thinnest subset of `go-go-mcp`.

Observed issues:

1. `jesus` builds and owns a raw `*goja.Runtime`, raw event loop, and custom job queue in `jesus/pkg/engine/engine.go` and `jesus/pkg/engine/dispatcher.go`.
2. Runtime access rules are not encoded in the engine API. `SubmitJob` serializes some work through the dispatcher, but `executeCode`, `executeCodeWithResult`, bootstrap execution, handler invocation, and REPL execution still work directly on the runtime.
3. Script execution does not use script-relative module root resolution, so the `require()` environment is flatter and less predictable than the newer `go-go-goja` model.
4. MCP uses one generic `executeJS` tool with loose `map[string]interface{}` arguments, so clients get a weak schema and little behavioral metadata.
5. `jesus` does not expose prompts or resources over MCP even though its docs, execution history, and saved scripts are natural candidates for those protocol surfaces.

Scope of this analysis:

1. Identify where `jesus` can directly benefit from current `go-go-goja` features.
2. Identify which newer `go-go-mcp` capabilities are worth adopting now versus later.
3. Recommend an implementation order that improves architecture without forcing a large rewrite.

## Current-State Analysis

### 1. jesus still owns its own runtime lifecycle

The engine constructor in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/engine.go` creates:

1. a raw `eventloop.EventLoop`,
2. a raw `goja.Runtime`,
3. a raw `require.Registry`,
4. a global module registry bridge,
5. a custom `jobs` channel for serialized evaluation.

Relevant code:

- `jesus/pkg/engine/engine.go:60-131`
- `jesus/pkg/engine/engine.go:285-359`
- `jesus/pkg/engine/dispatcher.go:15-205`

This design is workable, but it duplicates responsibilities that `go-go-goja` now centralizes:

- factory build-time validation,
- explicit module registration,
- runtime-scoped initialization hooks,
- owned runtime lifecycle,
- owner-aware scheduling for safe cross-goroutine access.

### 2. go-go-goja now provides the missing ownership model

The newer runtime stack in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go` and `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/runtimeowner/runner.go` introduces:

1. `FactoryBuilder` for explicit runtime composition,
2. `ModuleSpec` and `RuntimeInitializer` for deterministic setup,
3. `Runtime` as an owned object with `VM`, `Require`, `Loop`, and `Owner`,
4. `runtimeowner.Runner` with `Call` and `Post` for cross-goroutine execution.

Relevant code:

- `go-go-goja/engine/factory.go:31-180`
- `go-go-goja/engine/module_specs.go:14-82`
- `go-go-goja/pkg/runtimeowner/runner.go:62-159`

This is not just a nicer API. It is a stronger concurrency boundary for a long-lived shared runtime.

### 3. jesus only partially adopted the new REPL runtime

The REPL in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/repl/model.go:38-83` now creates a `go-go-goja` factory and runtime, but then drops back to the bare `runtime.VM`.

That means:

1. the REPL gets the module registry,
2. but it does not retain runtime ownership/lifecycle,
3. and it still operates as if the runtime were just a plain `*goja.Runtime`.

This is a useful signal: the repo already wants the new runtime model, but the integration is incomplete.

### 4. Script loading misses new module-root resolution

`jesus` bootstrap and script execution currently call `RunString` on code loaded from files, but the runtime construction does not derive module search roots from the script path.

Relevant code:

- `jesus/pkg/engine/engine.go:151-214`
- `jesus/pkg/mcp/server.go:330-357`

`go-go-goja` now provides:

- `ResolveModuleRootsFromScript`
- `RequireOptionWithModuleRootsFromScript`
- `WithModuleRootsFromScript`

Relevant code:

- `go-go-goja/engine/module_roots.go:11-119`

That is directly useful for `bootstrap.js`, `run-scripts`, and any saved MCP script file.

### 5. MCP uses the minimal embeddable tool path

`jesus` registers a single MCP tool via `embeddable.WithTool(...)` in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go:117-171`.

That path is perfectly valid, but it does not use:

1. enhanced typed argument access,
2. annotations like read-only/idempotent/destructive hints,
3. prompts,
4. resources,
5. streamable HTTP/auth capabilities for remote deployments.

Relevant `go-go-mcp` capabilities:

- `go-go-mcp/pkg/embeddable/server.go:20-220`
- `go-go-mcp/pkg/embeddable/command.go:24-254`
- `go-go-mcp/pkg/embeddable/enhanced_tools.go:27-220`
- `go-go-mcp/pkg/embeddable/mcpgo_backend.go:26-260`
- `go-go-mcp/pkg/prompts/registry.go:13-133`
- `go-go-mcp/pkg/resources/registry.go:12-166`

## Gap Analysis

### Gap 1: Runtime safety is implicit instead of structural

Today `jesus` has a queue (`jobs`) and dispatcher, but the real invariant is "do not touch the runtime concurrently". That invariant is not embodied in the public engine API.

Why this matters:

1. It becomes easy to introduce accidental direct runtime access from new features.
2. The shutdown story remains partial because the queue, event loop, runtime, and module state are not owned by one object.
3. It is hard to reason about which calls are safe from arbitrary goroutines.

### Gap 2: Runtime composition is engine-specific instead of declarative

The current `NewEngine` mixes:

1. runtime creation,
2. module registration,
3. database configuration,
4. repository initialization,
5. bindings setup,
6. bootstrap behavior.

This makes the engine harder to test and harder to vary for REPL, server, and MCP use cases.

### Gap 3: Script ergonomics are weaker than the new runtime supports

`jesus` saves MCP-executed code into `scripts/` and can execute files, but it does not use file-relative module roots. That means scripts cannot reliably carry a local module layout model without extra global setup.

### Gap 4: MCP exposes behavior, not structure

The current tool is understandable to a human, but not especially rich for MCP clients:

1. one tool does several things,
2. arguments are generic,
3. there are no tool behavior hints,
4. docs/history/scripts are not surfaced as prompts/resources.

## Proposed Solution

### Phase 1: Refactor jesus/pkg/engine onto go-go-goja Runtime

Introduce a `go-go-goja`-backed engine core that holds:

1. `*ggjengine.Runtime`,
2. `runtime.Owner`,
3. repository manager,
4. request logger,
5. route/file registries,
6. step settings.

Key rule: all runtime interactions go through the owner.

Sketch:

```go
type Engine struct {
    runtime *ggjengine.Runtime
    repos repository.RepositoryManager
    reqLogger *RequestLogger

    mu sync.RWMutex
    handlers map[string]map[string]*HandlerInfo
    files map[string]goja.Callable
    stepSettings *settings.StepSettings
}

func NewEngine(cfg Config) (*Engine, error) {
    builder := ggjengine.NewBuilder().
        WithModules(ggjengine.DefaultRegistryModules()).
        WithRuntimeInitializers(
            newDatabaseInitializer(cfg.AppDBPath),
            newBindingsInitializer(...),
        )

    factory, err := builder.Build()
    if err != nil { return nil, err }

    runtime, err := factory.NewRuntime(context.Background())
    if err != nil { return nil, err }

    return &Engine{runtime: runtime, ...}, nil
}
```

Then replace direct runtime execution with:

```go
func (e *Engine) Eval(ctx context.Context, code string) (*EvalResult, error) {
    value, err := e.runtime.Owner.Call(ctx, "eval", func(ctx context.Context, vm *goja.Runtime) (any, error) {
        return vm.RunString(code)
    })
    ...
}
```

### Phase 2: Replace the dispatcher queue with owner-routed execution

Once the runtime owner exists, the custom dispatcher in `jesus/pkg/engine/dispatcher.go` becomes optional.

Recommended direction:

1. keep a small evaluation API that handles logging/persistence/session metadata,
2. remove the exported "submit arbitrary job to shared channel" model,
3. use `Owner.Call` for synchronous eval and `Owner.Post` for handler-side async work when needed.

This reduces custom concurrency code and makes cancellation semantics explicit through `context.Context`.

### Phase 3: Add per-script module roots

For bootstrap and file-based execution:

1. resolve module roots from the script file,
2. build a runtime with `WithModuleRootsFromScript(...)`,
3. or maintain separate runtime factories for file-backed versus in-memory execution.

Recommendation:

- Use a file-backed execution path for `run-scripts` and MCP file execution.
- Keep ad hoc inline eval on the base runtime.

This gives scripts a predictable local `require()` story without making inline eval more complex.

### Phase 4: Upgrade MCP tools to enhanced typed tools

Move `executeJS` from `embeddable.WithTool` to `embeddable.WithEnhancedTool`.

Benefits:

1. better schema generation,
2. easier argument validation,
3. tool behavior hints for clients,
4. easier future tool splitting.

Recommended tool split:

1. `execute_code`
2. `execute_file`
3. `list_routes`
4. `read_execution_history`
5. `read_runtime_docs`

Suggested annotations:

- `execute_code`: destructive false, idempotent false, open-world false
- `execute_file`: destructive false, idempotent false
- `list_routes`: read-only true, idempotent true
- `read_execution_history`: read-only true, idempotent true

### Phase 5: Add MCP resources before prompts

Resources are the easiest next MCP surface because `jesus` already has stable read-only artifacts:

1. JavaScript API docs,
2. execution history,
3. saved scripts,
4. possibly current route inventory.

Example resource set:

1. `jesus://docs/javascript-api`
2. `jesus://executions/recent`
3. `jesus://scripts/{name}`
4. `jesus://routes`

Prompts are useful later, but they require a stronger opinion about user workflows. Resources have a clearer mapping to existing `jesus` data.

## Design Decisions

### Decision 1: Prioritize go-go-goja adoption before expanding MCP surface

Rationale:

The runtime layer is the architectural bottleneck. Improving MCP shape without fixing runtime ownership leaves the hardest correctness problem untouched.

### Decision 2: Prefer owner-routed execution over maintaining the custom dispatcher

Rationale:

`runtimeowner.Runner` already gives `Call`/`Post` semantics with cancellation and panic recovery. Keeping both the custom dispatcher and owner scheduling would duplicate control planes.

### Decision 3: Treat prompts/resources as second-phase protocol improvements

Rationale:

They are valuable, but they do not fix the engine architecture. Resources are lower-risk than prompts and should come first.

## Alternatives Considered

### Alternative A: Keep the current engine and only improve MCP tool schemas

Rejected because it improves client ergonomics but not runtime correctness or maintainability.

### Alternative B: Replace only the REPL runtime

Rejected because the REPL is not the architectural center of the product. It would improve one leaf command and leave server/MCP behavior unchanged.

### Alternative C: Adopt prompts/resources first

Rejected for now because the current embeddable usage in `jesus` is tool-only, and prompts/resources add more surface area before the runtime layer is cleaned up.

## Implementation Plan

### Phase 1: Introduce a go-go-goja-backed engine core

Files:

1. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/engine.go`
2. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/dispatcher.go`
3. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/bindings.go`

Steps:

1. Add an internal runtime wrapper field for `*ggjengine.Runtime`.
2. Move module and runtime setup into factory builder code.
3. Move DB setup and JS binding setup into runtime initializers or explicit post-runtime helpers.
4. Replace direct `e.rt.RunString(...)` calls with owner-mediated execution.

### Phase 2: Simplify public execution APIs

Files:

1. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/api/execute.go`
2. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go`

Steps:

1. Replace ad hoc `Done`/`Result` channel choreography with a simpler `Eval` or `EvalWithMetadata` API.
2. Preserve execution history persistence and request logging inside the engine layer.
3. Remove the need for callers to manufacture `EvalJob` directly.

### Phase 3: Add file-relative module roots

Files:

1. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/engine.go`
2. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/cmd/jesus/cmd/run_scripts.go`
3. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go`

Steps:

1. Create a file-execution helper that accepts `scriptPath`.
2. Resolve module roots with `go-go-goja/engine/module_roots.go`.
3. Use file-aware runtime/factory options for bootstrap and script execution.

### Phase 4: Improve MCP schemas and surfaces

Files:

1. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go`

Steps:

1. Convert `executeJS` to `WithEnhancedTool`.
2. Split read-only and mutating operations into separate tools.
3. Add tool annotations for client guidance.
4. If remote serving matters, enable `streamable_http` and auth flags in the operational docs.

### Phase 5: Add resources

Files:

1. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go`
2. new resource registration helpers under `jesus/pkg/mcp/`

Steps:

1. Build a small `resources.Registry`.
2. Expose docs, routes, scripts, and execution-history summaries as resources.
3. Add change notifications only where the data is naturally refreshable.

## Testing Strategy

1. Add engine tests that verify runtime access always happens through owner-mediated APIs.
2. Add cancellation tests for long-running eval.
3. Add module resolution tests for script-relative `require()` behavior.
4. Add MCP integration tests for:
   - tool listing,
   - tool schema shape,
   - `execute_code`,
   - `execute_file`,
   - resource listing and reading.
5. Keep `go test ./...` green in `jesus`, and add focused tests around runtime initialization and shutdown.

## Risks and Open Questions

### Risks

1. The current engine mixes route registration and runtime state heavily; extracting a new runtime core may surface hidden assumptions.
2. File-relative module roots may change script resolution behavior for users who implicitly relied on the flatter global module environment.
3. The Geppetto JS surface in `jesus` is currently incomplete, so a runtime refactor should avoid promising feature parity that does not exist yet.

### Open Questions

1. Should `jesus` keep one long-lived shared runtime, or move some execution modes to isolated per-script runtimes?
2. Is preserving persistent in-memory JavaScript state across all MCP and HTTP evaluations a hard product requirement?
3. Does `jesus` actually need prompts, or are resources plus better tools sufficient?
4. Is remote/authenticated MCP deployment a real target, or is `jesus` primarily a local development tool?

## References

### Key jesus files

1. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/engine.go`
2. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/dispatcher.go`
3. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/repl/model.go`
4. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go`
5. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/api/execute.go`

### Key go-go-goja files

1. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go`
2. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/module_specs.go`
3. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/module_roots.go`
4. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/runtimeowner/runner.go`

### Key go-go-mcp files

1. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/server.go`
2. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/command.go`
3. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/enhanced_tools.go`
4. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go`
5. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/prompts/registry.go`
6. `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/resources/registry.go`

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
