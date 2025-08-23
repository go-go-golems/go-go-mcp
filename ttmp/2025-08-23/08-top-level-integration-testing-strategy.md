### Top-level integration testing strategy for go-go-mcp

#### Purpose and scope

This document proposes a practical plan for end-to-end integration testing of go-go-mcp using real transports, the real CLI, and real internal tools. The approach focuses on tmux-based orchestration locally, with an optional Docker-based setup to isolate the environment.

Goals:
- Exercise server and client with real transports: stdio, StreamableHTTP, SSE
- Verify tool discovery and tool invocation using real internal tools (sqlite, fetch, echo)
- Produce deterministic, machine-verifiable outputs (JSON) for assertions
- Run locally in CI-friendly fashion; optionally isolate with Docker

Non-goals:
- Unit tests for individual components (covered elsewhere)
- Full network isolation guarantees (Docker option covers most needs)

---

### Test matrix

- Transports:
  - stdio (command)
  - streamable_http (HTTP) → default endpoint: /mcp
  - sse → SSE endpoint: /sse
- Internal servers/tools enabled during tests:
  - echo, fetch, sqlite
- Client flows to cover:
  - tools list (human + glaze JSON)
  - tools call echo (JSON args)
  - resources list/read (if resources are wired later)
  - prompts list/execute (if/when added)

---

### Local orchestration with tmux

We use tmux to run long-lived servers in background sessions, and the CLI as foreground processes for assertions.

Key patterns:
- Start server in a dedicated session/name per transport
- Wait for port to be open (simple retry loop with curl/nc)
- Run client commands with `--with-glaze-output --output json` and assert with jq/grep
- Tear down tmux session(s)

Example scripts:

```bash
#!/usr/bin/env bash
set -euo pipefail
ROOT="/home/manuel/workspaces/2025-08-20/migrate-jesus-go-mcp/go-go-mcp"
PORT=${PORT:-3001}
SESSION=${SESSION:-mcp_http_test}

# Start StreamableHTTP server
cmd="cd $ROOT && go run ./cmd/go-go-mcp server start \
  --transport streamable_http \
  --port $PORT \
  --internal-servers sqlite,fetch,echo \
  --log-level debug --with-caller"

tmux new-session -d -s "$SESSION" "$cmd"

# Wait for HTTP to respond
for i in {1..50}; do
  if curl -sf "http://localhost:${PORT}/mcp" >/dev/null; then
    break
  fi
  sleep 0.2
  if [ "$i" -eq 50 ]; then echo "Server did not start"; exit 1; fi
done

# List tools (JSON)
cd "$ROOT"
TOOLS_JSON=$(go run ./cmd/go-go-mcp client tools list \
  --transport streamable_http \
  --server "http://localhost:${PORT}/mcp" \
  --with-glaze-output --output json)

echo "$TOOLS_JSON" | jq -e 'map(.name) | index("echo")' >/dev/null

# Call echo
CALL_JSON=$(go run ./cmd/go-go-mcp client tools call echo \
  --transport streamable_http \
  --server "http://localhost:${PORT}/mcp" \
  --json '{"message":"hello"}' \
  --with-glaze-output --output json || true)

echo "$CALL_JSON" | jq -e '.[0].text == "hello"' >/dev/null

# Teardown
(tmux kill-session -t "$SESSION" >/dev/null 2>&1) || true
```

Adapt for SSE:
- Use `--transport sse` and `--server http://localhost:3002/sse`
- Start session name `mcp_sse_test` on port 3002
- The readiness probe can hit `/sse` for HTTP 200

---

### Optional Docker-based orchestration

Motivation: increased isolation, repeatability of environment and dependencies.

Options:
- Single container running both server and client (simpler, less isolated)
- Two containers on a user-defined network (server, client) (better network parity)

Skeleton Dockerfile:

```Dockerfile
FROM golang:1.24 as build
WORKDIR /app
# Copy only go-go-mcp module to speed up build cache; use go work for multi-module if needed
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/mcp ./cmd/go-go-mcp

FROM debian:stable-slim
RUN apt-get update && apt-get install -y ca-certificates curl jq && rm -rf /var/lib/apt/lists/*
COPY --from=build /usr/local/bin/mcp /usr/local/bin/mcp
ENTRYPOINT ["/usr/local/bin/mcp"]
```

Docker run examples:
- Server: `docker run --rm -p 3001:3001 mcp image server start --transport streamable_http --port 3001 --internal-servers sqlite,fetch,echo`
- Client: `docker run --rm --network host mcp image client tools list --transport streamable_http --server http://localhost:3001/mcp --with-glaze-output --output json`

Notes:
- For Linux CI runners, `--network host` is acceptable; on macOS it maps to a VM, adjust with container networks
- Alternatively, create a user-defined bridge network and reference the server container by name

---

### CI wiring

- Add a `Makefile` target `integration-http` and `integration-sse` that run the tmux scripts
- Optionally add a `docker-compose.yml` to bring up a server service; run client as a one-off job
- Use `jq -e` expressions for hard assertions
- Log at `--log-level debug --with-caller` to aid diagnosis on CI failures

---

### Next steps

- [ ] Add `scripts/integration/http_smoke.sh` and `scripts/integration/sse_smoke.sh`
- [ ] Add `make integration-http` and `make integration-sse`
- [ ] Add `docker/Dockerfile` and optional `docker-compose.yml`
- [ ] Create a GitHub Action job that executes the HTTP smoke test on push/PR
- [ ] Expand coverage to include `sqlite_open` + `sqlite_query` minimal round-trip
- [ ] Optionally add `fetch` test with a known stable URL and snapshot output
