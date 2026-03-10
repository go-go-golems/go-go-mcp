#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/../../../../../.." && pwd)"
PORT="${PORT:-$((4000 + RANDOM % 1000))}"
AUTH_KEY="${AUTH_KEY:-TEST_AUTH_KEY_123}"
LOG_FILE="${SCRIPT_DIR}/auth-mode-smoke.log"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
    wait "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

cd "${REPO_DIR}"

echo "== Focused external_oidc tests =="
go test ./pkg/embeddable -run 'TestNewHTTPAuthProviderSelectsExternalOIDC|TestExternalOIDCProvider' -count=1

echo "== Embedded_dev runtime smoke =="
go run ./pkg/embeddable/examples/oidc \
  mcp start \
  --transport streamable_http \
  --port "${PORT}" \
  --auth-mode embedded_dev \
  --embedded-issuer "http://127.0.0.1:${PORT}" \
  --embedded-auth-key "${AUTH_KEY}" \
  >"${LOG_FILE}" 2>&1 &
SERVER_PID=$!

for _ in $(seq 1 40); do
  if curl -fsS "http://127.0.0.1:${PORT}/.well-known/oauth-protected-resource" >/dev/null 2>&1; then
    break
  fi
  sleep 0.25
done

METADATA="$(curl -fsS "http://127.0.0.1:${PORT}/.well-known/oauth-protected-resource")"
echo "${METADATA}" | rg -q '"resource":"http://127.0.0.1:'"${PORT}"'/mcp"' || {
  echo "protected resource metadata did not advertise the expected resource URL" >&2
  exit 1
}

UNAUTH_HEADERS="$(mktemp)"
AUTH_HEADERS="$(mktemp)"
curl --max-time 2 -sS -X POST -D "${UNAUTH_HEADERS}" -o /dev/null "http://127.0.0.1:${PORT}/mcp" || true
rg -q 'HTTP/1.1 401 Unauthorized' "${UNAUTH_HEADERS}"
rg -qi '^Www-Authenticate: Bearer realm="mcp"' "${UNAUTH_HEADERS}"

curl --max-time 2 -sS -D "${AUTH_HEADERS}" -o /dev/null \
  -X POST \
  -H "Authorization: Bearer ${AUTH_KEY}" \
  "http://127.0.0.1:${PORT}/mcp" || true
if rg -q 'HTTP/1.1 401 Unauthorized' "${AUTH_HEADERS}"; then
  echo "authorized request still returned 401" >&2
  exit 1
fi
rg -q 'HTTP/1.1 (400|405)' "${AUTH_HEADERS}"

echo "Smoke passed. Log: ${LOG_FILE}"
