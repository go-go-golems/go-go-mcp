package tools

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-go-golems/geppetto/pkg/helpers"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/invopop/jsonschema"
)

// ReflectTool is a Tool implementation that uses reflection to create a tool from a Go function
type ReflectTool struct {
	*ToolImpl
	reflector *jsonschema.Reflector
	fn        interface{}
}

var _ Tool = &ReflectTool{}

// NewReflectTool creates a new ReflectTool from a function
func NewReflectTool(name string, description string, fn interface{}) (*ReflectTool, error) {
	reflector := new(jsonschema.Reflector)

	// Get the JSON schema for the function parameters
	schema, err := helpers.GetSimplifiedFunctionParametersJsonSchema(reflector, fn)
	if err != nil {
		return nil, fmt.Errorf("failed to get function parameters schema: %w", err)
	}

	// Create the base ToolImpl
	toolImpl, err := NewToolImpl(name, description, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool implementation: %w", err)
	}

	return &ReflectTool{
		ToolImpl:  toolImpl,
		reflector: reflector,
		fn:        fn,
	}, nil
}

// isPrimitiveType checks if a value is a primitive type (string, number, bool)
func isPrimitiveType(v interface{}) bool {

	if v == nil {
		return true
	}

	//nolint:exhaustive
	switch reflect.TypeOf(v).Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	default:
		return false
	}
}

// Call implements the Tool interface by using reflection to call the underlying function
func (t *ReflectTool) Call(ctx context.Context, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	// Call the function using reflection
	results, err := helpers.CallFunctionFromJson(t.fn, arguments)
	if err != nil {
		return protocol.NewToolResult(
			protocol.WithError(fmt.Sprintf("failed to call function: %v", err)),
		), nil
	}

	// Convert the results to a JSON-serializable format
	var resultValue interface{}
	if len(results) == 1 {
		resultValue = results[0].Interface()
	} else if len(results) > 1 {
		resultValues := make([]interface{}, len(results))
		for i, r := range results {
			resultValues[i] = r.Interface()
		}
		resultValue = resultValues
	}

	// If the result is a primitive type or string, return it as text content
	if isPrimitiveType(resultValue) {
		return protocol.NewToolResult(
			protocol.WithText(fmt.Sprintf("%v", resultValue)),
		), nil
	}

	// For complex types (maps, structs, slices), use JSON content
	content, err := protocol.NewJSONContent(resultValue)
	if err != nil {
		return protocol.NewToolResult(
			protocol.WithError(fmt.Sprintf("failed to convert result to JSON: %v", err)),
		), nil
	}

	return protocol.NewToolResult(
		protocol.WithContent(content),
	), nil
}
