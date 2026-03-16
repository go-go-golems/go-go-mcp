---
Title: Diary
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
    - Path: ../../../../../../../smailnail/Dockerfile
      Note: Captured the healthcheck-specific runtime fix that made the hosted deployment healthy
    - Path: ../../../../../../../smailnail/deployments/coolify/smailnail-dovecot.compose.yaml
      Note: Defines the raw-port Coolify Dovecot fixture used for hosted IMAP testing
    - Path: ../../../../../../../smailnail/docs/deployments/smailnail-dovecot-coolify.md
      Note: Documents the hosted Dovecot fixture UUID
    - Path: ../../../../../../../smailnail/docs/deployments/smailnail-imap-mcp-coolify.md
      Note: Captured the exact Coolify create command and hosted route expectations
    - Path: ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_mcp_oidc.go
      Note: Implements the authenticated hosted MCP smoke client
    - Path: ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_mcp_oidc.sh
      Note: Fetches the smoke client secret and runs the authenticated hosted MCP smoke
ExternalSources: []
Summary: Chronological implementation diary for packaging and deploying smailnail-mcp on Coolify with Keycloak OIDC and a hosted Dovecot target.
LastUpdated: 2026-03-16T05:07:00-04:00
WhatFor: Preserve the exact commands, reasoning, deployment constraints, and validation steps used during this implementation.
WhenToUse: Use when reviewing the deployment work or continuing it later.
---




# Diary

## Goal

Track the end-to-end implementation of the first hosted `smailnail-mcp` deployment, from repository packaging through live infrastructure verification, plus a hosted Dovecot target for remote testing.

## Context

The repository already has a working local MCP binary and a local Dovecot+Keycloak stack, but it does not yet have production packaging or deployment artifacts. Separately, the Hetzner host already runs Coolify at `https://hq.scapegoat.dev` and Keycloak at `https://auth.scapegoat.dev`, and the intended public MCP hostname is `https://smailnail.mcp.scapegoat.dev`.

## Quick Reference

Target URLs:

- Coolify dashboard: `https://hq.scapegoat.dev`
- Keycloak base: `https://auth.scapegoat.dev`
- Planned MCP base: `https://smailnail.mcp.scapegoat.dev/mcp`

Current infrastructure facts already verified:

- Hetzner/Coolify host is reachable over SSH as `root@89.167.52.236`
- Keycloak discovery currently responds on the public domain
- Coolify dashboard is live, but its web UI currently requires an explicit login session

Current implementation scope:

- package and validate a production MCP image
- add deployment docs and env reference
- deploy MCP
- then add a separate hosted Dovecot test target

## Usage Examples

## Step 1: Ticket setup and deployment scoping

Created a dedicated implementation ticket so the packaging, deployment, and hosted-Dovecot work can be tracked independently from the earlier design-only ticket. I also re-checked the current repo and infrastructure state to make sure the plan is anchored in what actually exists today rather than the original design assumptions.

This step established that the MCP binary already supports external OIDC, the target hostname is `smailnail.mcp.scapegoat.dev`, Keycloak is already publicly reachable at `auth.scapegoat.dev`, and the Coolify box is reachable over SSH. It also exposed the first likely deployment constraint: the Coolify dashboard is live but not already authenticated in the browser context.

### Prompt Context

**User prompt (verbatim):** "ok, create a new ticket, create a concrete deployment plan document, add to ticket, and then add tasks, and then continue working on them, one by one, committing as you go, and keeping a diary"

**Assistant interpretation:** Create a new implementation ticket for the MCP deployment slice, write the plan/tasks/diary, then immediately start executing the deployment work in small committed steps.

**Inferred user intent:** Move from architecture talk to tracked, concrete delivery for the hosted MCP and its production testing environment.

### What I did
- Created ticket `SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT`
- Added a design doc and this diary
- Inspected the current `smailnail-imap-mcp` command and `go-go-mcp` external OIDC wiring
- Inspected the previous `SMAILNAIL-003` design ticket to identify what remained design-only
- Inspected the Coolify install diary under `/home/manuel/code/wesen/2026-03-15--install-coolify`
- Verified current public endpoints:
  - `https://hq.scapegoat.dev`
  - `https://auth.scapegoat.dev`
- Verified current server reachability via `ssh root@89.167.52.236`

### Why
- The earlier work established architecture but not packaging or deployment
- The user explicitly wants implementation with intermediate commits and a diary
- The hosted Dovecot requirement changed the scope enough to warrant a clean ticket

### What worked
- Ticket scaffolding succeeded with `docmgr`
- Public Keycloak discovery responded successfully
- SSH access to the Hetzner host worked
- The Coolify install diary provided high-value operational context without re-discovery

### What didn't work
- The Playwright browser session reached `https://hq.scapegoat.dev/login` but was not pre-authenticated, so the Coolify UI cannot yet be driven headlessly without a login session or a server-side workaround

### What I learned
- `smailnail` has no production Docker packaging yet
- `smailnail-imap-mcp` already exposes `external_oidc`-compatible flags through `go-go-mcp`
- The production MCP hostname and the Keycloak issuer are now concrete, not hypothetical
- The separate hosted Dovecot target needs to be treated as part of the deployment slice, not as a later nice-to-have

### What was tricky to build
- The tricky part here was not code yet; it was determining where the real blockers are. The repository gap is straightforward, but the deployment gap is split between repo artifacts and platform access. The Coolify host itself is reachable over SSH, while the Coolify UI still requires a login session. That means the ticket has to preserve both paths: repo implementation first, then either UI-driven app creation or a server-side fallback if Coolify internals can be used safely.

### What warrants a second pair of eyes
- Whether direct server-side manipulation of Coolify-managed application state is acceptable if UI access remains inconvenient
- Whether the hosted Dovecot service should be modeled as a Coolify-managed app, a service, or a host-level compose stack

### What should be done in the future
- Implement the repository packaging and deployment docs
- Validate the container locally
- Re-assess whether the actual live deployment can be completed through Coolify UI automation or needs a different path

### Code review instructions
- Start with the ticket plan and task list in this workspace
- Review the current MCP command in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail-imap-mcp/main.go`
- Review the external OIDC support in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/command.go`
- Review the Coolify operational baseline in `/home/manuel/code/wesen/2026-03-15--install-coolify/ttmp/2026/03/15/COOLIFY-001--configure-coolify-on-hetzner-server/reference/01-diary.md`

### Technical details
- Ticket path: `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc`
- Verified SSH command:
  - `ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@89.167.52.236 'hostname && docker ps --format "table {{.Names}}\t{{.Status}}"'`
- Verified Keycloak discovery command:
  - `curl -fsS https://auth.scapegoat.dev/realms/master/.well-known/openid-configuration | jq -r '.issuer, .authorization_endpoint, .token_endpoint'`

## Step 2: Repository packaging and local container validation

I implemented the first production-ready repository slice in `smailnail` and committed it as `ab5df7b`. This step turned the MCP binary into something Coolify can actually run: it now has deployment-oriented defaults, a production Dockerfile, an env-driven container entrypoint, and deployment docs keyed to `smailnail.mcp.scapegoat.dev`.

The validation mattered as much as the code. The first packaging attempt failed because the current module graph does not support `CGO_ENABLED=0` due to a transitive tree-sitter JavaScript dependency. I adjusted the image to use a glibc-based runtime instead of forcing a static build, rebuilt successfully, and then verified the live container behavior against the public Keycloak issuer by checking both protected-resource metadata and the unauthenticated `401` response on `/mcp`.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement the first real deployment slice in the repository rather than leaving the ticket as planning only.

**Inferred user intent:** Produce deployable artifacts that can move directly into hosted rollout work.

**Commit (code):** `ab5df7b` — `feat(smailnail): package imap mcp for coolify deployment`

### What I did
- Updated `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go` to default the MCP binary to `streamable_http` on port `3201`
- Added `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/.dockerignore`
- Added `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile.smailnail-imap-mcp`
- Added `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnail-imap-mcp.sh`
- Added `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnail-imap-mcp-coolify.md`
- Updated `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/README.md`
- Added a helper build target in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Makefile`
- Ran `gofmt -w pkg/mcp/imapjs/server.go`
- Ran `go test ./...`
- Built the image with `docker build -f Dockerfile.smailnail-imap-mcp -t smailnail-imap-mcp:dev .`
- Ran a live smoke container with:
  - `SMAILNAIL_MCP_AUTH_MODE=external_oidc`
  - `SMAILNAIL_MCP_AUTH_RESOURCE_URL=http://127.0.0.1:33201/mcp`
  - `SMAILNAIL_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/master`
- Verified:
  - `curl -s http://127.0.0.1:33201/.well-known/oauth-protected-resource`
  - `curl -i -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' http://127.0.0.1:33201/mcp`

### Why
- Coolify deployment needed a real container image, not just a binary and a design note
- The deployment path is materially easier when the container can be configured with environment variables instead of a long hand-maintained command
- Local validation needed to prove the containerized OIDC path behaves the way the documentation claims

### What worked
- `go test ./...` passed
- The revised Docker image built successfully
- The smoke container returned protected-resource metadata with the expected issuer
- Unauthenticated `/mcp` returned `401 Unauthorized` and a populated `WWW-Authenticate` header
- The pre-commit hook passed both `go test` and `golangci-lint`

### What didn't work
- The first Docker build used `CGO_ENABLED=0` and failed with:
  - `github.com/tree-sitter/tree-sitter-javascript/bindings/go: build constraints exclude all Go files in /go/pkg/mod/github.com/tree-sitter/tree-sitter-javascript@v0.25.0/bindings/go`
- Reproduced separately with:
  - `CGO_ENABLED=0 go build -o /tmp/smailnail-imap-mcp-static ./cmd/smailnail-imap-mcp`
- Plain `go build -o /tmp/smailnail-imap-mcp ./cmd/smailnail-imap-mcp` worked, confirming the failure was specifically the forced static build path

### What I learned
- The current `smailnail` module graph is not compatible with a fully static `CGO_ENABLED=0` image build
- A Debian/glibc runtime is the pragmatic packaging choice for this slice
- The MCP protected-resource metadata path is a good public health probe when auth is enabled

### What was tricky to build
- The sharp edge here was the mismatch between “typical Go container image advice” and this specific dependency graph. The first instinct of a static binary plus tiny Alpine image was wrong for this repo because the transitive tree-sitter dependency excludes the `CGO_ENABLED=0` path. The symptom appeared only during the container build, not in normal local `go test`, so I had to reproduce it directly and then adjust the runtime strategy instead of continuing to chase a nonexistent Dockerfile bug.

### What warrants a second pair of eyes
- Whether the environment-variable entrypoint surface is the right long-term contract or whether some of those options should eventually move into first-class app config
- Whether the current deployment docs should also include an explicit image-publishing step once a registry target is chosen

### What should be done in the future
- Configure the real `smailnail` Keycloak realm and client settings on the server
- Deploy the MCP service on the Hetzner/Coolify machine
- Add the hosted Dovecot target

### Code review instructions
- Start with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile.smailnail-imap-mcp`
- Then review `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnail-imap-mcp.sh`
- Then review the default transport change in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go`
- Validate with:
  - `cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail && go test ./...`
  - `cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail && docker build -f Dockerfile.smailnail-imap-mcp -t smailnail-imap-mcp:dev .`

### Technical details
- Successful image tag: `smailnail-imap-mcp:dev`
- Successful smoke metadata response:
  - `{"authorization_servers":["https://auth.scapegoat.dev/realms/master"],"resource":"http://127.0.0.1:33201/mcp"}`
- Successful unauthenticated response status:
  - `HTTP/1.1 401 Unauthorized`

## Step 3: Coolify app creation, Keycloak setup, and first hosted rollout attempt

I finished the public-repo deployment path and moved the work from “local packaging” to a real hosted application. This step created the actual Coolify app, created the production Keycloak realm and MCP client, and wired the hosted env vars against `https://smailnail.mcp.scapegoat.dev` and `https://auth.scapegoat.dev/realms/smailnail`.

This also exposed two concrete platform issues that were not visible from local testing. First, the current `coolify` CLI env helpers are out of sync with the API and send `is_build_time`, while the server expects `is_buildtime`. Second, the first hosted deployment failed even though the app itself was fine, because Coolify runs health checks from inside the container and the runtime image did not yet include `curl` or `wget`.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue from the initial packaging work into real hosted provisioning, keeping the ticket and diary current while executing live deployment steps.

**Inferred user intent:** Turn the MCP from a prepared artifact into a reachable HTTPS service with real Keycloak-backed OIDC and a repeatable Coolify setup.

**Commit (code):** `f24629d` — `feat(smailnail): add coolify-friendly dockerfile`

### What I did
- Added a standard root `Dockerfile` to `smailnail` so Coolify can build directly from the public repository
- Pushed branch `task/update-imap-mcp` to `origin`
- Installed and configured `coolify` CLI on `root@89.167.52.236`
- Fixed Coolify API allowlisting so server-side CLI calls from `89.167.52.236` are accepted
- Identified the target Coolify project/server:
  - project UUID `n8xkgqpbjj04m4pishy3su5e`
  - environment `production`
  - server UUID `cgl105090ljoxitdf7gmvbrm`
- Created the public-repo app:
  - app UUID `fhp3mxqlfftdxdib3vxz89l3`
  - domain `https://smailnail.mcp.scapegoat.dev`
- Created Keycloak realm `smailnail`
- Created Keycloak client `smailnail-mcp` with Claude callback URIs:
  - `https://claude.ai/api/mcp/auth_callback`
  - `https://claude.com/api/mcp/auth_callback`
- Discovered the `coolify app env sync` / `coolify app env create` bug and switched to direct API calls with the saved token in `~/.apitoken`
- Set the hosted MCP env vars through the API:
  - `SMAILNAIL_MCP_TRANSPORT=streamable_http`
  - `SMAILNAIL_MCP_PORT=3201`
  - `SMAILNAIL_MCP_AUTH_MODE=external_oidc`
  - `SMAILNAIL_MCP_AUTH_RESOURCE_URL=https://smailnail.mcp.scapegoat.dev/mcp`
  - `SMAILNAIL_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail`
- Triggered the first deployment with `coolify deploy uuid fhp3mxqlfftdxdib3vxz89l3 --force`

### Why
- The public repo path is the simplest deploy surface for this repository
- Using the real Keycloak issuer and public HTTPS domain is required for remote MCP/OAuth consumers like Claude Desktop
- The hosted app had to be created before any real external OIDC verification could happen

### What worked
- Coolify successfully cloned and built from the public `wesen/smailnail` repository
- Keycloak realm and client creation succeeded on the live server
- The app build itself completed on the host
- The exact app UUID, domain, and issuer are now fixed and testable

### What didn't work
- `coolify app create public ... --domains smailnail.mcp.scapegoat.dev` failed with:
  - `Validation failed.`
  - `Invalid URL: smailnail.mcp.scapegoat.dev`
- The working create command required a full URL:
  - `--domains https://smailnail.mcp.scapegoat.dev`
- `coolify app env create` failed with:
  - `API error 422 on applications/fhp3mxqlfftdxdib3vxz89l3/envs: Validation failed.`
- The debug output showed the underlying payload bug:
  - request body included `"is_build_time":true`
  - server error returned `{"errors":{"is_build_time":["This field is not allowed."]}}`
- The first hosted deployment failed its health checks because the runtime image lacked a probing binary:
  - `/bin/sh: 1: curl: not found`
  - `/bin/sh: 1: wget: not found`

### What I learned
- Coolify’s application create flow for public repos is usable from the CLI once the dashboard API allowlist is correct
- The current Coolify CLI env code path is incompatible with this server build and needs either a direct API workaround or an upstream fix
- Coolify’s Dockerfile/image health checks are not external-only; they shell into the container and need `curl` or `wget`

### What was tricky to build
- The hardest part here was separating “the app is broken” from “the platform glue is wrong.” The initial deployment looked unhealthy from the top-level app status, but the actual application build and start were fine. The failure was entirely in the health-check contract: Coolify expects an HTTP client binary inside the image. The env configuration path had a similar shape, where the CLI looked correct at first glance but the server-side controller only accepts `is_buildtime`, not `is_build_time`.

### What warrants a second pair of eyes
- Whether the Coolify CLI env mismatch is a known upstream regression or something version-specific to this host
- Whether the Keycloak client should immediately enforce audience/scope claims instead of starting with issuer-only validation

### What should be done in the future
- Patch the image to satisfy Coolify health checks
- Re-run the deployment and verify the public endpoints
- Set up the hosted Dovecot target next

### Code review instructions
- Start with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile`
- Then review `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnail-imap-mcp-coolify.md`
- Then review the live deployment state with:
  - `ssh root@89.167.52.236 'export PATH=$PATH:/usr/local/bin:$HOME/go/bin; coolify app get fhp3mxqlfftdxdib3vxz89l3 --context scapegoat --format pretty'`
  - `ssh root@89.167.52.236 'docker exec keycloak-k12lm4blpo13louovn3pfsgs /opt/keycloak/bin/kcadm.sh get realms/smailnail'`

### Technical details
- Coolify app UUID: `fhp3mxqlfftdxdib3vxz89l3`
- First failed deployment UUID: `mwz6x6hr6a7zyuqginolpup8`
- Keycloak client ID: `smailnail-mcp`
- Keycloak issuer: `https://auth.scapegoat.dev/realms/smailnail`
- Coolify env API workaround:
  - `POST https://hq.scapegoat.dev/api/v1/applications/fhp3mxqlfftdxdib3vxz89l3/envs`
  - payload keys must use `is_buildtime`, not `is_build_time`

## Step 4: Health-check image fix and successful hosted MCP verification

I fixed the runtime image for Coolify’s health-check model, redeployed the application, and verified the public HTTPS behavior end to end. This is the point where `smailnail-mcp` became a real hosted service instead of an app record with a failed rollout.

I also cleaned up the duplicated environment rows created while working around the broken CLI env helper, so the hosted configuration is now minimal and deterministic. The deployed application is healthy in Coolify, the protected-resource metadata resolves publicly, and unauthenticated access to `/mcp` returns the expected bearer challenge.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue through the failure rather than stopping at diagnosis, and record the exact fix plus verification steps in the ticket diary.

**Inferred user intent:** Finish the hosted MCP slice to the point where it is actually reachable and behaves correctly for OAuth-protected MCP clients.

**Commit (code):** `6072f7c` — `fix(smailnail): add curl for coolify healthchecks`

### What I did
- Updated both deployment Dockerfiles to install `curl` in the Debian runtime image
- Updated the deployment doc to note Coolify’s in-container health-check requirement
- Rebuilt the image locally with:
  - `docker build -t smailnail-imap-mcp:dev /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail`
- Verified `curl` exists in the runtime image with:
  - `docker run --rm --entrypoint sh smailnail-imap-mcp:dev -lc 'command -v curl'`
- Committed and pushed the fix
- Removed the duplicated env rows through direct API deletes, preserving one row per key
- Triggered a new deployment:
  - deployment UUID `vhgfh6tubvan4g796fn3ue7u`
- Verified Coolify app status:
  - `running:healthy`
- Verified the public metadata endpoint:
  - `curl -fsS https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource | jq -c .`
- Verified the unauthenticated MCP behavior:
  - `curl -i -s -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' https://smailnail.mcp.scapegoat.dev/mcp`

### Why
- The hosted deployment could not be considered complete while Coolify rolled it back as unhealthy
- Duplicated env rows were avoidable operator debt after the manual API workaround and were worth cleaning before moving to the next slice
- Public endpoint verification is the real proof that this deployment is ready for an OAuth client to test against

### What worked
- The new deployment finished successfully in Coolify
- Coolify now reports the app as `running:healthy`
- The live metadata endpoint returns:
  - `{"authorization_servers":["https://auth.scapegoat.dev/realms/smailnail"],"resource":"https://smailnail.mcp.scapegoat.dev/mcp"}`
- The live unauthenticated MCP call returns:
  - `HTTP/2 401`
  - `www-authenticate: Bearer realm="mcp", resource="https://smailnail.mcp.scapegoat.dev/mcp", authorization_uri="https://auth.scapegoat.dev/realms/smailnail/.well-known/openid-configuration", resource_metadata="https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource"`
  - body: `missing bearer`

### What didn't work
- The first `git push` after the fix did not complete cleanly on the first attempt even though the hooks ran; it succeeded on retry
- `golangci-lint` in the pre-push hook still emits:
  - `typechecking error: pattern ./...: open dist: no such file or directory`
  - but still reports `0 issues` and the later retry completed successfully

### What I learned
- The Coolify-hosted path is now real and testable from outside the server
- The current MCP deploy path is stable enough for remote client/OIDC work before the hosted Dovecot target exists
- The remaining infrastructure slice is the companion IMAP target, not the MCP container itself

### What was tricky to build
- The sharp edge here was that the failing health check was outside the app’s own logs. The service binary could start correctly and still get rolled back if the container image did not satisfy Coolify’s probe mechanism. Once that was fixed, the rest of the deployment path behaved as expected, which confirms the underlying MCP/OIDC configuration was already sound.

### What warrants a second pair of eyes
- Whether the pre-push lint behavior around `dist/` should be cleaned up separately so future deployment commits do not look flaky
- Whether the Keycloak client needs additional claim mapping before Claude/Desktop testing

### What should be done in the future
- Set up the hosted Dovecot target on the same Coolify machine
- Document the end-to-end remote test procedure against the hosted MCP plus hosted IMAP target

### Code review instructions
- Start with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile`
- Then review `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile.smailnail-imap-mcp`
- Then verify the hosted service with:
  - `curl -fsS https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource | jq`
  - `curl -i -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' https://smailnail.mcp.scapegoat.dev/mcp`
  - `ssh root@89.167.52.236 'export PATH=$PATH:/usr/local/bin:$HOME/go/bin; coolify app get fhp3mxqlfftdxdib3vxz89l3 --context scapegoat --format pretty'`

### Technical details
- Healthy app container:
  - `fhp3mxqlfftdxdib3vxz89l3-044752750295`
- Healthy image tag:
  - `fhp3mxqlfftdxdib3vxz89l3:6072f7c1ab15840dbeb6cdde6a8458b5c49dc2ab`
- Remaining env rows after cleanup:
  - `SMAILNAIL_MCP_TRANSPORT`
  - `SMAILNAIL_MCP_PORT`
  - `SMAILNAIL_MCP_AUTH_MODE`
  - `SMAILNAIL_MCP_AUTH_RESOURCE_URL`
  - `SMAILNAIL_MCP_OIDC_ISSUER_URL`

## Step 5: Hosted Dovecot fixture deployment and remote IMAPS validation

I created the hosted Dovecot fixture on the same Coolify machine so the deployment now has a real remote IMAP target, not just the local docker-compose fixture. The important part was to keep this aligned with the existing local setup, so I used the same `ghcr.io/spezifisch/docker-test-dovecot:latest` image and the same raw mail ports rather than inventing a new mail server path.

The result is a direct-port Coolify service that binds `24`, `110`, `143`, `993`, `995`, and `4190` on the Hetzner host. I then validated it with the actual `smailnail` tooling by creating a mailbox remotely, appending a known message over IMAPS, and fetching that message back from `89.167.52.236:993` with `--insecure`.

### Prompt Context

**User prompt (verbatim):** "once done, also set up a separate dovecot coolify server so that we have something we can test against, just like our local setup"

**Assistant interpretation:** Provision a remote Dovecot fixture on the Coolify host that behaves like the local docker-test-dovecot environment and can be used for real hosted IMAP validation.

**Inferred user intent:** Avoid blocking hosted MCP testing on local-only infrastructure by having a real remote IMAP server with predictable test accounts and mailbox behavior.

**Commit (code):** `04f2762` — `feat(smailnail): add hosted dovecot fixture definition`

### What I did
- Added `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/coolify/smailnail-dovecot.compose.yaml`
- Added `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnail-dovecot-coolify.md`
- Validated the compose file locally with:
  - `docker compose -f deployments/coolify/smailnail-dovecot.compose.yaml config`
- Created a custom Coolify service via the API using base64-encoded compose content
- Created service:
  - UUID `gh32795yh1av2dpi2j6lhn6h`
  - name `smailnail-dovecot-fixture`
- Verified the container is running:
  - `dovecot-gh32795yh1av2dpi2j6lhn6h`
- Verified host port bindings exist for:
  - `24`
  - `110`
  - `143`
  - `993`
  - `995`
  - `4190`
- Created remote mailbox `Archive` for user `a`
- Stored remote message with subject `Hosted Coolify Dovecot Test`
- Fetched the stored message back with `smailnail fetch-mail`

### Why
- The user explicitly wanted a hosted Dovecot target similar to the local setup
- The MCP deployment is much more useful once there is a stable remote IMAP endpoint to point it at
- Reusing the exact fixture image keeps local and hosted testing semantics aligned

### What worked
- Coolify accepted a custom compose-based service create request
- The service started and bound all expected raw mail ports on the host
- `imap-tests create-mailbox` succeeded remotely against `89.167.52.236:993`
- `imap-tests store-text-message` succeeded remotely against `89.167.52.236:993`
- `smailnail fetch-mail` returned the seeded message:
  - subject `Hosted Coolify Dovecot Test`
  - content `Remote hosted IMAP fixture validation.`

### What didn't work
- The first service create attempt failed because the server rejected `connect_to_docker_network`:
  - `{"message":"Validation failed.","errors":{"connect_to_docker_network":["This field is not allowed."]}}`
- The container’s first internal mail generation attempt raced Dovecot startup and failed once:
  - `lda(a)<117><>: Error: auth-master: userdb lookup(a): connect(/run/dovecot/auth-userdb) failed: No such file or directory`
  - `mailgen failed with code 75`
- That startup race did not block real IMAP access; manual mailbox/message operations worked immediately afterward

### What I learned
- The service API path is more reliable than the current CLI for custom compose-based resources
- For this test fixture, direct host port binding is the right model; HTTP-style proxying is irrelevant
- The hosted Dovecot service does not need to be “healthy” in Coolify’s HTTP sense to be fully usable for IMAP testing

### What was tricky to build
- The tricky bit was choosing the right Coolify abstraction. The mail fixture is not an HTTP app and does not benefit from Traefik domains, so the right move was a compose-based service with raw host port bindings. The other sharp edge was the mismatch between what the controller source suggested and what the live API accepted: `connect_to_docker_network` looked valid in code, but this running API rejected it and required a smaller payload.

### What warrants a second pair of eyes
- Whether we want a DNS name for the hosted IMAP fixture instead of using the host IP plus `--insecure`
- Whether the startup mailgen race in the upstream fixture image is worth patching, or whether manual seeding is enough for our purposes

### What should be done in the future
- Add one end-to-end hosted MCP invocation that points at this hosted Dovecot fixture
- Optionally assign a dedicated IMAP hostname if that improves operator ergonomics

### Code review instructions
- Start with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/coolify/smailnail-dovecot.compose.yaml`
- Then review `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnail-dovecot-coolify.md`
- Validate the hosted fixture with:
  - `ssh root@89.167.52.236 'ss -ltnp | grep -E ":(24|110|143|993|995|4190)\\b"'`
  - `cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail && go run ./cmd/imap-tests create-mailbox --server 89.167.52.236 --port 993 --username a --password pass --mailbox INBOX --new-mailbox Archive --insecure --output json`
  - `cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail && go run ./cmd/imap-tests store-text-message --server 89.167.52.236 --port 993 --username a --password pass --mailbox INBOX --from 'Remote Seeder <seed@example.com>' --to 'User A <a@testcot>' --subject 'Hosted Coolify Dovecot Test' --body 'Remote hosted IMAP fixture validation.' --insecure --output json`
  - `cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail && go run ./cmd/smailnail fetch-mail --server 89.167.52.236 --port 993 --username a --password pass --mailbox INBOX --subject-contains 'Hosted Coolify Dovecot Test' --insecure --output json`

### Technical details
- Service UUID: `gh32795yh1av2dpi2j6lhn6h`
- Current service status: `running:unknown`
- Current remote test host: `89.167.52.236`
- Remote seeded mailbox: `Archive`
- Remote seeded subject: `Hosted Coolify Dovecot Test`

## Step 6: Retroactive ticket script capture

I consolidated the ad hoc deployment helpers into the ticket `scripts/` directory so the work can be replayed without relying on shell history or piecing commands back together from the diary. This step is small in code terms, but it materially improves traceability because the actual operator commands now live next to the ticket instead of only in prose.

It also closed a documentation gap created by moving quickly through the live rollout. The hosted MCP, Keycloak, and Dovecot steps already existed as successful commands, but they had not all been preserved as reusable files. This step fixed that retroactively and linked the scripts back into the recreate flow.

### Prompt Context

**User prompt (verbatim):** "store all scripts in the ticket's scripts/ folder, retroactively too"

**Assistant interpretation:** Move the deployment helper commands into checked-in files under the ticket and update the ticket docs so those scripts are discoverable and attributable.

**Inferred user intent:** Make the ticket reproducible from artifacts, not just from memory or shell history.

### What I did
- Added the following files under the ticket workspace:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/01-scripts-index.md`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_coolify_mcp_app.sh`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/set_coolify_mcp_envs.sh`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/dedupe_coolify_mcp_envs.sh`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_smoke_client.sh`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_coolify_dovecot_service.sh`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_dovecot.sh`
- Updated the recreate doc to point directly at the ticket `scripts/` directory
- Updated the changelog and task list to capture the retroactive script archival step

### Why
- The user explicitly asked that all scripts live in the ticket workspace
- The deployment already relied on several one-off helper commands that were too useful to leave undocumented
- Keeping the scripts in the ticket improves reviewability and future handoff

### What worked
- The existing ad hoc commands were straightforward to preserve as checked-in scripts
- The ticket now has a single obvious home for executable helpers
- The recreate doc can point to concrete files instead of paraphrasing every shell sequence

### What didn't work
- The scripts were not captured early enough during the live rollout, so this had to be done retroactively after the fact

### What I learned
- The helper surface for this deployment is large enough that script capture should happen during the first implementation step, not after the system is already working
- The `scripts/` directory is the right place to keep ticket-scoped operational glue, especially when the commands include Coolify and Keycloak interactions that are easy to mistype

### What was tricky to build
- The subtle part here was deciding what counts as a durable script versus a transient shell experiment. For this ticket, the right threshold is broad: if a command materially created, configured, or validated live infrastructure, it belongs in `scripts/`. That includes the env-var API workaround and the hosted Dovecot smoke, not just the obvious app-creation commands.

### What warrants a second pair of eyes
- Whether any additional throwaway commands from the hosted rollout still exist only in the diary and should also become scripts
- Whether the existing shell helpers should be parameterized further before re-use on another environment

### What should be done in the future
- Add the authenticated hosted MCP smoke as another first-class script in this same directory

### Code review instructions
- Start with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/01-scripts-index.md`
- Then review the individual scripts in the same directory to confirm they cover Coolify app creation, env setup, Keycloak setup, and Dovecot validation
- Confirm the recreate doc now points to the script directory:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/reference/02-recreate-and-verify-hosted-smailnail-mcp.md`

### Technical details
- The ticket script inventory intentionally reflects the exact live rollout sequence that was used:
  - create app
  - set env
  - dedupe env
  - create Keycloak realm/client
  - create smoke client
  - create hosted Dovecot service
  - smoke hosted Dovecot

## Step 7: Authenticated hosted MCP end-to-end smoke

I finished the last missing validation step by running a real bearer-authenticated MCP session against the hosted server and using it to call `executeIMAPJS` against the hosted Dovecot fixture. This is the first proof that the complete remote path works as deployed: Keycloak issues a token, the hosted MCP accepts it over streamable HTTP, and the tool runtime can connect to the remote IMAP server and return a result.

I also turned that validation into reusable ticket scripts instead of leaving it as a one-off terminal experiment. The resulting smoke pair is a small Go streamable-HTTP MCP client plus a shell wrapper that can fetch the Keycloak smoke-client secret automatically when only the admin credentials are available.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue from the hosted Dovecot rollout into the final end-to-end validation step and keep documenting as the work lands.

**Inferred user intent:** Prove that the hosted deployment is usable in practice, not just reachable and theoretically configured.

### What I did
- Added `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_mcp_oidc.go`
- Added `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_mcp_oidc.sh`
- Updated the ticket script README and recreate doc to include the authenticated smoke path
- Ran:
  - `cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp && KEYCLOAK_ADMIN_USER='...' KEYCLOAK_ADMIN_PASSWORD='...' ./ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_mcp_oidc.sh`
- Verified the live result:
  - `{"serverURL":"https://smailnail.mcp.scapegoat.dev/mcp","tokenEndpoint":"https://auth.scapegoat.dev/realms/smailnail/protocol/openid-connect/token","toolCount":2,"value":{"mailbox":"INBOX"}}`

### Why
- The hosted MCP deployment was not fully proven until an authenticated client actually initialized a session and called a tool
- Raw `curl` against `/mcp` was enough to verify `401` behavior, but not enough to validate the streamable-HTTP protocol flow
- The user asked to continue and keep the ticket reproducible, so the smoke needed to become a checked-in script

### What worked
- The wrapper successfully logged into Keycloak admin, resolved the `smailnail-mcp-smoke` client UUID, fetched the client secret, and exported it for the Go smoke client
- The Go smoke client initialized the streamable-HTTP MCP session with a bearer token
- `tools/list` returned the advertised tool set
- `executeIMAPJS` connected to the hosted Dovecot fixture and returned `{"mailbox":"INBOX"}`

### What didn't work
- The first version of the Go smoke client only reported `tool returned error result`, which hid the actual tool payload and made debugging slower
- After improving the error output, the rerun succeeded, so there was no underlying server defect to fix

### What I learned
- The hosted stack is now validated at the correct abstraction level: not just HTTP reachability, but a real authenticated MCP tool call
- A tiny dedicated client is the right tool for streamable-HTTP validation; hand-rolled JSON-RPC POSTs are too lossy for this workflow
- The `smailnail-mcp-smoke` service-account client is a good operational testing hook for future hosted regressions

### What was tricky to build
- The subtle issue here was separating auth transport problems from tool/runtime problems. A raw bearer `curl` POST to `/mcp` had already shown that authentication worked, but it still failed with `400` because streamable HTTP expects proper MCP initialization and session behavior. The fix was not on the server side; it was to use a real MCP client transport and only then judge the hosted stack.

### What warrants a second pair of eyes
- Whether the `smailnail-mcp-smoke` client should remain a long-lived operational client or be rotated/recreated per deployment
- Whether the smoke should eventually be promoted into CI against a staging environment rather than remaining a ticket-local script

### What should be done in the future
- Optionally add an authenticated Claude/Desktop-specific validation once a real remote connector test is needed
- Consider creating a dedicated hostname and valid certificate for the hosted IMAP fixture if operator ergonomics around `--insecure` become a problem

### Code review instructions
- Start with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_mcp_oidc.go`
- Then review `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_mcp_oidc.sh`
- Then confirm the recreate doc includes the authenticated smoke section:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/reference/02-recreate-and-verify-hosted-smailnail-mcp.md`

### Technical details
- Hosted MCP endpoint: `https://smailnail.mcp.scapegoat.dev/mcp`
- Keycloak token endpoint: `https://auth.scapegoat.dev/realms/smailnail/protocol/openid-connect/token`
- Hosted IMAPS target used by the tool: `89.167.52.236:993`
- Successful tool result payload:
  - `{"mailbox":"INBOX"}`

## Related

- Implementation plan: `../design-doc/01-concrete-deployment-plan-for-smailnail-mcp-on-coolify-with-keycloak.md`
- Ticket index: `../index.md`
