---
Title: Port smailnail to glazed facade packages and fix IMAP correctness issues
Ticket: SMAILNAIL-002-FACADE-MIGRATION
Status: active
Topics:
    - smailnail
    - glazed
    - refactor
    - go
    - email
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
      Note: Repository under migration and validation for this ticket
ExternalSources: []
Summary: Tracks the completed migration of smailnail to Glazed facade APIs, the IMAP correctness fixes, the updated docs, and the Docker-backed validation evidence.
LastUpdated: 2026-03-08T20:21:22.43675072-04:00
WhatFor: Use this ticket to review what changed in the smailnail facade migration, why the behavior fixes were necessary, and how the Docker validation was performed.
WhenToUse: Use when reviewing the migration implementation, reproducing the IMAP validation flow, or onboarding someone to the repaired smailnail command tree.
---


# Port smailnail to glazed facade packages and fix IMAP correctness issues

## Overview

This ticket migrated `smailnail` from removed Glazed legacy APIs to the current facade packages, repaired two correctness bugs that could target the wrong IMAP messages or generate malformed address headers, updated the user-facing docs, and validated the result against the Docker Dovecot fixture at `/home/manuel/code/others/docker-test-dovecot`.

The code changes are complete in the `smailnail` repository. The migration landed in commit `dbc9e00` and the correctness fixes plus runtime hardening landed in commit `cd446d2`.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Implementation guide**: [design-doc/01-smailnail-facade-migration-and-correctness-implementation-guide.md](./design-doc/01-smailnail-facade-migration-and-correctness-implementation-guide.md)
- **Diary**: [reference/01-implementation-diary.md](./reference/01-implementation-diary.md)
- **Validation script**: [scripts/docker-imap-validation.sh](./scripts/docker-imap-validation.sh)

## Status

Current status: **active**

Completed work:

- migrated all three CLI trees to `schema` / `fields` / `values`
- fixed UID-vs-sequence targeting in `pkg/dsl/actions.go`
- fixed mail header address serialization in `cmd/mailgen/cmds/generate.go`
- repaired two additional runtime defects discovered during Docker validation:
  - `mail-rules` was not executing configured actions
  - fetch logic treated missing body structures as fatal for non-MIME rules
- updated the root and command docs to match the current binaries
- added focused regression tests and a reproducible Docker validation script

Validation summary:

- `go test ./... -count=1`
- `go build ./cmd/smailnail ./cmd/mailgen ./cmd/imap-tests`
- help smoke for all three binaries
- end-to-end Docker IMAP validation script covering mailbox creation, fetch, rule actions, and `mailgen --store-imap`

## Topics

- smailnail
- glazed
- refactor
- go
- email

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts

## Commit Summary

- `dbc9e00` `Port smailnail commands to glazed facade APIs`
- `cd446d2` `Fix smailnail IMAP action and header handling`
