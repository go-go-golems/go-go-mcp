---
Title: IMAP JS MCP and Queryable Documentation Architecture Guide
Ticket: SMAILNAIL-006-IMAP-JS-MCP-DOCS-DESIGN
Status: active
Topics:
    - smailnail
    - mcp
    - javascript
    - documentation
    - oidc
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-goja/modules/glazehelp/glazehelp.go
      Note: Provides a query-first documentation API shape worth copying
    - Path: go-go-goja/pkg/jsdoc/extract/extract.go
      Note: Provides the sentinel-based JS documentation extraction model
    - Path: go-go-goja/pkg/jsdoc/model/store.go
      Note: Provides the query-oriented documentation store indexes
    - Path: jesus/pkg/mcp/server.go
      Note: Provides the thin executeJS-over-MCP reference pattern
    - Path: smailnail/pkg/js/modules/smailnail/module.go
      Note: Defines the current JavaScript module surface that the MCP should host
    - Path: smailnail/pkg/services/smailnailjs/service.go
      Note: Defines the canonical JS-facing service options and runtime behavior
ExternalSources: []
Summary: Recommends a dedicated smailnail MCP with executeIMAPJS and a structured getIMAPJSDocumentation tool backed by go-go-goja jsdoc extraction plus Go-side drift validation.
LastUpdated: 2026-03-09T00:43:01.911684934-04:00
WhatFor: Provide an evidence-based design for a minimal IMAP JavaScript MCP server with a rich documentation query tool.
WhenToUse: Use when implementing the first hosted MCP for the smailnail JavaScript surface or when deciding how to author and query JS API documentation in Goja-based systems.
---


# IMAP JS MCP and Queryable Documentation Architecture Guide

## Executive Summary

The right v1 is a dedicated MCP server that exposes exactly two tools: `executeIMAPJS` and `getIMAPJSDocumentation`. The execution tool should stay thin and delegate real behavior to the existing `smailnail` Go service layer and Goja native module surface. The documentation tool should not return one giant markdown document. It should query structured documentation extracted from JavaScript companion files using the already-migrated `go-go-goja/pkg/jsdoc` stack.

The strongest local evidence is already present in the workspace. `smailnail` has a reusable module and service layer, but its current module documentation is just a one-line string in `Module.Doc()` and is not rich enough for model guidance or onboarding. `jesus` demonstrates the minimal one-tool MCP pattern, but its documentation approach is static markdown loaded as a whole file. `go-go-goja` already contains the better answer: a parser, data model, store, and export path for structured JS documentation using sentinel patterns like `__package__`, `__doc__`, `__example__`, and `doc\`...\``.

The recommended design is a hybrid:

- Canonical prose, concepts, and examples live in JavaScript sentinel doc files close to the `smailnail` module.
- A Go-side doc registry loads those files into a `DocStore` at startup.
- `getIMAPJSDocumentation` provides structured queries by symbol, concept, example, and package.
- Optional Go reflection and descriptor checks are added as test-time validation to detect drift between exported Goja symbols and documented symbols.

This keeps the docs readable and example-rich for humans, queryable for agents, and checkable by Go-based tests.

## Problem Statement

The current `smailnail` JavaScript module exists, but there is no dedicated MCP host around it yet. The module implementation in `smailnail/pkg/js/modules/smailnail/module.go` exports `parseRule`, `buildRule`, and `newService`, and the service object exposes `parseRule`, `buildRule`, and `connect`. That is enough to begin, but not enough to self-document well for a model or a new engineer. The module currently exposes only a short summary string from `Module.Doc()` rather than detailed symbol-level documentation.

At the same time, the codebase already contains two nearby patterns that point in different directions:

- `jesus/pkg/mcp/server.go` shows how to expose a single JavaScript execution tool over MCP, but it loads documentation from a static markdown file using `doc.GetJavaScriptAPIReference()`.
- `go-go-goja/pkg/jsdoc` provides a richer documentation system that extracts structured docs from JavaScript sources and indexes them for lookup by package, symbol, concept, and example.

The design problem is therefore not just "how do we add one more MCP tool?" It is "how do we build a minimal IMAP JavaScript MCP that stays small while also exposing rich, queryable, low-drift API documentation?"

The documentation question matters because an eval-style tool with weak docs becomes brittle quickly:

- users and models guess symbol names incorrectly,
- examples drift away from the actual API,
- prose lives far away from exports,
- and every change requires manually editing broad markdown pages that are hard to query.

The requested outcome is more ambitious than that. The user explicitly wants a rich documentation querying tool and asked whether it makes sense to create a more general Go-side `jsdocex`-style facility for Goja scenarios. That means the design must compare alternatives, not just pick the first implementation that works.

## Scope

This ticket covers design only. It does not implement the new MCP in this turn.

In scope:

- architecture for a dedicated IMAP JavaScript MCP,
- design of the two MCP tools,
- documentation-authoring and querying strategy,
- file layout and package responsibilities,
- validation strategy and rollout plan.

Out of scope for this ticket:

- production implementation,
- OIDC deployment details beyond acknowledging future integration points,
- expanding the JS API beyond the current `smailnail` module surface,
- broader hosted multi-user configuration design from earlier tickets.

## Proposed Solution

### Current State

#### `smailnail` already has the reusable runtime layer needed for a thin MCP host

The native Goja module is in `smailnail/pkg/js/modules/smailnail/module.go`. The module registers itself in `init()` and exposes the name `smailnail` at lines 29 through 35. The exports are built in `Loader()` at lines 41 through 57:

- `parseRule(yamlString)`
- `buildRule(input)`
- `newService()`

The service object created by `newService()` exposes `parseRule`, `buildRule`, and `connect` at lines 59 through 85. That is already the correct architecture boundary for a future eval-style MCP: the handler can boot a runtime and register one module instead of owning IMAP logic itself.

The canonical business behavior is in `smailnail/pkg/services/smailnailjs/service.go`. Important pieces include:

- JS-facing option structs like `BuildRuleOptions` and `ConnectOptions` at lines 14 through 46,
- decoding helpers `DecodeBuildRuleOptions` and `DecodeConnectOptions` at lines 83 through 89,
- rule parsing and building methods at lines 91 through 196,
- connection handling at lines 202 through 258.

That service package is the right place to keep API-shape truth. The future MCP layer should depend on this package rather than duplicate argument and result semantics in handler code.

#### The current module documentation is too shallow

`Module.Doc()` in `smailnail/pkg/js/modules/smailnail/module.go` lines 37 through 39 returns one sentence: "smailnail exposes rule helpers and IMAP service construction for JavaScript runtimes." That is useful as a module label, but it is not enough for:

- onboarding,
- LLM guidance,
- example discovery,
- symbol-level lookup,
- or enforcing documentation completeness.

This is why a dedicated `getIMAPJSDocumentation` tool is necessary rather than just embedding the one-line module doc in the MCP tool description.

#### `jesus` demonstrates the minimal MCP hosting pattern, but not the right documentation architecture

`jesus/pkg/mcp/server.go` shows a workable model for the execution side. `AddMCPCommand()` registers a small MCP surface using `embeddable.AddMCPCommand` at lines 117 through 176, and the actual execution handler is `executeJSHandler` at lines 310 through 395.

Architecturally, the useful lessons are:

- keep the MCP surface thin,
- initialize runtime dependencies lazily when possible,
- register a very small number of tools,
- shape JSON responses in one place.

But the documentation path in `jesus` is weaker. `AddMCPCommand()` reads a whole markdown document using `doc.GetJavaScriptAPIReference()` at lines 126 through 131, which is defined in `jesus/pkg/doc/doc.go` lines 27 through 33 as a simple embedded file read. That approach is adequate for a static reference page, but not for structured querying by symbol, concept, or example.

#### `go-go-goja` already contains the documentation system this MCP should reuse

The key discovery is that the workspace already has a migrated jsdocex-like stack in `go-go-goja`.

The CLI entrypoint `go-go-goja/cmd/goja-jsdoc/main.go` lines 17 through 21 explicitly describes itself as "JavaScript doc extraction and browser (jsdocex migrated into go-go-goja)." That is strong evidence that a new parallel documentation system would be redundant unless there is a hard mismatch.

The data model is in `go-go-goja/pkg/jsdoc/model/model.go`:

- `SymbolDoc` at lines 17 through 33,
- `Example` at lines 35 through 48,
- `Package` at lines 50 through 63,
- `FileDoc` at lines 65 through 71.

This model already supports the kinds of data a rich MCP documentation tool needs:

- summaries,
- long-form prose,
- concepts,
- tags,
- params and returns,
- example bodies,
- source file and line anchors.

The aggregation layer is `go-go-goja/pkg/jsdoc/model/store.go`. `DocStore` at lines 3 through 14 maintains indexes by package, symbol, example, and concept. `AddFile()` at lines 26 through 45 integrates extracted docs into those indexes. That means the storage primitives for queryable documentation already exist.

The extraction layer is `go-go-goja/pkg/jsdoc/extract/extract.go`. It parses JavaScript files and extracts structured metadata from:

- `__package__`
- `__doc__`
- `__example__`
- `doc\`...\``

The relevant logic is:

- parse entrypoints at lines 21 through 87,
- call-expression dispatch at lines 159 through 206,
- tagged template prose attachment at lines 208 through 267,
- symbol and example parsing at lines 287 through 333.

This is exactly the authoring model needed for companion JS documentation files that live alongside the `smailnail` module.

#### The workspace also contains a query-first module pattern worth copying

The `go-go-goja/modules/glazehelp/glazehelp.go` module shows a good pattern for query-oriented docs access. Its loader exposes:

- `query`
- `section`
- `render`
- `topics`
- `keys`

at lines 23 through 123. The important design lesson is not the specific help-system API. It is the shape: expose structured lookup functions instead of a single opaque blob. `getIMAPJSDocumentation` should follow that philosophy.

#### There is already an export path if the documentation ever outgrows in-memory lookup

`go-go-goja/pkg/jsdoc/exportsq/exportsq.go` can write a `DocStore` into SQLite. The schema is created at lines 102 through 156 and includes tables for packages, symbols, symbol tags, symbol concepts, examples, and example-symbol links. This is not required for v1, but it matters because it shows the docs architecture can scale to richer offline indexing or future API/resource endpoints without redesigning the canonical format.

### Design Goals

The new MCP should satisfy these goals:

- expose only the two tools requested by the user,
- keep business logic out of the MCP handler layer,
- make documentation queryable instead of monolithic,
- keep prose and examples close to the JavaScript API surface,
- make documentation drift detectable,
- reuse existing libraries already present in the workspace,
- stay small enough for an intern to implement without hidden architecture debt.

Non-goals:

- building a general-purpose multi-tenant SaaS in the same step,
- exposing every possible IMAP operation in v1,
- inventing a brand-new documentation DSL when a good one already exists locally.

### Alternatives Considered

#### Alternative 1: static markdown only

This is the `jesus` model: store one or more markdown files, embed them, and return them from `getIMAPJSDocumentation`.

Advantages:

- minimal implementation effort,
- straightforward to render for humans,
- easy to boot quickly.

Disadvantages:

- weak queryability,
- hard to map docs to exact symbols and examples,
- poor fit for model/tool discovery,
- high drift risk because prose is disconnected from the exported module shape,
- hard to answer "show me examples for `connect`" without custom parsing layered on top later.

This should not be the default design. It optimizes for the first commit, not for maintainability.

#### Alternative 2: Go reflection only

This alternative would inspect Go module structures and exported bindings, then generate documentation from reflected signatures or descriptors.

Advantages:

- can help detect signature drift,
- keeps some metadata close to Go implementations,
- attractive if one wants a reusable system for many Goja modules.

Disadvantages:

- reflection alone cannot generate good prose,
- examples still need a home,
- Goja exports are closures registered through `SetExport`, which are not a great source of high-quality JS-facing docs,
- method signatures do not describe semantic constraints well,
- decoding helpers and lowerCamelCase JS shapes are not always obvious from reflected Go symbols.

Pure reflection is not sufficient. It should be used only as a verification layer, not as the canonical documentation source.

#### Alternative 3: JavaScript sentinel docs only

This alternative would author all documentation in JS companion files and extract them with `go-go-goja/pkg/jsdoc`.

Advantages:

- best prose and example locality,
- already supported by current tooling,
- naturally queryable via `DocStore`,
- maps well to how the JS API is actually consumed.

Disadvantages:

- documented symbols can still drift from actual exported Go bindings,
- requires author discipline,
- requires one more filesystem convention inside `smailnail`.

This is strong, but not quite enough by itself if we also care about drift detection.

#### Alternative 4: hybrid JS-doc canon plus Go-side verification

This is the recommended design.

Canonical docs are authored in JS sentinel files, extracted into a `DocStore`, and served by the documentation MCP tool. Separately, a Go-side validator checks that documented symbol names and exported runtime bindings stay aligned.

Advantages:

- good prose quality,
- rich examples,
- structured query support,
- low implementation risk because most primitives already exist,
- drift can be caught in tests instead of by humans at runtime.

Disadvantages:

- slightly more moving parts than a markdown-only approach,
- requires a small validation harness around exported symbol registration.

The trade-off is worth it. This is the best balance of practicality and rigor in the current workspace.

### Recommended architecture

Create a dedicated MCP binary in `smailnail` that exposes exactly two tools:

- `executeIMAPJS`
- `getIMAPJSDocumentation`

Use `go-go-mcp/pkg/embeddable` to register those tools, following the thin server shape proven by `jesus`, but do not copy the `jesus` documentation pattern.

Canonical IMAP JS API documentation should live in embedded JavaScript doc source files, extracted at startup into a `DocStore`.

```text
                   +-------------------------------+
                   |  smailnail IMAP JS MCP server |
                   +-------------------------------+
                         |                    |
                         |                    |
          +--------------+                    +------------------+
          |                                                     |
          v                                                     v
+-------------------------+                      +-----------------------------+
| executeIMAPJS handler   |                      | getIMAPJSDocumentation      |
|                         |                      | handler                      |
| - bind args             |                      | - bind query args            |
| - create runtime        |                      | - query DocStore             |
| - register smailnail    |                      | - shape structured result    |
| - eval JS               |                      +-----------------------------+
| - return result/error   |                                      |
+-------------------------+                                      |
          |                                                      |
          v                                                      v
+-------------------------+                      +-----------------------------+
| smailnail JS module     |                      | go-go-goja/pkg/jsdoc        |
| pkg/js/modules/...      |                      | - extract                    |
| pkg/services/...        |                      | - model                      |
+-------------------------+                      | - store                      |
                                                 +-----------------------------+
```

### Recommended file layout

The dedicated MCP should live in the `smailnail` repository, not in `go-go-mcp`. `go-go-mcp` should remain the framework dependency.

Recommended package layout:

```text
smailnail/
  cmd/
    smailnail-imap-mcp/
      main.go
  pkg/
    mcp/
      imapjs/
        server.go
        execute_tool.go
        docs_tool.go
        docs_registry.go
        docs_query.go
        docs_validation.go
        types.go
    js/
      modules/
        smailnail/
          module.go
          docs/
            package.js
            service.js
            connect-example.js
            build-rule-example.js
```

If more Goja-based systems need the same loader/query helpers later, the generic parts can be promoted into `go-go-goja/pkg/jsdoc/query` or `go-go-goja/pkg/jsdoc/registry`. That should happen only after the first implementation proves the abstractions. Right now, the safest path is to keep `smailnail`-specific query shaping local and only extract generic pieces that are obviously reusable.

### Tool 1: `executeIMAPJS`

This tool evaluates JavaScript against a runtime that exposes only the `smailnail` module and a minimal console capture facility.

The execution flow should be:

1. Bind and validate arguments with `embeddable.Arguments`.
2. Create a fresh Goja runtime per request.
3. Register the `smailnail` native module.
4. Optionally inject a connection descriptor or helper object if the API contract includes one.
5. Execute the submitted JavaScript.
6. Return:
   - success flag,
   - returned value,
   - captured console output,
   - error metadata if execution failed.

Suggested request shape:

```json
{
  "code": "const smailnail = require('smailnail'); ...",
  "timeoutMs": 30000,
  "captureConsole": true
}
```

Possible future extension, but not required for v1:

```json
{
  "connection": {
    "server": "imap.example.com",
    "port": 993,
    "username": "alice",
    "password": "secret",
    "mailbox": "INBOX",
    "insecure": false
  }
}
```

The tool should stay intentionally narrow. It should not add multiple execution modes, file loading, or persistent session storage in the first version.

Pseudocode:

```go
func executeIMAPJSHandler(ctx context.Context, raw map[string]interface{}) (*protocol.ToolResult, error) {
    args := embeddable.NewArguments(raw)

    var req ExecuteIMAPJSRequest
    if err := args.BindArguments(&req); err != nil {
        return errorResult(err), nil
    }
    if req.Code == "" {
        return errorResult(errors.New("code is required")), nil
    }

    rt := newRuntimeWithConsoleCapture()
    registerSmailnailModule(rt, smailnailmodule.NewModule())

    value, consoleLog, err := evalWithTimeout(ctx, rt, req.Code, req.TimeoutMS)
    if err != nil {
        return jsonResult(ExecuteIMAPJSResponse{
            Success: false,
            Error: serializeError(err),
            Console: consoleLog,
        }), nil
    }

    return jsonResult(ExecuteIMAPJSResponse{
        Success: true,
        Value: value,
        Console: consoleLog,
    }), nil
}
```

### Tool 2: `getIMAPJSDocumentation`

This tool is the real differentiator. It should provide structured queries against a loaded `DocStore`, not just one markdown string.

Recommended request contract:

```json
{
  "mode": "overview | package | symbol | concept | example | search | render",
  "package": "smailnail",
  "symbol": "connect",
  "concept": "imap-connection",
  "example": "connect-basic",
  "query": "mailbox connection",
  "limit": 10,
  "includeBody": true
}
```

Recommended response shape:

```json
{
  "mode": "symbol",
  "package": { "...": "..." },
  "symbols": [ { "...": "..." } ],
  "examples": [ { "...": "..." } ],
  "concepts": [ "imap-connection", "rule-building" ],
  "summary": "connect opens an IMAP session using ConnectOptions",
  "renderedMarkdown": "... optional when mode=render ..."
}
```

The internal query implementation should support:

- `overview`: package doc plus top symbols and top examples,
- `package`: exact package lookup,
- `symbol`: exact symbol lookup,
- `concept`: symbol and example lookup by concept tag,
- `example`: fetch one example including source body,
- `search`: simple substring search across names, summaries, prose, tags, and concepts,
- `render`: produce a single markdown page assembled from structured docs.

This approach gives both machines and humans what they need:

- models can ask focused symbol and example questions,
- humans can still render a full page when needed.

Pseudocode:

```go
func getIMAPJSDocumentationHandler(ctx context.Context, raw map[string]interface{}) (*protocol.ToolResult, error) {
    args := embeddable.NewArguments(raw)

    var req DocsQueryRequest
    if err := args.BindArguments(&req); err != nil {
        return errorResult(err), nil
    }

    result, err := docsRegistry.Query(req)
    if err != nil {
        return errorResult(err), nil
    }

    return jsonResult(result), nil
}
```

### Documentation source of truth

The canonical docs should be small JavaScript source files that use `go-go-goja/pkg/jsdoc` sentinel patterns.

Example sketch:

```javascript
__package__({
  name: "smailnail",
  title: "smailnail IMAP JavaScript API",
  category: "imap",
  description: "JavaScript API for building rules and opening IMAP sessions."
})

__doc__("connect", {
  summary: "Open an IMAP session using ConnectOptions.",
  concepts: ["imap-connection", "session-lifecycle"],
  params: [
    { name: "options", type: "ConnectOptions", description: "Server and mailbox settings." }
  ],
  returns: { type: "Session", description: "A live IMAP session wrapper." },
  related: ["newService", "close"],
  tags: ["network", "imap", "session"]
})

doc`
---
symbol: connect
---

Use \`connect\` when a script needs a live mailbox session.

Typical flow:
1. construct a service
2. call \`connect\`
3. work with the session
4. close it explicitly
`

__example__({
  id: "connect-basic",
  title: "Connect to INBOX",
  symbols: ["connect"],
  concepts: ["imap-connection"]
})
function connectBasic() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const session = service.connect({ server: "...", username: "...", password: "..." })
  session.close()
}
```

This is better than markdown-only docs because the docs are naturally broken down by symbol and example and can be indexed without custom parsing.

### Reflection and verification strategy

The user explicitly raised the possibility of "jsdocex on the go side of things, maybe with reflection there too, for go go goja scenarios." The recommended answer is:

- do not replace JS-authored docs with reflection,
- do add a Go-side verification layer that checks documentation completeness and symbol alignment.

Recommended verification checks:

1. Assert that every exported JS symbol has a corresponding `SymbolDoc`.
2. Assert that documented symbol names exist in the module export registry.
3. Assert that every example references real symbols.
4. Optionally assert that option field names mentioned in docs are real JSON field names from `BuildRuleOptions` and `ConnectOptions`.

The last point is especially valuable because `smailnailjs/service.go` is already the canonical source of JS-facing option keys like `withinDays`, `afterUid`, and `contentMaxLength`.

Pseudocode for a drift check:

```go
func TestSmailnailDocsMatchExports(t *testing.T) {
    exports := []string{
        "parseRule",
        "buildRule",
        "newService",
        "connect",
        "close",
    }

    store := loadSmailnailDocStore(t)
    for _, name := range exports {
        if _, ok := store.BySymbol[name]; !ok {
            t.Fatalf("missing documentation for symbol %q", name)
        }
    }
}
```

This is the correct role for Go-side reflection or descriptors: verification, not prose generation.

### Should a new generic Go-side `jsdocex` layer be created?

Not yet as a separate product name. The workspace already contains `go-go-goja/pkg/jsdoc`, and the CLI entrypoint already states that jsdocex has been migrated there. Creating another parallel package or tool would add naming confusion and more sync work.

What does make sense is a small amount of genericization inside `go-go-goja` after the first `smailnail` integration proves itself. Candidate reusable pieces:

- `pkg/jsdoc/registry` for embedded-FS loading into a `DocStore`,
- `pkg/jsdoc/query` for common lookups and simple search,
- `pkg/jsdoc/validate` for doc-to-export drift checks,
- possibly a lightweight `modules/jsdoc` Goja module if more runtimes need in-VM doc access.

That is a good follow-up direction because it builds on the migrated package instead of rebranding it again.

### API Reference Sketch

#### `executeIMAPJS`

Request:

- `code` `string` required
- `timeoutMs` `int` optional, default `30000`
- `captureConsole` `bool` optional, default `true`

Response:

- `success` `bool`
- `value` `any`
- `console` `[]string`
- `error` `object | null`

Error object:

- `message` `string`
- `kind` `string`
- `stack` `string | null`

#### `getIMAPJSDocumentation`

Request:

- `mode` `string` required
- `package` `string` optional
- `symbol` `string` optional
- `concept` `string` optional
- `example` `string` optional
- `query` `string` optional
- `limit` `int` optional
- `includeBody` `bool` optional

Response:

- `mode` `string`
- `package` `object | null`
- `symbols` `[]object`
- `examples` `[]object`
- `concepts` `[]string`
- `summary` `string`
- `renderedMarkdown` `string | null`

### Detailed Implementation Plan

#### Phase 1: create the dedicated MCP host

Add a new command entrypoint in `smailnail/cmd/` that wires `go-go-mcp/pkg/embeddable` and registers only the two tools.

Files to add:

- `smailnail/cmd/smailnail-imap-mcp/main.go`
- `smailnail/pkg/mcp/imapjs/server.go`
- `smailnail/pkg/mcp/imapjs/types.go`

Acceptance criteria:

- binary starts,
- advertises only the two tools,
- stdio transport works first.

#### Phase 2: implement `executeIMAPJS`

Add the execution handler and runtime bootstrapping.

Files to add:

- `smailnail/pkg/mcp/imapjs/execute_tool.go`

Files to reuse:

- `smailnail/pkg/js/modules/smailnail/module.go`
- `smailnail/pkg/services/smailnailjs/service.go`

Acceptance criteria:

- JavaScript can `require("smailnail")`,
- `newService().buildRule(...)` works,
- errors are serialized cleanly.

#### Phase 3: author structured docs and load them

Add JS sentinel documentation files and a doc registry loader.

Files to add:

- `smailnail/pkg/js/modules/smailnail/docs/package.js`
- `smailnail/pkg/js/modules/smailnail/docs/service.js`
- `smailnail/pkg/mcp/imapjs/docs_registry.go`
- `smailnail/pkg/mcp/imapjs/docs_query.go`

Acceptance criteria:

- docs load from embedded FS,
- `DocStore` contains package, symbol, and example entries,
- one command or test can print a package overview.

#### Phase 4: implement `getIMAPJSDocumentation`

Add the query handler that exposes exact lookups and search.

Files to add:

- `smailnail/pkg/mcp/imapjs/docs_tool.go`

Acceptance criteria:

- symbol lookup works,
- concept lookup works,
- render mode assembles a readable markdown page.

#### Phase 5: add drift validation

Add tests that compare documented symbols to exported module surfaces and example references.

Files to add:

- `smailnail/pkg/mcp/imapjs/docs_validation.go`
- `smailnail/pkg/mcp/imapjs/docs_validation_test.go`

Acceptance criteria:

- missing symbol docs fail tests,
- bad example references fail tests.

#### Phase 6: transport and smoke validation

Validate the new MCP over the same transport set already used elsewhere:

- `command`
- `sse`
- `streamable_http`

Recommended smoke artifacts:

- repo-maintained smoke script under `smailnail/scripts/`,
- ticket-local experiment scripts while iterating,
- end-to-end calls against both tools.

### Testing and Validation Strategy

#### Unit tests

- doc registry loads embedded files correctly,
- query modes return expected symbol/example sets,
- simple search matches names and prose,
- drift checks catch missing docs.

#### Integration tests

- Goja runtime can `require("smailnail")`,
- `executeIMAPJS` runs a small rule-building script,
- `getIMAPJSDocumentation` returns structured data for `symbol=connect`.

#### Smoke tests

- start the MCP over stdio,
- call `getIMAPJSDocumentation` for overview and symbol modes,
- call `executeIMAPJS` with a no-network script first,
- then optionally with Docker-backed IMAP validation if credentials or test fixtures are available.

#### Documentation review checks

- every top-level symbol has prose, params, and returns,
- every example names at least one symbol,
- every concept is spelled consistently.

### Risks

#### Risk: documentation drift

Mitigation: add doc validation tests and keep canonical docs close to module sources.

#### Risk: overly broad v1

Mitigation: enforce the two-tool boundary and reject extra execution modes in the first implementation.

#### Risk: abstraction premature extraction into `go-go-goja`

Mitigation: build the `smailnail`-specific layer first, then promote only clearly reusable query/registry helpers.

#### Risk: weak search semantics in the first documentation tool

Mitigation: start with exact lookup plus small substring search. Do not block v1 on a full-text indexing engine.

### Open Questions

1. Should `executeIMAPJS` accept raw connection credentials in v1, or only operate on scripts that construct connections manually through the module API?
2. Should `getIMAPJSDocumentation` include rendered markdown in every response, or only when `mode=render`?
3. Should a future `go-go-goja` generic query package be introduced in the same implementation ticket or deferred until after the first `smailnail` integration lands?

### Recommendation

Proceed with a dedicated `smailnail` MCP that exposes only the two requested tools and uses a hybrid documentation architecture:

- runtime behavior stays in `smailnailjs.Service` and the native module,
- docs are authored in JS sentinel files and loaded through `go-go-goja/pkg/jsdoc`,
- the documentation tool exposes structured queries,
- Go-side validation keeps symbol documentation aligned with exports.

That gives the user exactly the MCP they asked for while building a documentation system that is reusable in other Goja-based projects without forcing premature framework extraction.

## Design Decisions

The most important design decisions are:

- host the new MCP in `smailnail`, not in `go-go-mcp`,
- expose only two tools in v1,
- use JS sentinel docs as the canonical prose source,
- use Go-side validation as a guardrail rather than as the primary doc generator,
- defer generic `go-go-goja` extraction until the first concrete integration exists.

## Alternatives Considered

See the dedicated alternatives analysis above. The short version is that static markdown is too weak and pure reflection is too shallow.

## Implementation Plan

See the phased plan under "Detailed Implementation Plan" above.

## Open Questions

See the "Open Questions" section above.

## References

- `smailnail/pkg/js/modules/smailnail/module.go`
- `smailnail/pkg/services/smailnailjs/service.go`
- `jesus/pkg/mcp/server.go`
- `jesus/pkg/doc/doc.go`
- `go-go-goja/cmd/goja-jsdoc/main.go`
- `go-go-goja/pkg/jsdoc/model/model.go`
- `go-go-goja/pkg/jsdoc/model/store.go`
- `go-go-goja/pkg/jsdoc/extract/extract.go`
- `go-go-goja/pkg/jsdoc/exportsq/exportsq.go`
- `go-go-goja/modules/glazehelp/glazehelp.go`
- `go-go-mcp/pkg/embeddable/arguments.go`
