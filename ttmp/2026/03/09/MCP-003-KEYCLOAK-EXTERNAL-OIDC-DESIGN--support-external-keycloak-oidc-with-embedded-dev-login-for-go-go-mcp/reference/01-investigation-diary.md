---
Title: Investigation diary
Ticket: MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN
Status: active
Topics:
    - mcp
    - oidc
    - keycloak
    - authentication
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-mcp/cmd/go-go-mcp/cmds/oidc.go
      Note: |-
        Existing embedded OIDC admin CLI
        Primary evidence for embedded user/client/token administration
    - Path: go-go-mcp/pkg/auth/oidc/server.go
      Note: |-
        Primary source for the current embedded issuer and login flow
        Primary evidence for embedded issuer and login flow
    - Path: go-go-mcp/pkg/doc/topics/07-embedded-oidc.md
      Note: |-
        Current documentation claims about embedded OIDC
        Current embedded OIDC documentation reviewed during analysis
    - Path: go-go-mcp/pkg/embeddable/command.go
      Note: |-
        Current CLI auth surface
        Primary evidence for the current flat OIDC flag surface
    - Path: go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: |-
        Primary source for MCP auth middleware and protected-resource metadata
        Primary evidence for current in-process introspection coupling
    - Path: go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/current-capability-scan.sh
      Note: |-
        Ticket-local capability scan used during the investigation
        Commands actually run during the investigation
ExternalSources:
    - https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization
    - https://modelcontextprotocol.io/docs/tutorials/security/authorization
    - https://www.keycloak.org/guides#server
    - https://www.keycloak.org/securing-apps/oidc-layers
    - https://coolify.io/docs/services/keycloak
Summary: Chronological investigation diary for designing external Keycloak issuer support and embedded dev login coexistence in go-go-mcp.
LastUpdated: 2026-03-09T18:12:00-04:00
WhatFor: Record the evidence-gathering steps, command outputs, failures, and design conclusions behind the external-issuer refactor plan.
WhenToUse: Use when reviewing the auth design, continuing the implementation later, or validating why the recommended architecture separates external and embedded auth modes.
---


# Diary

## Goal

Capture the investigation behind a new auth architecture for `go-go-mcp`: Keycloak as an external production issuer, plus an embedded local password-login path that remains easy to use in development. The diary focuses on evidence, not assumptions, because the key question is whether the current repository already has enough embedded OIDC machinery to be refactored instead of replaced.

## Step 1: Create the research ticket and verify the current auth surface exists in code

I started by creating a new ticket rather than reusing the earlier `smailnail` product ticket. That matters because this is a library/server refactor problem inside `go-go-mcp`, not a hosted product deployment plan. The first task was to verify whether the repo already had real OIDC code or whether the previous design recommendation had leaned on an assumption.

The answer became clear very quickly: `go-go-mcp` already contains a substantial embedded OIDC implementation and already wires it into MCP HTTP transports. That finding changed the nature of the design problem. The question was no longer “how do we add OIDC?” It became “how do we decouple the existing embedded issuer from MCP resource protection so external Keycloak tokens can be trusted too?”

### Prompt Context

**User prompt (verbatim):** "ok, create a new ticket and then do a detailed analysis, keeping a diary. You can even run experiments in the scripts/ folder of that ticket if needed.\n\nCreate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file\n  references.\n  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new `docmgr` research ticket for the `go-go-mcp` auth refactor question, investigate the current embedded OIDC code and current standards guidance, write a detailed architecture guide and diary, and upload the ticket bundle to reMarkable.

**Inferred user intent:** Produce a design package that is immediately useful for an intern or future implementer, not a conversational summary.

### What I did

- Created ticket `MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN`.
- Added:
  - the primary design doc
  - the investigation diary
- Inspected:
  - [pkg/auth/oidc/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/auth/oidc/server.go)
  - [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go)
  - [pkg/embeddable/command.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/command.go)
  - [pkg/embeddable/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/server.go)
  - [cmd/go-go-mcp/cmds/oidc.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/cmd/go-go-mcp/cmds/oidc.go)
  - [pkg/doc/topics/07-embedded-oidc.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/doc/topics/07-embedded-oidc.md)

### Why

- I needed to establish whether the earlier Keycloak recommendation was compatible with the code that actually exists in `go-go-mcp`.
- I also needed to identify the exact coupling points before recommending any refactor shape.

### What worked

- The current embedded OIDC implementation is real and substantial, not a stub.
- The auth middleware already understands MCP protected-resource metadata and `WWW-Authenticate` advertising, which is a strong foundation for external issuer support.

### What didn't work

- N/A

### What I learned

- The repository already has the local password-login path the user wants to preserve.
- The main problem is coupling, not feature absence.

### What was tricky to build

- The subtle part in this step was resisting the temptation to jump straight to Keycloak design. The right first step was proving what `go-go-mcp` already does, because the quality of the final recommendation depends on whether we are extending a real auth subsystem or inventing one from scratch.

### What warrants a second pair of eyes

- Whether the current embedded issuer should remain in `pkg/auth/oidc` unchanged and be wrapped, or whether some of its current responsibilities should be split further during the refactor.

### What should be done in the future

- Run a reproducible capability scan and capture the current auth/admin surface in the ticket.

### Code review instructions

- Start with [pkg/auth/oidc/server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/auth/oidc/server.go).
- Then read [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go).

### Technical details

- The current embedded server includes login, authorization, token, dynamic client registration, SQLite user management, and token introspection.

## Step 2: Create and run a ticket-local capability scan

After establishing the broad architecture, I wanted one command in the ticket that could re-derive the same baseline quickly later. That led to a small ticket-local script under `scripts/`. It inventories the auth-related files, surfaces the key OIDC symbols, and runs the existing CLI help entrypoints that expose embedded auth functionality.

The first run failed for an uninteresting but still real reason: executing the script directly from the ticket path returned `permission denied` even after `chmod +x`. Running it via `bash` worked immediately. I preserved that in the diary because it affects how future reviewers should run ticket scripts in this environment.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Keep a detailed, reproducible investigation trail and use the ticket `scripts/` folder when useful.

**Inferred user intent:** Make the research repeatable and operational, not just descriptive.

### What I did

- Added:
  - [scripts/current-capability-scan.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/current-capability-scan.sh)
- Ran:
  - `chmod +x .../scripts/current-capability-scan.sh`
  - direct execution attempt
  - `bash .../scripts/current-capability-scan.sh`

### Why

- I wanted a cheap way to re-derive the current surface of the auth subsystem while writing the design and during future review.

### What worked

- The scan confirmed:
  - embedded OIDC admin commands already exist
  - the embeddable `mcp start` flags are currently embedded-issuer oriented
  - the middleware currently reaches directly into `oidcSrv.IntrospectAccessToken(...)`

### What didn't work

- Direct execution from the ticket path failed with:

```text
zsh:1: permission denied: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/current-capability-scan.sh
```

- Running the same script with `bash` worked:

```bash
bash /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/current-capability-scan.sh
```

### What I learned

- Ticket-local scripts should be invoked with `bash` in this environment for reliability.
- The repository already documents and exposes the embedded dev helpers publicly enough that removing them outright would be a regression.

### What was tricky to build

- The tricky part was deciding how much to automate. I kept the script small and diagnostic rather than trying to make it a full validation harness, because this ticket is about architecture analysis rather than implementation testing.

### What warrants a second pair of eyes

- None in the script itself; it is intentionally simple. The important part is the evidence it surfaces.

### What should be done in the future

- When the implementation ticket lands, add a real smoke harness for both `embedded_dev` and `external_oidc` modes.

### Code review instructions

- Read [scripts/current-capability-scan.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/current-capability-scan.sh).
- Re-run it with:
  - `bash /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/scripts/current-capability-scan.sh`

### Technical details

- The scan script invokes `go run ./cmd/go-go-mcp oidc --help` and `go run ./pkg/embeddable/examples/oidc mcp start --help` so the ticket records the actual exposed admin and server flag surfaces, not just source-level assumptions.

## Step 3: Compare the current code to the earlier hosted design and current standards guidance

Once I had the local evidence, I compared it to the earlier hosted `smailnail` design and to current official MCP and Keycloak guidance. The earlier `SMAILNAIL-003` ticket had already recommended Keycloak as the production issuer with GitHub as upstream social login. That recommendation still stands, but the important refinement from this ticket is that `go-go-mcp` should not try to become Keycloak-like. It should become a cleaner protected resource server that can trust an external issuer.

The MCP auth material reinforced that direction. The resource server should advertise protected-resource metadata and tell clients where the authorization server lives. That is already close to what `go-go-mcp` does today. Keycloak’s current documentation, including its OIDC layers and server guides, supports the idea that the production answer is “use Keycloak as issuer and validate tokens correctly,” not “grow `go-go-mcp` into a full identity provider.”

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Use the current code and the current standards/docs to design the right architecture, including alternatives and tradeoffs.

**Inferred user intent:** Get a recommendation that is technically current, not just a repo-local opinion.

### What I did

- Reviewed prior product-level auth design in:
  - [SMAILNAIL-003 architecture guide](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md)
- Checked current official guidance:
  - MCP authorization spec
  - MCP authorization tutorial
  - Keycloak server guides and OIDC layers
  - Coolify Keycloak service docs

### Why

- The design question here is not purely local. The correct answer depends on what an MCP protected resource server should do and what Keycloak already offers.

### What worked

- The external guidance aligned cleanly with the local refactor direction:
  - keep protected-resource metadata
  - trust an external issuer in production
  - keep embedded login only for local/dev workflows

### What didn't work

- N/A

### What I learned

- The existing `go-go-mcp` embedded OIDC code is a good dev-mode implementation, but not a good reason to avoid external issuer support.
- The right abstraction boundary is “auth provider,” not “more options on the current embedded issuer.”

### What was tricky to build

- The hard reasoning problem was architectural, not mechanical. It would be easy to bolt Keycloak-specific conditionals onto the current middleware, but that would leave the core coupling problem untouched. The better design is to separate issuer behavior from resource-server behavior and let both embedded and external modes implement the same validation contract.

### What warrants a second pair of eyes

- Whether the first implementation of `external_oidc` should use local JWT verification only, or support optional remote introspection from day one.

### What should be done in the future

- Write the guide in a way that makes the implementation sequence explicit for a new engineer.

### Code review instructions

- Compare:
  - [pkg/embeddable/mcpgo_backend.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/mcpgo_backend.go)
  - [pkg/doc/topics/07-embedded-oidc.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/doc/topics/07-embedded-oidc.md)
  - [SMAILNAIL-003 architecture guide](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/design-doc/01-coolify-deployment-github-sso-per-user-imap-settings-and-mcp-oidc-architecture-guide.md)

### Technical details

- The main production recommendation in this ticket is “external issuer with local JWT validation and provider abstraction,” not “replace Keycloak” and not “keep the current embedded issuer as the production story.”

## Step 4: Synthesize the implementation guide, finish ticket bookkeeping, and prepare delivery

With the evidence and external references in place, I wrote the main architecture guide with three goals: explain the current code clearly, explain why the current coupling blocks external issuer support, and explain exactly how to refactor without losing embedded dev auth. I kept the guide file-backed and explicit about alternatives because the easy failure mode here is producing a vague “support both” recommendation without showing where the current code must change.

The rest of the closeout work was ticket hygiene: update tasks, relate key files, update the changelog, run `docmgr doctor`, and then bundle the ticket for reMarkable. That work is not glamorous, but it is what makes the ticket usable later by someone who was not present during the investigation.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Finish the analysis package end to end and leave it in a documented, validated, distributable state.

**Inferred user intent:** Make the ticket the source of truth for the future implementation, not just notes from an exploratory session.

### What I did

- Wrote the primary guide:
  - [design-doc/01-go-go-mcp-external-keycloak-oidc-and-embedded-dev-login-architecture-guide.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/design-doc/01-go-go-mcp-external-keycloak-oidc-and-embedded-dev-login-architecture-guide.md)
- Updated:
  - [tasks.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/tasks.md)
  - `changelog.md`
  - file relationships via `docmgr doc relate`
- Ran:
  - `docmgr doctor --ticket MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN --stale-after 30`

### Why

- The guide is the actual deliverable.
- The bookkeeping and validation are what make the deliverable reviewable and shareable.

### What worked

- The repository evidence and the external guidance produced a consistent recommendation, not a split-brain answer.

### What didn't work

- N/A at the design level.

### What I learned

- This refactor can be staged cleanly. It does not require deleting the current embedded auth subsystem first.

### What was tricky to build

- The biggest challenge in the write-up was keeping the distinction between “issuer” and “protected resource” sharp. The current code blurs that boundary because the same process does both. The final guide needed to keep those roles conceptually separate even while explaining how the embedded mode still works.

### What warrants a second pair of eyes

- The exact first-version external JWT validation library choices and caching strategy once implementation starts.

### What should be done in the future

- Create a follow-up implementation ticket that executes the provider abstraction and external Keycloak mode in phases.

### Code review instructions

- Start with the architecture guide.
- Then review the ticket-local scan script and the current key code files it references.
- Finally run `docmgr doctor` and verify the reMarkable upload path listed in the final handoff.

### Technical details

- The design explicitly recommends a provider abstraction because it is the smallest architectural move that solves the production use case without deleting the current embedded development ergonomics.
