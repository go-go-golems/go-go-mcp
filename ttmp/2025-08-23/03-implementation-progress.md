### Implementation Progress: Porting `jesus` to `go-mcp`

#### Purpose and scope
Track concrete implementation steps as they land, with code references and next tasks.

---

### Reasoning walkthrough
- Current code already registers `executeJS` via `embeddable.AddMCPCommand` in `jesus/pkg/mcp/server.go` and spins up JS/admin servers on startup hook.
- Next work focuses on adding more tools and creating a `JSEngineService` adapter without changing transports.

---

### Completed
- Initial inventory and migration plan documents.

---

### In progress
- Designing `JSEngineService` adapter and identifying engine methods for list routes/logs.

Potential locations in engine:
- Handler registry: `jesus/pkg/engine/engine.go` has `handlers map[string]map[string]*HandlerInfo` and accessors used in router; we can add read-only getters to enumerate.
- Request logs: `engine.Engine` exposes `GetRequestLogger()` used in admin; add method to tail logs or expose last N entries.

---

### Next steps
- Add read-only getters to engine for routes and logs if missing.
- Implement `JSEngineService` and wire additional MCP tools in `jesus/pkg/mcp/server.go`.
- Validate via `go-go-mcp mcp start --transport stdio` and tool `list-tools`.

---

Requested output path: `go-go-mcp/ttmp/2025-08-23/03-implementation-progress.md`
