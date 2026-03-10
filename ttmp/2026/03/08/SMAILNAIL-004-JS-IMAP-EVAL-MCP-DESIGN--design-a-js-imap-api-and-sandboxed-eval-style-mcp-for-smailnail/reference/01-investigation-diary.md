---
Title: Investigation diary
Ticket: SMAILNAIL-004-JS-IMAP-EVAL-MCP-DESIGN
Status: active
Topics:
    - smailnail
    - go
    - email
    - review
    - mcp
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-goja/engine/factory.go
      Note: Documented for runtime factory evidence
    - Path: go-go-goja/modules/common.go
      Note: Documented for native module interface evidence
    - Path: jesus/pkg/engine/engine.go
      Note: Documented as the runtime host reference
    - Path: jesus/pkg/mcp/server.go
      Note: Documented as the eval-style MCP reference
    - Path: smailnail/pkg/dsl/processor.go
      Note: Documented as the core fetch execution path
    - Path: smailnail/pkg/dsl/types.go
    - Path: smailnail/pkg/imap/layer.go
      Note: Documented as the initial IMAP connection evidence
ExternalSources: []
Summary: Chronological notes for the analysis work that mapped existing smailnail IMAP code to a future go-go-goja module and eval-style MCP host.
LastUpdated: 2026-03-08T22:52:20.580149837-04:00
WhatFor: Use this diary to understand how the design was derived, what evidence mattered, and how to reproduce the codebase scan.
WhenToUse: Read this when continuing the ticket or reviewing whether the resulting design is grounded in the current repositories.
---


# Investigation diary

## Goal

Record how the architecture assessment for a JavaScript IMAP API and sandboxed eval-style MCP was performed, what code paths were inspected, and why the final design recommends a service-layer plus native-module plus runtime-host split.

## Step 1: Create the research ticket and map the existing packages

I started by creating a fresh `docmgr` ticket workspace because the user asked for a long-form analysis package, not code changes. From there, I focused on the three repos that matter most for the requested feature shape: `smailnail` for domain logic, `go-go-goja` for module/runtime mechanics, and `jesus` for the closest local eval-plus-MCP implementation.

The first useful result was confirming that `smailnail` is already more reusable than its CLI shape suggests. The domain behavior is mostly in packages, while the command layer is mainly input decoding and output formatting. That means the right next step is not “teach Cobra to speak JS”, but “create a service layer that JS and MCP can both call.”

### Prompt Context

**User prompt (verbatim):** "Then create another ticket with docmgr: I want to build a JS API for the functionality of IMAP and expose it as a sandbox JS \"eval\" kind of MCP, see go-go-goja and potentially jesus/ . 

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new research ticket, inspect the relevant repos, and write a detailed intern-facing design for a JavaScript IMAP API plus sandboxed MCP runtime.

**Inferred user intent:** Define the architecture before implementation so the future build has clear package boundaries, a safe runtime model, and concrete file-level guidance.

**Commit (code):** N/A — documentation and ticket artifacts only

### What I did
- Created ticket `SMAILNAIL-004-JS-IMAP-EVAL-MCP-DESIGN` with `docmgr ticket create-ticket`
- Added the primary design doc and diary documents with `docmgr doc add`
- Inspected these files with line-numbered reads:
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/fetch_mail.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/mail_rules.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailgen/mailgen.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/modules/common.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/engine.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go`
- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/repl/model.go`

### Why
- `smailnail` defines the domain behavior that must become scriptable.
- `go-go-goja` defines the native module and runtime composition model.
- `jesus` defines the closest local example of sandboxed JS evaluation exposed through MCP.

### What worked
- The codebase split is clean enough to support a layered design.
- `go-go-goja` already provides the exact `NativeModule` abstraction needed.
- `jesus` provides a concrete MCP registration pattern that can be adapted instead of invented from scratch.

### What didn't work
- There was no single existing file in `smailnail` that already represented an application-service boundary, so I had to derive that boundary from package responsibilities instead of finding it directly.

### What I learned
- `smailnail` is closer to “API-ready” than “CLI-only” because `pkg/dsl` and `pkg/mailgen` already hold the core logic.
- The safe path is to introduce service packages before exposing any JS or MCP API.

### What was tricky to build
- The main design challenge was avoiding a false shortcut. The CLI command files show how to use the system, but they are not the right long-term API boundary. The tempting path would be to wrap command code and call it from JS. That would be fast initially and brittle later. I resolved this by mapping what those commands actually do, then promoting that behavior into a proposed service layer in the guide instead of treating the commands themselves as reusable APIs.

### What warrants a second pair of eyes
- The proposed session model. A reviewer should explicitly decide whether the first version should be ephemeral per invocation or persistent across invocations.
- The sandbox policy. A reviewer should verify that the recommended allowlist is strict enough for the intended threat model.

### What should be done in the future
- Implement the `smailnail` service layer before writing any Goja adapter code.
- Decide whether `mailgen` ships in the same first-version JS module or follows in a second phase.

### Code review instructions
- Start in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go` and `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go`
- Then review `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/modules/common.go` and `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go`
- Finally compare the proposed MCP shape against `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go`

### Technical details
- Core commands run:
```bash
find /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-004-JS-IMAP-EVAL-MCP-DESIGN--design-a-js-imap-api-and-sandboxed-eval-style-mcp-for-smailnail -maxdepth 3 -type f | sort
nl -ba /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go | sed -n '1,220p'
nl -ba /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go | sed -n '1,260p'
nl -ba /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/modules/common.go | sed -n '1,220p'
nl -ba /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go | sed -n '1,260p'
nl -ba /home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go | sed -n '1,260p'
```

## Step 2: Derive the module and sandbox architecture

After the package scan, I wrote the guide around one central conclusion: there should be a service layer below the JS adapter and above the existing domain code. That keeps the future JavaScript API, MCP server, and any later hosted web surface aligned.

I also used `jesus` as a narrow reference rather than a template to clone. It proves the viability of a Goja runtime plus MCP exposure, but it also carries unrelated concerns like HTTP route registration and application databases. The guide therefore extracts the useful parts of `jesus` and deliberately leaves the rest behind.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Turn the code scan into a detailed, implementation-ready architecture document.

**Inferred user intent:** Give an intern a document that explains not only what to build, but why the package boundaries and safety model should look the way they do.

**Commit (code):** N/A — documentation and ticket artifacts only

### What I did
- Wrote the main design doc with:
- current-state architecture
- gap analysis
- recommended package boundaries
- JS API proposal
- sandbox policy
- MCP tool contract
- implementation phases
- testing strategy
- review checklist
- Added ASCII diagrams, pseudocode, and API-shape examples
- Added a ticket-local script `scripts/architecture-scan.sh` to reproduce the evidence scan

### Why
- A design document without package boundaries turns into implementation drift.
- A sandbox proposal without an explicit policy turns into accidental capability creep.

### What worked
- The final guide now clearly separates:
- domain logic,
- service API,
- JS adapter,
- runtime host,
- MCP adapter.

### What didn't work
- I did not find a pre-existing “one true runtime host” package in the local repos that was minimal enough to adopt directly. The design therefore recommends borrowing patterns from `jesus`, not embedding `jesus` itself.

### What I learned
- The most stable shared interface is not the CLI, not the future MCP tool, and not the JS binding. It is the service-layer contract that both JS and MCP should call.

### What was tricky to build
- The biggest documentation challenge was writing something detailed enough for a new intern without hiding the core recommendation. The solution was to keep the document layered: current state first, gaps second, design third, file-level plan fourth. That preserves clarity while still giving concrete implementation direction.

### What warrants a second pair of eyes
- The proposed JS surface area. It should be reviewed for minimality before implementation starts.
- The question of whether to expose `mailgen` in the first module release.

### What should be done in the future
- Follow this design with an implementation ticket that creates:
- `pkg/service/...`
- `pkg/js/modules/smailnail/...`
- `pkg/jshost/...`
- `pkg/mcp/...`

### Code review instructions
- Review the design doc from top to bottom once.
- Then inspect the referenced files in the order listed in the guide’s references section.
- Then decide whether the suggested `executeIMAPJS` tool contract needs a sibling structured-tool set from day one.

### Technical details
- Reproducible scan script:
```bash
bash /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-004-JS-IMAP-EVAL-MCP-DESIGN--design-a-js-imap-api-and-sandboxed-eval-style-mcp-for-smailnail/scripts/architecture-scan.sh
```

## Quick Reference

- Recommended architecture:
```text
smailnail domain code -> service layer -> go-go-goja native module -> sandbox host -> MCP tool
```
- Recommended first module entrypoint:
```javascript
const smailnail = require("smailnail");
```
- Recommended first MCP tool:
```text
executeIMAPJS(code, connection?, connectionId?, timeoutMs?, maxMessages?)
```
- Recommended first implementation priority:
```text
service layer first, JS adapter second, runtime host third, MCP adapter fourth
```

## Usage Examples

Example architecture scan:

```bash
bash /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-004-JS-IMAP-EVAL-MCP-DESIGN--design-a-js-imap-api-and-sandboxed-eval-style-mcp-for-smailnail/scripts/architecture-scan.sh
```

Example review order:

```text
1. smailnail/pkg/imap/layer.go
2. smailnail/pkg/dsl/types.go
3. smailnail/pkg/dsl/processor.go
4. go-go-goja/modules/common.go
5. go-go-goja/engine/factory.go
6. jesus/pkg/mcp/server.go
```

## Related

- [../design-doc/01-js-imap-api-and-sandbox-eval-mcp-architecture-guide.md](../design-doc/01-js-imap-api-and-sandbox-eval-mcp-architecture-guide.md)
- [../tasks.md](../tasks.md)
