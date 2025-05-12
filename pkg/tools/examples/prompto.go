package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// RegisterPromptoTools registers tools for interacting with prompto:
// prompto_list: Lists all available prompto configurations
// prompto_get: Gets a specific prompto content by name
func RegisterPromptoTools(registry *tool_registry.Registry) error {
	if err := registerPromptoListTool(registry); err != nil {
		return errors.Wrap(err, "failed to register prompto_list tool")
	}
	if err := registerPromptoGetTool(registry); err != nil {
		return errors.Wrap(err, "failed to register prompto_get tool")
	}
	return nil
}

// registerPromptoListTool registers a tool to list all available prompto configurations
func registerPromptoListTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {},
		"required": []
	}`

	tool, err := tools.NewToolImpl(
		"prompto_list",
		"Lists all available prompto configurations",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, _ tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			log.Debug().Msg("Executing prompto list command")

			cmd := exec.CommandContext(ctx, "prompto", "list")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing 'prompto list': %v\n%s", err, output)),
				), nil
			}

			// Format the output
			result := string(output)
			log.Debug().Str("result", result).Msg("prompto list completed successfully")

			return protocol.NewToolResult(
				protocol.WithText(result),
			), nil
		})

	return nil
}

// registerPromptoGetTool registers a tool to get specific prompto content by name
func registerPromptoGetTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "Name of the prompto configuration to retrieve"
			}
		},
		"required": ["name"]
	}`

	tool, err := tools.NewToolImpl(
		"prompto_get",
		"Gets a specific prompto content by name",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, _ tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			name, ok := arguments["name"].(string)
			if !ok {
				return protocol.NewToolResult(protocol.WithError("name argument must be a string")), nil
			}

			if strings.TrimSpace(name) == "" {
				return protocol.NewToolResult(protocol.WithError("name argument must not be empty")), nil
			}

			log.Debug().Str("name", name).Msg("Executing prompto get command")

			cmd := exec.CommandContext(ctx, "prompto", "get", name)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing 'prompto get %s': %v\n%s", name, err, output)),
				), nil
			}

			// Format the output
			result := string(output)
			log.Debug().Str("name", name).Str("result", result).Msg("prompto get completed successfully")

			return protocol.NewToolResult(
				protocol.WithText(result),
			), nil
		})

	return nil
}

// Ensure Tool interfaces are implemented
var _ tools.Tool = (*tools.ToolImpl)(nil)
