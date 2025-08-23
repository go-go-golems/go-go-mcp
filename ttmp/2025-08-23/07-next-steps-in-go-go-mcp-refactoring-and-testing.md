### Refactor report: go-go-mcp migration to mcp-go and CLI updates

#### Scope and intent

We decoupled `go-go-mcp`'s internal MCP protocol/server/transport from our codebase and migrated server and client paths to `github.com/mark3labs/mcp-go` (mcp-go). We preserved the embeddable façade and CLI UX, and added developer-focused debug logging.

References:
- `ttmp/2025-08-23/05-decoupling-mcp-go-in-go-go-mcp-refactoring-plan-and-approach.md`
- `ttmp/2025-08-23/06-cli-and-affordances-impact-analysis.md`
- Planning/supporting notes: `ttmp/2025-08-23/0{1,2,3,4}-*.md`

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
    - Added debug logs (tool registration, calls, selection of transport)

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
  - `cmd/go-go-mcp/cmds/client/tools.go` → list (human + glaze) and call (human + glaze)
  - `cmd/go-go-mcp/cmds/client/resources.go` → list/read (human + glaze)
  - `cmd/go-go-mcp/cmds/client/prompts.go` → list/execute (human + glaze)

Symbols of note:
- `mcp.ListToolsRequest`, `mcp.CallToolRequest`, `mcp.ListResourcesRequest`, `mcp.ReadResourceRequest`, `mcp.ListPromptsRequest`, `mcp.GetPromptRequest`
- Dual-command pattern via `cli.BuildCobraCommand(...)` and implementing both `RunIntoWriter` (human default) and `RunIntoGlazeProcessor` (structured)

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
- `cmd/go-go-mcp/cmds/server/start.go`:
  - Transport/port
  - Server settings (directories, config, internal servers, watch)
  - Tool registration count and per-tool logs; per-invocation logs

Use Viper/log level to enable debug output at runtime.

---

### Current issues and observations

1) Local SSE endpoint returns 404
- Symptom:
  - `go run ./cmd/go-go-mcp client tools list --transport sse --server http://localhost:3001` → `failed to start client transport: unexpected status code: 404`
- Likely causes:
  - SSE server base-path mismatch. `mcp-go` SSE server supports base paths/options (`WithBaseURL`, etc.). Client default endpoint may differ from server default.
- Hints:
  - Check `mcp-go/client/transport/sse.go` for default paths used
  - Configure server with `server.WithBaseURL` (SSE option) or align client to the server's expected path via client options
  - For local testing, prefer `streamable_http` first (HTTP POST is simpler), then bring up SSE

2) Local streamable-http server not started when tested
- Symptom:
  - Connect refused on `http://localhost:3001` when attempting streamable_http
- Likely causes:
  - We started SSE server in tmux, not streamable-http
- Hints:
  - Start `streamable_http` transport in tmux and test again
  - Confirm port and that POST to `http://localhost:3001` is open

3) Dual-command UX toggle
- Status:
  - We implemented both Writer and Glaze interfaces; by default, human output is produced via `RunIntoWriter`
- Next improvement:
  - Expose a conventional flag to force structured output (e.g. `--with-glaze-output`) using Glazed dual-command options (see `commands-reference.md`, Dual Command Builder)

4) Tool registration path parity
- Status:
  - `server/start.go` proxies `ToolProvider` into `tool-registry.Registry` and injects into `embeddable` backend
  - This keeps embeddable mappings consistent with our registry semantics
- Next improvement:
  - If we later add prompt/resource providers, add similar adapters to `embeddable` (optional for now)

5) Result mapping coverage
- Status:
  - We map `text`, `image`, and `resource` to `mcp` equivalents; JSON is emitted as `TextContent` with `application/json` per our current convention
- Next improvement:
  - Consider using `StructuredContent` for true structured returns where beneficial

---

### Concrete next steps for the next developer

1) Stand up streamable-http server in tmux and verify client list/call
- Start server:
  - `tmux new-session -d -s mcp_http 'cd /home/manuel/workspaces/2025-08-20/migrate-jesus-go-mcp/go-go-mcp && go run ./cmd/go-go-mcp server start --transport streamable_http --port 3001 --internal-servers sqlite,fetch,echo'`
- Verify:
  - `go run ./cmd/go-go-mcp client tools list --transport streamable_http --server http://localhost:3001`
  - Call a simple tool (e.g. `echo`):
    - `go run ./cmd/go-go-mcp client tools call echo --transport streamable_http --server http://localhost:3001 --args message=hello`
- If successful, repeat with glaze (structured) output by selecting a structured output format via Glazed (e.g., `--output json`) if exposed

2) Fix SSE 404 mismatch
- Inspect defaults:
  - `mcp-go/server/sse.go` for default base path and server options
  - `mcp-go/client/transport/sse.go` for client path expectations
- Apply server options:
  - In `pkg/embeddable/mcpgo_backend.go` for SSE transport construction, pass appropriate options (e.g., base URL/path) if client expects `/mcp` or similar
- Re-test:
  - `tmux new-session -d -s mcp_sse '... --transport sse --port 3001 ...'`
  - `go run ./cmd/go-go-mcp client tools list --transport sse --server http://localhost:3001`

3) Add explicit dual-command toggle flag
- In client commands builders (`tools.go`, `resources.go`, `prompts.go`), switch to Glazed dual toggles per `commands-reference.md`:
  - Example pattern:
    - `cobraCmd, err := cli.BuildCobraCommand(myDualCmd, cli.WithDualMode(true), cli.WithGlazeToggleFlag("with-glaze-output"))`
  - Document usage in help strings

4) Tighten logging and visibility
- Ensure debug level is set in dev runs to see:
  - Tool registration counts and names (`pkg/embeddable/mcpgo_backend.go`)
  - Per-call logs
  - Server settings logs (`cmd/go-go-mcp/cmds/server/start.go`)

5) Expand embeddable adapters (optional)
- If needed, add prompt/resource registration adapters mirroring tool adapter into `mcp-go` (currently tool path is covered)

6) Tests
- Add basic integration tests that:
  - Start a streamable-http server on an ephemeral port
  - Use the mcp-go client to `initialize`, `tools/list`, `tools/call` against a simple test tool
- Add golden tests for CLI human output (now dual-command) for stability

7) Docs
- Update README and `pkg/doc` to reflect new client/server usage and dual-output patterns

---

### File reference map (key edits)

- Backend/server:
  - `pkg/embeddable/mcpgo_backend.go` (new backend over mcp-go, tool/result mapping, debug logs)
  - `pkg/embeddable/command.go` (backend wiring)
  - `cmd/go-go-mcp/cmds/server/start.go` (tool provider proxy → registry, backend start, logs)
  - `cmd/go-go-mcp/main.go` (removed bridge wiring)

- Client:
  - `cmd/go-go-mcp/cmds/client/helpers/client.go` (mcp-go client creation/start/initialize)
  - `cmd/go-go-mcp/cmds/client/tools.go` (dual command)
  - `cmd/go-go-mcp/cmds/client/resources.go` (dual command)
  - `cmd/go-go-mcp/cmds/client/prompts.go` (dual command)

- Removals:
  - `pkg/server/*`, `pkg/transport/*`, `cmd/ui-server/*`, bridge command, legacy examples

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
