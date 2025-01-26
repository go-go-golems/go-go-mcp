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

Updated the SSE server implementation to fully comply with the MCP protocol specification:
- Added proper CORS headers for cross-origin requests
- Implemented unique session ID generation using UUID
- Added initial endpoint event with session ID
- Ensured proper SSE headers according to spec

# Async SSE Client

Made the SSE client run asynchronously to prevent blocking on subscription:
- Added initialization synchronization channel
- Moved SSE subscription to a goroutine
- Improved error handling and state management
- Added proper mutex protection for shared state 

# Context-Based Transport Control

Refactored all transports to use context.Context for better control over cancellation and timeouts:
- Added context support to Transport interface methods (Send and Close)
- Updated SSE transport to use context for initialization and event handling
- Updated stdio transport to use context for command execution and response handling
- Removed channel-based cancellation in favor of context
- Added proper context propagation throughout the client API 

## Optional Session ID and Request ID Handling

Made session ID optional in SSE transport by using a default session when none is provided. Also ensured request ID handling follows the MCP specification.

- Made session ID optional in server and client SSE transport
- Added default session support in server
- Removed session ID requirement from client
- Updated request ID handling to follow MCP spec 

# Simplified zerolog caller configuration

Replaced custom caller marshaler with zerolog's built-in Caller() for simpler and more maintainable code.

- Removed custom caller marshaler function in both client and server
- Using zerolog's built-in Caller() functionality
- Added proper zerolog/log import in server

## Logger Consolidation

Consolidated logging setup to use a consistent logger throughout the application:
- Set up global logger with console writer in main.go
- Removed duplicate logger creation in createClient
- Ensured client.go uses the local logger consistently

# Server Logger Consolidation

Consolidated logging setup in the server components:
- Set up global logger with console writer in server's main.go
- Removed duplicate logger creation in server startup
- Ensured consistent logger propagation through server components

# Client Transport Logger Consolidation

Consolidated logging setup in client transports:
- Updated SSE transport to use passed logger instead of creating its own
- Updated stdio transport to use passed logger instead of creating its own
- Modified transport constructors to accept logger parameter
- Ensured consistent logger propagation through all client components

# Improved SSE Server Client Management

Improved the SSE server's client management to better handle multiple clients and sessions:
- Added unique client IDs for better tracking and debugging
- Improved session management to handle multiple clients per session
- Added client metadata tracking (creation time, remote address, user agent)
- Fixed race conditions in client management
- Better error handling for invalid sessions 

# SSE Client Endpoint Handling

Enhanced SSE client to properly handle endpoint events and session management:
- Added proper endpoint event handling and waiting
- Added session ID extraction and storage
- Improved initialization flow to wait for endpoint event
- Added better error handling for endpoint event timeout 

## Empty Array Response Fix

Ensure list operations return empty arrays instead of null when no results are available. This fixes type validation errors in the client.

- Added structured response types (ListPromptsResult, ListResourcesResult, ListToolsResult) for list operations
- Improved type handling for list operation results with proper interface conversions
- Ensured empty arrays are always returned instead of null

## Structured List Operation Responses

Added proper response types for list operations to ensure consistent JSON encoding:

- Created ListPromptsResult, ListResourcesResult, and ListToolsResult types with correct protocol types
- Fixed type mismatches between service interfaces and response types
- Improved type handling with proper interface conversions
- Ensured empty arrays are always returned instead of null
- Fixed JSON response structure to match API specification 

# Fix SSE Server Shutdown

Fixed an issue where the SSE server would hang during shutdown due to improper handling of client connections.

- Added proper context cancellation for client goroutines
- Added WaitGroup to track and wait for client goroutines to finish
- Improved shutdown coordination between HTTP server and client cleanup
- Added timeout handling for client goroutine cleanup 

## Response Type Consolidation

Consolidated list operation response types into a shared location:

- Moved ListPromptsResult, ListResourcesResult, and ListToolsResult to responses.go
- Updated both SSE and stdio servers to use the shared types
- Ensured consistent response structure across all server implementations 

# Improved Logger Output

Added conditional logger output based on terminal detection. When output is not going to a terminal, the logger will not use color escape codes, making the output more readable in log files and non-terminal environments.

- Updated both client and server to detect terminal output
- Added golang.org/x/term dependency for terminal detection
- Logger now uses NoColor mode when output is not going to a terminal

## SQLite Tool and Tool Reorganization

Added a new SQLite tool to query the Cursor state database and reorganized tools into separate files for better maintainability.

- Added new SQLite tool to query Cursor's state.vscdb database
- Split tools into separate files in pkg/tools directory
- Moved echo tool to its own file
- Moved fetch tool to its own file
- Added go-sqlite3 dependency

## SQLite Tool YAML Output

Changed SQLite tool output format from JSON to YAML for better readability.

- Changed output format to YAML
- Added gopkg.in/yaml.v3 dependency
- Updated tool description to clarify SQLite dot command limitations

# Cursor Database Tools

Added tools for analyzing and querying the Cursor database:
- Added conversation management tools for retrieving and searching conversations
- Added code analysis tools for extracting and tracking code blocks
- Added context management tools for accessing file references and conversation context

# Changelog

## Unified Protocol Dispatcher

Refactored SSE and stdio servers to use a common protocol dispatcher to handle MCP protocol methods. This improves code reuse and maintainability by:
- Creating a central dispatcher package for handling all protocol methods
- Unifying error handling and response creation
- Adding proper context handling for session IDs
- Reducing code duplication between transport implementations

## Embed Static Files in Prompto Server Binary

Improved deployment by embedding static files into the binary instead of serving from disk.

- Updated serve.go to use Go's embed package for static files

# Refactor prompts commands to use glazed framework

Refactored the prompts list and execute commands to use the glazed framework for better structured data output and consistent command line interface.

- Converted ListPrompts to a GlazeCommand for structured output
- Converted ExecutePrompt to a WriterCommand for formatted text output
- Added proper parameter handling using glazed parameter layers
- Improved error handling and command initialization

# Change prompt name from argument to flag

Changed the execute prompt command to use a --prompt-name flag instead of a positional argument for better usability and consistency with glazed framework.

- Added --prompt-name flag to execute command
- Removed positional argument requirement
- Updated command help text to reflect new usage

# Add client settings layer

Added a dedicated settings layer for MCP client configuration to improve reusability and consistency:

- Created new ClientParameterLayer for transport and server settings
- Updated client helper to use settings layer instead of cobra flags
- Updated list and execute commands to use the new layer
- Improved error handling and configuration management

# Convert start and schema commands to Glazed commands

Refactored the start and schema commands to use the Glazed command framework for better parameter handling and consistency.

- Converted start command to BareCommand
- Converted schema command to WriterCommand
- Added proper parameter definitions and settings structs
- Improved error handling and command organization

# Extract start and schema commands to separate files

Moved start and schema commands to their own files in cmd/mcp-server/cmds for better code organization and maintainability.

- Created cmd/mcp-server/cmds/start.go for start command
- Created cmd/mcp-server/cmds/schema.go for schema command
- Updated main.go to use the extracted commands
- Improved code organization and readability

# Add settings struct for schema command

Added proper settings struct for schema command to better handle parameter initialization:

- Created SchemaCommandSettings struct with file parameter
- Updated schema command to use settings struct
- Improved error handling for file parameter

# Command Loading Documentation

Added comprehensive documentation for loading commands from YAML files and repositories:

- Added command loader interface explanation
- Added examples for loading single files and repositories
- Added best practices and troubleshooting guide
- Added example shell command YAML format

## Reflection-based Tool Implementation

Added a new ReflectTool implementation that uses reflection to create tools from Go functions:
- Added ReflectTool struct that implements the Tool interface
- Added NewReflectTool constructor that generates JSON schema from function parameters
- Implemented Call method that uses reflection to invoke functions and handle results
- Added smart result type handling (text for primitives, JSON for complex types)
- Uses protocol.ToolResult helpers for proper result handling

## Tool Registration Documentation

Added comprehensive documentation on registering tools with MCP:
- Detailed guide on using ReflectTool for Go functions
- Examples of implementing the Tool interface directly
- Guide to YAML-based shell commands
- Best practices and examples for all approaches
- Added doc/registering-tools.md

# Add Simple Weather Tool

Added a simple weather tool using reflection to demonstrate tool registration.

- Added getWeather reflect tool as an example
- Added WeatherData struct for weather information

## URL Content Fetching Command

Added a new shell command `fetch-url` that uses lynx to fetch and extract text content from URLs. This command is designed to be LLM-friendly with detailed descriptions and supports multiple URLs and link handling options.

- Added `examples/shell-commands/fetch-url.yaml` with comprehensive documentation
- Supports batch URL processing
- Optional link reference removal
- Simple stdout output with URL separators

## Simple Diary Command

Added a new shell command `diary-append` that appends timestamped entries to a diary file in markdown format.

- Added `examples/shell-commands/diary-append.yaml` with markdown formatting
- Automatically adds timestamps as headings
- Supports markdown in messages
- Maintains clean formatting with proper spacing

## YouTube Transcript Fetching Command

Added a new shell command `fetch-transcript` that downloads transcripts/subtitles from YouTube videos using youtube-dl.

- Added `examples/shell-commands/fetch-transcript.yaml` with comprehensive options
- Supports multiple languages and auto-generated subtitles
- Includes language listing capability
- Downloads in SRT format for easy reading

# Add Logging Support

Added comprehensive logging support using zerolog:
- Added `--debug` flag to enable logging (disabled by default)
- Added `--log-level` flag to control log level (default: debug)
- Logs are written to stderr with timestamp and caller information
- Added detailed logging throughout HTML processing for better debugging

# Fix Text Simplification for Important Elements

Fixed text simplification to preserve important HTML elements:
- Added detection of important elements (links, formatting, etc.)
- Prevented collapsing of nodes containing important elements
- Added more detailed trace logging for debugging
- Improved handling of inline formatting elements

# Add Markdown Conversion Support

Added support for converting HTML to Markdown format:
- Added `--markdown` flag to enable markdown conversion
- Converts important elements (links, formatting) to markdown syntax
- Preserves links with proper markdown link syntax
- Supports various markdown formatting (bold, italic, code, etc.)
- Reduces token usage while maintaining readability

## Test Organization and Cleanup

Improved test organization and readability:
- Reorganized text_simplifier_test.go into focused test functions
- Created common test struct and helper function
- Improved test case naming and organization

## HTML Simplification Strategy Improvements

Added new TryText strategy for HTML simplification that attempts text simplification without falling back to ExtractText, providing more control over text extraction behavior.

## Add Complete HTML Document Test Cases

Added test cases to verify the HTML simplifier's handling of complete HTML documents including DOCTYPE and top-level elements.

- Added TestSimplifier_ProcessHTML_CompleteDocument test function
- Test covers DOCTYPE, html, head, and body elements
- Verifies handling of meta tags and document attributes
- Tests both simple and complex document structures

Enhanced HTML Selector Modes

Added support for both select and filter modes in HTML simplification:
- New `mode` field in selectors config: "select" or "filter"
- Select mode keeps only matching elements and their parents
- Filter mode removes matching elements
- Selectors are applied in order: first selects, then filters

HTML Simplifier Tag Format Enhancement
Enhanced the HTML simplifier to include id and class attributes in the tag name for better readability and CSS-like format.

- Modified tag format to include id and classes (e.g. div#myid.class1.class2)
- Removed id and class from regular attributes list
- Improved readability of HTML structure in YAML output

# Convert test-html-selector to glazed framework

Converted the test-html-selector tool to use the glazed framework for better CLI integration and consistency with other tools. The file field has been removed from the config file and is now required as a command line argument.

- Converted test-html-selector to use glazed framework
- Removed file field from config file
- Added --sample-count and --context-chars command line flags
- Updated documentation to reflect new command structure

# Enhanced HTML Selector Output

Enhanced the test-html-selector tool to use the full HTML simplification structure:

- Added HTML simplification options from simplify-html tool
- Changed output format to use full Document structure for both HTML and context
- Added support for markdown and text simplification
- Updated documentation with new output format examples

# HTML Simplifier Document List Support

Changed the HTML simplifier to return lists of documents instead of single documents:

- Modified ProcessHTML to return []Document instead of Document
- Updated test-html-selector to handle document lists
- Changed output format to show arrays of documents for HTML and context
- Updated documentation with new output format examples

## Add Clay logging initialization to HTML tools

Added proper Clay logging initialization to test-html-selector and simplify-html tools to match mcp-server's logging setup.

# Added Show Context and Path Control Flags

Added flags to control the display of context and paths in the HTML selector tool output for better customization:

- Added --show-context flag to control display of context around matched elements (default: false)
- Added --show-path flag to control display of element paths (default: true)
- Updated documentation to reflect new options

# Added Command Line Selector Support

Added support for specifying selectors directly via command line flags, making the config file optional:

- Added --select-css flag for specifying CSS selectors (can be used multiple times)
- Added --select-xpath flag for specifying XPath selectors (can be used multiple times)
- Made config file optional while ensuring at least one selector is provided
- Selectors from both config file and command line are combined
- Added automatic naming for command line selectors (css_1, css_2, xpath_1, etc.)

# Added Enhanced Data Extraction Features

Added new extraction modes to the HTML selector tool for more flexible data processing:

- Added --extract-data flag to output all matches in a YAML map format (selector name -> matches)
- Added --extract-template flag to render matches through a Go template
- Both modes process all matches without sample count limits
- Template mode allows for custom formatting of extracted data

# Simplified Data Extraction

Simplified data extraction by merging --extract and --extract-data flags:

- Changed --extract to always output data in YAML format
- Removed redundant --extract-data flag
- Maintained template support with --extract-template and config file templates
- Improved consistency in data output formats

# Fixed Selector Match Count

Fixed the selector match count to show the total number of matches instead of the truncated sample count:

- Count now shows total number of matches before sample truncation
- Sample count limit only affects displayed samples, not the total count
- Provides more accurate match statistics while keeping output manageable

# Added Sprig Template Functions

Added Sprig template functions to enhance template capabilities:

- Added full set of Sprig text template functions
- Available in both file templates and config templates
- Includes string manipulation, math, date formatting, and more
- Removed custom add function in favor of Sprig's math functions

# Added Configuration Descriptions

Added description fields to improve configuration documentation:

- Added top-level description field to describe the overall configuration purpose
- Added per-selector description field to document each selector's purpose
- Updated example configurations with detailed descriptions
- Improved self-documentation of configuration files

HTML Selector Tutorial Examples
Added comprehensive tutorial examples to demonstrate HTML selector usage:
- Basic text extraction examples showing simple selectors
- Nested content examples showing parent-child relationships
- Tables and lists examples showing structured data extraction
- XPath examples showing advanced selection techniques

Each example includes both HTML and YAML files with detailed descriptions and comments.

# Raw Data Extraction Option

Added a new flag to allow extracting raw data without applying templates.

- Added --extract-data flag to skip template processing and output raw YAML data
- Updated documentation to reflect new option
- Maintains backwards compatibility with existing template functionality

# Multiple Source Support

Added support for processing multiple HTML files and URLs in a single run:

- Changed --input flag to --files for processing multiple files
- Added --urls flag for processing multiple URLs
- Updated output format to include source information
- Added proper error handling for each source
- Improved template support to handle multiple sources

## Add VSCode launch configuration for test-html-selector
Added a new launch configuration to make it easier to debug the test-html-selector tool with example files.

- Added Test HTML Selector launch configuration in .vscode/launch.json

## HTML Extraction Commands

Added two new MCP commands for HTML data extraction:
- `get-html-extraction-tutorial`: Command to display comprehensive HTML extraction documentation
- `html-extraction`: Command to extract data from HTML documents using selectors