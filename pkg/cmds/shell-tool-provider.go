package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ShellTool is a wrapper around a shell command that implements the Tool interface
type ShellTool struct {
	cmd *ShellCommand
}

// ShellToolProvider is a ToolProvider that exposes shell commands as tools
type ShellToolProvider struct {
	commands   map[string]*ShellTool
	debug      bool
	tracingDir string
}

type ShellToolProviderOption func(*ShellToolProvider)

func WithDebug(debug bool) ShellToolProviderOption {
	return func(p *ShellToolProvider) {
		p.debug = debug
	}
}

func WithTracingDir(dir string) ShellToolProviderOption {
	return func(p *ShellToolProvider) {
		p.tracingDir = dir
	}
}

// NewShellToolProvider creates a new ShellToolProvider with the given commands
func NewShellToolProvider(commands []cmds.Command, options ...ShellToolProviderOption) (*ShellToolProvider, error) {
	provider := &ShellToolProvider{
		commands: make(map[string]*ShellTool),
	}

	for _, option := range options {
		option(provider)
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
			Description: desc.Short + "\n\n" + desc.Long,
			InputSchema: schemaBytes,
		})
	}

	return tools, "", nil
}

// CallTool invokes a specific tool with the given arguments
func (p *ShellToolProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	if p.debug {
		log.Debug().
			Str("name", name).
			Interface("arguments", arguments).
			Msg("calling tool with arguments")
	}

	tool, ok := p.commands[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	if p.tracingDir != "" {
		timestamp := time.Now().Format("2006-01-02T15-04-05.000")
		inputFile := filepath.Join(p.tracingDir, fmt.Sprintf("%s-%s-input.json", name, timestamp))
		if err := p.writeTraceFile(inputFile, arguments); err != nil {
			log.Error().Err(err).Str("file", inputFile).Msg("failed to write input trace file")
		}
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

	if p.debug {
		log.Debug().Interface("dataMap", dataMap).Msg("processed arguments")
	}

	// Run the command with parsed parameters
	err = tool.cmd.ExecuteCommand(ctx, dataMap, buf)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	text := buf.String()
	l := 62 * 1024
	if len(text) > l {
		text = text[:l]
	}

	result := protocol.NewToolResult(protocol.WithText(text))

	if p.tracingDir != "" {
		timestamp := time.Now().Format("2006-01-02T15-04-05.000")
		outputFile := filepath.Join(p.tracingDir, fmt.Sprintf("%s-%s-output.json", name, timestamp))
		if err := p.writeTraceFile(outputFile, result); err != nil {
			log.Error().Err(err).Str("file", outputFile).Msg("failed to write output trace file")
		}
	}

	return result, nil
}

func (p *ShellToolProvider) writeTraceFile(filename string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create tracing directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write trace file: %w", err)
	}

	return nil
}
