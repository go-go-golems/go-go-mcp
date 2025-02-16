# RFC: Clean Transport Layer for MCP

## Status
- Status: Draft
- Date: 2025-02-14
- Authors: Claude
- Related Issues: N/A

## Abstract

This RFC proposes a clean, unified transport layer for the MCP protocol. The current implementation mixes concerns between transport, request handling, and business logic. The proposed design separates these concerns and provides a clear interface for implementing new transport mechanisms.

## Background

The current MCP implementation has two transport mechanisms:
1. Server-Sent Events (SSE) for real-time communication
2. Standard IO (stdio) for command-line usage

## Proposal

### 1. Core Interfaces

#### Transport Interface

```go
// Transport handles the low-level communication between client and server
type Transport interface {
    // Listen starts accepting requests and forwards them to the handler
    Listen(ctx context.Context, handler RequestHandler) error
    
    // Send transmits a response back to the client
    Send(ctx context.Context, response *Response) error
    
    // Close cleanly shuts down the transport
    Close(ctx context.Context) error
    
    // Info returns metadata about the transport
    Info() TransportInfo
}

// TransportInfo provides metadata about the transport
type TransportInfo struct {
    Type            string            // "sse", "stdio", etc.
    RemoteAddr      string            // Remote address if applicable
    Capabilities    map[string]bool   // Transport capabilities
    Metadata        map[string]string // Additional transport metadata
}
```

#### Request Handler Interface

```go
// RequestHandler processes incoming requests and notifications
type RequestHandler interface {
    // HandleRequest processes a request and returns a response
    HandleRequest(ctx context.Context, req *Request) (*Response, error)
    
    // HandleNotification processes a notification (no response expected)
    HandleNotification(ctx context.Context, notif *Notification) error
}

// Request represents an incoming JSON-RPC request
type Request struct {
    ID      string
    Method  string
    Params  json.RawMessage
    Headers map[string]string
}

// Response represents an outgoing JSON-RPC response
type Response struct {
    ID      string
    Result  json.RawMessage
    Error   *ResponseError
    Headers map[string]string
}

// Notification represents an incoming notification
type Notification struct {
    Method  string
    Params  json.RawMessage
    Headers map[string]string
}
```

### 2. Transport Options

```go
// Common options for all transports
type TransportOptions struct {
    // Common options
    MaxMessageSize int64
    Logger zerolog.Logger
    
    // Transport specific options
    SSE     *SSEOptions
    Stdio   *StdioOptions
}

// SSE-specific options
type SSEOptions struct {
    // HTTP server configuration
    Addr            string
    TLSConfig       *tls.Config
    
    // Middleware
    Middleware []func(http.Handler) http.Handler
}

// Stdio-specific options
type StdioOptions struct {
    // Buffer sizes
    ReadBufferSize  int
    WriteBufferSize int
    
    // Process management
    Command     string
    Args        []string
    WorkingDir  string
    Environment map[string]string
    
    // Signal handling
    SignalHandlers map[os.Signal]func()
}

// Option constructors for each transport type
func WithSSEOptions(opts SSEOptions) TransportOption
func WithStdioOptions(opts StdioOptions) TransportOption
```

### 3. SSE Transport Implementation

```go
// SSETransport implements Transport using Server-Sent Events
type SSETransport struct {
    opts     TransportOptions
    server   *http.Server
    upgrader *sse.Upgrader
    clients  sync.Map
    logger   zerolog.Logger
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(addr string, opts ...TransportOption) (*SSETransport, error)

// Implementation example
func (t *SSETransport) Listen(ctx context.Context, handler RequestHandler) error {
    // Set up HTTP server
    t.server = &http.Server{
        Addr:    t.addr,
        Handler: t.createHandler(handler),
    }
    
    // Start server
    go func() {
        if err := t.server.ListenAndServe(); err != http.ErrServerClosed {
            t.logger.Error().Err(err).Msg("SSE server error")
        }
    }()
    
    // Wait for context cancellation
    <-ctx.Done()
    return t.Close(ctx)
}
```

### 4. Stdio Transport Implementation

```go
// StdioTransport implements Transport using standard input/output
type StdioTransport struct {
    opts    TransportOptions
    stdin   *bufio.Reader
    stdout  *bufio.Writer
    logger  zerolog.Logger
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(opts ...TransportOption) (*StdioTransport, error)

// Implementation example
func (t *StdioTransport) Listen(ctx context.Context, handler RequestHandler) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
            
        default:
            // Read request
            req, err := t.readRequest()
            if err != nil {
                if err == io.EOF {
                    return nil
                }
                t.logger.Error().Err(err).Msg("Failed to read request")
                continue
            }
            
            // Handle request
            go t.handleRequest(ctx, req, handler)
        }
    }
}
```

### 5. Error Handling

```go
// ResponseError represents a JSON-RPC error response
type ResponseError struct {
    Code    int             `json:"code"`
    Message string          `json:"message"`
    Data    json.RawMessage `json:"data,omitempty"`
}

// Common error codes
const (
    ErrCodeParse           = -32700
    ErrCodeInvalidRequest  = -32600
    ErrCodeMethodNotFound  = -32601
    ErrCodeInvalidParams   = -32602
    ErrCodeInternal        = -32603
    ErrCodeTransport       = -32500
    ErrCodeTimeout         = -32501
)

// Error constructors
func NewParseError(msg string) *ResponseError
func NewInvalidRequestError(msg string) *ResponseError
func NewMethodNotFoundError(msg string) *ResponseError
func NewInvalidParamsError(msg string) *ResponseError
func NewInternalError(msg string) *ResponseError
func NewTransportError(msg string) *ResponseError
func NewTimeoutError(msg string) *ResponseError
```

## Implementation

The implementation will follow these steps:

1. Create new transport package with core interfaces
2. Implement SSE transport
3. Implement stdio transport
6. Update server to use new transport layer

## Migration Guide

1. Update imports to use new transport package
2. Replace direct transport usage with interface
3. Update error handling to use new error types
4. Add transport options where needed

Example:
```go
// Before
transport := client.NewSSETransport(logger, "http://localhost:8080")

// After
transport, err := NewSSETransport(
    "localhost:8080",
    WithLogger(logger),
    WithMaxMessageSize(1024*1024),
)
```

## Migration Plan

### Phase 1: Package Structure (Week 1)

1. Create new package structure:
```
pkg/
  transport/
    transport.go     # Core interfaces
    options.go       # Transport options
    errors.go        # Error types
    sse/
      transport.go   # SSE implementation
      options.go     # SSE specific options
    stdio/
      transport.go   # Stdio implementation
      options.go     # Stdio specific options
    testing/
      mock.go        # Mock transport for testing
```

2. Move existing transport code:
```bash
# Create new directories
mkdir -p pkg/transport/{sse,stdio,testing}

# Move existing files
mv pkg/client/sse.go pkg/transport/sse/transport.go
mv pkg/client/stdio.go pkg/transport/stdio/transport.go
```

### Phase 2: Interface Implementation (Week 2-3)

1. Update SSE Transport:

```go
// Old implementation
type SSETransport struct {
    mu                  sync.Mutex
    baseURL             string
    client              *http.Client
    sseClient           *sse.Client
    events              chan *sse.Event
    responses           chan *sse.Event
    notifications       chan *sse.Event
    closeOnce           sync.Once
    logger              zerolog.Logger
    initialized         bool
    sessionID           string
    endpoint            string
    notificationHandler func(*protocol.Response)
}

// New implementation
type SSETransport struct {
    opts     TransportOptions
    server   *http.Server
    upgrader *sse.Upgrader
    clients  sync.Map
    logger   zerolog.Logger
    handler  RequestHandler
}

// Migration steps:
1. Create new struct with required fields
2. Implement Transport interface
3. Add SSE-specific functionality
4. Update error handling
5. Add tests
```

2. Update Stdio Transport:

```go
// Old implementation
type StdioTransport struct {
    mu                  sync.Mutex
    scanner             *bufio.Scanner
    writer              *json.Encoder
    cmd                 *exec.Cmd
    logger              zerolog.Logger
    notificationHandler func(*protocol.Response)
}

// New implementation
type StdioTransport struct {
    opts    TransportOptions
    stdin   *bufio.Reader
    stdout  *bufio.Writer
    cmd     *exec.Cmd
    logger  zerolog.Logger
    handler RequestHandler
}

// Migration steps:
1. Create new struct with required fields
2. Implement Transport interface
3. Add process management
4. Update error handling
5. Add tests
```

### Phase 3: Client Updates (Week 4)

1. Update Client struct:

```go
// Old implementation
type Client struct {
    mu        sync.Mutex
    logger    zerolog.Logger
    transport Transport
    nextID    int
    // ...
}

// New implementation
type Client struct {
    mu        sync.Mutex
    logger    zerolog.Logger
    transport transport.Transport
    nextID    int
    // ...
}

// Migration steps:
1. Update import paths
2. Implement RequestHandler interface
3. Update error handling
4. Add transport options
```

2. Update client creation:

```go
// Old
client := NewClient(logger, NewSSETransport(baseURL))

// New
transport, err := sse.NewTransport(
    transport.WithSSEOptions(sse.Options{
        Addr: baseURL,
    }),
    transport.WithLogger(logger),
)
if err != nil {
    return nil, err
}
client := NewClient(logger, transport)
```

### Phase 4: Server Updates (Week 5)

1. Update Server struct:

```go
// Old implementation
type Server struct {
    logger    zerolog.Logger
    transport Transport
    // ...
}

// New implementation
type Server struct {
    logger    zerolog.Logger
    transport transport.Transport
    handler   RequestHandler
    // ...
}

// Migration steps:
1. Update import paths
2. Implement RequestHandler interface
3. Add transport configuration
4. Update error handling
```

2. Update server creation:

```go
// Old
server := NewServer(logger, NewSSETransport(":8080"))

// New
transport, err := sse.NewTransport(
    transport.WithSSEOptions(sse.Options{
        Addr: ":8080",
        TLSConfig: &tls.Config{...},
        Middleware: []func(http.Handler) http.Handler{
            cors.Handler,
            auth.Handler,
        },
    }),
    transport.WithLogger(logger),
)
if err != nil {
    return nil, err
}
server := NewServer(logger, transport)
```

### Phase 5: Testing Updates (Week 6)

1. Create mock transport:

```go
// pkg/transport/testing/mock.go
type MockTransport struct {
    handler RequestHandler
    requests chan *Request
    responses chan *Response
}

func (m *MockTransport) Listen(ctx context.Context, handler RequestHandler) error {
    m.handler = handler
    // Implementation
}

// Usage in tests
func TestClient(t *testing.T) {
    mock := testing.NewMockTransport()
    client := NewClient(logger, mock)
    
    // Test client with mock transport
}
```

2. Update existing tests:

```go
// Old tests
func TestSSETransport(t *testing.T) {
    transport := NewSSETransport(":0")
    // ...
}

// New tests
func TestSSETransport(t *testing.T) {
    transport, err := sse.NewTransport(
        transport.WithSSEOptions(sse.Options{
            Addr: ":0",
            HeartbeatInterval: time.Second,
        }),
    )
    require.NoError(t, err)
    // ...
}
```
