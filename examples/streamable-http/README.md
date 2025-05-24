# Streamable HTTP Transport Example

This example demonstrates the new Streamable HTTP transport implementation for MCP 2025-03-26.

## Features

- **Bidirectional streaming**: Supports both WebSocket and HTTP POST for client-server communication
- **Session management**: Maintains session state across connections
- **Concurrent connections**: Multiple clients can connect to the same session
- **Fallback support**: HTTP POST endpoint for clients that don't support WebSockets

## Usage

### Starting the Server

#### Using the Example Server

```bash
go run main.go
```

The server will start on port 8080 with the following endpoints:

#### Using the Embeddable API

You can also use the embeddable MCP server with streamable HTTP transport:

```go
package main

import (
	"context"
	"log"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func main() {
	mcpCmd := embeddable.NewMCPCommand(
		embeddable.WithName("My Streamable HTTP Server"),
		embeddable.WithDefaultTransport("streamable_http"),
		embeddable.WithDefaultPort(8081),
		embeddable.WithTool("echo", func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
			message, _ := args["message"].(string)
			return &protocol.ToolResult{
				Content: []protocol.ToolContent{{
					Type: "text",
					Text: "Echo: " + message,
				}},
			}, nil
		}, 
			embeddable.WithDescription("Echo back a message"),
			embeddable.WithStringArg("message", "Message to echo", false),
		),
	)
	
	if err := mcpCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
```

Both approaches will start servers with the following endpoints:

- **WebSocket**: `ws://localhost:8080/stream` - For bidirectional streaming
- **HTTP POST**: `http://localhost:8080/messages` - For one-way communication

### Testing with WebSocket

You can test the WebSocket endpoint using a simple JavaScript client:

```javascript
const ws = new WebSocket('ws://localhost:8080/stream');

ws.onopen = function() {
    console.log('Connected to streamable HTTP transport');
    
    // Send a test request
    ws.send(JSON.stringify({
        jsonrpc: "2.0",
        id: "test-1",
        method: "test/echo",
        params: {
            message: "Hello, Streamable HTTP!"
        }
    }));
};

ws.onmessage = function(event) {
    const response = JSON.parse(event.data);
    console.log('Received:', response);
};
```

### Testing with HTTP POST

You can also test the HTTP POST endpoint using curl:

```bash
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "test-1",
    "method": "test/echo",
    "params": {
      "message": "Hello, HTTP!"
    }
  }'
```

## Transport Configuration

The transport can be configured with various options:

```go
transport, err := streamable_http.NewStreamableHTTPTransport(
    transport.WithStreamableHTTPOptions(transport.StreamableHTTPOptions{
        Addr:            ":8080",           // Server address
        ReadBufferSize:  1024,              // WebSocket read buffer size
        WriteBufferSize: 1024,              // WebSocket write buffer size
        CheckOrigin: func(r *http.Request) bool {
            // Implement origin checking for security
            return true
        },
    }),
)
```

## Security Considerations

- The example allows all origins for demonstration purposes
- In production, implement proper origin checking in the `CheckOrigin` function
- Consider using TLS for production deployments
- Implement proper authentication and authorization

## Compatibility

This transport is compatible with MCP 2025-03-26 specification and provides:

- Bidirectional streaming over HTTP using WebSockets
- Fallback to HTTP POST for clients that don't support WebSockets
- Session-based client management
- Automatic reconnection handling (client-side implementation required)