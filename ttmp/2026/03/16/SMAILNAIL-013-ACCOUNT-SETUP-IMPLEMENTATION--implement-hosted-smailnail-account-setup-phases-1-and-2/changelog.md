# Changelog

## 2026-03-16

- Initial workspace created
- Added a detailed implementation plan for Phase 1 and Phase 2 of the hosted smailnail UI
- Added a granular milestone and task breakdown suitable for commit-by-commit execution
- Split the task list into backend work we will implement and frontend work owned by the UI frontend developer
- Completed the backend foundation slice: versioned schema bootstrap, Phase 1/2 tables, package layout, secret encryption helpers, and focused `smailnaild` tests
- Added a living API reference document for the UI handoff, marking which endpoints are already implemented and which are still draft contracts
- Implemented the backend account and rule slices: repositories, ownership-aware services, hosted JSON APIs, and local-user request resolution
- Added read-only IMAP account tests, mailbox listing, preview-message fetches, rule CRUD, dry-run persistence, and a DSL field-normalization fix needed for saved-rule execution
- Added Dovecot-backed integration coverage for account services, rule dry-runs, and the hosted HTTP flow
- Updated `smailnaild` README/help/docs with encryption-key setup, local account-flow walkthroughs, and exact verification commands
