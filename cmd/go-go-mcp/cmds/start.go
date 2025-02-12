package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/go-go-golems/go-go-mcp/pkg/config"
	config_provider "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/config-provider"
)

type StartCommandSettings struct {
	Transport    string   `glazed.parameter:"transport"`
	Port         int      `glazed.parameter:"port"`
	Repositories []string `glazed.parameter:"repositories"`
	Debug        bool     `glazed.parameter:"debug"`
	TracingDir   string   `glazed.parameter:"tracing-dir"`
	ConfigFile   string   `glazed.parameter:"config-file" help:"Path to the configuration file"`
	Profile      string   `glazed.parameter:"profile" help:"Profile to use from the configuration file"`
}

type StartCommand struct {
	*cmds.CommandDescription
}

func NewStartCommand() (*StartCommand, error) {
	defaultConfigFile, err := config.GetDefaultProfilesPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default profiles path")
	}

	return &StartCommand{
		CommandDescription: cmds.NewCommandDescription(
			"start",
			cmds.WithShort("Start the MCP server"),
			cmds.WithLong(`Start the MCP server using the specified transport.
		
Available transports:
- stdio: Standard input/output transport (default)
- sse: Server-Sent Events transport over HTTP`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"transport",
					parameters.ParameterTypeString,
					parameters.WithHelp("Transport type (stdio or sse)"),
					parameters.WithDefault("stdio"),
				),
				parameters.NewParameterDefinition(
					"port",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Port to listen on for SSE transport"),
					parameters.WithDefault(3001),
				),
				parameters.NewParameterDefinition(
					"repositories",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of directories containing shell command repositories"),
					parameters.WithDefault([]string{}),
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
			),
		),
	}, nil
}

func (c *StartCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &StartCommandSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Create server
	srv := server.NewServer(log.Logger)

	// Create tool provider options
	toolProviderOptions := []config_provider.ConfigToolProviderOption{
		config_provider.WithDebug(s.Debug),
		config_provider.WithWatch(true),
	}
	if s.TracingDir != "" {
		toolProviderOptions = append(toolProviderOptions, config_provider.WithTracingDir(s.TracingDir))
	}

	var toolProvider *config_provider.ConfigToolProvider
	var err error

	// Try to create tool provider from config file first
	if s.ConfigFile != "" {
		toolProvider, err = config_provider.CreateToolProviderFromConfig(s.ConfigFile, s.Profile, toolProviderOptions...)
		if err != nil {
			if !os.IsNotExist(err) || len(s.Repositories) == 0 {
				fmt.Fprintf(os.Stderr, "Run 'go-go-mcp config init' to create a starting configuration file, and further edit it with 'go-go-mcp config edit'\n")
				return errors.Wrap(err, "failed to create tool provider from config")
			}
			// Config file doesn't exist but we have repositories, continue with directories
		}
	}

	// If no tool provider yet and we have repositories, create from directories
	if toolProvider == nil && len(s.Repositories) > 0 {
		toolProvider, err = config_provider.CreateToolProviderFromDirectories(s.Repositories, toolProviderOptions...)
		if err != nil {
			return errors.Wrap(err, "failed to create tool provider from directories")
		}
	}

	if toolProvider == nil {
		return fmt.Errorf("no valid configuration source found (neither config file nor repositories)")
	}

	srv.GetRegistry().RegisterToolProvider(toolProvider)

	// Create a context that will be cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gctx := errgroup.WithContext(ctx)

	// Start file watcher
	g.Go(func() error {
		if err := toolProvider.Watch(gctx); err != nil {
			log.Error().Err(err).Msg("failed to start file watcher")
			return err
		}
		return nil
	})

	// Start server
	g.Go(func() error {
		var err error
		switch s.Transport {
		case "stdio":
			log.Info().Msg("Starting server with stdio transport")
			err = srv.StartStdio(gctx)
		case "sse":
			log.Info().Int("port", s.Port).Msg("Starting server with SSE transport")
			err = srv.StartSSE(gctx, s.Port)
		default:
			err = fmt.Errorf("invalid transport type: %s", s.Transport)
		}
		if err != nil && err != io.EOF {
			log.Error().Err(err).Msg("Server error")
			return err
		}
		return nil
	})

	// Add graceful shutdown handler
	g.Go(func() error {
		<-gctx.Done()
		log.Info().Msg("Initiating graceful shutdown")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := srv.Stop(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Error during shutdown")
			return err
		}
		log.Info().Msg("Server stopped gracefully")
		return nil
	})

	return g.Wait()
}
