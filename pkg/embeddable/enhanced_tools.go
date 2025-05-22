package embeddable

import (
	"context"
	"encoding/json"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
)

// EnhancedToolHandler is an enhanced function signature for tool handlers
// It provides an Arguments wrapper with convenient accessor methods
type EnhancedToolHandler func(ctx context.Context, args Arguments) (*protocol.ToolResult, error)

// PropertyOption configures a property in a tool's input schema
type PropertyOption func(map[string]interface{})

// ToolAnnotations provides metadata about tool behavior
type ToolAnnotations struct {
	Title           string `json:"title,omitempty"`
	ReadOnlyHint    *bool  `json:"readOnlyHint,omitempty"`
	DestructiveHint *bool  `json:"destructiveHint,omitempty"`
	IdempotentHint  *bool  `json:"idempotentHint,omitempty"`
	OpenWorldHint   *bool  `json:"openWorldHint,omitempty"`
}

// Enhanced tool registration with mark3labs/mcp-go inspired API
func WithEnhancedTool(name string, handler EnhancedToolHandler, opts ...EnhancedToolOption) ServerOption {
	return func(config *ServerConfig) error {
		// Create enhanced tool configuration
		toolConfig := &EnhancedToolConfig{
			Description: "",
			Schema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Annotations: ToolAnnotations{
				ReadOnlyHint:    boolPtr(false),
				DestructiveHint: boolPtr(true),
				IdempotentHint:  boolPtr(false),
				OpenWorldHint:   boolPtr(true),
			},
		}

		// Apply enhanced tool options
		for _, opt := range opts {
			if err := opt(toolConfig); err != nil {
				return err
			}
		}

		// Create the tool
		tool, err := createEnhancedTool(name, toolConfig)
		if err != nil {
			return err
		}

		// Register the tool with enhanced handler
		config.toolRegistry.RegisterToolWithHandler(tool, func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Wrap arguments with enhanced accessor
			args := NewArguments(arguments)

			// Apply middleware
			finalHandler := func(ctx context.Context, args Arguments) (*protocol.ToolResult, error) {
				return handler(ctx, args)
			}

			for i := len(config.middleware) - 1; i >= 0; i-- {
				// Convert middleware to work with enhanced handler
				middleware := config.middleware[i]
				currentHandler := finalHandler
				finalHandler = func(ctx context.Context, args Arguments) (*protocol.ToolResult, error) {
					legacyHandler := func(ctx context.Context, legacyArgs map[string]interface{}) (*protocol.ToolResult, error) {
						return currentHandler(ctx, NewArguments(legacyArgs))
					}
					wrappedHandler := middleware(legacyHandler)
					return wrappedHandler(ctx, args.Raw())
				}
			}

			// Apply hooks
			if config.hooks != nil && config.hooks.BeforeToolCall != nil {
				if err := config.hooks.BeforeToolCall(ctx, name, arguments); err != nil {
					return nil, err
				}
			}

			// Call the enhanced handler
			result, err := finalHandler(ctx, args)

			// Apply hooks
			if config.hooks != nil && config.hooks.AfterToolCall != nil {
				config.hooks.AfterToolCall(ctx, name, result, err)
			}

			return result, err
		})

		return nil
	}
}

// EnhancedToolConfig holds configuration for enhanced tools
type EnhancedToolConfig struct {
	Description string
	Schema      map[string]interface{}
	Annotations ToolAnnotations
	Examples    []ToolExample
}

// EnhancedToolOption configures enhanced tools
type EnhancedToolOption func(*EnhancedToolConfig) error

// Enhanced tool configuration options
func WithEnhancedDescription(desc string) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		config.Description = desc
		return nil
	}
}

func WithAnnotations(annotations ToolAnnotations) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		config.Annotations = annotations
		return nil
	}
}

func WithReadOnlyHint(readOnly bool) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		config.Annotations.ReadOnlyHint = boolPtr(readOnly)
		return nil
	}
}

func WithDestructiveHint(destructive bool) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		config.Annotations.DestructiveHint = boolPtr(destructive)
		return nil
	}
}

func WithIdempotentHint(idempotent bool) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		config.Annotations.IdempotentHint = boolPtr(idempotent)
		return nil
	}
}

func WithOpenWorldHint(openWorld bool) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		config.Annotations.OpenWorldHint = boolPtr(openWorld)
		return nil
	}
}

// Enhanced property configuration
func WithStringProperty(name string, opts ...PropertyOption) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		return addProperty(config, name, "string", opts...)
	}
}

func WithIntProperty(name string, opts ...PropertyOption) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		return addProperty(config, name, "integer", opts...)
	}
}

func WithNumberProperty(name string, opts ...PropertyOption) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		return addProperty(config, name, "number", opts...)
	}
}

func WithBooleanProperty(name string, opts ...PropertyOption) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		return addProperty(config, name, "boolean", opts...)
	}
}

func WithArrayProperty(name string, opts ...PropertyOption) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		return addProperty(config, name, "array", opts...)
	}
}

func WithObjectProperty(name string, opts ...PropertyOption) EnhancedToolOption {
	return func(config *EnhancedToolConfig) error {
		return addProperty(config, name, "object", opts...)
	}
}

// Property configuration options
func PropertyDescription(desc string) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["description"] = desc
	}
}

func PropertyRequired() PropertyOption {
	return func(schema map[string]interface{}) {
		schema["required"] = true
	}
}

func PropertyTitle(title string) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["title"] = title
	}
}

// String property options
func DefaultString(value string) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["default"] = value
	}
}

func StringEnum(values ...string) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["enum"] = values
	}
}

func MaxLength(maxLen int) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["maxLength"] = maxLen
	}
}

func MinLength(minLen int) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["minLength"] = minLen
	}
}

func StringPattern(pattern string) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["pattern"] = pattern
	}
}

// Number property options
func DefaultNumber(value float64) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["default"] = value
	}
}

func Maximum(maxVal float64) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["maximum"] = maxVal
	}
}

func Minimum(minVal float64) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["minimum"] = minVal
	}
}

func MultipleOf(value float64) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["multipleOf"] = value
	}
}

// Boolean property options
func DefaultBool(value bool) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["default"] = value
	}
}

// Array property options
func ArrayItems(itemSchema map[string]interface{}) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["items"] = itemSchema
	}
}

func MinItems(minItems int) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["minItems"] = minItems
	}
}

func MaxItems(maxItems int) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["maxItems"] = maxItems
	}
}

func UniqueItems(unique bool) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["uniqueItems"] = unique
	}
}

// Helper functions
func addProperty(config *EnhancedToolConfig, name, propType string, opts ...PropertyOption) error {
	schema := map[string]interface{}{
		"type": propType,
	}

	// Apply property options
	for _, opt := range opts {
		opt(schema)
	}

	// Handle required properties
	if required, ok := schema["required"].(bool); ok && required {
		delete(schema, "required")

		// Ensure required array exists in main schema
		requiredArray, ok := config.Schema["required"].([]interface{})
		if !ok {
			requiredArray = []interface{}{}
		}
		requiredArray = append(requiredArray, name)
		config.Schema["required"] = requiredArray
	}

	// Add to properties
	properties := config.Schema["properties"].(map[string]interface{})
	properties[name] = schema

	return nil
}

func createEnhancedTool(name string, config *EnhancedToolConfig) (tools.Tool, error) {
	// Convert schema to JSON
	schemaBytes, err := json.Marshal(config.Schema)
	if err != nil {
		return nil, err
	}

	return tools.NewToolImpl(name, config.Description, json.RawMessage(schemaBytes))
}

func boolPtr(b bool) *bool {
	return &b
}
