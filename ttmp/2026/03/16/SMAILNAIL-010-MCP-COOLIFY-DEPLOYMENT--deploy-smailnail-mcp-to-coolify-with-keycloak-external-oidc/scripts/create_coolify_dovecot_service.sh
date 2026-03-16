#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/coolify/smailnail-dovecot.compose.yaml"
COMPOSE_B64="$(base64 -w0 "$COMPOSE_FILE")"

ssh root@89.167.52.236 "TOKEN=\$(cat ~/.apitoken); curl -fsS -X POST \
  -H 'Authorization: Bearer '"\""\$TOKEN"\""' \
  -H 'Content-Type: application/json' \
  -d @- https://hq.scapegoat.dev/api/v1/services" <<EOF
{
  "project_uuid": "n8xkgqpbjj04m4pishy3su5e",
  "environment_name": "production",
  "server_uuid": "cgl105090ljoxitdf7gmvbrm",
  "name": "smailnail-dovecot-fixture",
  "description": "Hosted docker-test-dovecot fixture for remote smailnail MCP validation",
  "instant_deploy": true,
  "docker_compose_raw": "$COMPOSE_B64"
}
EOF
