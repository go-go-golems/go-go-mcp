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

type ServerSettings struct {
	Repositories []string `glazed.parameter:"repositories"`
	ConfigFile   string   `glazed.parameter:"config-file"`
	Profile      string   `glazed.parameter:"profile"`
	Debug        bool     `glazed.parameter:"debug"`
	TracingDir   string   `glazed.parameter:"tracing-dir"`
}

const ServerLayerSlug = "mcp-server"

func NewServerParameterLayer() (layers.ParameterLayer, error) {
	defaultConfigFile, err := config.GetDefaultProfilesPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default profiles path")
	}

	return layers.NewParameterLayer(ServerLayerSlug, "MCP Server Settings",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"repositories",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of directories containing shell command repositories"),
				parameters.WithDefault([]string{}),
			),
			parameters.NewParameterDefinition(
				"config-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to the configuration file"),
				parameters.WithDefault(defaultConfigFile),
			),
			parameters.NewParameterDefinition(
				"profile",
				parameters.ParameterTypeString,
				parameters.WithHelp("Profile to use from the configuration file"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"debug",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable debug mode for shell tool provider"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"tracing-dir",
				parameters.ParameterTypeString,
				parameters.WithHelp("Directory to store tool call traces"),
				parameters.WithDefault(""),
			),
		),
	)
}

// CreateToolProvider creates a tool provider from the given server settings
func CreateToolProvider(serverSettings *ServerSettings) (*config_provider.ConfigToolProvider, error) {
	// Create tool provider options
	toolProviderOptions := []config_provider.ConfigToolProviderOption{
		config_provider.WithDebug(serverSettings.Debug),
		config_provider.WithWatch(true),
	}
	if serverSettings.TracingDir != "" {
		toolProviderOptions = append(toolProviderOptions, config_provider.WithTracingDir(serverSettings.TracingDir))
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
			if !os.IsNotExist(err) || len(serverSettings.Repositories) == 0 {
				fmt.Fprintf(os.Stderr, "Run 'go-go-mcp config init' to create a starting configuration file, and further edit it with 'go-go-mcp config edit'\n")
				return nil, errors.Wrap(err, "failed to create tool provider from config")
			}
			// Config file doesn't exist but we have repositories, continue with directories
		}
	}

	// If no tool provider yet and we have repositories, create from directories
	if toolProvider == nil && len(serverSettings.Repositories) > 0 {
		toolProvider, err = config_provider.CreateToolProviderFromDirectories(serverSettings.Repositories, toolProviderOptions...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tool provider from directories")
		}
	}

	if toolProvider == nil {
		return nil, fmt.Errorf("no valid configuration source found (neither config file nor repositories)")
	}

	return toolProvider, nil
}
