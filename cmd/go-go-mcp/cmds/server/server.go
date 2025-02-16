package server

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ServerCmd is the root command for server-related operations
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Server management commands",
	Long:  `Commands for managing the MCP server, including starting the server and managing tools.`,
}

func init() {
	ServerCmd.AddCommand(ToolsCmd)
	startCmd, err := NewStartCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create start command")
	}
	cobraStartCmd, err := cli.BuildCobraCommandFromBareCommand(startCmd)
	cobra.CheckErr(err)
	ServerCmd.AddCommand(cobraStartCmd)
}
