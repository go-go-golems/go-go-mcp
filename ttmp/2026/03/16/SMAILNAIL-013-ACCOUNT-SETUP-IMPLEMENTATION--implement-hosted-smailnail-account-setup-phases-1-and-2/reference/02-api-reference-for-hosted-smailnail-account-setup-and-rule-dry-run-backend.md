---
Title: API reference for hosted smailnail account setup and rule dry-run backend
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
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go
      Note: Current implemented HTTP endpoints and the main entrypoint for future API expansion
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go
      Note: Current schema bootstrap that determines the stored backend resources this API will expose
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/types.go
      Note: Current account and account-test domain record shapes
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/rules/types.go
      Note: Current rule and rule-run domain record shapes
ExternalSources: []
Summary: Living API reference for the hosted smailnail backend, tracking implemented endpoints and draft contracts for account setup, mailbox previews, and rule dry-runs.
LastUpdated: 2026-03-16T10:18:00-04:00
WhatFor: Give the backend and UI developers one evolving source of truth for endpoint status, payload shapes, and contract decisions.
WhenToUse: Use when implementing backend handlers, building the frontend against draft or implemented APIs, or handing off backend status to the UI developer.
---

# API reference for hosted smailnail account setup and rule dry-run backend

## Goal

Track the hosted `smailnaild` API surface as it evolves through Phase 1 and Phase 2. This document distinguishes between:

- `Implemented`: endpoints that exist in the code today
- `Draft`: contracts that are planned and should guide the next backend and frontend slices, but are not yet implemented

## Context

This document is intentionally ahead of the code in a few places. The backend is being built in slices, but the frontend developer still needs early visibility into the contract direction.

As of the current slice:

- database schema and secret handling foundations are implemented
- hosted account/rule APIs are not yet implemented
- only health/info endpoints exist in `smailnaild`

## Quick reference

## Endpoint status table

| Endpoint | Method | Status | Purpose |
| --- | --- | --- | --- |
| `/healthz` | `GET` | Implemented | Liveness check |
| `/readyz` | `GET` | Implemented | Readiness check including DB ping |
| `/api/info` | `GET` | Implemented | Service metadata |
| `/api/accounts` | `GET` | Draft | List current user's saved IMAP accounts |
| `/api/accounts` | `POST` | Draft | Create a saved IMAP account |
| `/api/accounts/:id` | `GET` | Draft | Fetch one saved account |
| `/api/accounts/:id` | `PATCH` | Draft | Update a saved account |
| `/api/accounts/:id` | `DELETE` | Draft | Delete or archive a saved account |
| `/api/accounts/:id/test` | `POST` | Draft | Run a read-only or write-probe account test |
| `/api/accounts/:id/mailboxes` | `GET` | Draft | List mailboxes for one saved account |
| `/api/accounts/:id/messages` | `GET` | Draft | Fetch paginated preview messages |
| `/api/accounts/:id/messages/:uid` | `GET` | Draft | Fetch one preview message in more detail |
| `/api/rules` | `GET` | Draft | List saved rules |
| `/api/rules` | `POST` | Draft | Create a saved rule |
| `/api/rules/:id` | `GET` | Draft | Fetch one saved rule |
| `/api/rules/:id` | `PATCH` | Draft | Update one saved rule |
| `/api/rules/:id` | `DELETE` | Draft | Delete one saved rule |
| `/api/rules/:id/dry-run` | `POST` | Draft | Run a dry-run preview for a saved rule |

## Implemented endpoints

### `GET /healthz`

Purpose:

- liveness probe

Response:

```json
{
  "status": "ok"
}
```

### `GET /readyz`

Purpose:

- readiness probe
- verifies DB reachability

Success response:

```json
{
  "status": "ready"
}
```

Failure response:

```json
{
  "status": "not-ready",
  "error": "database is nil"
}
```

### `GET /api/info`

Purpose:

- expose service identity and DB target metadata

Response shape:

```json
{
  "service": "smailnaild",
  "version": "dev",
  "startedAt": "2026-03-15T18:00:00Z",
  "database": {
    "driver": "sqlite3",
    "target": ":memory:",
    "mode": "structured"
  }
}
```

## Draft conventions for Phase 1 and Phase 2

These are the contract conventions the next backend slices should follow unless there is a strong reason to change them.

### Authentication assumption

Draft assumption:

- all future `/api/accounts*` and `/api/rules*` endpoints require an authenticated application user
- the user identity should eventually come from the OIDC-backed app session or bearer-auth context

### Response format

For application APIs, prefer JSON objects over raw arrays so metadata can be added without breaking clients.

Preferred pattern:

```json
{
  "data": {},
  "meta": {}
}
```

For list endpoints:

```json
{
  "data": [],
  "meta": {
    "count": 0
  }
}
```

### Error format

Preferred draft error envelope:

```json
{
  "error": {
    "code": "account-test-login-failed",
    "message": "Login failed for the provided IMAP credentials.",
    "details": {
      "providerHint": "gmail"
    }
  }
}
```

Use:

- stable `code` for frontend logic
- human-readable `message`
- optional `details` for display or debugging

## Draft account API

### `GET /api/accounts`

Status:

- Draft

Purpose:

- return the current user's saved accounts with enough status data to render the accounts table

Draft response:

```json
{
  "data": [
    {
      "id": "acc_work_gmail",
      "label": "Work Gmail",
      "providerHint": "gmail",
      "server": "imap.gmail.com",
      "port": 993,
      "username": "user@example.com",
      "mailboxDefault": "INBOX",
      "insecure": false,
      "authKind": "password",
      "isDefault": true,
      "mcpEnabled": false,
      "latestTest": {
        "success": true,
        "warningCode": "provider-app-password-recommended",
        "createdAt": "2026-03-16T10:00:00Z"
      }
    }
  ],
  "meta": {
    "count": 1
  }
}
```

### `POST /api/accounts`

Status:

- Draft

Purpose:

- save a new account with encrypted credentials

Draft request:

```json
{
  "label": "Work Gmail",
  "providerHint": "gmail",
  "server": "imap.gmail.com",
  "port": 993,
  "username": "user@example.com",
  "password": "app-password-here",
  "mailboxDefault": "INBOX",
  "insecure": false,
  "authKind": "password"
}
```

Draft response:

```json
{
  "data": {
    "id": "acc_work_gmail",
    "label": "Work Gmail",
    "providerHint": "gmail",
    "server": "imap.gmail.com",
    "port": 993,
    "username": "user@example.com",
    "mailboxDefault": "INBOX",
    "insecure": false,
    "authKind": "password",
    "isDefault": false,
    "mcpEnabled": false,
    "createdAt": "2026-03-16T10:05:00Z",
    "updatedAt": "2026-03-16T10:05:00Z"
  }
}
```

Note:

- plaintext `password` is request-only and must never appear in responses

### `POST /api/accounts/:id/test`

Status:

- Draft

Purpose:

- run a structured test against one saved account

Draft request:

```json
{
  "mode": "read_only"
}
```

Allowed draft modes:

- `read_only`
- `write_probe`

Draft response:

```json
{
  "data": {
    "id": "acctest_01",
    "imapAccountId": "acc_work_gmail",
    "testMode": "read_only",
    "success": true,
    "tcpOk": true,
    "loginOk": true,
    "mailboxSelectOk": true,
    "listOk": true,
    "sampleFetchOk": true,
    "writeProbeOk": null,
    "warningCode": "provider-app-password-recommended",
    "errorCode": "",
    "errorMessage": "",
    "detailsJson": "{\"sampleMailboxes\":[\"INBOX\",\"Archive\"]}",
    "createdAt": "2026-03-16T10:07:00Z"
  }
}
```

### `GET /api/accounts/:id/mailboxes`

Status:

- Draft

Purpose:

- list the account’s mailboxes for the explorer UI

Draft response:

```json
{
  "data": [
    { "name": "INBOX", "path": "INBOX", "children": [] },
    { "name": "Archive", "path": "Archive", "children": [] }
  ]
}
```

### `GET /api/accounts/:id/messages`

Status:

- Draft

Purpose:

- fetch preview rows for one mailbox

Draft query parameters:

- `mailbox`
- `limit`
- `offset`
- `query`
- `unread_only`

Draft response:

```json
{
  "data": [
    {
      "uid": 91230,
      "date": "2026-03-15T12:11:00Z",
      "from": "Billing Co <billing@example.com>",
      "subject": "Invoice 9381",
      "flags": ["Seen"]
    }
  ],
  "meta": {
    "mailbox": "INBOX",
    "count": 1,
    "limit": 20,
    "offset": 0
  }
}
```

## Draft rule API

### `GET /api/rules`

Status:

- Draft

Purpose:

- list saved rules for the current user

Draft response:

```json
{
  "data": [
    {
      "id": "rule_invoice_triage",
      "imapAccountId": "acc_work_gmail",
      "name": "Invoice triage",
      "description": "Find invoice emails and move them to Receipts",
      "status": "draft",
      "sourceKind": "ui",
      "lastPreviewCount": 12,
      "lastRunAt": null,
      "createdAt": "2026-03-16T10:12:00Z",
      "updatedAt": "2026-03-16T10:12:00Z"
    }
  ],
  "meta": {
    "count": 1
  }
}
```

### `POST /api/rules`

Status:

- Draft

Purpose:

- save a new rule record

Draft request:

```json
{
  "imapAccountId": "acc_work_gmail",
  "name": "Invoice triage",
  "description": "Find invoice emails and move them to Receipts",
  "status": "draft",
  "sourceKind": "ui",
  "ruleYaml": "name: Invoice triage\nsearch:\n  subject_contains: invoice\noutput:\n  format: json\n  limit: 20\n  fields:\n    - subject\n    - from\n"
}
```

### `POST /api/rules/:id/dry-run`

Status:

- Draft

Purpose:

- execute a non-destructive preview for a saved rule using a stored account

Draft request:

```json
{
  "imapAccountId": "acc_work_gmail"
}
```

Draft response:

```json
{
  "data": {
    "ruleId": "rule_invoice_triage",
    "imapAccountId": "acc_work_gmail",
    "matchedCount": 12,
    "actionPlan": {
      "moveTo": "Receipts"
    },
    "sampleRows": [
      {
        "uid": 91230,
        "subject": "Invoice 9381",
        "from": "Billing Co <billing@example.com>",
        "whyMatched": ["subject_contains", "within_days"]
      }
    ],
    "createdAt": "2026-03-16T10:14:00Z"
  }
}
```

## Usage examples

### Health check

```bash
curl -s http://127.0.0.1:8080/healthz
```

### Readiness check

```bash
curl -s http://127.0.0.1:8080/readyz
```

### Info endpoint

```bash
curl -s http://127.0.0.1:8080/api/info | jq
```

## Related

- [01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md](../design-doc/01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md)
- [01-implementation-diary.md](./01-implementation-diary.md)
