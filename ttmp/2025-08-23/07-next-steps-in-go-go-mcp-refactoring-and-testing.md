### Refactor report: go-go-mcp migration to mcp-go and CLI updates

#### Scope and intent

We decoupled `go-go-mcp`'s internal MCP protocol/server/transport from our codebase and migrated server and client paths to `github.com/mark3labs/mcp-go` (mcp-go). We preserved the embeddable façade and CLI UX, and added developer-focused debug logging.

References:
- `ttmp/2025-08-23/05-decoupling-mcp-go-in-go-go-mcp-refactoring-plan-and-approach.md`
- `ttmp/2025-08-23/06-cli-and-affordances-impact-analysis.md`
- Planning/supporting notes: `ttmp/2025-08-23/0{1,2,3,4}-*.md`
- Integration strategy: `ttmp/2025-08-23/08-top-level-integration-testing-strategy.md`

---

### What changed (by area)

#### 1) Embeddable server rebuilt on mcp-go
- New backend wrapper over `mcp-go/server`:
  - `pkg/embeddable/mcpgo_backend.go`
    - New `Backend` interface and `NewBackend(*ServerConfig)`
    - Creates `mcpserver.NewMCPServer(...)` with `WithToolCapabilities(true)` and `WithLogging()`
    - Registers tools from our `tool-registry` into `mcp-go` via `AddTool`
    - Adapter converts our `protocol.ToolResult` into `mcp.CallToolResult` (`mapToolResultToMCP`)
    - Transports mapped to `mcp-go` equivalents:
      - stdio → `mcpserver.ServeStdio`
      - SSE → `mcpserver.NewSSEServer(server).Start(":PORT")`
      - streamable-http → `mcpserver.NewStreamableHTTPServer(server).Start(":PORT")`
    - Added debug logs (tool registration, calls, selection of transport, transport start address)

- Embeddable command now starts the backend:
  - `pkg/embeddable/command.go` → `startServer()` constructs backend via `NewBackend(config)` and `backend.Start(ctx)`

- CLI server start refactor to embeddable backend:
  - `cmd/go-go-mcp/cmds/server/start.go`
    - Builds a `tool_registry.Registry` proxy around the configured `ToolProvider`
    - Registers discovered tools with proxy handlers into the registry, then injects it into `embeddable` via `WithToolRegistry(reg)`
    - Starts backend (selected transport) and file watcher
    - Added debug logs (transport, tool counts, per-tool registration and invocation)

Symbols of note:
- `embeddable.NewBackend`, `registerToolsFromRegistry`, `mapToolResultToMCP`
- `server/start.go` `Run(...).` path: `toolsList, _, err := toolProvider.ListTools(...)` → register into `reg`

#### 2) Client CLI migrated to mcp-go client
- Client creation and initialize moved to mcp-go:
  - `cmd/go-go-mcp/cmds/client/helpers/client.go`
    - SSE: `mcpclient.NewSSEMCPClient`
    - Streamable HTTP: `mcpclient.NewStreamableHttpClient`
    - Command (stdio subprocess): `mcpclient.NewStdioMCPClient`
    - Start transport (non-stdio) via `c.Start(ctx)` before `Initialize`
    - Initialize with `mcp.InitializeRequest` and `mcp.ClientCapabilities{}`

- Client subcommands now use `mcp-go/mcp` types and support dual-output (human + structured):
  - `cmd/go-go-mcp/cmds/client/tools.go` → list (human + glaze) and call (human + glaze) [updated: dual-mode toggle]
  - `cmd/go-go-mcp/cmds/client/resources.go` → list/read (human + glaze) [updated: dual-mode toggle]
  - `cmd/go-go-mcp/cmds/client/prompts.go` → list/execute (human + glaze) [updated: dual-mode toggle]

Symbols of note:
- `mcp.ListToolsRequest`, `mcp.CallToolRequest`, `mcp.ListResourcesRequest`, `mcp.ReadResourceRequest`, `mcp.ListPromptsRequest`, `mcp.GetPromptRequest`
- Dual-command pattern via `cli.BuildCobraCommand(...)` with `cli.WithDualMode(true)`, `cli.WithGlazeToggleFlag("with-glaze-output")`

#### 3) Removal of legacy internals and examples
- Deleted legacy server, transport, and UI bits:
  - `pkg/server/*` (handlers, responses, SSE bridge, ui/*)
  - `pkg/transport/*` (stdio/sse/streamable_http implementations)
  - `cmd/ui-server/*`
  - Bridge command and usage removed:
    - `cmd/go-go-mcp/cmds/bridge.go` deleted
    - `cmd/go-go-mcp/main.go` no longer adds `NewBridgeCommand`

- Removed examples tied to deleted transports:
  - `examples/batch-cancellation/main.go`
  - `examples/streamable-http/{main.go, README.md}`

Note: We intentionally kept `pkg/protocol` for tool result structures used by our registry and adapter; mapping to `mcp-go` response types now happens at the boundary.

---

### Debug logging added (where to look)
- `pkg/embeddable/mcpgo_backend.go`:
  - Backend creation (name/version/transport/port)
  - Tool registration count and per-tool details
  - Middleware chain size
  - Per-call arguments and completion
  - Transport start address logs (SSE/HTTP)
- `cmd/go-go-mcp/cmds/server/start.go`:
  - Transport/port
  - Server settings (directories, config, internal servers, watch)
  - Tool registration count and per-tool logs; per-invocation logs

Use Viper/log level to enable debug output at runtime.

---

### Current issues and observations

1) Local SSE and StreamableHTTP base paths
- Finding: StreamableHTTP default endpoint is `/mcp`; SSE endpoint is `/sse` (client expects to be pointed to the SSE endpoint, server emits a `message` endpoint)
- Fix: Use `--server http://localhost:3001/mcp` for HTTP and `--server http://localhost:3002/sse` for SSE; both verified working

2) Dual-command UX toggle
- Implemented: `--with-glaze-output` now toggles structured output for tools/resources/prompts
- Example: `... tools list --with-glaze-output --output json`

3) Result mapping coverage
- Status unchanged: `text`, `image`, `resource` mapped; JSON remains `TextContent`
- Consider `StructuredContent` later if needed

---

### Completed tasks (checked off)

- [x] Stand up streamable-http server in tmux and verify client list/call (used `/mcp` path)
- [x] Bring up SSE server and verify client list/call (used `/sse` path)
- [x] Add explicit dual-command toggle flag to client commands (`--with-glaze-output`)
- [x] Tighten logging and visibility (transport start address logs added)
- [x] Remove temporary in-process integration test in favor of top-level tests (`pkg/embeddable/integration_http_test.go` deleted)

---

### Next steps for integration testing (top-level)

See `ttmp/2025-08-23/08-top-level-integration-testing-strategy.md`.

Actionable follow-ups:
- [ ] Add `scripts/integration/http_smoke.sh` and `scripts/integration/sse_smoke.sh` using tmux
- [ ] Makefile targets `integration-http` and `integration-sse`
- [ ] Optional Dockerfile and CI job to run HTTP smoke test
- [ ] Add sqlite round-trip (open + query) and snapshot output for fetch
- [ ] Document `--with-glaze-output` usage in README and help

---

### File reference map (key edits)

- Backend/server:
  - `pkg/embeddable/mcpgo_backend.go` (backend over mcp-go, tool/result mapping, debug logs, transport start logs)
  - `pkg/embeddable/command.go` (backend wiring)
  - `cmd/go-go-mcp/cmds/server/start.go` (tool provider proxy → registry, backend start, logs)
  - `cmd/go-go-mcp/main.go` (removed bridge wiring)

- Client:
  - `cmd/go-go-mcp/cmds/client/helpers/client.go` (mcp-go client creation/start/initialize)
  - `cmd/go-go-mcp/cmds/client/tools.go` (dual mode + glaze toggle)
  - `cmd/go-go-mcp/cmds/client/resources.go` (dual mode + glaze toggle)
  - `cmd/go-go-mcp/cmds/client/prompts.go` (dual mode + glaze toggle)

- Removals:
  - `pkg/server/*`, `pkg/transport/*`, `cmd/ui-server/*`, bridge command, legacy examples
  - `pkg/embeddable/integration_http_test.go` (replaced by top-level integration strategy)

- Design docs:
  - `ttmp/2025-08-23/05-decoupling-mcp-go-in-go-go-mcp-refactoring-plan-and-approach.md`
  - `ttmp/2025-08-23/06-cli-and-affordances-impact-analysis.md`
  - Additional: `01-initial-analysis-and-inventory.md`, `02-migration-planning.md`, `03-implementation-progress.md`, `04-mcp-protocol-decoupling.md`

---

### tmux notes (dev convenience)
- Start session (example):
  - `tmux new-session -d -s mcp_http 'cd .../go-go-mcp && go run ./cmd/go-go-mcp server start --transport streamable_http --port 3001 --internal-servers sqlite,fetch,echo'`
- Attach/detach/kill:
  - `tmux attach -t mcp_http`
  - `Ctrl-b d`
  - `tmux kill-session -t mcp_http`
