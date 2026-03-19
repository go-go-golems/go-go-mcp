#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${SMAILNAIL_BASE_URL:-https://smailnail.mcp.scapegoat.dev}"
USERNAME="${SMAILNAIL_USERNAME:-alice}"
PASSWORD="${SMAILNAIL_PASSWORD:-secret}"
COOKIE_JAR="${SMAILNAIL_COOKIE_JAR:-/tmp/smailnail-hosted.cookies}"
LOGIN_HTML="${SMAILNAIL_LOGIN_HTML:-/tmp/smailnail-hosted-login.html}"
FINAL_HTML="${SMAILNAIL_FINAL_HTML:-/tmp/smailnail-hosted-final.html}"

rm -f "$COOKIE_JAR" "$LOGIN_HTML" "$FINAL_HTML"

login_location="$(
  curl -sS -D - -o /dev/null -c "$COOKIE_JAR" "$BASE_URL/auth/login" |
    awk 'tolower($1)=="location:" {print $2}' |
    tr -d '\r'
)"

if [ -z "$login_location" ]; then
  echo "missing login redirect" >&2
  exit 1
fi

curl -sS -b "$COOKIE_JAR" -c "$COOKIE_JAR" "$login_location" >"$LOGIN_HTML"

action_url="$(
  LOGIN_HTML="$LOGIN_HTML" python - <<'PY'
import html
import os
import re
from pathlib import Path

raw = Path(os.environ["LOGIN_HTML"]).read_text()
match = re.search(r'<form[^>]+id="kc-form-login"[^>]+action="([^"]+)"', raw)
if not match:
    raise SystemExit("missing Keycloak login form action")
print(html.unescape(match.group(1)))
PY
)"

if [ -z "$action_url" ]; then
  echo "missing form action" >&2
  exit 1
fi

curl -sS -L -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode "username=$USERNAME" \
  --data-urlencode "password=$PASSWORD" \
  --data-urlencode 'credentialId=' \
  "$action_url" >"$FINAL_HTML"

echo "SESSION_COOKIES:"
cat "$COOKIE_JAR"

echo
echo "API_ME:"
curl -sS -b "$COOKIE_JAR" "$BASE_URL/api/me" | jq
