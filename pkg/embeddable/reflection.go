package embeddable

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
)

// RegisterStructTool automatically creates a tool from a struct and method
func RegisterStructTool(config *ServerConfig, name string, obj interface{}, methodName string) error {
	objValue := reflect.ValueOf(obj)
	objType := reflect.TypeOf(obj)

	// Find the method
	method := objValue.MethodByName(methodName)
	if !method.IsValid() {
		return fmt.Errorf("method %s not found on %T", methodName, obj)
	}

	methodType := method.Type()

	// Validate method signature
	if methodType.NumIn() < 1 {
		return fmt.Errorf("method %s must have at least context parameter", methodName)
	}

	// First parameter should be context.Context
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !methodType.In(0).Implements(ctxType) {
		return fmt.Errorf("method %s first parameter must be context.Context", methodName)
	}

	// Method should return (*protocol.ToolResult, error)
	if methodType.NumOut() != 2 {
		return fmt.Errorf("method %s must return (*protocol.ToolResult, error)", methodName)
	}

	toolResultType := reflect.TypeOf((*protocol.ToolResult)(nil))
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	if methodType.Out(0) != toolResultType {
		return fmt.Errorf("method %s first return value must be *protocol.ToolResult", methodName)
	}

	if !methodType.Out(1).Implements(errorType) {
		return fmt.Errorf("method %s second return value must implement error", methodName)
	}

	// Generate description from struct and method
	description := fmt.Sprintf("%s.%s", objType.Name(), methodName)

	// Generate schema based on method signature
	var schema map[string]interface{}
	var err error

	if methodType.NumIn() == 2 {
		// Second parameter should be the arguments struct
		argsType := methodType.In(1)
		schema, err = generateSchemaFromType(argsType)
		if err != nil {
			return fmt.Errorf("failed to generate schema for %s: %w", methodName, err)
		}
	} else {
		// No arguments
		schema = map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	// Create the tool
	tool, err := tools.NewToolImpl(name, description, schema)
	if err != nil {
		return fmt.Errorf("failed to create tool %s: %w", name, err)
	}

	// Create wrapper handler
	handler := func(ctx context.Context, arguments map[string]interface{}) (*protocol.ToolResult, error) {
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))

		if methodType.NumIn() == 2 {
			// Convert arguments to struct
			argsType := methodType.In(1)
			argsValue, err := convertArgumentsToStruct(arguments, argsType)
			if err != nil {
				return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Invalid arguments: %v", err))), nil
			}
			args = append(args, argsValue)
		}

		// Call the method
		results := method.Call(args)

		// Extract results
		resultValue := results[0]
		errorValue := results[1]

		var err error
		if !errorValue.IsNil() {
			err = errorValue.Interface().(error)
		}

		var result *protocol.ToolResult
		if !resultValue.IsNil() {
			result = resultValue.Interface().(*protocol.ToolResult)
		}

		return result, err
	}

	// Register the tool
	config.toolRegistry.RegisterToolWithHandler(tool, func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
		return handler(ctx, arguments)
	})

	return nil
}

// RegisterFunctionTool creates a tool from a function using reflection
func RegisterFunctionTool(config *ServerConfig, name string, fn interface{}) error {
	fnValue := reflect.ValueOf(fn)
	fnType := reflect.TypeOf(fn)

	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("expected function, got %T", fn)
	}

	// Validate function signature
	if fnType.NumIn() < 1 {
		return fmt.Errorf("function must have at least context parameter")
	}

	// First parameter should be context.Context
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !fnType.In(0).Implements(ctxType) {
		return fmt.Errorf("function first parameter must be context.Context")
	}

	// Function should return (*protocol.ToolResult, error)
	if fnType.NumOut() != 2 {
		return fmt.Errorf("function must return (*protocol.ToolResult, error)")
	}

	toolResultType := reflect.TypeOf((*protocol.ToolResult)(nil))
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	if fnType.Out(0) != toolResultType {
		return fmt.Errorf("function first return value must be *protocol.ToolResult")
	}

	if !fnType.Out(1).Implements(errorType) {
		return fmt.Errorf("function second return value must implement error")
	}

	// Generate schema based on function signature
	var schema map[string]interface{}
	var err error

	if fnType.NumIn() == 2 {
		// Second parameter should be the arguments struct
		argsType := fnType.In(1)
		schema, err = generateSchemaFromType(argsType)
		if err != nil {
			return fmt.Errorf("failed to generate schema: %w", err)
		}
	} else {
		// No arguments
		schema = map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	// Create the tool
	tool, err := tools.NewToolImpl(name, fmt.Sprintf("Generated tool for function %s", name), schema)
	if err != nil {
		return fmt.Errorf("failed to create tool %s: %w", name, err)
	}

	// Create wrapper handler
	handler := func(ctx context.Context, arguments map[string]interface{}) (*protocol.ToolResult, error) {
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))

		if fnType.NumIn() == 2 {
			// Convert arguments to struct
			argsType := fnType.In(1)
			argsValue, err := convertArgumentsToStruct(arguments, argsType)
			if err != nil {
				return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Invalid arguments: %v", err))), nil
			}
			args = append(args, argsValue)
		}

		// Call the function
		results := fnValue.Call(args)

		// Extract results
		resultValue := results[0]
		errorValue := results[1]

		var err error
		if !errorValue.IsNil() {
			err = errorValue.Interface().(error)
		}

		var result *protocol.ToolResult
		if !resultValue.IsNil() {
			result = resultValue.Interface().(*protocol.ToolResult)
		}

		return result, err
	}

	// Register the tool
	config.toolRegistry.RegisterToolWithHandler(tool, func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
		return handler(ctx, arguments)
	})

	return nil
}

// generateSchemaFromType generates a JSON schema from a Go type
func generateSchemaFromType(t reflect.Type) (map[string]interface{}, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", t.Kind())
	}

	properties := make(map[string]interface{})
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
			// Check for omitempty
			isOptional := false
			for _, part := range parts[1:] {
				if part == "omitempty" {
					isOptional = true
					break
				}
			}
			if !isOptional {
				required = append(required, fieldName)
			}
		} else {
			// Field without json tag is required by default
			required = append(required, fieldName)
		}

		// Get description from tag
		description := field.Tag.Get("description")

		// Convert Go type to JSON schema type
		fieldSchema := make(map[string]interface{})
		switch field.Type.Kind() {
		case reflect.String:
			fieldSchema["type"] = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldSchema["type"] = "integer"
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldSchema["type"] = "integer"
		case reflect.Float32, reflect.Float64:
			fieldSchema["type"] = "number"
		case reflect.Bool:
			fieldSchema["type"] = "boolean"
		case reflect.Slice, reflect.Array:
			fieldSchema["type"] = "array"
		case reflect.Map, reflect.Struct:
			fieldSchema["type"] = "object"
		case reflect.Invalid:
			fieldSchema["type"] = "null"
		case reflect.Uintptr:
			fieldSchema["type"] = "string" // Represent as string
		case reflect.Complex64, reflect.Complex128:
			fieldSchema["type"] = "string" // Represent complex numbers as strings
		case reflect.Chan:
			fieldSchema["type"] = "string" // Channels not supported, use string representation
		case reflect.Func:
			fieldSchema["type"] = "string" // Functions not supported, use string representation
		case reflect.Interface:
			fieldSchema["type"] = "object" // Interfaces as objects
		case reflect.Ptr:
			fieldSchema["type"] = "object" // Pointers as objects
		case reflect.UnsafePointer:
			fieldSchema["type"] = "string" // Unsafe pointers as strings
		default:
			fieldSchema["type"] = "string" // Default fallback
		}

		if description != "" {
			fieldSchema["description"] = description
		}

		properties[fieldName] = fieldSchema
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema, nil
}

// convertArgumentsToStruct converts a map of arguments to a struct
func convertArgumentsToStruct(arguments map[string]interface{}, structType reflect.Type) (reflect.Value, error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	// Create new struct instance (we'll use JSON marshaling instead)

	// Convert arguments to JSON and back to struct for proper type conversion
	jsonBytes, err := json.Marshal(arguments)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	structPtr := reflect.New(structType).Interface()
	err = json.Unmarshal(jsonBytes, structPtr)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	return reflect.ValueOf(structPtr).Elem(), nil
}
