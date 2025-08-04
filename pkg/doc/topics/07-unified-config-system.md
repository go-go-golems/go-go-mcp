---
Title: Unified Configuration System
Slug: unified-config-system
Short: Complete guide to the unified MCP configuration system for all supported editors.
Topics:
  - config
  - editors
  - mcp
  - servers
Commands:
  - editor
  - ui
Flags:
  - target
  - transport
  - env
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The unified configuration system in go-go-mcp provides a consistent interface for managing MCP server configurations across all supported editors. This system handles the complexity of different editor configurations while providing a unified command structure.

## Table of Contents

1. [Overview](#overview)
2. [Supported Editors](#supported-editors)
3. [Command Structure](#command-structure)
4. [Configuration Verbs](#configuration-verbs)
5. [Target Resolution](#target-resolution)
6. [Transport Types](#transport-types)
7. [Environment Variables](#environment-variables)
8. [Common Use Cases](#common-use-cases)
9. [TUI Interface](#tui-interface)
10. [Troubleshooting](#troubleshooting)

## Overview

The unified configuration system allows you to manage MCP server configurations using a consistent command structure:

```bash
go-go-mcp editor config <editor> <verb> [options]
```

This approach abstracts the differences between editor configuration formats while providing specialized functionality for each editor's unique features.

### Key Benefits

- **Consistent Interface**: Same commands work across all editors
- **Transport Auto-Detection**: Automatically selects appropriate transport type
- **Target Resolution**: Smart handling of global vs local configurations
- **Environment Management**: Unified environment variable handling
- **Editor-Specific Features**: Access to unique features per editor

## Supported Editors

The system supports five major editors with varying feature sets:

| Editor | Command | Transport Support | Target Support | Special Features |
|--------|---------|-------------------|----------------|------------------|
| Claude Desktop | `claude` | stdio | global | Log tailing, detailed server management |
| Cursor | `cursor` | stdio, HTTP, SSE | global, cwd, local | Project-specific configs |
| AmpCode/VS Code | `ampcode` | stdio, HTTP, SSE | global, cwd, local | Workspace integration |
| Standalone Amp | `amp` | stdio, HTTP, SSE | global, cwd, local | Development-focused |
| Crush | `crush` | stdio, HTTP, SSE | global, cwd, local | Lightweight editor |

### Editor Capabilities Matrix

```bash
# Check which editors are available
go-go-mcp editor config --help

# Each editor supports different configuration locations:
# - Claude: Global only (~/.claude/config.json)
# - Cursor: Global, CWD, Local (.cursor.json, cursor.json)
# - AmpCode: Global, CWD, Local (.vscode/settings.json variants)
# - Amp: Global, CWD, Local (.amp/config.json variants)  
# - Crush: Global, CWD, Local (.crush.json, crush.json)
```

## Command Structure

### Basic Syntax

All configuration commands follow this pattern:

```bash
go-go-mcp editor config <editor> <verb> [server-name] [command-path] [options]
```

### Universal Flags

These flags work across all editors and verbs:

- `--target <location>`: Specify configuration target (default, global, cwd, local)
- `--transport <type>`: Override transport type (stdio, http, sse)
- `--env <key=value>`: Add environment variables
- `--args <arguments>`: Specify command arguments
- `--enabled/--disabled`: Set server enabled state

### Editor-Specific Flags

Some editors support additional flags:

```bash
# Claude-specific
--logs-dir <path>     # Specify logs directory location

# Development editors (Cursor, AmpCode, Amp, Crush)
--workspace <path>    # Target specific workspace
--project <name>      # Target specific project
```

## Configuration Verbs

The system provides eight primary verbs that work consistently across all editors:

### add - Add MCP Server

Add a new MCP server configuration:

```bash
# Basic server addition
go-go-mcp editor config claude add myserver /path/to/command

# With custom transport
go-go-mcp editor config cursor add apiserver /path/to/server \
  --transport http \
  --args "--port 3001 --host localhost"

# With environment variables
go-go-mcp editor config amp add dbserver /usr/bin/db-server \
  --env "DB_HOST=localhost" \
  --env "DB_PORT=5432" \
  --target local
```

### remove - Remove MCP Server

Remove an existing server configuration:

```bash
# Remove from default target
go-go-mcp editor config claude remove myserver

# Remove from specific target
go-go-mcp editor config cursor remove oldserver --target global
```

### list - List Servers

Display configured servers:

```bash
# List all servers
go-go-mcp editor config claude list

# List servers for specific target
go-go-mcp editor config cursor list --target local

# Example output:
# Target: global
# ├── myserver (stdio) - enabled
# │   Command: /path/to/command
# │   Args: --config production
# └── apiserver (http) - disabled
#     Command: /path/to/api-server
#     Environment: API_KEY=***
```

### enable/disable - Toggle Server State

Control whether servers are active:

```bash
# Enable a server
go-go-mcp editor config claude enable myserver

# Disable a server
go-go-mcp editor config cursor disable oldserver --target local

# Check current state
go-go-mcp editor config claude list
```

### edit - Edit Configuration

Open configuration files in your default editor:

```bash
# Edit default configuration
go-go-mcp editor config claude edit

# Edit specific target
go-go-mcp editor config cursor edit --target local
```

### init - Initialize Configuration

Create initial configuration structure:

```bash
# Initialize with default settings
go-go-mcp editor config claude init

# Initialize with custom settings
go-go-mcp editor config cursor init --target local
```

### tail - Monitor Logs (Claude Only)

Monitor server logs in real-time:

```bash
# Tail all server logs
go-go-mcp editor config claude tail

# Tail specific server
go-go-mcp editor config claude tail myserver

# Tail with custom logs directory
go-go-mcp editor config claude tail --logs-dir /custom/logs/path
```

### add-go-go - Quick go-go-mcp Setup

Quickly add a go-go-mcp server with sensible defaults:

```bash
# Add with default profile
go-go-mcp editor config claude add-go-go

# Add with specific profile and settings
go-go-mcp editor config cursor add-go-go \
  --args "--profile development --log-level debug" \
  --target local
```

## Target Resolution

The target system determines where configuration files are stored and read from. Different editors support different targets:

### Target Types

1. **default**: Use editor's preferred location (varies by editor)
2. **global**: User-wide configuration
3. **cwd**: Current working directory 
4. **local**: Project-local configuration

### Target Support by Editor

```bash
# Claude: Only supports global
go-go-mcp editor config claude add server cmd  # Always uses global

# Cursor: Supports all targets
go-go-mcp editor config cursor add server cmd --target global    # ~/.cursor/config.json
go-go-mcp editor config cursor add server cmd --target cwd       # ./.cursor.json  
go-go-mcp editor config cursor add server cmd --target local     # ./cursor.json

# AmpCode: Supports all targets  
go-go-mcp editor config ampcode add server cmd --target global   # ~/.vscode/settings.json
go-go-mcp editor config ampcode add server cmd --target cwd      # ./.vscode/settings.json
go-go-mcp editor config ampcode add server cmd --target local    # ./settings.json

# Amp: Supports all targets
go-go-mcp editor config amp add server cmd --target global       # ~/.amp/config.json
go-go-mcp editor config amp add server cmd --target cwd          # ./.amp/config.json
go-go-mcp editor config amp add server cmd --target local        # ./amp.json

# Crush: Supports all targets  
go-go-mcp editor config crush add server cmd --target global     # ~/.crush/config.json
go-go-mcp editor config crush add server cmd --target cwd        # ./.crush.json
go-go-mcp editor config crush add server cmd --target local      # ./crush.json
```

### Target Selection Strategy

When `--target` is not specified, the system uses these defaults:

1. **Claude**: Always uses `global` (only supported target)
2. **Development Editors**: Uses `default` which maps to:
   - `local` if a project config file exists
   - `cwd` if in a project directory
   - `global` otherwise

## Transport Types

The system automatically selects appropriate transport types but allows manual override:

### Supported Transports

1. **stdio**: Standard input/output (most common)
2. **http**: HTTP-based communication
3. **sse**: Server-Sent Events

### Transport Auto-Detection

```bash
# System automatically chooses stdio for most cases
go-go-mcp editor config claude add server /path/to/cmd

# Override to use HTTP
go-go-mcp editor config cursor add apiserver /path/to/server \
  --transport http \
  --args "--port 3001"

# Override to use SSE  
go-go-mcp editor config amp add streaming /path/to/server \
  --transport sse \
  --args "--sse-port 3002"
```

### Transport Capabilities

| Transport | Real-time | Bidirectional | Setup Complexity |
|-----------|-----------|---------------|------------------|
| stdio | Yes | Yes | Low |
| http | No | No | Medium |
| sse | Yes | No | Medium |

## Environment Variables

All editors support environment variable configuration:

### Basic Environment Setup

```bash
# Single environment variable
go-go-mcp editor config claude add dbserver /path/to/server \
  --env "DATABASE_URL=postgresql://localhost/mydb"

# Multiple environment variables
go-go-mcp editor config cursor add apiserver /path/to/server \
  --env "API_KEY=secret123" \
  --env "API_HOST=api.example.com" \
  --env "DEBUG=true"
```

### Environment Variable Patterns

```bash
# Development vs Production
go-go-mcp editor config amp add server /path/to/server \
  --env "NODE_ENV=development" \
  --target local

go-go-mcp editor config amp add server /path/to/server \
  --env "NODE_ENV=production" \
  --target global

# Sensitive Data Handling
go-go-mcp editor config claude add secure-server /path/to/server \
  --env "SECRET_KEY=$(cat /path/to/secret)" \
  --env "CONFIG_FILE=/secure/path/config.json"
```

## Common Use Cases

### Development Workflow

```bash
# 1. Set up local development server
go-go-mcp editor config cursor add dev-server go-go-mcp \
  --args "server start --profile development --log-level debug" \
  --env "CONFIG_FILE=./dev-config.yaml" \
  --target local

# 2. Add production-like staging
go-go-mcp editor config cursor add staging-server go-go-mcp \
  --args "server start --profile staging" \
  --env "CONFIG_FILE=./staging-config.yaml" \
  --target cwd

# 3. Configure global tools
go-go-mcp editor config cursor add tools-server go-go-mcp \
  --args "server start --profile tools" \
  --target global
```

### Multi-Editor Setup

```bash
# Configure the same server across multiple editors
for editor in cursor ampcode amp; do
  go-go-mcp editor config $editor add shared-tools go-go-mcp \
    --args "server start --profile shared" \
    --env "TOOLS_DIR=/shared/tools" \
    --target global
done
```

### Project-Specific Configuration

```bash
# Set up project-specific MCP servers
cd /path/to/project

# Add project-specific tools
go-go-mcp editor config ampcode add project-tools /path/to/tools \
  --args "--project-dir $(pwd)" \
  --env "PROJECT_NAME=$(basename $(pwd))" \
  --target local

# Add development database access
go-go-mcp editor config cursor add project-db ./scripts/db-server.sh \
  --env "DB_NAME=project_dev" \
  --env "DB_CONFIG=./db-config.json" \
  --target local
```

### Server Management

```bash
# Disable servers temporarily
go-go-mcp editor config claude disable debug-server
go-go-mcp editor config cursor disable slow-server --target local

# Re-enable when needed
go-go-mcp editor config claude enable debug-server

# Remove unused servers
go-go-mcp editor config amp remove old-server --target global
```

## TUI Interface

The system includes a terminal user interface for interactive configuration management:

### Launching the TUI

```bash
# Start the interactive configuration UI
go-go-mcp ui
```

### TUI Features

1. **Visual Server Management**
   - Browse servers across all editors
   - View server details and status
   - Enable/disable servers with keystrokes

2. **Configuration Editing**
   - Edit server configurations inline
   - Add new servers through forms
   - Remove servers with confirmation

3. **Yank and Paste**
   - Copy server configurations between editors
   - Duplicate servers with modifications
   - Share configurations across targets

4. **Real-time Updates**
   - See configuration changes immediately
   - Monitor server status changes
   - View live log output (Claude)

### TUI Navigation

```
Key Bindings:
- ↑/↓: Navigate servers
- Enter: Edit selected server
- Space: Toggle server enabled/disabled
- d: Delete server (with confirmation)
- y: Yank (copy) server configuration
- p: Paste yanked configuration
- n: Add new server
- e: Edit configuration file
- q: Quit
- ?: Show help
```

### TUI vs CLI Usage

Use the TUI when:
- Exploring configurations interactively
- Making multiple changes across editors
- Learning the system and available options
- Visualizing complex configuration setups

Use CLI commands when:
- Scripting and automation
- Quick one-off changes
- CI/CD integration
- Shell command integration

## Troubleshooting

### Common Issues

#### 1. Configuration File Not Found

```bash
# Error: configuration file not found
# Solution: Initialize configuration first
go-go-mcp editor config claude init
```

#### 2. Permission Denied

```bash
# Error: permission denied writing config
# Solution: Check file permissions
ls -la ~/.claude/config.json
chmod 644 ~/.claude/config.json
```

#### 3. Server Not Starting

```bash
# Check server configuration
go-go-mcp editor config claude list

# Verify command path exists
which go-go-mcp

# Test server manually
/path/to/server --args-from-config
```

#### 4. Environment Variables Not Set

```bash
# Check current environment setup
go-go-mcp editor config cursor list --target local

# Verify environment variables
env | grep API_KEY

# Update environment variables
go-go-mcp editor config cursor add server cmd \
  --env "API_KEY=new_value"
```

### Debugging Commands

```bash
# List all configurations
for editor in claude cursor ampcode amp crush; do
  echo "=== $editor ==="
  go-go-mcp editor config $editor list 2>/dev/null || echo "Not configured"
done

# Check specific target configurations
go-go-mcp editor config cursor list --target global
go-go-mcp editor config cursor list --target local

# Verify configuration files exist
ls -la ~/.claude/config.json
ls -la ~/.cursor/config.json
ls -la ./.cursor.json
```

### Best Practices

1. **Use Consistent Naming**
   ```bash
   # Good: descriptive names
   go-go-mcp editor config claude add development-tools
   go-go-mcp editor config claude add production-api
   
   # Avoid: unclear names
   go-go-mcp editor config claude add server1
   go-go-mcp editor config claude add test
   ```

2. **Environment-Specific Targets**
   ```bash
   # Development: local/cwd configurations
   go-go-mcp editor config cursor add dev-server cmd --target local
   
   # Shared tools: global configuration
   go-go-mcp editor config cursor add shared-tools cmd --target global
   ```

3. **Documentation**
   ```bash
   # Document complex setups
   echo "# MCP Server Setup" > MCP_SETUP.md
   echo "go-go-mcp editor config cursor add myserver ..." >> MCP_SETUP.md
   ```

4. **Version Control**
   ```bash
   # Include local configs in version control
   git add .cursor.json .vscode/settings.json
   
   # Exclude sensitive data
   echo "*.env" >> .gitignore
   ```

For more detailed information on specific topics, see:
- [Configuration Tutorial](08-config-tutorial.md)
- [Shell Commands](02-shell-commands.md)
- [MCP in Practice](03-mcp-in-practice.md)
