# go-go-mcp

A Go implementation of the Model Context Protocol (MCP), providing a framework for building MCP servers and clients.

## Overview

This project implements the [Model Context Protocol](https://github.com/modelcontextprotocol/specification), which enables standardized communication between AI applications and language models. The implementation includes:

- Core protocol message types and interfaces
- A modular registry system for managing prompts, resources, and tools
- Thread-safe provider implementations
- A stdio server implementation
- Support for custom handlers and subscriptions

## Architecture

The project follows a modular, provider-based architecture with these main components:

1. Protocol Types (`pkg/protocol/types.go`)
2. Registry System
   - Prompts Registry (`pkg/prompts/registry.go`)
   - Resources Registry (`pkg/resources/registry.go`)
   - Tools Registry (`pkg/tools/registry.go`)
3. Server Implementation (`pkg/server/server.go`)
4. Error Handling (`pkg/registry.go`)

For detailed architecture documentation, see [Architecture Documentation](pkg/doc/architecture.md).

## Features

### Registry System

The registry system provides thread-safe management of prompts, resources, and tools:

```go
// Create registries
promptRegistry := prompts.NewRegistry()
resourceRegistry := resources.NewRegistry()
toolRegistry := tools.NewRegistry()

// Register a prompt with custom handler
promptRegistry.RegisterPromptWithHandler(protocol.Prompt{
    Name: "hello",
    Description: "A simple greeting prompt",
    Arguments: []protocol.PromptArgument{
        {
            Name: "name",
            Description: "Name to greet",
            Required: false,
        },
    },
}, func(prompt protocol.Prompt, args map[string]string) (*protocol.PromptMessage, error) {
    return &protocol.PromptMessage{
        Role: "user",
        Content: protocol.PromptContent{
            Type: "text",
            Text: fmt.Sprintf("Hello, %s!", args["name"]),
        },
    }, nil
})
```

### Custom Handlers

Each registry supports custom handlers for flexible behavior:

- **Prompts**: Custom message generation
- **Resources**: Custom content providers with subscription support
- **Tools**: Custom tool execution handlers

### Error Handling

Standardized error handling with JSON-RPC compatible error codes:

```go
var (
    ErrPromptNotFound   = NewError("prompt not found", -32000)
    ErrResourceNotFound = NewError("resource not found", -32001)
    ErrToolNotFound     = NewError("tool not found", -32002)
    ErrNotImplemented   = NewError("not implemented", -32003)
)
```

## Example Server

The project includes an example stdio server that demonstrates the registry system:

```go
package main

import (
    "io"

    "github.com/go-go-golems/go-go-mcp/pkg/prompts"
    "github.com/go-go-golems/go-go-mcp/pkg/protocol"
    "github.com/go-go-golems/go-go-mcp/pkg/resources"
    "github.com/go-go-golems/go-go-mcp/pkg/server"
    "github.com/go-go-golems/go-go-mcp/pkg/tools"
    "github.com/rs/zerolog/log"
)

func main() {
    srv := server.NewServer()

    // Create registries
    promptRegistry := prompts.NewRegistry()
    resourceRegistry := resources.NewRegistry()
    toolRegistry := tools.NewRegistry()

    // Register with server
    srv.GetRegistry().RegisterPromptProvider(promptRegistry)
    srv.GetRegistry().RegisterResourceProvider(resourceRegistry)
    srv.GetRegistry().RegisterToolProvider(toolRegistry)

    if err := srv.Start(); err != nil && err != io.EOF {
        log.Fatal().Err(err).Msg("Server error")
    }
}
```

### Supported Methods

The server implements the MCP specification methods:

- `initialize` - Protocol initialization and capability negotiation
- `ping` - Connection health check
- `prompts/list` - List available prompts
- `prompts/get` - Retrieve prompt content
- `resources/list` - List available resources
- `resources/read` - Read resource content
- `resources/subscribe` - Subscribe to resource updates
- `tools/list` - List available tools
- `tools/call` - Execute a tool

## Running

### Building the Binary

Build the binary which includes both server and client functionality:

```bash
go build -o go-go-mcp ./cmd/mcp-server/main.go
```

### Basic Usage

The binary can be run in two modes:
- Server mode (default): Run as an MCP server process
- Client mode: Use client commands to interact with an MCP server

#### Server Mode

Start the server with either stdio or SSE transport:

```bash
# Start with stdio transport (default)
./go-go-mcp start --transport stdio

# Start with SSE transport
./go-go-mcp start --transport sse --port 3001
```

#### Client Mode

Use the client subcommand to interact with an MCP server:

```bash
# List available prompts (uses default server: go-go-mcp start --transport stdio)
./go-go-mcp client prompts list

# List available tools
./go-go-mcp client tools list

# Execute a prompt with arguments
./go-go-mcp client prompts execute hello --args '{"name":"World"}'

# Call a tool with arguments
./go-go-mcp client tools call echo --args '{"message":"Hello, MCP!"}'
```

You can customize the server command and arguments if needed:

```bash
# Use a different server binary with custom arguments
./go-go-mcp client --command custom-server,start,--debug,--port,8001 prompts list

# Use a server with a specific configuration
./go-go-mcp client -c go-go-mcp,start,--config,config.yaml prompts list
```

#### Using SSE Transport

For web-based applications, use the SSE transport:

```bash
# Start the server with SSE transport
./go-go-mcp start --transport sse --port 3001

# In another terminal, connect using the client
./go-go-mcp client --transport sse --server http://localhost:3001 prompts list
```

### Debug Mode

Add the `--debug` flag to enable detailed logging:

```bash
./go-go-mcp client --debug prompts list
```

### Version Information

Check the version:

```bash
./go-go-mcp version
```

### Project Structure

- `pkg/`
  - `protocol/` - Core protocol types and interfaces
  - `prompts/` - Prompt registry and handlers
  - `resources/` - Resource registry and handlers
  - `tools/` - Tool registry and handlers
  - `server/` - Server implementation
  - `doc/` - Documentation
- `cmd/`
  - `mcp-server/` - MCP server and client implementation
    - `cmds/` - Command implementations
      - `client/` - Client subcommands

### Dependencies

- `github.com/rs/zerolog` - Structured logging
- `github.com/spf13/cobra` - Command-line interface
- Standard Go libraries

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

[Insert your chosen license here]