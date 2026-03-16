---
Title: Recreate and verify hosted smailnail-mcp
Ticket: SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - coolify
    - keycloak
    - deployments
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: smailnail/Dockerfile
      Note: Defines the Coolify-facing container image used by the hosted MCP deployment
    - Path: smailnail/docs/deployments/smailnail-dovecot-coolify.md
      Note: Companion hosted IMAP target used by the hosted MCP deployment
    - Path: smailnail/docs/deployments/smailnail-imap-mcp-coolify.md
      Note: Contains the operator-facing deployment and routing contract for the hosted MCP service
ExternalSources: []
Summary: Exact commands and notes required to recreate the hosted smailnail-mcp deployment and understand how the public /mcp path is routed.
LastUpdated: 2026-03-16T05:00:00-04:00
WhatFor: Re-run the hosted MCP deployment from scratch and retrace the routing, Keycloak, Coolify, and verification steps without relying on memory.
WhenToUse: Use when reproducing the hosted MCP deployment or reviewing how the public HTTPS endpoint reaches the smailnail MCP binary.
---



# Recreate and verify hosted smailnail-mcp

## Routing model

`smailnail-imap-mcp` listens on container port `3201` and serves both:

- `/.well-known/oauth-protected-resource`
- `/mcp`

Coolify exposes the app on `https://smailnail.mcp.scapegoat.dev` and routes the whole hostname to port `3201`.
There is no proxy-side path rewrite for `/mcp`.

The effective mapping is:

- `https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource` -> container `:3201/.well-known/oauth-protected-resource`
- `https://smailnail.mcp.scapegoat.dev/mcp` -> container `:3201/mcp`

## Preconditions

- Public repo: `https://github.com/wesen/smailnail`
- Branch: `task/update-imap-mcp`
- Coolify dashboard/API: `https://hq.scapegoat.dev`
- Keycloak base: `https://auth.scapegoat.dev`
- Coolify server SSH: `root@89.167.52.236`
- Coolify API token is stored on the server in `~/.apitoken`
- Coolify CLI is installed on the server and configured with context `scapegoat`

## Repo-side deployment shape

Files that matter:

- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile.smailnail-imap-mcp`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnail-imap-mcp.sh`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnail-imap-mcp-coolify.md`

Container default behavior:

```text
smailnail-imap-mcp mcp start --transport streamable_http --port 3201
```

Required runtime env:

```env
SMAILNAIL_MCP_TRANSPORT=streamable_http
SMAILNAIL_MCP_PORT=3201
SMAILNAIL_MCP_AUTH_MODE=external_oidc
SMAILNAIL_MCP_AUTH_RESOURCE_URL=https://smailnail.mcp.scapegoat.dev/mcp
SMAILNAIL_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail
```

Important image detail:

- the runtime image must include `curl` or `wget`, because Coolify health checks run from inside the container

## Coolify context bootstrap

Verify the server-side CLI context:

```bash
ssh root@89.167.52.236 '
  export PATH=$PATH:/usr/local/bin:$HOME/go/bin
  coolify context verify --context scapegoat
  coolify server list --context scapegoat --format json
  coolify project list --context scapegoat --format json
'
```

Known IDs from this deployment:

- server UUID: `cgl105090ljoxitdf7gmvbrm`
- project UUID: `n8xkgqpbjj04m4pishy3su5e`
- environment name: `production`

## Keycloak setup

Keycloak service container:

- `keycloak-k12lm4blpo13louovn3pfsgs`

Log in with `kcadm.sh`:

```bash
ssh root@89.167.52.236 '
  docker exec keycloak-k12lm4blpo13louovn3pfsgs \
    /opt/keycloak/bin/kcadm.sh config credentials \
    --server http://127.0.0.1:8080 \
    --realm master \
    --user "$KEYCLOAK_ADMIN" \
    --password "$KEYCLOAK_ADMIN_PASSWORD"
'
```

Create the realm:

```bash
ssh root@89.167.52.236 '
  docker exec keycloak-k12lm4blpo13louovn3pfsgs \
    /opt/keycloak/bin/kcadm.sh create realms \
    -s realm=smailnail \
    -s enabled=true
'
```

Create the client using a JSON file to avoid quoting problems:

```json
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
```

Issuer used by the MCP app:

- `https://auth.scapegoat.dev/realms/smailnail`

## Create the Coolify app

Exact command shape that worked:

```bash
ssh root@89.167.52.236 '
  export PATH=$PATH:/usr/local/bin:$HOME/go/bin
  coolify app create public \
    --context scapegoat \
    --server-uuid cgl105090ljoxitdf7gmvbrm \
    --project-uuid n8xkgqpbjj04m4pishy3su5e \
    --environment-name production \
    --name smailnail-imap-mcp \
    --git-repository https://github.com/wesen/smailnail \
    --git-branch task/update-imap-mcp \
    --build-pack dockerfile \
    --ports-exposes 3201 \
    --domains https://smailnail.mcp.scapegoat.dev \
    --health-check-enabled \
    --health-check-path /.well-known/oauth-protected-resource \
    --format json
'
```

Important detail:

- `--domains` must be a full URL, not just the hostname

Known app UUID from this deployment:

- `fhp3mxqlfftdxdib3vxz89l3`

## Configure environment variables

The current Coolify CLI env helpers are not reliable against this host because they send `is_build_time`, while the server expects `is_buildtime`.

Working workaround:

```bash
TOKEN=$(cat ~/.apitoken)
APP_UUID=fhp3mxqlfftdxdib3vxz89l3

curl -fsS -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key":"SMAILNAIL_MCP_TRANSPORT","value":"streamable_http","is_runtime":true,"is_buildtime":true}' \
  "https://hq.scapegoat.dev/api/v1/applications/$APP_UUID/envs"
```

Repeat that payload shape for:

- `SMAILNAIL_MCP_PORT=3201`
- `SMAILNAIL_MCP_AUTH_MODE=external_oidc`
- `SMAILNAIL_MCP_AUTH_RESOURCE_URL=https://smailnail.mcp.scapegoat.dev/mcp`
- `SMAILNAIL_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail`

If duplicates are created, list and dedupe them:

```bash
TOKEN=$(cat ~/.apitoken)
APP_UUID=fhp3mxqlfftdxdib3vxz89l3
curl -fsS -H "Authorization: Bearer $TOKEN" \
  "https://hq.scapegoat.dev/api/v1/applications/$APP_UUID/envs" \
  | jq -c '.[] | {uuid, key, value}'
```

## Deploy

```bash
ssh root@89.167.52.236 '
  export PATH=$PATH:/usr/local/bin:$HOME/go/bin
  coolify deploy uuid fhp3mxqlfftdxdib3vxz89l3 --context scapegoat --force --format pretty
'
```

Check result:

```bash
ssh root@89.167.52.236 '
  export PATH=$PATH:/usr/local/bin:$HOME/go/bin
  coolify app get fhp3mxqlfftdxdib3vxz89l3 --context scapegoat --format pretty
'
```

Expected healthy status:

```text
running:healthy
```

## Verify live behavior

Protected resource metadata:

```bash
curl -fsS https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource | jq
```

Expected shape:

```json
{
  "authorization_servers": [
    "https://auth.scapegoat.dev/realms/smailnail"
  ],
  "resource": "https://smailnail.mcp.scapegoat.dev/mcp"
}
```

Unauthenticated MCP call:

```bash
curl -i \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' \
  https://smailnail.mcp.scapegoat.dev/mcp
```

Expected result:

- status `401`
- `WWW-Authenticate` bearer challenge
- body `missing bearer`

## Companion hosted Dovecot target

The hosted IMAP fixture now exists separately as Coolify service `gh32795yh1av2dpi2j6lhn6h`.

Current remote IMAP test target:

- host: `89.167.52.236`
- IMAPS port: `993`
- username: `a`
- password: `pass`
- TLS mode: self-signed, so use `--insecure`

Related operator doc:

- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnail-dovecot-coolify.md`

## What is still missing

The remaining gap is a documented end-to-end hosted MCP invocation that authenticates through Keycloak and then uses the hosted Dovecot target as the IMAP backend.
