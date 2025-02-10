package helpers

import (
	"context"
	"fmt"

	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/client/layers"
	"github.com/go-go-golems/go-go-mcp/pkg/client"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type ClientSettings = layers.ClientSettings

// CreateClient initializes and returns a new MCP client based on the provided flags.
func CreateClient(cmd *cobra.Command) (*client.Client, error) {
	transport, err := cmd.Flags().GetString("transport")
	if err != nil {
		return nil, fmt.Errorf("failed to get transport flag: %w", err)
	}

	serverURL, err := cmd.Flags().GetString("server")
	if err != nil {
		return nil, fmt.Errorf("failed to get server flag: %w", err)
	}

	cmdArgs, err := cmd.Flags().GetStringSlice("command")
	if err != nil {
		return nil, fmt.Errorf("failed to get command flag: %w", err)
	}

	return createClient(&ClientSettings{
		Transport: transport,
		Server:    serverURL,
		Command:   cmdArgs,
	})
}

// CreateClientFromSettings initializes and returns a new MCP client based on the provided settings.
func CreateClientFromSettings(parsedLayers *glazed_layers.ParsedLayers) (*client.Client, error) {
	s := &ClientSettings{}
	if err := parsedLayers.InitializeStruct(layers.ClientLayerSlug, s); err != nil {
		return nil, err
	}

	return createClient(s)
}

func createClient(s *ClientSettings) (*client.Client, error) {
	var t client.Transport
	var err error

	switch s.Transport {
	case "command":
		if len(s.Command) == 0 {
			return nil, fmt.Errorf("command is required for command transport")
		}
		log.Debug().Msgf("Creating command transport with args: %v", s.Command)
		t, err = client.NewCommandStdioTransport(log.Logger, s.Command[0], s.Command[1:]...)
		if err != nil {
			return nil, fmt.Errorf("failed to create command transport: %w", err)
		}
	case "sse":
		log.Debug().Msgf("Creating SSE transport with server URL: %s", s.Server)
		t = client.NewSSETransport(s.Server, log.Logger)
	default:
		return nil, fmt.Errorf("invalid transport type: %s", s.Transport)
	}

	// Create and initialize client
	c := client.NewClient(log.Logger, t)
	log.Debug().Msgf("Initializing client")
	err = c.Initialize(context.Background(), protocol.ClientCapabilities{
		Sampling: &protocol.SamplingCapability{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	log.Debug().Msgf("Client initialized")

	return c, nil
}
