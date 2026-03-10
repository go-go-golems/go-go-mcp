# Changelog

## 2026-03-08

- Initial workspace created
- Ported all `smailnail`, `mailgen`, and `imap-tests` commands to Glazed facade APIs in commit `dbc9e00`
- Fixed IMAP UID action handling, mail header address serialization, two runtime validation defects, and updated the repo docs in commit `cd446d2`
- Finalized the ticket implementation guide, diary, task list, file relations, and doctor hygiene checks
- Promoted the Docker IMAP validation script into `smailnail/scripts/docker-imap-smoke.sh`, added `make smoke-docker-imap`, and recorded the exact `golangci-lint` hook failure mode
