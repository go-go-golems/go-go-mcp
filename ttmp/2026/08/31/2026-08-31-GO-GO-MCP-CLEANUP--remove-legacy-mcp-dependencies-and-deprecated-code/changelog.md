# Changelog

## 2026-08-31

- Initial workspace created


## 2026-08-31 - Initial cleanup design

Created a repository-local cleanup ticket and staged design. Phase 1 targets the unreferenced pkg/client and its exclusive r3labs/sse dependency; the active official SDK backend and custom protocol facade remain protected for a separate future migration.

### Related Files

- /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/client/client.go — Primary orphaned legacy client deletion candidate
- /home/manuel/workspaces/2026-08-28/coinvault-oidc-mcp/go-go-mcp/pkg/embeddable/official_backend.go — Current official SDK integration boundary

