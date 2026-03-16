---
Title: Implementation diary
Ticket: SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - mcp
    - documentation
    - architecture
    - authentication
    - sql
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-012-WEB-UI-UX-RESEARCH--research-web-ui-for-smailnail-mcp-account-inbox-and-filter-management/design-doc/01-intern-guide-to-smailnail-web-ui-ux-architecture-and-implementation-for-hosted-mcp-mail-management.md
      Note: Source research ticket from which the implementation phases were derived
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go
      Note: Reviewed to capture the current hosted baseline before expanding to account and rule APIs
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go
      Note: Reviewed to identify the current schema bootstrap gap
ExternalSources: []
Summary: Chronological notes for turning the hosted smailnail UI research into a concrete Phase 1 and Phase 2 implementation ticket.
LastUpdated: 2026-03-16T10:02:00-04:00
WhatFor: Preserve the reasoning behind the task ordering and the chosen implementation slices.
WhenToUse: Use when starting implementation or revisiting why the work was broken down in this order.
---

# Implementation diary

## Goal

Create a concrete execution ticket for the first two hosted `smailnail` UI phases so the team can start building account setup, mailbox previews, rule CRUD, and rule dry-runs without having to reinterpret the earlier research document.

## 2026-03-16

### Step 1: created the execution ticket

I created `SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION` as a follow-on to the UX research ticket. The intent is to separate discovery from implementation planning.

### Step 2: mapped the research phases to code

I revisited the `SMAILNAIL-012` guide and extracted:

- Phase 1: hosted backend primitives
- Phase 2: rule CRUD and dry-run

I also checked the current code in:

- [http.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go)
- [db.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go)

That made the initial implementation gap explicit:

- the hosted server exists
- the app DB exists
- but there are no domain packages, no CRUD APIs, no rule endpoints, and no UI

### Step 3: chose vertical slices instead of layer-first tasks

I structured the plan around milestones that can be committed independently:

- schema
- secrets
- accounts
- testing and preview
- APIs
- frontend shell
- rules
- dry-runs

That ordering should make it possible to keep the app runnable and reviewable throughout the work.

### Step 4: implemented the backend foundation slice

I started with the work that every later backend task depends on:

- new package layout under `pkg/smailnaild/accounts`, `pkg/smailnaild/rules`, and `pkg/smailnaild/secrets`
- versioned schema bootstrap in [db.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go)
- new schema for `imap_accounts`, `imap_account_tests`, `rules`, and `rule_runs`
- environment-backed encryption config and AES-GCM helpers in `pkg/smailnaild/secrets`

I kept this slice intentionally narrow: no repositories or HTTP handlers yet. The goal was to land stable foundations first.

### Step 5: verified the slice with focused tests

I ran:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/smailnaild/...
```

That covered:

- schema bootstrap and migration behavior
- fresh DB creation
- legacy version-1 DB upgrade
- secret config loading
- encrypt/decrypt round trips
- corrupt ciphertext handling

### Step 6: started the API handoff document

I added a separate API reference document to the ticket so the UI designer can track:

- which endpoints already exist
- which endpoints are still draft contracts
- what payload shapes are being proposed before implementation

This should evolve alongside each backend slice rather than being written only at the end.

### Step 7: implemented the account backend slice

I added the hosted account backend in `smailnail`:

- repository methods in `pkg/smailnaild/accounts/repository.go`
- ownership-aware service behavior in `pkg/smailnaild/accounts/service.go`
- local-development user resolution in `pkg/smailnaild/user.go`
- hosted account HTTP handlers in `pkg/smailnaild/http.go`

The service layer now supports:

- create, list, get, update, delete
- read-only IMAP account tests
- mailbox listing
- preview rows
- single-message detail previews

Important design choice:

- until hosted app auth is implemented, requests resolve to a local default user ID
- callers can override that with `X-Smailnail-User-ID`
- ownership is still enforced through the stored `user_id` column

### Step 8: implemented the rule backend slice

I added the hosted rule backend in:

- `pkg/smailnaild/rules/repository.go`
- `pkg/smailnaild/rules/service.go`

The rule path now supports:

- list, create, get, update, delete
- YAML validation using the current DSL parser
- dry-run execution against stored IMAP accounts
- persisted `rule_runs`
- `last_preview_count` and `last_run_at` updates on dry-run

### Step 9: found and fixed a saved-rule execution bug in the DSL layer

The first Dovecot-backed dry-run failed even though the saved YAML looked valid. The failure came from `FETCH` being sent with an empty field list.

The root cause was subtle:

- rules were normalized and stored in canonical YAML with `fields` entries like `name: uid`
- `dsl.OutputConfig.UnmarshalYAML` only converted string-based fields and a couple of older complex field shapes
- canonical `name:` field objects were left as generic maps
- later `BuildFetchOptions` skipped them because they were not typed `dsl.Field` values

I fixed that in `pkg/dsl/types.go` by teaching `OutputConfig.UnmarshalYAML` to rehydrate canonical `name` and `content` objects into typed `dsl.Field` values.

This was not just a hosted-backend problem. It was a general saved-rule correctness problem, so fixing it in the DSL package was the right boundary.

### Step 10: added verification coverage at three levels

I added:

- unit tests for account service behavior in `pkg/smailnaild/accounts/service_test.go`
- unit tests for rule service behavior in `pkg/smailnaild/rules/service_test.go`
- handler tests in `pkg/smailnaild/http_test.go`
- Dovecot-backed integration tests in:
  - `pkg/smailnaild/accounts/integration_test.go`
  - `pkg/smailnaild/rules/integration_test.go`
  - `pkg/smailnaild/http_integration_test.go`

### Step 11: restarted the local Dovecot fixture and ran the real backend flow

I started the fixture with:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker compose -f docker-compose.local.yml up -d dovecot
```

Then I verified the IMAPS port was live:

```bash
nc -z 127.0.0.1 993
```

After that I ran the focused Dovecot-backed tests:

```bash
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild/accounts -run TestServiceAgainstLocalDovecot -v
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild/rules -run TestDryRunAgainstLocalDovecot -v
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild -run TestHostedHTTPFlowAgainstLocalDovecot -v
```

All three passed after the DSL field-rehydration fix.

### Step 12: updated the docs for backend handoff and local recreation

I updated:

- `README.md`
- `cmd/smailnaild/commands/serve.go` help text
- `docs/smailnaild-local-account-flow.md`

Those now document:

- `SMAILNAILD_ENCRYPTION_KEY`
- the hosted account and rule API endpoints
- local Dovecot-backed testing
- curl examples for account creation, account testing, previews, and dry-runs

### Step 13: final verification sweep

I ran:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
SMAILNAILD_DOVECOT_TEST=1 go test ./...
go run ./cmd/smailnaild serve --help
```

The full repo test sweep passed with the Dovecot-backed hosted tests enabled, and the `serve` help output reflected the new backend capabilities and encryption-key requirement.

## Quick reference

### First delivery target

- account CRUD
- account test
- mailbox list
- message preview

### Second delivery target

- saved rules
- rule validation
- dry-run previews

### Key dependency

- app-side encrypted secret storage from the earlier identity design ticket

## Related

- [01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md](../design-doc/01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md)
