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

## Related

- Implementation plan: `../design-doc/01-concrete-deployment-plan-for-smailnail-mcp-on-coolify-with-keycloak.md`
- Ticket index: `../index.md`
