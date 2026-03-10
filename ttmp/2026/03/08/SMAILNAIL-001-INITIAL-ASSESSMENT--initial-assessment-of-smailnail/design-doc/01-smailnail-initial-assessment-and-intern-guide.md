---
Title: smailnail initial assessment and intern guide
Ticket: SMAILNAIL-001-INITIAL-ASSESSMENT
Status: active
Topics:
    - smailnail
    - go
    - email
    - review
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: smailnail/README.md
      Note: Root documentation drift and repo-level onboarding problems
    - Path: smailnail/cmd/mailgen/cmds/generate.go
      Note: Mail storage path that appears to serialize malformed address headers
    - Path: smailnail/cmd/smailnail/main.go
      Note: Broken root CLI wiring still depends on deleted Glazed packages
    - Path: smailnail/examples/complex-search.yaml
      Note: Stale example that no longer matches the parser contract
    - Path: smailnail/pkg/dsl/actions.go
      Note: High-risk message mutation logic with likely UID versus sequence-number mismatch
    - Path: smailnail/pkg/dsl/processor.go
      Note: Core staged IMAP search and fetch pipeline that remains the architectural center of the repo
    - Path: smailnail/pkg/imap/layer.go
      Note: Shared IMAP settings layer still uses legacy Glazed layers and parameters APIs
ExternalSources: []
Summary: Detailed architecture review, code audit, and intern-oriented guide for the smailnail repository, including current breakage, correctness risks, experiments, and a remediation roadmap.
LastUpdated: 2026-03-08T19:51:55.399848144-04:00
WhatFor: Understand what smailnail is, what currently works, what is stale or broken, and how to approach modernization and repairs.
WhenToUse: Use when onboarding to smailnail, planning follow-up tickets, or reviewing the repository's current health and architecture.
---


# smailnail initial assessment and intern guide

## Executive Summary

`smailnail` is an older Go repository that bundles three related but distinct tools: an IMAP rule runner (`cmd/smailnail`), a templated mail generator (`cmd/mailgen`), and an IMAP helper CLI (`cmd/imap-tests`). The core IMAP rule engine in `pkg/dsl` is still structurally useful: it has a coherent rule model, a reasonably clear search builder, a multi-stage fetch pipeline, and enough examples to show the intended usage model. The mail generator in `pkg/mailgen` is also conceptually straightforward and internally compact.

The repository is not currently healthy in the checked-out workspace. The biggest immediate problem is Glazed API drift: the command wiring still imports deleted legacy packages such as `glazed/pkg/cmds/layers`, `glazed/pkg/cmds/parameters`, and `glazed/pkg/cmds/middlewares`, so the main binaries do not build or run. On top of that build breakage, the code review found two behavior-level risks worth treating as high priority once the repo compiles again: IMAP actions appear to apply UID values through sequence-number sets, and `mailgen` produces malformed address headers when appending generated mail to an IMAP mailbox.

What is still good:

- The repository has a clear subsystem split between rule parsing, IMAP search/fetch execution, and email generation.
- The DSL package is organized around small files with focused responsibilities.
- Example rule coverage is broad enough to explain the intended feature surface to a new engineer.
- The processor implements a sensible staged fetch model instead of naively downloading everything.

What is bad or stale:

- The command layer is pinned to pre-facade Glazed APIs and no longer compiles.
- Root and command-specific docs are heavily stale and point to wrong repo names, paths, and binaries.
- Example drift exists: not all examples match the parser contract.
- Test coverage is thin and misses the most failure-prone paths.

Recommended order of work:

1. Restore buildability by migrating all command wiring to the current Glazed facade APIs.
2. Fix correctness issues in IMAP action targeting and mail header creation.
3. Add regression tests around the parser corpus, action semantics, and mail generation.
4. Rewrite the top-level documentation after the actual runtime contract is stable again.

## Scope And Method

This assessment reviews `smailnail/` as it exists in the current workspace on March 8, 2026. The goal was not to modernize the repository in this ticket, but to answer five questions:

1. What the system is intended to do.
2. Which parts still appear architecturally sound.
3. Which parts are stale, broken, or risky.
4. Whether the code still works today.
5. What a new engineer should understand before attempting repairs.

Evidence sources used in this review:

- static code inspection of `cmd/`, `pkg/`, and `examples/`
- targeted build and help-command execution attempts
- ticket-local experiments in `scripts/`
- targeted inspection of upstream library behavior in the local module cache

## What Smailnail Is

At a high level, `smailnail` is an email tooling repo with two product-facing workflows and one support workflow:

- `cmd/smailnail`: consume a YAML rule, connect to IMAP, search/fetch messages, and render selected fields
- `cmd/mailgen`: generate synthetic emails from YAML templates and optionally write or append them to IMAP
- `cmd/imap-tests`: issue focused IMAP helper commands for manual testing and fixture setup

The architectural center of gravity is `pkg/dsl`. That package defines the rule language, converts DSL search expressions into go-imap criteria, fetches message metadata and content, and turns the resulting messages into output rows. The mail generation side is separate and much smaller, centered on `pkg/mailgen` and `pkg/types/config.go`.

## Repository Map

### Entrypoints

- `cmd/smailnail/main.go`: CLI root for the IMAP rule runner
- `cmd/smailnail/commands/mail_rules.go`: main rule execution command
- `cmd/smailnail/commands/fetch_mail.go`: lower-level message fetch command
- `cmd/mailgen/main.go`: CLI root for email generation
- `cmd/mailgen/cmds/generate.go`: generate/write/store workflow
- `cmd/imap-tests/main.go`: helper CLI for mailbox/message setup

### Core packages

- `pkg/dsl/types.go`: rule/search/output/action model and validation
- `pkg/dsl/parser.go`: YAML loading and top-level validation
- `pkg/dsl/search.go`: DSL-to-go-imap search conversion
- `pkg/dsl/processor.go`: staged search/fetch/pagination/content pipeline
- `pkg/dsl/fetch.go`: MIME-part selection policy and fetch option derivation
- `pkg/dsl/message.go`: internal `EmailMessage` materialization
- `pkg/dsl/output.go`: row formatting and output shaping
- `pkg/dsl/actions.go`: flag/copy/move/delete/export behavior
- `pkg/imap/layer.go`: shared IMAP connection settings and dial/login logic
- `pkg/mailgen/mailgen.go`: template-driven email generation
- `pkg/types/config.go`: mailgen YAML schema and validation

### Examples and docs

- `examples/smailnail/*.yaml`: main rule examples used to explain the DSL
- `examples/mailgen/*.yaml`: mail generation examples
- `examples/complex-search.yaml`: older multi-document example that no longer matches the parser
- `README.md`, `cmd/smailnail/README.md`, `examples/smailnail/QUICK-START.md`: documentation, currently stale

## Current Runtime Status

Observed status in this workspace:

- `go test ./...` fails
- `go build ./cmd/smailnail` fails
- `go build ./cmd/mailgen` fails
- `go run ./cmd/mailgen --help` fails

Representative current failures:

```text
cmd/smailnail/main.go:11:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/layers
cmd/smailnail/main.go:12:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/middlewares
cmd/smailnail/commands/fetch_mail.go:12:2: no required module provides package github.com/go-go-golems/glazed/pkg/cmds/parameters
```

This is not an incidental environment problem. The code directly imports deleted legacy Glazed packages in multiple command entrypoints and shared layers, including `cmd/smailnail/main.go:9-12`, `pkg/imap/layer.go:7-10`, and `cmd/smailnail/commands/mail_rules.go:11-16`.

## Architecture Walkthrough For A New Intern

### 1. Rule-driven IMAP flow

The main product idea is: define what messages to search for in YAML, then let the binary translate that into IMAP operations and print only the fields you care about.

Conceptual flow:

```text
rule.yaml
   |
   v
pkg/dsl/parser.go
   |
   v
pkg/dsl/types.go validation
   |
   v
pkg/dsl/search.go -> go-imap SearchCriteria + SearchOptions
   |
   v
pkg/dsl/processor.go
   |
   +--> IMAP SEARCH
   +--> IMAP FETCH metadata/structure
   +--> decide needed MIME parts
   +--> IMAP FETCH body sections
   |
   v
pkg/dsl/message.go -> EmailMessage
   |
   v
pkg/dsl/output.go / command row creation
```

The critical call path begins in `cmd/smailnail/commands/mail_rules.go:89-131`. That command parses the YAML rule, validates credentials, connects to IMAP through `IMAPSettings.ConnectToIMAPServer`, selects a mailbox, and then calls `rule.FetchMessages(client)`.

`pkg/dsl/processor.go` is the most important file to understand. It implements a staged workflow rather than a single fetch:

- build search criteria (`processor.go:65-75`)
- execute search and inspect server return shape (`processor.go:77-103`)
- compensate when the server returns count without sequence numbers (`processor.go:109-178`)
- apply offset/limit and build the fetch set (`processor.go:180-217`)
- fetch message metadata and MIME structure
- decide which MIME parts actually need content fetches
- perform batched body fetches
- materialize final `EmailMessage` objects for output

That staged design is one of the better parts of the repo. It shows the author was trying to minimize network round trips and avoid downloading unnecessary content.

### 2. DSL rule model

The DSL is expressed as YAML and parsed into a `Rule`. The parser is intentionally thin in `pkg/dsl/parser.go:10-42`:

- read the file
- unmarshal once into a `Rule`
- call `rule.Validate()`
- apply a default output format if missing

That means the actual contract lives in `pkg/dsl/types.go`, not in the parser. When you need to understand which fields are legal, or why a rule is being rejected, start in `types.go` first and only then look at the parser.

Pseudo-code for how the parser behaves:

```text
func ParseRuleFile(path):
    bytes = read(path)
    rule = yaml.Unmarshal(bytes)
    validate(rule)
    if rule.output.format is empty:
        rule.output.format = "text"
    return rule
```

Important implication: because the parser performs a single `yaml.Unmarshal` into one struct, multi-document YAML is not supported by the current implementation.

### 3. Search compilation

`pkg/dsl/search.go` is the compiler from the YAML DSL into go-imap search structures. The file has two jobs:

- map simple fields like `from`, `subject_contains`, `since`, and flags into search criteria
- recursively build compound boolean conditions (`and`, `or`, `not`)

When debugging “the rule parses but returns wrong messages”, this file is usually where the semantic mismatch will be.

Pseudo-code:

```text
func BuildSearchCriteria(search, output):
    criteria = new SearchCriteria
    applySimpleFields(criteria, search)
    if search has nested conditions:
        criteria = combine(criteria, buildConditionTree(search.conditions))
    options = deriveSearchOptionsFromOutput(output)
    return criteria, options
```

### 4. MIME-part planning and fetch behavior

Message content retrieval is split across `pkg/dsl/fetch.go`, `pkg/dsl/processor.go`, and `pkg/dsl/message.go`.

The design intent appears to be:

- fetch headers/envelope/structure first
- inspect body structure to find relevant parts
- fetch only the MIME bodies actually requested by the output spec

This is the right architectural shape for IMAP. It lets the rule ask for `text/plain`, or HTML, or concatenated MIME output, without treating every message as “download the entire raw email”.

### 5. Action execution

`pkg/dsl/actions.go` applies write operations after message selection:

- add/remove flags
- copy
- move
- delete / move to Trash
- export

This subsystem is operationally sensitive because it mutates server state. That is why the UID/sequence-number bug described below matters so much: a mistake here is not just a bad display result, it can move or delete the wrong mail.

### 6. Mail generation subsystem

`pkg/mailgen/mailgen.go` and `pkg/types/config.go` are separate from the IMAP DSL flow. They implement synthetic email generation from YAML templates and variations.

Conceptual flow:

```text
mailgen config.yaml
   |
   v
pkg/types/config.go Validate()
   |
   v
pkg/mailgen/mailgen.go Generate()
   |
   +--> process variation templates
   +--> build execution context
   +--> render subject/from/to/body
   |
   v
[]types.Email
   |
   +--> print rows
   +--> write files
   +--> append to IMAP mailbox
```

The generator is compact and easy to understand. It validates the config first (`mailgen.go:53-57`), loops through `generate` entries (`mailgen.go:61-67`), renders per-variation values (`mailgen.go:72-89`), then renders the final email template (`mailgen.go:117-176`).

## Code Review Findings

### 1. Build is broken across the command layer due to legacy Glazed imports

Severity: high

Evidence:

- `cmd/smailnail/main.go:9-12` imports legacy `layers` and `middlewares`
- `pkg/imap/layer.go:8-9` imports legacy `layers` and `parameters`
- `cmd/smailnail/commands/mail_rules.go:11-16` imports legacy command interfaces and middleware types
- `go test ./...` fails immediately with missing-module errors for those deleted packages

Why it matters:

- The main binaries are not runnable.
- Any architectural strengths in `pkg/dsl` are currently inaccessible from the intended CLIs.
- This is also a documentation problem, because the docs imply the tool is usable when it is not.

What to do now:

1. Migrate all command definitions to the newer Glazed facade APIs.
2. Replace legacy struct tags and layer construction in `pkg/imap/layer.go`.
3. Remove deprecated `clay.InitViper` patterns while touching the roots if the current Clay setup has a supported replacement.

### 2. IMAP actions use UID values inside sequence-number sets

Severity: high

Evidence in `pkg/dsl/actions.go`:

- `executeFlags`: `uidSet.AddNum(uint32(msg.UID))` at `actions.go:80-84`
- `executeCopy`: `uidSet.AddNum(uint32(msg.UID))` at `actions.go:142-145`
- `executeMove`: `uidSet.AddNum(uint32(msg.UID))` at `actions.go:166-169`
- `executeDelete`: `uidSet.AddNum(uint32(msg.UID))` at `actions.go:211-214`

Why this is risky:

- The code builds `imap.SeqSet`, not a UID set.
- go-imap v2 chooses `STORE` vs `UID STORE`, `COPY` vs `UID COPY`, and `MOVE` vs `UID MOVE` based on the number-set kind.
- If UID values are sent as sequence numbers, the operation can target different messages from the ones the user inspected.

Observed inference from source inspection:

- The current code appears to intend UID-based safety.
- The actual API objects being used encode sequence semantics instead.

What to do now:

1. Switch the action layer to the correct UID number-set type.
2. Add regression tests for copy/move/delete/flags set construction.
3. Validate against a controlled IMAP fixture before trusting destructive actions.

### 3. `mailgen` writes malformed address headers when storing generated mail to IMAP

Severity: medium-high

Evidence in `cmd/mailgen/cmds/generate.go:210-224`:

- `SetAddressList("From", []*mail.Address{{Address: email.From}})`
- same pattern for `To`, `Cc`, `Bcc`, `Reply-To`

Why this is wrong:

- `email.From` and friends are rendered strings such as `"John Doe <john@example.com>"`
- `mail.Address.Address` expects the mailbox address field, not the full display string
- The code should parse the rendered address string into `mail.Address{Name, Address}` before setting the header

Experiment result from `scripts/address-header-experiment.sh`:

```text
from: <"John Doe <john"@example.com>>
to: <user@example.com>
```

That output is malformed for the display-name case.

What to do now:

1. Parse rendered address strings with the standard library address parser before calling `SetAddressList`.
2. Add mailgen regression tests that assert exact serialized headers.
3. Decide whether template outputs may be address lists or only single addresses.

### 4. Top-level documentation is badly stale and can mislead repairs

Severity: medium

Evidence:

- `README.md:1-22` documents `mailgen`, not the repo as a whole
- `README.md:83-130` contains duplicated ASCII-art noise
- `cmd/smailnail/README.md:24-29` tells the reader to clone `go-go-labs` and build `./cmd/apps/mail-app-rules`
- `examples/smailnail/QUICK-START.md:15-20` repeats those obsolete paths

Why it matters:

- A new engineer will waste time on non-existent paths before even reaching the real code.
- Documentation drift makes it harder to separate intentional design from archaeological leftovers.

What to do now:

1. Replace the root README with a repo overview.
2. Move product-specific usage docs under each CLI directory and make them match actual binaries.
3. Remove decorative noise that obscures actual content.

### 5. Example corpus is partially stale relative to the parser contract

Severity: medium

Evidence:

- `pkg/dsl/parser.go:22-42` unmarshals a single `Rule`
- `examples/complex-search.yaml:1-4` starts with a metadata-only YAML document, followed by more documents
- `scripts/parse-rule-examples.sh` passes all `examples/smailnail/*.yaml` but fails on `examples/complex-search.yaml`

Observed output:

```text
FAIL examples/complex-search.yaml: invalid output config: at least one output field is required
```

Why it matters:

- Examples are often treated as executable documentation.
- If examples do not match the real parser contract, new engineers will debug the wrong problem.

What to do now:

1. Either delete/replace the unsupported multi-document example.
2. Or deliberately add multi-document support, if that format is worth preserving.
3. Add an example-corpus parse test so drift shows up in CI.

### 6. Test coverage is too thin for the highest-risk paths

Severity: medium

Observed state:

- only one test file exists: `pkg/dsl/search_test.go`
- no direct tests for parser corpus loading, destructive actions, processor pagination, IMAP layer setup, or mailgen serialization

Why it matters:

- The most fragile code paths are precisely the ones that lack tests.
- Without regression tests, the upcoming Glazed migration and bug fixes will be harder to trust.

## What Is Good And Worth Keeping

Not everything here is stale. Several design choices are sound enough to preserve through modernization:

- `pkg/dsl` is separated by concern instead of being one large file.
- The staged fetch model in `pkg/dsl/processor.go` is a sensible IMAP optimization.
- Mail generation validation in `pkg/types/config.go` is compact and readable.
- Examples under `examples/smailnail/` cover a broad enough set of search cases to serve as future test fixtures.
- The repo already has a distinct support CLI (`imap-tests`) that can become part of a proper fixture/testing story.

## What Works Today

The following narrow slice still works in isolation:

- `pkg/dsl` search tests pass
- the parser can successfully load the rule files under `examples/smailnail/`
- `pkg/mailgen` template generation logic is structurally understandable and likely salvageable with limited repairs

The following does not work today in normal use:

- building the user-facing CLIs
- running help commands for `smailnail` or `mailgen`
- trusting IMAP action semantics
- trusting IMAP-stored mailgen headers for display-name addresses

## Suggested Remediation Plan

### Phase 1: Restore buildability

Goal: make the binaries compile again with current Glazed.

Files to start with:

- `cmd/smailnail/main.go`
- `cmd/smailnail/commands/mail_rules.go`
- `cmd/smailnail/commands/fetch_mail.go`
- `cmd/mailgen/main.go`
- `cmd/mailgen/cmds/generate.go`
- `cmd/imap-tests/main.go`
- `cmd/imap-tests/commands/*.go`
- `pkg/imap/layer.go`

Deliverable:

- `go test ./...`
- `go build ./cmd/smailnail ./cmd/mailgen ./cmd/imap-tests`
- working `--help` output for all three binaries

### Phase 2: Fix behavioral correctness issues

Goal: make destructive operations and generated IMAP mail safe to trust.

Tasks:

1. Fix UID-vs-sequence semantics in `pkg/dsl/actions.go`.
2. Fix address header parsing in `cmd/mailgen/cmds/generate.go`.
3. Add tests for those exact bugs.

### Phase 3: Lock down examples and docs

Goal: make documentation executable and trustworthy.

Tasks:

1. Convert valid example rules into parse tests.
2. Remove or rewrite `examples/complex-search.yaml`.
3. Rewrite the root README and the stale quick-start docs.

### Phase 4: Add runtime confidence

Goal: verify behavior against a controlled IMAP server or test fixture.

Tasks:

1. Use `cmd/imap-tests` to create fixture mailboxes/messages.
2. Run end-to-end fetch and action scenarios in scripts.
3. Add smoke scripts that can be rerun by future maintainers.

## Testing Strategy

Immediate validation gates to add:

1. Build and CLI smoke:
   - `go test ./... -count=1`
   - `go build ./cmd/smailnail ./cmd/mailgen ./cmd/imap-tests`
   - `go run ./cmd/smailnail --help`
   - `go run ./cmd/mailgen --help`
2. Parser corpus tests:
   - parse every file in `examples/smailnail/`
   - explicitly assert the status of `examples/complex-search.yaml`
3. Mailgen serialization tests:
   - single address
   - display-name address
   - multiple address-list policy if supported
4. IMAP action tests:
   - verify the correct number-set type is used
   - verify copy/move/delete/flag commands operate on the intended identifiers
5. Processor tests:
   - offset/limit behavior
   - server returns count but no sequence numbers
   - MIME-part content selection

Suggested end-to-end fixture flow:

```text
setup mailbox fixtures
    |
    +--> append plain text mail
    +--> append html mail
    +--> append attachment mail
    |
run rule examples
    |
    +--> validate output fields
    +--> validate pagination
    +--> validate MIME part inclusion
    |
run action scenarios
    |
    +--> copy
    +--> move
    +--> delete
    +--> flag add/remove
```

## Intern Onboarding Guide

If you are a new intern asked to work on `smailnail`, use this sequence:

1. Start with `pkg/dsl/types.go` and `pkg/dsl/parser.go` to understand the rule contract.
2. Read `pkg/dsl/search.go` to see how a rule becomes IMAP search criteria.
3. Read `pkg/dsl/processor.go` to understand the staged fetch pipeline.
4. Read `pkg/dsl/actions.go` only after you understand how messages are identified.
5. Read `pkg/mailgen/mailgen.go` and `pkg/types/config.go` separately; it is a different subsystem.
6. Treat the README files as historical artifacts until they are rewritten.

When debugging, ask these questions in order:

1. Did the rule parse and validate?
2. Did the search compiler translate the intended semantics?
3. Did the processor fetch the right identifiers?
4. Did output rendering lose information?
5. If the command mutates mail, are we operating on UIDs or sequence numbers?

Useful mental model:

```text
Rule authoring problem?
    -> parser.go + types.go

Wrong messages returned?
    -> search.go + processor.go

Wrong content in output?
    -> fetch.go + message.go + output.go

Wrong messages moved/deleted?
    -> actions.go first

Generated mail looks malformed?
    -> mailgen.go + cmd/mailgen/cmds/generate.go
```

## Risks And Open Questions

- The Glazed migration scope touches every CLI, so there is a moderate risk of fixing one binary while leaving another behind.
- The IMAP action bug is inferred from code and library API contracts; it should still be verified against a controlled mailbox before declaring it fully resolved.
- The repo may contain additional stale examples or hidden runtime drift not exposed by the current minimal tests.
- It is not yet clear whether the intended future is one combined repo with three CLIs, or a narrower product focus around only one of them.

## Recommended Immediate Ticket Follow-ups

1. `SMAILNAIL-002-GLAZED-FACADE-MIGRATION`
2. `SMAILNAIL-003-IMAP-ACTION-CORRECTNESS`
3. `SMAILNAIL-004-DOCS-AND-EXAMPLES-RECONCILIATION`
4. `SMAILNAIL-005-RUNTIME-FIXTURE-SMOKE-TESTS`

## References

Primary files reviewed:

- `cmd/smailnail/main.go`
- `cmd/smailnail/commands/mail_rules.go`
- `cmd/smailnail/commands/fetch_mail.go`
- `cmd/mailgen/main.go`
- `cmd/mailgen/cmds/generate.go`
- `cmd/imap-tests/main.go`
- `pkg/imap/layer.go`
- `pkg/dsl/parser.go`
- `pkg/dsl/types.go`
- `pkg/dsl/search.go`
- `pkg/dsl/fetch.go`
- `pkg/dsl/processor.go`
- `pkg/dsl/message.go`
- `pkg/dsl/output.go`
- `pkg/dsl/actions.go`
- `pkg/mailgen/mailgen.go`
- `pkg/types/config.go`
- `README.md`
- `cmd/smailnail/README.md`
- `examples/smailnail/QUICK-START.md`
- `examples/smailnail/*.yaml`
- `examples/complex-search.yaml`

Ticket-local experiments:

- `scripts/parse-rule-examples.sh`
- `scripts/address-header-experiment.sh`
