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

## Understanding the Configuration Format 🔧

The configuration file uses JSON format with one main section:

1. `mcpServers`: A map of named MCP server configurations

Here's the anatomy of an MCP server configuration:

```json
{
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

### 4. GitHub Integration Setup

```json
{
  "mcpServers": {
    "github": {
      "command": "go-go-mcp",
      "args": [
        "start",
        "--profile", "github",
      ],
      "env": {
        "GITHUB_TOKEN": "your-github-token"
      }
    }
  }
}
```

### 5. Development with Hot Reloading

```json
{
  "mcpServers": {
    "hot-reload": {
      "command": "go-go-mcp",
      "args": [
        "start",
        "--profile", "development",
        "--watch"
      ],
      "env": {
        "DEBUG": "true",
        "RELOAD_DELAY": "500ms"
      }
    }
  }
}
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