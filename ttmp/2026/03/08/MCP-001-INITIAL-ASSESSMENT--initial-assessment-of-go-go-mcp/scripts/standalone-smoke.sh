#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.." && pwd)"
cd "$REPO_ROOT"

SSE_PORT="${SSE_PORT:-4010}"
HTTP_PORT="${HTTP_PORT:-4012}"
SSE_LOG="$(mktemp)"
HTTP_LOG="$(mktemp)"
SSE_PID=""
HTTP_PID=""

cleanup() {
  if [[ -n "${SSE_PID}" ]]; then
    kill "${SSE_PID}" >/dev/null 2>&1 || true
    wait "${SSE_PID}" >/dev/null 2>&1 || true
  fi
  if [[ -n "${HTTP_PID}" ]]; then
    kill "${HTTP_PID}" >/dev/null 2>&1 || true
    wait "${HTTP_PID}" >/dev/null 2>&1 || true
  fi
  rm -f "$SSE_LOG" "$HTTP_LOG"
}
trap cleanup EXIT

echo "== standalone module test =="
env GOWORK=off go test ./...

echo
echo "== standalone module build =="
env GOWORK=off go build ./cmd/go-go-mcp

echo
echo "== embeddable basic example =="
env GOWORK=off go run ./pkg/embeddable/examples/basic mcp list-tools

echo
echo "== embeddable struct example =="
env GOWORK=off go run ./pkg/embeddable/examples/struct mcp list-tools

echo
echo "== embeddable enhanced example =="
env GOWORK=off go run ./pkg/embeddable/examples/enhanced mcp list-tools

echo
echo "== command transport smoke =="
env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport command \
  --command env --command GOWORK=off --command go --command run \
  --command ./pkg/embeddable/examples/basic --command mcp --command start

env GOWORK=off go run ./cmd/go-go-mcp client tools call greet --transport command \
  --command env --command GOWORK=off --command go --command run \
  --command ./pkg/embeddable/examples/basic --command mcp --command start \
  --json '{"name":"Intern"}'

echo
echo "== SSE transport smoke =="
env GOWORK=off go run ./pkg/embeddable/examples/basic mcp start --transport sse --port "$SSE_PORT" >"$SSE_LOG" 2>&1 &
SSE_PID="$!"
sleep 2
env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport sse --server "http://localhost:${SSE_PORT}/mcp/sse"

echo
echo "== streamable_http transport smoke =="
env GOWORK=off go run ./pkg/embeddable/examples/basic mcp start --transport streamable_http --port "$HTTP_PORT" >"$HTTP_LOG" 2>&1 &
HTTP_PID="$!"
sleep 2
env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport streamable_http --server "http://localhost:${HTTP_PORT}/mcp"

echo
echo "Smoke test completed."
