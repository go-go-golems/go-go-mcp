## Tool Result Helper Methods

Added helper methods to easily create ToolResults with different content types (text, JSON, images, resources).
This simplifies the creation of tool responses and reduces boilerplate code.

- Added NewToolResult and NewErrorToolResult constructors
- Added NewTextContent for text responses
- Added NewJSONContent and MustNewJSONContent for JSON data
- Added NewImageContent for base64-encoded images
- Added NewResourceContent for resource data
- Updated echo tool to use new helpers 

## Tool Result Functional Options

Added functional options pattern for creating ToolResults, providing a more flexible and chainable API.

- Added NewToolResultWithOptions constructor with functional options
- Added WithText for adding text content
- Added WithJSON for adding JSON-serialized content
- Added WithImage for adding image content
- Added WithResource for adding resource content
- Added WithError for creating error results
- Added WithContent for adding raw ToolContent
- Updated echo tool to use new functional options pattern 

## Command Transport and Argument Handling

Added support for launching external commands as MCP servers and improved argument handling.

- Added NewCommandStdioTransport to launch and communicate with external commands
- Added command transport type to client CLI
- Changed command arguments to use string slice for better argument handling
- Added proper process cleanup in Close method 

## Transport Simplification

Made command transport the default and only direct connection option.

- Removed redundant stdio transport option
- Made command transport the default with sensible defaults
- Updated documentation to reflect the simplified transport options
- Improved command transport examples in README

## Documentation Updates

Added comprehensive Running section to README.md with detailed examples.

- Added build instructions for client and server
- Added examples for all transport types (stdio, SSE, command)
- Added debug mode and version information usage
- Updated project structure to include client implementation 

## Command Flag Simplification

Combined command and arguments into a single flag for simpler usage.

- Merged --command and --args flags into a single --command flag
- First argument in the list is the command, remaining are arguments
- Updated documentation with new command usage examples
- Improved error handling for empty command list 

# Enhanced Transport Debugging

Added comprehensive debug logging to transport implementations:

- Added structured logging to SSE transport for connection lifecycle and events
- Added structured logging to stdio transport for command execution and I/O
- Added error logging with context for both transports
- Added RawJSON logging for request/response payloads

# Enhanced Logging Configuration

Added more granular control over logging configuration in both client and server:

- Added `--log-level` flag to configure zerolog level (default: info)
- Added `--with-caller` flag to show caller information in logs (default: true)
- Improved log level parsing with fallback to info level
- Added short file:line format for caller information 

# Service Layer Refactoring

Refactored the MCP server to use a proper service layer:
- Extracted service interfaces for prompts, resources, tools, and initialization
- Created default implementations for all services
- Moved stdio server to its own package
- Improved error handling by removing panic-based JSON marshaling
- Added proper context support throughout the service layer 

# Graceful Shutdown Implementation

Added graceful shutdown support to handle interrupt signals (SIGINT, SIGTERM) properly. This ensures that the server can clean up resources and close connections gracefully when stopped.

- Added Stop method to Server and Transport interface
- Implemented graceful shutdown for SSE server with proper client connection cleanup
- Added graceful shutdown for stdio server
- Updated main program to handle interrupt signals and coordinate shutdown
- Added proper error handling during shutdown process 

# Context-Based Server Control

Refactored the server to use context.Context for better control over server lifecycle and cancellation.

- Added context support to Transport interface methods (Start and Stop)
- Updated SSE server to use context for connection handling and shutdown
- Updated stdio server to handle context cancellation
- Added context with timeout for graceful shutdown in main program
- Improved error handling with context cancellation 

# Enhanced Server Logging

Improved server logging when running as a command:

- Added [SERVER] tag prefix to all server log messages
- Configured stderr output to be forwarded to client
- Preserved log level and timestamp formatting
- Improved log message readability for debugging 

# Command Server Logging

Enhanced command server logging:

- Forward command server's stderr to client's stderr
- Allows seeing server logs directly in client's terminal
- Helps with debugging command server issues 

# Improved Command Shutdown

Enhanced command server shutdown handling:

- Added fallback to Process.Kill() if interrupt signal fails
- Improved error handling for process termination
- Added detailed debug logging with process IDs
- Fixed issue with programmatic interrupt signals not working 

# Improved Signal Handling

Enhanced signal handling in stdio server:

- Added direct signal handling in stdio server
- Fixed issue with signals not breaking scanner reads
- Added detailed debug logging for signal flow
- Improved shutdown coordination between scanner and signals 

# Improved Process Group Handling

Enhanced process group and signal handling in stdio transport:

- Set up command server in its own process group
- Send signals to entire process group instead of just the main process
- Added fallback to direct process signals if process group handling fails
- Improved debug logging for process and signal management
- Fixed issue with signals not being properly received by the server 

# Fixed Channel Close Race Condition

Fixed a race condition in the stdio server where the done channel could be closed multiple times:

- Added sync.Once to ensure done channel is only closed once
- Fixed potential panic during server shutdown
- Improved shutdown coordination between signal handling and Stop method
- Enhanced logging around channel closing operations 

# Simplified Server Shutdown

Simplified stdio server shutdown by using only context-based coordination:

- Removed redundant done channel in favor of context cancellation
- Added dedicated scanner context for cleaner shutdown
- Simplified shutdown logic and error handling
- Improved logging messages for shutdown events 

# Tool Interface Improvements

Made Tool an interface with accessors and context-aware Call method for better extensibility and context propagation.

- Changed Tool from struct to interface with GetName, GetDescription, GetInputSchema and Call methods
- Added ToolImpl as a basic implementation of the Tool interface
- Updated Registry, Client and CLI to use the new Tool interface
- Added context support throughout the tool invocation chain 

# SSE Server Message Handling

Added support for all MCP protocol methods in the SSE server implementation:
- Updated SSEServer to use all services (prompt, resource, tool, initialize)
- Added handlers for all protocol methods (prompts/list, prompts/get, resources/list, resources/read, tools/list, tools/call)
- Improved error handling and response formatting 

# Improved JSON-RPC Error Handling

Enhanced error handling in SSE server to better comply with JSON-RPC 2.0 specification:

- Added proper error codes for different error types (parse error, invalid request, method not found, invalid params, internal error)
- Improved error message formatting and consistency
- Added error data field with detailed error information
- Fixed JSON-RPC version validation

# SSE Protocol Compliance

Removed non-standard endpoint event from SSE server to better align with the official protocol specification.

- Removed endpoint event from SSE server implementation 

# Async SSE Client

Made the SSE client run asynchronously to prevent blocking on subscription:
- Added initialization synchronization channel
- Moved SSE subscription to a goroutine
- Improved error handling and state management
- Added proper mutex protection for shared state 