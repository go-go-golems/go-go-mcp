---
Title: UX handover for account add and lightweight connection test
Ticket: SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION
Status: active
Topics:
    - smailnail
    - frontend
    - ux
    - onboarding
    - imap
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go
      Note: Hosted API routes the UX flow will call
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go
      Note: Backend account behavior and test-result semantics the UI must respect
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/smailnaild-local-account-flow.md
      Note: Manual local workflow and curl examples for backend recreation
    - Path: /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-013-ACCOUNT-SETUP-IMPLEMENTATION--implement-hosted-smailnail-account-setup-phases-1-and-2/reference/02-api-reference-for-hosted-smailnail-account-setup-and-rule-dry-run-backend.md
      Note: Current API contract and payload shapes
ExternalSources: []
Summary: Focused UX handoff for the first frontend slice only: adding an IMAP account and running the initial lightweight connection test.
LastUpdated: 2026-03-16T10:10:00-04:00
WhatFor: Give the UX/frontend developer a narrow, implementation-ready starting point without pulling them into mailbox exploration or rules yet.
WhenToUse: Use when designing or implementing the first hosted frontend slice for smailnail account onboarding.
---

# UX handover for account add and lightweight connection test

## Scope

This handoff is intentionally narrow.

The UX developer should focus only on:

- the first-run account onboarding view
- adding one IMAP account
- editing basic account settings
- running the initial lightweight connection test
- showing test success, warnings, and failure states clearly

Do not design these yet:

- mailbox explorer
- message preview
- rule library
- rule builder
- dry-run screens
- MCP settings or policy controls

Those can be designed after this first slice is stable.

## Product goal for this slice

A new user should be able to answer one question quickly:

> “Can I connect my mailbox to smailnail safely enough to continue?”

This first slice is not about power-user configuration. It is about reducing uncertainty and helping a user get to a confident “yes” or a legible “not yet.”

That means the UI should optimize for:

- trust
- clear field guidance
- low-friction setup
- immediate feedback
- understandable failure states

It should not yet optimize for:

- advanced IMAP options
- browsing or triaging mail
- multi-step automation

## User needs

### Primary user needs

- I need to understand what smailnail needs from me.
- I need to add one account without guessing what each field means.
- I need to know whether the connection works before I invest more time.
- If something fails, I need a concrete reason and a next step.
- I need to feel safe entering credentials.

### Secondary user needs

- I may want to label the account in a human way, not just by hostname.
- I may want to mark one account as my default.
- I may want reassurance that this is a read-only test.

### Non-needs for this slice

- I do not yet need inbox browsing.
- I do not yet need rule creation.
- I do not yet need a server dashboard.

## UX principles for this slice

### Principle 1: keep the form boring and explicit

This screen should feel practical, not clever. The user is entering infrastructure credentials. The UI should lean toward calm and legible over playful.

### Principle 2: test is the main action, save is only part of the journey

The user mentally thinks:

1. fill the fields
2. try the connection
3. see whether it works
4. decide whether to continue

So the screen should make “Test connection” or “Save and test” central, not hide testing as a secondary afterthought.

### Principle 3: success should explain what was proven

A generic green check is not enough. The user should understand that the backend validated:

- TLS connection
- login
- mailbox selection
- mailbox listing
- sample fetch

### Principle 4: failure should be stage-specific

A single red “connection failed” message is too weak. The UI should surface the backend’s stage-specific test result and make the next action obvious.

## Backend reality the UX should respect

The current backend already exists and works. Relevant endpoints:

- `POST /api/accounts`
- `PATCH /api/accounts/{id}`
- `POST /api/accounts/{id}/test`
- `GET /api/accounts`
- `GET /api/accounts/{id}`

Important backend facts:

- account secrets are stored encrypted at rest
- the test is currently read-only
- the backend returns structured booleans such as:
  - `tcpOk`
  - `loginOk`
  - `mailboxSelectOk`
  - `listOk`
  - `sampleFetchOk`
- the backend may return warnings like:
  - `tls-verification-disabled`
  - `provider-app-password-recommended`

The UX should present those as human states, not raw engineering jargon, but it should preserve their structure.

## Recommended first flow

### Flow summary

1. User lands on empty-state account screen.
2. User clicks `Add account`.
3. User fills a compact account form.
4. User clicks `Save and test connection`.
5. UI creates the account.
6. UI immediately runs the read-only test.
7. UI shows:
   - success summary
   - warning summary
   - failure summary with likely fix direction
8. UI keeps the form available for editing and retesting.

### Why this flow

This avoids a split mental model where “saving” and “testing” feel unrelated. Saving is only useful because it enables testing and later use.

## Information architecture for this slice

For now, the UX can be just one area with two states:

- `No accounts yet`
- `Account editor + latest test result`

Do not build a full left-nav or multi-page app experience yet unless the frontend shell already requires one. If a shell exists, keep navigation minimal:

- `Accounts`

That is enough for the first handoff.

## Recommended screen states

### State 1: empty state

Use this when `GET /api/accounts` returns zero rows.

Goals:

- explain what an account is
- reassure about the first test being read-only
- get the user to the form quickly

Suggested content:

- headline: `Connect your first mailbox`
- body: `Add an IMAP account so smailnail can test the connection and prepare mailbox previews. The first test does not modify mail.`
- primary action: `Add account`

ASCII wireframe:

```text
+--------------------------------------------------------------+
| Accounts                                                     |
|                                                              |
|  Connect your first mailbox                                  |
|  Add an IMAP account so smailnail can test the connection    |
|  and prepare mailbox previews. The first test is read-only.  |
|                                                              |
|  [ Add account ]                                             |
|                                                              |
+--------------------------------------------------------------+
```

### State 2: add-account form

This is the core first-screen experience.

Recommended fields, in order:

- Label
- Email / Username
- IMAP server
- Port
- Password / app password
- Default mailbox
- Advanced toggle:
  - `Skip TLS verification`
  - `Set as default account`

Do not lead with advanced settings.

ASCII wireframe:

```text
+--------------------------------------------------------------+
| Accounts                                                     |
|                                                              |
|  Add IMAP account                                            |
|  Use the mailbox credentials you already use in your mail    |
|  client. We will store the password encrypted and run a      |
|  read-only connection test.                                  |
|                                                              |
|  Label                  [ Work inbox                    ]    |
|  Username or email      [ a                             ]    |
|  IMAP server            [ localhost                     ]    |
|  Port                   [ 993                           ]    |
|  Password               [ ************                  ]    |
|  Default mailbox        [ INBOX                         ]    |
|                                                              |
|  [>] Advanced                                               |
|                                                              |
|  [ Cancel ]                          [ Save and test ]       |
|                                                              |
+--------------------------------------------------------------+
```

### State 3: testing in progress

This should not be a spinner with no explanation. Show progress as a calm checklist or staged status list.

Suggested stages:

- Connecting securely
- Signing in
- Opening mailbox
- Listing mailboxes
- Fetching a sample message

ASCII wireframe:

```text
+--------------------------------------------------------------+
| Testing connection                                           |
|                                                              |
|  We are checking this account in read-only mode.             |
|                                                              |
|  [✓] Connecting securely                                     |
|  [✓] Signing in                                              |
|  [~] Opening mailbox                                         |
|  [ ] Listing mailboxes                                       |
|  [ ] Fetching sample message                                 |
|                                                              |
+--------------------------------------------------------------+
```

### State 4: success

A successful result should do more than say “passed.”

Show:

- result headline
- what was validated
- any warnings
- next action

ASCII wireframe:

```text
+--------------------------------------------------------------+
| Connection looks good                                        |
|                                                              |
|  smailnail connected to this mailbox and completed the       |
|  initial read-only checks.                                   |
|                                                              |
|  [✓] Connected over TLS                                      |
|  [✓] Logged in                                               |
|  [✓] Opened INBOX                                            |
|  [✓] Listed mailboxes                                        |
|  [✓] Fetched a sample message                                |
|                                                              |
|  Sample mailbox: INBOX                                       |
|  Sample subject: “Quarterly invoice”                         |
|                                                              |
|  [ Edit account ]                 [ Continue later ]         |
|                                                              |
+--------------------------------------------------------------+
```

### State 5: success with warning

Warnings are not failures. Visually distinguish them from hard errors.

Example warnings:

- TLS verification disabled
- provider app password recommended

Recommended treatment:

- success banner remains green or positive
- warning sits beneath it in amber
- warning copy must explain consequence, not only name the warning

ASCII wireframe:

```text
+--------------------------------------------------------------+
| Connection works, but needs attention                        |
|                                                              |
|  smailnail reached the mailbox successfully.                 |
|                                                              |
|  Warning                                                     |
|  TLS verification is disabled for this account. This is      |
|  acceptable for local testing, but it should not stay this   |
|  way for a production mailbox.                               |
|                                                              |
|  [ Keep editing ]                 [ Save anyway ]            |
|                                                              |
+--------------------------------------------------------------+
```

### State 6: failure

Failure states should be stage-specific. The current backend can distinguish multiple stages. The UX should translate those into user-facing language.

Recommended mapping:

- `account-test-connect-failed`
  - “Could not reach the IMAP server.”
- `account-test-login-failed`
  - “The server was reachable, but the sign-in failed.”
- `account-test-mailbox-select-failed`
  - “The account worked, but the default mailbox could not be opened.”
- `account-test-mailbox-list-failed`
  - “Sign-in worked, but mailbox listing failed.”
- `account-test-sample-fetch-failed`
  - “Connection worked, but fetching a sample message failed.”

ASCII wireframe:

```text
+--------------------------------------------------------------+
| Connection failed                                            |
|                                                              |
|  The server was reachable, but sign-in failed.               |
|                                                              |
|  Try checking:                                               |
|  - username or email                                         |
|  - password or app password                                  |
|  - provider-specific IMAP access settings                    |
|                                                              |
|  Technical details                                           |
|  errorCode: account-test-login-failed                        |
|                                                              |
|  [ Edit details ]                 [ Test again ]             |
|                                                              |
+--------------------------------------------------------------+
```

## Field guidance and copy suggestions

### Label

Purpose:

- purely for the human user
- should support multiple accounts later

Helpful placeholder:

- `Work inbox`

### Username

Use label:

- `Username or email`

Why:

- some users think in email address form
- some providers still expect a plain username

### IMAP server

Use label:

- `IMAP server`

Helper text:

- `Example: imap.gmail.com or mail.example.com`

### Port

Default:

- `993`

Treat this as standard, not as an advanced option.

### Password

Use label:

- `Password or app password`

Why:

- this will prevent a common failure mode for Gmail, iCloud, Yahoo, and similar providers

### Default mailbox

Default:

- `INBOX`

Do not hide this field. The backend test uses it directly.

## Interaction recommendations

### Primary action wording

Recommended:

- `Save and test`

Why:

- it matches the actual behavior
- it removes ambiguity around whether the UI will validate the account right now

### Secondary action wording

Recommended:

- `Cancel`
- `Edit account`
- `Test again`

Avoid vague actions like:

- `Done`
- `Apply`

### Loading behavior

The form submit should:

- disable the primary action
- preserve all field values
- keep the form visible
- show test progress inline

Do not navigate away during testing.

## Recommended frontend data model for this slice

The frontend does not need a full normalized client data layer yet. A simple screen-level state model is enough.

Suggested state buckets:

- `formDraft`
- `saveState`
- `testState`
- `latestAccount`
- `latestTestResult`

Pseudocode:

```ts
type AccountSetupScreenState = {
  formDraft: {
    label: string
    username: string
    server: string
    port: number
    password: string
    mailboxDefault: string
    insecure: boolean
    isDefault: boolean
  }
  saveState: "idle" | "saving" | "saved" | "error"
  testState: "idle" | "running" | "success" | "warning" | "failure"
  latestAccount?: Account
  latestTestResult?: AccountTestResult
  error?: UIError
}
```

## Recommended frontend sequence

Pseudocode:

```ts
async function onSaveAndTest(formDraft) {
  const account = await api.createAccount(formDraft)
  setLatestAccount(account)

  const testResult = await api.runAccountTest(account.id, { mode: "read_only" })
  setLatestTestResult(testResult)

  if (testResult.success && testResult.warningCode) {
    setTestState("warning")
  } else if (testResult.success) {
    setTestState("success")
  } else {
    setTestState("failure")
  }
}
```

## What the UX developer should not worry about yet

- final auth or OIDC screens
- inbox exploration IA
- rule-writing UX
- background polling
- multiple-account comparison views
- production-grade settings navigation

The right move is to get the first account experience legible and trustworthy before broadening the surface.

## Backend edge cases the UX should expect

- The backend may return `200 OK` for a failed account test, with `success=false`.
- The backend may return warnings even when `success=true`.
- Delete endpoints return `204` with no JSON body.
- The current backend can simulate a user through `X-Smailnail-User-ID`, but the frontend does not need to expose that.

## Recommended design deliverables

The UX developer should produce:

- one empty-state screen
- one add/edit account form
- one testing-in-progress state
- one success state
- one warning state
- one failure state

That is enough to unblock the first frontend implementation slice.

## Exact backend references

- API contract: [02-api-reference-for-hosted-smailnail-account-setup-and-rule-dry-run-backend.md](../reference/02-api-reference-for-hosted-smailnail-account-setup-and-rule-dry-run-backend.md)
- implementation diary: [01-implementation-diary.md](../reference/01-implementation-diary.md)
- local backend walkthrough: [smailnaild-local-account-flow.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/smailnaild-local-account-flow.md)
