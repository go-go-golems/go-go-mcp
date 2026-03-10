# Tasks

## Completed

- [x] T1 Create and normalize the implementation ticket, task list, and diary baseline
- [x] T2 Add `go-go-goja` as a local dependency in `smailnail` and verify the workspace resolves cleanly
- [x] T3 Introduce a pure Go `pkg/services/smailnailjs` package for rule parsing, rule building, and result shaping
- [x] T4 Add a connection/session abstraction with an injectable dialer so JS-facing tests do not require live IMAP
- [x] T5 Implement a native `smailnail` goja module with top-level exports for `parseRule`, `buildRule`, and `newService`
- [x] T6 Add service-layer unit tests covering rule parsing/building and message shaping
- [x] T7 Add goja integration tests proving `require("smailnail")` works through `engine.NewBuilder(...).Build().NewRuntime(...)`
- [x] T8 Add a repo-maintained JS demo or smoke entrypoint plus a ticket-local script that exercises it
- [x] T9 Run `go test` validation, update docs/diary/changelog, and commit focused changes in stages
