package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// Tool represents a tool that can be invoked
type Tool interface {
	GetName() string
	GetDescription() string
	GetInputSchema() json.RawMessage
	GetToolDefinition() protocol.Tool
	Call(ctx context.Context, arguments map[string]interface{}) (*protocol.ToolResult, error)
}

// ToolImpl is a basic implementation of the Tool interface
type ToolImpl struct {
	name        string
	description string
	inputSchema json.RawMessage
}

// NewToolImpl creates a new ToolImpl with the given parameters
func NewToolImpl(name, description string, inputSchema interface{}) (*ToolImpl, error) {
	var schema json.RawMessage
	switch s := inputSchema.(type) {
	case json.RawMessage:
		schema = s
	case string:
		schema = json.RawMessage(s)
	default:
		var err error
		schema, err = json.Marshal(s)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input schema: %w", err)
		}
	}

	return &ToolImpl{
		name:        name,
		description: description,
		inputSchema: schema,
	}, nil
}

// GetName returns the tool's name
func (t *ToolImpl) GetName() string {
	return t.name
}

// GetDescription returns the tool's description
func (t *ToolImpl) GetDescription() string {
	return t.description
}

// GetInputSchema returns the tool's input schema
func (t *ToolImpl) GetInputSchema() json.RawMessage {
	return t.inputSchema
}

// GetToolDefinition returns the tool's definition
func (t *ToolImpl) GetToolDefinition() protocol.Tool {
	return protocol.Tool{
		Name:        t.name,
		Description: t.description,
		InputSchema: t.inputSchema,
	}
}

// Call implements the Tool interface but panics as it should be overridden
func (t *ToolImpl) Call(ctx context.Context, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	panic("Call not implemented for ToolImpl - must be overridden")
}

// MarshalJSON implements json.Marshaler
func (t *ToolImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		InputSchema json.RawMessage `json:"inputSchema"`
	}{
		Name:        t.name,
		Description: t.description,
		InputSchema: t.inputSchema,
	})
}
