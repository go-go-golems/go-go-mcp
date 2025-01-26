package cmds

import (
	"context"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/server/transports/stdio"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewBridgeCommand(logger zerolog.Logger) *cobra.Command {
	var sseURL string

	cmd := &cobra.Command{
		Use:   "bridge",
		Short: "Start a stdio server that bridges to an SSE server",
		Long: `Start a stdio server that forwards all requests to an SSE server.
This is useful when you want to connect a stdio client to a remote SSE server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sseURL == "" {
				return fmt.Errorf("SSE URL is required")
			}

			server := stdio.NewSSEBridgeServer(logger, sseURL)
			return server.Start(context.Background())
		},
	}

	cmd.Flags().StringVarP(&sseURL, "sse-url", "s", "", "URL of the SSE server to connect to (required)")
	_ = cmd.MarkFlagRequired("sse-url")

	return cmd
}
