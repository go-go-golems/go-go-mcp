---
Title: MCP implementation testing plan
Ticket: MCP-002-GLAZED-FACADE-MIGRATION
Status: active
Topics:
    - mcp
    - go
    - glazed
    - refactor
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/go-go-mcp/cmds/client/prompts.go
      Note: Evidence that prompt probing exists client-side but is not yet part of the hard gate
    - Path: cmd/go-go-mcp/cmds/client/resources.go
      Note: Evidence that resource probing exists client-side but is not yet part of the hard gate
    - Path: cmd/go-go-mcp/cmds/client/tools.go
      Note: Defines the tool-facing client checks used in the smoke plan
    - Path: cmd/go-go-mcp/cmds/server/start.go
      Note: Defines the currently wired MCP server surface and transport bootstrap path
    - Path: pkg/tools/providers/config-provider/tool-provider.go
      Note: Primary semantic-risk area that needs focused regression tests
    - Path: pkg/tools/providers/config-provider/tool-provider_test.go
      Note: Initial direct regression coverage for config-provider precedence and filtering
    - Path: ttmp/2026/03/08/MCP-002-GLAZED-FACADE-MIGRATION--migrate-go-go-mcp-to-glazed-schema-fields-values-sources-apis/scripts/tool-transport-smoke.sh
      Note: Ticket-local validated smoke harness for the currently supported transports
    - Path: ttmp/2026/03/08/MCP-002-GLAZED-FACADE-MIGRATION--migrate-go-go-mcp-to-glazed-schema-fields-values-sources-apis/sources/local/01-mcp-testing.md
      Note: Imported external note that informed the tooling recommendations
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-08T19:13:45.966123015-04:00
WhatFor: ""
WhenToUse: ""
---



# MCP implementation testing plan

## Executive Summary

`go-go-mcp` now passes workspace-mode compile validation after the Glazed facade migration, but that still leaves a testing gap at the MCP runtime level. The server startup path currently wires a tool registry into the embeddable backend, and this ticket's runtime experiments confirm that the tool surface works across `command`, `sse`, and `streamable_http` transports when exercised through the built-in `echo` internal server.

The immediate testing gate should therefore focus on what the runtime actually serves today: `tools/list` and `tools/call`. Prompt and resource client commands exist, but they should be treated as expected-gap probes until the server backend registers those providers. In parallel, the Glazed migration introduced one especially important semantic risk area in `pkg/tools/providers/config-provider/tool-provider.go`, where `defaults`, `overrides`, `whitelist`, and `blacklist` are now translated locally into `sources` middleware after the old Parka bridge was removed.

## Problem Statement

The imported note at `sources/local/01-mcp-testing.md` recommends the official MCP Inspector CLI and the official conformance runner. That recommendation is useful, but it is too generic on its own. `go-go-mcp` is not just "an MCP server"; it is a Go CLI with multiple transports, config-driven command loading, internal servers, and an authentication surface. A testing plan for this repository needs to distinguish between protocol-level tooling, repo-specific behavior, and runtime surfaces that are not fully wired yet.

There is also a mismatch between what the client can ask for and what the server currently exposes. `cmd/go-go-mcp/cmds/client/prompts.go` and `cmd/go-go-mcp/cmds/client/resources.go` exist, but `cmd/go-go-mcp/cmds/server/start.go` currently registers only tools with the embeddable backend. A naive "test every MCP capability equally" strategy would therefore produce failing checks that reflect known architecture limits rather than fresh regressions.

The plan needs to provide:

1. A hard green gate for behavior that must work now.
2. A separate bucket for exploratory probes and expected gaps.
3. Focused regression tests for the migrated config semantics.
4. A clear role for external MCP tooling without making the repo depend entirely on upstream test stability.

## Proposed Solution

Adopt a five-layer testing strategy.

### Layer 1: Workspace build validation

Keep the existing fast repository checks as the first gate:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp
go test ./...
go build ./cmd/go-go-mcp
```

This catches API drift against the local `glazed/` checkout, broken Cobra registration, and compile-time migration fallout. It does not verify MCP protocol behavior, but it remains the cheapest failure signal and should stay mandatory.

### Layer 2: Tool-transport smoke gate

Use the ticket-local harness at `scripts/tool-transport-smoke.sh` as the current runtime smoke gate. The script was validated during this ticket and now passes end to end.

What the script does:

- starts a stdio-backed server through `--transport command`
- validates `tools/list` against the internal `echo` server
- validates `tools/call echo` and checks the returned message text
- repeats the same assertions for `sse`
- repeats the same assertions for `streamable_http`
- wraps every client-side invocation in `timeout 20` to avoid indefinite hangs

Representative invocation:

```bash
bash ttmp/2026/03/08/MCP-002-GLAZED-FACADE-MIGRATION--migrate-go-go-mcp-to-glazed-schema-fields-values-sources-apis/scripts/tool-transport-smoke.sh
```

This layer should be part of everyday local validation and any CI job that claims to protect the MCP runtime.

### Layer 3: Config-provider semantic regression tests

The highest-risk migration area is the config-provider path in `pkg/tools/providers/config-provider/tool-provider.go`. During the Glazed migration, the old Parka adapter was removed and replaced with a local middleware chain built from:

- `sources.BlacklistSectionFieldsFirst(...)`
- `sources.WhitelistSectionFieldsFirst(...)`
- `sources.FromMap(... overrides ...)`
- `sources.FromMapAsDefaultFirst(... defaults ...)`
- followed by command defaults and user arguments

This logic needs direct Go tests, because a transport smoke test can tell us only that a tool call succeeded, not whether field precedence was correct.

The tests should cover:

- `defaults` filling only missing fields
- `overrides` taking precedence over user input
- `whitelist` removing non-allowed fields
- `blacklist` removing denied fields
- combinations where defaults, overrides, and caller args all reference the same field

Pseudocode for the desired shape:

```text
given a command schema with fields message, format, hidden
and config:
  defaults  = {format: "plain"}
  overrides = {message: "forced"}
  whitelist = ["message", "format"]
  blacklist = ["hidden"]
when executeCommand runs with args:
  {message: "user", hidden: "ignored"}
then the final parsed values should decode to:
  message = "forced"
  format = "plain"
and hidden should not survive filtering
```

These tests belong in normal `go test ./...`.

Initial coverage is now implemented in `pkg/tools/providers/config-provider/tool-provider_test.go`. The current tests cover precedence across config defaults, schema defaults, user arguments, and overrides, plus whitelist/blacklist filtering of removed fields.

### Layer 4: Capability probes with expected-gap classification

Prompt and resource checks should exist, but they should not yet be part of the hard green gate.

Current rationale:

- the client has prompt/resource commands
- the server startup path builds only a tool registry into the backend
- a prompt/resource failure today is not necessarily a new regression

Recommended handling now:

- `tools/*` failures are hard failures
- `prompts/*` and `resources/*` checks are recorded as exploratory probes
- once the backend wiring changes, those probes can be promoted into required tests

### Layer 5: External black-box protocol tooling

External MCP tools should complement repo-owned tests.

Recommended order:

1. MCP Inspector CLI as the primary scripted black-box probe.
2. MCP conformance runner as an optional second pass.

Why this order:

- Inspector CLI is practical for targeted scripted checks like `tools/list` and `tools/call`
- the imported note identifies it as the most agent-friendly black-box tool
- the official conformance runner is useful, but upstream still describes it as unstable

Suggested future wrapper:

```bash
./scripts/test-external-mcp.sh
```

Suggested flow inside that wrapper:

```text
1. start go-go-mcp with a fixed internal server set
2. run inspector CLI for tools/list
3. run inspector CLI for tools/call echo
4. optionally run conformance initialize/tools scenarios
5. leave prompts/resources optional until runtime wiring exists
```

## Design Decisions

### Decision: gate only the surfaces that the runtime actually serves

This keeps CI honest. A green result should mean the implemented server surface is healthy, not that contributors learned to ignore failures for unwired capabilities.

### Decision: keep config semantic checks inside Go tests

The migrated behavior in the config-provider layer is specific to this repo and cannot be validated precisely enough with protocol-level black-box tools alone.

### Decision: use the internal `echo` server as the smoke fixture

This keeps the transport smoke deterministic and self-contained. No external YAML configs, shell command directories, or third-party services are needed.

### Decision: treat Clay/Viper deprecation warnings as cleanup debt, not immediate test failures

The smoke runs consistently emitted warnings such as `clay.InitViper is deprecated; use InitGlazed and config middlewares`. They do not currently break functionality, but they should be tracked because they pollute output and can mask more important diagnostics.

## Alternatives Considered

### Conformance-only testing

Rejected because it would leave config middleware semantics under-tested and would make the repo too dependent on an upstream tool that is still marked unstable.

### Compile-only testing

Rejected because `go test ./...` and `go build` do not prove that the MCP transports start correctly or that client/server interaction still works.

### Immediate hard failures for prompts/resources

Rejected because the server backend does not currently register those capabilities, so those failures would not cleanly indicate a regression.

## Implementation Plan

1. Keep `go test ./...` and `go build ./cmd/go-go-mcp` as required baseline checks.
2. Keep `scripts/tool-transport-smoke.sh` as the current runtime smoke gate for tools.
3. Add focused Go tests around config defaults, overrides, whitelist, and blacklist behavior.
4. Add a second optional probe script for prompts/resources that records results without failing the main gate.
5. Add an optional Node-based external-tool script using Inspector CLI first and conformance second.
6. When prompt/resource registries are wired into the backend, promote those probes into required tests.
7. Open a cleanup follow-up for the deprecated Clay/Viper initialization path if the warnings remain.

## Test Matrix

| Area | Example check | Expected result now | Interpretation if it fails |
| --- | --- | --- | --- |
| Workspace compile | `go test ./...` | Pass | Build or API regression |
| CLI build | `go build ./cmd/go-go-mcp` | Pass | Build or link regression |
| Command transport | `client tools list/call` via `--transport command` | Pass | stdio or command transport regression |
| SSE transport | `client tools list/call` via `--transport sse` | Pass | SSE server/client regression |
| Streamable HTTP transport | `client tools list/call` via `--transport streamable_http` | Pass | HTTP transport regression |
| Config middleware semantics | direct Go tests in config-provider | Should be added, then pass | behavior drift from Parka replacement |
| Prompts | `client prompts ...` | Exploratory only | likely expected gap until backend wiring exists |
| Resources | `client resources ...` | Exploratory only | likely expected gap until backend wiring exists |
| OIDC metadata/protected resource | curl or Go smoke | Should pass where already implemented | auth regression |
| Inspector CLI / conformance | optional external script | Optional for now | black-box protocol issue or tooling drift |

## Open Questions

1. Should prompt/resource backend wiring be the next implementation ticket, or stay explicitly out of scope for now?
2. Do we want the Inspector CLI / conformance layer in CI immediately, or keep it manual until the local test layers settle?
3. Should the deprecation-warning cleanup happen in `go-go-mcp`, in shared initialization helpers, or both?

## References

- `sources/local/01-mcp-testing.md`
- `cmd/go-go-mcp/cmds/server/start.go`
- `cmd/go-go-mcp/cmds/client/tools.go`
- `cmd/go-go-mcp/cmds/client/prompts.go`
- `cmd/go-go-mcp/cmds/client/resources.go`
- `pkg/tools/providers/config-provider/tool-provider.go`
- `pkg/tools/providers/config-provider/tool-provider_test.go`
- `scripts/tool-transport-smoke.sh`
