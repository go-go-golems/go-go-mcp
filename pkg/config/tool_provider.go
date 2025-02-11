package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/pkg"
	mcp_cmds "github.com/go-go-golems/go-go-mcp/pkg/cmds"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/parka/pkg/handlers/config"
	"github.com/pkg/errors"
)

// ConfigToolProvider implements pkg.ToolProvider interface
type ConfigToolProvider struct {
	repository    *repositories.Repository
	shellCommands map[string]cmds.Command
	toolConfigs   map[string]*SourceConfig
}

func NewConfigToolProvider(config *Config, profile string) (*ConfigToolProvider, error) {
	if _, ok := config.Profiles[profile]; !ok {
		return nil, errors.Errorf("profile %s not found", profile)
	}

	directories := []repositories.Directory{}
	profileConfig := config.Profiles[profile]

	// Load directories using Clay's repository system
	for _, dir := range profileConfig.Tools.Directories {
		absPath, err := filepath.Abs(dir.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get absolute path for %s", dir.Path)
		}

		directories = append(directories, repositories.Directory{
			FS:            os.DirFS(absPath),
			RootDirectory: ".",
			Name:          dir.Path,
		})
	}

	provider := &ConfigToolProvider{
		repository:    repositories.NewRepository(repositories.WithDirectories(directories...)),
		shellCommands: make(map[string]cmds.Command),
		toolConfigs:   make(map[string]*SourceConfig),
	}

	if profileConfig.Tools == nil {
		return provider, nil
	}

	helpSystem := help.NewHelpSystem()
	// Load repository commands
	if err := provider.repository.LoadCommands(helpSystem); err != nil {
		return nil, errors.Wrap(err, "failed to load repository commands")
	}

	// Load individual files as ShellCommands
	for _, file := range profileConfig.Tools.Files {
		absPath, err := filepath.Abs(file.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get absolute path for %s", file.Path)
		}

		shellCmd, err := mcp_cmds.LoadShellCommand(absPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load shell command from %s", file.Path)
		}

		desc := shellCmd.Description()
		provider.shellCommands[desc.Name] = shellCmd
		provider.toolConfigs[desc.Name] = &file
	}

	return provider, nil
}

func ConvertCommandToTool(desc *cmds.CommandDescription) (protocol.Tool, error) {
	schema_, err := mcp_cmds.ToJsonSchema(desc)
	if err != nil {
		return protocol.Tool{}, errors.Wrapf(err, "failed to convert command to schema")
	}
	schemaBytes, err := json.Marshal(schema_)
	if err != nil {
		return protocol.Tool{}, errors.Wrapf(err, "failed to marshal schema")
	}
	tool := protocol.Tool{
		Name:        desc.Name,
		Description: desc.Short + "\n\n" + desc.Long,
		InputSchema: schemaBytes,
	}

	return tool, nil
}

// ListTools implements pkg.ToolProvider interface
func (p *ConfigToolProvider) ListTools(cursor string) ([]protocol.Tool, string, error) {
	var tools []protocol.Tool

	// Get tools from repositories
	repoCommands := p.repository.CollectCommands([]string{}, true)
	for _, cmd := range repoCommands {
		tool, err := ConvertCommandToTool(cmd.Description())
		if err != nil {
			return nil, "", errors.Wrapf(err, "failed to convert command to tool")
		}
		tools = append(tools, tool)
	}

	// Add shell commands
	for _, cmd := range p.shellCommands {
		tool, err := ConvertCommandToTool(cmd.Description())
		if err != nil {
			return nil, "", errors.Wrapf(err, "failed to convert command to tool")
		}
		tools = append(tools, tool)
	}

	// Handle cursor-based pagination if needed
	if cursor != "" {
		for i, tool := range tools {
			if tool.Name == cursor && i+1 < len(tools) {
				return tools[i+1:], "", nil
			}
		}
		return nil, "", nil
	}

	return tools, "", nil
}

// CallTool implements pkg.ToolProvider interface
func (p *ConfigToolProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	// Try repositories first
	if cmd, ok := p.repository.GetCommand(name); ok {
		return p.executeCommand(ctx, cmd, arguments)
	}

	// Try shell commands
	if cmd, ok := p.shellCommands[name]; ok {
		return p.executeCommand(ctx, cmd, arguments)
	}

	return nil, pkg.ErrToolNotFound
}

func (p *ConfigToolProvider) executeCommand(ctx context.Context, cmd cmds.Command, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	// Get parameter layers from command description
	parameterLayers := cmd.Description().Layers

	// Create empty parsed layers
	parsedLayers := layers.NewParsedLayers()

	// Create a map structure for the arguments
	argsMap := map[string]map[string]interface{}{
		layers.DefaultSlug: arguments,
	}

	// Build middleware chain
	var middlewareChain []middlewares.Middleware

	// Add parameter filtering middlewares if we have a config
	if config, ok := p.toolConfigs[cmd.Description().Name]; ok {
		// Convert our config to Parka's parameter filter
		paramFilter := p.createParkaParameterFilter(config)
		middlewareChain = append(middlewareChain, paramFilter.ComputeMiddlewares(false)...)
	}

	// Add the user arguments middleware last
	middlewareChain = append(middlewareChain,
		middlewares.SetFromDefaults(parameters.WithParseStepSource(parameters.SourceDefaults)),
		middlewares.UpdateFromMap(argsMap),
	)

	// Execute middlewares in order
	err := middlewares.ExecuteMiddlewares(
		parameterLayers,
		parsedLayers,
		middlewareChain...,
	)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	// Create a buffer to capture the command output
	buf := &strings.Builder{}

	// Run the command with parsed parameters
	switch c := cmd.(type) {
	case cmds.WriterCommand:
		if err := c.RunIntoWriter(ctx, parsedLayers, buf); err != nil {
			return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
		}
	case cmds.BareCommand:
		panic("BareCommand not supported yet")
	case cmds.GlazeCommand:
		panic("GlazeCommand not supported yet")
	default:
		panic("Unknown command type")
	}

	text := buf.String()
	l := 62 * 1024
	if len(text) > l {
		text = text[:l]
	}

	return protocol.NewToolResult(protocol.WithText(text)), nil
}

func (p *ConfigToolProvider) createParkaParameterFilter(sourceConfig *SourceConfig) *config.ParameterFilter {
	options := []config.ParameterFilterOption{}

	// Handle defaults
	if sourceConfig.Defaults != nil {
		layerParams := config.NewLayerParameters()
		for layer, params := range sourceConfig.Defaults {
			layerParams.Layers[layer] = params
		}
		options = append(options, config.WithReplaceDefaults(layerParams))
	}

	// Handle overrides
	if sourceConfig.Overrides != nil {
		layerParams := config.NewLayerParameters()
		for layer, params := range sourceConfig.Overrides {
			layerParams.Layers[layer] = params
		}
		options = append(options, config.WithReplaceOverrides(layerParams))
	}

	// Handle whitelist
	if sourceConfig.Whitelist != nil {
		whitelist := &config.ParameterFilterList{}
		for layer, params := range sourceConfig.Whitelist {
			whitelist.LayerParameters = map[string][]string{
				layer: params,
			}
		}
		options = append(options, config.WithWhitelist(whitelist))
	}

	// Handle blacklist
	if sourceConfig.Blacklist != nil {
		blacklist := &config.ParameterFilterList{}
		for layer, params := range sourceConfig.Blacklist {
			blacklist.LayerParameters = map[string][]string{
				layer: params,
			}
		}
		options = append(options, config.WithBlacklist(blacklist))
	}

	return config.NewParameterFilter(options...)
}
