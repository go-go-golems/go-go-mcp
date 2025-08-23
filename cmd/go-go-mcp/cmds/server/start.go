package server

import (
	"context"
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
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/resources"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type StartCommandSettings struct {
	Transport string `glazed.parameter:"transport"`
	Port      int    `glazed.parameter:"port"`
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
- sse: Server-Sent Events transport over HTTP
- streamable_http: Streamable HTTP transport with WebSocket support`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"transport",
					parameters.ParameterTypeString,
					parameters.WithHelp("Transport type (stdio, sse, or streamable_http)"),
					parameters.WithDefault("stdio"),
				),
				parameters.NewParameterDefinition(
					"port",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Port to listen on for SSE and streamable HTTP transport"),
					parameters.WithDefault(3001),
				),
			),
			cmds.WithLayersList(serverLayer),
		),
	}, nil
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
	_ = toolProvider

	// Create resource provider (not yet wired into mcp-go backend)
	_ = resources.NewRegistry()

	// Build a registry adapter that proxies calls to the tool provider
	reg := tool_registry.NewRegistry()
	// List tools once at startup and register into our registry with proxy handler
	toolsList, _, err := toolProvider.ListTools(ctx, "")
	if err != nil {
		return errors.Wrap(err, "failed to list tools from provider")
	}
	for _, t := range toolsList {
		toolImpl, err := tools.NewToolImpl(t.Name, t.Description, t.InputSchema)
		if err != nil {
			return errors.Wrapf(err, "failed to create tool %s", t.Name)
		}
		name := t.Name
		reg.RegisterToolWithHandler(toolImpl, func(ctx context.Context, _ tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			return toolProvider.CallTool(ctx, name, arguments)
		})
	}

	// Create embeddable server config
	cfg := embeddable.NewServerConfig()
	_ = embeddable.WithName("go-go-mcp")(cfg)
	_ = embeddable.WithDefaultTransport(transportType)(cfg)
	_ = embeddable.WithDefaultPort(port)(cfg)
	_ = embeddable.WithToolRegistry(reg)(cfg)
	if len(serverSettings.InternalServers) > 0 {
		_ = embeddable.WithInternalServers(serverSettings.InternalServers...)(cfg)
	}

	// Create backend
	backend, err := embeddable.NewBackend(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to create backend")
	}

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

	// Start backend
	g.Go(func() error {
		defer cancel()
		if err := backend.Start(gctx); err != nil && err != io.EOF {
			logger.Error().Err(err).Msg("Server error")
			return err
		}
		logger.Info().Msg("Server stopped")
		return nil
	})

	// Add graceful shutdown handler (best effort; transports may not support Shutdown yet)
	g.Go(func() error {
		<-gctx.Done()
		logger.Info().Msg("Initiating graceful shutdown")
		_, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		// TODO: add backend Shutdown() if/when needed
		return nil
	})

	return g.Wait()
}
