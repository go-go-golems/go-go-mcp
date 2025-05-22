# Embeddable MCP Server

The embeddable package provides a simple way for Go applications to add MCP (Model Context Protocol) server capabilities with minimal code changes.

## Features

- **Simple Integration**: Add MCP server capabilities with just a few lines of code
- **Cobra Integration**: Provides a standard `mcp` subcommand for existing cobra applications
- **Multiple Tool Registration Methods**: Support for function-based, struct-based, and reflection-based tool registration
- **Session Management**: Automatic session management with context-based access
- **Transport Flexibility**: Support for both stdio and SSE transports
- **Schema Generation**: Automatic JSON schema generation from Go structs and function signatures

## Quick Start

### Basic Usage

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
```

Usage:
```bash
# Start the MCP server with stdio transport
myapp mcp start

# Start with SSE transport on port 3001
myapp mcp start --transport sse --port 3001

# List available tools
myapp mcp list-tools
```

### Session Management

The embeddable API provides automatic session management through Go's context system:

```go
func counterHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    // Session is automatically available via context
    sess, ok := session.GetSessionFromContext(ctx)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
    }

    // Store and retrieve data from session
    sess.SetData("counter", 42)
    value, exists := sess.GetData("counter")
    
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Session: %s, Counter: %v", sess.ID, value)),
    ), nil
}
```

### Struct-Based Tool Registration

Register tools directly from struct methods:

```go
type DatabaseService struct {
    connectionString string
}

type QueryArgs struct {
    Query string `json:"query" description:"SQL query to execute"`
    Limit int    `json:"limit,omitempty" description:"Maximum number of rows to return"`
}

func (db *DatabaseService) ExecuteQuery(ctx context.Context, args QueryArgs) (*protocol.ToolResult, error) {
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Executed: %s", args.Query)),
    ), nil
}

func main() {
    // ... setup code ...
    
    config := embeddable.NewServerConfig()
    dbService := &DatabaseService{connectionString: "..."}
    
    err := embeddable.RegisterStructTool(config, "execute_query", dbService, "ExecuteQuery")
    if err != nil {
        log.Fatal(err)
    }
    
    // ... rest of setup ...
}
```

## Enhanced Features (v2)

Inspired by [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go), we now provide enhanced APIs for even more convenient tool development:

### Enhanced Tool Registration

```go
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
)
```

### Enhanced Argument Access

```go
func formatTextHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
    // Type-safe argument access with defaults and validation
    text, err := args.RequireString("text")
    if err != nil {
        return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
    }
    
    format := args.GetString("format", "lowercase")
    maxLength := args.GetInt("max_length", 100)
    enabled := args.GetBool("enabled", true)
    
    // Bind to struct
    var config MyConfig
    if err := args.BindArguments(&config); err != nil {
        return nil, err
    }
    
    // ... implementation
}
```

### Tool Annotations

Add semantic hints about tool behavior:

```go
embeddable.WithReadOnlyHint(true),        // Tool doesn't modify environment
embeddable.WithDestructiveHint(false),    // Tool won't cause destructive changes
embeddable.WithIdempotentHint(true),      // Repeated calls have no additional effect
embeddable.WithOpenWorldHint(false),      // Tool doesn't interact with external entities
```

## API Reference

### Server Configuration Options

#### Core Options
- `WithName(name string)` - Set server name
- `WithVersion(version string)` - Set server version
- `WithServerDescription(description string)` - Set server description

#### Transport Options
- `WithDefaultTransport(transport string)` - Set default transport ("stdio" or "sse")
- `WithDefaultPort(port int)` - Set default port for SSE transport

#### Tool Registration Options
- `WithTool(name, handler, opts...)` - Register a tool with handler function
- `WithEnhancedTool(name, handler, opts...)` - Register with enhanced argument handling
- `WithToolRegistry(registry)` - Use a custom tool registry

#### Advanced Options
- `WithSessionStore(store)` - Use a custom session store
- `WithMiddleware(middleware...)` - Add middleware functions
- `WithHooks(hooks)` - Add lifecycle hooks

### Tool Configuration Options

#### Basic Tool Options
- `WithDescription(desc string)` - Set tool description
- `WithSchema(schema interface{})` - Set custom JSON schema
- `WithExample(name, description, args)` - Add usage example

#### Enhanced Tool Options
- `WithEnhancedDescription(desc)` - Set tool description
- `WithReadOnlyHint(bool)` - Mark tool as read-only
- `WithDestructiveHint(bool)` - Mark tool as potentially destructive
- `WithIdempotentHint(bool)` - Mark tool as idempotent
- `WithOpenWorldHint(bool)` - Mark tool as interacting with external world

#### Property Configuration (Enhanced Tools)
- `WithStringProperty(name, opts...)` - Add string property with rich options
- `WithIntProperty(name, opts...)` - Add integer property
- `WithNumberProperty(name, opts...)` - Add number property
- `WithBooleanProperty(name, opts...)` - Add boolean property
- `WithArrayProperty(name, opts...)` - Add array property
- `WithObjectProperty(name, opts...)` - Add object property

#### Property Options
- `PropertyDescription(desc)` - Set property description
- `PropertyRequired()` - Mark property as required
- `PropertyTitle(title)` - Set display title
- `DefaultString(value)`, `DefaultNumber(value)`, `DefaultBool(value)` - Set defaults
- `StringEnum(values...)` - Restrict to enum values
- `MinLength(n)`, `MaxLength(n)` - String length constraints
- `Minimum(n)`, `Maximum(n)` - Number range constraints
- `StringPattern(regex)` - Regex pattern validation
- `MinItems(n)`, `MaxItems(n)` - Array size constraints
- `UniqueItems(bool)` - Array uniqueness constraint

#### Convenience Schema Builders (Legacy)
- `WithStringArg(name, description, required)` - Add string parameter
- `WithIntArg(name, description, required)` - Add integer parameter
- `WithBoolArg(name, description, required)` - Add boolean parameter
- `WithFileArg(name, description, required)` - Add file parameter

### Command Structure

The generated `mcp` command provides:

```
myapp mcp
├── start              # Start the MCP server
│   ├── --transport    # Transport type (stdio, sse)
│   ├── --port         # Port for SSE transport
│   └── --config       # Configuration file path
├── list-tools         # List available tools
└── test-tool          # Test a specific tool
```

## Advanced Features

### Middleware Support

Add middleware to process tool calls:

```go
func loggingMiddleware(next embeddable.ToolHandler) embeddable.ToolHandler {
    return func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
        log.Printf("Calling tool with args: %v", args)
        result, err := next(ctx, args)
        log.Printf("Tool result: %v, error: %v", result, err)
        return result, err
    }
}

// Usage
embeddable.AddMCPCommand(rootCmd,
    embeddable.WithMiddleware(loggingMiddleware),
    // ... other options
)
```

### Reflection-Based Registration

Register functions directly using reflection:

```go
func addNumbers(ctx context.Context, args AddArgs) (*protocol.ToolResult, error) {
    // Implementation
}

// Register the function
err := embeddable.RegisterFunctionTool(config, "add", addNumbers)
```

### Custom Tool Registry

Use a custom tool registry for advanced scenarios:

```go
registry := tool_registry.NewRegistry()
// ... register tools manually ...

err := embeddable.AddMCPCommand(rootCmd,
    embeddable.WithToolRegistry(registry),
    // ... other options
)
```

## Examples

See the `examples/` directory for complete working examples:

- `examples/basic/` - Basic tool registration and usage
- `examples/session/` - Session management demonstration  
- `examples/struct/` - Struct-based tool registration
- `examples/enhanced/` - Enhanced API with rich argument handling and property configuration

## Migration from Manual Setup

The embeddable API is designed to work alongside existing go-go-mcp implementations:

1. **Gradual Adoption**: Can coexist with manual server implementations
2. **Tool Reuse**: Existing tools can be easily wrapped for the embeddable API
3. **Configuration Compatibility**: Supports existing configuration patterns
4. **Feature Parity**: All core go-go-mcp features available

## Error Handling

The embeddable API provides consistent error handling:

```go
func myHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    if someCondition {
        // Return error result (shown to user)
        return protocol.NewErrorToolResult(
            protocol.NewTextContent("Something went wrong"),
        ), nil
    }
    
    // Return system error (logged, generic error shown to user)
    return nil, fmt.Errorf("system error: %w", err)
}
```

## Best Practices

1. **Tool Naming**: Use clear, descriptive names for tools
2. **Schema Design**: Provide comprehensive schemas with descriptions
3. **Error Messages**: Use user-friendly error messages
4. **Session Management**: Use sessions for stateful operations
5. **Validation**: Validate inputs before processing
6. **Documentation**: Add examples to help users understand tool usage

## Contributing

The embeddable package is part of the go-go-mcp project. See the main project README for contribution guidelines.