# New Shortcut Command: add-go-go-server

## Overview

Added a new shortcut command `add-go-go-server` to both the Claude desktop and Cursor MCP configuration management. This command simplifies the process of adding go-go-mcp servers with specific profiles to the configuration files.

## Command Purpose

The `add-go-go-server` command makes it easier to add go-go-mcp server configurations by automating the setup of a server that uses the `mcp server start --profile` command. This is particularly useful for developers who work with multiple profiles and want to quickly set up servers without specifying the full command path and arguments each time.

## Implementation Details

The command has been added to both:
- Claude desktop configuration (`mcp claude add-go-go-server`)
- Cursor MCP configuration (`mcp cursor add-go-go-server`)

The implementation:
1. Automatically finds the `mcp` executable in the system PATH
2. Sets up the command with proper arguments for starting a server with a specific profile
3. Allows additional arguments to be passed to the server command
4. Supports environment variables for server configuration

## Usage Examples

### Claude desktop configuration

```bash
# Basic usage
mcp claude add-go-go-server my-server my-profile

# With additional arguments
mcp claude add-go-go-server my-server my-profile --args="--port=8080" --args="--verbose"

# With environment variables
mcp claude add-go-go-server my-server my-profile --env API_KEY=your-api-key

# Overwrite existing server
mcp claude add-go-go-server my-server my-profile --overwrite
```

### Cursor MCP configuration

```bash
# Global configuration
mcp cursor add-go-go-server --global my-server my-profile

# Project-specific configuration
mcp cursor add-go-go-server my-server my-profile

# With additional arguments
mcp cursor add-go-go-server --global my-server my-profile --args="--port=8080" --args="--verbose"

# With environment variables
mcp cursor add-go-go-server --global my-server my-profile --env API_KEY=your-api-key
```

## Command Parameters

- `NAME`: The name of the server in the configuration
- `PROFILE`: The profile to use with the `mcp server start --profile` command
- `[ARGS...]`: Optional additional arguments to pass to the server command
- `--args`: Additional arguments to pass to the server command (can be specified multiple times)
- `--env`: Environment variables in KEY=VALUE format (can be specified multiple times)
- `--overwrite`: Overwrite existing server if it exists

### Cursor-specific parameters
- `--global`: Use global configuration (default: true)
- `--project-dir`: Project directory (defaults to current directory)
- `--config`: Path to config file

## Benefits

This command provides several advantages:
1. **Simplicity**: Reduces the complexity of adding go-go-mcp servers with profiles
2. **Consistency**: Ensures servers are configured consistently with the correct command format
3. **Automation**: Automatically finds the mcp binary path, removing the need to specify it manually
4. **Flexibility**: Supports additional arguments and environment variables for full customization 