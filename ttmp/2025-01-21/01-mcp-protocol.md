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

Detailed SSE Protocol Flow:

1. **Connection Establishment**
   - Client initiates connection by making a GET request to the server endpoint
   - Server responds with SSE headers:
     ```http
     Content-Type: text/event-stream
     Cache-Control: no-cache
     Connection: keep-alive
     ```
   - Server generates a unique session ID (UUID)
   - Server sends initial endpoint event:
     ```
     event: endpoint
     data: {endpoint-url}?sessionId={session-uuid}
     ```

2. **Message Transport**
   - Server → Client:
     - Messages sent as SSE events:
       ```
       event: message
       data: {JSON-RPC message}
       ```
   - Client → Server:
     - Messages sent as HTTP POST requests
     - Content-Type must be application/json
     - Maximum message size: 4MB
     - POST to the endpoint URL provided in initial connection
     - Server responds with:
       - 202: Message accepted
       - 400: Invalid message
       - 500: SSE connection not established

3. **Security Considerations**
   - Endpoint origin validation: Client verifies the POST endpoint matches the SSE connection origin
   - Session management: Each connection has a unique session ID for routing messages
   - Content validation: All messages must conform to the JSON-RPC message schema

4. **Error Handling**
   - Connection errors trigger onerror handlers
   - Invalid messages return HTTP 400 responses
   - Missing SSE connection returns HTTP 500
   - Content-type mismatches are rejected
   - Message size limits enforced

5. **Connection Termination**
   - Either side can initiate closure
   - Server cleans up session resources
   - Client closes EventSource connection
   - Both sides receive close notifications

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
