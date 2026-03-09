---
Title: Investigation diary
Ticket: MCP-001-INITIAL-ASSESSMENT
Status: active
Topics:
    - mcp
    - go
    - assessment
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-mcp/pkg/embeddable/examples/basic/main.go
      Note: Example server used for command
    - Path: go-go-mcp/pkg/embeddable/examples/oidc/main.go
      Note: OIDC example used to validate metadata and issuer mismatch behavior
    - Path: go-go-mcp/ttmp/2026/03/08/MCP-001-INITIAL-ASSESSMENT--initial-assessment-of-go-go-mcp/scripts/oidc-smoke.sh
      Note: Reproducible OIDC metadata and protected-resource experiments
    - Path: go-go-mcp/ttmp/2026/03/08/MCP-001-INITIAL-ASSESSMENT--initial-assessment-of-go-go-mcp/scripts/standalone-smoke.sh
      Note: Reproducible standalone validation and transport smoke commands used during investigation
ExternalSources: []
Summary: Chronological diary of the MCP-001 assessment, including repository inventory, workspace-vs-standalone build experiments, transport and OIDC smoke tests, and documentation conclusions.
LastUpdated: 2026-03-08T17:51:20.332321857-04:00
WhatFor: Capture what was investigated, what worked, what failed, and how to reproduce the assessment findings.
WhenToUse: Use this when reviewing the assessment, reproducing its experiments, or resuming work on go-go-mcp cleanup.
---


# Investigation diary

## Goal

Capture the end-to-end investigation for `MCP-001-INITIAL-ASSESSMENT`: what was inspected, how the repo behaves in practice, what evidence shaped the conclusions, and how to repeat the validation.

## Step 1: Create the ticket and map the repository surface

This step established the ticket workspace and identified the main code areas that could affect a realistic assessment. The repository immediately showed a warning sign: it contains both executable runtime code and a large amount of older design material, so any conclusion would have to separate current behavior from historical intent.

The initial inventory also showed that `go-go-mcp` is broader than a simple server/client package. There is a direct CLI, an embeddable integration layer, config-driven tool loading, OIDC support, a TUI/editor surface, and a separate scholarly application. That meant the assessment needed to focus on runtime-critical paths first.

### Prompt Context

**User prompt (verbatim):** "GOt hrough go-go-mcp/ which is an old MCP implementation of mine, and create a new ticket MCP-001-INITIAL-ASSESSMENT where you take a look at how out of date thismight be, what is good, what is bad, what we should address now, if it works, etc...

You can run any experiments you want (let me know if you want to install software to validate the MCP implementation, for example, or do OIDC experiments) in the scripts/ folder of the ticket.

Keep a detailed diary as you investigate, and at the end

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new docmgr ticket, investigate the live `go-go-mcp` repository in depth, record the investigation as a diary, run validation experiments, and produce a detailed intern-oriented assessment document that is uploaded to reMarkable.

**Inferred user intent:** Establish a trustworthy picture of the current quality and viability of `go-go-mcp` before deciding on future cleanup or reinvestment.

### What I did

- Opened the `ticket-research-docmgr-remarkable` and `diary` skill instructions.
- Ran `docmgr status --summary-only`.
- Read `README.md`, `AGENT.md`, and initial file inventories with `rg --files`.
- Created the ticket:
  - `docmgr ticket create-ticket --ticket MCP-001-INITIAL-ASSESSMENT --title 'Initial assessment of go-go-mcp' --topics mcp,go,assessment,architecture`
- Added the primary docs:
  - `docmgr doc add --ticket MCP-001-INITIAL-ASSESSMENT --doc-type design-doc --title 'go-go-mcp initial assessment and modernization guide'`
  - `docmgr doc add --ticket MCP-001-INITIAL-ASSESSMENT --doc-type reference --title 'Investigation diary'`

### Why

- The ticket needed to exist before any diary or design material could be tracked properly.
- The repo contains many historical notes, so I needed an up-front map before trusting any narrative.

### What worked

- `docmgr` ticket/document creation worked cleanly.
- The repository inventory was fast and exposed the major subsystems immediately.

### What didn't work

- `git -C /home/manuel/workspaces/2026-03-08/update-imap-mcp status --short`
- Exact error:

```text
fatal: not a git repository (or any of the parent directories): .git
```

- This mattered only for the workspace root; it did not block the assessment because the work product lives in docmgr files under `go-go-mcp/ttmp`.

### What I learned

- `go-go-mcp` already contains a lot of self-analysis from 2025, including planned migration and testing documents.
- The current assessment needed to verify what is live today, not just repeat those older plans.

### What was tricky to build

- The hard part at this stage was context discipline. The repo contains many plausible narratives: old design docs, current code, stale README claims, and separate apps. The only reliable approach was to map entry points and runtime wiring before making any judgment.

### What warrants a second pair of eyes

- Whether the existing 2025 internal docs still reflect the owner's intended direction, or whether they should now be treated as historical background only.

### What should be done in the future

- Keep future tickets scoped around validated runtime paths, because this repo makes it easy to confuse plans with implementation.

### Code review instructions

- Start with the ticket files created under `go-go-mcp/ttmp/2026/03/08/MCP-001-INITIAL-ASSESSMENT--initial-assessment-of-go-go-mcp/`.
- Then read the main CLI entrypoint in `cmd/go-go-mcp/main.go` and the server/client helpers referenced later in this diary.

### Technical details

- Key setup commands:

```bash
docmgr status --summary-only
docmgr ticket create-ticket --ticket MCP-001-INITIAL-ASSESSMENT --title 'Initial assessment of go-go-mcp' --topics mcp,go,assessment,architecture
docmgr doc add --ticket MCP-001-INITIAL-ASSESSMENT --doc-type design-doc --title 'go-go-mcp initial assessment and modernization guide'
docmgr doc add --ticket MCP-001-INITIAL-ASSESSMENT --doc-type reference --title 'Investigation diary'
rg --files go-go-mcp
```

## Step 2: Inspect the live architecture and compare code paths

This step focused on line-anchored code reading so the assessment would describe the implementation that actually runs today. The main conclusion was that the active runtime is already centered on `mcp-go`, while large parts of the repo still carry older local client/protocol abstractions and documentation.

The most important architectural distinction surfaced here: the CLI server path and the CLI client path are both current enough to use `mcp-go`, but the repository still contains a parallel local client stack and protocol/result types that make the overall system look more bespoke than it really is.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Identify which files define the live runtime and which files are legacy or partially wired.

**Inferred user intent:** Separate “working system” from “historical residue” so recommendations are grounded.

### What I did

- Read with line numbers:
  - `cmd/go-go-mcp/main.go`
  - `cmd/go-go-mcp/cmds/server/start.go`
  - `pkg/embeddable/server.go`
  - `pkg/embeddable/command.go`
  - `pkg/embeddable/mcpgo_backend.go`
  - `cmd/go-go-mcp/cmds/client/helpers/client.go`
  - `cmd/go-go-mcp/cmds/client/tools.go`
  - `cmd/go-go-mcp/cmds/server/tools.go`
  - `pkg/tools/providers/config-provider/tool-provider.go`
  - `pkg/tools/providers/tool-registry/registry.go`
  - `pkg/auth/oidc/server.go`
  - `pkg/client/client.go`
  - `pkg/client/sse.go`
  - `pkg/client/stdio.go`
  - `pkg/resources/registry.go`
  - `pkg/prompts/registry.go`
  - `pkg/protocol/*`
- Searched for:
  - bridge references
  - prompt/resource registration usage
  - `2024-11-05`
  - `LATEST_PROTOCOL_VERSION`
  - OIDC metadata handling

### Why

- The ticket needed precise claims such as “the server is effectively tool-only” and “the CLI client is `mcp-go`-backed”, which required direct code evidence.

### What worked

- The runtime boundary became clear quickly:
  - server start -> config provider -> proxy registry -> `embeddable.NewBackend`
  - client helpers -> `mcp-go` clients
- The search pass clearly exposed dead or drifting documentation around the removed bridge command.

### What didn't work

- No hard failures in this step; the main issue was ambiguity rather than broken commands.

### What I learned

- The active runtime already migrated farther toward `mcp-go` than the README suggests.
- The server path currently wires tools only; resources and prompts exist as packages but are not first-class in the live server startup flow.
- `pkg/client` is now best understood as a legacy/internal stack, not the primary CLI client implementation.

### What was tricky to build

- The tricky part was avoiding overstatement. The repo has enough partial support for prompts/resources that a quick skim could wrongly call them implemented. The actual server wiring had to be followed end-to-end to show that only tools are truly wired in the current path.

### What warrants a second pair of eyes

- Whether the local `pkg/protocol` types should be retained as stable internal handler/result types or treated as more migration debt to remove later.

### What should be done in the future

- Add explicit architecture notes in the repo that say:
  - which client path is canonical,
  - which server backend is canonical,
  - which packages are legacy.

### Code review instructions

- Read `cmd/go-go-mcp/cmds/server/start.go` and `pkg/embeddable/mcpgo_backend.go` together.
- Then compare `cmd/go-go-mcp/cmds/client/helpers/client.go` with `pkg/client/client.go`.
- Verify the bridge mismatch by comparing `README.md` with `cmd/go-go-mcp/main.go`.

### Technical details

- Important commands:

```bash
rg -n "bridge|Bridge" go-go-mcp/cmd/go-go-mcp go-go-mcp/README.md go-go-mcp/pkg/doc -S
rg -n "RegisterPrompt|RegisterResource|resources.NewRegistry\\(\\)" go-go-mcp/pkg go-go-mcp/cmd/go-go-mcp -S
rg -n "LATEST_PROTOCOL_VERSION|2024-11-05" go-go-mcp /home/manuel/go/pkg/mod/github.com/mark3labs/mcp-go@*/mcp -S
```

## Step 3: Run build and test experiments to separate workspace breakage from module health

This step answered the most important practical question: does the repo still work? The answer is “yes, but not in the checked-out workspace form.” With the active `go.work`, the repo fails immediately because the local `glazed/` checkout is incompatible. With `GOWORK=off`, the standalone `go-go-mcp` module builds and tests cleanly.

That distinction materially changes the assessment. Without this step, it would be easy to conclude the project is just broken. In reality, the working core still exists; the broken part is the workspace coupling and the lack of a single reliable validation path.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Determine whether `go-go-mcp` currently builds and passes tests, and identify whether failures come from the repo itself or from environmental/workspace assumptions.

**Inferred user intent:** Avoid making product or cleanup decisions based on a false-negative build result.

### What I did

- Ran failing workspace-scope commands:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp
go test ./...
go build ./...
```

- Investigated module resolution:

```bash
go list -f '{{.Dir}}' -m github.com/go-go-golems/glazed
go env GOWORK
sed -n '1,220p' /home/manuel/workspaces/2026-03-08/update-imap-mcp/go.work
ls -la /home/manuel/workspaces/2026-03-08/update-imap-mcp/glazed/pkg/cmds
find /home/manuel/workspaces/2026-03-08/update-imap-mcp/glazed/pkg/cmds -maxdepth 2 -type d | sort
go list github.com/go-go-golems/glazed/pkg/cmds/layers github.com/go-go-golems/glazed/pkg/cmds/parameters github.com/go-go-golems/glazed/pkg/cmds/middlewares
```

- Re-ran validation in standalone mode:

```bash
env GOWORK=off go list -deps ./cmd/go-go-mcp
env GOWORK=off go list -deps ./cmd/apps/scholarly
env GOWORK=off go test ./...
env GOWORK=off go build ./cmd/go-go-mcp
```

### Why

- The difference between workspace failure and module failure directly affects the quality assessment and the recommended next steps.

### What worked

- `GOWORK=off` resolved the full dependency graph.
- `env GOWORK=off go test ./...` passed.
- `env GOWORK=off go build ./cmd/go-go-mcp` succeeded.

### What didn't work

- `go test ./...`
- Exact failure excerpt:

```text
cmd/apps/scholarly/cmd/arxiv.go:13:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/layers; to add it:
	go get github.com/go-go-golems/glazed/pkg/cmds/layers
cmd/apps/scholarly/cmd/arxiv.go:14:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/parameters; to add it:
	go get github.com/go-go-golems/glazed/pkg/cmds/parameters
pkg/tools/providers/config-provider/tool-provider.go:15:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/middlewares; to add it:
	go get github.com/go-go-golems/glazed/pkg/cmds/middlewares
```

- `go build ./...` failed for the same reason.

### What I learned

- The active `go.work` file is part of the problem:

```text
go 1.25.7

use (
	./geppetto
	./glazed
	./go-go-goja
	./go-go-mcp
	./pinocchio
	./smailnail
)
```

- The local `glazed/` checkout is missing the required `pkg/cmds/layers`, `parameters`, and `middlewares` directories, while the tagged module cache contains them.
- The repo is therefore not self-contained in its current checked-out workspace form.

### What was tricky to build

- The initial failure message looked like a normal stale dependency issue, but the real cause was subtler: Go was resolving `github.com/go-go-golems/glazed` to the sibling workspace module from `go.work`, not to the standalone published dependency graph.

### What warrants a second pair of eyes

- Whether this workspace coupling is intentional and temporary, or whether `go-go-mcp` is supposed to be independently buildable as the default developer experience.

### What should be done in the future

- Make one validation path canonical and automate it.
- If `go.work` is intended to stay, pin compatible sibling modules or treat that workspace as a release-engineering artifact rather than the default path.

### Code review instructions

- Compare:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go.work`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/glazed/pkg/cmds`
  - `/home/manuel/go/pkg/mod/github.com/go-go-golems/glazed@v0.6.12/pkg/cmds`
- Then rerun `go test ./...` and `env GOWORK=off go test ./...` to see the difference directly.

### Technical details

- Relevant command outputs:

```bash
go list -f '{{.Dir}}' -m github.com/go-go-golems/glazed
# /home/manuel/workspaces/2026-03-08/update-imap-mcp/glazed

go env GOWORK
# /home/manuel/workspaces/2026-03-08/update-imap-mcp/go.work
```

## Step 4: Run end-to-end transport and OIDC smoke tests

This step validated the behavior that matters most to a practical user: can the current CLI and embeddable examples actually serve tools and talk over real transports? The answer was yes. The basic example worked as an MCP server, the CLI client could discover tools over command, SSE, and streamable HTTP, and a real tool call returned the expected greeting.

The OIDC path also proved more functional than the repo’s general drift might suggest. Discovery endpoints and protected-resource metadata were present, unauthenticated MCP access returned `401` with the expected metadata hints, and authenticated SSE access succeeded. The main defect found here was in the example configuration: the issuer stayed pinned to `http://localhost:3001` even when the example server was started on another port.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Run practical MCP experiments, including auth experiments, and record what happens exactly.

**Inferred user intent:** Learn whether the current MCP implementation is viable in real workflows, not just compilable.

### What I did

- Validated embeddable examples:

```bash
env GOWORK=off go run ./pkg/embeddable/examples/basic mcp list-tools
env GOWORK=off go run ./pkg/embeddable/examples/struct mcp list-tools
env GOWORK=off go run ./pkg/embeddable/examples/enhanced mcp list-tools
```

- Started the basic example over SSE and tested the CLI client:

```bash
env GOWORK=off go run ./pkg/embeddable/examples/basic mcp start --transport sse --port 4010
env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport sse --server http://localhost:4010/mcp/sse
```

- Tested command transport:

```bash
env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport command \
  --command env --command GOWORK=off --command go --command run \
  --command ./pkg/embeddable/examples/basic --command mcp --command start

env GOWORK=off go run ./cmd/go-go-mcp client tools call greet --transport command \
  --command env --command GOWORK=off --command go --command run \
  --command ./pkg/embeddable/examples/basic --command mcp --command start \
  --json '{"name":"Intern"}'
```

- Started the basic example over streamable HTTP:

```bash
env GOWORK=off go run ./pkg/embeddable/examples/basic mcp start --transport streamable_http --port 4012
env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport streamable_http --server http://localhost:4012/mcp
```

- Started the OIDC example and exercised auth endpoints:

```bash
env GOWORK=off go run ./pkg/embeddable/examples/oidc mcp start --transport sse --port 4011
curl -i -s http://localhost:4011/.well-known/openid-configuration
curl -i -s http://localhost:4011/.well-known/oauth-protected-resource
curl -i -s http://localhost:4011/mcp/sse
curl -i -s --max-time 2 -H 'Authorization: Bearer TEST_AUTH_KEY_123' http://localhost:4011/mcp/sse
```

### Why

- Compilation is necessary but insufficient. MCP infrastructure needs transport-level validation.
- OIDC is one of the most advanced parts of the repo and needed direct evidence.

### What worked

- Basic example listed tools successfully.
- Struct example listed tools successfully.
- Enhanced example listed tools successfully.
- CLI client listed tools over `sse`.
- CLI client listed tools over `command`.
- CLI client called `greet` over `command` and returned `Hello, Intern!`.
- CLI client listed tools over `streamable_http`.
- OIDC metadata endpoints responded successfully.
- Unauthenticated `GET /mcp/sse` returned `401 Unauthorized`.
- Authenticated `GET /mcp/sse` returned an SSE `endpoint` event.

### What didn't work

- `curl -i -s --max-time 2 -H 'Authorization: Bearer TEST_AUTH_KEY_123' http://localhost:4011/mcp/sse`
- Exit code:

```text
28
```

- This was expected here because the connection is a live SSE stream and `curl --max-time 2` intentionally times out after printing the initial event stream lines. The relevant successful output was:

```text
HTTP/1.1 200 OK
Content-Type: text/event-stream

event: endpoint
data: /mcp/message?sessionId=a807f37f-4f97-47ab-8cc9-6ba9a2bcb885
```

- Functional defect discovered:
  - `curl -i -s http://localhost:4011/.well-known/openid-configuration`
  - Discovery metadata advertised issuer `http://localhost:3001` even though the server was started on `4011`.

### What I learned

- The active runtime is meaningfully healthier than the docs imply.
- `streamable_http` is not aspirational here; it works in the current path.
- The OIDC implementation is viable enough for real experimentation, but the example configuration has correctness bugs that undermine confidence.

### What was tricky to build

- The main sharp edge was transport process management: long-running `go run ... start` commands do not emit useful incremental output beyond startup and need explicit timeout or background-session handling for clean scripted validation.

### What warrants a second pair of eyes

- The OIDC example's issuer handling when `--port` changes.
- The exact contract expected by external MCP clients around SSE and auth headers, especially if the repo is upgraded to newer transport/auth spec revisions.

### What should be done in the future

- Turn these manual smoke steps into committed repo-level integration tests or scripts.
- Add one automated auth-path check so future auth changes do not regress protected-resource metadata or `WWW-Authenticate` behavior.

### Code review instructions

- Review:
  - `pkg/embeddable/examples/basic/main.go`
  - `pkg/embeddable/examples/oidc/main.go`
  - `pkg/embeddable/mcpgo_backend.go`
  - `cmd/go-go-mcp/cmds/client/helpers/client.go`
- Re-run the transport and curl commands from this step to validate behavior end to end.

### Technical details

- Successful tool-list output excerpt:

```text
- calculate: Perform basic calculations
- greet: Greet a person
```

- Successful tool-call output:

```text
Hello, Intern!
```

- Successful unauthorized OIDC response excerpt:

```text
HTTP/1.1 401 Unauthorized
Www-Authenticate: Bearer realm="mcp", resource="http://localhost:3001/mcp", authorization_uri="http://localhost:3001/.well-known/oauth-authorization-server", resource_metadata="http://localhost:3001/.well-known/oauth-protected-resource"
```

## Related

- See the design doc in `../design-doc/01-go-go-mcp-initial-assessment-and-modernization-guide.md`.
- Reproducible smoke scripts are stored in `../scripts/`.

<!-- Provide background context needed to use this reference -->

## Quick Reference

<!-- Provide copy/paste-ready content, API contracts, or quick-look tables -->

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
