# Changelog

## 2026-03-09

- Initial workspace created


## 2026-03-09

Landed the first runtime slice in commit ff584b4 with the new smailnail-imap-mcp binary, executeIMAPJS handler, focused tests, and a placeholder docs tool.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail-imap-mcp/main.go — New dedicated MCP binary
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go — First working MCP execution tool


## 2026-03-09

Landed the documentation-query slice in commit dc0c5f3 with embedded JS sentinel docs, the docs registry/query layer, and a real getIMAPJSDocumentation tool.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/docs_query.go — Query modes and rendered markdown response shaping
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/docs_registry.go — Embedded docs registry and local example-body extraction


## 2026-03-09

Landed the drift/smoke slice in commit 1ac7866 with runtime-backed documentation drift tests, a maintained smoke script, and README/Makefile wiring for smailnail-imap-mcp.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/README.md — User-facing binary and smoke instructions
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module_test.go — Detect drift between documented and exported JS symbols
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/imap-js-mcp-smoke.sh — Repo-maintained smoke entrypoint for the new MCP binary

