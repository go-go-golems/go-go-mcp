# Tasks

## Backend track

### Ready

- [x] Add schema/bootstrap support for `imap_accounts`, `imap_account_tests`, `rules`, and `rule_runs`
- [x] Add encrypted secret storage helpers and tests
- [ ] Add account repository and service packages
- [ ] Add account CRUD APIs
- [ ] Add account test API with read-only validation
- [ ] Add mailbox list and message preview APIs
- [ ] Add rule repository and service packages
- [ ] Add rule CRUD APIs
- [ ] Add rule dry-run API and persistence
- [ ] Add Dovecot-backed integration coverage and an end-to-end hosted smoke

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
- [ ] Backend C1: add account repository create/get/list/update/delete
- [ ] Backend C2: add account service ownership checks
- [ ] Backend C3: add account service tests
- [ ] Backend D1: add read-only IMAP account test runner
- [ ] Backend D2: classify account test results into UI-safe warnings/errors
- [ ] Backend D3: add mailbox list helper
- [ ] Backend D4: add recent-message preview helper
- [ ] Backend D5: add message-detail preview helper
- [ ] Backend D6: add Dovecot-backed integration tests for account testing
- [ ] Backend E1: split `smailnaild` health routes from app API routes
- [ ] Backend E2: add account CRUD handlers
- [ ] Backend E3: add account test handler
- [ ] Backend E4: add mailbox list handler
- [ ] Backend E5: add message preview handlers
- [ ] Backend E6: add HTTP handler tests
- [ ] Backend G1: add rule repository create/get/list/update/delete
- [ ] Backend G2: add rule YAML validation using current DSL
- [ ] Backend G3: add rule service tests
- [ ] Backend H1: add dry-run service path using stored accounts
- [ ] Backend H2: persist dry-run runs
- [ ] Backend H3: add Dovecot-backed integration tests for dry-run
- [ ] Backend J1: add end-to-end hosted smoke
- [ ] Backend J2: update README and hosted testing docs

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
