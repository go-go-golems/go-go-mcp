# go-go-mcp

A Go implementation of the Model Context Protocol (MCP), providing a framework for building MCP servers and clients.

### Installation

To install the `go-go-mcp` command line tool with homebrew, run:

```bash
brew tap go-go-golems/go-go-go
brew install go-go-golems/go-go-go/go-go-mcp
```

To install the `go-go-mcp` command using apt-get, run:

```bash
echo "deb [trusted=yes] https://apt.fury.io/go-go-golems/ /" >> /etc/apt/sources.list.d/fury.list
apt-get update
apt-get install go-go-mcp
```

To install using `yum`, run:

```bash
echo "
[fury]
name=Gemfury Private Repo
baseurl=https://yum.fury.io/go-go-golems/
enabled=1
gpgcheck=0
" >> /etc/yum.repos.d/fury.repo
yum install go-go-mcp
```

To install using `go get`, run:

```bash
go get -u github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp
```

Finally, install by downloading the binaries straight from [github](https://github.com/go-go-golems/go-go-mcp/releases).

## Overview

This project implements the [Model Context Protocol](https://github.com/modelcontextprotocol/specification), which enables standardized communication between AI applications and language models. The implementation includes:

- MCP server and client
- Support for SSE and stdio transports
- Bridge to expose an SSE server as a stdio server
- Templated shell scripts as MCP tools
- Configuration files for profiles

- Core protocol message types and interfaces
- A modular registry system for managing prompts, resources, and tools
- Support for custom handlers and subscriptions
- 

### Supported Methods

The server implements the MCP specification methods:

- `initialize` - Protocol initialization and capability negotiation
- `ping` - Connection health check
- `prompts/list` - List available prompts
- `prompts/get` - Retrieve prompt content
- `resources/list` - List available resources
- `resources/read` - Read resource content
- `tools/list` - List available tools
- `tools/call` - Execute a tool

Not yet implemented:
- notifications
- `resources/subscribe` - Subscribe to resource updates

## Running

### Basic Usage

The binary can be run in a few different modes:
- Server mode: Run as an MCP server process
- Client mode: Use client commands to interact with an MCP server
- Bridge mode: Expose an SSE server as a stdio server

#### Server Mode

Start the server with either stdio or SSE transport:

```bash
# Start with stdio transport (default)
go-go-mcp start --transport stdio

# Start with SSE transport
go-go-mcp start --transport sse --port 3001
```

The server automatically watches configured repositories and files for changes, reloading tools when:
- Files are added or removed from watched directories
- Tool configuration files are modified
- Repository structure changes

This allows for dynamic tool updates without server restarts.

#### Server Tools

You can interact with tools directly without starting a server using the `server tools` commands:

```bash
# List available tools using the 'all' profile
go-go-mcp server tools list --profile all

# List only system monitoring tools
go-go-mcp server tools list --profile system

# Call a tool directly
go-go-mcp server tools call system-monitor --args format=json,metrics=cpu,memory

# Call a tool with JSON arguments
go-go-mcp server tools call calendar-event --json '{
  "title": "Team Meeting",
  "start_time": "2024-02-01 10:00",
  "duration": 60
}'
```

The available tools depend on the selected profile:
```bash
# System monitoring profile
go-go-mcp server tools list --profile system
# Shows: system-monitor, disk-usage, etc.

# Calendar profile
go-go-mcp server tools list --profile calendar
# Shows: calendar-event, calendar-availability, etc.

# Data analysis profile
go-go-mcp server tools list --profile data
# Shows: data-analyzer, data-visualizer, etc.
```

#### Client Mode

Use the client subcommand to interact with an MCP server:

```bash
# List available prompts (uses default server: go-go-mcp start --transport stdio)
go-go-mcp client prompts list

# List available tools
go-go-mcp client tools list

# Execute a prompt with arguments
go-go-mcp client prompts execute hello --args '{"name":"World"}'

# Call a tool with arguments
go-go-mcp client tools call echo --args '{"message":"Hello, MCP!"}'
```

You can customize the server command and arguments if needed:

```bash
# Use a different server binary with custom arguments
go-go-mcp client --command custom-server,start,--debug,--port,8001 prompts list

# Use a server with a specific configuration
go-go-mcp client -c go-go-mcp,start,--config,config.yaml prompts list
```

#### Using go-go-mcp as a Bridge

go-go-mcp can be used as a bridge to expose an SSE server as a stdio server. This is useful when you need to connect a stdio-based client to an SSE server:

```bash
# Start an SSE server on port 3000
go-go-mcp start --transport sse --port 3000

# In another terminal, start the bridge to expose the SSE server as stdio
go-go-mcp bridge --sse-url http://localhost:3000 --log-level debug

# Now you can use stdio-based clients to connect to the bridge
```

This is particularly useful when integrating with tools that only support stdio communication but need to connect to a web-based MCP server.

### Debug Mode

Add the `--debug` flag to enable detailed logging:

```bash
go-go-mcp start --debug 
```

### Version Information

Check the version:

```bash
go-go-mcp -v
```

## Configuration

go-go-mcp can be configured using YAML configuration files that allow you to:
- Define multiple profiles for different environments
- Configure tool and prompt sources
- Set parameter defaults and overrides
- Control access through blacklists/whitelists
- Manage security through parameter filtering

### Command-Line Configuration Management

The `config` command group provides tools to manage your configuration:

```bash
# Create a new configuration file
go-go-mcp config init

# Edit configuration in your default editor
go-go-mcp config edit

# List available profiles
go-go-mcp config list-profiles

# Show details of a specific profile
go-go-mcp config show-profile default

# Add a tool directory to a profile
go-go-mcp config add-tool default --dir ./tools/system

# Add a specific tool file to a profile
go-go-mcp config add-tool default --file ./tools/special-tool.yaml

# Create a new profile
go-go-mcp config add-profile production "Production environment configuration"

# Duplicate an existing profile
go-go-mcp config duplicate-profile development staging "Staging environment configuration"

# Set the default profile
go-go-mcp config set-default-profile production
```

For detailed configuration documentation, use:
```bash
# View configuration file documentation
go-go-mcp help config-file

# View example configuration
go-go-mcp help config-file --example
```

## Shell Commands

go-go-mcp supports defining custom shell commands in YAML files, providing:
- Template-based command generation with Go templates and Sprig functions
- Rich parameter type system
- Environment variable management
- Working directory control
- Error handling and output capture

### Example Commands

The `examples/` directory contains various ready-to-use commands. You can view the schema for any command using `go-go-mcp schema <command.yaml>`, which shows the full parameter documentation that is passed to the LLM:

```bash
go-go-mcp schema examples/github/list-github-issues.yaml
```

#### GitHub Integration
- [`examples/github/list-github-issues.yaml`](examples/github/list-github-issues.yaml) - List and filter GitHub issues
- [`examples/github/list-pull-requests.yaml`](examples/github/list-pull-requests.yaml) - List and filter pull requests

#### Web Content Tools
- [`examples/shell-commands/fetch-url.yaml`](examples/shell-commands/fetch-url.yaml) - Fetch and process web content
- [`examples/html-extract/pubmed.yaml`](examples/html-extract/pubmed.yaml) - Search and extract data from PubMed

#### Research and Documentation
- [`examples/shell-commands/diary-append.yaml`](examples/shell-commands/diary-append.yaml) - Maintain a timestamped diary
- [`examples/shell-commands/fetch-transcript.yaml`](examples/shell-commands/fetch-transcript.yaml) - Download YouTube video transcripts

For any command, you can view its full schema and documentation using:
```bash
# View the full parameter schema and documentation
go-go-mcp schema examples/github/list-github-issues.yaml

# View the command help
go-go-mcp run-command examples/github/list-github-issues.yaml --help
```

### Running Shell Commands Directly

You can run shell commands directly using the `run-command` subcommand. This allows you to execute any shell command YAML file without loading it into a server first:

```bash
# View command help and available flags
go-go-mcp run-command examples/github/list-github-issues.yaml --help

# Run a command with specific parameters
go-go-mcp run-command examples/github/list-github-issues.yaml --author wesen

# Run a Google Calendar command
go-go-mcp run-command examples/google/get-calendar.yaml --start today --end "next week"

# Run a URL fetching command
go-go-mcp run-command examples/shell-commands/fetch-url.yaml --url https://example.com
```

This provides a simpler way to use shell commands as standalone command-line tools without setting up a server.

### Generating Shell Commands with Pinocchio

You can use [Pinocchio](https://github.com/go-go-golems/pinocchio) to generate shell commands for go-go-mcp. First, add your local go-go-mcp repository as a Pinocchio repository:

```bash
pinocchio config repositories add $(pwd)/pinocchio
```

Then generate a new command using:

```bash
pinocchio go-go-mcp create-command --description "A command description"
```

This will create a new shell command YAML file with the appropriate structure and configuration.

For detailed shell commands documentation, use:
```bash
# View shell commands documentation
go-go-mcp help shell-commands

# View example shell commands
go-go-mcp help shell-commands --example
```

## Help System

go-go-mcp comes with comprehensive built-in documentation. To explore:

```bash
# List all available help topics
go-go-mcp help --all

# Get help on a specific topic
go-go-mcp help <topic>

# Show examples for a topic
go-go-mcp help <topic> --example
```
