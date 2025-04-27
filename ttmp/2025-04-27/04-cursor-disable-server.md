# Enhanced Cursor MCP Configuration with Server Enable/Disable Functionality

## Overview

Added the ability to temporarily disable and re-enable MCP servers in the Cursor configuration without removing them. This feature mirrors the functionality already available in the Claude desktop configuration commands.

## Key Features

- Support for disabling servers without removing their configurations
- Support for re-enabling previously disabled servers
- Visual indication of disabled status when listing servers
- Separate section listing all disabled servers

## New Commands

Two new commands have been added to the Cursor MCP configuration tools:

```
mcp cursor
  ├── ...
  ├── disable-server    # Disable an MCP server without removing its configuration
  ├── enable-server     # Re-enable a previously disabled MCP server
  └── ...
```

## Implementation Details

The implementation involved:

1. Enhancing the `CursorMCPConfig` structure to include a `DisabledServers` map
2. Adding methods to the `CursorMCPEditor` for:
   - Disabling servers (`DisableMCPServer`)
   - Enabling servers (`EnableMCPServer`)
   - Checking server status (`IsServerDisabled`)
   - Listing disabled servers (`ListDisabledServers`)
3. Adding command handlers for:
   - `disable-server`
   - `enable-server`
4. Updating the `list-servers` command to show disabled status

## Usage Examples

### Disabling a Server

```bash
# Disable a server in the global configuration
mcp cursor disable-server --global my-server

# Disable a server in the project configuration
mcp cursor disable-server my-server
```

### Re-enabling a Server

```bash
# Enable a previously disabled server in the global configuration
mcp cursor enable-server --global my-server

# Enable a previously disabled server in the project configuration
mcp cursor enable-server my-server
```

### Listing Servers (with disabled status)

```bash
# List all servers including disabled ones
mcp cursor list-servers --global
```

Example output:
```
Configured MCP servers in ~/.cursor/mcp.json:

github:
  Command: mcp
  Args: [server start --profile github]
  Environment:
    GITHUB_TOKEN: your-token

dev-tools (disabled):
  Command: mcp
  Args: [server start --profile dev --log-level debug]
  Environment:
    DEVELOPMENT: true

Disabled servers:
  - dev-tools
```

## Benefits

This feature provides several advantages:

1. **Flexibility**: Easily switch between different server configurations without losing settings
2. **Organization**: Temporarily disable servers that are not currently needed
3. **Testing**: Quickly enable/disable servers for testing different configurations
4. **Consistency**: Same functionality available for both Claude and Cursor configurations 