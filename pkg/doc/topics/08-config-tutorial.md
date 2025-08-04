---
Title: Configuration Tutorial
Slug: config-tutorial
Short: Step-by-step tutorial for getting started with the unified MCP configuration system.
Topics:
  - config
  - tutorial
  - getting-started
  - editors
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

This tutorial will guide you through setting up and managing MCP server configurations using the unified configuration system. You'll learn how to configure servers across different editors, manage multiple environments, and use advanced features.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [First Time Setup](#first-time-setup)
3. [Adding Your First Server](#adding-your-first-server)
4. [Managing Multiple Editors](#managing-multiple-editors)
5. [Working with Different Targets](#working-with-different-targets)
6. [Environment Variables and Arguments](#environment-variables-and-arguments)
7. [Copying and Migration](#copying-and-migration)
8. [Multi-Editor Workflows](#multi-editor-workflows)
9. [Using the TUI](#using-the-tui)
10. [Advanced Scenarios](#advanced-scenarios)
11. [Best Practices](#best-practices)
12. [Troubleshooting Guide](#troubleshooting-guide)

## Prerequisites

Before starting, ensure you have:

1. **go-go-mcp installed**: The unified configuration system is part of go-go-mcp
2. **At least one supported editor**: Claude Desktop, Cursor, AmpCode/VS Code, Standalone Amp, or Crush
3. **Basic command line knowledge**: Familiarity with terminal/command prompt

### Checking Your Setup

```bash
# Verify go-go-mcp is installed
go-go-mcp --version

# Check available editors
go-go-mcp editor config --help

# See which editors you have installed
which claude-desktop cursor code amp crush
```

## First Time Setup

Let's start with Claude Desktop as it's the simplest editor to configure (only supports global configuration).

### Step 1: Initialize Claude Configuration

```bash
# Initialize Claude Desktop configuration
go-go-mcp editor config claude init

# Verify initialization worked
go-go-mcp editor config claude list
```

Expected output:
```
Target: global
No servers configured
Configuration file: /home/user/.claude/config.json
```

### Step 2: Check Configuration File

```bash
# View the configuration file that was created
go-go-mcp editor config claude edit
```

This opens your default editor with the configuration file. You should see something like:
```json
{
  "mcpServers": {}
}
```

## Adding Your First Server

Now let's add your first MCP server to Claude Desktop.

### Basic Server Addition

```bash
# Add a simple server with go-go-mcp
go-go-mcp editor config claude add my-first-server go-go-mcp \
  --args "server start --profile default"

# Verify it was added
go-go-mcp editor config claude list
```

Expected output:
```
Target: global
├── my-first-server (stdio) - enabled
│   Command: go-go-mcp
│   Args: server start --profile default
```

### Testing Your Server

1. **Restart Claude Desktop** to load the new configuration
2. **Start a new conversation** in Claude
3. **Look for MCP indicators** showing tools are available
4. **Try using a tool** by asking Claude to use an available MCP tool

### Understanding What Happened

The command above:
1. Created a server entry named `my-first-server`
2. Set the command to `go-go-mcp` 
3. Added arguments to start the server with the default profile
4. Used the `stdio` transport (automatically detected)
5. Enabled the server by default

## Managing Multiple Editors

Now let's expand to managing configurations across multiple editors.

### Adding the Same Server to Cursor

```bash
# Initialize Cursor configuration
go-go-mcp editor config cursor init

# Add the same server to Cursor
go-go-mcp editor config cursor add my-first-server go-go-mcp \
  --args "server start --profile default"

# List both configurations
echo "=== Claude ==="
go-go-mcp editor config claude list
echo "=== Cursor ==="
go-go-mcp editor config cursor list
```

### Bulk Configuration

You can configure multiple editors at once using shell loops:

```bash
# Configure the same server across multiple editors
for editor in cursor ampcode amp; do
  echo "Configuring $editor..."
  go-go-mcp editor config $editor init 2>/dev/null || true
  go-go-mcp editor config $editor add shared-tools go-go-mcp \
    --args "server start --profile tools"
done

# Verify all configurations
for editor in claude cursor ampcode amp; do
  echo "=== $editor ==="
  go-go-mcp editor config $editor list 2>/dev/null || echo "Not configured"
done
```

## Working with Different Targets

Unlike Claude (which only supports global configuration), other editors support multiple configuration targets.

### Understanding Targets

```bash
# Global: Affects all projects for this user
go-go-mcp editor config cursor add global-tools go-go-mcp \
  --args "server start --profile tools" \
  --target global

# Current directory: Affects this directory and subdirectories  
go-go-mcp editor config cursor add project-tools go-go-mcp \
  --args "server start --profile development" \
  --target cwd

# Local: Affects only this specific directory
go-go-mcp editor config cursor add local-tools go-go-mcp \
  --args "server start --profile local" \
  --target local
```

### Target Priority

When multiple targets are configured, editors typically use this priority order:
1. **local** (highest priority)
2. **cwd** 
3. **global** (lowest priority)

### Viewing Target-Specific Configurations

```bash
# View all targets for Cursor
go-go-mcp editor config cursor list --target global
go-go-mcp editor config cursor list --target cwd  
go-go-mcp editor config cursor list --target local

# Compare configurations
echo "=== Global ==="
go-go-mcp editor config cursor list --target global
echo "=== Local ==="
go-go-mcp editor config cursor list --target local
```

## Environment Variables and Arguments

MCP servers often need configuration through environment variables and command-line arguments.

### Basic Environment Variables

```bash
# Add a server with environment variables
go-go-mcp editor config claude add api-server go-go-mcp \
  --args "server start --profile api" \
  --env "API_HOST=localhost" \
  --env "API_PORT=3001" \
  --env "DEBUG=true"

# Verify the configuration
go-go-mcp editor config claude list
```

### Complex Configuration Example

```bash
# Database server with multiple environment variables
go-go-mcp editor config cursor add database-server ./scripts/db-mcp-server.sh \
  --env "DB_HOST=localhost" \
  --env "DB_PORT=5432" \
  --env "DB_NAME=development" \
  --env "DB_USER=dev_user" \
  --env "DB_SSLMODE=disable" \
  --target local
```

### Sensitive Data Handling

```bash
# For sensitive data, use environment variable references
go-go-mcp editor config claude add secure-api /path/to/server \
  --env "API_KEY=$(echo $MY_API_KEY)" \
  --env "SECRET_CONFIG=/path/to/secrets.json"

# Or reference external files
go-go-mcp editor config cursor add secure-server ./server.sh \
  --env "CONFIG_FILE=./config/production.env" \
  --target global
```

## Copying and Migration

The copy functionality allows you to duplicate server configurations within the same editor or migrate them between different editors.

### Basic Copying

```bash
# Copy within the same editor (Claude)
go-go-mcp editor config claude copy my-first-server backup-server

# Verify the copy was created
go-go-mcp editor config claude list
```

Expected output:
```
Target: global
├── my-first-server (stdio) - enabled
│   Command: go-go-mcp
│   Args: server start --profile default
└── backup-server (stdio) - enabled
    Command: go-go-mcp
    Args: server start --profile default
```

### Cross-Editor Migration

Let's migrate a server from Claude to Cursor:

```bash
# Copy server from Claude to Cursor
go-go-mcp editor config claude copy my-first-server my-first-server --target-editor cursor

# Verify it was copied to Cursor
go-go-mcp editor config cursor list
```

### Migration Workflow Exercise

Follow this step-by-step migration workflow:

```bash
# 1. Set up a development server in Claude
go-go-mcp editor config claude add dev-tools go-go-mcp \
  --args "server start --profile development" \
  --env "LOG_LEVEL=debug"

# 2. Test that it works in Claude, then migrate to Cursor
go-go-mcp editor config claude copy dev-tools dev-tools --target-editor cursor

# 3. Customize the Cursor version with local configuration
go-go-mcp editor config cursor copy dev-tools local-dev-tools --target local

# 4. Verify all configurations
echo "=== Claude ==="
go-go-mcp editor config claude list
echo "=== Cursor Global ==="
go-go-mcp editor config cursor list --target global
echo "=== Cursor Local ==="
go-go-mcp editor config cursor list --target local
```

### Copy Use Cases

**Backup Before Changes**:
```bash
# Create backup before modifying important server
go-go-mcp editor config cursor copy production-server backup-prod-$(date +%Y%m%d)

# Now safely modify the original
go-go-mcp editor config cursor remove production-server
go-go-mcp editor config cursor add production-server /new/path/to/server
```

**Template Creation**:
```bash
# Create a template with common settings
go-go-mcp editor config claude add template-server go-go-mcp \
  --args "server start --profile template" \
  --env "TEMPLATE=true" \
  --disabled

# Use template to create environment-specific servers
go-go-mcp editor config claude copy template-server dev-server
go-go-mcp editor config claude copy template-server staging-server
```

**Editor Migration**:
```bash
# When switching from Claude to Cursor, migrate all servers
go-go-mcp editor config claude list --format json > claude-servers.json

# Copy each server (manual approach for complex configurations)
go-go-mcp editor config claude copy server1 server1 --target-editor cursor
go-go-mcp editor config claude copy server2 server2 --target-editor cursor
```

## Multi-Editor Workflows

The multi-editor syntax allows you to manage configurations across multiple editors simultaneously, making it easy to standardize setups.

### Basic Multi-Editor Commands

```bash
# Add the same server to multiple editors
go-go-mcp editor config claude,cursor,amp add shared-tools go-go-mcp \
  --args "server start --profile shared" \
  --env "SHARED_CONFIG=true"

# List configurations from multiple editors
go-go-mcp editor config claude,cursor,amp list

# Remove server from multiple editors
go-go-mcp editor config cursor,amp remove old-server
```

### Team Development Setup

Set up a consistent development environment across all editors:

```bash
# 1. Initialize all editors that support multi-target
for editor in cursor ampcode amp crush; do
  go-go-mcp editor config $editor init 2>/dev/null || true
done

# 2. Set up team tools across all editors
go-go-mcp editor config claude,cursor,ampcode,amp add team-tools go-go-mcp \
  --args "server start --profile team" \
  --env "TEAM_CONFIG=/shared/config/team.yaml" \
  --target global

# 3. Add development-specific tools to development editors
go-go-mcp editor config cursor,ampcode,amp add dev-tools go-go-mcp \
  --args "server start --profile development" \
  --env "DEV_MODE=true" \
  --target local

# 4. Verify setup across all editors
echo "=== Team Setup Verification ==="
go-go-mcp editor config claude,cursor,ampcode,amp list | grep -E "(===|team-tools)"
```

### Multi-Editor Maintenance

```bash
# Update configuration across multiple editors
go-go-mcp editor config cursor,ampcode,amp remove outdated-server
go-go-mcp editor config claude,cursor,amp enable new-production-tools

# Check for inconsistencies
echo "Checking configurations across editors..."
for editor in claude cursor ampcode amp; do
  echo "=== $editor ==="
  go-go-mcp editor config $editor list --format brief 2>/dev/null || echo "Not configured"
done
```

### Selective Multi-Editor Operations

```bash
# Development editors only (editors that support local targets)
go-go-mcp editor config cursor,ampcode,amp add project-tools ./tools/server.sh \
  --target local

# All editors including Claude
go-go-mcp editor config claude,cursor,ampcode,amp add global-utilities go-go-mcp \
  --args "server start --profile utilities" \
  --target global

# Subset operations based on use case
go-go-mcp editor config cursor,amp add experimental-server ./experimental/server \
  --disabled
```

## Using the TUI

The Terminal User Interface (TUI) provides an interactive way to manage configurations.

### Launching the TUI

```bash
# Start the interactive configuration interface
go-go-mcp ui
```

### TUI Tutorial

1. **Navigation**:
   - Use ↑/↓ arrows to navigate between servers
   - Use Tab to switch between editor sections
   - Use Enter to select/edit items

2. **Adding Servers**:
   - Press `n` to add a new server
   - Fill in the form fields
   - Press Enter to save

3. **Editing Servers**:
   - Navigate to a server and press Enter
   - Modify fields as needed
   - Save with Enter, cancel with Esc

4. **Copying Configurations**:
   - Navigate to a server and press `y` to yank (copy)
   - Navigate to another editor section
   - Press `p` to paste the configuration

5. **Managing Server State**:
   - Press Space to toggle enabled/disabled
   - Press `d` to delete (with confirmation)

### TUI Exercise

1. Launch the TUI: `go-go-mcp ui`
2. Navigate to any existing server
3. Press `y` to copy it
4. Navigate to a different editor section
5. Press `p` to paste it with a new name
6. Press Space to toggle its enabled state
7. Press `q` to quit

## Advanced Scenarios

### Project-Specific Development Setup

Let's set up a complete development environment for a project:

```bash
# Navigate to your project directory
cd /path/to/your/project

# 1. Set up local development server
go-go-mcp editor config cursor add dev-server go-go-mcp \
  --args "server start --profile development --log-level debug" \
  --env "PROJECT_ROOT=$(pwd)" \
  --env "CONFIG_FILE=./dev-config.yaml" \
  --target local

# 2. Add project-specific tools
go-go-mcp editor config cursor add project-tools ./scripts/mcp-tools.sh \
  --env "PROJECT_NAME=$(basename $(pwd))" \
  --env "TOOLS_DIR=./tools" \
  --target local

# 3. Add database access (if applicable)
go-go-mcp editor config cursor add project-db ./scripts/db-server.sh \
  --env "DATABASE_URL=postgresql://localhost/$(basename $(pwd))_dev" \
  --target local
```

### Multi-Environment Configuration

Set up different servers for different environments:

```bash
# Development environment (local target)
go-go-mcp editor config ampcode add dev-env go-go-mcp \
  --args "server start --profile development" \
  --env "NODE_ENV=development" \
  --env "LOG_LEVEL=debug" \
  --target local

# Staging environment (cwd target)  
go-go-mcp editor config ampcode add staging-env go-go-mcp \
  --args "server start --profile staging" \
  --env "NODE_ENV=staging" \
  --env "LOG_LEVEL=info" \
  --target cwd

# Production-like tools (global target)
go-go-mcp editor config ampcode add production-tools go-go-mcp \
  --args "server start --profile production" \
  --env "NODE_ENV=production" \
  --env "LOG_LEVEL=error" \
  --target global
```

### Transport Type Customization

Configure servers with different transport types:

```bash
# stdio transport (default, most compatible)
go-go-mcp editor config amp add stdio-server go-go-mcp \
  --args "server start --transport stdio"

# HTTP transport for API-like access
go-go-mcp editor config amp add http-server go-go-mcp \
  --args "server start --transport http --port 3001" \
  --transport http

# SSE transport for real-time features
go-go-mcp editor config amp add sse-server go-go-mcp \
  --args "server start --transport sse --port 3002" \
  --transport sse
```

## Best Practices

### 1. Naming Conventions

Use descriptive, consistent names:

```bash
# Good naming examples
go-go-mcp editor config claude add development-tools go-go-mcp
go-go-mcp editor config claude add production-database ./db-server
go-go-mcp editor config claude add project-assistant ./assistant-server

# Avoid unclear names
go-go-mcp editor config claude add server1 go-go-mcp
go-go-mcp editor config claude add test ./server
```

### 2. Copy and Migration Strategy

Plan your copying and migration workflows:

```bash
# Always backup before major changes
go-go-mcp editor config cursor copy important-server backup-important-$(date +%Y%m%d)

# Use descriptive names for copies
go-go-mcp editor config claude copy base-server dev-modified-server
go-go-mcp editor config claude copy base-server staging-server --target-editor cursor

# Create templates for repeated configurations
go-go-mcp editor config claude add template-base go-go-mcp \
  --args "server start --profile template" \
  --disabled
```

### 3. Multi-Editor Management

Leverage multi-editor syntax for consistency:

```bash
# Use multi-editor syntax for common operations
go-go-mcp editor config claude,cursor,amp add shared-utilities go-go-mcp \
  --args "server start --profile utilities" \
  --target global

# But use specific editors for specialized configurations
go-go-mcp editor config cursor add cursor-specific-tools ./cursor-tools \
  --target local

# Document which editors are used in your team
echo "# Team uses: Claude (main), Cursor (development), Amp (experiments)" > MCP_EDITORS.md
```

### 4. Environment Organization

Organize servers by environment and scope:

```bash
# Global tools (available everywhere)
go-go-mcp editor config cursor add global-utilities go-go-mcp \
  --args "server start --profile utilities" \
  --target global

# Project tools (specific to current project)
go-go-mcp editor config cursor add project-tools go-go-mcp \
  --args "server start --profile project" \
  --target local

# Development tools (temporary/experimental)
go-go-mcp editor config cursor add dev-experiments go-go-mcp \
  --args "server start --profile experimental" \
  --target local
```

### 5. Configuration Documentation

Document your setups:

```bash
# Create a setup script for your project
cat > setup-mcp.sh << 'EOF'
#!/bin/bash
# MCP Server Setup for Project XYZ
# Run this script to configure MCP servers for development

echo "Setting up MCP servers for development..."

# Development server
go-go-mcp editor config cursor add project-dev go-go-mcp \
  --args "server start --profile development" \
  --env "PROJECT_ROOT=$(pwd)" \
  --target local

# Database tools
go-go-mcp editor config cursor add db-tools ./scripts/db-server.sh \
  --env "DB_NAME=project_dev" \
  --target local

echo "MCP servers configured. Restart your editor to use them."
EOF

chmod +x setup-mcp.sh
```

### 6. Version Control Integration

Include relevant configurations in version control:

```bash
# Include local editor configs
git add .cursor.json .vscode/settings.json

# Create .gitignore entries for sensitive data
echo "# MCP secrets" >> .gitignore
echo "*.env.local" >> .gitignore
echo "config/secrets.json" >> .gitignore

# Document the setup in README
echo "" >> README.md
echo "## MCP Configuration" >> README.md
echo "Run \`./setup-mcp.sh\` to configure MCP servers for development." >> README.md
```

## Troubleshooting Guide

### Common Issues and Solutions

#### 1. Server Not Appearing in Editor

**Problem**: Added server but it doesn't show up in the editor.

**Solutions**:
```bash
# Check the configuration was saved
go-go-mcp editor config claude list

# Restart the editor completely
# (Close all windows, quit application, restart)

# Check the configuration file directly
go-go-mcp editor config claude edit
```

#### 2. Permission Errors

**Problem**: Getting permission denied errors.

**Solutions**:
```bash
# Check file permissions
ls -la ~/.claude/config.json

# Fix permissions if needed
chmod 644 ~/.claude/config.json

# Check directory permissions
ls -la ~/.claude/
```

#### 3. Command Not Found

**Problem**: MCP server command not found.

**Solutions**:
```bash
# Verify the command exists
which go-go-mcp

# Use full path if needed
go-go-mcp editor config claude add server /full/path/to/go-go-mcp \
  --args "server start"

# Check PATH environment
echo $PATH
```

#### 4. Environment Variables Not Working

**Problem**: Environment variables not being passed to server.

**Solutions**:
```bash
# Check current configuration
go-go-mcp editor config claude list

# Verify environment variables are set
env | grep YOUR_VAR

# Update the server configuration
go-go-mcp editor config claude remove problematic-server
go-go-mcp editor config claude add problematic-server go-go-mcp \
  --args "server start" \
  --env "YOUR_VAR=correct_value"
```

#### 5. Multiple Targets Confusion

**Problem**: Not sure which target is being used.

**Solutions**:
```bash
# Check all targets
go-go-mcp editor config cursor list --target global
go-go-mcp editor config cursor list --target cwd  
go-go-mcp editor config cursor list --target local

# Use specific target explicitly
go-go-mcp editor config cursor add server cmd --target local

# Check configuration files directly
ls -la ~/.cursor/config.json    # global
ls -la ./.cursor.json           # cwd
ls -la ./cursor.json            # local
```

#### 6. Copy Operation Failures

**Problem**: Copy command fails or creates unexpected results.

**Solutions**:
```bash
# Check source server exists
go-go-mcp editor config claude list | grep source-server

# Verify target editor supports the operation
go-go-mcp editor config cursor init

# Check for naming conflicts
go-go-mcp editor config cursor list | grep target-name

# Use overwrite flag if intentional
go-go-mcp editor config claude copy source target --target-editor cursor --overwrite
```

#### 7. Multi-Editor Command Issues

**Problem**: Multi-editor commands fail for some editors.

**Solutions**:
```bash
# Test each editor individually
for editor in claude cursor amp; do
  echo "Testing $editor..."
  go-go-mcp editor config $editor list 2>/dev/null || echo "$editor not configured"
done

# Initialize missing editors
go-go-mcp editor config cursor init
go-go-mcp editor config amp init

# Use subset that works
go-go-mcp editor config claude,cursor add working-server go-go-mcp
```

#### 8. Transport Compatibility Issues

**Problem**: Copied servers don't work due to transport incompatibility.

**Solutions**:
```bash
# Check transport support for target editor
# Claude: stdio only
# Others: stdio, http, sse

# Copy will automatically fall back to stdio
go-go-mcp editor config cursor copy http-server http-server --target-editor claude

# Manually specify transport if needed
go-go-mcp editor config amp add server cmd --transport stdio
```

### Debugging Commands

```bash
# Comprehensive configuration check
echo "=== Configuration Overview ==="
for editor in claude cursor ampcode amp crush; do
  echo "--- $editor ---"
  go-go-mcp editor config $editor list 2>/dev/null || echo "Not configured"
done

# Multi-editor diagnostics
echo "=== Multi-Editor Test ==="
go-go-mcp editor config claude,cursor,amp list 2>/dev/null || echo "Multi-editor operation failed"

# Copy operation test  
echo "=== Copy Operation Test ==="
go-go-mcp editor config claude copy test-server test-copy 2>/dev/null && echo "Copy works" || echo "Copy failed"

# Check file system
echo "=== Configuration Files ==="
find ~ -name "*claude*config*" -o -name "*cursor*" -o -name "*.vscode*" 2>/dev/null
find . -name "*cursor*" -o -name "*.vscode*" -o -name "*amp*" 2>/dev/null

# Test server command manually
echo "=== Manual Server Test ==="
go-go-mcp server start --help
```

### Getting Help

If you encounter issues not covered here:

1. **Check the logs**: Many editors provide MCP-specific logs
2. **Use the TUI**: `go-go-mcp ui` provides visual feedback
3. **Test manually**: Run server commands manually to isolate issues
4. **Check documentation**: See [Unified Configuration System](07-unified-config-system.md)

### Recovery Commands

If configurations become corrupted:

```bash
# Back up existing configuration
cp ~/.claude/config.json ~/.claude/config.json.backup

# Start fresh
go-go-mcp editor config claude init

# Restore servers one by one
go-go-mcp editor config claude add server1 command1
go-go-mcp editor config claude add server2 command2
```

For more detailed information, see:
- [Unified Configuration System](07-unified-config-system.md)
- [MCP in Practice](03-mcp-in-practice.md)
- [Configuration File Guide](01-config-file.md)
