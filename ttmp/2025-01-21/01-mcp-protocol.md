MCP uses a client-server architecture with two main transport options:

1. **Local stdio Transport**
- Messages are sent over standard input/output streams
- Each message is a single line of JSON
- Messages are terminated with newlines
- stderr is used for logging and doesn't interfere with the protocol

2. **SSE (Server-Sent Events) Transport** 
- Used for remote connections
- Server exposes an HTTP endpoint
- Uses SSE for server->client communication
- Uses HTTP POST requests for client->server communication

### Message Format

All messages follow this basic JSON structure:
```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "method": "method-name",
  "params": { /* method-specific parameters */ }
}
```

### Core Protocol Flow

1. **Initialization**
```json
// Client sends initialization request
{
  "jsonrpc": "2.0", 
  "id": "1",
  "method": "initialize",
  "params": {
    "capabilities": {
      "tools": {},
      "prompts": {},
      "resources": {}
    }
  }
}

// Server responds with capabilities
{
  "jsonrpc": "2.0",
  "id": "1", 
  "result": {
    "capabilities": {
      "tools": {},
      "prompts": {},
      "resources": {}
    }
  }
}
```

2. **Tool Listing**
```json
// Client requests available tools
{
  "jsonrpc": "2.0",
  "id": "2",
  "method": "listTools",
  "params": {}
}

// Server responds with tool definitions
{
  "jsonrpc": "2.0",
  "id": "2",
  "result": {
    "tools": [
      {
        "name": "get-forecast",
        "description": "Get weather forecast for a location",
        "inputSchema": {
          "type": "object",
          "properties": {
            "latitude": {"type": "number"},
            "longitude": {"type": "number"}
          }
        }
      }
    ]
  }
}
```

3. **Tool Execution**
```json
// Client requests tool execution
{
  "jsonrpc": "2.0",
  "id": "3",
  "method": "callTool",
  "params": {
    "name": "get-forecast",
    "arguments": {
      "latitude": 37.7749,
      "longitude": -122.4194
    }
  }
}

// Server responds with tool result
{
  "jsonrpc": "2.0",
  "id": "3",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Forecast for San Francisco..."
      }
    ]
  }
}
```

4. **Resource Operations**
```json
// Client requests resource listing
{
  "jsonrpc": "2.0",
  "id": "4", 
  "method": "listResources",
  "params": {}
}

// Server responds with available resources
{
  "jsonrpc": "2.0",
  "id": "4",
  "result": {
    "resources": [
      {
        "name": "config.json",
        "mimeType": "application/json",
        "description": "Configuration file"
      }
    ]
  }
}
```

5. **Logging Messages**
```json
// Server can send log messages as notifications
{
  "jsonrpc": "2.0",
  "method": "logging",
  "params": {
    "level": "info",
    "data": "Server started successfully"
  }
}
```

### Error Handling

When errors occur, responses include an error object:
```json
{
  "jsonrpc": "2.0",
  "id": "5",
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": { /* additional error details */ }
  }
}
```

### Key Implementation Notes

1. All messages must include the `jsonrpc: "2.0"` field to identify the protocol version

2. Request IDs must be matched in responses to allow asynchronous operation

3. Notifications (like logging) don't include an ID since they don't expect responses

4. The protocol is bidirectional - both client and server can send requests and receive responses

5. Error codes follow the JSON-RPC 2.0 specification
