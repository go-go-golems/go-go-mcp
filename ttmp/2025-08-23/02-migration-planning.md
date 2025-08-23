### Migration Plan: Porting `jesus/` to `go-mcp` and setting up for MCP protocol decoupling

#### Purpose and scope
Define a stepwise plan to port `jesus` features to `go-mcp` (specifically the embeddable server APIs), minimize protocol coupling, and prepare `go-go-mcp` to replace internal protocol/transport logic with `mcp-go` later.

---

### Reasoning walkthrough

- Inventory outcome: `jesus` already integrates `go-go-mcp` embeddable, exposes `executeJS`, and runs its own HTTP servers under MCP lifecycle hooks. Most porting is about stabilizing service boundaries and exposing richer tools, not reworking transports.
- `go-go-mcp` embeddable `WithTool` is the primary integration surface; ToolResult formats match `go-go-mcp/pkg/protocol`.
- `mcp-go` defines richer `CallToolResult` and typed request helpers. Future decoupling should provide adapters to/from `go-go-mcp` `protocol.ToolResult`.

---

### Stepwise plan (high-level)

1) Stabilize jesus engine service boundary
- Define a Go interface `JSEngineService` encapsulating: execute code, load file, list routes, reset VM, fetch logs, health.
- Provide an adapter over `engine.Engine` to implement `JSEngineService`.

2) Expand MCP tools in jesus using embeddable
- Add tools: `execute_js_file`, `list_routes`, `reset_vm`, `read_logs`, `health`, `load_scripts_dir`.
- Use `embeddable.WithTool` with simple schemas; return `protocol.ToolResult`.

3) Optional: resources and prompts
- Register resources for admin links and docs; prompts for preset code snippets.
- Keep in jesus initially; later we can move generic parts into `go-go-mcp` examples.

4) Session strategy
- Start with single engine; tag results with MCP session ID from context.
- Evaluate per-session sandboxes later using `session.SessionStore`.

5) Logging/observability
- Ensure admin endpoints remain live; consider adding `notifications/logging/message` via `go-go-mcp` if available.

6) Prepare decoupling of protocol logic in go-go-mcp
- Introduce narrow interfaces in `go-go-mcp` around: protocol types, transports, and server handler mapping.
- Add adapter layer to convert between `go-go-mcp/pkg/protocol.ToolResult` and `mcp-go/mcp.CallToolResult`.
- Gate transport creation behind interfaces so `mcp-go` transports can be swapped in without changing embeddable API.

---

### Detailed tasks

- Interfaces
  - Create `JSEngineService` in `jesus/pkg/mcp/` (or `pkg/service/`) with methods: `Execute(code)`, `ExecuteFile(path)`, `ListRoutes()`, `Reset()`, `ReadLogs(limit)`, `Health()`.
  - Implement `JSEngineAdapter` wrapping `*engine.Engine`.

- Tools (in `jesus/pkg/mcp/server.go` via `embeddable.WithTool`)
  - `executeJS` (exists): verify schema and add examples.
  - `executeJSFile(absolutePath)`.
  - `listRoutes()` -> returns JSON array of `{method,path}` from engine state.
  - `resetVM()` -> reinitialize engine; ensure dispatcher running.
  - `readLogs(limit:int=100)` -> tail request logs.
  - `health()` -> `{ok:true, jsBaseURL, adminBaseURL}`.

- Wiring and flags
  - Ensure JS/admin ports and DB paths are pulled from `GetCommandFlags(ctx)` in startup hook.
  - Confirm scripts dir handling is configurable by MCP flags if needed.

- Adapters for decoupling (`go-go-mcp` side)
  - Add `adapter/protocol` with mappers: `ToMark3CallToolResult`, `FromMark3CallToolResult`.
  - Abstract transport constructors behind small factory interface.

- Compatibility checks
  - Verify `ToolResult` text/json mapping stays consistent with clients.
  - Confirm SSE/stdio start flags: `--transport sse|stdio|streamable_http`, `--port`.

---

### Risks and mitigations
- Engine concurrency and resets: guard with locks; ensure dispatcher lifecycle survives resets.
- Session isolation: start with shared state; document risk; add session-aware sandbox later.
- Protocol differences: maintain adapters and integration tests for result conversion.

---

### Next steps
- Implement `JSEngineService` adapter and register additional tools in `jesus`.
- Add protocol adapter scaffolding in `go-go-mcp` (no behavior change yet).
- Document testing matrix (stdio, sse, streamable-http) and sample invocations.

---

Requested output path: `go-go-mcp/ttmp/2025-08-23/02-migration-planning.md`
