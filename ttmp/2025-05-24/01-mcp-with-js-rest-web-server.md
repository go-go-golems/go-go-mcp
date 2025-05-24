- mcp server which has embedded goja js using go-go-mcp

## Javascript side

- access to sqlite
- able to store functions and global state
- session is scoped to a MCP session
- tool call for executing javascript in the vm: execute_js(code: string)

- API to register functions as REST handlers
- API to register function that returns a file under a URL

## TODO

- [ ] Hot reload on the browser side when an endpoint is updated