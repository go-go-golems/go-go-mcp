---
Title: JS IMAP API and sandbox eval MCP architecture guide
Ticket: SMAILNAIL-004-JS-IMAP-EVAL-MCP-DESIGN
Status: active
Topics:
    - smailnail
    - go
    - email
    - review
    - mcp
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-goja/engine/factory.go
      Note: Runtime composition model for sandbox host
    - Path: go-go-goja/modules/common.go
      Note: Native module contract for require smailnail
    - Path: go-go-goja/pkg/doc/02-creating-modules.md
    - Path: jesus/pkg/engine/engine.go
    - Path: jesus/pkg/mcp/server.go
      Note: Local reference for executeJS style MCP exposure
    - Path: jesus/pkg/repl/model.go
    - Path: smailnail/cmd/smailnail/commands/fetch_mail.go
      Note: Evidence that command adapters should be separated from services
    - Path: smailnail/cmd/smailnail/commands/mail_rules.go
    - Path: smailnail/pkg/dsl/processor.go
      Note: Core message fetch pipeline to wrap behind services
    - Path: smailnail/pkg/dsl/types.go
      Note: Current rule and search model that should inform JS API shape
    - Path: smailnail/pkg/imap/layer.go
      Note: Current IMAP settings and connection entrypoint
    - Path: smailnail/pkg/mailgen/mailgen.go
ExternalSources: []
Summary: Detailed architecture and implementation guide for exposing smailnail IMAP operations through a go-go-goja native module and a sandboxed eval-style MCP host.
LastUpdated: 2026-03-08T22:52:20.19880282-04:00
WhatFor: Use this guide to design and implement a safe, intern-friendly JavaScript scripting layer over smailnail and to expose it through MCP.
WhenToUse: Read this before writing any smailnail JS bindings, runtime sandbox code, or executeJS-style MCP tools.
---


# JS IMAP API and sandbox eval MCP architecture guide

## Executive Summary

`smailnail` already contains the hard part of the domain problem: real IMAP connection settings, search/rule modeling, message fetching, and mail generation. What it does not yet contain is a reusable application-service layer, a JavaScript-facing API, or an MCP surface for running user-supplied scripts. The codebase therefore has a strong core but no scripting boundary.

The cleanest design is to treat this as a three-layer system rather than “add JavaScript directly to the CLI.” First, introduce or formalize pure Go services around IMAP sessions, rule execution, and message generation. Second, wrap those services as a `go-go-goja` native module so JavaScript code can call `require("smailnail")`. Third, host that runtime inside a sandbox controller that exposes an `executeJS`-style MCP tool, using `jesus` as the closest local reference for runtime bootstrapping and MCP registration.

This guide recommends a conservative sandbox. The default runtime should expose only the `smailnail` module and a very small set of utility globals. It should not expose arbitrary filesystem, shell, or unrestricted network modules unless an explicit product requirement later justifies them. That is the key product boundary between “email scripting” and “general code execution”.

## Problem Statement

The user wants to build a JavaScript API for IMAP functionality and expose it as a sandboxed eval-style MCP. That means a future client should be able to submit JavaScript like:

```javascript
const smailnail = require("smailnail");

const session = await smailnail.connect({
  server: "imap.example.com",
  port: 993,
  username: "user@example.com",
  password: secret,
  mailbox: "INBOX"
});

const messages = await session.search({
  subjectContains: "invoice",
  withinDays: 7,
  limit: 20
});

return messages.map(m => ({
  uid: m.uid,
  subject: m.subject,
  from: m.from,
}));
```

Today, there is no stable API layer for that. The current `smailnail` entrypoints are command-oriented:

- [`pkg/imap/layer.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go) defines connection settings and low-level `ConnectToIMAPServer()`.
- [`pkg/dsl/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go) defines the rule and search/output model.
- [`pkg/dsl/processor.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go) executes searches and fetches messages.
- [`cmd/smailnail/commands/fetch_mail.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/fetch_mail.go) and [`cmd/smailnail/commands/mail_rules.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/mail_rules.go) adapt CLI flags into rules and outputs.

Those are useful building blocks, but they are not yet a safe or ergonomic scripting surface.

## Scope

In scope for this ticket:

- a detailed design for a JavaScript API over IMAP and mail-generation functionality,
- an implementation plan for a `go-go-goja` native module,
- an implementation plan for a sandboxed eval-style runtime host,
- an MCP design for exposing that runtime to remote clients,
- testing and validation guidance for an intern or new maintainer.

Out of scope for this ticket:

- implementing the feature,
- picking a final auth/storage model for hosted deployment,
- building a browser UI,
- defining every future business workflow that could be scripted on top of the API.

## Current-State Architecture

### Layer 1: Existing smailnail domain logic

`smailnail` already has reusable behavior in package code, which is exactly what makes this design feasible.

Connection setup lives in [`pkg/imap/layer.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go). `IMAPSettings` at lines 13-20 is the current connection DTO, and `ConnectToIMAPServer()` at lines 67-87 performs TLS dialing and login. That means the repo already has a single obvious starting point for any future session abstraction.

The rule/search model lives in [`pkg/dsl/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go). `Rule` at lines 21-28 combines `Search`, `Output`, and `Actions`. `SearchConfig` at lines 52-82 and `OutputConfig` at lines 198-206 already model most of the inputs a scripting user would want. This is a strong sign that the future JS API should map to these concepts rather than inventing entirely new semantics.

The execution engine lives in [`pkg/dsl/processor.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go). `FetchMessages()` at lines 49-246 builds search criteria, executes IMAP search, paginates results, fetches metadata and content, and returns `EmailMessage` results. That is the core unit of work a JS API must eventually call.

Mail generation already exists too. [`pkg/mailgen/mailgen.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailgen/mailgen.go) defines `MailGenerator`, takes a validated template config, and renders email payloads through `Generate()` at lines 53-115. That creates a second natural module area besides IMAP fetch/search.

### Layer 2: Current CLI adapters

The current CLI commands are useful evidence for API shape, but they should not become the service boundary.

[`cmd/smailnail/commands/fetch_mail.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/fetch_mail.go) shows how flag-driven settings are translated into a `dsl.Rule`. The important design signal is not the Cobra code itself; it is the fact that the command already performs three conceptual steps:

- decode input settings,
- build a rule,
- connect to IMAP and execute it.

[`cmd/smailnail/commands/mail_rules.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/mail_rules.go) shows a sibling flow where the rule comes from YAML instead of CLI flags. That strongly suggests a future service API should accept pre-built rule objects and separate “input parsing” from “rule execution”.

### Layer 3: go-go-goja patterns

`go-go-goja` already provides the native-module and runtime-composition model this feature needs.

[`modules/common.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/modules/common.go) defines the `NativeModule` interface at lines 29-33. A module needs a `Name()`, `Doc()`, and `Loader(*goja.Runtime, *goja.Object)`. The registry at lines 35-102 is the shared mechanism for registering and enabling modules into a Goja `require` registry.

[`engine/factory.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go) shows the right runtime-building style. `FactoryBuilder` and `Factory` build an immutable runtime composition plan, register modules into a `require.Registry`, and then create isolated runtimes with event loops and initializers through `NewRuntime()` at lines 134-179. This is the right substrate for a sandbox host because it makes module exposure explicit instead of implicit.

### Layer 4: jesus as runtime-host reference

`jesus` is the closest existing example of “boot a JS runtime, preload modules, and expose eval through MCP.”

[`pkg/engine/engine.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/engine.go) shows a production-ish engine wrapper around Goja, the Node-style `require` registry, and an event loop. The most relevant lines are 71-75, where the default module registry is enabled, and lines 146-149, where code execution is wrapped as `ExecuteScript()`.

[`pkg/repl/model.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/repl/model.go) is useful because it proves the expected developer ergonomics: a runtime created through `ggjengine.New()` can support `require()`-based interactive scripting. Lines 45-49 say that explicitly.

[`pkg/mcp/server.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go) is the most important local MCP reference. It uses `go-go-mcp/pkg/embeddable`, registers an `executeJS` tool at lines 148-171, and defers runtime initialization depending on transport at lines 179-197. This is very close to the product shape requested by the user.

## Gap Analysis

The missing pieces are structural, not conceptual.

### What is already good

- `smailnail` already separates IMAP connection, rule modeling, fetching, and mail generation into packages.
- `go-go-goja` already gives a clean native-module abstraction and runtime factory.
- `jesus` already demonstrates how to host Goja plus MCP and how to structure an `executeJS` tool.

### What is missing

- `smailnail` has no dedicated service package for non-CLI callers.
- There is no `smailnail` JavaScript module contract yet.
- There is no sandbox policy layer.
- There is no serialization contract between IMAP Go types and JS values.
- There is no persistent execution-session concept for JS-driven IMAP work.
- There are no integration tests that boot a runtime and call `require("smailnail")`.

### What should not be done

- Do not expose Cobra commands directly to JavaScript.
- Do not let the first implementation inherit CLI flag names blindly where they hurt JS ergonomics.
- Do not default to exposing all available `go-go-goja` modules to the eval environment.
- Do not allow arbitrary unbounded IMAP fetches or uncontrolled attachment reads in the first version.

## Proposed Solution

The system should be split into four packages or package groups.

### 1. `smailnail` service layer

Create a service-oriented package tree that sits between existing package logic and the future JS adapter. A likely structure is:

```text
smailnail/
  pkg/
    service/
      imap/
        service.go
        session.go
        types.go
      mailgen/
        service.go
        types.go
```

This layer should:

- own connection/session lifecycle,
- convert JS/API-friendly option structs into existing `dsl.Rule` and related types,
- return simplified result objects suitable for JSON/JS export,
- centralize validation, limits, and guardrails.

It should not:

- know anything about Goja,
- know anything about Cobra/Glazed,
- expose raw `imapclient.Client` pointers outside carefully scoped internals.

### 2. Native `go-go-goja` module

Create a dedicated module package in either `smailnail` or a sibling repo, for example:

```text
smailnail/
  pkg/js/modules/smailnail/
    module.go
    session_wrappers.go
    codecs.go
```

This package should implement the `modules.NativeModule` interface from [`go-go-goja/modules/common.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/modules/common.go).

The exports should feel like a normal JavaScript library:

```javascript
const smailnail = require("smailnail");

const session = await smailnail.connect(opts);
const rows = await session.search(query);
const ruleResult = await session.runRule(rule);
const generated = await smailnail.generateEmails(templateConfig);
await session.close();
```

### 3. Sandboxed runtime host

Build a small host package that uses the `go-go-goja` factory pattern instead of constructing runtimes ad hoc. This runtime host should:

- register only approved modules,
- attach execution metadata such as request id, session id, and time limits,
- capture console output and execution results,
- optionally persist code snippets or execution logs,
- support transport-friendly inputs and outputs for MCP.

Use `jesus` as the conceptual reference, but avoid copying unnecessary web-app concerns.

### 4. MCP adapter

Add an MCP server that exposes at least one tool:

- `executeJS` or `executeIMAPJS`

The tool should accept:

- JavaScript source code,
- execution mode flags such as `awaitResult`,
- optionally a stored connection reference instead of raw credentials,
- optionally runtime limits like max messages or timeout.

It should return:

- the exported result value,
- captured console logs,
- structured error data,
- optionally execution metadata such as duration and result size.

## Recommended Package Boundaries

Use these ownership boundaries to avoid long-term drift.

### smailnail domain packages keep doing domain work

- IMAP connection and protocol handling remain in `pkg/imap` and service-layer packages.
- Rule and search semantics remain grounded in `pkg/dsl`.
- Mail generation remains grounded in `pkg/mailgen`.

### service packages become the stable application API

- The future JS module should call service types only.
- Any future HTTP API or hosted web UI should call the same services.
- This keeps JS, MCP, and web surfaces aligned.

### JS adapter packages do translation only

- decode JS objects,
- call Go services,
- convert results back into JS values,
- surface errors consistently.

## Detailed JS API Proposal

The first version should be intentionally small.

### Top-level module contract

```javascript
const smailnail = require("smailnail");

smailnail.connect(opts) -> Promise<Session>
smailnail.generateEmails(templateConfig) -> Promise<Email[]>
smailnail.validateRule(rule) -> ValidationResult
smailnail.buildRule(queryOpts) -> Rule
```

### Session object contract

```javascript
session.selectMailbox(name) -> Promise<MailboxInfo>
session.search(queryOpts) -> Promise<MessageSummary[]>
session.fetchByUid(uids, fetchOpts) -> Promise<Message[]>
session.runRule(rule) -> Promise<RuleResult>
session.append(email, appendOpts) -> Promise<AppendResult>
session.close() -> Promise<void>
```

### Suggested TypeScript-like shapes

```typescript
type ConnectOptions = {
  server: string;
  port?: number;
  username: string;
  password: string;
  mailbox?: string;
  insecure?: boolean;
};

type SearchOptions = {
  since?: string;
  before?: string;
  withinDays?: number;
  from?: string;
  to?: string;
  subject?: string;
  subjectContains?: string;
  bodyContains?: string;
  hasFlags?: string[];
  notHasFlags?: string[];
  largerThan?: string;
  smallerThan?: string;
  limit?: number;
  offset?: number;
  afterUid?: number;
  beforeUid?: number;
  includeContent?: boolean;
  contentType?: string;
  contentMaxLength?: number;
};

type MessageSummary = {
  uid: number;
  subject?: string;
  from?: string;
  to?: string[];
  date?: string;
  flags?: string[];
  size?: number;
  content?: string;
  mimeParts?: MimePart[];
};
```

### Why this API shape

It mirrors the existing semantics in `dsl.SearchConfig` and `dsl.OutputConfig`, but renames fields into lowerCamelCase and groups operations around a session object. That is easier for JavaScript users than forcing them to think in Glazed/Cobra sections or CLI flag names.

## Data Conversion Strategy

This part matters more than it first appears. A bad JS bridge becomes impossible to maintain.

### Conversion rules

- Use lowerCamelCase in JS-facing APIs.
- Return plain objects, arrays, booleans, numbers, and strings.
- Convert Go `time.Time` values to RFC3339 strings.
- Convert Go zero values into omitted or `null`-equivalent fields only when the absence is semantically meaningful.
- Avoid returning raw Go structs with accidental method exposure.

### Error rules

- Validation problems should throw normal JS errors with actionable messages.
- IMAP connection/auth failures should include a stable error code.
- Execution-limits failures should be clearly distinguishable from domain failures.

Suggested shape:

```javascript
{
  code: "imap_auth_failed",
  message: "failed to login: ...",
  retryable: false
}
```

## Sandbox Design

This is the highest-risk area because the user explicitly wants “sandbox JS eval”.

### Security goal

Allow users to script email workflows, not to turn the runtime into a general-purpose shell.

### Minimum sandbox policy

- expose only `smailnail` and minimal safe helpers,
- no raw file I/O module by default,
- no shell/process execution,
- no ambient environment-variable dump,
- no arbitrary TCP or HTTP clients unless explicitly approved,
- enforce per-execution timeout,
- enforce max messages/max content bytes,
- enforce session cleanup after script completion,
- capture and log execution metadata.

### Runtime diagram

```text
MCP Client
    |
    v
executeJS tool
    |
    v
Sandbox Controller
    |
    +--> go-go-goja Factory
    |        |
    |        +--> require("smailnail")
    |        +--> console
    |        +--> limited utility globals
    |
    v
smailnail service layer
    |
    +--> IMAP service
    +--> Rule execution service
    +--> Mail generation service
```

### Execution model choices

Option A: ephemeral runtime per invocation.

- Pros: strongest isolation, easiest cleanup.
- Cons: no cross-invocation state unless explicitly persisted.

Option B: persistent runtime sessions.

- Pros: can cache connections and script state.
- Cons: more complicated lifecycle and stronger isolation requirements.

Recommendation: start with ephemeral runtimes and explicit connection handles or short-lived session objects inside one invocation. Add persistent sessions only if there is a proven workflow that needs them.

## MCP Design

### Minimum tool surface

Start with one primary eval tool and consider a few structured tools later.

```text
Tool: executeIMAPJS
Args:
  code: string
  connection: object?        # optional direct credentials
  connectionId: string?      # preferred if a future hosted service stores accounts
  timeoutMs: int?
  maxMessages: int?

Result:
  value: any
  consoleLog: string[]
  durationMs: int
  error?: object
```

### Why not only structured tools

Purely structured tools are safer, but the request is specifically for a sandbox JS eval MCP. If the product goal is user programmability, the structured tools should complement the eval tool rather than replace it. They are useful for:

- health checks,
- account validation,
- common fetch patterns,
- easier automated testing.

### Why not only eval

Eval is flexible but harder to reason about. A mixed model is better:

- `executeIMAPJS` for power users,
- `listMailboxes`, `testConnection`, `fetchRecentMessages` for simpler or lower-risk consumers.

## Intern-Facing Implementation Plan

### Phase 1: stabilize Go services

Goal: create pure Go APIs that are safe to call from JS or HTTP.

Files likely to add:

- `smailnail/pkg/service/imap/service.go`
- `smailnail/pkg/service/imap/session.go`
- `smailnail/pkg/service/imap/types.go`
- `smailnail/pkg/service/mailgen/service.go`

Tasks:

- define input and output structs,
- wrap `ConnectToIMAPServer()` behind a service API,
- add `SearchOptions -> dsl.Rule` conversion helpers,
- add tests that do not involve Goja.

Pseudocode:

```go
type IMAPService struct{}

func (s *IMAPService) Connect(ctx context.Context, opts ConnectOptions) (*Session, error) {
    settings := imap.IMAPSettings{...}
    client, err := settings.ConnectToIMAPServer()
    if err != nil { return nil, wrapErr(err) }
    return NewSession(client, opts.Mailbox), nil
}

func (s *Session) Search(ctx context.Context, opts SearchOptions) ([]MessageSummary, error) {
    rule := BuildRuleFromSearchOptions(opts)
    msgs, err := rule.FetchMessages(s.client)
    if err != nil { return nil, wrapErr(err) }
    return summarizeMessages(msgs), nil
}
```

### Phase 2: implement the native module

Goal: make `require("smailnail")` work in a Goja runtime.

Files likely to add:

- `smailnail/pkg/js/modules/smailnail/module.go`
- `smailnail/pkg/js/modules/smailnail/codecs.go`
- `smailnail/pkg/js/modules/smailnail/session.go`

Tasks:

- implement `modules.NativeModule`,
- register it via `init()` and `modules.Register(...)`,
- export top-level functions,
- wrap service-layer objects as JS-visible session objects,
- add runtime integration tests.

Pseudocode:

```go
type Module struct {
    service *imapservice.IMAPService
}

func (m *Module) Name() string { return "smailnail" }

func (m *Module) Loader(rt *goja.Runtime, module *goja.Object) {
    exports := module.Get("exports").(*goja.Object)
    exports.Set("connect", m.connect(rt))
    exports.Set("generateEmails", m.generateEmails(rt))
    exports.Set("validateRule", m.validateRule(rt))
}
```

### Phase 3: build the sandbox host

Goal: create a reusable engine that boots only approved modules.

Files likely to add:

- `smailnail/pkg/jshost/factory.go`
- `smailnail/pkg/jshost/runtime.go`
- `smailnail/pkg/jshost/limits.go`
- `smailnail/pkg/jshost/result.go`

Tasks:

- define allowed modules,
- set up console capture,
- attach timeouts and result-size limits,
- serialize errors consistently.

### Phase 4: expose MCP

Goal: publish the runtime over `go-go-mcp`.

Files likely to add:

- `smailnail/pkg/mcp/server.go`
- `smailnail/cmd/smailnaild/main.go` or a sibling hosted binary

Tasks:

- register `executeIMAPJS`,
- add help/docs,
- optionally add `testConnection` and `fetchRecentMessages`,
- add transport smoke tests.

## Testing Strategy

This feature needs three test layers. Do not skip any of them.

### 1. Service tests

These should prove:

- connection validation,
- rule conversion correctness,
- result shaping,
- bounded fetch behavior.

### 2. Goja integration tests

These should boot a real runtime and execute:

```javascript
const smailnail = require("smailnail");
typeof smailnail.connect
```

Then test one or two real flows, ideally against the existing Docker IMAP setup used in the `smailnail` workstream.

### 3. MCP smoke tests

These should start the MCP server and execute a script over:

- `command`,
- `sse`,
- `streamable_http`

Return assertions should cover:

- result value shape,
- console capture,
- timeout enforcement,
- sanitized errors.

## Review Checklist For A New Intern

When reviewing an implementation, start with these questions.

1. Does the JS module call only service-layer packages, or is it reaching back into Cobra commands?
2. Are JS option names lowerCamelCase and documented?
3. Are returned objects plain JSON-like values rather than raw Go structs?
4. Does the runtime expose only intended modules?
5. Are timeout and fetch-size limits enforced in both Go services and runtime control paths?
6. Does the MCP tool return stable machine-readable errors?
7. Are there both service tests and Goja integration tests?

## Alternatives Considered

### Alternative 1: expose CLI commands to JS

Rejected because it would couple the JS API to Glazed/Cobra argument parsing, make testing awkward, and create long-term drift between CLI behavior and API behavior.

### Alternative 2: skip Goja native modules and use JSON-RPC between JS and Go

Rejected for the first version because the local codebase already has a solid native-module model in `go-go-goja`, and adding an internal RPC boundary would be more moving parts with no clear gain yet.

### Alternative 3: put all logic inside the MCP server binary

Rejected because it would make the runtime host own domain logic that should instead live in reusable `smailnail` packages.

## Risks

### Risk 1: sandbox creep

If unrelated modules get exposed “for convenience”, the product stops being an IMAP scripting environment and becomes a general code-execution environment. Resist that.

### Risk 2: result-shape drift

If CLI types, Go structs, and JS exports all evolve independently, consumers will get a brittle API. Fix this with a single service-layer contract and runtime tests.

### Risk 3: session lifecycle leaks

Persistent IMAP clients inside runtimes can leak sockets or mailbox state. Start with ephemeral execution unless persistence is clearly needed.

## Open Questions

- Should the first version allow raw credentials in `executeIMAPJS`, or require stored connection profiles?
- Should `mailgen` ship in the same JS module from day one, or wait until the IMAP API stabilizes?
- Does the hosted product need persistent JS state, or is per-invocation state enough?
- Should attachments be exposed in the first JS API, or only text content and metadata?

## References

- [`smailnail/pkg/imap/layer.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go)
- [`smailnail/pkg/dsl/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go)
- [`smailnail/pkg/dsl/processor.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go)
- [`smailnail/cmd/smailnail/commands/fetch_mail.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/fetch_mail.go)
- [`smailnail/cmd/smailnail/commands/mail_rules.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/mail_rules.go)
- [`smailnail/pkg/mailgen/mailgen.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailgen/mailgen.go)
- [`go-go-goja/modules/common.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/modules/common.go)
- [`go-go-goja/engine/factory.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/engine/factory.go)
- [`go-go-goja/pkg/doc/02-creating-modules.md`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-goja/pkg/doc/02-creating-modules.md)
- [`jesus/pkg/engine/engine.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/engine/engine.go)
- [`jesus/pkg/mcp/server.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/mcp/server.go)
- [`jesus/pkg/repl/model.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/jesus/pkg/repl/model.go)

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
