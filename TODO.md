## Tool API

- [x] Add context to the tool call
- [x] Don't allow a tool to be registered without interface
- [x] Make tool an interface?

## MCP client

- [x] mcp-client should be able to run stdio servers
- [ ] allow loading of .env file
- [ ] make it easy to run from docker image
- [ ] load config file
- [ ] REPL mode / TUI
- [X] Add debug logging
- [ ] make web ui to easily debug / interact

### Bugs
- [ ] BUG: figure out why closing the client seems to hang

## MCP server

- [ ] plugin API to register servers
- [ ] register glazed commands
- [ ] allow config file for all settings
- [ ] figure out how to easily register bash commands to the MCP
- [ ] dynamic loading / enabling / removing servers

- [X] Allow debug logging
- [ ] Implement missing SSE methods