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
      Note: Hosted JSON API router and response-envelope implementation
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/user.go
      Note: Local-development user resolution and header override
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go
      Note: Account CRUD, IMAP testing, mailbox listing, and preview behavior
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/rules/service.go
      Note: Rule CRUD, DSL validation, and dry-run behavior
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/smailnaild-local-account-flow.md
      Note: Manual local recreation and curl walkthrough
ExternalSources: []
Summary: Implemented API reference for the hosted smailnail backend, covering account management, mailbox previews, rule CRUD, dry-runs, local-user auth assumptions, and exact verification commands.
LastUpdated: 2026-03-16T10:02:00-04:00
WhatFor: Give the backend team and UI developer one source of truth for the implemented hosted backend contract.
WhenToUse: Use when wiring the frontend, validating local behavior, or checking whether an endpoint is implemented versus still future work.
---

# API reference for hosted smailnail account setup and rule dry-run backend

## Goal

Describe the implemented Phase 1 and Phase 2 hosted backend surface for `smailnaild`:

- account CRUD
- read-only IMAP account tests
- mailbox listing
- preview-message listing and message detail
- rule CRUD
- rule dry-runs

This document replaces the earlier draft assumptions with the actual current contract.

## Current development auth model

Hosted app login and OIDC session handling are not implemented yet. The backend currently uses a local development ownership model:

- default user ID: `local-user`
- per-request override header: `X-Smailnail-User-ID`
- all account and rule records are still keyed by `user_id`
- ownership enforcement happens server-side through `(user_id, id)` lookups

This means the frontend can already build as if a current user exists, but production auth still needs to be wired later.

## Required startup configuration

The hosted backend requires:

- `SMAILNAILD_ENCRYPTION_KEY`

That value must be a base64-encoded 32-byte AES-GCM key. The backend will not start without it, because stored IMAP passwords are encrypted at rest.

## Conventions

### Response envelope

Successful application responses use:

```json
{
  "data": {},
  "meta": {}
}
```

List endpoints omit `meta` only when there is nothing useful to add. Most list endpoints currently return `count`, and preview endpoints also return mailbox/pagination metadata.

### Error envelope

Errors use:

```json
{
  "error": {
    "code": "validation-error",
    "message": "invalid account input: label is required"
  }
}
```

Current stable error-code families:

- `validation-error`
- `not-found`
- `imap-error`
- `accounts-unavailable`
- `internal-error`
- `invalid-body`
- `invalid-query`
- `invalid-uid`

### Deletion behavior

Delete endpoints return:

- `204 No Content`

No response body is sent on success.

## Endpoint status table

| Endpoint | Method | Status | Purpose |
| --- | --- | --- | --- |
| `/healthz` | `GET` | Implemented | Liveness check |
| `/readyz` | `GET` | Implemented | Readiness check including DB ping |
| `/api/info` | `GET` | Implemented | Service metadata |
| `/api/accounts` | `GET` | Implemented | List current user's saved IMAP accounts |
| `/api/accounts` | `POST` | Implemented | Create a saved IMAP account |
| `/api/accounts/{id}` | `GET` | Implemented | Fetch one saved account |
| `/api/accounts/{id}` | `PATCH` | Implemented | Update a saved account |
| `/api/accounts/{id}` | `DELETE` | Implemented | Delete a saved account |
| `/api/accounts/{id}/test` | `POST` | Implemented | Run a read-only account test |
| `/api/accounts/{id}/mailboxes` | `GET` | Implemented | List mailboxes for one saved account |
| `/api/accounts/{id}/messages` | `GET` | Implemented | Fetch paginated preview messages |
| `/api/accounts/{id}/messages/{uid}` | `GET` | Implemented | Fetch one preview message in more detail |
| `/api/rules` | `GET` | Implemented | List saved rules |
| `/api/rules` | `POST` | Implemented | Create a saved rule |
| `/api/rules/{id}` | `GET` | Implemented | Fetch one saved rule |
| `/api/rules/{id}` | `PATCH` | Implemented | Update one saved rule |
| `/api/rules/{id}` | `DELETE` | Implemented | Delete one saved rule |
| `/api/rules/{id}/dry-run` | `POST` | Implemented | Run a dry-run preview for a saved rule |

## Health and service endpoints

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

Success:

```json
{
  "status": "ready"
}
```

Failure example:

```json
{
  "status": "not-ready",
  "error": "database is nil"
}
```

### `GET /api/info`

Purpose:

- expose service identity and DB target metadata

Response example:

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

## Account API

### `GET /api/accounts`

Purpose:

- return the current user's saved accounts
- include the latest saved account-test summary when available

Example response:

```json
{
  "data": [
    {
      "id": "acc-1",
      "userId": "local-user",
      "label": "Local Dovecot",
      "providerHint": "local",
      "server": "localhost",
      "port": 993,
      "username": "a",
      "mailboxDefault": "INBOX",
      "insecure": true,
      "authKind": "password",
      "secretKeyId": "env:smailnaild-encryption-key",
      "isDefault": true,
      "mcpEnabled": false,
      "createdAt": "2026-03-16T14:00:00Z",
      "updatedAt": "2026-03-16T14:00:00Z",
      "latestTest": {
        "success": true,
        "warningCode": "tls-verification-disabled",
        "createdAt": "2026-03-16T14:01:00Z"
      }
    }
  ],
  "meta": {
    "count": 1
  }
}
```

Notes:

- plaintext passwords never appear in responses
- ciphertext and nonce are server-only fields and are stripped before JSON responses

### `POST /api/accounts`

Purpose:

- save a new account with encrypted credentials

Request:

```json
{
  "label": "Local Dovecot",
  "providerHint": "local",
  "server": "localhost",
  "port": 993,
  "username": "a",
  "password": "pass",
  "mailboxDefault": "INBOX",
  "insecure": true,
  "authKind": "password",
  "isDefault": true,
  "mcpEnabled": false
}
```

Behavior:

- first account for a user is forced to `isDefault=true`
- if `isDefault=true`, other accounts for the same user are unset as default

Success:

```json
{
  "data": {
    "id": "acc-1",
    "label": "Local Dovecot",
    "providerHint": "local",
    "server": "localhost",
    "port": 993,
    "username": "a",
    "mailboxDefault": "INBOX",
    "insecure": true,
    "authKind": "password",
    "secretKeyId": "env:smailnaild-encryption-key",
    "isDefault": true,
    "mcpEnabled": false,
    "createdAt": "2026-03-16T14:00:00Z",
    "updatedAt": "2026-03-16T14:00:00Z"
  }
}
```

### `GET /api/accounts/{id}`

Purpose:

- fetch one saved account

Success:

```json
{
  "data": {
    "id": "acc-1",
    "label": "Local Dovecot",
    "providerHint": "local",
    "server": "localhost",
    "port": 993,
    "username": "a",
    "mailboxDefault": "INBOX",
    "insecure": true,
    "authKind": "password",
    "secretKeyId": "env:smailnaild-encryption-key",
    "isDefault": true,
    "mcpEnabled": false,
    "createdAt": "2026-03-16T14:00:00Z",
    "updatedAt": "2026-03-16T14:00:00Z"
  }
}
```

### `PATCH /api/accounts/{id}`

Purpose:

- partially update one saved account

Request fields are optional. Example:

```json
{
  "label": "Work mailbox",
  "mailboxDefault": "Archive",
  "password": "new-secret"
}
```

Notes:

- when `password` is provided, it is re-encrypted and replaces the stored secret
- if `isDefault=true`, other accounts for the same user are unset as default

### `DELETE /api/accounts/{id}`

Purpose:

- delete one saved account

Success:

- HTTP `204 No Content`

### `POST /api/accounts/{id}/test`

Purpose:

- run a structured read-only IMAP test against one saved account

Current supported mode:

- `read_only`

Request:

```json
{
  "mode": "read_only"
}
```

Response example:

```json
{
  "data": {
    "id": "test-1",
    "imapAccountId": "acc-1",
    "testMode": "read_only",
    "success": true,
    "tcpOk": true,
    "loginOk": true,
    "mailboxSelectOk": true,
    "listOk": true,
    "sampleFetchOk": true,
    "warningCode": "tls-verification-disabled",
    "details": {
      "mailbox": "INBOX",
      "server": "localhost",
      "port": 993,
      "selectedMailbox": "INBOX",
      "numMessages": 7,
      "sampleMailboxes": ["INBOX"],
      "sampleSubject": "smailnaild account integration 1773669355448646322"
    },
    "createdAt": "2026-03-16T14:01:00Z"
  }
}
```

Current warning behavior:

- `tls-verification-disabled` when the stored account has `insecure=true`
- `provider-app-password-recommended` on login failures for a few known consumer providers

Current error behavior:

- account-test failures are usually returned as `200 OK` with `success=false` and populated `errorCode` and `errorMessage`
- request-shape or account-ownership issues still return normal HTTP errors

### `GET /api/accounts/{id}/mailboxes`

Purpose:

- list available mailboxes for one saved account

Response:

```json
{
  "data": [
    { "name": "INBOX", "path": "INBOX" }
  ],
  "meta": {
    "count": 1
  }
}
```

### `GET /api/accounts/{id}/messages`

Purpose:

- fetch preview rows for one mailbox using the existing DSL fetch engine

Supported query parameters:

- `mailbox`
- `limit`
- `offset`
- `query`
- `unread_only`

Current implementation notes:

- `query` maps to `subject_contains`
- `unread_only=true` maps to `flags.not_has=["seen"]`
- the current preview fields are fixed to `uid`, `subject`, `from`, `to`, `date`, `flags`, `size`

Example:

```json
{
  "data": [
    {
      "uid": 7,
      "seqNum": 7,
      "subject": "smailnaild http integration 1773669405135960830",
      "from": [
        {
          "name": "Seeder",
          "address": "seed@example.com"
        }
      ],
      "to": [
        {
          "name": "User A",
          "address": "a@testcot"
        }
      ],
      "date": "2026-03-16T09:56:45-04:00",
      "flags": [],
      "size": 214,
      "totalCount": 1
    }
  ],
  "meta": {
    "mailbox": "INBOX",
    "count": 1,
    "limit": 20,
    "offset": 0,
    "totalCount": 1
  }
}
```

### `GET /api/accounts/{id}/messages/{uid}`

Purpose:

- fetch one preview message in more detail

Supported query parameters:

- `mailbox`

Current implementation notes:

- the backend resolves an exact UID by using the DSL fetch path with `after_uid` / `before_uid`
- MIME-part content is currently limited to text parts with a maximum preview length of 4096 bytes

Example:

```json
{
  "data": {
    "uid": 7,
    "seqNum": 7,
    "subject": "smailnaild account integration 1773669355448646322",
    "from": [
      {
        "name": "Seeder",
        "address": "seed@example.com"
      }
    ],
    "to": [
      {
        "name": "User A",
        "address": "a@testcot"
      }
    ],
    "date": "2026-03-16T09:55:55-04:00",
    "flags": [],
    "size": 221,
    "mimeParts": [
      {
        "type": "text/plain",
        "subtype": "plain",
        "size": 73,
        "content": "Hosted test body for smailnaild account integration 1773669355448646322\r\n"
      }
    ],
    "totalCount": 1
  },
  "meta": {
    "mailbox": "INBOX"
  }
}
```

## Rule API

### `GET /api/rules`

Purpose:

- list saved rules for the current user

Example:

```json
{
  "data": [
    {
      "id": "rule-1",
      "userId": "local-user",
      "imapAccountId": "acc-1",
      "name": "Invoice triage",
      "description": "Invoice triage",
      "status": "draft",
      "sourceKind": "ui",
      "ruleYaml": "name: Invoice triage\n...",
      "lastPreviewCount": 1,
      "lastRunAt": "2026-03-16T14:05:00Z",
      "createdAt": "2026-03-16T14:04:00Z",
      "updatedAt": "2026-03-16T14:05:00Z"
    }
  ],
  "meta": {
    "count": 1
  }
}
```

### `POST /api/rules`

Purpose:

- create a new saved rule

Request:

```json
{
  "imapAccountId": "acc-1",
  "name": "Invoice triage",
  "description": "Invoice triage",
  "status": "draft",
  "sourceKind": "ui",
  "ruleYaml": "name: placeholder\ndescription: placeholder\nsearch:\n  subject_contains: invoice\noutput:\n  format: json\n  limit: 10\n  fields:\n    - uid\n    - subject\n"
}
```

Behavior:

- the current DSL parser validates `ruleYaml`
- if `name` or `description` is provided in the request, they override the parsed YAML values
- the rule is then re-normalized and stored back as YAML so the record and DSL stay aligned

### `GET /api/rules/{id}`

Purpose:

- fetch one saved rule

### `PATCH /api/rules/{id}`

Purpose:

- partially update a saved rule

Supported optional fields:

- `imapAccountId`
- `name`
- `description`
- `status`
- `sourceKind`
- `ruleYaml`

Behavior:

- if `ruleYaml` is omitted, the stored YAML is reused
- if `name` or `description` changes, the backend reapplies those values to the parsed rule and stores normalized YAML again

### `DELETE /api/rules/{id}`

Purpose:

- delete one saved rule and its stored `rule_runs`

Success:

- HTTP `204 No Content`

### `POST /api/rules/{id}/dry-run`

Purpose:

- execute a non-destructive preview for a saved rule using a stored account

Request:

```json
{
  "imapAccountId": "acc-1"
}
```

Behavior:

- if `imapAccountId` is empty, the rule's stored `imap_account_id` is used
- the backend resolves the stored account credentials
- the backend selects the account's default mailbox
- the existing DSL fetch engine runs without executing actions
- a `rule_runs` record is persisted with:
  - `mode="dry_run"`
  - `matched_count`
  - `action_summary_json`
  - `sample_results_json`
- the parent rule is updated with:
  - `last_preview_count`
  - `last_run_at`

Current preview safeguards:

- if the rule limit is zero, the dry-run defaults to 20 rows
- if the rule limit is above 25, it is clamped to 25 rows for preview

Example:

```json
{
  "data": {
    "ruleId": "rule-1",
    "imapAccountId": "acc-1",
    "matchedCount": 1,
    "actionPlan": {
      "moveTo": "Archive"
    },
    "sampleRows": [
      {
        "uid": 7,
        "seqNum": 7,
        "subject": "smailnaild http integration 1773669405135960830",
        "from": [
          {
            "name": "Seeder",
            "address": "seed@example.com"
          }
        ],
        "flags": [],
        "size": 214,
        "totalCount": 1
      }
    ],
    "createdAt": "2026-03-16T14:05:00Z"
  }
}
```

## Manual local verification

Start the local Dovecot fixture:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker compose -f docker-compose.local.yml up -d dovecot
```

Set the encryption key and start the backend:

```bash
export SMAILNAILD_ENCRYPTION_KEY="$(openssl rand -base64 32)"
go run ./cmd/smailnaild serve
```

Then follow the curl walkthrough in:

- [smailnaild-local-account-flow.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/smailnaild-local-account-flow.md)

## Automated verification

Focused hosted-backend verification:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/smailnaild/...
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild/accounts -run TestServiceAgainstLocalDovecot -v
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild/rules -run TestDryRunAgainstLocalDovecot -v
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild -run TestHostedHTTPFlowAgainstLocalDovecot -v
SMAILNAILD_DOVECOT_TEST=1 go test ./...
```

## Related

- [01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md](../design-doc/01-implementation-plan-for-hosted-smailnail-account-setup-and-rule-dry-run-phases.md)
- [01-implementation-diary.md](./01-implementation-diary.md)
