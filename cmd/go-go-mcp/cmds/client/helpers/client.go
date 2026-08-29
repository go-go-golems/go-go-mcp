package helpers

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/client/layers"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type ClientSettings = layers.ClientSettings

// CreateClient initializes and connects an MCP client based on command flags.
func CreateClient(cmd *cobra.Command) (*mcp.ClientSession, error) {
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
	return createClient(&ClientSettings{Transport: transport, Server: serverURL, Command: cmdArgs})
}

// CreateClientFromSettings initializes and connects an MCP client from Glazed settings.
func CreateClientFromSettings(parsedValues *values.Values) (*mcp.ClientSession, error) {
	s := &ClientSettings{}
	if err := parsedValues.DecodeSectionInto(layers.ClientLayerSlug, s); err != nil {
		return nil, err
	}
	return createClient(s)
}

func createClient(s *ClientSettings) (*mcp.ClientSession, error) {
	var transport mcp.Transport
	switch s.Transport {
	case "sse":
		log.Debug().Msgf("Creating SSE client with server URL: %s", s.Server)
		transport = &mcp.SSEClientTransport{Endpoint: s.Server}
	case "streamable_http":
		log.Debug().Msgf("Creating Streamable HTTP client with server URL: %s", s.Server)
		transport = &mcp.StreamableClientTransport{Endpoint: s.Server}
	case "command":
		if len(s.Command) == 0 {
			return nil, fmt.Errorf("command is required for command transport")
		}
		log.Debug().Msgf("Creating stdio client to command: %v", s.Command)
		transport = &mcp.CommandTransport{Command: exec.Command(s.Command[0], s.Command[1:]...)} // #nosec G204 -- explicit user-requested MCP subprocess command
	default:
		return nil, fmt.Errorf("invalid transport type: %s", s.Transport)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "go-go-mcp", Version: "dev"}, nil)
	log.Debug().Msg("Connecting and initializing client")
	session, err := client.Connect(context.Background(), transport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect and initialize client: %w", err)
	}
	log.Debug().Msg("Client initialized")
	return session, nil
}
