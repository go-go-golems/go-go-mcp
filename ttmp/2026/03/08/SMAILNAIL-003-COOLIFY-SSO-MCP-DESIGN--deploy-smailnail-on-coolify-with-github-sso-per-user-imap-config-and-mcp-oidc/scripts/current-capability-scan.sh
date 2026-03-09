#!/usr/bin/env bash
set -euo pipefail

WORKSPACE_ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp"
SMAILNAIL_ROOT="$WORKSPACE_ROOT/smailnail"
GO_GO_MCP_ROOT="$WORKSPACE_ROOT/go-go-mcp"

echo "== smailnail command roots =="
rg -n "^func main\\(" "$SMAILNAIL_ROOT/cmd" -S

echo
echo "== smailnail auth/http/db capability scan =="
rg -n "http\\.ListenAndServe|chi|gin|echo|fiber|oidc|oauth|session|cookie|sqlite|postgres|gorm|sqlx" \
  "$SMAILNAIL_ROOT/cmd" "$SMAILNAIL_ROOT/pkg" "$SMAILNAIL_ROOT/go.mod" -S || true

echo
echo "== direct IMAP credential entry points =="
rg -n "Password|ConnectToIMAPServer|StoreIMAP|store-imap" \
  "$SMAILNAIL_ROOT/cmd" "$SMAILNAIL_ROOT/pkg" -S

echo
echo "== go-go-mcp OIDC and HTTP transport reuse points =="
rg -n "WithOIDC|oauth-protected-resource|openid-configuration|streamable_http|NewBackend" \
  "$GO_GO_MCP_ROOT/pkg" "$GO_GO_MCP_ROOT/cmd" -S
