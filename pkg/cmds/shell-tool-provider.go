package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/pkg/errors"
)

// ShellTool is a wrapper around a shell command that implements the Tool interface
type ShellTool struct {
	cmd *ShellCommand
}

// ShellToolProvider is a ToolProvider that exposes shell commands as tools
type ShellToolProvider struct {
	commands map[string]*ShellTool
}

// NewShellToolProvider creates a new ShellToolProvider with the given commands
func NewShellToolProvider(commands []cmds.Command) (*ShellToolProvider, error) {
	provider := &ShellToolProvider{
		commands: make(map[string]*ShellTool),
	}

	for _, cmd := range commands {
		if shellCmd, ok := cmd.(*ShellCommand); ok {
			provider.commands[shellCmd.Description().Name] = &ShellTool{cmd: shellCmd}
		}
	}

	return provider, nil
}

// ListTools returns a list of available tools
func (p *ShellToolProvider) ListTools(cursor string) ([]protocol.Tool, string, error) {
	tools := make([]protocol.Tool, 0, len(p.commands))

	for _, tool := range p.commands {
		desc := tool.cmd.Description()
		schema, err := tool.cmd.ToJsonSchema()
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to generate JSON schema")
		}

		schemaBytes, err := json.Marshal(schema)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to marshal JSON schema")
		}

		tools = append(tools, protocol.Tool{
			Name:        desc.Name,
			Description: desc.Short,
			InputSchema: schemaBytes,
		})
	}

	return tools, "", nil
}

// CallTool invokes a specific tool with the given arguments
func (p *ShellToolProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	tool, ok := p.commands[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// Get parameter layers from command description
	parameterLayers := tool.cmd.Description().Layers

	// Create empty parsed layers
	parsedLayers := layers.NewParsedLayers()

	// Create a map structure for the arguments
	argsMap := map[string]map[string]interface{}{
		layers.DefaultSlug: arguments,
	}

	// Execute middlewares in order:
	// 1. Set defaults from parameter definitions
	// 2. Update with provided arguments
	err := middlewares.ExecuteMiddlewares(
		parameterLayers,
		parsedLayers,
		middlewares.SetFromDefaults(parameters.WithParseStepSource(parameters.SourceDefaults)),
		middlewares.UpdateFromMap(argsMap),
	)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	// Create a buffer to capture the command output
	buf := &strings.Builder{}

	dataMap := parsedLayers.GetDataMap()
	// Run the command with parsed parameters
	err = tool.cmd.ExecuteCommand(ctx, dataMap, buf)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	// Return the output as a text result
	return protocol.NewToolResult(protocol.WithText(buf.String())), nil
}
