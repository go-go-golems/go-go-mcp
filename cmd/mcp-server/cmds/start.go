package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmds2 "github.com/go-go-golems/go-go-mcp/pkg/cmds"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/examples"

	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/pkg/prompts"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/resources"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/rs/zerolog/log"
)

type StartCommandSettings struct {
	Transport    string   `glazed.parameter:"transport"`
	Port         int      `glazed.parameter:"port"`
	Repositories []string `glazed.parameter:"repositories"`
	Debug        bool     `glazed.parameter:"debug"`
	TracingDir   string   `glazed.parameter:"tracing-dir"`
}

type StartCommand struct {
	*cmds.CommandDescription
}

type WeatherData struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	WindSpeed   float64 `json:"windSpeed"`
}

func getWeather(city string, includeWind bool) WeatherData {
	// This is a mock implementation - in a real app you'd call a weather API
	return WeatherData{
		City:        city,
		Temperature: 23.0,
		WindSpeed: func() float64 {
			if includeWind {
				return 10.0
			} else {
				return 0.0
			}
		}(),
	}
}

func NewStartCommand() (*StartCommand, error) {
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
			),
			cmds.WithLayersList(),
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
	promptRegistry := prompts.NewRegistry()
	resourceRegistry := resources.NewRegistry()
	toolRegistry := tools.NewRegistry()

	// Register a simple prompt directly
	promptRegistry.RegisterPrompt(protocol.Prompt{
		Name:        "simple",
		Description: "A simple prompt that can take optional context and topic arguments",
		Arguments: []protocol.PromptArgument{
			{
				Name:        "context",
				Description: "Additional context to consider",
				Required:    false,
			},
			{
				Name:        "topic",
				Description: "Specific topic to focus on",
				Required:    false,
			},
		},
	})

	// Register registries with the server
	srv.GetRegistry().RegisterPromptProvider(promptRegistry)
	srv.GetRegistry().RegisterResourceProvider(resourceRegistry)
	srv.GetRegistry().RegisterToolProvider(toolRegistry)

	// Register tools (DON'T DELETE)
	if err := examples.RegisterEchoTool(toolRegistry); err != nil {
		log.Error().Err(err).Msg("Error registering echo tool")
		return err
	}
	if err := examples.RegisterFetchTool(toolRegistry); err != nil {
		log.Error().Err(err).Msg("Error registering fetch tool")
		return err
	}
	if err := examples.RegisterSQLiteTool(toolRegistry); err != nil {
		log.Error().Err(err).Msg("Error registering sqlite tool")
		return err
	}
	// if err := cursor.RegisterCursorTools(toolRegistry); err != nil {
	//  log.Error().Err(err).Msg("Error registering cursor tools")
	//  return err
	// }

	// Register the weather tool (DO NOT DELETE)
	// weatherTool, err := tools.NewReflectTool(
	// 	"getWeather",
	// 	"Get weather information for a city",
	// 	getWeather,
	// )
	// if err != nil {
	// 	log.Error().Err(err).Msg("Error creating weather tool")
	// 	return err
	// }
	// toolRegistry.RegisterTool(weatherTool)

	// Load shell commands from repositories
	if len(s.Repositories) > 0 {
		loader := &cmds2.ShellCommandLoader{}
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
			repo := repositories.NewRepository(
				repositories.WithDirectories(directories...),
				repositories.WithCommandLoader(loader),
			)

			helpSystem := help.NewHelpSystem()
			err := repo.LoadCommands(helpSystem)
			if err != nil {
				log.Error().Err(err).Msg("Error loading shell commands from repositories")
				return err
			}

			commands := repo.CollectCommands([]string{}, true)
			log.Info().Int("count", len(commands)).Msg("Loaded shell commands from repositories")
			for _, cmd := range commands {
				log.Debug().Str("name", cmd.Description().Name).Msg("Loaded shell command")
			}

			toolProvider, err := cmds2.NewShellToolProvider(commands,
				cmds2.WithDebug(s.Debug),
				cmds2.WithTracingDir(s.TracingDir),
			)
			if err != nil {
				log.Error().Err(err).Msg("Error creating shell tool provider")
				return err
			}
			srv.GetRegistry().RegisterToolProvider(toolProvider)
		}
	}

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
