package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/server/layers"
	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/resources"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/examples"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/sse"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/stdio"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type StartCommandSettings struct {
	Transport       string   `glazed.parameter:"transport"`
	Port            int      `glazed.parameter:"port"`
	InternalServers []string `glazed.parameter:"internal-servers"`
}

type StartCommand struct {
	*cmds.CommandDescription
}

func NewStartCommand() (*StartCommand, error) {
	serverLayer, err := layers.NewServerParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create server parameter layer")
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
					"internal-servers",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of internal servers to register (comma-separated). Available: sqlite,fetch,echo"),
					parameters.WithDefault([]string{}),
				),
			),
			cmds.WithLayersList(serverLayer),
		),
	}, nil
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
		if err := examples.RegisterSQLiteTool(registry); err != nil {
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

	return nil
}

func (c *StartCommand) Run(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
) error {
	logger := log.Logger

	s_ := &StartCommandSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s_); err != nil {
		return err
	}

	// Get transport type from flags
	transportType := s_.Transport
	port := s_.Port

	// Create transport based on type
	var t transport.Transport
	var err error

	switch transportType {
	case "sse":
		t, err = sse.NewSSETransport(
			transport.WithLogger(logger),
			transport.WithSSEOptions(transport.SSEOptions{
				Addr: fmt.Sprintf(":%d", port),
			}),
		)
	case "stdio":
		t, err = stdio.NewStdioTransport(
			transport.WithLogger(logger),
		)
	default:
		return fmt.Errorf("unsupported transport type: %s", transportType)
	}

	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	// Get server settings
	serverSettings := &layers.ServerSettings{}
	if err := parsedLayers.InitializeStruct(layers.ServerLayerSlug, serverSettings); err != nil {
		return err
	}

	// Create tool provider
	configToolProvider, err := layers.CreateToolProvider(serverSettings)
	if err != nil {
		return err
	}

	// Initialize the final tool provider
	var toolProvider pkg.ToolProvider = configToolProvider

	// Register internal servers if specified
	if len(s_.InternalServers) > 0 {
		registry := tool_registry.NewRegistry()

		// Register the internal servers
		if err := registerInternalServers(registry, s_.InternalServers); err != nil {
			return err
		}

		// Combine the registry with the config tool provider
		toolProvider = tools.CombineProviders(configToolProvider, registry)
	}

	// Create resource provider
	resourceProvider := resources.NewRegistry()

	// Create and start server with transport and providers
	s := server.NewServer(logger, t,
		server.WithToolProvider(toolProvider),
		server.WithResourceProvider(resourceProvider))

	// Create a context that will be cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cancelCtx, cancel := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(cancelCtx)

	// Start file watcher
	g.Go(func() error {
		if err := configToolProvider.Watch(gctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error().Err(err).Msg("failed to run file watcher")
			} else {
				logger.Debug().Msg("File watcher cancelled")
			}
			return err
		}
		logger.Info().Msg("File watcher finished")
		return nil
	})

	// Start server
	g.Go(func() error {
		defer cancel()
		if err := s.Start(gctx); err != nil && err != io.EOF {
			logger.Error().Err(err).Msg("Server error")
			return err
		}
		if err != nil {
			logger.Warn().Err(err).Msg("Server stopped with error")
		}
		logger.Info().Msg("Server stopped")
		return nil
	})

	// Add graceful shutdown handler
	g.Go(func() error {
		<-gctx.Done()
		logger.Info().Msg("Initiating graceful shutdown")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := s.Stop(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("Error during shutdown")
			return err
		}
		logger.Info().Msg("Server stopped gracefully")
		return nil
	})

	return g.Wait()
}
