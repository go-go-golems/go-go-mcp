#!/usr/bin/env bash
set -euo pipefail

APP_UUID="${APP_UUID:-fhp3mxqlfftdxdib3vxz89l3}"

cat <<'SCRIPT' >/tmp/remote-set-coolify-mcp-envs.sh
#!/usr/bin/env bash
set -euo pipefail
TOKEN=$(cat ~/.apitoken)
APP_UUID="${APP_UUID:-fhp3mxqlfftdxdib3vxz89l3}"
cat > /tmp/smailnail-mcp-envs.json <<'EOF'
[
  {"key":"SMAILNAIL_MCP_TRANSPORT","value":"streamable_http"},
  {"key":"SMAILNAIL_MCP_PORT","value":"3201"},
  {"key":"SMAILNAIL_MCP_AUTH_MODE","value":"external_oidc"},
  {"key":"SMAILNAIL_MCP_AUTH_RESOURCE_URL","value":"https://smailnail.mcp.scapegoat.dev/mcp"},
  {"key":"SMAILNAIL_MCP_OIDC_ISSUER_URL","value":"https://auth.scapegoat.dev/realms/smailnail"}
]
EOF

jq -c '.[]' /tmp/smailnail-mcp-envs.json | while read -r item; do
  payload=$(echo "$item" | jq -c '. + {is_runtime:true,is_buildtime:true}')
  curl -fsS -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "https://hq.scapegoat.dev/api/v1/applications/$APP_UUID/envs"
  echo
done

curl -fsS -H "Authorization: Bearer $TOKEN" \
  "https://hq.scapegoat.dev/api/v1/applications/$APP_UUID/envs" \
  | jq -c '.[] | {uuid, key, value, is_runtime, is_buildtime}'

rm -f /tmp/smailnail-mcp-envs.json
SCRIPT

scp /tmp/remote-set-coolify-mcp-envs.sh root@89.167.52.236:/tmp/remote-set-coolify-mcp-envs.sh >/dev/null
ssh root@89.167.52.236 "APP_UUID='$APP_UUID' bash /tmp/remote-set-coolify-mcp-envs.sh && rm -f /tmp/remote-set-coolify-mcp-envs.sh"
rm -f /tmp/remote-set-coolify-mcp-envs.sh
