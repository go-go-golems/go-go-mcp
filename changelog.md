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