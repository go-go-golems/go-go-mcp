# Cursor MCP Configuration Support

## Overview

Added functionality to manage Cursor MCP configuration files, similar to the existing Claude desktop configuration functionality. This allows users to manage MCP servers in the Cursor IDE configuration files.

## Key Features

- Support for both global (`~/.cursor/mcp.json`) and project-specific (`.cursor/mcp.json`) configuration files
- Support for both stdio format (command and arguments) and SSE format (URL) MCP servers
- Full set of management commands (init, edit, add, remove, list)
- Shortcut command for adding go-go-mcp servers with specific profiles

## Command Structure

```
mcp cursor
  ├── init                # Initialize a new Cursor MCP configuration file
  ├── edit                # Edit the configuration file in your default editor
  ├── add-mcp-server      # Add a new MCP server in stdio format
  ├── add-mcp-server-sse  # Add a new MCP server in SSE format
  ├── add-go-go-server    # Add a go-go-mcp server with a specific profile
  ├── remove-mcp-server   # Remove an MCP server
  └── list-servers        # List all configured MCP servers
```

## Usage Examples

### Initialize a configuration file

```bash
# Initialize global configuration
mcp cursor init --global

# Initialize project configuration
mcp cursor init
```

### Adding an MCP server (stdio format)

```bash
# Add a server to the global configuration
mcp cursor add-mcp-server --global server-name npx -y mcp-server --env API_KEY=your-api-key

# Add a server to the project configuration
mcp cursor add-mcp-server server-name npx -y mcp-server --env API_KEY=your-api-key
```

### Adding an MCP server (SSE format)

```bash
# Add a server to the global configuration
mcp cursor add-mcp-server-sse --global server-name http://localhost:3000/sse --env API_KEY=your-api-key

# Add a server to the project configuration
mcp cursor add-mcp-server-sse server-name http://localhost:3000/sse --env API_KEY=your-api-key
```

### Adding a go-go-mcp server with a profile

```bash
# Add a go-go-mcp server with a specific profile to the global configuration
mcp cursor add-go-go-server --global server-name my-profile

# Add a go-go-mcp server with a specific profile to the project configuration
mcp cursor add-go-go-server server-name my-profile

# Add a go-go-mcp server with additional arguments
mcp cursor add-go-go-server server-name my-profile --args="--port=8080" --args="--verbose"
```

### Removing an MCP server

```bash
# Remove a server from the global configuration
mcp cursor remove-mcp-server --global server-name

# Remove a server from the project configuration
mcp cursor remove-mcp-server server-name
```

### Listing configured servers

```bash
# List servers in the global configuration
mcp cursor list-servers --global

# List servers in the project configuration
mcp cursor list-servers
```

## Implementation

The implementation follows the pattern established by the Claude desktop configuration functionality:

1. Added `CursorMCPConfig` type in `pkg/config/cursor.go` with support for both stdio and SSE server formats
2. Implemented configuration editor in `pkg/config/cursor.go` with methods for managing servers
3. Added command handlers in `cmd/go-go-mcp/cmds/cursor_config.go`
4. Updated the main CLI entry point in `cmd/go-go-mcp/main.go` to include the new commands
5. Added convenience commands for common use cases (e.g., add-go-go-server)

## Next Steps

1. Add tests for the new functionality
2. Consider adding a `copy` command to copy servers between global and project configurations
3. Add validation for server configuration (e.g., URL format for SSE servers) 