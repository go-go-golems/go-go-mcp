# Embeddable MCP Server API Design

## Overview

This document outlines the design for an embeddable MCP (Model Context Protocol) server API that allows existing Go applications to easily add MCP server capabilities. The goal is to provide a simple, library-based approach that enables applications to expose their functionality as MCP tools with minimal code changes.

**Implementation Status: ✅ MVP COMPLETED**

The embeddable MCP server has been successfully implemented and is ready for use. See [Implementation Status](#implementation-status) for details.

## Inspiration

The design takes inspiration from the [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) library, which provides a clean, simple API for creating MCP servers. However, our implementation is tailored to work seamlessly with the existing go-go-mcp architecture and provides additional features like configuration management and built-in tool registration.

## Design Goals

1. **Minimal Integration**: Applications should be able to add MCP server capabilities with just a few lines of code ✅
2. **Cobra Integration**: Provide a standard `mcp` subcommand that can be easily added to existing cobra applications ✅
3. **Simple Tool Registration**: Make it easy to register existing functions as MCP tools ✅
4. **Configuration Management**: Leverage existing go-go-mcp configuration and tool discovery mechanisms ✅
5. **Transport Flexibility**: Support both stdio and SSE transports out of the box ✅
6. **Session Management**: Automatic session management with context-based access for stateful tools ✅
7. **Extensibility**: Allow for custom tool providers and advanced configurations ✅

## Implementation Status

### ✅ Completed Features (MVP)

#### Core API (v1)
- [x] `AddMCPCommand()` main entry point
- [x] Server configuration with options pattern
- [x] Tool registration API with `ToolHandler` signature
- [x] Basic cobra command integration
- [x] Integration with existing go-go-mcp architecture

#### Server Configuration Options
- [x] `WithName()` - Set server name
- [x] `WithVersion()` - Set server version  
- [x] `WithServerDescription()` - Set server description
- [x] `WithDefaultTransport()` - Configure transport type
- [x] `WithDefaultPort()` - Configure SSE port
- [x] `WithTool()` - Register individual tools
- [x] `WithToolRegistry()` - Use custom registry
- [x] `WithSessionStore()` - Custom session storage
- [x] `WithMiddleware()` - Tool middleware support
- [x] `WithHooks()` - Lifecycle hooks

#### Tool Registration
- [x] Simple function-based tool registration
- [x] Schema generation with builder pattern
- [x] Convenience functions (`WithStringArg`, `WithIntArg`, etc.)
- [x] Description and example support
- [x] `RegisterSimpleTools()` for batch registration

#### Advanced Features
- [x] Session management via context
- [x] Struct-based tool registration with reflection
- [x] Function-based tool registration with reflection
- [x] Automatic JSON schema generation from Go types
- [x] Middleware support for tool calls
- [x] Both stdio and SSE transport support

#### Command Structure
- [x] `mcp start` - Start the server
- [x] `mcp list-tools` - List available tools
- [x] `mcp test-tool` - Test individual tools
- [x] `mcp config` - Configuration management (placeholder)

#### Enhanced API (v2) - Inspired by mark3labs/mcp-go
- [x] Enhanced argument access with `Arguments` wrapper
- [x] Type-safe argument getters (`GetString`, `RequireInt`, etc.)
- [x] Flexible type conversion for all basic types
- [x] Slice argument support with validation
- [x] `BindArguments()` for struct binding
- [x] `WithEnhancedTool()` for enhanced tool registration
- [x] Rich property configuration API
- [x] Tool annotations (ReadOnlyHint, DestructiveHint, etc.)
- [x] Advanced JSON schema generation
- [x] Property options (defaults, validation, constraints)

## Session Management Approach

The embeddable API uses go-go-mcp's existing context-based session management. Every tool handler receives a `context.Context` that contains session information accessible via `session.GetSessionFromContext(ctx)`. This provides:

- **Automatic Setup**: No explicit session handling required ✅
- **Simple Access**: Session available in every tool handler via context ✅
- **Persistent State**: Data stored in sessions persists across tool calls within the same connection ✅
- **Thread Safety**: All session operations are automatically thread-safe ✅

## Core API Design

### 1. Main Entry Point

```go
package mcp

import (
    "github.com/spf13/cobra"
    "github.com/go-go-golems/go-go-mcp/pkg/embeddable"
)

// AddMCPCommand adds a standard 'mcp' subcommand to an existing cobra application
func AddMCPCommand(rootCmd *cobra.Command, opts ...ServerOption) error {
    mcpCmd := embeddable.NewMCPCommand(opts...)
    rootCmd.AddCommand(mcpCmd)
    return nil
}
```

### 2. Server Configuration

```go
package embeddable

// ServerOption configures the embeddable MCP server
type ServerOption func(*ServerConfig) error

// ServerConfig holds the configuration for the embeddable server
type ServerConfig struct {
    Name        string
    Version     string
    Description string
    
    // Tool registration
    toolRegistry *tool_registry.Registry
    
    // Transport options
    defaultTransport string
    defaultPort      int
    
    // Configuration options
    enableConfig     bool
    configFile       string
    
    // Internal servers
    internalServers  []string
    
    // Advanced options
    sessionStore     session.SessionStore
    middleware       []ToolMiddleware
    hooks           *Hooks
}

// Core options
func WithName(name string) ServerOption
func WithVersion(version string) ServerOption  
func WithServerDescription(description string) ServerOption

// Transport options
func WithDefaultTransport(transport string) ServerOption
func WithDefaultPort(port int) ServerOption

// Tool registration options
func WithTool(name string, handler ToolHandler, opts ...ToolOption) ServerOption
func WithEnhancedTool(name string, handler EnhancedToolHandler, opts ...EnhancedToolOption) ServerOption
func WithToolRegistry(registry *tool_registry.Registry) ServerOption

// Advanced options
func WithSessionStore(store session.SessionStore) ServerOption
func WithMiddleware(middleware ...ToolMiddleware) ServerOption
func WithHooks(hooks *Hooks) ServerOption
```

### 3. Tool Registration API

```go
// ToolHandler is a simplified function signature for tool handlers
// Session information is available via session.GetSessionFromContext(ctx)
type ToolHandler func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error)

// EnhancedToolHandler provides enhanced argument access (v2 API)
type EnhancedToolHandler func(ctx context.Context, args Arguments) (*protocol.ToolResult, error)

// ToolOption configures individual tools
type ToolOption func(*ToolConfig) error

type ToolConfig struct {
    Description string
    Schema      interface{} // Can be a struct, JSON schema string, or json.RawMessage
    Examples    []ToolExample
}

func WithDescription(desc string) ToolOption
func WithSchema(schema interface{}) ToolOption
func WithExample(name, description string, args map[string]interface{}) ToolOption

// Convenience functions for common tool patterns
func WithStringArg(name, description string, required bool) ToolOption
func WithIntArg(name, description string, required bool) ToolOption
func WithBoolArg(name, description string, required bool) ToolOption
func WithFileArg(name, description string, required bool) ToolOption

// Enhanced tool options (v2 API)
type EnhancedToolOption func(*EnhancedToolConfig) error

func WithEnhancedDescription(desc string) EnhancedToolOption
func WithReadOnlyHint(readOnly bool) EnhancedToolOption
func WithDestructiveHint(destructive bool) EnhancedToolOption
func WithIdempotentHint(idempotent bool) EnhancedToolOption
func WithOpenWorldHint(openWorld bool) EnhancedToolOption

// Property configuration for enhanced tools
func WithStringProperty(name string, opts ...PropertyOption) EnhancedToolOption
func WithIntProperty(name string, opts ...PropertyOption) EnhancedToolOption
func WithNumberProperty(name string, opts ...PropertyOption) EnhancedToolOption
func WithBooleanProperty(name string, opts ...PropertyOption) EnhancedToolOption
func WithArrayProperty(name string, opts ...PropertyOption) EnhancedToolOption
func WithObjectProperty(name string, opts ...PropertyOption) EnhancedToolOption
```

### 4. Enhanced Argument Access (v2 API)

```go
// Arguments provides enhanced argument access with type safety
type Arguments interface {
    // Type-safe getters with defaults
    GetString(key string, defaultValue ...string) string
    GetInt(key string, defaultValue ...int) int
    GetFloat64(key string, defaultValue ...float64) float64
    GetBool(key string, defaultValue ...bool) bool
    
    // Required getters that return errors
    RequireString(key string) (string, error)
    RequireInt(key string) (int, error)
    RequireFloat64(key string) (float64, error)
    RequireBool(key string) (bool, error)
    
    // Slice support
    GetStringSlice(key string, defaultValue ...[]string) []string
    GetIntSlice(key string, defaultValue ...[]int) []int
    
    // Struct binding
    BindArguments(target interface{}) error
    
    // Raw access
    Raw() map[string]interface{}
}
```

### 5. Simplified Registration Helpers

```go
// RegisterSimpleTools provides a very easy way to register multiple tools
func RegisterSimpleTools(config *ServerConfig, tools map[string]ToolHandler) error

// RegisterStructTool automatically creates a tool from a struct and method
func RegisterStructTool(config *ServerConfig, name string, obj interface{}, methodName string) error

// RegisterFunctionTool creates a tool from a function using reflection
func RegisterFunctionTool(config *ServerConfig, name string, fn interface{}) error
```

## Usage Examples

### 1. Basic Usage - Simple Tool Registration

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/spf13/cobra"
    "github.com/go-go-golems/go-go-mcp/pkg/embeddable"
    "github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "My application",
    }
    
    // Add MCP server capability
    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("MyApp MCP Server"),
        embeddable.WithVersion("1.0.0"),
        embeddable.WithServerDescription("Example MCP server"),
        embeddable.WithTool("greet", greetHandler,
            embeddable.WithDescription("Greet a person"),
            embeddable.WithStringArg("name", "Name of the person to greet", true),
        ),
        embeddable.WithTool("calculate", calculateHandler,
            embeddable.WithDescription("Perform basic calculations"),
            embeddable.WithIntArg("a", "First number", true),
            embeddable.WithIntArg("b", "Second number", true),
            embeddable.WithStringArg("operation", "Operation to perform (+, -, *, /)", true),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

func greetHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    name, ok := args["name"].(string)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("name must be a string")), nil
    }
    
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Hello, %s!", name)),
    ), nil
}

func calculateHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    a, ok := args["a"].(float64)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("a must be a number")), nil
    }
    
    b, ok := args["b"].(float64)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("b must be a number")), nil
    }
    
    operation, ok := args["operation"].(string)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("operation must be a string")), nil
    }
    
    var result float64
    switch operation {
    case "+":
        result = a + b
    case "-":
        result = a - b
    case "*":
        result = a * b
    case "/":
        if b == 0 {
            return protocol.NewErrorToolResult(protocol.NewTextContent("division by zero")), nil
        }
        result = a / b
    default:
        return protocol.NewErrorToolResult(protocol.NewTextContent("unsupported operation")), nil
    }
    
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Result: %g", result)),
    ), nil
}
```

Usage:
```bash
# Start the MCP server with stdio transport
myapp mcp start

# Start with SSE transport on port 3001
myapp mcp start --transport sse --port 3001

# List available tools
myapp mcp list-tools

# Test a tool
myapp mcp test-tool greet
```

### 2. Enhanced API Usage (v2)

```go
package main

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/spf13/cobra"
    "github.com/go-go-golems/go-go-mcp/pkg/embeddable"
    "github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "enhanced-app",
        Short: "Enhanced application with v2 API",
    }
    
    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("Enhanced MCP Server"),
        embeddable.WithVersion("2.0.0"),
        embeddable.WithEnhancedTool("format_text", formatTextHandler,
            embeddable.WithEnhancedDescription("Format text with various options"),
            embeddable.WithReadOnlyHint(true),
            embeddable.WithIdempotentHint(true),
            embeddable.WithStringProperty("text",
                embeddable.PropertyDescription("Text to format"),
                embeddable.PropertyRequired(),
                embeddable.MinLength(1),
            ),
            embeddable.WithStringProperty("format",
                embeddable.PropertyDescription("Format type"),
                embeddable.StringEnum("uppercase", "lowercase", "title"),
                embeddable.DefaultString("lowercase"),
            ),
            embeddable.WithIntProperty("max_length",
                embeddable.PropertyDescription("Maximum length of output"),
                embeddable.Minimum(1),
                embeddable.Maximum(1000),
                embeddable.DefaultNumber(100),
            ),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    rootCmd.Execute()
}

func formatTextHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
    // Type-safe argument access with defaults and validation
    text, err := args.RequireString("text")
    if err != nil {
        return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
    }
    
    format := args.GetString("format", "lowercase")
    maxLength := args.GetInt("max_length", 100)
    
    // Format the text
    var result string
    switch format {
    case "uppercase":
        result = strings.ToUpper(text)
    case "lowercase":
        result = strings.ToLower(text)
    case "title":
        result = strings.Title(text)
    default:
        result = text
    }
    
    // Truncate if necessary
    if len(result) > maxLength {
        result = result[:maxLength] + "..."
    }
    
    return protocol.NewToolResult(
        protocol.WithText(result),
    ), nil
}
```

### 3. Advanced Usage - Custom Registry and Configuration

```go
package main

import (
    "context"
    "log"
    
    "github.com/spf13/cobra"
    "github.com/go-go-golems/go-go-mcp/pkg/embeddable"
    "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "advanced-app",
        Short: "Advanced application with custom MCP setup",
    }
    
    // Create custom tool registry
    registry := tool_registry.NewRegistry()
    
    // Register tools using the existing pattern
    registerCustomTools(registry)
    
    // Add MCP command with advanced configuration
    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("Advanced MCP Server"),
        embeddable.WithVersion("2.0.0"),
        embeddable.WithToolRegistry(registry),
        embeddable.WithDefaultTransport("sse"),
        embeddable.WithDefaultPort(3002),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

func registerCustomTools(registry *tool_registry.Registry) {
    // Use existing tool registration patterns
    // This allows leveraging existing tools and patterns
}
```

### 4. Struct-based Tool Registration

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/spf13/cobra"
    "github.com/go-go-golems/go-go-mcp/pkg/embeddable"
    "github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

type DatabaseService struct {
    connectionString string
}

type QueryArgs struct {
    Query string `json:"query" description:"SQL query to execute"`
    Limit int    `json:"limit,omitempty" description:"Maximum number of rows to return"`
}

func (db *DatabaseService) ExecuteQuery(ctx context.Context, args QueryArgs) (*protocol.ToolResult, error) {
    // Implementation here
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Executed query: %s", args.Query)),
    ), nil
}

func main() {
    rootCmd := &cobra.Command{
        Use:   "db-app",
        Short: "Database application",
    }
    
    config := embeddable.NewServerConfig()
    dbService := &DatabaseService{connectionString: "..."}
    
    err := embeddable.RegisterStructTool(config, "execute_query", dbService, "ExecuteQuery")
    if err != nil {
        log.Fatal(err)
    }
    
    err = embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("Database MCP Server"),
        embeddable.WithToolRegistry(config.GetToolRegistry()),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    rootCmd.Execute()
}
```

### 5. Session Management Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/spf13/cobra"
    "github.com/go-go-golems/go-go-mcp/pkg/embeddable"
    "github.com/go-go-golems/go-go-mcp/pkg/protocol"
    "github.com/go-go-golems/go-go-mcp/pkg/session"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "session-app",
        Short: "Application demonstrating session management",
    }
    
    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("Session Demo MCP Server"),
        embeddable.WithVersion("1.0.0"),
        embeddable.WithTool("get_counter", getCounterHandler,
            embeddable.WithDescription("Get the current counter value for this session"),
        ),
        embeddable.WithTool("increment_counter", incrementCounterHandler,
            embeddable.WithDescription("Increment the counter for this session"),
            embeddable.WithIntArg("amount", "Amount to increment by", false),
        ),
        embeddable.WithTool("reset_counter", resetCounterHandler,
            embeddable.WithDescription("Reset the counter for this session"),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

func getCounterHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    // Session information is automatically available via context
    sess, ok := session.GetSessionFromContext(ctx)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
    }
    
    // Get counter value from session
    counterVal, ok := sess.GetData("counter")
    counter := 0
    if ok {
        counter, _ = counterVal.(int)
    }
    
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Counter value: %d (Session: %s)", counter, sess.ID)),
    ), nil
}

func incrementCounterHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    // Access session via context
    sess, ok := session.GetSessionFromContext(ctx)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
    }
    
    // Get increment amount (default to 1)
    amount := 1
    if amountVal, ok := args["amount"].(float64); ok {
        amount = int(amountVal)
    }
    
    // Get current counter value
    counterVal, ok := sess.GetData("counter")
    counter := 0
    if ok {
        counter, _ = counterVal.(int)
    }
    
    // Increment and store back to session
    counter += amount
    sess.SetData("counter", counter)
    
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Counter incremented by %d to %d (Session: %s)", 
            amount, counter, sess.ID)),
    ), nil
}

func resetCounterHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    sess, ok := session.GetSessionFromContext(ctx)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
    }
    
    sess.SetData("counter", 0)
    
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Counter reset to 0 (Session: %s)", sess.ID)),
    ), nil
}
```

Usage:
```bash
# Start the server
session-app mcp start

# The session will be maintained across tool calls within the same connection
# Each tool call can access and modify session state via context
```

## Implementation Architecture

### 1. Package Structure

```
pkg/embeddable/
├── README.md           # Comprehensive documentation
├── IMPLEMENTATION.md   # Implementation status and details
├── server.go          # Main server configuration and creation
├── command.go         # Cobra command creation
├── tools.go           # Tool registration helpers
├── reflection.go      # Reflection-based tool registration
├── arguments.go       # Enhanced argument handling (v2)
├── enhanced_tools.go  # Enhanced tool API (v2)
└── examples/          # Usage examples
    ├── basic/         # Basic usage example
    ├── session/       # Session management example
    ├── struct/        # Struct-based registration example
    └── enhanced/      # Enhanced API demonstration
```

### 2. Integration with Existing Architecture

The embeddable server leverages the existing go-go-mcp architecture:

- **Transport Layer**: Uses existing `pkg/transport/stdio` and `pkg/transport/sse` ✅
- **Server Core**: Uses existing `pkg/server.Server` with custom configuration ✅
- **Tool Registry**: Extends `pkg/tools/providers/tool-registry.Registry` ✅
- **Configuration**: Optionally integrates with `pkg/config` for advanced setups ✅
- **Session Management**: Uses existing `pkg/session` for session-aware tools ✅

#### Session Management Integration

The embeddable API directly leverages go-go-mcp's existing session management system:

- **Automatic Session Handling**: Sessions are automatically created and managed by the transport layer ✅
- **Context-Based Access**: Session information is always available in tool handlers via `session.GetSessionFromContext(ctx)` ✅
- **Thread-Safe Operations**: All session operations are thread-safe using mutex protection ✅
- **Persistent State**: Session data persists across tool calls within the same MCP connection ✅

```go
// Example of accessing session in any tool handler
func myToolHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    // Session is automatically available via context
    sess, ok := session.GetSessionFromContext(ctx)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
    }
    
    // Use session for persistent state
    sess.SetData("key", "value")          // Store data
    value, exists := sess.GetData("key")  // Retrieve data
    sess.DeleteData("key")                // Delete data
    
    return protocol.NewToolResult(protocol.WithText("Operation completed")), nil
}
```

The implementation wraps the user's `ToolHandler` to work with the existing `tool_registry.Registry`:

```go
func (r *Registry) registerToolHandler(name string, handler ToolHandler) {
    r.RegisterToolWithHandler(tool, func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
        return handler(ctx, arguments)
    })
}
```

### 3. Command Structure

The generated `mcp` command has the following structure:

```
myapp mcp
├── start              # Start the MCP server ✅
│   ├── --transport    # Transport type (stdio, sse) ✅
│   ├── --port         # Port for SSE transport ✅
│   ├── --config       # Configuration file path ✅
│   └── --internal-servers # Built-in tools to enable (planned)
├── list-tools         # List available tools ✅
├── test-tool          # Test a specific tool ✅
└── config             # Configuration management (placeholder) ✅
    ├── init           # Initialize configuration
    ├── edit           # Edit configuration
    └── show           # Show current configuration
```

### 4. Tool Schema Generation

The library provides automatic schema generation from:

1. **Struct tags**: Using JSON tags and custom description tags ✅
2. **Function signatures**: Using reflection to analyze parameters ✅
3. **Manual schemas**: JSON Schema objects or strings ✅
4. **Builder pattern**: Fluent API for schema construction ✅

### 5. Error Handling and Validation

- **Input validation**: Automatic validation based on schemas ✅
- **Error wrapping**: Consistent error handling and reporting ✅
- **Graceful degradation**: Handle missing or invalid tools gracefully ✅
- **Logging integration**: Use structured logging for debugging ✅

## Advanced Features

### 1. Session Management

The embeddable API leverages go-go-mcp's existing robust session management system:

- **Automatic Session Creation**: Sessions are automatically created and managed by the transport layer ✅
- **Context-based Access**: Session information is always available via `session.GetSessionFromContext(ctx)` ✅
- **Persistent State**: Session data persists across tool calls within the same connection ✅
- **Thread-safe Operations**: Session state operations are thread-safe using mutex protection ✅

```go
// Access session in any tool handler
func myHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    sess, ok := session.GetSessionFromContext(ctx)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
    }
    
    // Store data in session
    sess.SetData("key", "value")
    
    // Retrieve data from session
    value, exists := sess.GetData("key")
    
    // Delete data from session
    sess.DeleteData("key")
    
    return protocol.NewToolResult(protocol.WithText("Session operation completed")), nil
}
```

### 2. Middleware Support

```go
type ToolMiddleware func(next ToolHandler) ToolHandler

// Built-in middleware (implemented)
func LoggingMiddleware() ToolMiddleware
func ValidationMiddleware() ToolMiddleware
func RateLimitingMiddleware(limit int) ToolMiddleware
func AuthenticationMiddleware(validator func(ctx context.Context) error) ToolMiddleware
```

### 3. Resource Provider Integration

```go
// Future enhancement
func WithResourceProvider(provider pkg.ResourceProvider) ServerOption
func WithResource(uri string, handler ResourceHandler, opts ...ResourceOption) ServerOption
```

### 4. Prompt Provider Integration

```go
// Future enhancement
func WithPromptProvider(provider pkg.PromptProvider) ServerOption
func WithPrompt(name string, handler PromptHandler, opts ...PromptOption) ServerOption
```
