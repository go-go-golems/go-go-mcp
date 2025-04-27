# Adding Internal Servers Flag to Server Start Command

## Overview

This document outlines how to add a new `--internal-servers` flag to the `start` command in `go-go-mcp/cmd/go-go-mcp/cmds/server/start.go`. This flag will allow users to enable built-in example tools provided in the `pkg/tools/examples` directory.

## Current State

The `go-go-mcp` server command currently supports different transport types (stdio and sse) but doesn't provide a way to easily register the example tools that are included in the codebase:

- `sqlite.go`: A tool for executing SQL queries against Cursor's SQLite database
- `fetch.go`: A tool for fetching content from URLs and converting it to markdown
- `echo.go`: A simple tool that echoes the input back to the user

These tools are defined in the `pkg/tools/examples` directory with registration functions:
- `RegisterSQLiteTool(registry *tool_registry.Registry)`
- `RegisterFetchTool(registry *tool_registry.Registry)`
- `RegisterEchoTool(registry *tool_registry.Registry)`

## Implementation Plan

### 1. Modify the StartCommandSettings Struct

Add a new field to track the internal servers that should be registered:

```go
type StartCommandSettings struct {
    Transport      string   `glazed.parameter:"transport"`
    Port           int      `glazed.parameter:"port"`
    InternalServers []string `glazed.parameter:"internal-servers"`
}
```

### 2. Add the Flag to Command Definition

Modify the `NewStartCommand()` function to include the new flag:

```go
cmds.WithFlags(
    // existing flags...
    parameters.NewParameterDefinition(
        "internal-servers",
        parameters.ParameterTypeStringList,
        parameters.WithHelp("List of internal servers to register (comma-separated). Available: sqlite,fetch,echo"),
        parameters.WithDefault([]string{}),
    ),
),
```

### 3. Create a Function to Register Internal Servers

Add a new function to register the requested internal servers:

```go
func registerInternalServers(registry *tool_registry.Registry, serverList []string) error {
    // Create a map for faster lookups
    serversMap := make(map[string]bool)
    for _, server := range serverList {
        serversMap[server] = true
    }

    // Register the requested servers
    if serversMap["sqlite"] {
        if err := examples.RegisterSQLiteTool(registry); err != nil {
            return errors.Wrap(err, "failed to register sqlite tool")
        }
    }
    
    if serversMap["fetch"] {
        if err := examples.RegisterFetchTool(registry); err != nil {
            return errors.Wrap(err, "failed to register fetch tool")
        }
    }
    
    if serversMap["echo"] {
        if err := examples.RegisterEchoTool(registry); err != nil {
            return errors.Wrap(err, "failed to register echo tool")
        }
    }

    return nil
}
```

### 4. Modify the Run Method

Update the `Run` method to use the registered internal servers:

```go
// Inside the Run method, after creating the tool provider:
toolProvider, err := layers.CreateToolProvider(serverSettings)
if err != nil {
    return err
}

// Register internal servers if specified
if len(s_.InternalServers) > 0 {
    registry, ok := toolProvider.(*tool_registry.Registry)
    if !ok {
        // If the tool provider is not a registry, create a new one
        registry = tool_registry.NewRegistry()
        
        // Register internal servers
        if err := registerInternalServers(registry, s_.InternalServers); err != nil {
            return err
        }
        
        // Combine with the existing provider
        toolProvider = tools.CombineProviders(toolProvider, registry)
    } else {
        // If it's already a registry, register directly
        if err := registerInternalServers(registry, s_.InternalServers); err != nil {
            return err
        }
    }
}
```

## Usage Examples

Once implemented, users can enable specific internal servers:

```bash
# Start the server with SQLite tool enabled
go run cmd/go-go-mcp/main.go server start --internal-servers sqlite

# Start the server with multiple tools
go run cmd/go-go-mcp/main.go server start --internal-servers sqlite,fetch,echo

# Start the server with SSE transport and the fetch tool
go run cmd/go-go-mcp/main.go server start --transport sse --port 3001 --internal-servers fetch
```

## Implementation Notes

1. The `internal-servers` flag accepts a comma-separated list of server names
2. Invalid server names are silently ignored
3. If the tool provider created by `CreateToolProvider` is not a registry, we create a new registry and combine them
4. This implementation maintains backward compatibility with existing code

## Testing

To test this feature:

1. Start the server with different combinations of internal servers
2. Verify that the specified tools are available and working correctly
3. Test with both transport types (stdio and sse)
4. Ensure that invalid server names don't cause errors

## Next Steps

After implementing this feature, consider:

1. Adding more built-in tools to the examples directory
2. Improving error handling for invalid server names
3. Adding a command to list available internal servers
4. Adding documentation for each internal server 