# Design, Architecture, and Control Flow of go-go-mcp

## Overview

This document outlines the design, architecture, and request flow of go-go-mcp, a Go implementation of the Model Context Protocol (MCP) server that exposes RPC-like capabilities. The server enables **session state management** and communication between clients and server through different transport mechanisms (SSE, stdio).

## Architectural Components

### Core Components

1. **Server**: Manages the lifecycle of the MCP server, holds provider and **session store** references, and coordinates request processing.
2. **Transports**: Handles communication protocols (HTTP/SSE, stdio) between clients and the server. **Responsible for session identification and context enhancement.**
3. **RequestHandler**: Processes incoming requests (with session context) and dispatches them to appropriate method handlers.
4. **Providers**: Interfaces for serving prompts, resources, and tools. Receive session context.
5. **Tools**: Implementations of various tools that can be invoked via the MCP protocol. Receive session context.
6. **Session Management**: Handles session creation, storage (`SessionStore`), and context propagation (`pkg/session`).

### Key Files and Packages

1. **Server Package (`pkg/server/`)**
   - `server.go`: Main server implementation (holds `SessionStore`).
   - `handler.go`: Request handling and routing (uses session context).
   - `options.go`: Server configuration options (includes `WithSessionStore`).

2. **Session Package (`pkg/session/`)**
   - `session.go`: `Session`, `SessionState` types; `WithSession`, `GetSessionFromContext` utilities.
   - `store.go`: `SessionStore` interface, `InMemorySessionStore` implementation.

3. **Transport Package (`pkg/transport/`)**
   - `transport.go`: Common `Transport` interface (includes `SetSessionStore`), `RequestHandler` interface.
   - `errors.go`: Common error types for MCP, `ProcessError`, `ErrorToHTTPStatus` helpers.
   - `options.go`: Transport configuration options.
   - `sse/transport.go`: Server-Sent Events transport implementation (handles cookie-based sessions).
   - `stdio/transport.go`: Standard I/O transport implementation (handles connection-based sessions).

4. **Protocol Package (`pkg/protocol/`)**
   - Contains message types and protocol definitions.

5. **Provider Interfaces (`pkg/providers.go`)**
   - Defines interfaces for PromptProvider, ResourceProvider, and ToolProvider (methods accept `context.Context`).

6. **Tool Registry (`pkg/tools/providers/tool-registry/registry.go`)**
   - Registry for tool registration and invocation (methods accept `context.Context`).

7. **Command Package (`cmd/go-go-mcp/`)**
   - `main.go`: Application entry point.
   - `cmds/server/start.go`: Server startup logic.

## Initialization Flow

The initialization flow of the MCP server starts in `cmd/go-go-mcp/cmds/server/start.go` and proceeds as follows:

1. **Transport Creation**:
   ```go
   // Create transport based on type
   var t transport.Transport
   switch transportType {
   case "sse":
     t, err = sse.NewSSETransport(
       transport.WithLogger(logger),
       transport.WithSSEOptions(transport.SSEOptions{
         Addr: fmt.Sprintf(":%d", port),
       }),
     )
   case "stdio":
     t, err = stdio.NewStdioTransport(
       transport.WithLogger(logger),
     )
   }
   ```

2. **Provider Creation**:
   ```go
   // Create tool provider
   configToolProvider, err := layers.CreateToolProvider(serverSettings)
   
   // Initialize the final tool provider
   var toolProvider pkg.ToolProvider = configToolProvider
   
   // Register internal servers if specified
   if len(s_.InternalServers) > 0 {
     registry := tool_registry.NewRegistry()
     
     // Register the internal servers
     if err := registerInternalServers(registry, s_.InternalServers); err != nil {
       return err
     }
     
     // Combine the registry with the config tool provider
     toolProvider = tools.CombineProviders(configToolProvider, registry)
   }
   
   // Create resource provider
   resourceProvider := resources.NewRegistry()
   ```

3. **Server Creation**:
   ```go
   // Create server with transport, providers, and potentially session store option
   s := server.NewServer(logger, t,
     server.WithToolProvider(toolProvider),
     server.WithResourceProvider(resourceProvider),
     // Optionally: server.WithSessionStore(customSessionStore),
   )
   // NewServer initializes a default InMemorySessionStore if none is provided.
   ```

4. **Server Start**:
   ```go
   // Start server in a goroutine
   g.Go(func() error {
     defer cancel()
     // Server.Start injects the session store into the transport
     // via transport.SetSessionStore()
     if err := s.Start(gctx); err != nil && err != io.EOF {
       logger.Error().Err(err).Msg("Server error")
       return err
     }
     return nil
   })
   ```

## Request Processing Flow (with Session Management)

The request processing flow in the MCP server follows these steps:

1. **Request Reception (Transport Layer)**:
   - `SSETransport` (`handleSSE`, `handleMessages`) or `StdioTransport` (read loop) receives the raw request.

2. **Session Identification & Context Enhancement (Transport Layer)**:
   - **SSE**: `getSessionFromRequest` retrieves/creates a session using the `SessionStore` based on the `mcp_session_id` cookie. `setSessionCookie` is called. Context is enhanced: `ctx := session.WithSession(r.Context(), currentSession)`.
   - **Stdio**: `manageSession` retrieves/creates the `currentSession` for the connection (new session on `initialize`). Context is enhanced: `ctx := session.WithSession(baseCtx, s.currentSession)`.

3. **Forward to Handler (Transport Layer)**:
   - The transport calls `s.handler.HandleRequest(ctx, &request)` or `s.handler.HandleNotification(ctx, &notification)`, passing the **session-enhanced context**.

4. **Request Routing (`RequestHandler`)**:
   - `RequestHandler.HandleRequest` in `pkg/server/handler.go` receives the enhanced context.
   - It validates the request and routes it based on the method to a specific handler (e.g., `handleInitialize`), passing the **enhanced context** along.

5. **Method Handler Processing (`RequestHandler`)**:
   - Each method-specific handler (e.g., `handleInitialize`, `handleToolsCall`) receives the **enhanced context**.
   - `handleInitialize` retrieves the session from the context using `session.GetSessionFromContext` and logs the ID.
   - Parameters are extracted.
   - Provider methods are called with the **enhanced context**.

6. **Provider Invocation (Providers)**:
   - Provider methods (e.g., `ToolProvider.CallTool`) receive the **enhanced context**.
   - They can optionally access session data: `s, ok := session.GetSessionFromContext(ctx)`.
   - They execute their logic and return results.

7. **Response Creation (`RequestHandler`)**:
   - The handler creates a success response (`newSuccessResponse`). `handleInitialize` adds the SessionID to the result.

8. **Response Transmission (Transport Layer)**:
   - The response is sent back via the appropriate transport mechanism (SSE client channel or stdio writer).

## Session Handling (Implementation Details)

Session state is managed primarily within the transport layer, leveraging the `pkg/session` package.

### Session Package (`pkg/session`)

- Provides `Session`, `SessionState`, `SessionStore`, `InMemorySessionStore`.
- Provides context utilities: `WithSession(ctx, session)` and `GetSessionFromContext(ctx)`. `InMemorySessionStore` uses `sync.RWMutex`.

### Server Integration

- `Server` holds a `sessionStore session.SessionStore`.
- `NewServer` creates a default `InMemorySessionStore`.
- `WithSessionStore` allows injecting custom stores.
- `Server.Start` calls `transport.SetSessionStore` to pass the store to the active transport.

### SSE Transport Session Handling (`pkg/transport/sse/transport.go`)

- Implements `SetSessionStore` and holds a `sessionStore` field.
- Uses `http.Cookie` named `mcp_session_id`.
- `getSessionFromRequest(r)`:
  - Reads cookie.
  - Calls `sessionStore.Get()` or `sessionStore.Create()`.
  - Returns `(*session.Session, string)`.
- `setSessionCookie(w, sessionID)`: Sets the cookie in the response header.
- `handleSSE` & `handleMessages`:
  - Call `getSessionFromRequest`.
  - Call `setSessionCookie`.
  - Enhance context: `ctx := session.WithSession(r.Context(), currentSession)`.
  - Call `handler.HandleRequest/HandleNotification` with the enhanced `ctx`.
- Maintains maps `clients` (clientID -> client) and `sessions` (sessionID -> []*SSEClient) for message routing.

### Stdio Transport Session Handling (`pkg/transport/stdio/transport.go`)

- Implements `SetSessionStore` and holds `sessionStore` and `currentSession *session.Session` fields.
- `manageSession(baseCtx, req)`:
  - Checks if `req.Method == "initialize"` or `s.currentSession == nil`.
  - Calls `sessionStore.Create()` to get/replace `s.currentSession`.
  - Returns `session.WithSession(baseCtx, s.currentSession)`.
- Message Read Loop:
  - Parses request.
  - Calls `ctx := s.manageSession(scannerCtx, &request)`.
  - Calls `s.handleMessage(ctx, &request)` with the enhanced `ctx`.
- `handleRequest`/`handleNotification` pass the enhanced `ctx` to the main `RequestHandler`.

## Tool Registry and Invocation

1. **Tool Registration**:
   ```go
   func (r *Registry) RegisterTool(tool tools.Tool) {
     r.mu.Lock()
     defer r.mu.Unlock()
     r.tools[tool.GetName()] = tool
   }
   ```

2. **Tool Listing**:
   ```go
   func (r *Registry) ListTools(_ context.Context, cursor string) ([]protocol.Tool, string, error) {
     r.mu.RLock()
     defer r.mu.RUnlock()
     
     tools := make([]protocol.Tool, 0, len(r.tools))
     for _, t := range r.tools {
       tools = append(tools, t.GetToolDefinition())
     }
     // ...
   }
   ```

3. **Tool Invocation**:
   ```go
   func (r *Registry) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
     r.mu.RLock()
     defer r.mu.RUnlock()
     
     tool, ok := r.tools[name]
     if !ok {
       return nil, pkg.ErrToolNotFound
     }
     
     if handler, ok := r.handlers[name]; ok {
       return handler(ctx, tool, arguments)
     }
     
     // If no handler is registered, use the tool's Call method
     return tool.Call(ctx, arguments)
   }
   ```

## Shutdown Flow

The server shutdown flow handles graceful termination:

1. **Signal Handling**:
   ```go
   ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
   defer stop()
   ```

2. **Graceful Shutdown**:
   ```go
   g.Go(func() error {
     <-gctx.Done()
     logger.Info().Msg("Initiating graceful shutdown")
     shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
     defer shutdownCancel()
     if err := s.Stop(shutdownCtx); err != nil {
       logger.Error().Err(err).Msg("Error during shutdown")
       return err
     }
     logger.Info().Msg("Server stopped gracefully")
     return nil
   })
   ```

3. **Transport Cleanup**:
   - For SSE, clients are notified and connections closed.
   - For stdio, ongoing operations are cancelled.

## Conclusion

The go-go-mcp server now incorporates session management:

- **Centralized Session Logic**: `pkg/session` provides core types and context utilities.
- **Transport-Specific Handling**: SSE uses cookies, Stdio uses connection state, both managed within their respective transport implementations.
- **Context Propagation**: Session information is consistently passed down the call stack via `context.Context`.
- **Flexible Storage**: Supports `InMemorySessionStore` by default, with the ability to inject custom persistent stores via `WithSessionStore`.

This enables stateful interactions within the MCP framework, allowing tools and providers to access session-specific data when needed. 