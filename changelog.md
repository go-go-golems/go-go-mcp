# Added GitHub Pull Request Listing Command

Added a new command to list pull requests from GitHub repositories:
- Created list-github-pull-requests command with support for filtering by state, assignee, author, labels, and base branch
- Added draft PR filtering support
- Included comprehensive JSON output options for PR-specific fields

# Added Shell Tool Provider Debug Flags

Added command line flags to control ShellToolProvider debugging and tracing:
- Added --debug flag to enable detailed logging of tool calls and arguments
- Added --tracing-dir flag to save tool call input/output as JSON files
- Updated start command to pass debug settings to ShellToolProvider

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

Moved start and schema commands to their own files in cmd/go-go-mcp/cmds for better code organization and maintainability.

- Created cmd/go-go-mcp/cmds/start.go for start command
- Created cmd/go-go-mcp/cmds/schema.go for schema command
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

## HTML Fetch Command

Added `fetch-html` command to simplify HTML content from URLs:
- Supports fetching and simplifying multiple URLs
- Configurable limits for list items and table rows
- Optional markdown output format

## Enhanced simplify-html command with multiple input sources

Added support for processing multiple files and URLs in the simplify-html command. The command now accepts:
- Multiple files via --files flag (including stdin with -)
- Multiple URLs via --urls flag
- Outputs results as a list of source and data pairs

# Enhanced ShellToolProvider Debugging and Tracing

Added debug mode and tracing capabilities to ShellToolProvider for better debugging and audit capabilities:
- Added debug mode to log detailed information about tool calls and arguments
- Added tracing directory support to save input/output JSON files for each tool call
- Implemented functional options pattern for configuration
- Added timestamp-based file naming for trace files

# Switch to html-to-markdown Library

Replaced manual markdown conversion with html-to-markdown library for better and more consistent HTML to Markdown conversion:
- Removed manual markdown conversion code and maps
- Added html-to-markdown library integration
- Simplified TextSimplifier implementation
- Maintained existing logging and error handling

# HTML Selector Show Simplified Flag

Added a new flag to control whether simplified HTML is shown in the output to reduce verbosity.

- Added `--show-simplified` flag (default: false) to control whether simplified HTML is included in output
- Modified output to only include simplified HTML and context when explicitly requested

HTML Selector Template Control
Added ability to disable template rendering in the HTML selector tool.

- Added --no-template flag to disable template rendering
- Template rendering can now be explicitly disabled even when config file or extract options are used

## PubMed Search Command

Added a new shell command for searching PubMed and extracting structured data from search results.

- Added `examples/html-extract/pubmed.yaml` with search term and config file parameters
- Support for configurable maximum pages to scrape

## Prompto Shell Commands

Added shell command wrappers for prompto CLI:
- `prompto-list.yaml`: Lists all prompto entries
- `prompto-get.yaml`: Retrieves a specific prompto entry

## RAG Shell Commands

Added shell command wrappers for mento-service RAG operations:
- `rag-recent-documents.yaml`: Lists recent documents in the RAG system
- `rag-search.yaml`: Searches documents in the RAG system with a query

## Coaching Conversation Tools

Added shell command wrappers for accessing coaching conversation history:
- `recent-coaching-conversations.yaml`: Lists recent coaching conversations with detailed metadata
- `search-coaching-conversations.yaml`: Performs semantic search through coaching conversation history

These tools are specifically designed for LLMs to use in the context of coaching discussions.

## Google Calendar Integration

Added a new shell command for retrieving Google Calendar agenda using gcalcli with configurable date ranges.

- Added `examples/google/get-calendar.yaml` with support for custom start and end dates
- Added `examples/google/add-calendar-event.yaml` with comprehensive event creation options
- Added `examples/google/search-calendar.yaml` with extensive search and display options

## GitHub Issue Management

Added shell command for listing GitHub issues with comprehensive filtering options:
- Added `examples/github/list-issues.yaml` with support for filtering by state, assignee, author, labels, etc.
- Added JSON output formatting and web browser viewing options

# Fix string joining in GitHub issues list command
Fixed the join syntax in the GitHub issues list command template to use proper Go template string joining.

- Fixed join syntax in list-github-issues.yaml to use printf with join function

## Merge MCP Client into Server

Merged the MCP client functionality into the server as a subcommand for better code organization and maintainability.

- Moved client commands to `cmd/mcp-server/cmds/client/`
- Added client subcommand to server binary
- Updated package names and imports

## Update README for Merged Client/Server

Updated the README.md to reflect the new unified client/server architecture:
- Updated build instructions to show single binary build
- Added documentation for server and client modes
- Updated all command examples to use client subcommand
- Updated project structure to show new organization

## Rename Binary to go-go-mcp

Renamed the binary from mcp-server to go-go-mcp for consistency:
- Updated build instructions to use go-go-mcp as binary name
- Updated all command examples to use new binary name
- Updated configuration examples

## Update README for Merged Client/Server

Updated the README.md to reflect the new unified client/server architecture:
- Updated build instructions to show single binary build
- Added documentation for server and client modes
- Updated all command examples to use client subcommand
- Updated project structure to show new organization

Repository GetCommand Method
Added a GetCommand method to the Repository struct to easily find a single command by its full path name.
- Added GetCommand(name string) method that returns a single command by its full path name
- Provides a convenient way to find commands without dealing with prefix slices directly

Proper Command Type Handling
- Added proper type checking for different command types (WriterCommand, BareCommand, GlazeCommand)
- Implemented proper handling for WriterCommand
- Added panic stubs for BareCommand and GlazeCommand for future implementation

Parka Parameter Filter Integration
- Replaced custom parameter filtering with Parka's parameter filter system
- Improved middleware handling for parameter defaults, overrides, and filtering
- Removed custom parameter manager in favor of Parka's implementation
- Better consistency with other tools in the ecosystem

# Multi-Profile Configuration

Added a new configuration file with three profiles (all, github, google) to demonstrate profile-based tool loading from different directories.

- Created config.yaml with three distinct profiles
- Set up directory paths for GitHub and Google tools
- Configured default profile as 'all' to load all available tools

# XDG Config Path Support

Added support for XDG config directory for configuration files:
- Set default config file path to ~/.config/go-go-mcp/profiles.yaml
- Maintains backward compatibility with explicit --config-file flag
- Improved configuration file discovery and organization

## CORS Support for SSE Messages Endpoint

Added CORS headers and OPTIONS request handling to the /messages endpoint to fix cross-origin request issues.

- Added CORS headers to /messages endpoint
- Added support for OPTIONS preflight requests
- Fixed 405 Method Not Allowed errors for cross-origin requests

# Improved notification handling in SSE transport

Improved the SSE transport to handle notifications more efficiently by not waiting for responses when handling notification messages.

- Modified SSE transport to skip response waiting for notifications
- Added support for both empty ID and notifications/ prefix detection

# Improved SSE Notification Handling

Enhanced the SSE transport to handle notifications and responses separately for better efficiency and clarity:

- Added separate channels for notifications and responses
- Added notification handler support to Transport interface
- Updated SSE bridge to forward notifications to stdout
- Added default notification logging in client
- Improved notification detection and routing

# Added Notification Support to Stdio Transport

Added notification handling support to the stdio transport:
- Added notification handler to StdioTransport struct
- Implemented SetNotificationHandler method
- Added notification detection and handling in Send method
- Improved response handling to properly handle interleaved notifications

## Documentation: Added Pinocchio Integration Instructions

Added documentation about using Pinocchio to generate shell commands for go-go-mcp.

- Added section about adding go-go-mcp as a Pinocchio repository
- Added instructions for using the create-command template

## Documentation: Added Run Command Usage Instructions

Added documentation about using the run-command subcommand to execute shell commands directly as standalone command-line tools.

- Added section about using run-command with shell command YAML files
- Added examples for different types of commands (GitHub, Google Calendar, URL fetching)

## Documentation: Added Example Commands Overview

Added a comprehensive overview of example commands available in the examples directory.

- Added categorized listing of example commands (GitHub, Google Calendar, Web Content, etc.)
- Added brief descriptions for each command
- Improved discoverability of available tools

## Documentation: Added Schema Command and Improved Example Links

Enhanced the example commands documentation:
- Added information about using the schema command to view parameter documentation
- Converted example command paths to clickable markdown links
- Added examples of viewing command documentation using both schema and help

## Fix: Resolve Predeclared Identifier Conflict

Fixed linter error in randomInt function by renaming variables to avoid conflict with predeclared identifiers.

- Renamed min to minVal and max to maxVal in randomInt function

# Configuration Management Commands

Added a new `config` command group to manage go-go-mcp configuration files and profiles:
- `init`: Create a new minimal configuration file
- `edit`: Edit configuration in default editor
- `list-profiles`: List all available profiles
- `show-profile`: Show full configuration of a profile
- `add-tool`: Add tool directory or file to a profile
- `add-profile`: Create a new profile
- `duplicate-profile`: Create a new profile by duplicating an existing one

# YAML Editor and Config Management

Added a new YAML editor package in clay for manipulating YAML files while preserving comments and structure:
- Added YAMLEditor type with methods for manipulating YAML nodes
- Added helper functions for creating and manipulating YAML nodes
- Added support for preserving comments and formatting
- Added deep copy functionality for YAML nodes

Updated the config commands to use the new YAML editor:
- Added ConfigEditor type for managing go-go-mcp configuration
- Updated all config commands to use the new editor
- Added set-default-profile command
- Improved error handling and validation

# Profiles Path Helper

Added a helper package for managing profiles configuration paths:
- Added GetDefaultProfilesPath to get the default XDG config path
- Added GetProfilesPath to handle both default and custom paths
- Updated all config commands to use the new helpers
- Updated start command to use the new helpers

# Update Configuration Documentation with CLI Tools

Updated the configuration documentation to include the command-line configuration management tools:

- Added config command examples to README.md
- Updated configuration file tutorial with CLI tool usage
- Added CLI-based configuration workflow to MCP in Practice guide
- Improved documentation organization and clarity

## Claude Desktop Configuration Editor

Added functionality to manage Claude desktop configuration files through the go-go-mcp CLI.
This allows users to manage MCP server configurations for the Claude desktop application.

- ✨ Added `claude-config` command group with the following subcommands:
  - `init`: Initialize a new Claude desktop configuration file
  - `edit`: Edit the configuration file in your default editor
  - `add-mcp-server`: Add or update an MCP server configuration
  - `remove-mcp-server`: Remove an MCP server configuration
  - `list-servers`: List all configured MCP servers
- 🏗️ Added `ClaudeDesktopEditor` type for managing Claude desktop configuration files
- 📝 Configuration files are stored in the XDG config directory under `Claude/claude_desktop_config.json`

## Environment Variable Support for Claude Desktop Configuration

Added support for environment variables in MCP server configurations:
- ✨ Added `--env` flag to `claude-config add-mcp-server` command to set environment variables
- 🔧 Environment variables are stored in the `env` field of server configurations
- 📝 Updated `list-servers` command to display configured environment variables

## Improved Claude Desktop Configuration Command Output

Enhanced the output of Claude desktop configuration commands:
- 📝 Added detailed success messages for add-mcp-server and remove-mcp-server commands
- 🔍 Added configuration file path to command output for better visibility
- 💡 Added empty state handling in list-servers command

## Added Server Existence Check and Overwrite Flag

Added safety check when adding MCP servers and option to overwrite:
- 🛡️ Added check for existing servers in `add-mcp-server` command
- ✨ Added `--overwrite` flag to force update of existing servers
- 📝 Updated success messages to indicate whether server was added or updated

# Enhanced Claude Desktop Configuration Documentation

Updated the Claude desktop configuration documentation with:
- Comprehensive command-line examples for all claude-config commands
- Detailed examples of using init, edit, add-mcp-server, remove-mcp-server, and list-servers
- Clear explanations of command flags and options
- Improved organization and readability

# Added Server Enable/Disable Support

Added ability to temporarily disable MCP servers without removing their configuration:
- ✨ Added `disable-server` and `enable-server` commands to claude-config
- 🏗️ Added `DisabledServers` map to store configurations of disabled servers
- 📝 Updated list-servers command to show disabled status
- 🔧 Added helper functions for managing server state

# Added Log Tailing Support

Added ability to tail Claude log files in real-time:
- ✨ Added `tail` command to claude-config for monitoring log files
- 🔍 Support for tailing specific server logs by name
- 🎯 Added `--all` flag to tail all log files simultaneously
- 🛠️ Graceful shutdown support with Ctrl+C
- 📝 Real-time log monitoring with automatic file reopening

# Enhanced Log Tailing with Line History

Enhanced the tail command with line history support:
- ✨ Added `--lines/-n` flag to show last N lines when starting to tail
- 🔍 Efficient seeking to last N lines without reading entire file
- 🛠️ Proper handling of files without newline at end
- 📝 Updated command documentation with new flag

## Improved MCP Server Enable/Disable Functionality

Improved the enable/disable functionality for MCP servers to properly preserve server configurations when enabling/disabling them.

- Added `DisabledServers` map to store configurations of disabled servers
- Updated `EnableMCPServer` to move server config from disabled to enabled state
- Updated `DisableMCPServer` to move server config from enabled to disabled state
- Updated `ListServers` to show both enabled and disabled server configurations

## Parameter Value Validation and Casting

Enhanced parameter validation to return cast values along with validation errors. This allows for proper type conversion and sanitization of input values.

- Modified `CheckValueValidity` to return both the cast value and any validation errors
- Added `setReflectValue` method to handle setting reflect values with proper type casting
- Updated tests to verify cast values
- Improved error messages for invalid choices

# Improved YAML Editor Value Node Creation

Refactored YAML editor to have a more maintainable and recursive value node creation system:

- Extracted value node creation logic into a dedicated CreateValueNode method
- Made value creation process recursive for nested structures
- Improved error handling with more specific error messages
- Centralized value node creation logic for better maintainability

# Refactor Tool Provider Creation

Improved code organization by moving tool provider creation to server layer package.

- Moved tool provider creation to server layer package for better organization
- Made CreateToolProvider a public function for reuse
- Updated start command to use the new package function

## Server Tools Commands

Added server-side tool management commands for direct interaction with tool providers.

- Added `server tools list` command to list available tools directly from tool provider
- Added `server tools call` command to call tools directly without starting the server
- Reused server layer for configuration consistency

## Add Minimal Glazed Command Layer

Added a minimal version of the Glazed command layer (NewGlazedMinimalCommandLayer) that contains just the most commonly used parameters: print-yaml, print-parsed-parameters, load-parameters-from-file, and print-schema. This provides a simpler interface for basic command configuration.

- Added GlazedMinimalCommandSlug constant
- Added NewGlazedMinimalCommandLayer function

## Enhanced Glazed Command Layer Handling

Updated cobra command handling to support both full and minimal Glazed command layers:

- Added support for GlazedMinimalCommandLayer in cobra command processing
- Unified handling of common flags (print-yaml, print-parsed-parameters, etc.) between both layers
- Maintained backward compatibility with full GlazedCommandLayer features
- Added placeholder for schema printing functionality

# Transport Layer Refactoring

Implemented new transport layer architecture as described in RFC-01. This change:
- Creates a clean interface for different transport mechanisms
- Separates transport concerns from business logic
- Provides consistent error handling across transports
- Adds support for transport-specific options and capabilities

- Created new transport package with core interfaces and types
- Implemented SSE transport using new architecture
- Added transport options system
- Added standardized error handling

# Transport Layer Implementation

Added stdio transport implementation using new transport layer architecture:
- Implemented stdio transport with proper signal handling and graceful shutdown
- Added support for configurable buffer sizes and logging
- Added proper error handling and JSON-RPC message processing
- Added context-based cancellation and cleanup

# Server Layer Updates

Updated server implementation to use new transport layer:
- Refactored Server struct to use transport interface
- Added RequestHandler to implement transport.RequestHandler interface
- Updated server command to support multiple transport types
- Improved error handling and logging throughout server layer

# Enhanced SSE Transport

Added support for integrating SSE transport with existing HTTP servers:
- Added standalone and integrated modes for SSE transport
- Added GetHandlers method to get SSE endpoint handlers
- Added RegisterHandlers method for router integration
- Added support for path prefixes and middleware
- Improved configuration options for HTTP server integration

# Transport Interface Refactoring

Simplified transport interface to use protocol types directly instead of custom types.
- Removed duplicate type definitions from transport package
- Use protocol.Request/Response/Notification types directly
- Improved type safety by removing interface{} usage

# Transport Request ID Handling

Added proper request ID handling to transport package:
- Added IsNotification helper to check for empty/null request IDs
- Improved notification detection for JSON-RPC messages
- Consistent handling of request IDs across transports

# Transport ID Type Conversion

Added helper functions for converting between string and JSON-RPC ID types:
- Added StringToID to convert string to json.RawMessage
- Added IDToString to convert json.RawMessage to string
- Improved type safety in ID handling across transports

# Simplified UI DSL

Simplified the UI DSL by removing class attributes and creating distinct title and text elements:
- Removed class attributes from all components
- Added dedicated title element for headings
- Simplified text element to be just for paragraphs
- Updated documentation to reflect changes

# UI DSL Implementation

Created a YAML-based UI DSL for defining simple user interfaces. The DSL supports common UI components like buttons, text, inputs, textareas, checkboxes, and lists with a clean and intuitive syntax.

- Added `ui-dsl.yaml` with component definitions and examples
- Added documentation in `README.md`
- Included support for common attributes across all components
- Added nested component support for complex layouts

# UI Server Implementation
Added a new UI server that can render YAML UI definitions using HTMX and Bootstrap:
- Created a new command `ui-server` that serves UI definitions from YAML files
- Implemented templ templates for rendering UI components
- Added support for various UI components like buttons, inputs, forms, etc.
- Used HTMX for dynamic interactions and Bootstrap for styling

# Halloween-themed UI Examples
Added a collection of Halloween-themed example pages using the UI DSL:
- Created welcome page with spooky navigation
- Added haunted house tour booking form
- Created costume contest voting interface
- Added Halloween party RSVP form with fun options
- Created trick-or-treat checklist for safety

# UI DSL Structure Update
Updated the UI DSL to use a top-level components list for better sequence handling:
- Changed UIDefinition to use a list of components instead of a map
- Updated all example pages to use the new structure
- Modified templates to handle the new component list format
- Improved component rendering to handle nested components

# SSE Transport Port Configuration

Improved port configuration handling in SSE transport by properly parsing the provided address.

- Added proper port parsing from SSE options address
- Ensures port configuration is correctly propagated from command line to transport

## Google Maps CLI Tool

Added a new CLI tool for interacting with Google Maps API. The tool uses API key authentication and provides a foundation for implementing various Google Maps API features.

- Created basic command structure with API key authentication
- Set up Maps client initialization
- Added environment variable support for API key

## Dinosaur-themed Example Pages

Added new example pages showcasing the UI DSL capabilities with a fun dinosaur theme:
- Added dino-facts.yaml with interactive dinosaur facts and newsletter signup
- Added build-a-dino.yaml with a form to create custom dinosaurs
- Added dino-quiz.yaml featuring an interactive dinosaur knowledge quiz
- Added dino-park-guide.yaml with park areas, safety guidelines and tour booking

## UI Server File Watching

Added file watching capabilities to the UI server for automatic page reloading:
- Added Clay watcher integration to monitor YAML page files
- Implemented automatic page reloading on file changes
- Added page removal on file deletion
- Added graceful shutdown handling

## UI Server Logging Improvements

Improved logging in the UI server by:
- Switched to zerolog for structured logging
- Implemented glazed command pattern for better CLI structure
- Added proper log levels and context fields

## Fixed Page Handler Registration

Fixed issues with page handler registration in UI server:
- Added mutex protection for concurrent access to pages and handlers
- Created single ServeMux instance instead of recreating on each request
- Fixed handler registration to avoid conflicts during page reloading
- Added proper locking around map access and handler registration
- Improved logging to show both file path and URL path

## Improved Page Loading

Enhanced page loading in UI server:
- Added proper recursive page loading from subdirectories
- Clear existing pages before reloading to avoid stale entries
- Added debug logging for directory traversal
- Improved error handling for directory walking

## Improved Dynamic Page Routing

Enhanced page routing in UI server:
- Added dynamic page routing using a routes map
- Fixed handler registration conflicts during page reloading
- Moved all pages under /pages/ prefix for better organization
- Added proper cleanup of routes on page removal
- Improved URL path handling and logging

## Fixed Route Handler Order

Fixed routing issues in UI server:
- Fixed handler registration order to ensure correct path matching
- Added proper path validation for root handler
- Added debug logging for request routing
- Improved handler type consistency