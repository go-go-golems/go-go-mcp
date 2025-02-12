package server

import (
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
}
