---
Title: Local Keycloak external OIDC testing playbook
Ticket: MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN
Status: active
Topics:
    - mcp
    - oidc
    - keycloak
    - authentication
    - documentation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-mcp/pkg/embeddable/command.go
      Note: Explicit external and embedded auth flags used by this playbook
    - Path: go-go-mcp/pkg/embeddable/auth_provider_external.go
      Note: External discovery, JWKS, JWT, audience, and scope validation logic exercised by this playbook
    - Path: go-go-mcp/pkg/embeddable/examples/oidc/main.go
      Note: Runnable example server used as the local protected resource
    - Path: go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/keycloak-local/docker-compose.yml
      Note: Ticket-local Keycloak Compose stack for this playbook
    - Path: go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/keycloak-local/import/mcp-local-realm.json
      Note: Imported test realm with basic and strict service-account clients
    - Path: go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/local-keycloak-smoke.sh
      Note: End-to-end automation for the playbook
Summary: Step-by-step local setup and verification guide for testing go-go-mcp in external_oidc mode against a Docker Compose Keycloak instance.
LastUpdated: 2026-03-09T19:23:09-04:00
WhatFor: Give a new engineer a reproducible local workflow for standing up Keycloak, minting tokens, running go-go-mcp against an external issuer, and validating both lenient and strict auth modes.
WhenToUse: Use when you want to validate external_oidc locally, debug Keycloak issuer integration, or verify audience and scope enforcement before deploying.
---

# Local Keycloak external OIDC testing playbook

## Goal

This playbook gives you a local, repeatable way to test `go-go-mcp` in `external_oidc` mode against a real Keycloak instance running in Docker Compose. It does not rely on a hosted Keycloak, a manually configured realm, or hand-edited client settings in the UI. The ticket includes all of the local assets needed to stand up Keycloak, import a test realm, mint tokens, and verify the MCP auth boundary.

The local setup deliberately tests two cases:

- a basic external issuer path where `go-go-mcp` validates an externally issued bearer token without additional audience or scope checks
- a strict path where `go-go-mcp` enforces both `--oidc-audience mcp-resource` and `--oidc-required-scope mcp-invoke`

That split matters because the first case proves the external issuer integration works at all, while the second proves the extra policy gates are actually enforced.

## What the ticket contains

- Compose stack:
  - [docker-compose.yml](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/keycloak-local/docker-compose.yml)
- Imported realm:
  - [mcp-local-realm.json](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/keycloak-local/import/mcp-local-realm.json)
- Automated smoke harness:
  - [local-keycloak-smoke.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/local-keycloak-smoke.sh)

## Preconditions

You need:

- Docker
- Docker Compose
- `curl`
- `jq`
- Go toolchain

This was validated in the current workspace on:

- Docker `25.0.2`
- Docker Compose `v2.27.0`
- `jq 1.7`
- `curl 8.5.0`

## Imported realm layout

The imported realm is named `mcp-local` and contains two confidential service-account clients:

- `mcp-cli-basic`
  - secret: `mcp-cli-basic-secret`
  - purpose: basic external issuer validation without required audience or scope gates
- `mcp-cli-strict`
  - secret: `mcp-cli-strict-secret`
  - purpose: strict validation with both:
    - audience `mcp-resource`
    - scope `mcp-invoke`

The strict client uses imported Keycloak client scopes so the token arrives with the extra audience and scope values needed for the stricter `go-go-mcp` flags.

## Fast path: run the automated smoke

The quickest validation path is:

```bash
bash /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/local-keycloak-smoke.sh
```

Expected result:

- Keycloak comes up in Docker Compose
- the imported realm becomes reachable
- the script mints a basic token and a strict token
- `go-go-mcp` passes the basic external issuer test
- `go-go-mcp` rejects the basic token in strict mode
- `go-go-mcp` accepts the strict token in strict mode

If the script prints `Local Keycloak smoke passed.`, the local playbook is working.

## Manual setup

If you want to inspect each step manually instead of running the harness, use this sequence.

### 1. Start Keycloak

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/keycloak-local
KEYCLOAK_PORT=18080 docker compose up -d
```

Wait until discovery is reachable:

```bash
curl -s http://127.0.0.1:18080/realms/mcp-local/.well-known/openid-configuration | jq
```

Expected fields:

- `issuer`
- `jwks_uri`
- `token_endpoint`

### 2. Mint a basic token

```bash
BASIC_TOKEN=$(
  curl -s \
    -X POST http://127.0.0.1:18080/realms/mcp-local/protocol/openid-connect/token \
    -d grant_type=client_credentials \
    -d client_id=mcp-cli-basic \
    -d client_secret=mcp-cli-basic-secret \
  | jq -r .access_token
)
```

Sanity check:

```bash
test -n "$BASIC_TOKEN" && test "$BASIC_TOKEN" != "null"
```

### 3. Run `go-go-mcp` in basic external mode

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp

go run ./pkg/embeddable/examples/oidc \
  mcp start \
  --transport streamable_http \
  --port 3001 \
  --auth-mode external_oidc \
  --auth-resource-url http://127.0.0.1:3001/mcp \
  --oidc-issuer-url http://127.0.0.1:18080/realms/mcp-local
```

### 4. Verify unauthenticated vs authenticated behavior

Protected-resource metadata:

```bash
curl -s http://127.0.0.1:3001/.well-known/oauth-protected-resource | jq
```

Expected:

- `authorization_servers[0] == "http://127.0.0.1:18080/realms/mcp-local"`
- `resource == "http://127.0.0.1:3001/mcp"`

Unauthenticated request:

```bash
curl -i -X POST http://127.0.0.1:3001/mcp
```

Expected:

- `401 Unauthorized`
- `Www-Authenticate` header present

Authenticated request:

```bash
curl -i \
  -X POST \
  -H "Authorization: Bearer $BASIC_TOKEN" \
  http://127.0.0.1:3001/mcp
```

Expected:

- not `401`
- typically `400` because this is only an auth-boundary probe, not a full MCP session exchange

That non-`401` result is the important part. It proves Keycloak-issued JWT validation succeeded and the request made it past the auth middleware into the transport layer.

### 5. Mint a strict token

```bash
STRICT_TOKEN=$(
  curl -s \
    -X POST http://127.0.0.1:18080/realms/mcp-local/protocol/openid-connect/token \
    -d grant_type=client_credentials \
    -d client_id=mcp-cli-strict \
    -d client_secret=mcp-cli-strict-secret \
  | jq -r .access_token
)
```

Inspect the payload if you want:

```bash
echo "$STRICT_TOKEN" | awk -F. '{print $2}' | tr '_-' '/+' | base64 -d 2>/dev/null | jq
```

You should see claims that include:

- `aud` containing `mcp-resource`
- `scope` containing `mcp-invoke`

### 6. Run `go-go-mcp` in strict external mode

Use a different port so you can keep the earlier process separate if you want:

```bash
go run ./pkg/embeddable/examples/oidc \
  mcp start \
  --transport streamable_http \
  --port 3002 \
  --auth-mode external_oidc \
  --auth-resource-url http://127.0.0.1:3002/mcp \
  --oidc-issuer-url http://127.0.0.1:18080/realms/mcp-local \
  --oidc-audience mcp-resource \
  --oidc-required-scope mcp-invoke
```

### 7. Verify strict failure and strict success

Basic token should now fail:

```bash
curl -i \
  -X POST \
  -H "Authorization: Bearer $BASIC_TOKEN" \
  http://127.0.0.1:3002/mcp
```

Expected:

- `401 Unauthorized`

Strict token should pass auth:

```bash
curl -i \
  -X POST \
  -H "Authorization: Bearer $STRICT_TOKEN" \
  http://127.0.0.1:3002/mcp
```

Expected:

- not `401`
- usually `400` because, again, the request is only probing the auth gate

## What this does not test

This playbook validates the auth boundary and metadata behavior. It does not perform a full MCP client handshake and tool session over streamable HTTP. That is intentional. For auth integration work, the high-signal question is whether the bearer token gets you past the auth layer. Once the request reaches the transport and changes from `401` to a transport-level response, the auth integration has done its job.

## Teardown

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/keycloak-local
KEYCLOAK_PORT=18080 docker compose down -v
```

## What you need to do to test it fully

If you want the exact full validation I ran, do this:

1. Run the automated smoke script.
2. If it passes, rerun the manual strict-mode steps once so you can inspect the actual claims and HTTP headers yourself.
3. If you want to go deeper, change one thing at a time:
   - wrong issuer URL
   - wrong audience
   - wrong required scope
   - wrong client secret
4. Confirm the failure mode changes to `401` again.

That gives you both the happy path and the main policy-failure paths locally.
