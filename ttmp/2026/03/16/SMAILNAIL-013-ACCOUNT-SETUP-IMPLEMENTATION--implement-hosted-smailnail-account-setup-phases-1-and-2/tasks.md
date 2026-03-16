# Tasks

## Backend track

### Ready

- [x] Add schema/bootstrap support for `imap_accounts`, `imap_account_tests`, `rules`, and `rule_runs`
- [x] Add encrypted secret storage helpers and tests
- [x] Add account repository and service packages
- [x] Add account CRUD APIs
- [x] Add account test API with read-only validation
- [x] Add mailbox list and message preview APIs
- [x] Add rule repository and service packages
- [x] Add rule CRUD APIs
- [x] Add rule dry-run API and persistence
- [x] Add Dovecot-backed integration coverage and an end-to-end hosted smoke

### Granular breakdown

- [x] Backend A1: create package layout under `pkg/smailnaild/accounts`, `rules`, and `secrets`
- [x] Backend A2: refactor DB bootstrap to support versioned schema creation
- [x] Backend A3: add `imap_accounts` schema
- [x] Backend A4: add `imap_account_tests` schema
- [x] Backend A5: add `rules` schema
- [x] Backend A6: add `rule_runs` schema
- [x] Backend B1: add encryption key config loading
- [x] Backend B2: add secret encrypt/decrypt helpers
- [x] Backend B3: add secret helper tests
- [x] Backend C1: add account repository create/get/list/update/delete
- [x] Backend C2: add account service ownership checks
- [x] Backend C3: add account service tests
- [x] Backend D1: add read-only IMAP account test runner
- [x] Backend D2: classify account test results into UI-safe warnings/errors
- [x] Backend D3: add mailbox list helper
- [x] Backend D4: add recent-message preview helper
- [x] Backend D5: add message-detail preview helper
- [x] Backend D6: add Dovecot-backed integration tests for account testing
- [x] Backend E1: split `smailnaild` health routes from app API routes
- [x] Backend E2: add account CRUD handlers
- [x] Backend E3: add account test handler
- [x] Backend E4: add mailbox list handler
- [x] Backend E5: add message preview handlers
- [x] Backend E6: add HTTP handler tests
- [x] Backend G1: add rule repository create/get/list/update/delete
- [x] Backend G2: add rule YAML validation using current DSL
- [x] Backend G3: add rule service tests
- [x] Backend H1: add dry-run service path using stored accounts
- [x] Backend H2: persist dry-run runs
- [x] Backend H3: add Dovecot-backed integration tests for dry-run
- [x] Backend J1: add end-to-end hosted smoke
- [x] Backend J2: update README and hosted testing docs

## Frontend track

### For UI frontend developer

- [ ] Add hosted frontend shell and account management screens
- [ ] Add rule library, rule builder, and dry-run screens

### Granular breakdown

- [ ] Frontend F1: scaffold hosted frontend build and embed path
- [ ] Frontend F2: add app shell and navigation
- [ ] Frontend F3: add accounts list screen
- [ ] Frontend F4: add add/edit account form
- [ ] Frontend F5: add account test results screen
- [ ] Frontend F6: add mailbox explorer and message preview UI
- [ ] Frontend I1: add rule library screen
- [ ] Frontend I2: add rule builder form
- [ ] Frontend I3: add YAML preview panel
- [ ] Frontend I4: add dry-run results screen
