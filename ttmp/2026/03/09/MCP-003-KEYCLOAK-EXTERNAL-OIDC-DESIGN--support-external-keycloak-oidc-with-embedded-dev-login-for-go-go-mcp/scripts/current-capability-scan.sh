#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp"

cd "$ROOT"

echo "==> Auth and embeddable auth files"
find pkg/auth pkg/embeddable cmd/go-go-mcp/cmds -maxdepth 3 -type f | sort | grep -E '(oidc|auth|mcpgo_backend|command\.go|server\.go)$' || true

echo
echo "==> OIDC/auth symbol hits"
rg -n "WithOIDC|oidcEnabled|IntrospectAccessToken|oauth-protected-resource|WWW-Authenticate|Authenticator|oidc-auth-key|oidc-user|oidc-pass" \
  pkg/auth pkg/embeddable cmd/go-go-mcp/cmds pkg/doc/topics || true

echo
echo "==> Embedded OIDC admin CLI"
go run ./cmd/go-go-mcp oidc --help || true

echo
echo "==> Embeddable MCP start flags (OIDC section)"
go run ./pkg/embeddable/examples/oidc mcp start --help | sed -n '1,220p' || true
