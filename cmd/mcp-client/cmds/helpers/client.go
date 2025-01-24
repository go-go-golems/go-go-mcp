package helpers

import (
	"context"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/client"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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

	var t client.Transport

	switch transport {
	case "command":
		if len(cmdArgs) == 0 {
			return nil, fmt.Errorf("command is required for command transport")
		}
		log.Debug().Msgf("Creating command transport with args: %v", cmdArgs)
		t, err = client.NewCommandStdioTransport(log.Logger, cmdArgs[0], cmdArgs[1:]...)
		if err != nil {
			return nil, fmt.Errorf("failed to create command transport: %w", err)
		}
	case "sse":
		log.Debug().Msgf("Creating SSE transport with server URL: %s", serverURL)
		t = client.NewSSETransport(serverURL, log.Logger)
	default:
		return nil, fmt.Errorf("invalid transport type: %s", transport)
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
