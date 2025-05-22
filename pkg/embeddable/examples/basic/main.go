package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/spf13/cobra"
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
		embeddable.WithServerDescription("Example MCP server for demonstration"),
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
