---
Title: Investigation diary
Ticket: SMAILNAIL-012-WEB-UI-UX-RESEARCH
Status: complete
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
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go
      Note: Reviewed to confirm the hosted UI does not yet exist and only minimal endpoints are present
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go
      Note: Reviewed to ground account onboarding and test flows in the current IMAP connection primitive
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go
      Note: Reviewed to map current search and action primitives to UI rule-builder requirements
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go
      Note: Reviewed to understand how mailbox preview and rule dry-run can reuse existing fetch and pagination logic
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/actions.go
      Note: Reviewed to derive safe confirmation flows for move, copy, delete, and export actions
ExternalSources:
    - https://support.google.com/accounts/answer/185833
    - https://support.google.com/mail/answer/7126229
    - https://learn.microsoft.com/en-us/exchange/clients-and-mobile-in-exchange-online/deprecation-of-basic-authentication-exchange-online
    - https://support.apple.com/en-us/102578
    - https://support.apple.com/en-il/guide/mail/mlhlp1040/mac
    - https://support.apple.com/en-il/guide/mail/mlhlp1190/mac
Summary: Chronological notes for researching a hosted smailnail web UI from user needs through screen design, storage, and implementation planning.
LastUpdated: 2026-03-16T09:48:00-04:00
WhatFor: Record the reasoning and evidence behind the proposed web UI, account-management, inbox-preview, and rule-management design.
WhenToUse: Use when extending the design work, checking what current code and external constraints informed the proposal, or planning implementation phases from this research.
---

# Investigation diary

## Goal

Produce a detailed intern-facing research and design guide for a hosted `smailnail` web UI that helps users manage IMAP credentials for multiple servers, test those connections safely, preview inboxes, build and test filters, and understand how the UI and hosted MCP should converge on shared account state.

## 2026-03-16

### Step 1: created a dedicated ticket

I created `SMAILNAIL-012-WEB-UI-UX-RESEARCH` as a separate workspace because this work is product and UX research, not just another credential-storage note. It needed its own design narrative, screen concepts, and implementation sequence.

### Step 2: inspected the current hosted process boundary

I reviewed [http.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go) and [serve.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnaild/commands/serve.go).

Key finding:

- `smailnaild` currently exposes only `healthz`, `readyz`, and `api/info`
- there is no UI, no session flow, no account CRUD, and no mailbox preview surface yet

That set the baseline: this ticket is designing the first real product UX, not refining an existing one.

### Step 3: inspected the reusable mail and rule primitives

I reviewed:

- [layer.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go)
- [types.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go)
- [processor.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go)
- [actions.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/actions.go)

Key findings:

- IMAP connection logic already exists and is sufficient for read-only and optional write-probe tests
- the rule DSL already covers search criteria, pagination, MIME filtering, and actions
- the processor already supports efficient fetch and paging, which fits mailbox preview and dry-run UX well
- destructive actions exist and therefore need careful confirmation in the UI

### Step 4: inspected the current MCP surface

I reviewed [server.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go), [execute_tool.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go), and [module.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go).

Key finding:

- the hosted MCP and the future web UI should share the same stored account model
- the UI should not be a separate product; it should become the configuration and visibility layer for the hosted MCP runtime

### Step 5: reviewed prior hosted and identity design work

I revisited:

- `SMAILNAIL-003` for the earlier hosted architecture direction
- `SMAILNAIL-011` for the provider-neutral identity and app-side credential storage design

Key finding:

- the web UI design must assume app-side storage and `(issuer, subject)` user mapping rather than Keycloak-owned application data

### Step 6: gathered external provider constraints

I reviewed official docs from:

- Google
- Microsoft
- Apple
- the IMAP RFC

The important product implications were:

- Gmail setup often requires IMAP enablement plus app passwords
- Microsoft-hosted mailboxes may fail under simple username/password due to basic-auth deprecation
- users expect multi-account mailbox organization and smart saved views because mainstream mail clients teach those mental models

### Step 7: design conclusion

The product should not start as "settings for credentials." It should start as a hosted mail operations console with four core flows:

1. add and validate accounts
2. browse mailboxes and sample messages
3. create and dry-run rules
4. bind stored accounts and rules to hosted MCP use

That led directly to the screen concepts and API design in the main design document.

### Step 8: validation and publication

I ran:

- `docmgr doctor --ticket SMAILNAIL-012-WEB-UI-UX-RESEARCH --stale-after 30`
- `remarquee upload bundle ... --name "SMAILNAIL-012 Web UI UX and MCP Mail Management Guide" --remote-dir "/ai/2026/03/16/SMAILNAIL-012-WEB-UI-UX-RESEARCH"`

The ticket validated cleanly, and the PDF bundle uploaded successfully to reMarkable. I then verified both the remote directory and the uploaded document name.

## Quick reference

### Core UX decisions

- multi-account first
- read-only test by default
- write test opt-in
- dry-run before destructive actions
- account scope visible on every rule and mailbox view

### Core technical decisions

- hosted UI and MCP should share account state
- encrypted account secrets live in app DB
- rule builder emits current DSL YAML
- mailbox preview reuses the existing DSL processor

## Related

- [01-intern-guide-to-smailnail-web-ui-ux-architecture-and-implementation-for-hosted-mcp-mail-management.md](../design-doc/01-intern-guide-to-smailnail-web-ui-ux-architecture-and-implementation-for-hosted-mcp-mail-management.md)
- [01-intern-guide-to-oidc-identity-user-mapping-and-imap-credential-storage-in-smailnail.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-011-OIDC-IDENTITY-CREDENTIALS-GUIDE--explain-oidc-identity-user-mapping-and-imap-credential-storage-design-for-smailnail/design-doc/01-intern-guide-to-oidc-identity-user-mapping-and-imap-credential-storage-in-smailnail.md)
