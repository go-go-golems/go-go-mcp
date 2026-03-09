---
Title: Investigation diary
Ticket: SMAILNAIL-001-INITIAL-ASSESSMENT
Status: active
Topics:
    - smailnail
    - go
    - email
    - review
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-mcp/ttmp/2026/03/08/SMAILNAIL-001-INITIAL-ASSESSMENT--initial-assessment-of-smailnail/scripts/address-header-experiment.sh
      Note: Ticket-local address header serialization experiment used during the assessment
    - Path: go-go-mcp/ttmp/2026/03/08/SMAILNAIL-001-INITIAL-ASSESSMENT--initial-assessment-of-smailnail/scripts/parse-rule-examples.sh
      Note: Ticket-local parser corpus experiment used during the assessment
    - Path: smailnail/cmd/mailgen/cmds/generate.go
      Note: Source file tied to the mail header experiment and related diary notes
    - Path: smailnail/pkg/dsl/parser.go
      Note: Parser contract reviewed and exercised by the example-corpus experiment
ExternalSources: []
Summary: Chronological diary of the smailnail assessment, including ticket setup, repository inventory, build/runtime experiments, path issues, and reporting decisions.
LastUpdated: 2026-03-08T19:51:55.476781419-04:00
WhatFor: Provide a continuation-friendly record of how the smailnail assessment was performed and what evidence supported the conclusions.
WhenToUse: Use when reviewing the investigation process, reproducing experiments, or continuing the assessment in follow-up tickets.
---


# Investigation diary

## Goal

This diary records the chronological investigation of `smailnail/` for ticket `SMAILNAIL-001-INITIAL-ASSESSMENT`. It captures what I inspected, what worked, what failed, what experiments I ran, and what conclusions those steps justified.

## Step 1: Create The Ticket And Establish Scope

I started by creating a new ticket workspace for a fresh `smailnail` assessment instead of reusing the earlier `go-go-mcp` tickets. The immediate goal was to set up the standard ticket structure, then verify where `docmgr` believed the documentation root lived so I would not accidentally write to the wrong place.

This step also clarified an important repo-structure detail: the shared `.ttmp.yaml` at the workspace root points `docmgr` at `go-go-mcp/ttmp`, so this `smailnail` ticket lives under that docs root even though the code under review is in `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail`.

### Prompt Context

**User prompt (verbatim):** "Ok, now make a new ticket for smailnail initial assessment, and do an in depth analysis and code review of smailnail/ . Keep a frequent diary as you go, feel free to run experiments in thes cripts/ folder of the ticket, and write a detailed report.

Create a detailed analysis / review / user guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new ticketed assessment for `smailnail`, keep a detailed diary during the work, run experiments as needed from the ticket `scripts/` directory, write a very detailed analysis/intern guide, and deliver the final bundle to reMarkable.

**Inferred user intent:** Produce an evidence-backed, continuation-friendly audit of `smailnail` that explains the codebase to a new engineer and surfaces the highest-value modernization and correctness work.

### What I did
- Ran `docmgr status --summary-only` to confirm the active docs root and config.
- Created ticket `SMAILNAIL-001-INITIAL-ASSESSMENT`.
- Added the primary design doc and diary doc.
- Added ticket tasks for architecture inventory, experiments, writing, and delivery.

### Why
- I needed a stable place to store notes, scripts, and final deliverables before touching any analysis work.
- Confirming the docs root early avoids subtle ticket-placement errors later.

### What worked
- `docmgr` was already configured and ticket creation succeeded cleanly.
- The ticket structure was created in the expected docs tree.

### What didn't work
- N/A

### What I learned
- `docmgr` is workspace-scoped here, not repo-scoped to `smailnail`.
- Ticket paths need to be handled carefully when scripts infer the code repository location.

### What was tricky to build
- The main subtlety was path topology, not code. The docs root lives under `go-go-mcp/ttmp`, while the code under review lives in sibling repo `smailnail/`. That makes naive “walk up to repo root” script logic easy to get wrong.

### What warrants a second pair of eyes
- Only the doc-root placement if someone expects each repo to own its own `ttmp`.

### What should be done in the future
- Consider whether the workspace-level `docmgr` layout should be documented more explicitly for multi-repo workspaces.

### Code review instructions
- Verify the ticket exists and includes the standard scaffold files.
- Confirm the docs root with `docmgr status --summary-only`.

### Technical details
- Ticket path: `go-go-mcp/ttmp/2026/03/08/SMAILNAIL-001-INITIAL-ASSESSMENT--initial-assessment-of-smailnail`

## Step 2: Inventory Repository Structure And Build Health

With the ticket in place, I moved to code discovery and build validation. The goal in this step was to separate structural understanding from behavioral speculation: identify entrypoints, core packages, examples, and tests, then run the minimum build/test commands needed to answer whether the repo still works at all.

This step immediately showed that the repo’s central runtime problem is Glazed drift, not subtle business logic. The command layer imports packages that do not exist in the current workspace version of Glazed, so the main binaries do not build.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Build a concrete architecture map and determine whether the repo still runs in the current workspace.

**Inferred user intent:** Establish a trustworthy baseline before deeper code review, so the final report distinguishes “good architecture trapped behind build drift” from “currently operational code.”

### What I did
- Ran `rg --files cmd pkg examples | sort` in `smailnail/` to map the repository layout.
- Inspected key files:
  - `cmd/smailnail/main.go`
  - `cmd/smailnail/commands/mail_rules.go`
  - `cmd/mailgen/main.go`
  - `cmd/imap-tests/main.go`
  - `pkg/imap/layer.go`
  - `pkg/dsl/*`
  - `pkg/mailgen/mailgen.go`
  - `pkg/types/config.go`
- Ran:
  - `cd smailnail && go test ./...`
  - `cd smailnail && go build ./cmd/smailnail`
  - `cd smailnail && go run ./cmd/mailgen --help`
- Counted test files with `rg --files -g '*_test.go'`.

### Why
- I needed to know whether the repo was fundamentally runnable before spending time on examples and docs.
- File inventory helps orient the final intern guide around actual package boundaries.

### What worked
- Static inspection clearly exposed the subsystem boundaries.
- `pkg/dsl/search_test.go` still passes through `go test` even though the full repo fails.

### What didn't work
- `go test ./...` failed with missing-module errors:

```text
cmd/imap-tests/commands/create_mailbox.go:10:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/layers
cmd/imap-tests/commands/create_mailbox.go:11:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/parameters
cmd/mailgen/main.go:9:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/middlewares
```

- `go build ./cmd/smailnail` failed with the same root cause:

```text
cmd/smailnail/main.go:11:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/layers
cmd/smailnail/main.go:12:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/middlewares
cmd/smailnail/commands/fetch_mail.go:12:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/parameters
```

- `go run ./cmd/mailgen --help` failed for the same reason.

### What I learned
- The repo is broken at the CLI integration layer, not in the core DSL package.
- Only one `_test.go` file exists in the whole repo: `pkg/dsl/search_test.go`.

### What was tricky to build
- The tricky part was separating compile-breakage from logic-breakage. Once the Glazed failures appeared, it would have been easy to stop there. Instead, I kept digging into `pkg/dsl` and `pkg/mailgen` because the user asked for an in-depth assessment, not just a build-error note.

### What warrants a second pair of eyes
- The eventual migration will touch every CLI, so it is worth reviewing whether all three command trees should continue to exist together.

### What should be done in the future
- Create a dedicated migration ticket for current Glazed facade APIs.

### Code review instructions
- Start with `cmd/smailnail/main.go` and `pkg/imap/layer.go` to see the legacy imports quickly.
- Re-run `go test ./...` and `go build ./cmd/smailnail` to confirm the current failure mode.

### Technical details
- Command roots reviewed: `cmd/smailnail`, `cmd/mailgen`, `cmd/imap-tests`
- Core runtime package reviewed: `pkg/dsl`

## Step 3: Run Targeted Experiments On Examples And Header Serialization

After confirming the command layer is broken, I shifted to targeted experiments that could still run without fully restoring the CLI. I created ticket-local scripts to answer two concrete questions: whether the rule examples still match the parser contract, and whether `mailgen`’s IMAP storage path serializes email addresses correctly.

Both experiments were useful. They exposed one stale example contract and one concrete header-formatting bug that would otherwise remain a code-review suspicion rather than observed behavior.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Use the ticket `scripts/` folder for real experiments that validate current assumptions.

**Inferred user intent:** Move past static reading and prove at least some of the suspected issues through runnable checks.

### What I did
- Added `scripts/parse-rule-examples.sh` to compile and run a small Go helper that calls `dsl.ParseRuleFile(...)` on the example corpus.
- Added `scripts/address-header-experiment.sh` to isolate `mail.Header.SetAddressList(...)` behavior with the address-string shape generated by `mailgen`.
- Ran both scripts.

### Why
- Example drift and header serialization were both cheap to test without fixing the full CLI stack.
- Keeping the experiments in the ticket makes the assessment reproducible for the next engineer.

### What worked
- The parser experiment passed for all files under `examples/smailnail/`.
- The parser experiment failed for `examples/complex-search.yaml` with:

```text
FAIL examples/complex-search.yaml: invalid output config: at least one output field is required
```

- The address experiment produced:

```text
from: <"John Doe <john"@example.com>>
to: <user@example.com>
```

That strongly supports the conclusion that the current IMAP-store header path mishandles display-name addresses.

### What didn't work
- The first version of the ticket scripts resolved the repo root incorrectly because the ticket lives under `go-go-mcp/ttmp`, not under `smailnail`.
- I initially climbed too few `..` segments and the scripts tried to enter a non-existent `.../go-go-mcp/smailnail` path.

### What I learned
- `examples/smailnail/*.yaml` are still valuable as future regression fixtures.
- The top-level `examples/complex-search.yaml` is stale relative to the current single-document parser.
- The `mailgen` bug is not theoretical; the current address serialization really is malformed for display-name input.

### What was tricky to build
- The path correction in the scripts was the main sharp edge. Because the ticket lives in a different repo subtree than the code, script portability depends on explicitly deriving the workspace root first and only then entering `smailnail/`.

### What warrants a second pair of eyes
- The exact future policy for address parsing: single-address only versus proper comma-separated lists.

### What should be done in the future
- Promote the successful example-parser experiment into a real Go test.
- Add mail header serialization tests before fixing the bug.

### Code review instructions
- Read the scripts in the ticket `scripts/` folder.
- Re-run them directly from the workspace root to confirm behavior.

### Technical details
- `scripts/parse-rule-examples.sh`
- `scripts/address-header-experiment.sh`

## Step 4: Review Behavioral Risk Areas And Write The Final Assessment

The last investigation step focused on correctness risks that are easy to miss if you only look at build errors. I reviewed the IMAP action path, the parser, the docs, and the example corpus to decide which issues belong in the final report as immediate work rather than background noise.

The highest-signal conclusion from this step is that the codebase has one “stop everything” infrastructure issue and two genuine product-correctness issues. The infrastructure issue is Glazed drift. The product-correctness issues are the action identifier mismatch in `pkg/dsl/actions.go` and malformed stored headers in `cmd/mailgen/cmds/generate.go`.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Finish the in-depth code review and turn it into an intern-friendly assessment.

**Inferred user intent:** Produce a final document that helps someone new to the repo understand both how it is supposed to work and where it is unsafe or outdated.

### What I did
- Inspected `pkg/dsl/actions.go` for identifier handling across flag/copy/move/delete paths.
- Cross-checked local go-imap v2 API usage expectations in the module cache.
- Inspected stale docs:
  - `README.md`
  - `cmd/smailnail/README.md`
  - `examples/smailnail/QUICK-START.md`
- Wrote the design doc with:
  - executive summary
  - architecture map
  - code review findings
  - intern onboarding guidance
  - phased remediation plan
  - testing strategy

### Why
- The final report needed to prioritize issues by operational impact, not by whichever files looked oldest.
- The user explicitly asked for an intern-oriented guide, not only a bug list.

### What worked
- The evidence supports a coherent story:
  - the DSL core is salvageable
  - the command layer is outdated
  - docs are stale
  - examples are partly trustworthy
  - some mutation paths should not be trusted until repaired

### What didn't work
- N/A for this step; this was analysis and documentation work.

### What I learned
- `smailnail` is best understood as three CLIs sharing a good-enough core rather than as one polished product.
- The repo’s strongest reusable asset is the staged IMAP fetch logic in `pkg/dsl/processor.go`.

### What was tricky to build
- The main challenge was balancing “intern guide” clarity with “code review” sharpness. The report needed to explain the system sympathetically enough for onboarding, while still being direct about the parts that are broken or risky.

### What warrants a second pair of eyes
- The UID-versus-sequence-number interpretation in `pkg/dsl/actions.go` should be validated against a controlled mailbox fixture when repaired.
- The exact modernization target for Clay/Glazed initialization should be reviewed against the current preferred patterns in those repos.

### What should be done in the future
- Create follow-up tickets for migration, correctness fixes, docs/examples reconciliation, and runtime smoke coverage.

### Code review instructions
- Read the design doc first.
- Then inspect `pkg/dsl/actions.go`, `cmd/mailgen/cmds/generate.go`, and the stale README files as the highest-signal evidence for the major findings.

### Technical details
- Final assessment doc: `design-doc/01-smailnail-initial-assessment-and-intern-guide.md`

## Quick Reference

Key commands used during the investigation:

```bash
docmgr status --summary-only
docmgr ticket create-ticket --ticket SMAILNAIL-001-INITIAL-ASSESSMENT --title "Initial assessment of smailnail" --topics smailnail,go,email,review
docmgr doc add --ticket SMAILNAIL-001-INITIAL-ASSESSMENT --doc-type design-doc --title "smailnail initial assessment and intern guide"
docmgr doc add --ticket SMAILNAIL-001-INITIAL-ASSESSMENT --doc-type reference --title "Investigation diary"

cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
rg --files cmd pkg examples | sort
go test ./...
go build ./cmd/smailnail
go run ./cmd/mailgen --help

/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-001-INITIAL-ASSESSMENT--initial-assessment-of-smailnail/scripts/parse-rule-examples.sh
/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/SMAILNAIL-001-INITIAL-ASSESSMENT--initial-assessment-of-smailnail/scripts/address-header-experiment.sh
```

## Related

- `design-doc/01-smailnail-initial-assessment-and-intern-guide.md`
