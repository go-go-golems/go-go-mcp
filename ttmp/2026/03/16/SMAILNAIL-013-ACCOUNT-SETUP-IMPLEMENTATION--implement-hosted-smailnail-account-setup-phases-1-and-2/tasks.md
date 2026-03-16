# Tasks

## Ready

- [ ] Add schema/bootstrap support for `imap_accounts`, `imap_account_tests`, `rules`, and `rule_runs`
- [ ] Add encrypted secret storage helpers and tests
- [ ] Add account repository and service packages
- [ ] Add account CRUD APIs
- [ ] Add account test API with read-only validation
- [ ] Add mailbox list and message preview APIs
- [ ] Add frontend shell and account management screens
- [ ] Add rule repository and service packages
- [ ] Add rule CRUD APIs
- [ ] Add rule dry-run API and persistence
- [ ] Add rule library, rule builder, and dry-run screens
- [ ] Add Dovecot-backed integration coverage and an end-to-end hosted smoke

## Granular breakdown

- [ ] Milestone A1: create package layout under `pkg/smailnaild/accounts`, `rules`, and `secrets`
- [ ] Milestone A2: refactor DB bootstrap to support versioned schema creation
- [ ] Milestone A3: add `imap_accounts` schema
- [ ] Milestone A4: add `imap_account_tests` schema
- [ ] Milestone A5: add `rules` schema
- [ ] Milestone A6: add `rule_runs` schema
- [ ] Milestone B1: add encryption key config loading
- [ ] Milestone B2: add secret encrypt/decrypt helpers
- [ ] Milestone B3: add secret helper tests
- [ ] Milestone C1: add account repository create/get/list/update/delete
- [ ] Milestone C2: add account service ownership checks
- [ ] Milestone C3: add account service tests
- [ ] Milestone D1: add read-only IMAP account test runner
- [ ] Milestone D2: classify account test results into UI-safe warnings/errors
- [ ] Milestone D3: add mailbox list helper
- [ ] Milestone D4: add recent-message preview helper
- [ ] Milestone D5: add message-detail preview helper
- [ ] Milestone D6: add Dovecot-backed integration tests for account testing
- [ ] Milestone E1: split `smailnaild` health routes from app API routes
- [ ] Milestone E2: add account CRUD handlers
- [ ] Milestone E3: add account test handler
- [ ] Milestone E4: add mailbox list handler
- [ ] Milestone E5: add message preview handlers
- [ ] Milestone E6: add HTTP handler tests
- [ ] Milestone F1: scaffold hosted frontend build and embed path
- [ ] Milestone F2: add app shell and navigation
- [ ] Milestone F3: add accounts list screen
- [ ] Milestone F4: add add/edit account form
- [ ] Milestone F5: add account test results screen
- [ ] Milestone F6: add mailbox explorer and message preview UI
- [ ] Milestone G1: add rule repository create/get/list/update/delete
- [ ] Milestone G2: add rule YAML validation using current DSL
- [ ] Milestone G3: add rule service tests
- [ ] Milestone H1: add dry-run service path using stored accounts
- [ ] Milestone H2: persist dry-run runs
- [ ] Milestone H3: add Dovecot-backed integration tests for dry-run
- [ ] Milestone I1: add rule library screen
- [ ] Milestone I2: add rule builder form
- [ ] Milestone I3: add YAML preview panel
- [ ] Milestone I4: add dry-run results screen
- [ ] Milestone J1: add end-to-end hosted smoke
- [ ] Milestone J2: update README and hosted testing docs
