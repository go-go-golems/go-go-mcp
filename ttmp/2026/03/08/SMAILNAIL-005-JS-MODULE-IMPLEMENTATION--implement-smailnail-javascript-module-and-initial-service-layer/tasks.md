# Tasks

## Completed

- [x] T1 Create and normalize the implementation ticket, task list, and diary baseline

## In Progress

- [ ] T2 Add `go-go-goja` as a local dependency in `smailnail` and verify the workspace resolves cleanly

## Planned

- [ ] T3 Introduce a pure Go `pkg/services/smailnailjs` package for rule parsing, rule building, and result shaping
- [ ] T4 Add a connection/session abstraction with an injectable dialer so JS-facing tests do not require live IMAP
- [ ] T5 Implement a native `smailnail` goja module with top-level exports for `parseRule`, `buildRule`, and `newService`
- [ ] T6 Add service-layer unit tests covering rule parsing/building and message shaping
- [ ] T7 Add goja integration tests proving `require("smailnail")` works through `engine.NewBuilder(...).Build().NewRuntime(...)`
- [ ] T8 Add a repo-maintained JS demo or smoke entrypoint plus a ticket-local script that exercises it
- [ ] T9 Run `go test` validation, update docs/diary/changelog, and commit focused changes in stages
