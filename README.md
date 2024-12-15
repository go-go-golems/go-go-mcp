# go-go-mcp

A Go implementation of the Model Context Protocol (MCP), providing a framework for building MCP servers and clients.

## Overview

This project implements the [Model Context Protocol](https://github.com/modelcontextprotocol/specification), which enables standardized communication between AI applications and language models. The implementation includes:

- Core protocol message types
- A stdio server implementation
- Support for prompts and logging capabilities

## Example Server

The project includes an example stdio server that demonstrates basic MCP functionality:

```go
package main

import (
    "github.com/go-go-golems/go-mcp/pkg"
)

func main() {
    server := NewServer()
    if err := server.Start(); err != nil && err != io.EOF {
        log.Fatal().Err(err).Msg("Server error")
    }
}
```

### Features

The example server implements:

- JSON-RPC 2.0 message handling
- Protocol version negotiation
- Capability declaration
- Structured logging
- Simple prompt system

### Supported Methods

- `initialize` - Protocol initialization and capability negotiation
- `ping` - Connection health check
- `prompts/list` - List available prompts
- `prompts/get` - Retrieve prompt content

### Example Prompt

The server includes a simple prompt that demonstrates prompt arguments:

```json
{
  "name": "simple",
  "description": "A simple prompt that can take optional context and topic arguments",
  "arguments": [
    {
      "name": "context",
      "description": "Additional context to consider",
      "required": false
    },
    {
      "name": "topic",
      "description": "Specific topic to focus on",
      "required": false
    }
  ]
}
```

## Usage

### Running the Server

Build and run the example stdio server:

```bash
go build -o stdio-server go/cmd/stdio-server/main.go
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
    "capabilities": {},
    "clientInfo": {
      "name": "example-client",
      "version": "1.0.0"
    }
  }
}
```

## Development

### Project Structure

- `pkg/` - Core protocol types and utilities
- `cmd/stdio-server/` - Example stdio server implementation

### Dependencies

- `github.com/rs/zerolog` - Structured logging
- Standard Go libraries

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Insert your chosen license here]
