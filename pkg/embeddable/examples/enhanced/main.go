package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "enhanced-app",
		Short: "Application demonstrating enhanced embeddable MCP features",
	}

	// Add MCP server capability with enhanced tools
	err := embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("Enhanced MCP Server"),
		embeddable.WithVersion("1.0.0"),
		embeddable.WithServerDescription("Demonstration of enhanced embeddable MCP features"),

		// Enhanced tool with rich argument handling
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
				embeddable.StringEnum("uppercase", "lowercase", "title", "reverse"),
				embeddable.DefaultString("lowercase"),
			),
			embeddable.WithBooleanProperty("trim",
				embeddable.PropertyDescription("Whether to trim whitespace"),
				embeddable.DefaultBool(true),
			),
			embeddable.WithIntProperty("max_length",
				embeddable.PropertyDescription("Maximum length of output"),
				embeddable.Minimum(1),
				embeddable.Maximum(1000),
				embeddable.DefaultNumber(100),
			),
		),

		// Mathematical operations tool
		embeddable.WithEnhancedTool("math_operation", mathOperationHandler,
			embeddable.WithEnhancedDescription("Perform mathematical operations on arrays"),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithStringProperty("operation",
				embeddable.PropertyDescription("Operation to perform"),
				embeddable.StringEnum("sum", "product", "average", "max", "min"),
				embeddable.PropertyRequired(),
			),
			embeddable.WithArrayProperty("numbers",
				embeddable.PropertyDescription("Array of numbers to operate on"),
				embeddable.PropertyRequired(),
				embeddable.ArrayItems(map[string]interface{}{
					"type": "number",
				}),
				embeddable.MinItems(1),
			),
		),

		// Configuration tool with object properties
		embeddable.WithEnhancedTool("configure_service", configureServiceHandler,
			embeddable.WithEnhancedDescription("Configure a service with complex settings"),
			embeddable.WithDestructiveHint(true),
			embeddable.WithStringProperty("service_name",
				embeddable.PropertyDescription("Name of the service to configure"),
				embeddable.PropertyRequired(),
				embeddable.StringPattern("^[a-zA-Z][a-zA-Z0-9_-]*$"),
			),
			embeddable.WithObjectProperty("config",
				embeddable.PropertyDescription("Service configuration object"),
				embeddable.PropertyRequired(),
			),
			embeddable.WithArrayProperty("tags",
				embeddable.PropertyDescription("Service tags"),
				embeddable.ArrayItems(map[string]interface{}{
					"type": "string",
				}),
				embeddable.UniqueItems(true),
			),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func formatTextHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	// Use the enhanced argument accessors
	text, err := args.RequireString("text")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	format := args.GetString("format", "lowercase")
	trim := args.GetBool("trim", true)
	maxLength := args.GetInt("max_length", 100)

	// Apply formatting
	if trim {
		text = strings.TrimSpace(text)
	}

	switch format {
	case "uppercase":
		text = strings.ToUpper(text)
	case "lowercase":
		text = strings.ToLower(text)
	case "title":
		caser := cases.Title(language.English)
		text = caser.String(text)
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		text = string(runes)
	}

	// Apply length limit
	if len(text) > maxLength {
		text = text[:maxLength] + "..."
	}

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Formatted text: %s", text)),
	), nil
}

func mathOperationHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	operation, err := args.RequireString("operation")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	// Get numbers array - we need to handle the interface{} slice
	numbersRaw, ok := args.Raw()["numbers"]
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("numbers array is required")), nil
	}

	numbersSlice, ok := numbersRaw.([]interface{})
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("numbers must be an array")), nil
	}

	// Convert to float64 slice
	var numbers []float64
	for i, val := range numbersSlice {
		switch v := val.(type) {
		case float64:
			numbers = append(numbers, v)
		case int:
			numbers = append(numbers, float64(v))
		default:
			return protocol.NewErrorToolResult(
				protocol.NewTextContent(fmt.Sprintf("item %d is not a number", i)),
			), nil
		}
	}

	if len(numbers) == 0 {
		return protocol.NewErrorToolResult(protocol.NewTextContent("numbers array cannot be empty")), nil
	}

	var result float64
	switch operation {
	case "sum":
		for _, n := range numbers {
			result += n
		}
	case "product":
		result = 1
		for _, n := range numbers {
			result *= n
		}
	case "average":
		for _, n := range numbers {
			result += n
		}
		result /= float64(len(numbers))
	case "max":
		result = numbers[0]
		for _, n := range numbers {
			if n > result {
				result = n
			}
		}
	case "min":
		result = numbers[0]
		for _, n := range numbers {
			if n < result {
				result = n
			}
		}
	default:
		return protocol.NewErrorToolResult(
			protocol.NewTextContent(fmt.Sprintf("unknown operation: %s", operation)),
		), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Result of %s operation: %g", operation, result)),
	), nil
}

func configureServiceHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	serviceName, err := args.RequireString("service_name")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	// Get config object
	configRaw, ok := args.Raw()["config"]
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("config object is required")), nil
	}

	// Get tags array (optional)
	tags := args.GetStringSlice("tags", []string{})

	// For demonstration, just show what we received
	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf(
			"Service '%s' configured with %d settings and %d tags: %v",
			serviceName, len(configRaw.(map[string]interface{})), len(tags), tags,
		)),
	), nil
}
