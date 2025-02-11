package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-mcp/pkg/config"
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
	toolProviderOptions := []config.ConfigToolProviderOption{
		config.WithDebug(s.Debug),
	}
	if s.TracingDir != "" {
		toolProviderOptions = append(toolProviderOptions, config.WithTracingDir(s.TracingDir))
	}

	// Handle configuration file if provided
	if s.ConfigFile != "" {
		cfg, err := config.LoadFromFile(s.ConfigFile)
		if err != nil {
			return errors.Wrap(err, "failed to load configuration file")
		}

		// Determine profile
		profile := s.Profile
		if profile == "" {
			profile = cfg.DefaultProfile
		}

		toolProviderOptions = append(toolProviderOptions, config.WithConfig(cfg, profile))
	}

	// Handle repository directories
	if len(s.Repositories) > 0 {
		directories := []repositories.Directory{}
		for _, repoPath := range s.Repositories {
			dir := os.ExpandEnv(repoPath)
			// check if dir exists
			if fi, err := os.Stat(dir); os.IsNotExist(err) || !fi.IsDir() {
				log.Warn().Str("path", dir).Msg("Repository directory does not exist or is not a directory")
				continue
			}
			directories = append(directories, repositories.Directory{
				FS:               os.DirFS(dir),
				RootDirectory:    ".",
				RootDocDirectory: "doc",
				WatchDirectory:   dir,
				Name:             dir,
				SourcePrefix:     "file",
			})
		}

		if len(directories) > 0 {
			toolProviderOptions = append(toolProviderOptions, config.WithDirectories(directories))
		}
	}

	// Create and register tool provider
	toolProvider, err := config.NewConfigToolProvider(toolProviderOptions...)
	if err != nil {
		return errors.Wrap(err, "failed to create tool provider")
	}
	srv.GetRegistry().RegisterToolProvider(toolProvider)

	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		// Start server with selected transport
		var err error
		switch s.Transport {
		case "stdio":
			log.Info().Msg("Starting server with stdio transport")
			err = srv.StartStdio(ctx)
		case "sse":
			log.Info().Int("port", s.Port).Msg("Starting server with SSE transport")
			err = srv.StartSSE(ctx, s.Port)
		default:
			err = fmt.Errorf("invalid transport type: %s", s.Transport)
		}
		errChan <- err
	}()

	// Wait for either server error or interrupt signal
	select {
	case err := <-errChan:
		if err != nil && err != io.EOF {
			log.Error().Err(err).Msg("Server error")
			return err
		}
		return nil
	case sig := <-sigChan:
		log.Info().Str("signal", sig.String()).Msg("Received signal, initiating graceful shutdown")
		// Cancel context to initiate shutdown
		cancel()
		// Create a timeout context for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := srv.Stop(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Error during shutdown")
			return err
		}
		log.Info().Msg("Server stopped gracefully")
		return nil
	}
}
