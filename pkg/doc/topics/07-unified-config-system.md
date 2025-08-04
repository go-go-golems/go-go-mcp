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
4. [Structured Output Capabilities](#structured-output-capabilities)
5. [Configuration Verbs](#configuration-verbs)
6. [Target Resolution](#target-resolution)
7. [Transport Types](#transport-types)
8. [Environment Variables](#environment-variables)
9. [Automation and Scripting](#automation-and-scripting)
10. [Data Analysis and Reporting](#data-analysis-and-reporting)
11. [Output Format Reference](#output-format-reference)
12. [Common Use Cases](#common-use-cases)
13. [Advanced Workflows](#advanced-workflows)
14. [TUI Interface](#tui-interface)
15. [Troubleshooting](#troubleshooting)

## Overview

The unified configuration system allows you to manage MCP server configurations using a consistent command structure:

```bash
go-go-mcp editor config <editor> <verb> [options]
```

This approach abstracts the differences between editor configuration formats while providing specialized functionality for each editor's unique features.

### Key Benefits

- **Consistent Interface**: Same commands work across all editors
- **Structured Output**: JSON, YAML, CSV, and table formats for automation
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

### Multi-Editor Support

Commands can target multiple editors simultaneously using comma-separated syntax:

```bash
# Add server to multiple editors
go-go-mcp editor config claude,cursor,amp add shared-server /path/to/cmd

# List configurations from multiple editors
go-go-mcp editor config claude,cursor list

# Remove server from multiple editors
go-go-mcp editor config claude,cursor,amp remove old-server
```

### Universal Flags

These flags work across all editors and verbs:

- `--target <location>`: Specify configuration target (default, global, cwd, local)
- `--transport <type>`: Override transport type (stdio, http, sse)
- `--env <key=value>`: Add environment variables
- `--args <arguments>`: Specify command arguments
- `--enabled/--disabled`: Set server enabled state
- `--target-editor <editor>`: Target editor for copy operations

### Editor-Specific Flags

Some editors support additional flags:

```bash
# Claude-specific
--logs-dir <path>     # Specify logs directory location

# Development editors (Cursor, AmpCode, Amp, Crush)
--workspace <path>    # Target specific workspace
--project <name>      # Target specific project
```

## Structured Output Capabilities

All configuration commands now support structured output formats for automation, scripting, and data analysis. Commands provide both human-readable text output (default) and machine-parseable structured data.

### Available Output Formats

- **table** (default): Human-readable tabular format
- **json**: Machine-parseable JSON format
- **yaml**: YAML format for configuration files
- **csv**: Comma-separated values for spreadsheets

### Output Control Flags

```bash
--output <format>         # Set output format (table, json, yaml, csv)
--fields <field-list>     # Select specific fields to display
--sort-columns <columns>  # Sort output by specified columns
--structured-output       # Enable structured output for dual commands
```

### Command Output Types

#### Pure Structured Commands

The `list` command provides pure structured output:

```bash
# Default table format
go-go-mcp editor config claude list

# JSON output for automation
go-go-mcp editor config claude list --output json

# CSV for spreadsheet import
go-go-mcp editor config claude,cursor list --output csv

# Field selection
go-go-mcp editor config claude list --fields name,command,enabled
```

#### Dual Output Commands

Commands like `add`, `remove`, `copy`, `enable`, and `disable` support both human-readable and structured output:

```bash
# Human-readable output (default)
go-go-mcp editor config claude add server1 /path/to/cmd

# Structured output
go-go-mcp editor config claude add server1 /path/to/cmd --structured-output --output json

# CSV format for bulk operations
go-go-mcp editor config claude,cursor add shared-server /path/to/cmd --structured-output --output csv
```

## Configuration Verbs

The system provides nine primary verbs that work consistently across all editors:

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

### copy - Copy MCP Server

Copy server configurations within the same editor or between different editors:

```bash
# Copy within same editor
go-go-mcp editor config claude copy server1 backup-server1

# Copy between editors
go-go-mcp editor config claude copy server1 server1 --target-editor cursor

# Copy with different target
go-go-mcp editor config cursor copy dev-server prod-server \
  --target global \
  --target-editor amp

# Copy with overwrite
go-go-mcp editor config claude copy old-server new-server --overwrite
```

**Copy Use Cases:**
- **Migration**: Move servers between editors when switching workflows
- **Backup**: Create backup copies of important configurations
- **Standardization**: Ensure consistent server setups across editors
- **Template Creation**: Copy and modify servers for different environments

**Transport Compatibility**: The copy operation automatically handles transport compatibility between editors. If the target editor doesn't support the source transport type, it will fall back to `stdio`.

### remove - Remove MCP Server

Remove an existing server configuration:

```bash
# Remove from default target
go-go-mcp editor config claude remove myserver

# Remove from specific target
go-go-mcp editor config cursor remove oldserver --target global
```

### list - List Servers

Display configured servers with full structured output support:

```bash
# Default table format (human-readable)
go-go-mcp editor config claude list

# JSON output for automation and scripting
go-go-mcp editor config claude list --output json

# CSV output for spreadsheets and reporting
go-go-mcp editor config claude,cursor list --output csv

# YAML output for configuration management
go-go-mcp editor config claude list --output yaml

# Select specific fields
go-go-mcp editor config claude list --fields name,command,enabled

# Sort by column
go-go-mcp editor config claude list --sort-columns enabled,name

# Multi-editor structured output
go-go-mcp editor config claude,cursor,amp list --output json
```

**Example Table Output:**
```
Target: global
├── myserver (stdio) - enabled
│   Command: /path/to/command
│   Args: --config production
└── apiserver (http) - disabled
    Command: /path/to/api-server
    Environment: API_KEY=***
```

**Example JSON Output:**
```json
[
  {
    "editor": "claude",
    "name": "myserver",
    "command": "/path/to/command",
    "args": ["--config", "production"],
    "env": {},
    "url": "",
    "transport": "stdio",
    "enabled": true,
    "target": "global"
  }
]
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

## Automation and Scripting

The structured output capabilities enable powerful automation and scripting workflows. Commands provide machine-parseable data without sacrificing human readability.

### JSON Output for Scripts

Extract specific information using JSON output and jq:

```bash
# Get all enabled servers
go-go-mcp editor config claude list --output json | jq '.[] | select(.enabled == true)'

# Extract server names only
go-go-mcp editor config claude,cursor list --output json | jq -r '.[].name'

# Find servers using specific commands
go-go-mcp editor config claude list --output json | jq '.[] | select(.command | contains("go-go-mcp"))'

# Count servers by editor
go-go-mcp editor config claude,cursor,amp list --output json | jq 'group_by(.editor) | map({editor: .[0].editor, count: length})'
```

### Bulk Operations with Structured Output

Perform bulk operations and track results:

```bash
# Add servers to multiple editors with JSON tracking
go-go-mcp editor config claude,cursor,amp add bulk-server /path/to/cmd \
  --structured-output --output json > operation_results.json

# Track which operations succeeded
cat operation_results.json | jq '.[] | select(.success == true) | .editor'

# Export operations to CSV for reporting
go-go-mcp editor config claude,cursor add monitoring-server /path/to/monitor \
  --structured-output --output csv >> operations_log.csv
```

### Configuration Monitoring

Monitor configuration changes over time:

```bash
#!/bin/bash
# config-monitor.sh - Track configuration changes

TIMESTAMP=$(date '+%Y-%m-%d_%H-%M-%S')
OUTPUT_DIR="config-snapshots"
mkdir -p "$OUTPUT_DIR"

# Snapshot all configurations
for editor in claude cursor ampcode amp crush; do
  go-go-mcp editor config $editor list --output json > "$OUTPUT_DIR/${editor}_${TIMESTAMP}.json" 2>/dev/null || true
done

echo "Configuration snapshot saved to $OUTPUT_DIR"
```

### Environment-Specific Deployments

Deploy different configurations per environment:

```bash
#!/bin/bash
# deploy-config.sh - Deploy environment-specific configurations
ENVIRONMENT=${1:-development}

case $ENVIRONMENT in
  development)
    go-go-mcp editor config cursor,amp add dev-tools go-go-mcp \
      --args "server start --profile development" \
      --env "LOG_LEVEL=debug" \
      --target local \
      --structured-output --output json
    ;;
  production)
    go-go-mcp editor config claude,cursor add prod-tools go-go-mcp \
      --args "server start --profile production" \
      --env "LOG_LEVEL=error" \
      --target global \
      --structured-output --output json
    ;;
esac
```

## Data Analysis and Reporting

Structured output enables comprehensive analysis and reporting of MCP server configurations.

### Configuration Analysis

Analyze your MCP setup across editors:

```bash
# Export all configurations to CSV for analysis
go-go-mcp editor config claude,cursor,ampcode,amp,crush list --output csv > all_configs.csv

# Generate configuration summary
go-go-mcp editor config claude,cursor,amp list --output json | \
jq '{
  total: length,
  enabled: [.[] | select(.enabled == true)] | length,
  disabled: [.[] | select(.enabled == false)] | length,
  editors: [.[] | .editor] | unique,
  transports: [.[] | .transport] | unique,
  targets: [.[] | .target] | unique
}'
```

### Usage Reporting

Generate usage reports for team or compliance purposes:

```bash
#!/bin/bash
# generate-report.sh - Generate MCP configuration report

echo "# MCP Configuration Report - $(date)"
echo

# Summary statistics
echo "## Summary"
go-go-mcp editor config claude,cursor,ampcode,amp list --output json | \
jq -r '
  group_by(.editor) | 
  map("\(.[]|.editor): \(length) servers") | 
  .[]
'

echo
echo "## Enabled Servers by Editor"
go-go-mcp editor config claude,cursor,ampcode,amp list --output json | \
jq -r '
  map(select(.enabled == true)) |
  group_by(.editor) |
  map("### \(.[0].editor)\n" + (map("- \(.name): \(.command)") | join("\n"))) |
  join("\n\n")
'

# Export detailed CSV
go-go-mcp editor config claude,cursor,ampcode,amp list --output csv > mcp_report.csv
echo
echo "Detailed CSV report saved to: mcp_report.csv"
```

### Health Monitoring

Monitor server health and configuration drift:

```bash
# Check for configuration inconsistencies
go-go-mcp editor config claude,cursor,amp list --output json | \
jq 'group_by(.name) | map(select(length > 1)) | 
    map({name: .[0].name, editors: [.[] | .editor], commands: [.[] | .command] | unique})'

# Find disabled servers that might need attention
go-go-mcp editor config claude,cursor,amp list --output json | \
jq '.[] | select(.enabled == false) | {editor, name, command}'
```

## Output Format Reference

### List Command Schema

The `list` command outputs the following fields:

```json
{
  "editor": "string",      // Editor name (claude, cursor, etc.)
  "name": "string",        // Server name
  "command": "string",     // Command path
  "args": ["string"],      // Command arguments array
  "env": {"key": "value"}, // Environment variables object
  "url": "string",         // Server URL (if applicable)
  "transport": "string",   // Transport type (stdio, http, sse)
  "enabled": boolean,      // Server enabled state
  "target": "string"       // Configuration target (global, local, cwd)
}
```

### Operation Command Schema

Commands like `add`, `remove`, `copy`, `enable`, and `disable` output:

```json
{
  "operation": "string",    // Operation performed (add, remove, copy, etc.)
  "editor": "string",       // Target editor
  "server_name": "string",  // Server name
  "success": boolean,       // Operation success status
  "error": "string",        // Error message (if success is false)
  "server": {               // Server details (same as list schema)
    "name": "string",
    "command": "string",
    "args": ["string"],
    "env": {"key": "value"},
    "url": "string",
    "transport": "string",
    "enabled": boolean,
    "target": "string"
  }
}
```

### CSV Format

CSV output includes all fields as columns with proper escaping:

```csv
editor,name,command,args,env,url,transport,enabled,target
claude,myserver,/path/to/cmd,"arg1,arg2","KEY1=val1,KEY2=val2",,stdio,true,global
```

### Field Selection

Use `--fields` to select specific columns:

```bash
# Common field combinations
--fields name,enabled                    # Basic status
--fields editor,name,command            # Server overview  
--fields name,transport,target          # Configuration details
--fields editor,name,enabled,target     # Multi-editor status
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

Using the new multi-editor syntax for streamlined configuration:

```bash
# Configure the same server across multiple editors (new syntax)
go-go-mcp editor config cursor,ampcode,amp add shared-tools go-go-mcp \
  --args "server start --profile shared" \
  --env "TOOLS_DIR=/shared/tools" \
  --target global

# Traditional loop approach (still supported)
for editor in cursor ampcode amp; do
  go-go-mcp editor config $editor add shared-tools go-go-mcp \
    --args "server start --profile shared" \
    --env "TOOLS_DIR=/shared/tools" \
    --target global
done

# Bulk operations across editors
go-go-mcp editor config claude,cursor,amp list
go-go-mcp editor config cursor,ampcode remove old-server
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

## Advanced Workflows

### Bulk Operations

Multi-editor support enables powerful bulk operations:

```bash
# Set up team development environment across all editors
go-go-mcp editor config claude,cursor,ampcode,amp add team-tools go-go-mcp \
  --args "server start --profile team" \
  --env "TEAM_CONFIG=/shared/team-config.yaml" \
  --target global

# Update configurations across multiple editors
go-go-mcp editor config cursor,ampcode,amp remove outdated-server
go-go-mcp editor config claude,cursor,amp enable production-tools

# Check configurations across all editors
go-go-mcp editor config claude,cursor,ampcode,amp,crush list
```

### Server Migration

Use copy operations to migrate servers between editors or create backups:

```bash
# Migrate from Claude to Cursor when switching editors
go-go-mcp editor config claude copy my-tools my-tools --target-editor cursor
go-go-mcp editor config claude copy api-server api-server --target-editor cursor

# Create backup before major changes
go-go-mcp editor config cursor copy production-server backup-prod-server

# Standardize configuration across editors
go-go-mcp editor config claude copy reference-server reference-server --target-editor cursor
go-go-mcp editor config claude copy reference-server reference-server --target-editor amp

# Copy with target modification for different environments
go-go-mcp editor config cursor copy dev-server staging-server \
  --target global \
  --target-editor ampcode
```

### Template-Based Configuration

Create reusable server templates:

```bash
# Create template server
go-go-mcp editor config claude add template-server go-go-mcp \
  --args "server start --profile template" \
  --env "ENV=template" \
  --disabled

# Use template to create environment-specific servers
go-go-mcp editor config claude copy template-server dev-server
go-go-mcp editor config claude copy template-server staging-server \
  --target-editor cursor

# Copy template across editors for consistency
go-go-mcp editor config claude copy template-server template-server \
  --target-editor cursor,ampcode,amp
```

### Cross-Editor Standardization

Ensure consistent configurations across development environments:

```bash
# Set up standard development tools across all editors
go-go-mcp editor config claude,cursor,ampcode,amp add standard-dev go-go-mcp \
  --args "server start --profile development" \
  --env "LOG_LEVEL=debug" \
  --target global

# Copy project-specific configuration to all editors
go-go-mcp editor config cursor copy project-server project-server \
  --target-editor claude,ampcode,amp \
  --target local

# Synchronize configuration changes
go-go-mcp editor config claude copy updated-server updated-server \
  --target-editor cursor,ampcode,amp \
  --overwrite
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
