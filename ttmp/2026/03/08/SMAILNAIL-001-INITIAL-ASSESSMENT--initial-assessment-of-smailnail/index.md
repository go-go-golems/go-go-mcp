---
Title: Initial assessment of smailnail
Ticket: SMAILNAIL-001-INITIAL-ASSESSMENT
Status: active
Topics:
    - smailnail
    - go
    - email
    - review
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: smailnail
      Note: Repository under assessment for this ticket
ExternalSources: []
Summary: Ticket index for the initial assessment of smailnail, linking the main analysis, diary, experiments, and current status.
LastUpdated: 2026-03-08T19:51:28.264205386-04:00
WhatFor: Track the overall scope, outputs, and status of the smailnail initial assessment.
WhenToUse: Use as the landing page for the ticket before opening the detailed design doc or diary.
---


# Initial assessment of smailnail

## Overview

This ticket captures an evidence-based initial assessment of `smailnail/`. The deliverable set includes a detailed code review, an intern-oriented architecture/user guide, a chronological investigation diary, and two ticket-local experiments used to validate parser/example and mail-header assumptions.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

- Design doc: `design-doc/01-smailnail-initial-assessment-and-intern-guide.md`
- Diary: `reference/01-investigation-diary.md`
- Scripts: `scripts/parse-rule-examples.sh`, `scripts/address-header-experiment.sh`

## Status

Current status: **active**

## Topics

- smailnail
- go
- email
- review

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Assessment Summary

Current conclusion:

- the repo has a useful core in `pkg/dsl` and `pkg/mailgen`
- the user-facing CLIs are currently broken by legacy Glazed imports
- docs and examples are partially stale
- there are at least two correctness risks worth addressing after build restoration

Highest-priority findings:

1. Build breakage caused by deleted Glazed packages in command wiring and shared IMAP layer code.
2. IMAP mutation actions appear to send UID values through sequence-number sets.
3. `mailgen` appears to serialize malformed display-name headers when storing to IMAP.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
