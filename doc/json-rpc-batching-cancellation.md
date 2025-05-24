# JSON-RPC Batching and Cancellation Support

This document describes the JSON-RPC batching and cancellation features implemented in go-go-mcp according to the MCP 2025-03-26 specification.

## JSON-RPC Batching

### Overview

JSON-RPC batching allows clients to send multiple requests in a single HTTP request or message, improving efficiency and reducing latency. The server processes all requests in the batch and returns a batch response.

### Protocol Types

```go
// BatchRequest represents an array of JSON-RPC 2.0 requests
type BatchRequest []Request

// BatchResponse represents an array of JSON-RPC 2.0 responses  
type BatchResponse []Response
```

### Usage Examples

#### Single Request
```json
{"jsonrpc":"2.0","id":"1","method":"ping"}
```

#### Batch Request
```json
[
  {"jsonrpc":"2.0","id":"1","method":"ping"},
  {"jsonrpc":"2.0","id":"2","method":"echo","params":{"message":"hello"}},
  {"jsonrpc":"2.0","method":"notifications/initialized"}
]
```

#### Batch Response
```json
[
  {"jsonrpc":"2.0","id":"1","result":{"pong":true}},
  {"jsonrpc":"2.0","id":"2","result":{"echo":"hello"}}
]
```

Note that notifications in a batch don't generate responses.

### Implementation Details

- **Transport Layer**: Both `stdio` and `sse` transports support batch processing
- **Message Parsing**: `transport.ParseMessage()` automatically detects single vs batch requests
- **Handler Interface**: New `HandleBatchRequest()` method added to `RequestHandler` interface
- **Validation**: Batch requests are validated for proper JSON-RPC format
- **Error Handling**: Individual request errors don't fail the entire batch

### Transport Support

All transports now implement batch processing:

- **stdio**: Line-delimited batch requests/responses
- **SSE/HTTP**: HTTP POST with JSON batch in request body
- **Streamable HTTP**: Full batch support over HTTP

## Request Cancellation

### Overview

Request cancellation allows clients or servers to cancel in-progress requests using notification messages. This is particularly useful for long-running operations.

### Cancellation Flow

1. Client sends a request with a unique ID
2. Server begins processing the request
3. Client (or server) sends a cancellation notification referencing the request ID
4. Server attempts to cancel the in-progress request
5. Server cleans up resources and stops processing

### Protocol Messages

#### Cancellation Notification
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/cancelled",
  "params": {
    "requestId": "123",
    "reason": "User requested cancellation"
  }
}
```

#### Cancellation Parameters
```go
type CancellationParams struct {
    RequestID string `json:"requestId"`
    Reason    string `json:"reason,omitempty"`
}
```

### Usage Examples

#### Long-running Request
```json
{"jsonrpc":"2.0","id":"slow-task-1","method":"slow_operation","params":{"data":"large_dataset"}}
```

#### Cancellation Notification
```json
{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"slow-task-1","reason":"User cancelled operation"}}
```

### Implementation Guidelines

#### Server Implementation
```go
func (h *Handler) HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
    reqID := string(req.ID)
    
    // Create cancellable context for this request
    requestCtx, cancel := context.WithCancel(ctx)
    h.storeCancellationFunction(reqID, cancel)
    
    defer h.removeCancellationFunction(reqID)
    
    // Use requestCtx for long-running operations
    return h.processRequest(requestCtx, req)
}

func (h *Handler) HandleNotification(ctx context.Context, notif *protocol.Notification) error {
    if notif.Method == "notifications/cancelled" {
        var params protocol.CancellationParams
        json.Unmarshal(notif.Params, &params)
        
        // Cancel the request if it exists
        if cancel := h.getCancellationFunction(params.RequestID); cancel != nil {
            cancel()
        }
    }
    return nil
}
```

#### Client Implementation
```go
// Send cancellation notification
cancellation := protocol.NewCancellationNotification("request-123", "User cancelled")
client.SendNotification(ctx, cancellation)
```

### Behavior Requirements

1. **Cancellation Scope**: Only requests in the same direction can be cancelled
2. **Initialize Protection**: The `initialize` request cannot be cancelled by clients  
3. **Best Effort**: Servers should attempt cancellation but may ignore if request is complete
4. **Race Conditions**: Handle gracefully when cancellation arrives after completion
5. **Resource Cleanup**: Free associated resources when cancelling requests

### Error Handling

- **Unknown Request IDs**: Ignore cancellation requests for unknown IDs
- **Completed Requests**: Ignore cancellation for already completed requests
- **Invalid Notifications**: Malformed cancellations should be ignored
- **No Response**: Cancellation notifications never generate responses

## Examples

### Basic Batch Example

See `examples/batch-cancellation/` for a complete working example that demonstrates:

- Processing batch requests with mixed operations
- Long-running tasks that can be cancelled
- Proper context handling for cancellation
- Error handling in batch processing

### Running the Example

```bash
go run examples/batch-cancellation/main.go
```

Then send test messages via stdin:

```bash
# Batch request
echo '[{"jsonrpc":"2.0","id":"1","method":"ping"},{"jsonrpc":"2.0","id":"2","method":"echo","params":{"test":true}}]' | go run examples/batch-cancellation/main.go

# Long task + cancellation
go run examples/batch-cancellation/main.go &
echo '{"jsonrpc":"2.0","id":"slow-1","method":"slow_task"}' 
sleep 1
echo '{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"slow-1","reason":"User requested cancellation"}}'
```

## Best Practices

### Batch Processing
- Keep batch sizes reasonable (< 100 requests typically)
- Mix different operation types in batches for efficiency
- Handle partial failures gracefully
- Validate each request in the batch independently

### Cancellation
- Always use context.Context for cancellable operations
- Store cancellation functions with request IDs
- Clean up resources on cancellation
- Log cancellation events for debugging
- Test race conditions between completion and cancellation

### Performance
- Batch related operations for better throughput
- Use cancellation for user-initiated stops
- Implement timeouts alongside cancellation
- Monitor batch processing times