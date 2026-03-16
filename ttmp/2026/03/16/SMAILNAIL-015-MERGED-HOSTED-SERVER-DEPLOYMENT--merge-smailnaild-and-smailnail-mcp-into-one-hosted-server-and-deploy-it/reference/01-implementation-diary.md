---
Title: Implementation diary for merging smailnaild and MCP into one hosted server
Ticket: SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - oidc
    - keycloak
    - coolify
    - deployments
    - diary
DocType: reference
Intent: long-term
Summary: Chronological record of the server-merge and deployment work, including failed attempts, validation commands, and rollout notes.
---

# Implementation diary for merging smailnaild and MCP into one hosted server

## Kickoff

### Prompt context

The user wants to stop treating the hosted web app and the hosted MCP as separate production services and instead serve them from the same binary and the same `http.Server`. The immediate request is to create a new ticket and make it actionable enough to drive the actual refactor and deployment work.

### What I did

I created ticket `SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT` and added:

- a detailed implementation plan
- a granular task breakdown
- this diary as the running execution log

### Why

The merge is large enough that it needs a fresh execution record rather than being buried inside older identity or deployment tickets. The older tickets explain why the current split exists; this new ticket should explain how to remove that split safely.

### Initial assumptions

- the single-server design is desirable
- the merge should keep browser-session auth and bearer-token auth distinct
- the deployment cutover should happen only after a local proof and a hosted smoke

### Immediate next step

Start with the code-shape refactor:

- extract a mountable MCP handler from the standalone server package
- then mount it into `smailnaild`

## Implementation step 1: build the mounted-MCP path and thread it into smailnaild

### What I changed

I made the first structural merge slice across both repositories.

In `go-go-mcp`, I added a mounted-HTTP path in [mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go):

- exported `MountHTTPHandlers(...)`
- factored the existing SSE and streamable-HTTP route setup into reusable mount helpers
- kept the existing standalone backends working by making them reuse the same mount logic internally

In `smailnail`, I then added a hosted mounting layer in [server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go):

- `MountHTTPHandlers(...)` for the smailnail MCP package
- a shared `baseServerOptions(...)` helper so the mounted and standalone forms register the same tools and middleware

I also introduced hosted MCP config in [hosted_config.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/hosted_config.go) and threaded it into [serve.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnaild/commands/serve.go).

Finally, I updated [http.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go) so `smailnaild` can mount:

- `/.well-known/oauth-protected-resource`
- `/mcp`
- `/mcp/`

before the SPA fallback handler is registered.

### Why this matters

This is the minimum structural change needed before any deployment work can happen. Without it, `smailnaild` cannot own the MCP HTTP routes, and we would still be stuck with two separately booted servers.

### Validation

I ran:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp
go test ./pkg/embeddable/...
```

and:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/mcp/imapjs ./pkg/smailnaild ./cmd/smailnaild/...
```

I also added a route-level proof in [mounted_handler_test.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/mounted_handler_test.go) showing that the mounted MCP mux can be served by the hosted app handler without breaking `/api/info`.

### What remains after this step

The mounted route exists, but the ticket is not close to done yet. The next concrete slice is the real local full-stack proof:

- run the merged `smailnaild` with the local Keycloak and Dovecot stack
- verify browser login still works
- verify account setup still works
- verify `/mcp` on the same server still works with bearer auth
