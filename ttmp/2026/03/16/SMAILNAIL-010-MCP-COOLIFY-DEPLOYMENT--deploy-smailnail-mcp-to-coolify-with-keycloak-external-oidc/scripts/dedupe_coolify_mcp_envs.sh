#!/usr/bin/env bash
set -euo pipefail

APP_UUID="${APP_UUID:-fhp3mxqlfftdxdib3vxz89l3}"

cat <<'SCRIPT' >/tmp/remote-dedupe-coolify-mcp-envs.sh
#!/usr/bin/env bash
set -euo pipefail
TOKEN=$(cat ~/.apitoken)
APP_UUID="${APP_UUID:-fhp3mxqlfftdxdib3vxz89l3}"
json=$(curl -fsS -H "Authorization: Bearer $TOKEN" "https://hq.scapegoat.dev/api/v1/applications/$APP_UUID/envs")
echo "$json" | jq -r 'group_by(.key)[] | .[1:][]?.uuid' | while read -r env_uuid; do
  [ -n "$env_uuid" ] || continue
  curl -fsS -X DELETE \
    -H "Authorization: Bearer $TOKEN" \
    "https://hq.scapegoat.dev/api/v1/applications/$APP_UUID/envs/$env_uuid" >/dev/null
  echo "deleted $env_uuid"
done
echo "--- remaining ---"
curl -fsS -H "Authorization: Bearer $TOKEN" \
  "https://hq.scapegoat.dev/api/v1/applications/$APP_UUID/envs" \
  | jq -c '.[] | {uuid, key, value, is_runtime, is_buildtime}'
SCRIPT

scp /tmp/remote-dedupe-coolify-mcp-envs.sh root@89.167.52.236:/tmp/remote-dedupe-coolify-mcp-envs.sh >/dev/null
ssh root@89.167.52.236 "APP_UUID='$APP_UUID' bash /tmp/remote-dedupe-coolify-mcp-envs.sh && rm -f /tmp/remote-dedupe-coolify-mcp-envs.sh"
rm -f /tmp/remote-dedupe-coolify-mcp-envs.sh
