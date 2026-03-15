# Tasks

## TODO

- [x] Write the implementation plan and implementation diary for the hosted `smailnaild` SQL bootstrap slice
- [ ] Scaffold the `smailnaild` binary root command with Glazed help/logging setup
- [ ] Add a `serve` command that includes Clay SQL and dbt sections plus hosted server flags
- [ ] Implement app-database config loading with SQLite-first defaults while preserving Postgres compatibility
- [ ] Implement SQL bootstrap that opens the configured database and initializes a minimal application metadata table
- [ ] Implement a minimal hosted HTTP server with `/healthz`, `/readyz`, and `/api/info`
- [ ] Add focused tests for DB config/defaulting and hosted HTTP handlers
- [ ] Update the smailnail README with `smailnaild` build and run instructions
- [ ] Relate the implementation files to the ticket docs and keep the diary/changelog current
- [ ] Run validation (`go test`, ticket doctor) and commit code and docs in focused increments
