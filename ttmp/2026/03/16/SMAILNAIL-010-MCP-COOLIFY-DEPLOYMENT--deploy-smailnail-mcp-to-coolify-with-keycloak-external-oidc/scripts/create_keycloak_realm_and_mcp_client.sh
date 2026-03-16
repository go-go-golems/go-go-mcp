#!/usr/bin/env bash
set -euo pipefail

: "${KEYCLOAK_ADMIN_USER:?set KEYCLOAK_ADMIN_USER}"
: "${KEYCLOAK_ADMIN_PASSWORD:?set KEYCLOAK_ADMIN_PASSWORD}"

cat <<'EOF' >/tmp/smailnail-mcp-client.json
{
  "clientId": "smailnail-mcp",
  "enabled": true,
  "publicClient": true,
  "protocol": "openid-connect",
  "standardFlowEnabled": true,
  "directAccessGrantsEnabled": false,
  "serviceAccountsEnabled": false,
  "redirectUris": [
    "https://claude.ai/api/mcp/auth_callback",
    "https://claude.com/api/mcp/auth_callback",
    "https://smailnail.mcp.scapegoat.dev/*"
  ],
  "webOrigins": ["+"]
}
EOF

scp /tmp/smailnail-mcp-client.json root@89.167.52.236:/tmp/smailnail-mcp-client.json >/dev/null
ssh root@89.167.52.236 "docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh config credentials --server http://127.0.0.1:8080 --realm master --user '$KEYCLOAK_ADMIN_USER' --password '$KEYCLOAK_ADMIN_PASSWORD' >/dev/null && docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh create realms -s realm=smailnail -s enabled=true || true && docker cp /tmp/smailnail-mcp-client.json keycloak-k12lm4blpo13louovn3pfsgs:/tmp/smailnail-mcp-client.json && docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh create clients -r smailnail -f /tmp/smailnail-mcp-client.json || true && docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh get clients -r smailnail -q clientId=smailnail-mcp"
rm -f /tmp/smailnail-mcp-client.json
