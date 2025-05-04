# Design for Passing Session Information through Context in MCP

## Overview

This document outlines a design for adding session state management to the Model Context Protocol (MCP) implementation. The goal is to enable the MCP server to maintain state throughout a client session by passing a sessionId and sessionState through the context object.

## Components Involved

### Core Files

1.  `pkg/server/server.go`: Server structure and lifecycle management.
2.  `pkg/server/handler.go`: Request routing and base handling.
3.  `pkg/server/options.go`: Server configuration options.
4.  `pkg/session/session.go`: Session types and context utilities.
5.  `pkg/session/store.go`: SessionStore interface and InMemorySessionStore implementation.
6.  `pkg/transport/transport.go`: Transport interface definition.
7.  `pkg/transport/sse/transport.go`: SSE transport implementation.
8.  `pkg/transport/stdio/transport.go`: Stdio transport implementation.
9.  `pkg/providers.go`: Provider interface definitions.
10. `pkg/tools/providers/tool-registry/registry.go`: Example ToolProvider implementation.

## Design Approach

1.  **Create Session Package (`pkg/session`)**: Centralize session-related types (`Session`, `SessionState`), storage interface (`SessionStore`), an in-memory implementation (`InMemorySessionStore`), and context utilities (`WithSession`, `GetSessionFromContext`).
2.  **Integrate SessionStore into Server**: Add `sessionStore` field to `Server` struct, provide `WithSessionStore` option, and initialize a default `InMemorySessionStore` in `NewServer`.
3.  **Modify Transport Interface**: Add `SetSessionStore(store session.SessionStore)` method to `transport.Transport` interface.
4.  **Inject SessionStore into Transports**: Call `transport.SetSessionStore` in `Server.Start` to pass the configured store to the active transport.
5.  **Implement Session Handling in Transports**:
    *   **SSE (`pkg/transport/sse/transport.go`)**:
        *   Implement `SetSessionStore`.
        *   Use HTTP cookies (`mcp_session_id`) for session identification (`getSessionFromRequest`, `setSessionCookie`).
        *   Retrieve existing or create new sessions using the `SessionStore`.
        *   Enhance the request context using `session.WithSession` in `handleSSE` and `handleMessages` before passing it to the `RequestHandler`.
    *   **Stdio (`pkg/transport/stdio/transport.go`)**:
        *   Implement `SetSessionStore`.
        *   Maintain a single `currentSession` per connection.
        *   Create a new session via `SessionStore` upon receiving an `initialize` request (or implicitly on the first request if not `initialize`) using the `manageSession` helper.
        *   Enhance the context using `session.WithSession` in the message reading loop before passing it to the `RequestHandler`.
6.  **Utilize Session Context in Handler**: The `RequestHandler` methods receive the session-enhanced context from the transport layer. `handleInitialize` was updated to log the session ID and potentially include it in the response (protocol spec dependent). Other handlers pass the context downstream to providers.
7.  **Provider/Tool Access**: Providers and tools can retrieve the session from the context using `session.GetSessionFromContext(ctx)` if they need access to session state.

## Implementation Details (Summary)

### Session Management Package (`pkg/session`)

*   **Types**: `Session`, `SessionState` defined in `session.go`.
*   **Store**: `SessionStore` interface and `InMemorySessionStore` implementation in `store.go`. Uses `sync.RWMutex` for thread safety.
*   **Context Utilities**: `WithSession` and `GetSessionFromContext` in `session.go` using an unexported `contextKey`.

### Server Modifications (`pkg/server`)

*   `Server` struct in `server.go` now includes `sessionStore session.SessionStore`.
*   `NewServer` in `server.go` initializes `sessionStore` with `session.NewInMemorySessionStore()`.
*   `WithSessionStore` option added in `options.go`.
*   `Server.Start` in `server.go` calls `s.transport.SetSessionStore(s.sessionStore)`.

### Transport Modifications (`pkg/transport`)

*   `Transport` interface in `transport.go` includes `SetSessionStore(store session.SessionStore)`.
*   **SSE (`sse/transport.go`)**:
    *   Implements `SetSessionStore`.
    *   Adds `sessionStore` field.
    *   Uses `getSessionFromRequest` (checks cookie `mcp_session_id`, interacts with `sessionStore`) and `setSessionCookie`.
    *   Calls `session.WithSession(r.Context(), currentSession)` in handlers.
    *   Manages client-to-session mapping (`clients`, `sessions` maps).
*   **Stdio (`stdio/transport.go`)**:
    *   Implements `SetSessionStore`.
    *   Adds `sessionStore` and `currentSession` fields.
    *   Uses `manageSession` helper (checks for `initialize` method, interacts with `sessionStore`).
    *   Calls `session.WithSession(baseCtx, s.currentSession)` in the read loop.

### Request Handler Modifications (`pkg/server/handler.go`)

*   `handleInitialize` now uses `session.GetSessionFromContext(ctx)` to retrieve the session passed by the transport and logs its ID. It was also updated to potentially add the session ID to the `InitializeResult`.
*   Other handlers (`handlePromptsList`, `handleToolsCall`, etc.) receive the enhanced context and pass it to providers.

## Function Flow (Updated)

1.  Client connects (SSE establishes HTTP connection, Stdio starts).
2.  **Transport Layer**:
    *   Receives request (`handleSSE`/`handleMessages` for SSE, read loop for Stdio).
    *   **Identifies/Creates Session**:
        *   SSE: Checks `mcp_session_id` cookie, uses `sessionStore.Get` or `sessionStore.Create`. Sets cookie in response.
        *   Stdio: Checks if `initialize` request or if `currentSession` is nil, uses `sessionStore.Create` via `manageSession`.
    *   **Enhances Context**: Calls `session.WithSession(ctx, currentSession)`.
    *   Passes request and *enhanced context* to `RequestHandler`.
3.  `RequestHandler.HandleRequest(ctx, req)`:
    *   Receives enhanced context.
    *   Routes to specific handler (e.g., `handleInitialize`, `handleToolsCall`) based on `req.Method`, passing the enhanced context.
4.  Method-specific handler (e.g., `handleInitialize`, `handleToolsCall`):
    *   Operates using the enhanced context.
    *   `handleInitialize` logs session ID from context.
    *   Other handlers call provider methods, passing the enhanced context.
5.  Provider implementation (e.g., `Registry.CallTool`):
    *   Receives enhanced context.
    *   **Can optionally** access session info using `session.GetSessionFromContext(ctx)`.
    *   Passes context to tool handler or `tool.Call`.
6.  Tool implementation:
    *   Receives context.
    *   **Can optionally** access session info using `session.GetSessionFromContext(ctx)`.

## Considerations

1.  **Concurrency**: `InMemorySessionStore` uses `sync.RWMutex`. Ensure any custom stores or direct access to `Session.State` is thread-safe.
2.  **Session Lifetime**: `InMemorySessionStore` currently has no expiration. A cleanup mechanism (e.g., background ticker checking `LastAccessTime`) should be added for production use, especially for SSE. Stdio sessions implicitly end with the connection unless explicitly persisted.
3.  **Session Identification**:
    *   SSE relies on the `mcp_session_id` HTTP cookie. Security attributes (Secure, HttpOnly, SameSite) are set.
    *   Stdio manages one session per connection, reset by `initialize`.
4.  **State Size**: No limits currently enforced on `SessionState` size. Consider adding limits or using external stores for large states.
5.  **Persistence**: `InMemorySessionStore` is not persistent. Implementations requiring persistence need a different `SessionStore` (e.g., Redis, database).
6.  **Error Handling**: Added `transport.ProcessError` and `transport.ErrorToHTTPStatus` for better error reporting. 