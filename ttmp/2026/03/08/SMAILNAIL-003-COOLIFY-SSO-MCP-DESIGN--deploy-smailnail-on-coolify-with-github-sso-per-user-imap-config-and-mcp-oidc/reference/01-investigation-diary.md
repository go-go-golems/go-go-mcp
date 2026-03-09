---
Title: Investigation diary
Ticket: SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN
Status: active
Topics:
    - smailnail
    - glazed
    - go
    - email
    - review
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/main.go
      Note: Evidence that the repo still starts from CLI roots
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go
      Note: Evidence of the direct IMAP credential model
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: Local OIDC and protected-resource reference examined during research
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/scripts/current-capability-scan.sh
      Note: Reproducible investigation script created during the ticket
ExternalSources: []
Summary: Chronological record of the research that led to the recommended Coolify, GitHub SSO, per-user IMAP, and MCP OIDC architecture for smailnail.
LastUpdated: 2026-03-08T22:32:50.520363359-04:00
WhatFor: Capture what was inspected, what external constraints mattered, what experiments were run, and how the final design recommendation was chosen.
WhenToUse: Use when reviewing or continuing the hosted smailnail architecture work.
---


# Investigation diary

## Goal

This diary records how the Coolify/GitHub/OIDC deployment request was converted into a concrete architecture recommendation for `smailnail`.

## Context

The user asked for a new ticket that explains how to deploy `smailnail` on Coolify, let users sign in via GitHub, store user-specific IMAP settings after login, and also use OIDC for the remote MCP surface.

## Quick Reference

## Step 1: Create The Ticket And Confirm The Current Boundary

I created `SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN` first so the research had a stable home. Then I verified whether `smailnail` already had any hosted-app substrate. It does not. The repo is still organized as three CLIs with shared IMAP logic and test helpers.

### Prompt Context

**User prompt (verbatim):** "Then create a new ticket with docmgr: I want to deploy this to coolify with SSO through github, and allowing users to configure their imap connection after logging in. (the OIDC is also for the MCP itself). 

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a research ticket and produce a detailed architecture guide for a future hosted `smailnail` product with browser auth, stored IMAP settings, and authenticated MCP.

**Inferred user intent:** Get a real system design deliverable, not just deployment notes.

### What I did
- Ran `docmgr status --summary-only`.
- Created the ticket.
- Added the design doc and diary.
- Inspected:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/main.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/mailgen/main.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/imap-tests/main.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go`

### Why
- I needed to know whether the request was about deployment packaging or adding major application architecture.

### What worked
- The evidence made the boundary clear quickly: useful IMAP core, but no hosted-app layers.

### What didn't work
- N/A

### What I learned
- `smailnail` has reusable domain logic.
- It does not yet have HTTP, sessions, persistence, or MCP.

### What was tricky to build
- The subtle part was not over-assuming that “deploy this” meant the app layer already existed.

### What warrants a second pair of eyes
- Whether the future hosted binary belongs in the same repo or a sibling repo.

### What should be done in the future
- Start implementation with a hosted skeleton ticket before touching MCP.

### Code review instructions
- Verify the three CLI roots and the IMAP settings struct.

### Technical details
- Ticket path:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc`

## Step 2: Check Current External Docs For The Parts That Move Fast

I then checked current official docs for Coolify, GitHub OAuth, Keycloak, and MCP authorization. This mattered because the design touches standards and products that change over time, and the user explicitly wants OIDC for the MCP surface.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Ground the architecture in current platform/auth guidance, not memory.

**Inferred user intent:** Avoid stale advice on deployment and auth.

### What I did
- Researched current official docs for:
  - Coolify applications, Docker Compose, env vars, health checks, and Keycloak service
  - GitHub OAuth auth-code flow
  - Keycloak admin and identity brokering
  - MCP authorization guidance

### Why
- The product architecture depends on what these platforms currently support.

### What worked
- The docs supported a clean design:
  - Coolify can host the app and Keycloak.
  - GitHub works well as upstream social login.
  - Keycloak is a suitable OIDC issuer for both app and MCP.
  - MCP auth guidance aligns with a bearer-token protected HTTP resource server.

### What didn't work
- Some guessed doc URLs were too noisy at first; targeted official docs were more reliable.

### What I learned
- The strongest design is “GitHub upstream, Keycloak issuer,” not “GitHub directly as everything.”

### What was tricky to build
- The hard part was keeping upstream identity and product authorization clearly separate.

### What warrants a second pair of eyes
- The IdP choice if the team already has a preferred platform.

### What should be done in the future
- If the team has an established IdP, adapt the design before implementation starts.

### Code review instructions
- Read the external reference list in the design doc and compare it with the recommended auth flow.

### Technical details
- Most important external docs:
  - `https://coolify.io/docs/applications/`
  - `https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps`
  - `https://www.keycloak.org/docs/latest/server_admin/`
  - `https://modelcontextprotocol.io/docs/tutorials/security/authorization`

## Step 3: Add A Capability Scan Script And Fix A Shell Bug

I added one ticket-local scan script so the ticket contains a runnable summary of the current repo boundary and the local OIDC/MCP reuse points. The first version had a shell quoting mistake: I left a backtick-bearing pattern in the `rg` expression and the shell treated it as command substitution.

I fixed the script and reran it successfully.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Make the research reproducible and leave behind at least one concrete investigation artifact.

**Inferred user intent:** Future readers should be able to verify the current-state claims quickly.

### What I did
- Added:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/scripts/current-capability-scan.sh`
- First run failed with:

```text
current-capability-scan.sh: line 18: glazed:password: command not found
```

- Removed the unsafe pattern and reran successfully.

### Why
- The ticket should not rely entirely on prose.

### What worked
- The final script clearly showed:
  - three CLI roots only
  - direct IMAP password usage still present
  - reusable OIDC/MCP patterns in `go-go-mcp`

### What didn't work
- My initial shell pattern was not properly quoted.

### What I learned
- Ticket scripts need the same shell hygiene as product code if they are meant to be rerun.

### What was tricky to build
- Only the quoting bug; the repo scan itself was straightforward.

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Add a second script once implementation starts, covering local Keycloak + app smoke tests.

### Code review instructions
- Re-run the scan script and compare its output with the design doc’s current-state section.

### Technical details
- Final script path:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/scripts/current-capability-scan.sh`

## Step 4: Write The Final Guide

With the local evidence and external constraints in place, I wrote the design guide around one main recommendation:

- add `smailnaild`
- use Keycloak as the OIDC issuer
- use GitHub as Keycloak’s social login provider
- store encrypted IMAP settings per user in Postgres
- expose authenticated MCP over `streamable_http`

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Produce the final intern-ready architecture guide and connect it cleanly to the ticket.

**Inferred user intent:** Lower the ambiguity of the next implementation ticket as much as possible.

### What I did
- Filled the design doc with:
  - current-state analysis
  - external constraints
  - target architecture
  - auth model
  - data model
  - API and MCP design
  - phased implementation plan
  - testing strategy
  - risks and alternatives

### Why
- The user asked for a detailed implementation guide, not a minimal RFC.

### What worked
- The final recommendation fits both the current code and the current auth/deployment guidance.

### What didn't work
- N/A

### What I learned
- The real problem is not deployment mechanics. It is identity and secret ownership in a codebase that was originally designed as a CLI.

### What was tricky to build
- The guide had to be detailed enough for an intern while staying honest about what is proposed rather than already implemented.

### What warrants a second pair of eyes
- The exact session-store implementation choice.
- The final choice of IdP product if the team has a preferred standard.

### What should be done in the future
- The next implementation ticket should focus on hosted app skeleton + auth + IMAP settings before MCP.

### Code review instructions
- Read the design doc first, then rerun the scan script.

### Technical details
- Primary guide:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md`

## Step 5: Validate And Upload The Ticket Bundle

Once the docs were complete, I finished the ticket the same way I handle the other research tickets: run `docmgr doctor`, then do a dry-run reMarkable bundle upload, then the real upload, then verify the remote listing. That made the ticket self-consistent and confirmed that the long-form guide renders into a deliverable artifact cleanly.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the ticket cleanly and deliver it to reMarkable after validation.

**Inferred user intent:** Leave behind a polished research artifact, not just local markdown files.

### What I did
- Ran:

```bash
docmgr doctor --ticket SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN --stale-after 30
```

- Verified `remarquee` and the cloud account:

```bash
remarquee status
remarquee cloud account --non-interactive
```

- Dry-ran the bundle upload, then uploaded:

```bash
remarquee upload bundle --dry-run <index> <design-doc> <diary> <tasks> \
  --name 'SMAILNAIL-003 Hosted Deployment and OIDC Design' \
  --remote-dir '/ai/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN' \
  --toc-depth 2

remarquee upload bundle <index> <design-doc> <diary> <tasks> \
  --name 'SMAILNAIL-003 Hosted Deployment and OIDC Design' \
  --remote-dir '/ai/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN' \
  --toc-depth 2
```

- Verified the remote listing:

```bash
remarquee cloud ls /ai/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN --long --non-interactive
```

### Why
- The user explicitly asked for the ticket to be uploaded to reMarkable.
- The dry-run catches bundling mistakes before a real upload.

### What worked
- `docmgr doctor` passed cleanly.
- The upload succeeded.
- The remote listing showed:

```text
[f]    SMAILNAIL-003 Hosted Deployment and OIDC Design
```

### What didn't work
- N/A

### What I learned
- The ticket bundle order of `index`, `design-doc`, `diary`, and `tasks` works well for this style of long-form handoff.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- If this architecture is accepted, open the next implementation ticket directly from the Phase 1 plan in the design doc.

### Code review instructions
- Open the uploaded bundle on reMarkable and compare it with the local docs if formatting matters.

### Technical details
- Upload destination:
  - `/ai/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN`

## Usage Examples

- Review the current-state evidence before opening an implementation PR.
- Re-run the capability scan before starting implementation if the repo has changed significantly.
- Use the phase plan in the design doc to split implementation into follow-up tickets.

## Related

- [design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md](../design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md)
