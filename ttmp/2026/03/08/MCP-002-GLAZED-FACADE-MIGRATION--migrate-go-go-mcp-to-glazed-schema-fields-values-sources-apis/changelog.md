# Changelog

## 2026-03-08

- Initial workspace created


## 2026-03-08

Migrated go-go-mcp from legacy Glazed layers/parameters/middlewares APIs to schema/fields/values/sources, removed the stale Parka parameter-filter bridge, and restored workspace-mode go test/go build validation (commit ea4bc44).

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/cmd/go-go-mcp/cmds/server/start.go — Server command updated to values/sections APIs
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/tools/providers/config-provider/tool-provider.go — Local replacement for Parka-based config middleware translation


## 2026-03-08

Imported external MCP testing guidance, added a repo-specific testing plan, and validated a new transport smoke harness across command, SSE, and streamable HTTP.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/MCP-002-GLAZED-FACADE-MIGRATION--migrate-go-go-mcp-to-glazed-schema-fields-values-sources-apis/design-doc/02-mcp-implementation-testing-plan.md — Records the layered testing strategy and expected-gap classification
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/08/MCP-002-GLAZED-FACADE-MIGRATION--migrate-go-go-mcp-to-glazed-schema-fields-values-sources-apis/scripts/tool-transport-smoke.sh — Provides the validated runtime smoke checks for the current transport surface


## 2026-03-08

Added direct config-provider regression tests for defaults, overrides, whitelist, and blacklist behavior, and revalidated the full repository test suite (commit bec5cc6).

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/tools/providers/config-provider/tool-provider.go — Production middleware ordering validated by the new tests
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/tools/providers/config-provider/tool-provider_test.go — New executable coverage for the Parka replacement semantics

