### MCP Protocol Decoupling: Replacing `go-go-mcp` internals with `mcp-go`

#### Purpose and scope
Outline an approach to reduce or remove MCP protocol/transport duplication in `go-go-mcp` by leveraging `mcp-go`, without breaking the `embeddable` API surface used by apps like `jesus`.

---

### Reasoning walkthrough

- `go-go-mcp` currently defines protocol types (`pkg/protocol`), server dispatcher (`pkg/server`), and transports (`pkg/transport`).
- `mcp-go` provides a full-featured protocol implementation with stdio/SSE/streamable-http transports and typed results (`CallToolResult`).
- The `embeddable` API mostly depends on: registering tools, session store, and starting a server with a chosen transport. These can be backed by `mcp-go` if we provide thin adapters.

---

### Decoupling plan

1) Define narrow internal interfaces in `go-go-mcp`
- Protocol types: `Tool`, `ToolResult`, `ToolContent` – keep local types for now but add adapter functions.
- Server facade: interface that takes a tool registry and exposes `Start(context.Context)`.
- Transport factory: interface to create stdio/SSE/streamable-http transports.

2) Adapters between `go-go-mcp` and `mcp-go`
- Protocol adapter
  - `ToMark3CallToolResult(*protocol.ToolResult) *mcp.CallToolResult`
  - `FromMark3CallToolResult(*mcp.CallToolResult) *protocol.ToolResult`
- Tool registry adapter
  - Wrap `tool-registry` provider so its `CallTool` is invoked from `mcp-go` handler, with argument map passthrough.

3) Progressive replacement
- Step A: Implement protocol adapter and integration tests for result parity.
- Step B: Add optional build tag or flag in `embeddable` to select backend (`go-go-mcp` server vs `mcp-go` server) – default stays current.
- Step C: Implement `mcp-go` backend: create `mcp-go/server.MCPServer`, register tools, and start selected transport (stdio/SSE/streamable-http).
- Step D: Mark legacy `pkg/transport` and `pkg/server` as deprecated once parity is proven.

4) Compatibility and tests
- Ensure `list-tools`, `test-tool`, and `start` flows behave identically across backends.
- Validate transports with simple probe clients.

---

### Risks and mitigations
- Type mismatches (`ToolResult` vs `CallToolResult`): comprehensive adapter tests.
- Behavior drift in initialization and capabilities: align server capabilities; expose flags as needed.
- Maintenance of two paths during transition: use build tags or minimal backend switch to reduce complexity.

---

### Next steps
- Implement protocol adapters and a minimal `mcp-go` backend behind a flag.
- Run integration tests across stdio/SSE/streamable-http.
- Document migration timeline and deprecation notes.

---

Requested output path: `go-go-mcp/ttmp/2025-08-23/04-mcp-protocol-decoupling.md`
