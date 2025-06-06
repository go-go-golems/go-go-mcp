---
Title: Configuring Claude Desktop with MCP Servers
Slug: claude-desktop-config
Short: Learn how to configure Claude desktop to work with multiple MCP servers
Topics:
  - claude
  - desktop
  - config
  - mcp
  - tutorial
Commands:
  - claude-config
  - start
  - bridge
Flags:
  - config-file
  - profile
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Configuring Claude Desktop with MCP Servers 🎮

Want to supercharge your Claude desktop experience with powerful MCP servers? You've come to the right place! This guide will walk you through configuring multiple MCP servers in Claude desktop, unlocking a world of automation possibilities.

## Configuration File Location 📁

Claude desktop looks for its configuration file in your system's config directory:

- Linux: `~/.config/Claude/claude_desktop_config.json`
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

## Command Line Configuration 🛠️

The `go-go-mcp claude-config` command provides a suite of tools to manage your Claude desktop configuration:

### Initialize Configuration

Create a new configuration file:

```bash
# Create with default path
go-go-mcp claude-config init

# Create with custom path
go-go-mcp claude-config init --config /path/to/config.json
```

### Edit Configuration

Open the configuration file in your default editor:

```bash
# Edit default config
go-go-mcp claude-config edit

# Edit custom config
go-go-mcp claude-config edit --config /path/to/config.json
```

### Add MCP Server

Add or update an MCP server configuration:

```bash
# Add a basic server
go-go-mcp claude-config add-mcp-server dev \
  --command go-go-mcp \
  --args start --profile development --log-level debug

# Add with environment variables
go-go-mcp claude-config add-mcp-server github \
  --command go-go-mcp \
  --args start --profile github \
  --env GITHUB_TOKEN=your-token \
  --env DEBUG=true

# Update existing server
go-go-mcp claude-config add-mcp-server dev \
  --command go-go-mcp \
  --args start --profile development \
  --overwrite
```

### Remove MCP Server

Remove an existing server configuration:

```bash
# Remove a server
go-go-mcp claude-config remove-mcp-server dev

# Remove from custom config
go-go-mcp claude-config remove-mcp-server dev --config /path/to/config.json
```

### List Servers

View all configured servers:

```bash
# List all servers
go-go-mcp claude-config list-servers

# List from custom config
go-go-mcp claude-config list-servers --config /path/to/config.json
```

## Understanding the Configuration Format 🔧

The configuration file uses JSON format with two main sections:

1. `mcpServers`: A map of named MCP server configurations
2. `disabledMCPServers`: An optional list of disabled server names

Here's the anatomy of an MCP server configuration:

```json
{
  "mcpServers": {
    "server1": {
      "command": "executable-name",    // The command to run
      "args": [                       // List of command arguments
        "--flag1", "value1",
        "--flag2", "value2"
      ],
      "env": {                        // Optional environment variables
        "ENV_VAR1": "value1",
        "ENV_VAR2": "value2"
      }
    }
  },
  "disabledMCPServers": [            // Optional list of disabled servers
    "server2",
    "server3"
  ]
}
```

## Exciting Configuration Examples! 🚀

Let's explore some powerful configurations that showcase what you can do:

### 1. Basic Development Setup

```json
{
  "mcpServers": {
    "dev": {
      "command": "go-go-mcp",
      "args": [
        "start",
        "--profile", "development",
        "--log-level", "debug"
      ],
      "env": {
        "DEVELOPMENT": "true"
      }
    }
  }
}
```

### 2. Multi-Server Production Environment

```json
{
  "mcpServers": {
    "api-tools": {
      "command": "go-go-mcp",
      "args": [
        "start",
        "--profile", "api",
      ]
    },
    "data-tools": {
      "command": "go-go-mcp",
      "args": [
        "start",
        "--profile", "data",
      ]
    },
    "system-tools": {
      "command": "go-go-mcp",
      "args": [
        "start",
        "--profile", "system",
      ]
    }
  }
}
```

### 3. Bridge Configuration for Remote Servers

```json
{
  "mcpServers": {
    "remote-bridge": {
      "command": "go-go-mcp",
      "args": [
        "bridge",
        "--sse-url", "https://remote-mcp.example.com",
        "--log-level", "debug"
      ],
      "env": {
        "MCP_AUTH_TOKEN": "your-secret-token"
      }
    }
  }
}
```

### Enable/Disable Servers

You can temporarily disable servers without removing their configuration:

```bash
# Disable a server
go-go-mcp claude-config disable-server dev

# Enable a previously disabled server
go-go-mcp claude-config enable-server dev
```

When listing servers, disabled servers will be marked:

```bash
# List all servers including disabled ones
go-go-mcp claude-config list-servers
```

Example output:
```
Configured MCP servers in ~/.config/Claude/claude_desktop_config.json:

dev (disabled):
  Command: go-go-mcp
  Args: [start --profile development --log-level debug]
  Environment:
    DEVELOPMENT: true

github:
  Command: go-go-mcp
  Args: [start --profile github]
  Environment:
    GITHUB_TOKEN: your-token

Disabled servers:
  - dev
```

## Pro Tips 💡

1. **Server Names**: Choose descriptive names for your MCP servers that reflect their purpose
2. **Environment Variables**: Keep sensitive data in environment variables
3. **Logging**: Use `--log-level debug` during development for detailed logs
4. **Profiles**: Create different profiles for different use cases

## Common Patterns 🎯

For more details on specific topics, see:
- [MCP Protocol](01-mcp-protocol.md)
- [Shell Commands](02-shell-commands.md)
- [MCP in Practice](03-mcp-in-practice.md) 