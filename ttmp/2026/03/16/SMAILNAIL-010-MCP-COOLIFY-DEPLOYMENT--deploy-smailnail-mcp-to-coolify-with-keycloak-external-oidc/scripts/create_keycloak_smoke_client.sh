#!/usr/bin/env bash
set -euo pipefail

: "${KEYCLOAK_ADMIN_USER:?set KEYCLOAK_ADMIN_USER}"
: "${KEYCLOAK_ADMIN_PASSWORD:?set KEYCLOAK_ADMIN_PASSWORD}"

cat <<'EOF' >/tmp/smailnail-mcp-smoke-client.json
{
  "clientId": "smailnail-mcp-smoke",
  "enabled": true,
  "protocol": "openid-connect",
  "publicClient": false,
  "standardFlowEnabled": false,
  "directAccessGrantsEnabled": false,
  "serviceAccountsEnabled": true,
  "clientAuthenticatorType": "client-secret"
}
EOF

scp /tmp/smailnail-mcp-smoke-client.json root@89.167.52.236:/tmp/smailnail-mcp-smoke-client.json >/dev/null
ssh root@89.167.52.236 "docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh config credentials --server http://127.0.0.1:8080 --realm master --user '$KEYCLOAK_ADMIN_USER' --password '$KEYCLOAK_ADMIN_PASSWORD' >/dev/null && docker cp /tmp/smailnail-mcp-smoke-client.json keycloak-k12lm4blpo13louovn3pfsgs:/tmp/smailnail-mcp-smoke-client.json && docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh create clients -r smailnail -f /tmp/smailnail-mcp-smoke-client.json || true && docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh get clients -r smailnail -q clientId=smailnail-mcp-smoke"
rm -f /tmp/smailnail-mcp-smoke-client.json
