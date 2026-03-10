---
Title: Testing MCP Servers with go-go-mcp
Slug: testing-mcp-servers
Short: Playbook for exercising MCP servers over stdio and SSE using the go-go-mcp client.
Topics:
- mcp
- testing
- playbooks
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Testing MCP Servers with go-go-mcp

This playbook walks through verifying that an MCP server (the `jesus` server in this example) behaves correctly over both stdio and SSE transports using the `go-go-mcp` CLI. The same workflow applies to any embeddable MCP server built with go-go-mcp.

## Prerequisites

Before running the tests, make sure:

- The `go-go-mcp` and `jesus` repositories are checked out under the same workspace so relative `go run ../jesus/...` paths resolve.
- `GOCACHE` points to directories inside each repo (the snippets below reuse `$(pwd)/.gocache`) to avoid permission issues.
- Ports `4010`, `9930`, and `9095` are free for the SSE test. Adjust them if your environment requires different values.

## Stdio Transport (command)

The command transport launches the server as a child process, giving you a quick end-to-end test without leaving stray servers running. The client takes an explicit `--command` flag per argument, so you can pass the entire `go run` invocation.

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-go-mcp

GOCACHE=$(pwd)/.gocache \
  go run ./cmd/go-go-mcp client tools list \
  --transport command \
  --command env \
  --command GOCACHE=/home/manuel/code/wesen/corporate-headquarters/jesus/.gocache \
  --command go \
  --command run \
  --command ../jesus/cmd/jesus \
  --command mcp \
  --command start \
  --command --log-level \
  --command debug \
  --command --with-caller
```

This spawns `jesus mcp start` under the hood, waits for it to initialize, and prints the available tools. You should see the `executeJS` tool description, confirming that stdio transport works end-to-end. When the command exits, the child server shuts down automatically.

## SSE Transport

SSE testing requires a long-running server instance plus an HTTP endpoint for event streaming. Start `jesus` in another terminal (or tmux pane) with explicit ports to avoid collisions:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/jesus

GOCACHE=$(pwd)/.gocache \
  go run ./cmd/jesus mcp start \
  --transport sse \
  --port 4010 \
  --js-port 9930 \
  --admin-port 9095 \
  --log-level debug \
  --with-caller
```

Once the server logs show “Starting SSE server (single-port) addr=:4010 endpoint=/mcp”, run the client against the SSE endpoint. Note that the client expects the full SSE URL (`/mcp/sse`), not just the base address.

```bash
cd /home/manuel/code/wesen/corporate-headquarters/go-go-mcp

GOCACHE=$(pwd)/.gocache \
  go run ./cmd/go-go-mcp client tools list \
  --transport sse \
  --server http://localhost:4010/mcp/sse
```

You should again see the `executeJS` tool definition. Keep the SSE server running for further manual tests (e.g., `client tools call ...`). When finished, stop the server with `Ctrl+C` or `tmux kill-session`.

## Cleanup Checklist

- Terminate any tmux sessions or background servers started for SSE tests.
- Remove temporary caches if necessary (`rm -rf .gocache` in either repo).
- If you changed ports for the SSE test, document them in your `.env` or profile so future runs stay consistent.

Following this playbook ensures that both stdio and SSE transports remain healthy whenever you touch the MCP integration.
