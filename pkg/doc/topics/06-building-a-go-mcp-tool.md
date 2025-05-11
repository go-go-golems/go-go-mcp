---
Title: Building MCP Tools in Go Go MCP
Slug: mcp-tools
Short: Learn how to create custom MCP tools for go-go-mcp.
Topics:
  - tools
  - development
  - extensions
Commands:
  - run-tools
Flags:
  - tool-name
  - arguments
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This guide will walk you through the process of creating custom MCP tools for go-go-mcp. MCP tools allow you to extend the functionality of go-go-mcp by implementing interfaces that can be called by AI models and other clients through the MCP protocol.

## Table of Contents

1. [Introduction to MCP Tools](#introduction-to-mcp-tools)
2. [Tool Architecture](#tool-architecture)
3. [Creating a Basic Tool](#creating-a-basic-tool)
4. [Tool Registration](#tool-registration)
5. [Input Validation with JSON Schema](#input-validation-with-json-schema)
6. [Returning Results](#returning-results)
7. [Advanced Tool Development](#advanced-tool-development)
8. [Session Management](#session-management)
9. [Real-World Examples](#real-world-examples)
10. [Best Practices](#best-practices)
11. [Troubleshooting](#troubleshooting)

## Introduction to MCP Tools

MCP (Model Control Protocol) tools are executable components that can be invoked by AI models and other clients. They provide a way to extend the functionality of the go-go-mcp system by implementing new capabilities that can be used during conversations.

### What is an MCP Tool?

An MCP tool consists of:

- A **name** that uniquely identifies the tool
- A **description** that explains what the tool does
- An **input schema** that defines the parameters the tool accepts
- A **handler function** that executes the tool's logic
- A **result** that is returned to the caller

### When to Create a Tool

You should consider creating a custom tool when:

- You need to provide AI models with access to external systems
- You want to implement specialized functionality not available in existing tools
- You need to integrate with specific data sources or APIs
- You want to automate complex workflows

## Tool Architecture

The go-go-mcp system uses a layered architecture for tools:

```
┌─────────────┐
│   Client    │
└─────────────┘
       │
       ▼
┌─────────────┐
│   Server    │
└─────────────┘
       │
       ▼
┌─────────────┐
│ToolProvider │
└─────────────┘
       │
       ▼
┌─────────────┐
│    Tool     │
└─────────────┘
```

- **Client**: Makes requests to call tools
- **Server**: Handles incoming requests and routes them to tool providers
- **ToolProvider**: Manages registration and lookup of available tools
- **Tool**: Implements the actual functionality

Each layer communicates using the MCP protocol, which defines the format of requests and responses.

## Creating a Basic Tool

Let's start by creating a simple "hello world" tool that echoes a message back to the caller.

### Step 1: Define the Tool

```go
package mytools

import (
	"context"
	"encoding/json"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
)

func RegisterHelloTool(registry *tool_registry.Registry) error {
	// Define the JSON schema for input parameters
	schemaJson := `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "Name to greet"
			}
		},
		"required": ["name"]
	}`

	// Create a new tool implementation
	tool, err := tools.NewToolImpl(
		"hello",                // Tool name
		"Greet a person by name",  // Tool description
		json.RawMessage(schemaJson), // Input schema
	)
	if err != nil {
		return err
	}

	// Register the tool with its handler
	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Extract the name parameter
			name, ok := arguments["name"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("name argument must be a string"),
				), nil
			}

			// Generate and return the response
			message := "Hello, " + name + "!"
			return protocol.NewToolResult(
				protocol.WithText(message),
			), nil
		})

	return nil
}
```

### Step 2: Register the Tool

In your main application, you need to register the tool with the MCP system:

```go
package main

import (
	"github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"mypackage/mytools"
)

func main() {
	// Create a new tool registry
	registry := tool_registry.NewRegistry()

	// Register your custom tool
	err := mytools.RegisterHelloTool(registry)
	if err != nil {
		panic(err)
	}

	// Use the registry as a tool provider for your MCP server
	// ...
}
```

## Tool Registration

The tool registry is a central component that manages the registration and lookup of tools. It implements the `pkg.ToolProvider` interface, which allows the MCP server to discover and call tools.

### Registry Methods

- **RegisterTool**: Registers a tool that implements the `tools.Tool` interface
- **RegisterToolWithHandler**: Registers a tool with a custom handler function
- **UnregisterTool**: Removes a tool from the registry
- **ListTools**: Lists all registered tools
- **CallTool**: Invokes a specific tool by name

### Combining Tool Providers

You can combine multiple tool providers using the `tools.CombineProviders` function:

```go
package main

import (
	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
)

func createCombinedProvider() pkg.ToolProvider {
	// Create individual providers
	provider1 := createProvider1()
	provider2 := createProvider2()

	// Combine them into a single provider
	return tools.CombineProviders(provider1, provider2)
}
```

## Input Validation with JSON Schema

Tools use JSON Schema to define and validate their input parameters. This provides a structured way to describe the data format and constraints expected by your tool.

### Basic Schema Elements

```json
{
  "type": "object",
  "properties": {
    "stringParam": {
      "type": "string",
      "description": "A string parameter"
    },
    "numberParam": {
      "type": "number",
      "description": "A number parameter"
    },
    "boolParam": {
      "type": "boolean",
      "description": "A boolean parameter"
    }
  },
  "required": ["stringParam"]
}
```

### Schema Types

- **string**: Text values
- **number**: Numeric values (integer or floating point)
- **integer**: Integer values only
- **boolean**: True/false values
- **array**: Lists of values
- **object**: Nested objects with their own properties

### Validation Rules

You can add additional validation rules to your schema:

```json
{
  "type": "object",
  "properties": {
    "username": {
      "type": "string",
      "minLength": 3,
      "maxLength": 20,
      "pattern": "^[a-zA-Z0-9_]+$"
    },
    "age": {
      "type": "integer",
      "minimum": 18,
      "maximum": 120
    },
    "tags": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "minItems": 1,
      "maxItems": 5
    }
  }
}
```

## Returning Results

Tools return their results using the `protocol.ToolResult` structure, which can contain different types of content.

### Creating Results

```go
result := protocol.NewToolResult(
	protocol.WithText("This is a text response"),
)
```

### Result Types

- **Text**: Plain text content
- **JSON**: Structured data in JSON format
- **Image**: Base64-encoded image data
- **Resource**: References to external resources

### Returning Errors

If your tool encounters an error, you can return it using the `WithError` option:

```go
return protocol.NewToolResult(
	protocol.WithError("An error occurred: " + err.Error()),
), nil
```

### Multiple Content Types

You can combine multiple content types in a single result:

```go
return protocol.NewToolResult(
	protocol.WithText("Operation completed successfully."),
	protocol.WithJSON(data),
), nil
```

## Advanced Tool Development

Let's explore more complex tool development patterns using the SQLite tool example from the go-go-mcp codebase.

### Stateful Tools

Some tools need to maintain state between invocations. The SQLite tool provides an example of how to create a stateful tool that keeps a database connection open across multiple calls.

```go
// Example of a stateful tool setup
func registerSQLiteOpenTool(registry *tool_registry.Registry) error {
	// ... schema definition ...

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Get the session from context
			s, ok := session.GetSessionFromContext(ctx)
			if !ok {
				return protocol.NewToolResult(protocol.WithError("no session found in context")), nil
			}

			// Store state in the session
			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database '%s': %v", dbPath, err)),
				), nil
			}

			// Store connection in session for later use
			s.SetData(sessionDBConnectionKey, db)
			s.SetData(sessionDBPathKey, dbPath)

			return protocol.NewToolResult(
				protocol.WithText(fmt.Sprintf("Successfully opened database: %s", dbPath)),
			), nil
		})

	return nil
}
```

### Tool Dependencies

Tools can depend on each other, where one tool's output becomes another tool's input. For example, the SQLite query tool depends on the SQLite open tool to establish a database connection.

```go
func registerSQLiteQueryTool(registry *tool_registry.Registry) error {
	// ... schema definition ...

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Get the session from context
			s, sessionOk := session.GetSessionFromContext(ctx)
			if sessionOk {
				// Look for existing database connection in session
				if connVal, connOk := s.GetData(sessionDBConnectionKey); connOk {
					db, dbOk := connVal.(*sql.DB)
					if dbOk {
						// Use the existing connection
						// ...
					}
				}
			}
			
			// ... tool implementation ...
		})

	return nil
}
```

## Session Management

The session system in go-go-mcp allows tools to store and retrieve state across multiple invocations within the same conversation.

### Accessing the Session

```go
import "github.com/go-go-golems/go-go-mcp/pkg/session"

func toolHandler(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	// Get the session from the context
	s, ok := session.GetSessionFromContext(ctx)
	if !ok {
		return protocol.NewToolResult(protocol.WithError("no session found in context")), nil
	}

	// ... tool implementation ...
}
```

### Managing Session Data

- **SetData**: Store data in the session
- **GetData**: Retrieve data from the session
- **DeleteData**: Remove data from the session

```go
// Store data
s.SetData("key", value)

// Retrieve data
value, ok := s.GetData("key")
if ok {
	// Use the value
}

// Delete data
s.DeleteData("key")
```

### Session Cleanup

When your tool allocates resources that need to be cleaned up, make sure to provide a way to release them:

```go
// Example cleanup function
func closeSessionDB(s *session.Session) (string, bool) {
	connVal, ok := s.GetData(sessionDBConnectionKey)
	if !ok {
		return "", false // No connection to close
	}

	db, ok := connVal.(*sql.DB)
	if !ok {
		// Remove potentially corrupted data
		s.DeleteData(sessionDBConnectionKey)
		s.DeleteData(sessionDBPathKey)
		return "", false
	}

	err := db.Close()
	s.DeleteData(sessionDBConnectionKey)
	s.DeleteData(sessionDBPathKey)

	// ... logging and error handling ...
	
	return dbPath, true
}
```

## Real-World Examples

Let's look at a couple of real-world example tools to better understand how to implement different types of functionality.

### Example 1: File System Tool

This tool allows listing files in a directory:

```go
func RegisterListDirTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Directory path to list"
			}
		},
		"required": ["path"]
	}`

	tool, err := tools.NewToolImpl(
		"list_directory",
		"List files in a directory",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			path, ok := arguments["path"].(string)
			if !ok {
				return protocol.NewToolResult(protocol.WithError("path argument must be a string")), nil
			}

			files, err := os.ReadDir(path)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error listing directory '%s': %v", path, err)),
				), nil
			}

			var fileInfos []map[string]interface{}
			for _, file := range files {
				info, err := file.Info()
				if err != nil {
					continue
				}
				
				fileInfos = append(fileInfos, map[string]interface{}{
					"name":  file.Name(),
					"size":  info.Size(),
					"isDir": file.IsDir(),
					"mode":  info.Mode().String(),
				})
			}

			return protocol.NewToolResult(
				protocol.WithJSON(fileInfos),
			), nil
		})

	return nil
}
```

### Example 2: API Integration Tool

This tool fetches data from an external API:

```go
func RegisterWeatherTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"city": {
				"type": "string",
				"description": "City name"
			},
			"units": {
				"type": "string",
				"enum": ["metric", "imperial"],
				"default": "metric",
				"description": "Units of measurement"
			}
		},
		"required": ["city"]
	}`

	tool, err := tools.NewToolImpl(
		"get_weather",
		"Get current weather for a city",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			city, ok := arguments["city"].(string)
			if !ok {
				return protocol.NewToolResult(protocol.WithError("city argument must be a string")), nil
			}

			units := "metric"
			if unitsArg, ok := arguments["units"].(string); ok {
				units = unitsArg
			}

			// Call external API (implementation omitted for brevity)
			weatherData, err := fetchWeatherData(ctx, city, units)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error fetching weather data: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithJSON(weatherData),
			), nil
		})

	return nil
}
```

## Best Practices

### 1. Input Validation

Always validate input parameters before using them:

```go
value, ok := arguments["param"].(string)
if !ok {
	return protocol.NewToolResult(protocol.WithError("param must be a string")), nil
}
```

### 2. Error Handling

Provide clear error messages that help users understand what went wrong:

```go
if err != nil {
	return protocol.NewToolResult(
		protocol.WithError(fmt.Sprintf("failed to process data: %v", err)),
	), nil
}
```

### 3. Resource Cleanup

Ensure that any resources (files, network connections, etc.) are properly cleaned up:

```go
file, err := os.Open(path)
if err != nil {
	return protocol.NewToolResult(protocol.WithError(err.Error())), nil
}
defer file.Close()
```

### 4. Context Awareness

Respect the context passed to your tool, which may include cancellation signals:

```go
select {
case <-ctx.Done():
	return protocol.NewToolResult(protocol.WithError("operation cancelled")), ctx.Err()
case result := <-resultChan:
	return protocol.NewToolResult(protocol.WithText(result)), nil
}
```

### 5. Logging

Include appropriate logging to help with debugging:

```go
import "github.com/rs/zerolog/log"

func toolHandler(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	log.Debug().
		Str("tool", tool.GetName()).
		Interface("arguments", arguments).
		Msg("Tool invoked")
		
	// ... tool implementation ...
}
```

For more information, see the other guides in this documentation:
- [Configuration File Guide](01-config-file.md)
- [Shell Commands Guide](02-shell-commands.md)
- [MCP in Practice](03-mcp-in-practice.md) 