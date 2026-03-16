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
- Added a focused UX handover document for the frontend developer covering only the add-account and lightweight connection-test slice
- Moved hosted encryption-key config from direct `os.Getenv` reads to an explicit Glazed `Encryption Settings` section on `smailnaild serve`

## 2026-03-16

Step 15: React/Vite SPA frontend with account setup UI (commit cc59315)

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Makefile — Added dev-backend/dev-frontend/build-embed targets
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go — Wired SPA handler into NewHandler
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/web/spa.go — SPA handler with API guard
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui/src/features/accounts/accountsSlice.ts — Redux state machine and async thunks

## 2026-03-16

Step 16: Mailbox explorer with sidebar, message list, detail, and account delete (commit e6ee50b)

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui/src/features/mailbox/MailboxExplorer.tsx — Mailbox explorer page component


## 2026-03-16

Step 17: Rules CRUD and dry-run UI covering full backend API surface (commit e58bfe4)

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui/src/features/rules/rulesSlice.ts — Redux rules slice with CRUD and dry-run thunks

