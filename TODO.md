## Embeddable MCP

- [ ] Add more arguments to list-tools
- [ ] Allow passing arguments to test-tool (or maybe even expose one verb per tool)
- [ ] Call the hooks when calling a test-tool

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