#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.." && pwd)"
cd "$REPO_ROOT"

SSE_PORT="${SSE_PORT:-4110}"
HTTP_PORT="${HTTP_PORT:-4111}"
SSE_LOG="$(mktemp)"
HTTP_LOG="$(mktemp)"
SSE_PID=""
HTTP_PID=""

cleanup() {
  if [[ -n "$SSE_PID" ]]; then
    kill "$SSE_PID" >/dev/null 2>&1 || true
    wait "$SSE_PID" >/dev/null 2>&1 || true
  fi
  if [[ -n "$HTTP_PID" ]]; then
    kill "$HTTP_PID" >/dev/null 2>&1 || true
    wait "$HTTP_PID" >/dev/null 2>&1 || true
  fi
  rm -f "$SSE_LOG" "$HTTP_LOG"
}
trap cleanup EXIT

echo "== command transport: tools list =="
command_list="$(
  timeout 20 go run ./cmd/go-go-mcp client tools list \
    --transport command \
    --command go \
    --command run \
    --command ./cmd/go-go-mcp \
    --command server \
    --command start \
    --command --internal-servers \
    --command echo \
    --command --watch=false \
    --command --transport \
    --command stdio \
    --command --log-level \
    --command error
)"
printf '%s\n' "$command_list"
grep -q "echo" <<<"$command_list"

echo
echo "== command transport: tools call =="
command_call="$(
  timeout 20 go run ./cmd/go-go-mcp client tools call echo \
    --transport command \
    --command go \
    --command run \
    --command ./cmd/go-go-mcp \
    --command server \
    --command start \
    --command --internal-servers \
    --command echo \
    --command --watch=false \
    --command --transport \
    --command stdio \
    --command --log-level \
    --command error \
    --json '{"message":"hello from command transport"}'
)"
printf '%s\n' "$command_call"
grep -q "hello from command transport" <<<"$command_call"

echo
echo "== sse transport: tools list + call =="
go run ./cmd/go-go-mcp server start \
  --internal-servers echo \
  --watch=false \
  --transport sse \
  --port "$SSE_PORT" \
  --log-level error >"$SSE_LOG" 2>&1 &
SSE_PID="$!"
sleep 2

sse_list="$(
  timeout 20 go run ./cmd/go-go-mcp client tools list \
    --transport sse \
    --server "http://127.0.0.1:${SSE_PORT}/mcp/sse"
)"
printf '%s\n' "$sse_list"
grep -q "echo" <<<"$sse_list"

sse_call="$(
  timeout 20 go run ./cmd/go-go-mcp client tools call echo \
    --transport sse \
    --server "http://127.0.0.1:${SSE_PORT}/mcp/sse" \
    --json '{"message":"hello from sse transport"}'
)"
printf '%s\n' "$sse_call"
grep -q "hello from sse transport" <<<"$sse_call"

echo
echo "== streamable_http transport: tools list + call =="
go run ./cmd/go-go-mcp server start \
  --internal-servers echo \
  --watch=false \
  --transport streamable_http \
  --port "$HTTP_PORT" \
  --log-level error >"$HTTP_LOG" 2>&1 &
HTTP_PID="$!"
sleep 2

http_list="$(
  timeout 20 go run ./cmd/go-go-mcp client tools list \
    --transport streamable_http \
    --server "http://127.0.0.1:${HTTP_PORT}/mcp"
)"
printf '%s\n' "$http_list"
grep -q "echo" <<<"$http_list"

http_call="$(
  timeout 20 go run ./cmd/go-go-mcp client tools call echo \
    --transport streamable_http \
    --server "http://127.0.0.1:${HTTP_PORT}/mcp" \
    --json '{"message":"hello from streamable http transport"}'
)"
printf '%s\n' "$http_call"
grep -q "hello from streamable http transport" <<<"$http_call"

echo
echo "Tool transport smoke completed."
