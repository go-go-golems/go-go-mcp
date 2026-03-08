package layers

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	config_provider "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/config-provider"
	"github.com/pkg/errors"
)

// ServerSettings contains settings for the server
type ServerSettings struct {
	ServerConfigFile string   `glazed:"server-config-file"`
	Profile          string   `glazed:"profile"`
	Directories      []string `glazed:"directories"`
	Files            []string `glazed:"files"`
	Debug            bool     `glazed:"debug"`
	TracingDir       string   `glazed:"tracing-dir"`
	Watch            bool     `glazed:"watch"`
	ConvertDashes    bool     `glazed:"convert-dashes"`
	InternalServers  []string `glazed:"internal-servers"`
}

const ServerLayerSlug = "mcp-server"

// NewServerParameterLayer creates a new parameter layer for server settings
func NewServerParameterLayer() (schema.Section, error) {
	defaultServerConfigFile, err := config.GetDefaultProfilesPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default profiles path")
	}

	return schema.NewSection(ServerLayerSlug, "MCP Server Settings",
		schema.WithFields(
			fields.New(
				"server-config-file",
				fields.TypeString,
				fields.WithHelp("Server configuration file path"),
				fields.WithDefault(defaultServerConfigFile),
			),
			fields.New(
				"profile",
				fields.TypeString,
				fields.WithHelp("Profile to use from configuration file"),
				fields.WithDefault(""),
			),
			fields.New(
				"directories",
				fields.TypeStringList,
				fields.WithHelp("Directories to load commands from"),
				fields.WithDefault([]string{}),
			),
			fields.New(
				"files",
				fields.TypeStringList,
				fields.WithHelp("Files to load commands from"),
				fields.WithDefault([]string{}),
			),
			fields.New(
				"debug",
				fields.TypeBool,
				fields.WithHelp("Enable debug mode"),
				fields.WithDefault(false),
			),
			fields.New(
				"tracing-dir",
				fields.TypeString,
				fields.WithHelp("Directory to store tracing files"),
				fields.WithDefault(""),
			),
			fields.New(
				"watch",
				fields.TypeBool,
				fields.WithHelp("Watch for file changes"),
				fields.WithDefault(true),
			),
			fields.New(
				"convert-dashes",
				fields.TypeBool,
				fields.WithHelp("Convert dashes to underscores in tool names and arguments"),
				fields.WithDefault(false),
			),
			fields.New(
				"internal-servers",
				fields.TypeStringList,
				fields.WithHelp("List of internal servers to register (comma-separated). Available: sqlite,fetch,echo,scholarly"),
				fields.WithDefault([]string{}),
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
	if serverSettings.ServerConfigFile != "" {
		toolProvider, err = config_provider.CreateToolProviderFromConfig(
			serverSettings.ServerConfigFile,
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
