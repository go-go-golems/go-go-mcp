package config_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/mcp"
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
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/examples"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	parka_config "github.com/go-go-golems/parka/pkg/handlers/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ConfigToolProvider implements pkg.ToolProvider interface
type ConfigToolProvider struct {
	repository      *repositories.Repository
	shellCommands   map[string]cmds.Command
	toolConfigs     map[string]*config.SourceConfig
	helpSystem      *help.HelpSystem
	debug           bool
	tracingDir      string
	directories     []repositories.Directory
	files           []string
	watching        bool
	convertDashes   bool // controls whether to convert dashes to underscores in tool names
	internalServers []string
}

type ConfigToolProviderOption func(*ConfigToolProvider) error

var _ pkg.ToolProvider = &ConfigToolProvider{}

func WithDebug(debug bool) ConfigToolProviderOption {
	return func(p *ConfigToolProvider) error {
		p.debug = debug
		return nil
	}
}

func WithTracingDir(dir string) ConfigToolProviderOption {
	return func(p *ConfigToolProvider) error {
		p.tracingDir = dir
		return nil
	}
}

func WithDirectories(directories []repositories.Directory) ConfigToolProviderOption {
	return func(p *ConfigToolProvider) error {
		p.directories = append(p.directories, directories...)
		return nil
	}
}

func WithFiles(files []string) ConfigToolProviderOption {
	return func(p *ConfigToolProvider) error {
		p.files = append(p.files, files...)
		return nil
	}
}

func WithConfig(config_ *config.Config, profile string) ConfigToolProviderOption {
	return func(p *ConfigToolProvider) error {
		if profile == "" {
			profile = config_.DefaultProfile
		}
		if profile == "" {
			return nil
		}
		profileConfig, ok := config_.Profiles[profile]
		if !ok {
			return errors.Errorf("profile %s not found", profile)
		}

		if profileConfig.Tools == nil {
			return nil
		}

		// Load directories
		directories := []repositories.Directory{}
		for _, dir := range profileConfig.Tools.Directories {
			absPath, err := filepath.Abs(dir.Path)
			if err != nil {
				return errors.Wrapf(err, "failed to get absolute path for %s", dir.Path)
			}

			directories = append(directories, repositories.Directory{
				FS:               os.DirFS(absPath),
				RootDirectory:    ".",
				RootDocDirectory: "doc",
				WatchDirectory:   absPath,
				Name:             dir.Path,
				SourcePrefix:     "file",
			})
		}
		p.directories = directories

		// Collect file paths
		files := []string{}
		for _, file := range profileConfig.Tools.Files {
			absPath, err := filepath.Abs(file.Path)
			if err != nil {
				return errors.Wrapf(err, "failed to get absolute path for %s", file.Path)
			}
			files = append(files, absPath)
			p.toolConfigs[filepath.Base(absPath)] = &file
		}

		p.files = files

		return nil
	}
}

func WithWatch(watch bool) ConfigToolProviderOption {
	return func(p *ConfigToolProvider) error {
		p.watching = watch
		return nil
	}
}

func WithConvertDashes(convert bool) ConfigToolProviderOption {
	return func(p *ConfigToolProvider) error {
		p.convertDashes = convert
		return nil
	}
}

func WithInternalServers(internalServers []string) ConfigToolProviderOption {
	return func(p *ConfigToolProvider) error {
		p.internalServers = internalServers
		return nil
	}
}

// NewConfigToolProvider creates a new ConfigToolProvider with the given options
func NewConfigToolProvider(options ...ConfigToolProviderOption) (*ConfigToolProvider, error) {
	provider := &ConfigToolProvider{
		shellCommands: make(map[string]cmds.Command),
		toolConfigs:   make(map[string]*config.SourceConfig),
		helpSystem:    help.NewHelpSystem(),
		directories:   []repositories.Directory{},
		files:         []string{},
		watching:      false,
		convertDashes: false,
	}

	for _, option := range options {
		if err := option(provider); err != nil {
			return nil, err
		}
	}

	// Create repository with collected directories and shell command loader
	provider.repository = repositories.NewRepository(
		repositories.WithDirectories(provider.directories...),
		repositories.WithFiles(provider.files...),
		repositories.WithCommandLoader(&mcp_cmds.ShellCommandLoader{}),
	)

	// Load repository commands
	if err := provider.repository.LoadCommands(provider.helpSystem); err != nil {
		return nil, errors.Wrap(err, "failed to load repository commands")
	}

	return provider, nil
}

func ConvertCommandToTool(desc *cmds.CommandDescription) (protocol.Tool, error) {
	schema_, err := desc.ToJsonSchema()
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

// convertToolName converts tool names between dash and underscore format based on the provider's configuration
func (p *ConfigToolProvider) convertToolName(name string) string {
	if !p.convertDashes {
		return name
	}
	return strings.ReplaceAll(name, "-", "_")
}

// revertToolName converts tool names from underscore back to dash format based on the provider's configuration
func (p *ConfigToolProvider) revertToolName(name string) string {
	if !p.convertDashes {
		return name
	}
	return strings.ReplaceAll(name, "_", "-")
}

// ListTools implements pkg.ToolProvider interface
func (p *ConfigToolProvider) ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	var tools []protocol.Tool

	// If internal servers are enabled, create a registry with those tools
	var internalRegistry *tool_registry.Registry
	if len(p.internalServers) > 0 {
		internalRegistry = tool_registry.NewRegistry()
		if err := registerInternalServers(internalRegistry, p.internalServers); err != nil {
			return nil, "", errors.Wrap(err, "failed to register internal servers")
		}

		// List internal tools first
		internalTools, _, err := internalRegistry.ListTools(ctx, "")
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to list internal tools")
		}
		tools = append(tools, internalTools...)
	}

	repoCommands := p.repository.CollectCommands([]string{}, true)

	for _, cmd := range repoCommands {
		tool, err := ConvertCommandToTool(cmd.Description())
		if err != nil {
			return nil, "", errors.Wrapf(err, "failed to convert command to tool")
		}
		if p.convertDashes {
			tool.Name = p.convertToolName(tool.Name)
		}
		tools = append(tools, tool)
	}

	// Add shell commands
	for _, cmd := range p.shellCommands {
		tool, err := ConvertCommandToTool(cmd.Description())
		if err != nil {
			return nil, "", errors.Wrapf(err, "failed to convert command to tool")
		}
		if p.convertDashes {
			tool.Name = p.convertToolName(tool.Name)
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
	// Convert the tool name back to its original form if needed
	originalName := p.revertToolName(name)

	// First check if the tool is available from internal servers
	if len(p.internalServers) > 0 {
		internalRegistry := tool_registry.NewRegistry()
		if err := registerInternalServers(internalRegistry, p.internalServers); err != nil {
			return nil, errors.Wrap(err, "failed to register internal servers")
		}

		result, err := internalRegistry.CallTool(ctx, originalName, arguments)
		if err == nil {
			return result, nil
		}
		// If the error is not "tool not found", return the error
		if !errors.Is(err, pkg.ErrToolNotFound) {
			return nil, err
		}
		// Otherwise continue looking in other providers
	}

	cmd, ok := p.repository.GetCommand(originalName)

	if ok {
		// Convert argument keys if needed
		if p.convertDashes {
			newArgs := make(map[string]interface{}, len(arguments))
			for k, v := range arguments {
				newArgs[p.revertToolName(k)] = v
			}
			arguments = newArgs
		}
		return p.executeCommand(ctx, cmd, arguments)
	}

	// Try shell commands
	if cmd, ok := p.shellCommands[originalName]; ok {
		// Convert argument keys if needed
		if p.convertDashes {
			newArgs := make(map[string]interface{}, len(arguments))
			for k, v := range arguments {
				newArgs[p.revertToolName(k)] = v
			}
			arguments = newArgs
		}
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
			return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("%s\n\nOutput so far:\n%s", err.Error(), buf.String()))), nil
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

func (p *ConfigToolProvider) createParkaParameterFilter(sourceConfig *config.SourceConfig) *parka_config.ParameterFilter {
	options := []parka_config.ParameterFilterOption{}

	// Handle defaults
	if sourceConfig.Defaults != nil {
		layerParams := parka_config.NewLayerParameters()
		for layer, params := range sourceConfig.Defaults {
			layerParams.Layers[layer] = params
		}
		options = append(options, parka_config.WithReplaceDefaults(layerParams))
	}

	// Handle overrides
	if sourceConfig.Overrides != nil {
		layerParams := parka_config.NewLayerParameters()
		for layer, params := range sourceConfig.Overrides {
			layerParams.Layers[layer] = params
		}
		options = append(options, parka_config.WithReplaceOverrides(layerParams))
	}

	// Handle whitelist
	if sourceConfig.Whitelist != nil {
		whitelist := &parka_config.ParameterFilterList{}
		for layer, params := range sourceConfig.Whitelist {
			whitelist.LayerParameters = map[string][]string{
				layer: params,
			}
		}
		options = append(options, parka_config.WithWhitelist(whitelist))
	}

	// Handle blacklist
	if sourceConfig.Blacklist != nil {
		blacklist := &parka_config.ParameterFilterList{}
		for layer, params := range sourceConfig.Blacklist {
			blacklist.LayerParameters = map[string][]string{
				layer: params,
			}
		}
		options = append(options, parka_config.WithBlacklist(blacklist))
	}

	return parka_config.NewParameterFilter(options...)
}

// Watch starts watching all configured directories for changes
func (p *ConfigToolProvider) Watch(ctx context.Context) error {
	if !p.watching {
		return nil
	}

	// Use the repository's built-in watching functionality
	return p.repository.Watch(ctx)
}

// CreateToolProviderFromConfig creates a tool provider from a config file and profile
func CreateToolProviderFromConfig(configFile string, profile string, options ...ConfigToolProviderOption) (*ConfigToolProvider, error) {
	// Handle configuration file if provided
	if configFile != "" {
		cfg, err := config.LoadFromFile(configFile)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrap(err, "failed to load configuration file")
			}
			// Config file doesn't exist, continue with other options
			log.Warn().Str("config", configFile).Msg("Configuration file not found")
		} else {
			// Determine profile
			if profile == "" {
				profile = cfg.DefaultProfile
			}

			options = append(options, WithConfig(cfg, profile))
		}
	}

	provider, err := NewConfigToolProvider(options...)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// CreateToolProviderFromDirectories creates a tool provider from a list of directories
func CreateToolProviderFromDirectories(directories []string, options ...ConfigToolProviderOption) (*ConfigToolProvider, error) {
	if len(directories) > 0 {
		dirs := []repositories.Directory{}
		for _, repoPath := range directories {
			dir := os.ExpandEnv(repoPath)
			// check if dir exists
			if fi, err := os.Stat(dir); os.IsNotExist(err) || !fi.IsDir() {
				log.Warn().Str("path", dir).Msg("Repository directory does not exist or is not a directory")
				continue
			}
			dirs = append(dirs, repositories.Directory{
				FS:               os.DirFS(dir),
				RootDirectory:    ".",
				RootDocDirectory: "doc",
				WatchDirectory:   dir,
				Name:             dir,
				SourcePrefix:     "file",
			})
		}

		if len(dirs) > 0 {
			options = append(options, WithDirectories(dirs))
		}
	}

	provider, err := NewConfigToolProvider(options...)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// registerInternalServers registers the requested internal servers in the registry
func registerInternalServers(registry *tool_registry.Registry, serverList []string) error {
	// Create a map for faster lookups
	serversMap := make(map[string]bool)
	for _, server := range serverList {
		serversMap[server] = true
	}

	// Register the requested servers
	if serversMap["sqlite"] {
		if err := examples.RegisterSQLiteSessionTools(registry); err != nil {
			return errors.Wrap(err, "failed to register sqlite tool")
		}
	}

	if serversMap["fetch"] {
		if err := examples.RegisterFetchTool(registry); err != nil {
			return errors.Wrap(err, "failed to register fetch tool")
		}
	}

	if serversMap["echo"] {
		if err := examples.RegisterEchoTool(registry); err != nil {
			return errors.Wrap(err, "failed to register echo tool")
		}
	}

	if serversMap["prompto"] {
		if err := examples.RegisterPromptoTools(registry); err != nil {
			return errors.Wrap(err, "failed to register prompto tools")
		}
	}

	if serversMap["scholarly"] {
		if err := mcp.RegisterScholarlyTools(registry); err != nil {
			return errors.Wrap(err, "failed to register scholarly tools")
		}
	}

	return nil
}
