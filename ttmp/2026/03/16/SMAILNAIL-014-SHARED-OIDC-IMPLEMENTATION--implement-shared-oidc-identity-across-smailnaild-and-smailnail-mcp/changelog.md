# Changelog

## 2026-03-16

- Initial workspace created


## 2026-03-16

Create shared OIDC identity implementation ticket with detailed design, implementation phases, and granular task breakdown.


## 2026-03-16

Add ticket-local git-history-to-sqlite exporter and record the deeper MCP history result.


## 2026-03-16

Implement identity schema version 6 and add the shared user resolution foundation in smailnail.


## 2026-03-16

Add hosted auth-mode configuration, session-backed user resolution, /api/me, and 401 responses for protected API calls.


## 2026-03-16

Add hosted OIDC login, callback, logout, and fake-provider session flow tests for `smailnaild`.


## 2026-03-16

Carry richer verified OIDC principals through `go-go-mcp` and resolve the same local user for MCP bearer-authenticated requests in `smailnail`.


## 2026-03-16

Allow MCP JavaScript execution to select hosted stored IMAP accounts by `accountId` with ownership enforcement and browser-to-MCP account reuse tests.


## 2026-03-16

Add a live local Keycloak plus Dovecot end-to-end regression test for shared OIDC identity and stored-account MCP execution.


## 2026-03-16

Document `smailnail-web` Keycloak setup, shared OIDC playbooks, and the MCP container env needed to resolve browser-created stored accounts in hosted deployments.


## 2026-03-16

Add a frontend auth bootstrap shell around `/api/me`, including logged-out handling, login CTA, authenticated user display, and logout entry points.
