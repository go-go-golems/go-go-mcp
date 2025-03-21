package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/rs/zerolog/log"
)

type StartCommand struct {
	*cmds.CommandDescription
}

type StartSettings struct {
	Directory string `glazed.parameter:"directory"`
	Port      int    `glazed.parameter:"port"`
}

func NewStartCommand() (*StartCommand, error) {
	return &StartCommand{
		CommandDescription: cmds.NewCommandDescription(
			"start",
			cmds.WithShort("Start the UI server"),
			cmds.WithLong("Start a UI server that renders YAML UI definitions from a directory"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"port",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Port to run the server on"),
					parameters.WithDefault(8080),
					parameters.WithShortFlag("p"),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"directory",
					parameters.ParameterTypeString,
					parameters.WithHelp("Directory containing YAML UI definitions"),
					parameters.WithRequired(true),
				),
			),
		),
	}, nil
}

func (c *StartCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &StartSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Create a context that can be cancelled by signals
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server, err := NewServer(s.Directory)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	if err := server.Start(ctx, s.Port); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	log.Info().Msg("Server stopped gracefully")
	return nil
}

var _ cmds.BareCommand = &StartCommand{}
