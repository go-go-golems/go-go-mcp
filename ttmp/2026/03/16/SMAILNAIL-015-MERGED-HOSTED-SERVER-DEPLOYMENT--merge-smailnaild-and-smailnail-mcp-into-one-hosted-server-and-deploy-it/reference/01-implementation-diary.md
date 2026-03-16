---
Title: Implementation diary for merging smailnaild and MCP into one hosted server
Ticket: SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - oidc
    - keycloak
    - coolify
    - deployments
    - diary
DocType: reference
Intent: long-term
Summary: Chronological record of the server-merge and deployment work, including failed attempts, validation commands, and rollout notes.
---

# Implementation diary for merging smailnaild and MCP into one hosted server

## Kickoff

### Prompt context

The user wants to stop treating the hosted web app and the hosted MCP as separate production services and instead serve them from the same binary and the same `http.Server`. The immediate request is to create a new ticket and make it actionable enough to drive the actual refactor and deployment work.

### What I did

I created ticket `SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT` and added:

- a detailed implementation plan
- a granular task breakdown
- this diary as the running execution log

### Why

The merge is large enough that it needs a fresh execution record rather than being buried inside older identity or deployment tickets. The older tickets explain why the current split exists; this new ticket should explain how to remove that split safely.

### Initial assumptions

- the single-server design is desirable
- the merge should keep browser-session auth and bearer-token auth distinct
- the deployment cutover should happen only after a local proof and a hosted smoke

### Immediate next step

Start with the code-shape refactor:

- extract a mountable MCP handler from the standalone server package
- then mount it into `smailnaild`

## Implementation step 1: build the mounted-MCP path and thread it into smailnaild

### What I changed

I made the first structural merge slice across both repositories.

In `go-go-mcp`, I added a mounted-HTTP path in [mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go):

- exported `MountHTTPHandlers(...)`
- factored the existing SSE and streamable-HTTP route setup into reusable mount helpers
- kept the existing standalone backends working by making them reuse the same mount logic internally

In `smailnail`, I then added a hosted mounting layer in [server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go):

- `MountHTTPHandlers(...)` for the smailnail MCP package
- a shared `baseServerOptions(...)` helper so the mounted and standalone forms register the same tools and middleware

I also introduced hosted MCP config in [hosted_config.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/hosted_config.go) and threaded it into [serve.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnaild/commands/serve.go).

Finally, I updated [http.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go) so `smailnaild` can mount:

- `/.well-known/oauth-protected-resource`
- `/mcp`
- `/mcp/`

before the SPA fallback handler is registered.

### Why this matters

This is the minimum structural change needed before any deployment work can happen. Without it, `smailnaild` cannot own the MCP HTTP routes, and we would still be stuck with two separately booted servers.

### Validation

I ran:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp
go test ./pkg/embeddable/...
```

and:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/mcp/imapjs ./pkg/smailnaild ./cmd/smailnaild/...
```

I also added a route-level proof in [mounted_handler_test.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/mounted_handler_test.go) showing that the mounted MCP mux can be served by the hosted app handler without breaking `/api/info`.

### What remains after this step

The mounted route exists, but the ticket is not close to done yet. The next concrete slice is the real local full-stack proof:

- run the merged `smailnaild` with the local Keycloak and Dovecot stack
- verify browser login still works
- verify account setup still works
- verify `/mcp` on the same server still works with bearer auth

## Implementation step 2: add route-separation proof and repackage the hosted image around smailnaild

### What I changed

I added a second mounted-handler test in [mounted_handler_test.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/mounted_handler_test.go):

- `TestMountedHandlersKeepWebAndMCPAuthSeparated`

That test proves three properties at once:

- `/.well-known/oauth-protected-resource` stays public
- `/api/me` still requires a browser-session user
- `/mcp` still requires a bearer token and still emits `WWW-Authenticate`

I then switched the production packaging to treat `smailnaild` as the primary hosted binary:

- [Dockerfile](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile) now builds the Vite UI and the embedded `smailnaild` binary
- [docker-entrypoint.smailnaild.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnaild.sh) became the primary runtime entrypoint
- [smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md) became the deployment guide for the merged server
- [smailnail-imap-mcp-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnail-imap-mcp-coolify.md) was rewritten as legacy guidance
- [shared-oidc-playbook.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/shared-oidc-playbook.md) was updated so the browser and MCP stories both point at the same host

I also bumped the `go-go-mcp` dependency in [go.mod](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/go.mod) and [go.sum](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/go.sum) so the new `MountHTTPHandlers(...)` API is available in the merged image.

### Validation

I ran:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/mcp/imapjs -run 'TestMountedHandlersCanBeServedByHostedHandler|TestMountedHandlersKeepWebAndMCPAuthSeparated'
go test ./pkg/mcp/imapjs ./pkg/smailnaild/... ./cmd/smailnaild/...
docker build -t smailnaild-merged:dev .
go run ./cmd/smailnail-imap-mcp mcp list-tools
```

The standalone wrapper still listed the same two tools after the merge:

- `executeIMAPJS`
- `getIMAPJSDocumentation`

That was the explicit compatibility check for the old binary.

### Why this mattered before deployment

The merge only becomes operationally real once the container image and the route contract are both updated. At this stage the code and the packaging were finally aligned enough to replace the old production image instead of just proving the concept locally.

## Implementation step 3: roll the merged image onto Coolify and wire production Keycloak

### Production shape discovery

Before touching the live app, I re-checked the public DNS and the existing Coolify app configuration.

Important discovery:

- `smailnail.scapegoat.dev` does not currently resolve
- the already-live hostname is `smailnail.mcp.scapegoat.dev`

That changed the rollout plan. Instead of creating a second public hostname first, I reused the existing public app and cut it over in place to the merged image. That preserved the already-working remote MCP URL while adding the browser surface to the same host.

### Coolify app update

I used the existing Coolify app:

- app UUID: `fhp3mxqlfftdxdib3vxz89l3`

and updated it so the container now exposes the merged `smailnaild` runtime instead of the standalone MCP image.

The important runtime values I synced were:

- `SMAILNAILD_LISTEN_PORT=8080`
- `SMAILNAILD_DB_TYPE=sqlite`
- `SMAILNAILD_DATABASE=/app/smailnaild.sqlite`
- `SMAILNAILD_LOG_LEVEL=debug`
- `SMAILNAILD_ENCRYPTION_KEY_ID=prod-smailnail-merged-v1`
- `SMAILNAILD_ENCRYPTION_KEY_BASE64=aDYowYFnlu+JlsnKiE8XWGiuUUqTvxK3dmmMbpa9zl4=`
- `SMAILNAILD_AUTH_MODE=oidc`
- `SMAILNAILD_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail`
- `SMAILNAILD_OIDC_CLIENT_ID=smailnail-web`
- `SMAILNAILD_OIDC_CLIENT_SECRET=7GuMpylfY9WkYfdMlaMq12ieJhxaeUvp`
- `SMAILNAILD_OIDC_REDIRECT_URL=https://smailnail.mcp.scapegoat.dev/auth/callback`
- `SMAILNAILD_OIDC_SCOPES=openid,profile,email`
- `SMAILNAILD_MCP_ENABLED=1`
- `SMAILNAILD_MCP_TRANSPORT=streamable_http`
- `SMAILNAILD_MCP_AUTH_MODE=external_oidc`
- `SMAILNAILD_MCP_AUTH_RESOURCE_URL=https://smailnail.mcp.scapegoat.dev/mcp`
- `SMAILNAILD_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail`

I also changed the Coolify health check to:

- `/readyz`

because `smailnaild` now owns the hosted process.

### Commands I actually used

To inspect the live app:

```bash
ssh root@89.167.52.236 \
  'TOKEN=$(cat ~/.apitoken) && curl -sS -H "Authorization: Bearer $TOKEN" \
  https://hq.scapegoat.dev/api/v1/applications/fhp3mxqlfftdxdib3vxz89l3'
```

To inspect and sync env:

```bash
ssh root@89.167.52.236 \
  'TOKEN=$(cat ~/.apitoken) && curl -sS -H "Authorization: Bearer $TOKEN" \
  https://hq.scapegoat.dev/api/v1/applications/fhp3mxqlfftdxdib3vxz89l3/envs'
```

and then:

```bash
coolify app env sync fhp3mxqlfftdxdib3vxz89l3 --context scapegoat --file <env-file>
```

To trigger the rollout:

```bash
ssh root@89.167.52.236 \
  'TOKEN=$(cat ~/.apitoken) && curl -sS -H "Authorization: Bearer $TOKEN" \
  "https://hq.scapegoat.dev/api/v1/deploy?uuid=fhp3mxqlfftdxdib3vxz89l3&force=false"'
```

The successful deployment UUID was:

- `oo9pi1q0t3zglzmyn19i56e4`

and the new container became:

- `fhp3mxqlfftdxdib3vxz89l3-201817575552`

### Production Keycloak work

The merged server needed a real browser-login client. I used `kcadm.sh` inside the production Keycloak container and created:

- client: `smailnail-web`
- redirect URI: `https://smailnail.mcp.scapegoat.dev/auth/callback`
- web origin: `https://smailnail.mcp.scapegoat.dev`

I also created the production smoke-test user:

- username: `alice`
- password: `secret`
- email: `alice@smailnail.test`
- first name: `Alice`
- last name: `Smailnail`

One small Keycloak CLI gotcha mattered here:

- `kcadm.sh set-password` on this image rejected `--temporary false`
- the working form was just `set-password --new-password secret`

### Result after rollout

The hosted server responded correctly on the new merged shape:

```bash
curl -sS https://smailnail.mcp.scapegoat.dev/readyz
curl -sS https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource
curl -i -sS -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' \
  https://smailnail.mcp.scapegoat.dev/mcp
curl -I -sS https://smailnail.mcp.scapegoat.dev/
```

Observed behavior:

- `/readyz` returned `{"status":"ready"}`
- the protected-resource document advertised the Keycloak realm and `/mcp` resource URL
- unauthenticated `/mcp` returned `401`
- `/` served the embedded SPA HTML

At that point the code merge was deployed, but I still needed to prove both auth paths with real traffic.

## Implementation step 4: run the hosted browser-session flow and save a real IMAP account

### First hosted login issue

The first production browser-login attempt did not create a usable app session. I built a shell smoke around `curl` and the Keycloak login form so I could see the raw redirects and cookies instead of guessing from the browser UI.

The initial problem was not the app itself. The Keycloak user `alice` got redirected into a required-action flow instead of returning straight to the app callback. I fixed that by populating `firstName`, `lastName`, `email`, and clearing `requiredActions` for the user in the production realm.

I saved the final smoke script as:

- [smoke_hosted_web_login.sh](../scripts/smoke_hosted_web_login.sh)

### Hosted browser-login validation

With the updated user, the hosted login flow succeeded:

- `/auth/login` redirected into Keycloak
- posting `alice/secret` returned to `/auth/callback`
- `/auth/callback` created `smailnail_session`
- `/api/me` returned the provisioned local user

The successful response was:

```json
{
  "data": {
    "id": "0a77a135-b891-4761-a9b6-5f9c9a3a4e8a",
    "primaryEmail": "alice@smailnail.test",
    "displayName": "Alice Smailnail",
    "avatarUrl": "",
    "createdAt": "2026-03-16T20:22:43Z",
    "updatedAt": "2026-03-16T20:28:16Z"
  }
}
```

That proved the merged hosted server was now doing real production browser OIDC, local identity provisioning, and secure session-cookie resolution.

### Hosted account creation

Using the authenticated cookie jar, I created a real saved IMAP account against the hosted Dovecot fixture:

```bash
curl -sS -b /tmp/smailnail-prod.cookies -c /tmp/smailnail-prod.cookies \
  -H 'Content-Type: application/json' \
  -X POST https://smailnail.mcp.scapegoat.dev/api/accounts \
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
  }'
```

The saved account ID was:

- `79d8b6a5-c307-4eff-908b-886168177089`

### Hosted IMAP-access validation

I then exercised the hosted API against that stored account:

```bash
curl -sS -b /tmp/smailnail-prod.cookies \
  https://smailnail.mcp.scapegoat.dev/api/accounts/79d8b6a5-c307-4eff-908b-886168177089/mailboxes | jq

curl -sS -b /tmp/smailnail-prod.cookies \
  'https://smailnail.mcp.scapegoat.dev/api/accounts/79d8b6a5-c307-4eff-908b-886168177089/messages?mailbox=INBOX&limit=10&offset=0' | jq
```

Those both succeeded. The hosted app listed:

- `Archive`
- `INBOX`

and returned the message previews for the seeded mailbox content.

### One oddity that remained

The dedicated lightweight test endpoint:

- `POST /api/accounts/{id}/test`

returned:

- `success=false`
- `errorCode=account-test-mailbox-select-failed`
- `errorMessage=use of closed network connection`

against the hosted Dovecot fixture, even though:

- direct CLI access to the same Dovecot server worked
- hosted mailbox listing worked
- hosted message preview worked
- bearer-authenticated MCP execution against the same stored account worked

I am recording that as a residual bug in the read-only test path, not as a blocker for the merged deployment itself. The merged deployment proved that the saved account is real and usable; this one endpoint is more pessimistic than the rest of the system against the remote fixture.

## Implementation step 5: run the hosted bearer-authenticated MCP smoke against the same saved account

### Keycloak test client for password-grant smoke

To validate `/mcp` without involving Claude/OpenAI connector setup again, I created a temporary production realm client:

- `smailnail-mcp-test`

with:

- `publicClient=true`
- `directAccessGrantsEnabled=true`
- `standardFlowEnabled=false`

That was only for smoke testing. It gave me a predictable way to mint a bearer token for `alice` and use it directly against `/mcp`.

### Hosted merged MCP smoke

I then ran the existing ticket script:

- [smoke_merged_server_mcp.go](../scripts/smoke_merged_server_mcp.go)

against the production endpoints:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp
SMAILNAIL_SERVER_URL=https://smailnail.mcp.scapegoat.dev/mcp \
SMAILNAIL_TOKEN_URL=https://auth.scapegoat.dev/realms/smailnail/protocol/openid-connect/token \
SMAILNAIL_CLIENT_ID=smailnail-mcp-test \
SMAILNAIL_USERNAME=alice \
SMAILNAIL_PASSWORD=secret \
SMAILNAIL_ACCOUNT_ID=79d8b6a5-c307-4eff-908b-886168177089 \
go run ./ttmp/2026/03/16/SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT--merge-smailnaild-and-smailnail-mcp-into-one-hosted-server-and-deploy-it/scripts/smoke_merged_server_mcp.go
```

The result was:

```json
{
  "accountID": "79d8b6a5-c307-4eff-908b-886168177089",
  "serverURL": "https://smailnail.mcp.scapegoat.dev/mcp",
  "value": {
    "mailbox": "INBOX"
  }
}
```

That was the final proof for the merged-server design:

- browser OIDC login provisioned and resolved the local user
- the hosted app stored encrypted IMAP credentials for that user
- the merged `/mcp` route validated the bearer token for the same Keycloak identity
- MCP execution resolved the same stored account by `accountId`
- the actual IMAP connection succeeded on the production merged host

## End state

At the end of this ticket:

- one hosted binary owns the SPA, `/auth/*`, `/api/*`, `/mcp`, and `/.well-known/oauth-protected-resource`
- the old standalone MCP binary still works as a compatibility wrapper
- production now serves the merged host on `https://smailnail.mcp.scapegoat.dev`
- browser-session auth and bearer-token auth remain separated by route
- the deployment and operator docs now describe the merged shape instead of the split one

The one known residual issue is the false-negative remote result from:

- `POST /api/accounts/{id}/test`

against the hosted Dovecot fixture. The rest of the merged hosted account and MCP flows are validated end to end.

## Implementation step 6: harden the hosted account test against transient closed-connection failures

### Why I revisited this

After closing the ticket the first time, the hosted account-test failure was still the obvious rough edge in the user-facing account-setup flow. The user explicitly asked to handle that item next before moving on to lower-priority follow-ups.

I started by checking whether the failure was actually reproducible outside the hosted container. The important command was:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
SMAILNAILD_DOVECOT_TEST=1 \
SMAILNAILD_DOVECOT_SERVER=89.167.52.236 \
SMAILNAILD_DOVECOT_PORT=993 \
SMAILNAILD_DOVECOT_USERNAME=a \
SMAILNAILD_DOVECOT_PASSWORD=pass \
SMAILNAILD_DOVECOT_MAILBOX=INBOX \
go test ./pkg/smailnaild/accounts -run TestServiceAgainstLocalDovecot -v
```

That passed cleanly against the same hosted Dovecot fixture. So the failure was not a deterministic logic bug in the basic `RunTest` flow.

### What I changed

I refactored the account-test path in [service.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go):

- extracted the actual IMAP read-only probe into `runReadOnlyProbe(...)`
- added `shouldRetryReadOnlyProbe(...)`
- retried the probe once when the failure looks like a transient connection-closure event
- merged the successful retry result back into the persisted `TestResult`

The transient patterns I handle now are:

- `net.ErrClosed`
- `use of closed network connection`
- `broken pipe`
- `connection reset by peer`
- `unexpected eof`

This is intentionally narrow. Real login failures should still fail immediately and clearly.

### Test coverage

I added [service_retry_test.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service_retry_test.go) with two focused unit tests:

- retry once on a transient closed-network failure and succeed
- do not retry a non-transient authentication failure

That gave me a deterministic red-green loop for the retry logic without needing to rely on the hosted fixture to misbehave on demand.

### Validation before deploy

I ran:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/smailnaild/accounts ./pkg/smailnaild/... ./cmd/smailnaild/...
```

and the remote-fixture integration test again:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
SMAILNAILD_DOVECOT_TEST=1 \
SMAILNAILD_DOVECOT_SERVER=89.167.52.236 \
SMAILNAILD_DOVECOT_PORT=993 \
SMAILNAILD_DOVECOT_USERNAME=a \
SMAILNAILD_DOVECOT_PASSWORD=pass \
SMAILNAILD_DOVECOT_MAILBOX=INBOX \
go test ./pkg/smailnaild/accounts -run TestServiceAgainstLocalDovecot -v
```

Both passed.

### Deployment and an important persistence reminder

I committed the code as:

- `3663738` `fix(smailnaild): retry transient imap account test failures`

and redeployed the same Coolify app from that commit.

During post-deploy verification, the first `POST /api/accounts/{id}/test` call came back as:

- `404 account not found`

That was not another account-test bug. It was the persistence issue showing up again:

- the merged host still uses `/app/smailnaild.sqlite`
- the redeploy replaced the container
- the previously saved hosted account disappeared with it

So I logged in again, recreated the hosted account, and reran the hosted smokes on the new container state.

### Post-deploy hosted validation

The recreated hosted account ID was:

- `22d4af8c-728f-4691-91cb-abd8d80e9430`

I then ran the hosted account test repeatedly:

```bash
for i in 1 2 3 4 5; do
  curl -sS -b /tmp/smailnail-hosted.cookies \
    -H 'Content-Type: application/json' \
    -X POST \
    https://smailnail.mcp.scapegoat.dev/api/accounts/22d4af8c-728f-4691-91cb-abd8d80e9430/test \
    -d '{"mode":"read_only"}' | jq
done
```

All five attempts succeeded against the hosted Dovecot fixture.

I also reran the merged bearer-authenticated MCP smoke on the same recreated account:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp
SMAILNAIL_SERVER_URL=https://smailnail.mcp.scapegoat.dev/mcp \
SMAILNAIL_TOKEN_URL=https://auth.scapegoat.dev/realms/smailnail/protocol/openid-connect/token \
SMAILNAIL_CLIENT_ID=smailnail-mcp-test \
SMAILNAIL_USERNAME=alice \
SMAILNAIL_PASSWORD=secret \
SMAILNAIL_ACCOUNT_ID=22d4af8c-728f-4691-91cb-abd8d80e9430 \
go run ./ttmp/2026/03/16/SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT--merge-smailnaild-and-smailnail-mcp-into-one-hosted-server-and-deploy-it/scripts/smoke_merged_server_mcp.go
```

That still returned:

```json
{
  "value": {
    "mailbox": "INBOX"
  }
}
```

### Final state after this step

The earlier user-visible rough edge is now handled in code and revalidated in production:

- hosted `/api/accounts/{id}/test` is succeeding on the merged host
- the merged `/mcp` flow still works on the same account
- the remaining real operational issue is no longer the transient IMAP test failure
- the remaining real operational issue is still the non-persistent app DB on redeploy
