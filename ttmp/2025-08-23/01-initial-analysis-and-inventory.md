### Initial Analysis and Inventory: Porting `jesus/` to `go-mcp` and decoupling MCP in `go-go-mcp`

#### Purpose and scope
This document inventories the `jesus/` project, identifies the core functionality to port, outlines the `go-mcp` architecture and extension points we will target, and maps existing MCP protocol implementations (`mcp-go` and `go-go-mcp`) to plan eventual decoupling from protocol specifics in `go-go-mcp`.

---

### jesus/ feature inventory

#### Core runtime and bindings
- Express-like HTTP in JS powered by Goja
  - Engine setup: `jesus/pkg/engine/engine.go`
  - Express API surface: `app.get|post|put|delete|patch`, legacy `registerHandler`
    - Registration: `jesus/pkg/engine/handlers.go` (e.g., `registerHandler`, `appGet`, `appUse`)
    - Request/Response shims: `ExpressRequest`, `ExpressResponse` with `Status`, `Send`, `Json`, `Redirect`, `Set`, `Cookie`, `End`
  - HTTP client bindings: `fetch` and `HTTP.get|post|...` via `setupHTTPBindings`
  - Geppetto JS APIs and database modules loaded at boot

Key refs:
- `jesus/pkg/engine/handlers.go`
- `jesus/pkg/engine/http_bindings.go`
- `jesus/pkg/engine/engine.go`

#### Routing and servers
- JS server router (dynamic endpoints): `web.SetupJSRoutes`
- Admin server router (playground, logs, docs, API): `web.SetupAdminServerRoutes`
- Execute API handler wiring via `web.SetupRoutesWithAPI`
- Dynamic dispatch into JS via `web.HandleDynamicRoute`

Key refs:
- `jesus/pkg/web/routes.go`
- `jesus/pkg/web/router.go`

#### CLI and bootstrapping
- Main entrypoint wires cobra cmds and MCP command: `cmd/jesus/main.go`
- `serve` starts two servers and boots the engine
- `execute` posts JS to `/v1/execute`
- HTTP Execute API handler submits `EvalJob`

Key refs:
- `jesus/cmd/jesus/cmd/serve.go`
- `jesus/cmd/jesus/cmd/execute.go`
- `jesus/pkg/api/execute.go`

#### MCP integration in jesus
- Uses `go-go-mcp` embeddable, exposes `executeJS` tool, runs JS/admin servers on free ports, hooks on server start.
- Tool handler executes JS and returns `go-go-mcp/pkg/protocol.ToolResult`.

Key refs:
- `jesus/pkg/mcp/server.go`

#### Persistence layer
- Repository interfaces and SQLite implementation for executions/history

Key refs:
- `jesus/pkg/repository/interfaces.go`

---

### go-go-mcp architecture overview (target platform)

- Server facade with transport abstraction and request handler dispatch: `pkg/server` (`Server`, `RequestHandler`)
- Transport abstraction and helpers: `pkg/transport` (`Transport` interface)
- Protocol types and `ToolResult`: `pkg/protocol`
- Embeddable integration for cobra apps, tool registry, middleware, hooks: `pkg/embeddable`

Key refs:
- `go-go-mcp/pkg/server/server.go`
- `go-go-mcp/pkg/server/handler.go`
- `go-go-mcp/pkg/transport/transport.go`
- `go-go-mcp/pkg/protocol/tools.go`
- `go-go-mcp/pkg/embeddable/server.go`

Extension points to map `jesus` functionality:
- Tools: implement `WithTool` handlers for `executeJS`, `executeJSFile`, `listRoutes`, `adminLinks`.
- Session: use `session.SessionStore` to tie JS engine session IDs to MCP sessions if needed.
- Transport: reuse stdio/SSE via embeddable without protocol coupling.

---

### mcp-go (mark3labs) inventory for decoupling reference

- Full MCP server with request handler and transports (stdio, SSE, streamable-http)
- Client libs and transport interfaces separate from `go-go-mcp`

Key refs:
- `mcp-go/server/server.go`
- `mcp-go/server/stdio.go`
- `mcp-go/server/streamable_http.go`

Observation: `jesus` already uses `go-go-mcp` embeddable (not `mcp-go`) for its MCP CLI. `go-go-mcp` internally defines protocol and transports, overlapping with `mcp-go`.

---

### Preliminary mapping: `jesus` -> `go-go-mcp`

- JS runtime lifecycle: initialize engine, start dispatcher, start mux servers
  - In MCP mode, already invoked via embeddable hooks in `jesus/pkg/mcp/server.go`.
- Tool surface:
  - `executeJS(code: string) -> ToolResult` exists; candidates to add: `executeJSFile(path)`, `listRoutes()`, `readLogs()`, `health()`.
- HTTP endpoints remain outside MCP; embeddable does not constrain local HTTP servers.

Risks/gaps:
- Engine state vs MCP session: currently per-process; consider per-session isolation or namespacing.
- Duplicate `HTTP` symbol in JS (status constants vs client bindings); ensure no collisions.
- Repository coupling: DB paths via flags; validate dispatcher concurrency.
- Module alignment: ensure single workspace `go.mod` usage consistent with repo rules.

---

### Actionable next steps
- Map `jesus` engine APIs into a stable service interface for MCP tools.
- Add additional MCP tools for observability and control (list routes, reset VM, load script by path/URL).
- Define session strategy (single engine vs per-session sandboxing).
- Outline plan to replace protocol/transport handling in `go-go-mcp` with `mcp-go` equivalents while keeping `embeddable` API stable.
- Prepare migration tasks list for the next doc.

---

Requested output path: `go-go-mcp/ttmp/2025-08-23/01-initial-analysis-and-inventory.md`
