package layers

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	config_provider "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/config-provider"
	"github.com/pkg/errors"
)

// ServerSettings contains settings for the server
type ServerSettings struct {
	ConfigFile      string   `glazed.parameter:"config-file"`
	Profile         string   `glazed.parameter:"profile"`
	Directories     []string `glazed.parameter:"directories"`
	Files           []string `glazed.parameter:"files"`
	Debug           bool     `glazed.parameter:"debug"`
	TracingDir      string   `glazed.parameter:"tracing-dir"`
	Watch           bool     `glazed.parameter:"watch"`
	ConvertDashes   bool     `glazed.parameter:"convert-dashes"`
	InternalServers []string `glazed.parameter:"internal-servers"`
}

const ServerLayerSlug = "mcp-server"

// NewServerParameterLayer creates a new parameter layer for server settings
func NewServerParameterLayer() (layers.ParameterLayer, error) {
	defaultConfigFile, err := config.GetDefaultProfilesPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default profiles path")
	}

	return layers.NewParameterLayer(ServerLayerSlug, "MCP Server Settings",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"config-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Configuration file path"),
				parameters.WithDefault(defaultConfigFile),
			),
			parameters.NewParameterDefinition(
				"profile",
				parameters.ParameterTypeString,
				parameters.WithHelp("Profile to use from configuration file"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"directories",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Directories to load commands from"),
				parameters.WithDefault([]string{}),
			),
			parameters.NewParameterDefinition(
				"files",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Files to load commands from"),
				parameters.WithDefault([]string{}),
			),
			parameters.NewParameterDefinition(
				"debug",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable debug mode"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"tracing-dir",
				parameters.ParameterTypeString,
				parameters.WithHelp("Directory to store tracing files"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"watch",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Watch for file changes"),
				parameters.WithDefault(true),
			),
			parameters.NewParameterDefinition(
				"convert-dashes",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Convert dashes to underscores in tool names and arguments"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"internal-servers",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of internal servers to register (comma-separated). Available: sqlite,fetch,echo,scholarly"),
				parameters.WithDefault([]string{}),
			),
		),
	)
}

// CreateToolProvider creates a tool provider from the given server settings
func CreateToolProvider(serverSettings *ServerSettings) (*config_provider.ConfigToolProvider, error) {
	// Create tool provider options
	toolProviderOptions := []config_provider.ConfigToolProviderOption{
		config_provider.WithDebug(serverSettings.Debug),
		config_provider.WithWatch(serverSettings.Watch),
		config_provider.WithConvertDashes(serverSettings.ConvertDashes),
	}
	if serverSettings.TracingDir != "" {
		toolProviderOptions = append(toolProviderOptions, config_provider.WithTracingDir(serverSettings.TracingDir))
	}

	// Add internal servers if specified
	if len(serverSettings.InternalServers) > 0 {
		toolProviderOptions = append(toolProviderOptions, config_provider.WithInternalServers(serverSettings.InternalServers))
	}

	var toolProvider *config_provider.ConfigToolProvider
	var err error

	// Try to create tool provider from config file first
	if serverSettings.ConfigFile != "" {
		toolProvider, err = config_provider.CreateToolProviderFromConfig(
			serverSettings.ConfigFile,
			serverSettings.Profile,
			toolProviderOptions...)
		if err != nil {
			if !os.IsNotExist(err) || len(serverSettings.Directories) == 0 {
				fmt.Fprintf(os.Stderr, "Run 'go-go-mcp config init' to create a starting configuration file, and further edit it with 'go-go-mcp config edit'\n")
				return nil, errors.Wrap(err, "failed to create tool provider from config")
			}
			// Config file doesn't exist but we have directories, continue with directories
		}
	}

	// If no tool provider yet and we have directories, create from directories
	if toolProvider == nil && len(serverSettings.Directories) > 0 {
		toolProvider, err = config_provider.CreateToolProviderFromDirectories(serverSettings.Directories, toolProviderOptions...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tool provider from directories")
		}
	}

	if toolProvider == nil {
		return nil, fmt.Errorf("no valid configuration source found (neither config file nor directories)")
	}

	return toolProvider, nil
}
