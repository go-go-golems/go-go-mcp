### CLI and Affordances Impact Analysis: Requirements for the `mcp-go` refactor

#### Purpose and scope
Assess how `go-go-mcp` CLI commands and other affordances (UI, resources, config, loaders) depend on current internals, and list what they need from the refactor to work unchanged or improved.

---

### Server-side CLI (cmd/go-go-mcp/cmds/server)

- `server start` (start.go)
  - Today: constructs `pkg/transport` (stdio/sse/streamable_http), builds `pkg/server.Server`, and injects tool/resource providers; manages graceful shutdown.
  - After refactor: construct `embeddable` backend (mcp-go), set transport via flag, call `backend.Start(ctx)`. Keep flags `--transport`, `--port`. Resource/provider wiring must move into `embeddable` or a thin layer above it.
  - Needs:
    - Backend constructor that accepts tool/resource providers or a registry and exposes `Start(ctx)`.
    - SSE/stdio/streamable-http transport selection via flags (mapped to `mcp-go`).

- `server tools list` and `server tools call` (tools.go)
  - Today: build a tool provider from config, list or call tools directly (no running server required).
  - After refactor: unchanged; relies on provider/registry code, not on `pkg/server` internals.
  - Needs:
    - Keep `tools/providers/tool-registry` and config layers intact.

- `server` root (server.go)
  - Simply composes subcommands; unchanged.

---

### Bridge command (bridge.go)

- Today: `server.NewSSEBridgeServer(logger, sseURL)` then `.Start(context.Background())` – depends on `pkg/server` bridge implementation.
- After refactor: implement a small `mcp-go`-based bridge (stdio server that proxies to remote streamable-http/SSE). If not essential, mark as deferred or rebuild with `mcp-go` client/server.
- Needs:
  - New bridge built on `mcp-go` client + stdio server, or temporarily disable if not needed.

---

### Client-side CLI (cmd/go-go-mcp/cmds/client)

- `client tools|resources|prompts`
  - Today: use `helpers.CreateClientFromSettings` and a client that speaks to servers (independent of server internals).
  - After refactor: unchanged; ensure client package remains compatible (either keep existing client or migrate to `mcp-go` client later).
  - Needs:
    - Keep current client helper API stable for now.

---

### UI (tui)

- TUI command uses `pkg/ui/tui` to manage profiles. No dependency on server internals beyond config format.
- After refactor: unchanged; optionally add awareness of streamable-http endpoints.

---

### Resources and providers

- `pkg/resources/registry.go` is used by `server start` to expose resources. This must plug into the `mcp-go` server backend.
- After refactor: add adapter to register resources/prompts with `mcp-go` server during backend initialization.
- Needs:
  - Backend hooks to register resources and prompts from registry implementations.

---

### Config and loaders

- `pkg/config` types and paths; YAML command loaders (`pkg/cmds`) used by `schema` and `run-command` affordances are independent of MCP internals.
- After refactor: unchanged.

---

### Summary of required changes

- Replace `server start` internals to call `embeddable` mcp-go backend; maintain flags and graceful shutdown.
- Provide backend APIs to accept tool/resource/prompt providers; wire them to `mcp-go` server.
- Reimplement or postpone SSE bridge using `mcp-go` building blocks.
- Keep client CLI and loaders intact; consider migrating client to `mcp-go` in a later phase.
- Ensure `embeddable` exposes: transport selection, hooks, middleware, session access, and registration for tools/resources/prompts.

---

### Actionable next steps
- [ ] Add backend constructor that accepts providers and starts `mcp-go` transports.
- [ ] Refactor `server start` to use the new backend.
- [ ] Add adapters to register resources/prompts into `mcp-go` during backend setup.
- [ ] Decide path for bridge: reimplement with `mcp-go` or deprecate.
- [ ] Verify client CLI remains functional against refactored server.

---

Requested output path: `go-go-mcp/ttmp/2025-08-23/06-cli-and-affordances-impact-analysis.md`
