# Client Documentation

This document describes the client implementation of the Model Context Protocol (MCP).

## Overview

The client implementation provides a clean interface for communicating with MCP servers. It handles:
- Protocol initialization and capability negotiation
- Message formatting and transport abstraction
- Request/response lifecycle management
- Error handling and logging

## Core Components

### Client

The main client type (`pkg/client/client.go`) provides the primary interface:

```go
type Client struct {
    mu        sync.Mutex
    logger    zerolog.Logger
    transport Transport
    nextID    int
    
    capabilities       protocol.ClientCapabilities
    serverCapabilities protocol.ServerCapabilities
    initialized        bool
}
```

Key methods:
```go
func NewClient(logger zerolog.Logger, transport Transport) *Client
func (c *Client) Initialize(capabilities protocol.ClientCapabilities) error
func (c *Client) ListPrompts(cursor string) ([]protocol.Prompt, string, error)
func (c *Client) GetPrompt(name string, args map[string]string) (*protocol.PromptMessage, error)
func (c *Client) ListResources(cursor string) ([]protocol.Resource, string, error)
func (c *Client) ReadResource(uri string) (*protocol.ResourceContent, error)
func (c *Client) ListTools(cursor string) ([]protocol.Tool, string, error)
func (c *Client) CallTool(name string, args map[string]interface{}) (*protocol.ToolResult, error)
func (c *Client) CreateMessage(messages []protocol.Message, ...) (*protocol.Message, error)
```

### Transport System

The Transport interface abstracts communication methods:

```go
type Transport interface {
    Send(request *protocol.Request) (*protocol.Response, error)
    Close() error
}
```

#### StdioTransport

Line-based JSON communication over stdin/stdout:
```go
type StdioTransport struct {
    mu      sync.Mutex
    scanner *bufio.Scanner
    writer  *json.Encoder
}
```

Features:
- Simple and reliable
- No network dependencies
- Synchronous operation
- Suitable for CLI tools

#### SSETransport

Server-Sent Events based transport:
```go
type SSETransport struct {
    mu          sync.Mutex
    baseURL     string
    client      *http.Client
    sseClient   *sse.Client
    events      chan *sse.Event
    sessionID   string
    closeOnce   sync.Once
    closeChan   chan struct{}
    initialized bool
}
```

Features:
- HTTP-based messaging
- Asynchronous events
- Session management
- Automatic reconnection

## Usage Examples

### Basic Initialization

```go
// Create transport (stdio or SSE)
transport := client.NewStdioTransport()
// or
transport := client.NewSSETransport("http://localhost:3001")

// Create client
logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
client := client.NewClient(logger, transport)

// Initialize with capabilities
err := client.Initialize(protocol.ClientCapabilities{
    Sampling: &protocol.SamplingCapability{},
})
if err != nil {
    log.Fatal(err)
}
```

### Working with Prompts

```go
// List available prompts
prompts, nextCursor, err := client.ListPrompts("")
if err != nil {
    log.Fatal(err)
}

// Get specific prompt
message, err := client.GetPrompt("example", map[string]string{
    "context": "some context",
    "topic":   "example topic",
})
```

### LLM Sampling

```go
response, err := client.CreateMessage(
    []protocol.Message{{
        Role: "user",
        Content: protocol.MessageContent{
            Type: "text",
            Text: "What is the capital of France?",
        },
    }},
    protocol.ModelPreferences{
        Hints: []protocol.ModelHint{{
            Name: "claude-3-sonnet",
        }},
        IntelligencePriority: 0.8,
        SpeedPriority: 0.5,
    },
    "You are a helpful assistant.",
    100,
)
```

## Error Handling

The client provides structured error handling:

### Transport Errors
```go
if err != nil {
    switch {
    case errors.Is(err, io.EOF):
        // Handle connection closed
    case errors.Is(err, context.DeadlineExceeded):
        // Handle timeout
    default:
        // Handle other transport errors
    }
}
```

### Protocol Errors
```go
if response.Error != nil {
    switch response.Error.Code {
    case -32700:
        // Parse error
    case -32600:
        // Invalid request
    case -32601:
        // Method not found
    default:
        // Other protocol errors
    }
}
```

## Best Practices

1. **Initialization**
   - Always check initialization errors
   - Set appropriate capabilities
   - Handle version compatibility

2. **Transport Selection**
   - Use stdio for CLI tools
   - Use SSE for web applications
   - Consider custom transport for special needs

3. **Error Handling**
   - Check all errors
   - Provide context in errors
   - Log appropriate details
   - Handle transport failures

4. **Resource Management**
   - Close client when done
   - Handle timeouts
   - Clean up resources
   - Use context for cancellation

5. **Thread Safety**
   - Use client from one goroutine
   - Or implement your own synchronization
   - Respect transport limitations
   - Handle concurrent requests

## Implementation Notes

1. **Request IDs**
   - Automatically generated
   - Monotonically increasing
   - Thread-safe generation
   - Used for response matching

2. **Capabilities**
   - Declared during initialization
   - Matched with server
   - Affects available features
   - Can be extended

3. **Logging**
   - Structured logging
   - Level-based filtering
   - Transport events
   - Error details