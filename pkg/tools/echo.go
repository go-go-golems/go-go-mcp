package tools

import (
	"context"
	"encoding/json"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func RegisterEchoTool(registry *Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"message": {
				"type": "string"
			}
		}
	}`

	tool, err := NewToolImpl("echo", "Echo the input arguments", json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			message, ok := arguments["message"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("message argument must be a string"),
				), nil
			}
			return protocol.NewToolResult(
				protocol.WithText(message),
			), nil
		})

	return nil
}
