#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.." && pwd)"
cd "$REPO_ROOT"

PORT="${PORT:-4011}"
SERVER_LOG="$(mktemp)"
SERVER_PID=""

cleanup() {
  if [[ -n "${SERVER_PID}" ]]; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
    wait "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
  rm -f "$SERVER_LOG"
}
trap cleanup EXIT

env GOWORK=off go run ./pkg/embeddable/examples/oidc mcp start --transport sse --port "$PORT" >"$SERVER_LOG" 2>&1 &
SERVER_PID="$!"
sleep 3

echo "== discovery =="
curl -i -s "http://localhost:${PORT}/.well-known/openid-configuration"

echo
echo "== protected resource metadata =="
curl -i -s "http://localhost:${PORT}/.well-known/oauth-protected-resource"

echo
echo "== unauthenticated MCP SSE request =="
curl -i -s "http://localhost:${PORT}/mcp/sse"

echo
echo "== authenticated MCP SSE request (first event only) =="
curl -i -s --max-time 2 -H 'Authorization: Bearer TEST_AUTH_KEY_123' "http://localhost:${PORT}/mcp/sse" || true

echo
echo "OIDC smoke test completed."
