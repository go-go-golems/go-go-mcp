#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_GO_MCP_ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp"

: "${KEYCLOAK_ADMIN_USER:=}"
: "${KEYCLOAK_ADMIN_PASSWORD:=}"
: "${SMAILNAIL_MCP_SMOKE_CLIENT_SECRET:=}"

if [[ -z "${SMAILNAIL_MCP_SMOKE_CLIENT_SECRET}" ]]; then
  if [[ -z "${KEYCLOAK_ADMIN_USER}" || -z "${KEYCLOAK_ADMIN_PASSWORD}" ]]; then
    echo "set SMAILNAIL_MCP_SMOKE_CLIENT_SECRET or export KEYCLOAK_ADMIN_USER and KEYCLOAK_ADMIN_PASSWORD" >&2
    exit 1
  fi

  CLIENT_JSON="$(ssh root@89.167.52.236 "docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh config credentials --server http://127.0.0.1:8080 --realm master --user '$KEYCLOAK_ADMIN_USER' --password '$KEYCLOAK_ADMIN_PASSWORD' >/dev/null && docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh get clients -r smailnail -q clientId=smailnail-mcp-smoke")"
  CLIENT_UUID="$(printf '%s\n' "$CLIENT_JSON" | jq -r '.[0].id')"
  if [[ -z "$CLIENT_UUID" || "$CLIENT_UUID" == "null" ]]; then
    echo "could not find Keycloak client smailnail-mcp-smoke" >&2
    exit 1
  fi

  SMAILNAIL_MCP_SMOKE_CLIENT_SECRET="$(ssh root@89.167.52.236 "docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh get clients/$CLIENT_UUID/client-secret -r smailnail" | jq -r '.value')"
  export SMAILNAIL_MCP_SMOKE_CLIENT_SECRET
fi

cd "$GO_GO_MCP_ROOT"
go run "$SCRIPT_DIR/smoke_hosted_mcp_oidc.go" "$@"
