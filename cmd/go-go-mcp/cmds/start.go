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
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"

	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/server/layers"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
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
			),
			cmds.WithLayersList(serverLayer),
		),
	}, nil
}

func (c *StartCommand) Run(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
) error {
	s := &StartCommandSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	serverSettings := &layers.ServerSettings{}
	if err := parsedLayers.InitializeStruct(layers.ServerLayerSlug, serverSettings); err != nil {
		return err
	}

	// Create server
	srv := server.NewServer(log.Logger)

	toolProvider, err := layers.CreateToolProvider(serverSettings)
	if err != nil {
		return err
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
