package cmds

import (
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/client"
	"github.com/spf13/cobra"
)

var ClientCmd = &cobra.Command{
	Use:   "client",
	Short: "MCP client functionality",
	Long:  `Client commands for interacting with MCP servers`,
}

func InitClientCommand(helpSystem *help.HelpSystem) error {
	// Add client subcommands
	ClientCmd.AddCommand(client.ToolsCmd)
	ClientCmd.AddCommand(client.ResourcesCmd)
	ClientCmd.AddCommand(client.PromptsCmd)

	return nil
}
