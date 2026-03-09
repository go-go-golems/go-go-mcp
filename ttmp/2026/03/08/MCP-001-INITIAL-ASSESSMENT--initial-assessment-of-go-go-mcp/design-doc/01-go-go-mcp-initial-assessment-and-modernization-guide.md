---
Title: go-go-mcp initial assessment and modernization guide
Ticket: MCP-001-INITIAL-ASSESSMENT
Status: active
Topics:
    - mcp
    - go
    - assessment
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-mcp/README.md
      Note: Documentation drift around bridge and prompts/resources claims
    - Path: go-go-mcp/cmd/go-go-mcp/cmds/client/helpers/client.go
      Note: Canonical CLI client path using mcp-go
    - Path: go-go-mcp/cmd/go-go-mcp/cmds/server/start.go
      Note: Live server startup path and tool-only wiring evidence
    - Path: go-go-mcp/cmd/go-go-mcp/main.go
      Note: Root CLI composition and removed bridge evidence
    - Path: go-go-mcp/pkg/auth/oidc/server.go
      Note: Embedded OIDC provider and protected-resource metadata behavior
    - Path: go-go-mcp/pkg/client/client.go
      Note: Legacy in-house client still fixed at protocol version 2024-11-05
    - Path: go-go-mcp/pkg/embeddable/mcpgo_backend.go
      Note: Canonical mcp-go backend and transport/OIDC wiring
    - Path: go.work
      Note: Workspace override causing incompatible sibling module resolution
ExternalSources: []
Summary: Evidence-backed assessment of go-go-mcp covering current architecture, standalone vs workspace build health, transport and OIDC validation, documentation drift, and a phased modernization plan for future maintainers.
LastUpdated: 2026-03-08T17:51:20.32923992-04:00
WhatFor: Understand what go-go-mcp currently is, what still works, what is stale or misleading, and what should be fixed first before further product work.
WhenToUse: Use this when onboarding to go-go-mcp, planning a cleanup/refactor, or deciding whether to invest in the current CLI and embeddable MCP surface.
---


# go-go-mcp initial assessment and modernization guide

## Executive Summary

`go-go-mcp` is not a dead codebase, but it is not in a clean or trustworthy state either. The core embeddable/server/client path still works when the module is evaluated on its own with `GOWORK=off`: `go test ./...` passes, the `go-go-mcp` CLI can list and call tools over `command`, `sse`, and `streamable_http`, and the embedded OIDC layer can protect HTTP transports. The project therefore still has usable engineering value.

The main problem is drift. The checked-out multi-module workspace currently breaks the repo because the local `glazed/` checkout is missing packages that `go-go-mcp` imports, so `go test ./...` and `go build ./...` fail when `go.work` is active. The documentation also overstates what the repo does today: the README still advertises a removed bridge command, prompt/resource support is documented more broadly than the runtime wiring actually provides, and the codebase still contains an older internal client/protocol stack beside the newer `mcp-go`-backed path.

The immediate recommendation is not a rewrite. The immediate recommendation is to stabilize the working path that already exists: make the repo self-contained, declare the `mcp-go`-backed embeddable/CLI flow the canonical runtime, remove or quarantine stale surfaces, and add transport-level integration tests so future refactors do not recreate the same ambiguity.

## Problem Statement

The user asked for a new-ticket assessment of how out of date `go-go-mcp` is, whether it still works, what is good, what is bad, and what should be addressed now. This document answers that by inspecting the live repository, running builds/tests/smoke experiments, and comparing the implementation to current MCP specification direction.

This assessment is intentionally evidence-first. It distinguishes:

1. What code is actually exercised by the current CLI and embeddable examples.
2. What code still exists but appears partially abandoned or superseded.
3. What documentation claims are no longer accurate.
4. What modernization work is urgent versus optional.

## Scope

Included in scope:

1. `cmd/go-go-mcp` CLI entrypoints and server/client flows.
2. `pkg/embeddable` as the main reusable integration surface.
3. Tool/provider/config loading for the server.
4. `pkg/auth/oidc` and HTTP transport protection.
5. Local legacy protocol/client packages insofar as they affect maintainability.
6. Repository health under both workspace and standalone-module execution.

Excluded from deep implementation review:

1. The scholarly application internals, except where they reveal repository health.
2. UI/TUI affordances beyond noting their presence and current role.
3. Feature design for new MCP capabilities not already hinted at by the current architecture.

## Validation Snapshot

### Commands run

Repository-health experiments:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp
go test ./...
go build ./...
env GOWORK=off go test ./...
env GOWORK=off go build ./cmd/go-go-mcp
```

Transport smoke tests:

```bash
env GOWORK=off go run ./pkg/embeddable/examples/basic mcp list-tools
env GOWORK=off go run ./pkg/embeddable/examples/struct mcp list-tools
env GOWORK=off go run ./pkg/embeddable/examples/enhanced mcp list-tools

env GOWORK=off go run ./pkg/embeddable/examples/basic mcp start --transport sse --port 4010
env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport sse --server http://localhost:4010/mcp/sse

env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport command \
  --command env --command GOWORK=off --command go --command run \
  --command ./pkg/embeddable/examples/basic --command mcp --command start

env GOWORK=off go run ./cmd/go-go-mcp client tools call greet --transport command \
  --command env --command GOWORK=off --command go --command run \
  --command ./pkg/embeddable/examples/basic --command mcp --command start \
  --json '{"name":"Intern"}'

env GOWORK=off go run ./pkg/embeddable/examples/basic mcp start --transport streamable_http --port 4012
env GOWORK=off go run ./cmd/go-go-mcp client tools list --transport streamable_http --server http://localhost:4012/mcp
```

OIDC smoke tests:

```bash
env GOWORK=off go run ./pkg/embeddable/examples/oidc mcp start --transport sse --port 4011
curl -i -s http://localhost:4011/.well-known/openid-configuration
curl -i -s http://localhost:4011/.well-known/oauth-protected-resource
curl -i -s http://localhost:4011/mcp/sse
curl -i -s --max-time 2 -H 'Authorization: Bearer TEST_AUTH_KEY_123' http://localhost:4011/mcp/sse
```

### Results

1. With the workspace active, `go test ./...` and `go build ./...` fail immediately due to unresolved `glazed/pkg/cmds/layers`, `parameters`, and `middlewares`.
2. With `GOWORK=off`, `go test ./...` passes and `go build ./cmd/go-go-mcp` succeeds.
3. The embeddable examples successfully list tools.
4. The CLI client successfully lists tools over `command`, `sse`, and `streamable_http`.
5. A real tool call succeeded over command transport and returned `Hello, Intern!`.
6. OIDC discovery, protected-resource metadata, `401` with `WWW-Authenticate`, and authorized SSE endpoint access all worked.
7. The OIDC example reveals a bug: discovery metadata still advertises `http://localhost:3001` after launching on port `4011`.

## Current-State Architecture

### 1. Top-level shape

The repository currently contains three overlapping layers:

1. A user-facing CLI in `cmd/go-go-mcp`.
2. A reusable embeddable MCP integration layer in `pkg/embeddable`.
3. Older internal MCP protocol/client abstractions in `pkg/protocol` and `pkg/client`.

The active runtime path is not the older internal path. The current server path delegates transport work to `github.com/mark3labs/mcp-go`, while the current CLI client helpers also use `mcp-go`.

### 2. Entry points and ownership boundaries

Key entry points:

1. `cmd/go-go-mcp/main.go`
   - Builds the root Cobra command.
   - Adds client, server, config, cursor, Claude, UI, and OIDC admin commands.
   - Explicitly comments that the bridge command was removed.
2. `cmd/go-go-mcp/cmds/server/start.go`
   - Builds tool providers from config/directories.
   - Creates an embeddable backend.
   - Starts file watching and the selected transport in an `errgroup`.
3. `pkg/embeddable/mcpgo_backend.go`
   - Converts the local tool registry into `mcp-go` tool registrations.
   - Starts `stdio`, `sse`, or `streamable_http`.
   - Adds OIDC protection to the HTTP transports when enabled.
4. `cmd/go-go-mcp/cmds/client/helpers/client.go`
   - Uses `mcp-go` clients for `sse`, `streamable_http`, and `command`.
   - Initializes with `mcp.LATEST_PROTOCOL_VERSION`.

### 3. Runtime flow

Server startup flow:

```text
cmd/go-go-mcp main
  -> server start command
     -> server layer parses config/directories/files/internal servers
        -> ConfigToolProvider loads Glazed commands and shell-command tools
           -> tool-registry proxy is built
              -> embeddable.NewBackend
                 -> mcp-go MCPServer
                 -> register tool handlers
                 -> start stdio / sse / streamable_http transport
                 -> optional OIDC wrapping for HTTP transports
```

Tool call flow:

```text
MCP client request
  -> mcp-go transport/server
     -> embeddable tool adapter
        -> cfg hooks/middleware
           -> local tool registry
              -> ConfigToolProvider or explicit tool handler
                 -> Glazed command or Go handler executes
                    -> local protocol.ToolResult
                       -> mapped back into mcp-go CallToolResult
```

OIDC flow for HTTP transports:

```text
HTTP request to /mcp or /mcp/*
  -> oidcAuthMiddleware
     -> allow public metadata/auth endpoints
     -> require Bearer token otherwise
     -> on failure: 401 + WWW-Authenticate + protected-resource metadata hints
     -> on success: pass through to mcp-go SSE/streamable_http handler
```

### 4. What is implemented versus what exists on the shelf

Implemented and exercised now:

1. Tool registration and tool calling.
2. `command`, `sse`, and `streamable_http` client/server flows.
3. Config-driven tool loading for server-side tools.
4. Embedded OIDC for HTTP transports.
5. Embeddable examples covering basic, enhanced, struct, session, and OIDC server setup.

Present in the repository but not convincingly part of the current runtime:

1. Local `pkg/client` transport/client stack.
2. Local `pkg/protocol` types as a standalone protocol implementation.
3. Resource and prompt registries as general runtime capabilities.
4. Bridge-related documentation and help content.

## Evidence-Based Findings

### Finding 1: The repository is currently broken in its checked-out workspace form

Severity: high

Observed behavior:

1. `go test ./...` fails under the active `go.work`.
2. `go build ./...` fails the same way.
3. `go.work` pulls in a sibling `./glazed` checkout.
4. That local `glazed/` tree is missing `pkg/cmds/layers`, `pkg/cmds/parameters`, and `pkg/cmds/middlewares`, which `go-go-mcp` imports.

Why this matters:

1. A contributor cannot trust a plain checkout.
2. CI or local development will behave differently depending on whether `GOWORK` is set.
3. The repo has hidden environmental coupling.

Evidence:

1. `go.work` includes `./glazed`, `./go-go-mcp`, and other sibling modules.
2. `glazed/pkg/cmds` in the local checkout lacks the required subdirectories.
3. The same imports do resolve with `GOWORK=off`.

### Finding 2: The standalone module still works and is worth preserving

Severity: high, but positive

Observed behavior:

1. `env GOWORK=off go test ./...` passed.
2. `env GOWORK=off go build ./cmd/go-go-mcp` succeeded.
3. The basic/struct/enhanced embeddable examples all listed tools successfully.
4. The CLI client successfully listed tools over `command`, `sse`, and `streamable_http`.
5. The CLI client successfully called `greet` over command transport.

Why this matters:

1. The codebase is not obsolete in the sense of “non-functional”.
2. There is a stable-enough core to modernize rather than replace.
3. The current public value is concentrated in the embeddable layer plus the `mcp-go`-backed CLI.

### Finding 3: Documentation drift is severe enough to mislead a new contributor

Severity: high

Examples of drift:

1. The README still advertises a bridge mode, but `cmd/go-go-mcp/main.go` says `// bridge command removed`.
2. The README still broadly presents prompts/resources support, but the current server startup path only wires tools.
3. Some internal docs still describe earlier transport paths or planned refactors as if they are the next step, even though the refactor partially happened.

Why this matters:

1. A new engineer will infer capabilities from docs that are not present in the live runtime.
2. It increases the cost of every future change because the first task is re-deriving reality.

### Finding 4: The current server runtime is effectively tool-only

Severity: high

Observed behavior:

1. `cmd/go-go-mcp/cmds/server/start.go` builds a tool provider and a proxy tool registry.
2. It creates a resource registry only as `_ = resources.NewRegistry()`.
3. `pkg/embeddable/mcpgo_backend.go` enables tool capabilities and registers only tools into `mcp-go`.

Impact:

1. Prompt and resource registries exist in the repo, but they are not first-class in the current server runtime path.
2. README claims about prompts/resources should not be trusted without additional wiring.
3. The project needs an explicit decision: either wire these capabilities for real or document the server as tool-centric.

### Finding 5: The codebase contains two MCP client/protocol worlds

Severity: medium-high

Observed behavior:

1. `cmd/go-go-mcp/cmds/client/helpers/client.go` uses `mcp-go` and negotiates `mcp.LATEST_PROTOCOL_VERSION`.
2. `pkg/client/client.go` is a separate internal client stack with a hard-coded `"2024-11-05"` protocol version.
3. `pkg/client/sse.go` and `pkg/client/stdio.go` implement their own transports.

Impact:

1. The repo pays maintenance cost for code paths that are no longer the primary client runtime.
2. A contributor can easily update the wrong client stack.
3. Compatibility stories become ambiguous.

### Finding 6: The `mcp-go` migration is real but incomplete

Severity: medium

Observed behavior:

1. The active embeddable backend already runs on `mcp-go`.
2. The active CLI client already uses `mcp-go`.
3. Local protocol/result types are still part of handler signatures and adapters.
4. Stale docs and local client code still suggest a not-fully-finished migration.

Interpretation:

This repo is not “pre-migration”; it is “mid-consolidation”. The right framing for future work is consolidation and cleanup, not greenfield design.

### Finding 7: The OIDC layer is interesting and mostly functional, but the example is wrong in an important way

Severity: medium

Observed behavior:

1. OIDC discovery endpoints, authorization-server metadata, protected-resource metadata, and `WWW-Authenticate` hints are implemented.
2. Unauthorized `GET /mcp/sse` correctly returns `401`.
3. Authorized `GET /mcp/sse` with the static auth key returns an SSE endpoint event.
4. The example hardcodes `Issuer: "http://localhost:3001"` while allowing `--port` overrides; when launched on `4011`, discovery still advertises `3001`.

Impact:

1. The core auth machinery is promising.
2. The shipped example can mislead anyone trying to validate the flow on a non-default port.
3. Example correctness matters disproportionately because this is the area most likely to be tested by humans before reading code.

### Finding 8: Several runtime surfaces still panic instead of failing cleanly

Severity: medium

Examples:

1. `pkg/tools/tool.go` panics when `ToolImpl.Call` is reached without an override.
2. `pkg/tools/providers/config-provider/tool-provider.go` panics for `BareCommand`, `GlazeCommand`, and unknown command types.

Impact:

1. This is acceptable in an internal prototype.
2. It is not acceptable in a CLI/framework that aims to be an integration platform.
3. The panic paths are exactly the kind of failure that a new contributor will only discover late.

### Finding 9: Test coverage exists for only a small fraction of the interesting behavior

Severity: medium

Observed behavior:

1. `go test ./...` passes standalone.
2. Only a handful of packages actually contain tests.
3. The transport matrix, OIDC behavior, and config-provider execution path are primarily untested in automated form.

Impact:

1. The repo can regress silently while still appearing “healthy”.
2. Real functionality is protected mostly by manual knowledge rather than CI.

## What Is Good

### 1. The embeddable façade is the right product shape

For a Go-heavy ecosystem, `pkg/embeddable` is the most convincing part of the project. It gives another Cobra application a concise way to expose an MCP surface without re-implementing the server.

Why it is good:

1. It keeps integration logic near the host app.
2. It supports multiple tool registration styles.
3. It now rides on a more standard MCP runtime (`mcp-go`) instead of a fully homegrown transport stack.

### 2. The transport story is stronger than the repo health suggests

The validated transport matrix is materially useful:

1. `command` works.
2. `sse` works.
3. `streamable_http` works.

That means the project already supports the transport modes that matter most for practical MCP interoperability today.

### 3. Config-driven tool loading is still a meaningful differentiator

The config-provider pattern lets the server expose tools backed by Glazed commands and shell-command YAML definitions. That is still valuable and worth preserving if the project continues.

### 4. The auth work is directionally aligned with modern MCP needs

Even though the auth layer needs cleanup, it is not cargo-culted. It exposes protected-resource metadata and uses `WWW-Authenticate` hints, which is the right direction for contemporary MCP HTTP deployments.

## What Is Bad

### 1. The repo does not currently have a single trustworthy source of truth

Docs, code, examples, and workspace behavior disagree with one another. That is the central quality problem.

### 2. Legacy and current implementation layers are interleaved

The codebase never cleanly severed the old internal protocol/client world from the newer `mcp-go`-based world. That makes every architecture discussion noisy.

### 3. The project overpromises capabilities

Prompts, resources, bridge behavior, and some workflow docs appear more mature in documentation than they are in the live runtime.

### 4. Contributor ergonomics are poor

A new engineer can do the “correct” thing (`go test ./...`) and immediately receive a false-negative signal because of workspace coupling rather than `go-go-mcp` itself.

## Comparison Against Current MCP Direction

Official MCP sources show the ecosystem has continued moving after the repo's older local protocol types were written:

1. The repo's legacy local client still initializes with protocol version `2024-11-05`.
2. `mcp-go` v0.38.0, which the current CLI client uses, advertises `LATEST_PROTOCOL_VERSION = "2025-06-18"`.
3. Official MCP transport and authorization docs now include newer revisions, including 2025-11-25 transport/auth material.

Practical implication:

1. The canonical runtime path in this repo is closer to current MCP than the legacy local packages are.
2. The worst “out of date” part is not the `mcp-go`-backed runtime path.
3. The worst “out of date” part is the leftover local implementation and doc surface that still frames the repo around older assumptions.

## Detailed Intern Guide

### Start here

If you are new to this repo, read it in this order:

1. `cmd/go-go-mcp/main.go`
2. `cmd/go-go-mcp/cmds/server/start.go`
3. `pkg/embeddable/server.go`
4. `pkg/embeddable/mcpgo_backend.go`
5. `cmd/go-go-mcp/cmds/client/helpers/client.go`
6. `pkg/tools/providers/config-provider/tool-provider.go`
7. `pkg/auth/oidc/server.go`

This order mirrors how a real request gets from CLI to server to tool execution.

### Core concepts

#### Concept 1: There are two layers of abstraction

The repo separates:

1. MCP transport/runtime concerns.
2. Tool-definition and tool-execution concerns.

The active runtime is `mcp-go`. The tool-definition side is still custom to this repo.

#### Concept 2: Tools are the only fully wired server capability today

There are registries for prompts and resources, but the active server path is centered on tools.

#### Concept 3: `embeddable` is the product; `pkg/client` is legacy baggage

If you are extending the server or integrating MCP into another app, work from `pkg/embeddable`.

If you are maintaining the CLI client, work from `cmd/go-go-mcp/cmds/client/*`.

Do not start new work in `pkg/client` unless you intentionally plan to resurrect that internal stack.

### Key files and what they mean

`cmd/go-go-mcp/main.go`

1. Root CLI composition.
2. Good place to verify which commands actually exist.
3. Important because it explicitly tells you the bridge command is gone.

`cmd/go-go-mcp/cmds/server/start.go`

1. Main server wiring for the CLI.
2. Shows how config, watching, internal servers, and the backend are combined.
3. Also reveals that resources are currently created and discarded, not wired.

`pkg/embeddable/server.go`

1. Defines `ServerConfig`.
2. Defines server options such as `WithTool`, `WithMiddleware`, `WithHooks`, and `WithOIDC`.
3. This file tells you what the library wants its public API to be.

`pkg/embeddable/mcpgo_backend.go`

1. The actual current runtime engine.
2. Registers tools into `mcp-go`.
3. Starts `stdio`, `sse`, or `streamable_http`.
4. Adds auth middleware for HTTP transports.

`cmd/go-go-mcp/cmds/client/helpers/client.go`

1. The active CLI client creation logic.
2. Uses `mcp-go` clients for transport handling.
3. Negotiates the latest protocol version exposed by the dependency.

`pkg/client/*`

1. Older in-house client/transport implementation.
2. Useful mostly as historical context or as a source of ideas.
3. Dangerous if mistaken for the mainline runtime.

`pkg/tools/providers/config-provider/tool-provider.go`

1. Bridges config files/directories and Glazed commands into MCP tools.
2. This is where the "MCP as a wrapper around YAML/shell/Glazed commands" story really lives.
3. Contains panic paths that should eventually be converted into real errors.

`pkg/auth/oidc/server.go`

1. Embedded OAuth/OIDC provider.
2. Protected-resource metadata and discovery support live here.
3. High-value, high-complexity area.

### Pseudocode for the active server startup path

```go
func serverStart() {
    settings := parseServerFlags()

    provider := CreateToolProvider(settings)
    reg := newRegistry()

    for tool in provider.ListTools() {
        reg.RegisterToolWithHandler(tool, func(args) {
            return provider.CallTool(tool.Name, args)
        })
    }

    cfg := embeddable.NewServerConfig()
    cfg.toolRegistry = reg
    cfg.transport = settings.transport
    cfg.port = settings.port
    cfg.oidc = settings.oidc

    backend := embeddable.NewBackend(cfg)

    runFileWatcherInErrgroup(provider)
    runBackendInErrgroup(backend)
}
```

### Pseudocode for the active tool call path

```go
func mcpGoHandler(ctx, toolName, requestArgs) {
    if cfg.hooks.BeforeToolCall != nil {
        cfg.hooks.BeforeToolCall(ctx, toolName, requestArgs)
    }

    result, err := wrappedRegistryHandler(ctx, requestArgs)

    mcpResult := mapLocalToolResultToMCPGo(result)

    if cfg.hooks.AfterToolCall != nil {
        cfg.hooks.AfterToolCall(ctx, toolName, result, err)
    }

    return mcpResult, err
}
```

### Pseudocode for HTTP auth behavior

```go
func oidcAuthMiddleware(req) {
    if req.path is public_metadata_or_oauth_endpoint {
        allow()
        return
    }

    token := extractBearerToken(req)
    if token missing {
        return 401 with WWW-Authenticate metadata hints
    }

    if token == static_dev_key {
        allow()
        return
    }

    if oidc.IntrospectAccessToken(token) succeeds {
        allow()
        return
    }

    return 401 with WWW-Authenticate metadata hints
}
```

### Architecture diagrams

Server-side diagram:

```text
                           +-------------------+
                           |   Cobra CLI       |
                           | cmd/go-go-mcp     |
                           +---------+---------+
                                     |
                                     v
                           +-------------------+
                           | server/start      |
                           | parse flags       |
                           +---------+---------+
                                     |
                                     v
                           +-------------------+
                           | ConfigToolProvider|
                           | load commands     |
                           +---------+---------+
                                     |
                                     v
                           +-------------------+
                           | tool-registry     |
                           | proxy handlers    |
                           +---------+---------+
                                     |
                                     v
                           +-------------------+
                           | embeddable        |
                           | mcpgo_backend     |
                           +----+--------+-----+
                                |        |
                        tools ->|        |-> stdio / sse / streamable_http
                                |        |
                                v        v
                           +-------------------+
                           |   mcp-go server   |
                           +-------------------+
```

Client-side diagram:

```text
user command
  -> cmd/go-go-mcp client ...
     -> client/helpers.CreateClient
        -> mcp-go client transport
           -> initialize
              -> list tools / call tool
```

OIDC diagram:

```text
HTTP client
  -> /mcp or /mcp/sse
     -> oidcAuthMiddleware
        -> allow public OIDC metadata routes
        -> require Bearer token for MCP routes
        -> pass request to mcp-go server on success
```

## What To Address Now

### Priority 0: Make the repo self-contained and deterministic

Do this first.

Actions:

1. Decide whether `go-go-mcp` is expected to build inside this `go.work`.
2. If yes, fix the sibling `glazed` checkout or pin the workspace to a compatible commit.
3. If no, remove the hidden assumption by documenting and enforcing `GOWORK=off` in validation scripts and CI.
4. Add CI that runs `env GOWORK=off go test ./...` at minimum.

Success criteria:

1. A fresh checkout has one obvious validation command.
2. CI and local developer guidance agree.

### Priority 1: Declare the canonical runtime surface

Recommended choice:

1. Canonical server runtime: `pkg/embeddable` + `mcp-go` backend.
2. Canonical CLI client: `cmd/go-go-mcp/cmds/client/*`.
3. Legacy/internal surface: `pkg/client` and any older protocol-only abstractions.

Actions:

1. Mark legacy packages as deprecated in docs.
2. Stop mentioning them as the main implementation path.
3. Move new work only into the canonical path.

### Priority 2: Reconcile docs with the actual runtime

Actions:

1. Remove or rewrite README bridge sections.
2. Rewrite capability claims to say tool support is real, while prompts/resources are partial or not wired.
3. Update transport examples to match `/mcp/sse` and `/mcp`.
4. Add a small validation section that mirrors the smoke scripts in this ticket.

### Priority 3: Fix example correctness bugs

Actions:

1. Make the OIDC example derive issuer from the effective port or flag value.
2. Review example defaults for transport paths and auth expectations.
3. Ensure example code is treated as tested product surface, not disposable sample text.

### Priority 4: Replace panics with typed errors in framework code

Actions:

1. `ToolImpl.Call` should not be reachable as a hard panic in a user-facing framework.
2. Config-provider execution should reject unsupported command shapes with returned errors.
3. Convert contributor footguns into debuggable failures.

### Priority 5: Decide whether prompts/resources are product goals

Decision point:

1. If prompts/resources are core goals, wire them fully into the server runtime and test them.
2. If not, remove their prominence from the public narrative.

The current middle ground is the worst state because it advertises optionality but guarantees nothing.

## Phased Modernization Plan

### Phase 1: Stabilize and document reality

1. Fix workspace-vs-module ambiguity.
2. Update README and help docs.
3. Add smoke scripts and CI entrypoints.
4. Fix the OIDC example issuer bug.

### Phase 2: Remove architectural ambiguity

1. Deprecate or archive `pkg/client` if it will not remain supported.
2. Clarify whether `pkg/protocol` remains only as an internal result-shape library or should be reduced further.
3. Delete references to removed bridge behavior.

### Phase 3: Add automated transport coverage

1. Test `command`, `sse`, and `streamable_http`.
2. Test one real tool call in each transport.
3. Test OIDC unauthenticated and authenticated HTTP access.

### Phase 4: Either wire or prune optional capabilities

1. Wire prompts/resources into the live runtime if they matter.
2. Otherwise reduce public complexity and remove the false surface area.

## Testing and Validation Strategy

Recommended permanent validation matrix:

1. `env GOWORK=off go test ./...`
2. `env GOWORK=off go build ./cmd/go-go-mcp`
3. Embeddable example smoke:
   - basic `mcp list-tools`
   - struct `mcp list-tools`
   - enhanced `mcp list-tools`
4. Transport smoke:
   - CLI client to basic example over `command`
   - CLI client to basic example over `sse`
   - CLI client to basic example over `streamable_http`
5. OIDC smoke:
   - discovery endpoints
   - `401` on unauthenticated MCP route
   - successful authenticated SSE handshake

## Risks

### Risk 1: Cleaning up legacy code may break undocumented internal consumers

Mitigation:

1. Deprecate first.
2. Add smoke tests before deleting.
3. Publish a migration note inside the repo.

### Risk 2: Documentation cleanup may expose how much is only partial

Mitigation:

That is a worthwhile short-term pain. The current ambiguity is worse than explicit reduction.

### Risk 3: OIDC cleanup could accidentally regress the currently working HTTP auth path

Mitigation:

1. Preserve the current curl-based smoke procedure.
2. Automate it before larger auth refactors.

## Alternatives Considered

### Alternative 1: Rewrite around a fresh MCP library and ignore the current code

Rejected for now.

Reason:

1. The embeddable layer already works.
2. The current transport story is usable.
3. The bigger problem is consolidation, not missing primitives.

### Alternative 2: Keep everything and only add new features

Rejected.

Reason:

1. The current ambiguity would compound.
2. Every new feature would have to navigate duplicated runtime stories.

### Alternative 3: Freeze the repo as historical and start a new one

Possible, but not the best immediate move.

Reason:

1. The current repo still has a functional core.
2. A stabilization pass is cheaper than starting over unless strategic goals have changed dramatically.

## Open Questions

1. Is `go-go-mcp` expected to live inside this workspace permanently, or should it be independently buildable first and foremost?
2. Do prompts/resources still matter as product goals, or is the repo now intentionally tool-first?
3. Should `pkg/client` survive as a supported API, or should it be deprecated immediately?
4. Is the long-term ambition to keep `go-go-mcp` as a distinct framework around `mcp-go`, or to thin it down into mostly config/provider affordances?

## References

### Key repository files

1. `cmd/go-go-mcp/main.go`
2. `cmd/go-go-mcp/cmds/server/start.go`
3. `cmd/go-go-mcp/cmds/server/layers/server.go`
4. `cmd/go-go-mcp/cmds/client/helpers/client.go`
5. `cmd/go-go-mcp/cmds/client/tools.go`
6. `cmd/go-go-mcp/cmds/server/tools.go`
7. `pkg/embeddable/server.go`
8. `pkg/embeddable/command.go`
9. `pkg/embeddable/mcpgo_backend.go`
10. `pkg/auth/oidc/server.go`
11. `pkg/tools/providers/config-provider/tool-provider.go`
12. `pkg/tools/providers/tool-registry/registry.go`
13. `pkg/client/client.go`
14. `pkg/client/sse.go`
15. `pkg/client/stdio.go`
16. `pkg/protocol/initialization.go`
17. `pkg/resources/registry.go`
18. `pkg/prompts/registry.go`
19. `README.md`
20. `go.work`

### Existing internal design notes that still matter

1. `ttmp/2025-08-23/05-decoupling-mcp-go-in-go-go-mcp-refactoring-plan-and-approach.md`
2. `ttmp/2025-08-23/08-top-level-integration-testing-strategy.md`

### External sources

1. MCP transport specification: https://modelcontextprotocol.io/specification/2025-11-25/basic/transports/
2. MCP authorization specification: https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization/
3. MCP lifecycle/versioning references: https://modelcontextprotocol.io/specification/2025-11-05/basic/lifecycle/ and https://modelcontextprotocol.io/specification/2025-06-18/basic/versioning/

## Problem Statement

<!-- Describe the problem this design addresses -->

## Proposed Solution

<!-- Describe the proposed solution in detail -->

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
