# Registering Tools with MCP Server

The MCP server supports multiple ways to register tools, each with its own use cases and benefits. This guide covers the three main approaches:

1. Using Go Functions with Reflection
2. Implementing the Tool Interface
3. Using YAML-based Shell Commands

## 1. Using Go Functions with Reflection

The simplest way to create a tool is by using a regular Go function and the `ReflectTool` implementation. This approach automatically generates JSON schema from your function's parameters.

### Example:

```go
func GetWeather(city string, includeWind bool) WeatherData {
    return WeatherData{
        City: city,
        Temperature: 23.0,
        WindSpeed: 10.0,
    }
}

func RegisterTools(registry *tools.Registry) error {
    // Create a new ReflectTool from the function
    weatherTool, err := tools.NewReflectTool(
        "getWeather",
        "Get weather information for a city",
        GetWeather,
    )
    if err != nil {
        return err
    }

    // Register the tool
    registry.RegisterTool(weatherTool)
    return nil
}
```

### Benefits:
- Automatic JSON schema generation from function parameters
- Type safety through Go's type system
- Simple to write and maintain
- Smart result handling (primitives as text, complex types as JSON)

### Limitations:
- Less control over the exact JSON schema
- Limited to Go function signatures
- No direct access to raw arguments

## 2. Implementing the Tool Interface

For more control over tool behavior, you can implement the `Tool` interface directly. This gives you full control over schema definition and argument handling.

### The Tool Interface:

```go
type Tool interface {
    GetName() string
    GetDescription() string
    GetInputSchema() json.RawMessage
    GetToolDefinition() protocol.Tool
    Call(ctx context.Context, arguments map[string]interface{}) (*protocol.ToolResult, error)
}
```

### Example Implementation:

```go
type CustomTool struct {
    *tools.ToolImpl
}

func NewCustomTool() (*CustomTool, error) {
    schema := `{
        "type": "object",
        "properties": {
            "query": {
                "type": "string",
                "description": "Search query"
            },
            "limit": {
                "type": "integer",
                "default": 10
            }
        },
        "required": ["query"]
    }`

    toolImpl, err := tools.NewToolImpl(
        "search",
        "Search for items",
        json.RawMessage(schema),
    )
    if err != nil {
        return nil, err
    }

    return &CustomTool{ToolImpl: toolImpl}, nil
}

func (t *CustomTool) Call(ctx context.Context, arguments map[string]interface{}) (*protocol.ToolResult, error) {
    query := arguments["query"].(string)
    limit := 10
    if l, ok := arguments["limit"].(float64); ok {
        limit = int(l)
    }

    results := performSearch(query, limit)
    
    return protocol.NewToolResult(
        protocol.WithJSON(results),
    ), nil
}

func RegisterTools(registry *tools.Registry) error {
    tool, err := NewCustomTool()
    if err != nil {
        return err
    }
    registry.RegisterTool(tool)
    return nil
}
```

### Benefits:
- Complete control over JSON schema
- Custom argument handling and validation
- Access to raw argument map
- Can implement complex tool behavior

### Limitations:
- More code to write and maintain
- Need to handle type assertions manually
- Must create JSON schema manually

## 3. Using YAML-based Shell Commands

For tools that primarily execute shell commands or scripts, MCP provides a YAML-based configuration system. This is particularly useful for operations tasks, system administration, or wrapping existing CLI tools.

See the [Shell Commands Documentation](shell-commands.md) for a complete reference of the YAML format and features.

### Example YAML Tool:

```yaml
name: git-sync
short: Synchronize git repositories
flags:
  - name: repos-dir
    type: string
    help: Directory containing git repos
    required: true
  - name: branch
    type: string
    help: Branch to sync
    default: main
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  find {{ .Args.repos_dir }} -type d -name ".git" | while read -r gitdir; do
    repo=$(dirname "$gitdir")
    cd "$repo"
    git fetch origin
    git checkout {{ .Args.branch }}
    git pull origin {{ .Args.branch }}
  done
```

### Registering Shell Commands:

```go
func RegisterShellTools(registry *tools.Registry) error {
    // Create a shell tool provider
    provider := cmds.NewShellToolProvider()

    // Load commands from a directory
    err := provider.LoadCommandsFromDirectory("./examples/shell-commands")
    if err != nil {
        return err
    }

    // Or load a single command
    err = provider.LoadCommandFromFile("./examples/git-sync.yaml")
    if err != nil {
        return err
    }

    // Register the provider with the registry
    registry.RegisterProvider(provider)
    return nil
}
```

### Benefits:
- No coding required for simple command-line tools
- Easy to maintain and modify
- Supports complex shell scripts
- Built-in templating for arguments
- Great for system administration tasks

### Limitations:
- Limited to shell commands and scripts
- May be less performant than native Go tools
- Error handling is more challenging
- Limited type safety

## Best Practices

1. **Choose the Right Approach**:
   - Use **ReflectTool** for simple Go functions with clear input/output
   - Use **Custom Tool Implementation** for complex tools needing full control
   - Use **YAML Shell Commands** for system operations and CLI wrapping

2. **Error Handling**:
   - Always return errors through the `protocol.ToolResult` with `WithError`
   - Include meaningful error messages
   - Consider adding error context when wrapping errors

3. **Documentation**:
   - Provide clear descriptions for tools and parameters
   - Document expected input formats
   - Include examples in tool descriptions

4. **Type Safety**:
   - Use strong typing when possible (Go functions)
   - Validate arguments early
   - Handle type conversions safely

5. **Testing**:
   - Write tests for tool implementations
   - Test edge cases and error conditions
   - Consider integration tests for shell commands

## Example: Combining Approaches

Here's an example of registering tools using all three approaches:

```go
func RegisterAllTools(registry *tools.Registry) error {
    // 1. Register reflection-based tools
    weatherTool, err := tools.NewReflectTool(
        "getWeather",
        "Get weather information",
        GetWeather,
    )
    if err != nil {
        return err
    }
    registry.RegisterTool(weatherTool)

    // 2. Register custom tool implementation
    searchTool, err := NewCustomTool()
    if err != nil {
        return err
    }
    registry.RegisterTool(searchTool)

    // 3. Register shell commands
    shellProvider := cmds.NewShellToolProvider()
    err = shellProvider.LoadCommandsFromDirectory("./examples/shell-commands")
    if err != nil {
        return err
    }
    registry.RegisterProvider(shellProvider)

    return nil
}
```

## Conclusion

MCP's flexible tool registration system allows you to choose the most appropriate approach for each tool. By combining these approaches, you can create a powerful and maintainable tool ecosystem that serves various use cases efficiently. 