## Embeddable MCP

- [ ] Add more arguments to list-tools
- [ ] Allow passing arguments to test-tool (or maybe even expose one verb per tool)
- [ ] Call the hooks when calling a test-tool

## MCP JS server

- [ ] expose the session ID to the JS
- [ ] reload last sessions JS to recreate the full app as it was (all the code in a session?). Or maybe store which code was used for rest handlers and all that.
- [ ] make sure session ID is scoped to MCP
- [ ] Add view with all the code from a single session
- [ ] Give some name to the code so that it can be used for debugging: TypeError: Value is not object coercible at <eval>:82:47(3)

- [ ] req.body is apparently a string, not a parsed object (prompt? funcitonality?)
- [ ] give better error messages showing the relevant code as well instead of <eval>
- [ ] function to retrieve logs maybe?
- [x] allow for a richer o-TTP API allowing redirects and all that
- [ ] Output registered endpoints
- [ ] Allow for defining functions that can be used across executeJS calls (sadly we have the wrapped stuff right now, maybe we could parse for let / const?)
- [ ] allow querying the logs as well
- [ ] allow call for querying the endpoints to do debugging
- [ ] allow to load plugins / interface with external APIs
- [ ] get errors back from db.exec / db.query
- [ ] catch panics
- [ ] allow for a repl to a session on the CLI as well
- [ ] unit test suite for the JS too
- [ ] async notification for error 500? get a way to get the logs. maybe a self-fix function
- [ ] get rid of the function wrapped stuff
- [ ] add function to restart the VM from scratch
- [ ] allow loading thirdparty libs
- [ ] live view of both MCP calls and REST calls / JS evaluation
- [ ] add JS repl to the web UI
- [ ] have live console
- [ ] allow editing of sent JS and reload it for manual fixing
- [ ] download session zip
- [ ] MCP session ID is wrong
- [ ] Prompt engineering for separating js / css /html into multiple endpoints
- [ ] More clever routes (and params capture?)
- [ ] how to dispatch to multiple sessions on the rest side, since they will all have different routes? multiple ports? subpaths? use cool names for session ids instead of uuids / cool encoding?
- [ ] make registerHandler sytnax more express.js like to leverage training corpus

## Features

- [x] Add tool/prompt profiles to switch between different collections of prompts and resources and other things
- [ ] Add environment variable pass through / .env loading to shell script tools
- [ ] Add server log viewing integration

- [ ] Refactor claude_config.go and cursor_config.go to use the new CommonServer abstraction and unify into one set of commands
- [x] Add config editing UI


## Scholarly

- [ ] Web API / Web UI for search doesn't seem to rerank
- [ ] Merge both web applications 
- [ ] Better weighting of search results (crossref seems very empty)
- [ ] Integrate bge-reranker into go using ollama (see stuff in labs)
- [ ] Make Web UI for the other verbs
- [ ] Add query intent for reranking 

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
- [ ] add notification handler
- [ ] add resource templates

### Bugs
- [x] BUG: figure out why closing the client seems to hang
- [ ] Still using the wrong loggers in server (one with terminal, the other, for example)

## MCP server

- [ ] plugin API to register servers
- [ ] register glazed commands
- [ ] allow config file for all settings
- [x] figure out how to easily register bash commands to the MCP
- [ ] dynamic loading / enabling / removing servers
- [ ] add resource templates
- [ ] add tools from openapi json
- [ ] support openai actions protocol
- [ ] Add type field to CommandDescription to allow go-go-mcp to load any number of them (escuse-me, sqleton, pinocchio, etc...)

- [X] Allow debug logging
- [x] Implement missing SSE methods
- [x] BUG: killing server doesn't seem to kill hanging connections (when using inspector, for example)
- [ ] send out notifications
- [ ] pass session id as environment variable
- [ ] cancelling running shell scripts through KILL

- [ ] Allow logging to separate file (to debug claude for example)
  - seems kind of broken, there is a different logger running after the initial logger is setup

- [ ] Register commands using go introspection, like in pinocchio's tools
- [ ] Make it easy to register a struct with multiple tool handlers (say, to keep a single handle to a resource), linked to the session_id
- [x] Pass the session id to the tool  (maybe as part of the context?)
- [ ] Track crashes in a log file

- [ ] Build stdio server that forwards to sse server


## MCP ideas

- [ ] Diary entries / summary
- [ ] Arduino connection