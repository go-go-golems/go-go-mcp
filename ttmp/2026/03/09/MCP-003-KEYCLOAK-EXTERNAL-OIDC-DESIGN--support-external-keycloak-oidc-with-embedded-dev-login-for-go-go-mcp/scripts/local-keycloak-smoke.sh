#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TICKET_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_DIR="$(cd "${TICKET_DIR}/../../../../.." && pwd)"
KEYCLOAK_DIR="${SCRIPT_DIR}/keycloak-local"
KEYCLOAK_PORT="${KEYCLOAK_PORT:-18080}"
MCP_BASIC_PORT="${MCP_BASIC_PORT:-43001}"
MCP_STRICT_PORT="${MCP_STRICT_PORT:-43002}"
KEYCLOAK_ISSUER="http://127.0.0.1:${KEYCLOAK_PORT}/realms/mcp-local"
COMPOSE_FILE="${KEYCLOAK_DIR}/docker-compose.yml"
KEYCLOAK_LOG="${SCRIPT_DIR}/local-keycloak.log"
MCP_BASIC_LOG="${SCRIPT_DIR}/local-mcp-basic.log"
MCP_STRICT_LOG="${SCRIPT_DIR}/local-mcp-strict.log"

cleanup() {
  if [[ -n "${MCP_BASIC_PID:-}" ]]; then
    kill "${MCP_BASIC_PID}" >/dev/null 2>&1 || true
    wait "${MCP_BASIC_PID}" >/dev/null 2>&1 || true
  fi
  if [[ -n "${MCP_STRICT_PID:-}" ]]; then
    kill "${MCP_STRICT_PID}" >/dev/null 2>&1 || true
    wait "${MCP_STRICT_PID}" >/dev/null 2>&1 || true
  fi
  KEYCLOAK_PORT="${KEYCLOAK_PORT}" docker compose -f "${COMPOSE_FILE}" down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT

wait_for_url() {
  local url="$1"
  local attempts="${2:-60}"
  for _ in $(seq 1 "${attempts}"); do
    if curl --max-time 3 -fsS "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  return 1
}

mint_token() {
  local client_id="$1"
  local client_secret="$2"
  curl --max-time 10 -fsS \
    -X POST "${KEYCLOAK_ISSUER}/protocol/openid-connect/token" \
    -d grant_type=client_credentials \
    -d "client_id=${client_id}" \
    -d "client_secret=${client_secret}" | jq -r .access_token
}

assert_status() {
  local expected="$1"
  local url="$2"
  shift 2
  local status
  status="$(curl --max-time 5 -sS -o /dev/null -w '%{http_code}' "$@" "${url}" || true)"
  if [[ "${status}" != "${expected}" ]]; then
    echo "expected ${expected} from ${url}, got ${status}" >&2
    return 1
  fi
}

assert_not_status() {
  local forbidden="$1"
  local url="$2"
  shift 2
  local status
  status="$(curl --max-time 5 -sS -o /dev/null -w '%{http_code}' "$@" "${url}" || true)"
  if [[ "${status}" == "${forbidden}" ]]; then
    echo "unexpected ${forbidden} from ${url}" >&2
    return 1
  fi
}

cd "${REPO_DIR}"

echo "== Start local Keycloak =="
KEYCLOAK_PORT="${KEYCLOAK_PORT}" docker compose -f "${COMPOSE_FILE}" up -d >"${KEYCLOAK_LOG}" 2>&1
wait_for_url "${KEYCLOAK_ISSUER}/.well-known/openid-configuration" 90

echo "== Mint tokens =="
BASIC_TOKEN="$(mint_token mcp-cli-basic mcp-cli-basic-secret)"
STRICT_TOKEN="$(mint_token mcp-cli-strict mcp-cli-strict-secret)"
test -n "${BASIC_TOKEN}" && [[ "${BASIC_TOKEN}" != "null" ]]
test -n "${STRICT_TOKEN}" && [[ "${STRICT_TOKEN}" != "null" ]]

STRICT_PAYLOAD="$(echo "${STRICT_TOKEN}" | awk -F. '{print $2}' | tr '_-' '/+' | base64 -d 2>/dev/null || true)"
echo "${STRICT_PAYLOAD}" | jq -e '(.aud == "mcp-resource") or ((.aud | type) == "array" and (.aud | index("mcp-resource") != null))' >/dev/null 2>&1 || {
  echo "strict token did not contain expected audience" >&2
  exit 1
}
echo "${STRICT_PAYLOAD}" | jq -e '.scope | strings and contains("mcp-invoke")' >/dev/null 2>&1 || {
  echo "strict token did not contain expected scope" >&2
  exit 1
}

echo "== Basic external_oidc validation =="
go run ./pkg/embeddable/examples/oidc \
  mcp start \
  --transport streamable_http \
  --port "${MCP_BASIC_PORT}" \
  --auth-mode external_oidc \
  --auth-resource-url "http://127.0.0.1:${MCP_BASIC_PORT}/mcp" \
  --oidc-issuer-url "${KEYCLOAK_ISSUER}" \
  >"${MCP_BASIC_LOG}" 2>&1 &
MCP_BASIC_PID=$!
wait_for_url "http://127.0.0.1:${MCP_BASIC_PORT}/.well-known/oauth-protected-resource" 60

curl -fsS "http://127.0.0.1:${MCP_BASIC_PORT}/.well-known/oauth-protected-resource" | jq -e \
  --arg issuer "${KEYCLOAK_ISSUER}" \
  --arg resource "http://127.0.0.1:${MCP_BASIC_PORT}/mcp" \
  '.authorization_servers[0] == $issuer and .resource == $resource' >/dev/null

assert_status 401 "http://127.0.0.1:${MCP_BASIC_PORT}/mcp" -X POST
assert_not_status 401 "http://127.0.0.1:${MCP_BASIC_PORT}/mcp" -X POST -H "Authorization: Bearer ${BASIC_TOKEN}"

echo "== Strict external_oidc validation =="
go run ./pkg/embeddable/examples/oidc \
  mcp start \
  --transport streamable_http \
  --port "${MCP_STRICT_PORT}" \
  --auth-mode external_oidc \
  --auth-resource-url "http://127.0.0.1:${MCP_STRICT_PORT}/mcp" \
  --oidc-issuer-url "${KEYCLOAK_ISSUER}" \
  --oidc-audience mcp-resource \
  --oidc-required-scope mcp-invoke \
  >"${MCP_STRICT_LOG}" 2>&1 &
MCP_STRICT_PID=$!
wait_for_url "http://127.0.0.1:${MCP_STRICT_PORT}/.well-known/oauth-protected-resource" 60

assert_status 401 "http://127.0.0.1:${MCP_STRICT_PORT}/mcp" -X POST -H "Authorization: Bearer ${BASIC_TOKEN}"
assert_not_status 401 "http://127.0.0.1:${MCP_STRICT_PORT}/mcp" -X POST -H "Authorization: Bearer ${STRICT_TOKEN}"

echo "Local Keycloak smoke passed."
echo "Keycloak issuer: ${KEYCLOAK_ISSUER}"
echo "Basic MCP log: ${MCP_BASIC_LOG}"
echo "Strict MCP log: ${MCP_STRICT_LOG}"
