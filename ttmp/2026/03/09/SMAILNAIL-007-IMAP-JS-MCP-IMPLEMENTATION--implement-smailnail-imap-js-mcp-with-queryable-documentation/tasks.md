# Tasks

- [x] Create the execution ticket and map the current runtime, module, and jsdoc dependencies.
- [x] Write the implementation plan and detailed task list.
- [x] Add the new `smailnail-imap-mcp` binary and base MCP server wiring.
- [x] Implement `executeIMAPJS` with request binding, runtime bootstrapping, and JSON result shaping.
- [x] Add unit tests for `executeIMAPJS`.
- [ ] Add JS sentinel documentation sources for the `smailnail` module.
- [ ] Implement embedded doc loading and query helpers using `go-go-goja/pkg/jsdoc`.
- [ ] Implement `getIMAPJSDocumentation` with exact lookup, overview, render, and simple search modes.
- [ ] Add documentation drift validation tests for documented/exported symbol alignment.
- [ ] Add maintained smoke scripts for the new MCP and wire them into `Makefile`.
- [ ] Update README/help text for the new binary and tools.
- [ ] Run full validation, update the diary/changelog, and commit the final docs.
