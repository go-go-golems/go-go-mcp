---
Title: Investigation diary
Ticket: SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN
Status: active
Topics:
    - smailnail
    - mcp
    - javascript
    - documentation
    - oidc
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-goja/pkg/jsdoc/model/model.go
      Note: Primary jsdoc model reviewed during the investigation
    - Path: go-go-mcp/ttmp/2026/03/09/SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN--design-imap-js-mcp-with-queryable-documentation/scripts/mcp-docs-architecture-scan.sh
      Note: Replays the high-value architecture evidence gathered during the investigation
    - Path: smailnail/pkg/js/modules/smailnail/module.go
      Note: Primary JS module reviewed during the investigation
ExternalSources: []
Summary: Chronological record of how the dedicated smailnail IMAP JS MCP and documentation architecture design was derived from the current codebase.
LastUpdated: 2026-03-09T00:43:02.075790567-04:00
WhatFor: Record the investigation and design decisions for the dedicated smailnail IMAP JS MCP and its documentation architecture.
WhenToUse: Use when reviewing how the design was derived, what evidence shaped it, and how to continue implementation.
---


# Diary

## Goal

Capture the chronological investigation behind the design of a dedicated IMAP JavaScript MCP for `smailnail` with exactly two tools: `executeIMAPJS` and `getIMAPJSDocumentation`.

## Step 1: Map the current JS runtime, service, and MCP boundaries

I started by locating the existing `smailnail` JavaScript surface and checking whether the project already had enough reusable runtime logic to justify a thin MCP adapter. It does. The current module and service split is already the right architectural boundary, which changes the design question from "how do we expose IMAP?" to "how do we expose the existing JS module safely and document it well?"

That first pass also established that documentation is currently the weak point rather than execution. The module is usable, but its self-description is just a one-line `Doc()` string. That is not enough for a model-facing eval tool, which made the documentation architecture the main design decision in this ticket.

### Prompt Context

**User prompt (verbatim):** "ok make an mcp that has only executeIMAPJS and getIMAPJSDocumentation. Use jsdocex or something similar to create and expose the documentation as something queryable. 

Create a new ticket for that and create a detailed implementation plan. I want a rich documentation querying tool. Maybe it makes sense to create jsdocex on the go side of things, maybe with reflection there too, for go go goja scenarios, or maybe just keep the doc in sync in a js example, etc... 

You figure it out, think hard, design alternatives. 

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a research-and-design ticket for a minimal `smailnail` MCP with only two tools, compare documentation architectures carefully, and produce a detailed intern-facing guide plus diary and upload.

**Inferred user intent:** Establish a low-risk implementation direction before writing code, with special attention to how JavaScript API documentation should be queried and kept in sync in Goja-based systems.

### What I did

- Created the ticket workspace with `docmgr ticket create-ticket --ticket SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN --title "Design IMAP JS MCP with queryable documentation" --topics smailnail,mcp,javascript,documentation,oidc`.
- Added the primary design doc and diary reference with `docmgr doc add`.
- Inspected:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/arguments.go`

### Why

- I needed to confirm whether the future MCP should embed behavior directly or simply host the existing JS module.
- I also needed a concrete argument-binding and tool-registration path before proposing any handler contracts.

### What worked

- The current `smailnail` module/service split is already clean enough for a thin MCP host.
- `jesus` provided a concrete reference for a single-purpose JS execution tool over MCP.
- `go-go-mcp/pkg/embeddable/arguments.go` confirmed that request structs can be bound cleanly without hand-decoding maps.

### What didn't work

- N/A in this step. The investigation confirmed the expected boundaries without command failures.

### What I learned

- The core implementation risk is not "how to execute JS." That part is relatively straightforward.
- The real architecture decision is how to author and query documentation without creating another drifting static reference page.

### What was tricky to build

- The tricky part was resisting the obvious but weak path of copying the `jesus` markdown approach. It looks simple, but it would have produced the wrong long-term system for a queryable docs tool. I had to separate "fastest thing to ship" from "best thing to maintain."

### What warrants a second pair of eyes

- Whether the future execution tool should allow raw connection credentials in the MCP request or force all connection logic to happen inside the JS code.

### What should be done in the future

- Implement the dedicated MCP host in `smailnail` rather than in `go-go-mcp`.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go` and `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go`.
- Validate that the exposed JS symbol surface is small enough to document and host cleanly.

### Technical details

- Key exports observed:
  - module-level: `parseRule`, `buildRule`, `newService`
  - service-level: `parseRule`, `buildRule`, `connect`

## Step 2: Evaluate documentation-system alternatives against the local codebase

Once the runtime boundary was clear, I focused on the documentation side. The crucial discovery was that the workspace already contains a migrated jsdocex-style stack inside `go-go-goja`, so building a fresh parallel system would likely create duplication rather than clarity.

I explicitly compared three documentation sources of truth: static markdown, Go reflection, and JavaScript sentinel docs. The result was a hybrid recommendation: JS-authored docs as the canonical source, plus Go-side validation for drift detection.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Determine whether the new documentation tool should be backed by static docs, reflection, or an extracted JS documentation model.

**Inferred user intent:** Avoid a weak v1 that paints the project into a corner, especially for model-facing documentation queries.

### What I did

- Inspected:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/cmd/goja-jsdoc/main.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/jsdoc/model/model.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/jsdoc/model/store.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/jsdoc/extract/extract.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/jsdoc/exportsq/exportsq.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/modules/glazehelp/glazehelp.go`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/doc/doc.go`
- Confirmed that `goja-jsdoc` already identifies itself as "jsdocex migrated into go-go-goja."
- Compared the structured `DocStore` model to the static `GetJavaScriptAPIReference()` pattern used by `jesus`.

### Why

- The user specifically asked whether it makes sense to create `jsdocex` on the Go side or keep docs in sync through JS examples. I needed evidence on what infrastructure already exists before recommending either path.

### What worked

- The local code strongly supports reusing `go-go-goja/pkg/jsdoc` rather than inventing a new docs stack.
- `DocStore` already has the right index model for a query tool: package, symbol, example, and concept indexes.
- `glazehelp` demonstrates a good query-oriented API shape worth copying conceptually.

### What didn't work

- There was no existing off-the-shelf query helper for `DocStore`; the indexes exist, but the future MCP will still need a local query layer.

### What I learned

- Static markdown is adequate for a reference page but wrong for the requested queryable documentation tool.
- Reflection is useful for validation but not strong enough to be the main prose source.
- The best answer is to author canonical docs in JS sentinel files and layer tests on top to catch drift.

### What was tricky to build

- The main nuance here was conceptual rather than mechanical. A pure Go reflection system sounds attractive because it feels more "authoritative," but the actual JS-facing API is defined through Goja export wiring and JSON field naming, not just through reflected Go method signatures. That makes reflection insufficient as a documentation authoring system even if it remains valuable for verification.

### What warrants a second pair of eyes

- Whether common query/registry helpers should be extracted into `go-go-goja` immediately or left local to the first `smailnail` implementation.

### What should be done in the future

- After the first `smailnail` integration, consider extracting generic `pkg/jsdoc/query`, `pkg/jsdoc/registry`, and `pkg/jsdoc/validate` helpers into `go-go-goja`.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/jsdoc/model/store.go` and `/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/jsdoc/extract/extract.go`.
- Confirm that the sentinel extraction model already supports symbol-level metadata, package-level prose, and examples.

### Technical details

- Sentinel patterns supported by the extractor:
  - `__package__`
  - `__doc__`
  - `__example__`
  - `doc\`...\``

## Step 3: Write the recommended architecture, tasks, and evidence script

With the evidence gathered, I wrote the detailed architecture guide around a minimal dedicated MCP with two tools and a hybrid documentation architecture. I also added a ticket-local shell script that replays the key file scans so another engineer can re-anchor the design quickly.

This is the point where the ticket became self-contained: the design doc explains what to build, the task list shows the concrete deliverables, and the scan script preserves the high-value evidence without requiring someone to reconstruct the shell history manually.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Turn the investigation into a detailed intern-facing design package and keep the work traceable.

**Inferred user intent:** Produce a ticket that can guide implementation directly and survive handoff to another engineer.

### What I did

- Wrote the detailed design doc in `design-doc/01-imap-js-mcp-and-queryable-documentation-architecture-guide.md`.
- Updated `tasks.md` with the completed research/design tasks.
- Added `scripts/mcp-docs-architecture-scan.sh` to replay the core evidence.
- Updated `index.md` to summarize the ticket and point to the key artifacts.

### Why

- The user asked for a detailed implementation plan and an intern-friendly system guide, so the ticket needed more than notes. It needed a concrete recommended architecture, API sketches, diagrams, and a phased implementation plan.

### What worked

- The design converged cleanly around a two-tool MCP in `smailnail`.
- The documentation architecture is now justified by concrete local code instead of by abstract preference.
- The scan script makes the ticket easier to audit later.

### What didn't work

- N/A in this step. The work was documentation and artifact creation rather than code execution against a runtime.

### What I learned

- The strongest design documents in this workspace are the ones that separate "canonical authoring source" from "verification source." That pattern applies cleanly here too.

### What was tricky to build

- The hardest part was keeping the implementation plan specific without prematurely over-extracting reusable framework pieces. The design now recommends a local `smailnail` implementation first and only then extracting generic `go-go-goja` helpers if they prove stable across use cases.

### What warrants a second pair of eyes

- The proposed request/response contracts for `getIMAPJSDocumentation`, especially whether `renderedMarkdown` belongs only in `render` mode or in all responses.

### What should be done in the future

- Run `docmgr doctor`.
- Upload the bundle to reMarkable and verify the remote listing.

### Code review instructions

- Read the design doc first.
- Then run `bash scripts/mcp-docs-architecture-scan.sh`.
- Finally review the proposed file layout and tool contracts against the current `smailnail` module/service split.

### Technical details

- The scan script is intentionally ticket-local so the design evidence is stored with the ticket rather than being implicit in terminal history.

## Step 4: Validate ticket hygiene and deliver the bundle to reMarkable

After the design content and file relations were in place, I switched to ticket hygiene and delivery. That meant running `docmgr doctor`, fixing a small vocabulary issue, and then doing a dry-run and real bundle upload to reMarkable.

This last step matters because the user asked for the deliverable to be stored in the ticket and uploaded, not just drafted locally. I also corrected a small remote-path typo during upload by re-uploading to the intended ticket directory and verifying the final remote listing there.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Complete the ticket as a deliverable, not merely as a local draft.

**Inferred user intent:** Ensure the documentation package is durable, validated, and accessible from the reMarkable destination used in prior tickets.

### What I did

- Ran `docmgr doctor --ticket SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN --stale-after 30`.
- Added missing vocabulary slugs:
  - `documentation`
  - `oidc`
- Reran `docmgr doctor` until it passed cleanly.
- Ran a dry-run upload with `remarquee upload bundle --dry-run ...`.
- Uploaded the bundle to `/ai/2026/03/09/SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN`.
- Verified the upload with `remarquee cloud ls /ai/2026/03/09/SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN --long --non-interactive`.

### Why

- The ticket needed to be complete and reproducible, not just written.
- `docmgr doctor` prevents avoidable metadata drift.
- The remote listing check proves the upload actually landed where the ticket says it did.

### What worked

- `doctor` passed after the two missing vocabulary entries were added.
- The dry-run upload rendered cleanly.
- The final bundle upload succeeded and the remote listing shows `SMAILNAIL-006 IMAP JS MCP Docs Design`.

### What didn't work

- My first real upload command used the wrong remote directory suffix: `/ai/2026/03/09/SMAILNAIL-006-IMAP-JS-MCP-DESIGN`.
- I corrected that immediately by uploading again to the intended path: `/ai/2026/03/09/SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN`.

### What I learned

- Running `doctor` serially after vocabulary updates is safer than doing it in parallel with the vocabulary writes.
- The bundle upload flow is reliable, but the remote directory string is easy to mistype when it closely resembles the ticket ID.

### What was tricky to build

- The tricky part here was operational precision rather than technical complexity. The ticket ID and desired remote directory were long and similar enough that a one-word omission changed the path. The fix was simple, but it reinforced that upload verification should be treated as mandatory rather than optional.

### What warrants a second pair of eyes

- N/A for code correctness. The remaining concern is only whether the stray upload in the typo path should be cleaned up later for repository hygiene on the reMarkable side.

### What should be done in the future

- Implement the design in a new execution ticket.
- Optionally remove the mistaken upload at `/ai/2026/03/09/SMAILNAIL-006-IMAP-JS-MCP-DESIGN` if keeping only one canonical remote copy matters.

### Code review instructions

- Confirm that `tasks.md` is fully checked off.
- Confirm that `docmgr doctor --ticket SMAILNAIL-006-IMAP-JS-MCP-DESIGN --stale-after 30` passes.
- Confirm that the remote listing contains the bundle under the intended ticket path.

### Technical details

- Final verified remote path: `/ai/2026/03/09/SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN`
- Verified remote listing entry: `SMAILNAIL-006 IMAP JS MCP Docs Design`

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
