# Tasks

## TODO

- [x] Write the implementation plan and implementation diary for the hosted `smailnaild` SQL bootstrap slice
- [x] Scaffold the `smailnaild` binary root command with Glazed help/logging setup
- [x] Add a `serve` command that includes Clay SQL and dbt sections plus hosted server flags
- [x] Implement app-database config loading with SQLite-first defaults while preserving Postgres compatibility
- [x] Implement SQL bootstrap that opens the configured database and initializes a minimal application metadata table
- [x] Implement a minimal hosted HTTP server with `/healthz`, `/readyz`, and `/api/info`
- [x] Add focused tests for DB config/defaulting and hosted HTTP handlers
- [x] Update the smailnail README with `smailnaild` build and run instructions
- [x] Relate the implementation files to the ticket docs and keep the diary/changelog current
- [x] Run validation (`go test`, ticket doctor) and commit code and docs in focused increments
