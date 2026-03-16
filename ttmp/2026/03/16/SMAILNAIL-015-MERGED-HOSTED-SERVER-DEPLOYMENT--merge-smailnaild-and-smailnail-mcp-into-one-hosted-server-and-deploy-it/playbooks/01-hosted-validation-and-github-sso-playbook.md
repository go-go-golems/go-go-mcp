---
Title: Hosted validation and GitHub SSO playbook
Ticket: SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - oidc
    - keycloak
    - github
    - coolify
    - deployment
    - playbook
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go
      Note: Hosted account-test retry hardening and read-only IMAP probe logic
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service_retry_test.go
      Note: Unit coverage for retrying transient closed-connection failures
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT--merge-smailnaild-and-smailnail-mcp-into-one-hosted-server-and-deploy-it/scripts/smoke_hosted_web_login.sh
      Note: Browser-session login smoke used against the hosted merged server
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT--merge-smailnaild-and-smailnail-mcp-into-one-hosted-server-and-deploy-it/scripts/smoke_merged_server_mcp.go
      Note: Bearer-authenticated merged MCP smoke using a stored accountId
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md
      Note: Merged hosted deployment guide and runtime variable surface
Summary: Operator playbook for validating the live merged smailnail host after deploy and then enabling GitHub SSO in the production Keycloak realm.
LastUpdated: 2026-03-16T17:41:00-04:00
WhatFor: Give an operator one reproducible checklist for proving the hosted server works after deploy and then wiring GitHub as a Keycloak identity provider without changing the smailnail auth model.
WhenToUse: Use after a merged-host deployment, after account-test changes, or when preparing GitHub-backed login in the production Keycloak realm.
---

# Hosted validation and GitHub SSO playbook

## Goal

This playbook has two purposes:

1. validate that the live merged host still works end to end after a deploy
2. enable GitHub login in the production Keycloak realm at `auth.scapegoat.dev`

The important architecture point is that GitHub SSO is introduced at the Keycloak layer, not by changing `smailnaild` into a GitHub-specific app. `smailnaild` and `/mcp` should continue to trust the same OIDC issuer:

- issuer: `https://auth.scapegoat.dev/realms/smailnail`

The difference after the change is just that the human user can choose GitHub inside Keycloak.

## Current hosted shape

The merged host is:

- app + MCP host: `https://smailnail.mcp.scapegoat.dev`
- browser login entry: `https://smailnail.mcp.scapegoat.dev/auth/login`
- browser callback: `https://smailnail.mcp.scapegoat.dev/auth/callback`
- MCP endpoint: `https://smailnail.mcp.scapegoat.dev/mcp`
- MCP metadata: `https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource`

The production issuer is:

- `https://auth.scapegoat.dev/realms/smailnail`

Current caveat:

- the app database is still container-local SQLite at `/app/smailnaild.sqlite`
- any redeploy clears saved hosted accounts

That matters for validation. After a redeploy, you may need to recreate the hosted IMAP account before testing `/api/accounts/{id}/test` or `/mcp` against a stored account.

## Part 1: hosted validation after deploy

### 1. Check the merged host is alive

```bash
curl -sS https://smailnail.mcp.scapegoat.dev/readyz | jq
curl -sS https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource | jq
curl -i -sS \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' \
  https://smailnail.mcp.scapegoat.dev/mcp | sed -n '1,40p'
```

Expected:

- `/readyz` returns `{"status":"ready"}`
- protected-resource metadata points at the Keycloak realm and `/mcp`
- unauthenticated `/mcp` returns `401` plus `WWW-Authenticate`

### 2. Log in through the hosted browser flow

Use the ticket script:

```bash
bash /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT--merge-smailnaild-and-smailnail-mcp-into-one-hosted-server-and-deploy-it/scripts/smoke_hosted_web_login.sh
```

Defaults:

- username: `alice`
- password: `secret`

The script writes:

- `/tmp/smailnail-hosted.cookies`
- `/tmp/smailnail-hosted-login.html`
- `/tmp/smailnail-hosted-final.html`

and prints the `/api/me` result.

Expected:

- a `smailnail_session` cookie exists
- `/api/me` returns the provisioned user JSON

### 3. Recreate the hosted IMAP account if the app was redeployed

Because the DB is not persistent yet, run this after any deploy that replaced the container:

```bash
curl -sS \
  -b /tmp/smailnail-hosted.cookies \
  -c /tmp/smailnail-hosted.cookies \
  -H 'Content-Type: application/json' \
  -X POST \
  https://smailnail.mcp.scapegoat.dev/api/accounts \
  -d '{
    "label":"Hosted Dovecot Fixture",
    "providerHint":"coolify-fixture",
    "server":"89.167.52.236",
    "port":993,
    "username":"a",
    "password":"pass",
    "mailboxDefault":"INBOX",
    "insecure":true,
    "authKind":"password",
    "isDefault":true,
    "mcpEnabled":true
  }' | jq
```

Save the returned `data.id`. That is the live `accountId`.

### 4. Validate the hosted account test path

```bash
ACCOUNT_ID='<replace-with-live-account-id>'

curl -sS \
  -b /tmp/smailnail-hosted.cookies \
  -H 'Content-Type: application/json' \
  -X POST \
  "https://smailnail.mcp.scapegoat.dev/api/accounts/${ACCOUNT_ID}/test" \
  -d '{"mode":"read_only"}' | jq
```

Expected:

- `success: true`
- `mailboxSelectOk: true`
- `listOk: true`
- `sampleFetchOk: true`

If this intermittently fails with `use of closed network connection`, note that the account-test code now retries one transient closed-connection failure. If repeated failures return again, inspect the live merged container logs and the hosted Dovecot fixture.

### 5. Validate the merged MCP path against the same stored account

Use the ticket smoke program:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp

SMAILNAIL_SERVER_URL=https://smailnail.mcp.scapegoat.dev/mcp \
SMAILNAIL_TOKEN_URL=https://auth.scapegoat.dev/realms/smailnail/protocol/openid-connect/token \
SMAILNAIL_CLIENT_ID=smailnail-mcp-test \
SMAILNAIL_USERNAME=alice \
SMAILNAIL_PASSWORD=secret \
SMAILNAIL_ACCOUNT_ID="$ACCOUNT_ID" \
go run ./ttmp/2026/03/16/SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT--merge-smailnaild-and-smailnail-mcp-into-one-hosted-server-and-deploy-it/scripts/smoke_merged_server_mcp.go
```

Expected:

```json
{
  "value": {
    "mailbox": "INBOX"
  }
}
```

That is the proof that:

- the browser user and the bearer-token user are the same Keycloak identity
- the stored account belongs to that local user
- `/mcp` can use that stored account successfully

## Part 2: enable GitHub SSO in production Keycloak

### Why this works without changing smailnaild

`smailnaild` already relies on Keycloak as its OIDC issuer. That should stay true. GitHub should be configured as an upstream identity provider in Keycloak, not as a direct login integration inside `smailnaild`.

So the flow becomes:

```text
browser -> smailnaild /auth/login
        -> Keycloak login screen
        -> GitHub button
        -> GitHub OAuth
        -> Keycloak brokered identity
        -> smailnaild /auth/callback
```

`smailnaild` still only needs to understand:

- issuer
- subject
- session cookie

not any GitHub-specific token format.

### 1. Create the GitHub OAuth app

In GitHub Developer Settings, create a new OAuth App.

Use:

- application name: `smailnail auth`
- homepage URL: `https://smailnail.mcp.scapegoat.dev`
- authorization callback URL:
  `https://auth.scapegoat.dev/realms/smailnail/broker/github/endpoint`

Save:

- client ID
- client secret

### 2. Add GitHub as an identity provider in Keycloak

In the `smailnail` realm on `auth.scapegoat.dev`:

1. go to `Identity Providers`
2. add provider `GitHub`
3. paste the GitHub client ID and client secret
4. configure scopes:
   - `read:user user:email`

Suggested initial settings:

- enabled: `true`
- store token: `false`
- trust email: `true` only if you are comfortable relying on GitHub email identity
- first login flow: default `first broker login`

### 3. Keep the existing OIDC clients unchanged at first

Do not change these yet:

- `smailnail-web`
- `smailnail-mcp`
- `smailnail-mcp-test`

The purpose of the first GitHub step is just to add a new login option at the Keycloak layer. The app and MCP clients should continue to trust the same realm issuer.

### 4. Test GitHub login on Keycloak before touching the app

Open the production Keycloak realm login page and confirm:

- the GitHub button appears
- GitHub login succeeds
- Keycloak creates or links the user
- the resulting Keycloak user has the expected profile and email

Do this before testing `smailnaild`. It keeps “Keycloak broker config broken” separate from “app callback broken”.

### 5. Test GitHub login through smailnaild

After Keycloak-side success:

1. clear browser cookies for `smailnail.mcp.scapegoat.dev`
2. open:
   - `https://smailnail.mcp.scapegoat.dev/auth/login`
3. click `GitHub`
4. complete GitHub auth
5. confirm `/api/me` returns the expected brokered user

At this point, `smailnaild` should behave exactly the same as with password login, because the issuer is still the same Keycloak realm.

### 6. What to verify in the local user model

After the first GitHub login, verify that the app still maps the browser user through:

- issuer: `https://auth.scapegoat.dev/realms/smailnail`
- subject: Keycloak `sub`

Do not key local user data off:

- GitHub username
- GitHub numeric ID
- email alone

GitHub is now just the upstream login method. The stable app identity should still be the Keycloak-issued `(issuer, sub)` pair.

### 7. Recommended post-setup checks

After GitHub SSO works:

- confirm old local-password test login still works if you want to keep it
- decide whether to disable password login for normal users
- test that the same GitHub-authenticated user can:
  - reach `/api/me`
  - save an IMAP account
  - use `/mcp` if you mint a bearer token through the same realm

## Known caveats and next likely work

### Persistence is still the main operational gap

The merged host currently loses saved accounts on redeploy because it still stores app data in container-local SQLite.

Before depending on GitHub SSO for real users, the next important infrastructure fix is:

- persistent volume for SQLite, or
- Postgres via the Clay SQL path

### Temporary MCP smoke client

The client:

- `smailnail-mcp-test`

exists to make direct bearer-token smoke testing easier. It is useful now, but it should be reviewed later and either:

- kept as an explicit smoke-only client, or
- removed once the preferred operator smoke flow is settled

## Short checklist

- verify `/readyz`
- verify protected-resource metadata
- run [smoke_hosted_web_login.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT--merge-smailnaild-and-smailnail-mcp-into-one-hosted-server-and-deploy-it/scripts/smoke_hosted_web_login.sh)
- recreate hosted IMAP account if the app was redeployed
- run hosted `/api/accounts/{id}/test`
- run [smoke_merged_server_mcp.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT--merge-smailnaild-and-smailnail-mcp-into-one-hosted-server-and-deploy-it/scripts/smoke_merged_server_mcp.go)
- create GitHub OAuth app with callback `https://auth.scapegoat.dev/realms/smailnail/broker/github/endpoint`
- add GitHub provider in Keycloak
- test Keycloak-side GitHub login
- test `smailnaild` login through GitHub
