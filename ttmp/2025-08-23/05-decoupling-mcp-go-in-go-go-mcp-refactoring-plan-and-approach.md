### Refactoring Plan: Decoupling MCP in `go-go-mcp` by adopting `mcp-go`

#### Purpose and scope
Replace `go-go-mcp`'s internal protocol, server, and transport implementations with `mcp-go`, while retaining and enhancing the `embeddable` affordances. Backwards compatibility with internal code is not required; we will keep CLI affordances and developer UX stable or improved.

---

### Strategic goals
- Remove `go-go-mcp/pkg/server`, `go-go-mcp/pkg/transport`, and `go-go-mcp/pkg/protocol` implementations.
- Rebuild `embeddable` on top of `mcp-go` server and transports.
- Preserve tool registry affordances, hooks, middleware, sessions, and command UX.
- Offer more `mcp-go`-like features where beneficial (annotations, typed schemas, streamable-http).
- Remove the SSE bridge command entirely.
- Migrate client CLI to use `mcp-go` client.

---

### Repository map and affected areas

- Server/Transport/Protocol (to be removed):
  - `go-go-mcp/pkg/server/*` (Server, RequestHandler, options)
  - `go-go-mcp/pkg/transport/*` (interfaces, sse, stdio, streamable_http)
  - `go-go-mcp/pkg/protocol/*` (message and ToolResult types)
- Embeddable (keep; refactor internals):
  - `go-go-mcp/pkg/embeddable/*` (`server.go`, `command.go`, `tools.go`, `enhanced_tools.go`, `arguments.go`)
- Tooling/Registry (keep):
  - `go-go-mcp/pkg/tools/*` and subproviders (config-provider, tool-registry)
- CLI (adapt):
  - Server: `go-go-mcp/cmd/go-go-mcp/cmds/server/*` (`start.go`, `tools.go`)
  - Client: `go-go-mcp/cmd/go-go-mcp/cmds/client/*` (helpers, layers, prompts/resources/tools)
  - Bridge: `go-go-mcp/cmd/go-go-mcp/cmds/bridge.go` (remove)
  - UI: `go-go-mcp/cmd/go-go-mcp/cmds/ui_cmd.go` (keep)
- Config/Docs (keep):
  - `go-go-mcp/pkg/config/*`, `go-go-mcp/pkg/doc/*`

---

### Target end-state architecture

- `embeddable/`: A façade that wraps an `mcp-go/server.MCPServer` instance and registers tools/resources/prompts from our registries.
  - CLI entrypoints (`embeddable/command.go`) call into a backend constructor and `Start(ctx)`.
  - Transport selection maps to `mcp-go` transports (stdio, sse, streamable-http).
  - Hooks and middleware are executed around tool handler invocation.
- Client CLI uses `mcp-go` client to talk to servers over stdio/SSE/streamable-http.
- SSE bridge command is removed.

---

### File-by-file plan and symbol mapping

- Remove internals:
  - Delete `pkg/server` (symbols: `type Server`, `func NewServer`, `(*Server).Start`, options like `WithToolProvider`, `WithSessionStore`)
  - Delete `pkg/transport` (symbols: `type Transport`, `NewSSETransport`, `NewStdioTransport`, `NewStreamableHTTPTransport`)
  - Delete `pkg/protocol` (symbols: `ToolResult`, `ToolContent`, `NewToolResult`, `WithText`, etc.)

- Embeddable backend (new): `pkg/embeddable/mcpgo_backend.go`
  - New symbols:
    - `type Backend interface { Start(ctx context.Context) error }`
    - `func NewBackend(config *ServerConfig) (Backend, error)`
  - Responsibilities:
    - Build `mcpServer := mcpserver.NewMCPServer(name, version, options...)`
    - Register tools from `config.toolRegistry` as `mcp-go` tools
    - Optionally register resources/prompts if present
    - Start transport:
      - stdio: `server.NewStdioServer(mcpServer).Listen(ctx, os.Stdin, os.Stdout)`
      - sse: `server.NewSSEServer(mcpServer).Start(":PORT")`
      - streamable-http: `server.NewStreamableHTTPServer(mcpServer).Start(":PORT")`

- Embeddable command wiring: `pkg/embeddable/command.go`
  - Replace direct use of `pkg/server`/`pkg/transport` with `NewBackend(config)` + `backend.Start(ctx)`
  - Keep flags: `--transport`, `--port`, `--internal-servers`; store flags in context as today
  - Preserve `WithTool`, `WithEnhancedTool`, `WithHooks`, `WithMiddleware`, `WithSessionStore` semantics

- Tool registration adapter: could live in `pkg/embeddable/server.go`
  - Map our `tools.Tool` to `mcp-go/mcp.Tool`:
    - If we hold raw JSON schema, pass through to `mcp-go`
    - Map annotations to `mcp.ToolAnnotation`
  - Handler adapter pseudocode:
    - Input: `req mcp.CallToolRequest`
    - Convert `req.GetArguments()` to `map[string]interface{}`
    - Build context with session from `mcp-go` (if any)
    - Apply `BeforeToolCall` hooks and middleware chain
    - Call `config.toolRegistry.CallTool(ctx, name, args)`
    - Convert our `protocol.ToolResult` to `mcp.CallToolResult`
    - Apply `AfterToolCall` hooks

- Result mapping (our -> mcp-go):
  - text: `{ Type: "text", Text: <text> }`
  - JSON: we already serialize to text with `application/json`; convert to `TextContent` (or optionally use structured content later)
  - image: `{ Type: "image", Data: <b64>, MIMEType: <mime> }`
  - resource: `{ Type: "resource", Resource: EmbeddedResource{...} }`

- Resources/Prompts registration adapter (optional, if used by CLI):
  - Provide functions to add resources and prompts from our registry to `mcp-go` server, mirroring current providers

- Client migration: `cmd/go-go-mcp/cmds/client/helpers/client.go`
  - Replace `go-go-mcp/pkg/client` with `mcp-go` client
  - Symbol plan:
    - For SSE: `mcp-go/client.NewSSEClient(baseURL)` or build equivalent HTTP/SSE client
    - For stdio command: use `mcp-go` stdio transport (or a `Process` transport) if available; otherwise keep command-stdio temporarily until equivalent exists
  - Pseudocode:
    - Parse flags: transport/server/command
    - Build `transport` via `mcp-go/client` helpers
    - Create client, call `Initialize`
    - Expose methods: `ListTools`, `CallTool`, `ListResources`, `ReadResource`, `ListPrompts`, `GetPrompt`

- CLI server start refactor: `cmd/go-go-mcp/cmds/server/start.go`
  - Replace transport/server creation with embeddable backend call
  - Preserve file-watcher for config-provider and graceful shutdown using errgroup

- Remove bridge: delete `cmd/go-go-mcp/cmds/bridge.go` and references in `main.go`

---

### Pseudocode snippets

- New backend constructor:
```go
func NewBackend(cfg *ServerConfig) (Backend, error) {
    // Build mcp-go server
    s := mcpserver.NewMCPServer(cfg.Name, cfg.Version,
        mcpserver.WithToolCapabilities(true),
        // map cfg options to mcp-go capabilities
    )

    // Register tools
    for _, tool := range cfg.toolRegistry.List() { // pseudo
        mt := mapToMark3Tool(tool)
        s.AddTool(mt, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            args := req.GetArguments()
            if cfg.hooks != nil && cfg.hooks.BeforeToolCall != nil {
                if err := cfg.hooks.BeforeToolCall(ctx, tool.Name, args); err != nil { return nil, err }
            }
            res, err := cfg.toolRegistry.CallTool(ctx, tool.Name, args)
            mr := mapToMark3Result(res)
            if cfg.hooks != nil && cfg.hooks.AfterToolCall != nil { cfg.hooks.AfterToolCall(ctx, tool.Name, res, err) }
            return mr, err
        })
    }

    // Start selected transport
    switch cfg.defaultTransport {
    case "stdio":
        return &stdioBackend{server: s}, nil
    case "sse":
        return &sseBackend{server: s, port: cfg.defaultPort}, nil
    case "streamable_http":
        return &streamBackend{server: s, port: cfg.defaultPort}, nil
    }
    return nil, fmt.Errorf("unknown transport")
}
```

- Backend Start implementations:
```go
type stdioBackend struct { server *mcpserver.MCPServer }
func (b *stdioBackend) Start(ctx context.Context) error {
    return server.ServeStdio(b.server) // mcp-go stdio
}

type sseBackend struct { server *mcpserver.MCPServer; port int }
func (b *sseBackend) Start(ctx context.Context) error {
    return server.NewSSEServer(b.server).Start(fmt.Sprintf(":%d", b.port))
}

type streamBackend struct { server *mcpserver.MCPServer; port int }
func (b *streamBackend) Start(ctx context.Context) error {
    return server.NewStreamableHTTPServer(b.server).Start(fmt.Sprintf(":%d", b.port))
}
```

- Client helper (mcp-go):
```go
func CreateClientFromSettings(parsed *layers.ParsedLayers) (*mcpclient.Client, error) {
    s := &ClientSettings{}
    if err := parsed.InitializeStruct(ClientLayerSlug, s); err != nil { return nil, err }

    switch s.Transport {
    case "sse":
        c := mcpclient.NewHTTPClient(s.Server) // or HTTP/SSE variant per mcp-go
        if err := c.Initialize(ctx, mcp.ClientCapabilities{}); err != nil { return nil, err }
        return c, nil
    case "stdio":
        // Use mcp-go stdio client to a spawned process if provided; else not supported
    }
    return nil, fmt.Errorf("unsupported transport: %s", s.Transport)
}
```

---

### Testing strategy

- Unit tests:
  - Tool/result mapping adapters (round-trip text/JSON/image/resource)
  - Backend constructor error cases (bad transport, nil registry)
- Integration tests:
  - `embeddable` server start in stdio, sse, streamable-http; `list-tools` and `test-tool` correctness
  - Client CLI verbs against a running server for each transport
- Golden tests for CLI output stability (tools/resources/prompts)

---

### Risks and mitigations

- Removal of internals may break unanticipated callers: document removals and provide clear alternatives via `embeddable` and client CLI.
- Client transport parity: ensure SSE and streamable-http paths are covered; stdio client to spawned processes may require a small wrapper.

---

### Cutover plan

- Implement backend and client migration behind a short-lived feature flag.
- Validate all commands end-to-end.
- Remove legacy code and the flag; ship.

---

### Comprehensive TODO checklist

- [ ] Annotate legacy packages (`pkg/server`, `pkg/transport`, `pkg/protocol`) as deprecated in docs
- [ ] Implement `pkg/embeddable/mcpgo_backend.go` with `Backend` and `NewBackend`
- [ ] Map tool registration (our `tools.Tool` -> `mcp-go` `Tool`)
- [ ] Implement result adapters (our `ToolResult` -> `mcp-go` `CallToolResult`)
- [ ] Wire hooks and middleware into handler path
- [ ] Update `pkg/embeddable/command.go` to use backend
- [ ] Refactor `cmds/server/start.go` to call into `embeddable` backend
- [ ] Remove `cmds/bridge.go` and references from `main.go`
- [ ] Migrate `cmds/client/helpers/client.go` to `mcp-go` client
- [ ] Verify `client tools/resources/prompts` work with `mcp-go` client
- [ ] Remove `pkg/server`, `pkg/transport`, `pkg/protocol` sources
- [ ] Update docs in `pkg/doc` and `README.md`
- [ ] Add integration tests (start/list/call across transports)
- [ ] Add golden tests for CLI output
- [ ] CI: run tests on all transports

---

Requested output path: `go-go-mcp/ttmp/2025-08-23/05-decoupling-mcp-go-in-go-go-mcp-refactoring-plan-and-approach.md`
