---
Title: Implementation diary
Ticket: SMAILNAIL-014-SHARED-OIDC-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - mcp
    - oidc
    - authentication
    - keycloak
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/user.go
      Note: Starting point for the current dev-only identity behavior that motivated this ticket
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/embeddable/auth_provider_external.go
      Note: Existing verified-token path reviewed while scoping the shared identity work
ExternalSources:
    - https://openid.net/specs/openid-connect-core-1_0.html
    - https://openid.net/specs/openid-connect-discovery-1_0.html
    - https://datatracker.ietf.org/doc/html/rfc7591
    - https://www.keycloak.org/docs/latest/server_admin/
Summary: Diary capturing the creation of the shared OIDC identity implementation ticket and the reasoning that connects the current browser stub to the existing MCP OIDC flow.
LastUpdated: 2026-03-16T11:28:46.488167485-04:00
WhatFor: Record the concrete reasoning, references, and scoping decisions behind the shared identity implementation ticket.
WhenToUse: Use when reviewing why the ticket exists, what assumptions it makes, and which code paths were inspected at kickoff.
---

# Implementation diary

## Goal

Capture the starting context for the shared OIDC identity implementation ticket so future implementation work can see what was already established.

## Context

Existing state at ticket creation:

- `smailnaild` still resolves users through `HeaderUserResolver` and falls back to `local-user`
- `smailnail-mcp` already validates OIDC access tokens through `go-go-mcp`
- account storage and rule storage now exist in the hosted SQL backend
- the product needs the same human to be recognized identically in the browser and in MCP calls

## Quick Reference

### Notes from kickoff

- Browser auth and MCP auth should share identity semantics, not necessarily the same transport mechanism.
- The stable local key should be `(issuer, subject)`.
- The web app should use server-side OIDC code exchange and session cookies.
- The MCP should keep bearer-token validation and then call the same local-user mapping layer.
- IMAP accounts should stay in the application database and belong to the resolved local user.

### References consulted

- OIDC Core for claim semantics, especially `iss`, `sub`, and audience rules
- OIDC Discovery for issuer metadata conventions
- RFC 7591 because hosted MCP connectors may rely on dynamic client registration
- Keycloak server admin docs for client scopes, mappers, and client setup

## Usage Examples

Use this diary when:

- starting implementation on the backend auth layer
- onboarding a contributor to the shared identity problem
- justifying why email or `client_id` should not become the primary local user key

## Related

- Main plan: [../design-doc/01-implementation-plan-for-shared-oidc-identity-across-smailnaild-and-smailnail-mcp.md](../design-doc/01-implementation-plan-for-shared-oidc-identity-across-smailnaild-and-smailnail-mcp.md)
- Background guide: [../../SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE--explain-oidc-identity-user-mapping-and-imap-credential-storage-design-for-smailnail/index.md](../../SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE--explain-oidc-identity-user-mapping-and-imap-credential-storage-design-for-smailnail/index.md)
- Execution dependency: [../../SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION--implement-hosted-smailnail-account-setup-phases-1-and-2/index.md](../../SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION--implement-hosted-smailnail-account-setup-phases-1-and-2/index.md)

## History Analysis Addendum

To support wider repo-history searches, a reusable exporter was added under:

- [../scripts/export_git_history_to_sqlite.py](../scripts/export_git_history_to_sqlite.py)

That script exports full reachable git history into SQLite so future questions about timeline, path introduction, and file-level evolution can be answered with SQL instead of repeated ad hoc `git log` commands.

It was exercised against the `smailnail` repo and confirmed:

- the repo is not shallow
- the oldest reachable commit is `27c2460` (`Initial commit`)
- the first reachable MCP files appear in `ff584b4` (`Add smailnail IMAP JS MCP runtime slice`)
- there is no earlier non-JS MCP surface in the reachable history before that commit
