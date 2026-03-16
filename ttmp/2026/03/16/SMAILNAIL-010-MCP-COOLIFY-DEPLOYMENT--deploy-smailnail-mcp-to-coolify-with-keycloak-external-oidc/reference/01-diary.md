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
RelatedFiles: []
ExternalSources: []
Summary: Chronological implementation diary for packaging and deploying smailnail-mcp on Coolify with Keycloak OIDC and a hosted Dovecot target.
LastUpdated: 2026-03-16T00:10:00-04:00
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

## Related

- Implementation plan: `../design-doc/01-concrete-deployment-plan-for-smailnail-mcp-on-coolify-with-keycloak.md`
- Ticket index: `../index.md`
