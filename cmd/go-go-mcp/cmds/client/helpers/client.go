package helpers

import (
	"context"
	"fmt"

	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/client/layers"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type ClientSettings = layers.ClientSettings

// CreateClient initializes and returns a new MCP client based on the provided flags.
func CreateClient(cmd *cobra.Command) (*mcpclient.Client, error) {
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
func CreateClientFromSettings(parsedLayers *glazed_layers.ParsedLayers) (*mcpclient.Client, error) {
	s := &ClientSettings{}
	if err := parsedLayers.InitializeStruct(layers.ClientLayerSlug, s); err != nil {
		return nil, err
	}

	return createClient(s)
}

func createClient(s *ClientSettings) (*mcpclient.Client, error) {
	var c *mcpclient.Client
	var err error
	switch s.Transport {
	case "sse":
		log.Debug().Msgf("Creating SSE client with server URL: %s", s.Server)
		c, err = mcpclient.NewSSEMCPClient(s.Server)
		if err != nil {
			return nil, err
		}
	case "streamable_http":
		log.Debug().Msgf("Creating Streamable HTTP client with server URL: %s", s.Server)
		c, err = mcpclient.NewStreamableHttpClient(s.Server)
		if err != nil {
			return nil, err
		}
	case "command":
		if len(s.Command) == 0 {
			return nil, fmt.Errorf("command is required for command transport")
		}
		log.Debug().Msgf("Creating stdio client to command: %v", s.Command)
		c, err = mcpclient.NewStdioMCPClient(s.Command[0], nil, s.Command[1:]...)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid transport type: %s", s.Transport)
	}

	// Start transport for non-stdio clients
	if s.Transport != "command" {
		if err := c.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to start client transport: %w", err)
		}
	}

	log.Debug().Msgf("Initializing client")
	_, err = c.Initialize(context.Background(), mcp.InitializeRequest{
		Request: mcp.Request{Method: string(mcp.MethodInitialize)},
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo:      mcp.Implementation{Name: "go-go-mcp", Version: "dev"},
			Capabilities:    mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}
	log.Debug().Msgf("Client initialized")
	return c, nil
}
