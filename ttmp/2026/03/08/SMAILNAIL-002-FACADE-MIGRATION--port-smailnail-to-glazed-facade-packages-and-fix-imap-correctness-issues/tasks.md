# Tasks

## Completed

- [x] Inventory and relate all legacy Glazed imports, tags, and parser wiring affected by the facade migration
- [x] Port `pkg/imap/layer.go` to facade sections and fields, then update downstream slugs and tags
- [x] Port `cmd/smailnail` root plus `mail-rules` and `fetch-mail` to facade parser and values APIs
- [x] Port `cmd/mailgen` root and `generate` to facade parser and values APIs
- [x] Port `cmd/imap-tests` root and helper commands to facade parser and values APIs
- [x] Fix UID-based IMAP action handling in `pkg/dsl/actions.go`
- [x] Fix address parsing and header serialization in `cmd/mailgen/cmds/generate.go`
- [x] Add regression tests for action targeting, fetch behavior, and address serialization
- [x] Repair runtime defects discovered during Docker validation in `pkg/dsl/fetch.go` and `cmd/smailnail/commands/mail_rules.go`
- [x] Update stale user-facing docs to match the migrated CLI surface
- [x] Validate builds, tests, and help output locally
- [x] Validate mailbox creation, fetch, rule actions, and `mailgen --store-imap` flows against `/home/manuel/code/others/docker-test-dovecot`
- [x] Finalize the diary, changelog, file relations, and doctor pass for the ticket
