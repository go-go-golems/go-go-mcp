package editor

import (
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/editor/config"
	"github.com/spf13/cobra"
)

// NewEditorCommand creates the main editor command group
func NewEditorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "editor",
		Short: "Manage editor configurations and settings",
		Long: `Commands for managing various editor configurations and settings.
		
This command group provides functionality for managing MCP server configurations
across different editors and their specific settings.`,
	}

	// Add config subcommand
	cmd.AddCommand(config.NewConfigCommand())

	return cmd
}
