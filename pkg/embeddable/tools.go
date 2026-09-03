package embeddable

import (
	"encoding/json"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// ToolOption configures individual tools
type ToolOption func(*ToolConfig) error

// ToolConfig holds configuration for a tool
type ToolConfig struct {
	Description  string
	Schema       interface{} // Can be a struct, JSON schema string, or json.RawMessage
	OutputSchema interface{} // Optional output schema (struct, JSON schema string, or json.RawMessage)
	Annotations  *protocol.ToolAnnotations
	Examples     []ToolExample
}

// ToolExample represents an example usage of a tool
type ToolExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Arguments   map[string]interface{} `json:"arguments"`
}

// Tool configuration options
func WithDescription(desc string) ToolOption {
	return func(config *ToolConfig) error {
		config.Description = desc
		return nil
	}
}

func WithSchema(schema interface{}) ToolOption {
	return func(config *ToolConfig) error {
		config.Schema = schema
		return nil
	}
}

// WithOutputSchema attaches an output schema to a simple tool. The schema
// can be a struct, a JSON schema string, or a json.RawMessage.
func WithOutputSchema(schema interface{}) ToolOption {
	return func(config *ToolConfig) error {
		config.OutputSchema = schema
		return nil
	}
}

// WithToolAnnotations attaches MCP behavioral annotations to a simple tool.
// Annotations are descriptive hints for clients and are never enforced.
func WithToolAnnotations(annotations protocol.ToolAnnotations) ToolOption {
	return func(config *ToolConfig) error {
		annotationsCopy := annotations
		config.Annotations = &annotationsCopy
		return nil
	}
}

func WithExample(name, description string, args map[string]interface{}) ToolOption {
	return func(config *ToolConfig) error {
		config.Examples = append(config.Examples, ToolExample{
			Name:        name,
			Description: description,
			Arguments:   args,
		})
		return nil
	}
}

// Convenience functions for common tool patterns
func WithStringArg(name, description string, required bool) ToolOption {
	return func(config *ToolConfig) error {
		schema := ensureObjectSchema(config.Schema)
		properties := schema["properties"].(map[string]interface{})

		properties[name] = map[string]interface{}{
			"type":        "string",
			"description": description,
		}

		if required {
			requiredList, ok := schema["required"].([]interface{})
			if !ok {
				requiredList = []interface{}{}
			}
			requiredList = append(requiredList, name)
			schema["required"] = requiredList
		}

		config.Schema = schema
		return nil
	}
}

func WithIntArg(name, description string, required bool) ToolOption {
	return func(config *ToolConfig) error {
		schema := ensureObjectSchema(config.Schema)
		properties := schema["properties"].(map[string]interface{})

		properties[name] = map[string]interface{}{
			"type":        "integer",
			"description": description,
		}

		if required {
			requiredList, ok := schema["required"].([]interface{})
			if !ok {
				requiredList = []interface{}{}
			}
			requiredList = append(requiredList, name)
			schema["required"] = requiredList
		}

		config.Schema = schema
		return nil
	}
}

func WithBoolArg(name, description string, required bool) ToolOption {
	return func(config *ToolConfig) error {
		schema := ensureObjectSchema(config.Schema)
		properties := schema["properties"].(map[string]interface{})

		properties[name] = map[string]interface{}{
			"type":        "boolean",
			"description": description,
		}

		if required {
			requiredList, ok := schema["required"].([]interface{})
			if !ok {
				requiredList = []interface{}{}
			}
			requiredList = append(requiredList, name)
			schema["required"] = requiredList
		}

		config.Schema = schema
		return nil
	}
}

func WithFileArg(name, description string, required bool) ToolOption {
	return func(config *ToolConfig) error {
		schema := ensureObjectSchema(config.Schema)
		properties := schema["properties"].(map[string]interface{})

		properties[name] = map[string]interface{}{
			"type":        "string",
			"description": description,
			"format":      "file",
		}

		if required {
			requiredList, ok := schema["required"].([]interface{})
			if !ok {
				requiredList = []interface{}{}
			}
			requiredList = append(requiredList, name)
			schema["required"] = requiredList
		}

		config.Schema = schema
		return nil
	}
}

// ensureObjectSchema ensures that the schema is a proper object schema
func ensureObjectSchema(schema interface{}) map[string]interface{} {
	switch s := schema.(type) {
	case map[string]interface{}:
		if s["type"] == nil {
			s["type"] = "object"
		}
		if s["properties"] == nil {
			s["properties"] = make(map[string]interface{})
		}
		return s
	case string:
		// Try to parse JSON string
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(s), &parsed); err == nil {
			if parsed["type"] == nil {
				parsed["type"] = "object"
			}
			if parsed["properties"] == nil {
				parsed["properties"] = make(map[string]interface{})
			}
			return parsed
		}
		// Create a new object schema if parsing fails
		return map[string]interface{}{
			"type":       "object",
			"properties": make(map[string]interface{}),
		}
	default:
		// Create a new object schema
		return map[string]interface{}{
			"type":       "object",
			"properties": make(map[string]interface{}),
		}
	}
}

// RegisterSimpleTools provides a very easy way to register multiple tools
func RegisterSimpleTools(config *ServerConfig, tools map[string]ToolHandler) error {
	for name, handler := range tools {
		err := WithTool(name, handler)(config)
		if err != nil {
			return err
		}
	}
	return nil
}
