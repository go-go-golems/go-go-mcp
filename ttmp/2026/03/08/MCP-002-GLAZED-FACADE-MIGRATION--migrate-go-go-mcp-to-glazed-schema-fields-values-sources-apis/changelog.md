# Changelog

## 2026-03-08

- Initial workspace created


## 2026-03-08

Migrated go-go-mcp from legacy Glazed layers/parameters/middlewares APIs to schema/fields/values/sources, removed the stale Parka parameter-filter bridge, and restored workspace-mode go test/go build validation (commit ea4bc44).

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/cmd/go-go-mcp/cmds/server/start.go — Server command updated to values/sections APIs
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/pkg/tools/providers/config-provider/tool-provider.go — Local replacement for Parka-based config middleware translation

