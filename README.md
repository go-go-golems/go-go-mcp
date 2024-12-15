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

    "github.com/go-go-golems/go-mcp/pkg/prompts"
    "github.com/go-go-golems/go-mcp/pkg/protocol"
    "github.com/go-go-golems/go-mcp/pkg/resources"
    "github.com/go-go-golems/go-mcp/pkg/server"
    "github.com/go-go-golems/go-mcp/pkg/tools"
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

## Usage

### Running the Server

Build and run the example stdio server:

```bash
go build -o stdio-server cmd/stdio-server/main.go
./stdio-server
```

### Client Connection

The server accepts JSON-RPC messages on stdin and writes responses to stdout. Example initialization:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "prompts": {
        "listChanged": true
      },
      "resources": {
        "subscribe": true,
        "listChanged": true
      },
      "tools": {
        "listChanged": true
      }
    },
    "clientInfo": {
      "name": "example-client",
      "version": "1.0.0"
    }
  }
}
```

## Development

### Project Structure

- `pkg/`
  - `protocol/` - Core protocol types and interfaces
  - `prompts/` - Prompt registry and handlers
  - `resources/` - Resource registry and handlers
  - `tools/` - Tool registry and handlers
  - `server/` - Server implementation
  - `doc/` - Documentation
- `cmd/stdio-server/` - Example stdio server implementation

### Dependencies

- `github.com/rs/zerolog` - Structured logging
- Standard Go libraries

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

[Insert your chosen license here]
