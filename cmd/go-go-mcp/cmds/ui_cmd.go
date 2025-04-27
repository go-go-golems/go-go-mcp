package cmds

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-mcp/pkg/ui/tui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Interactive terminal UI for managing MCP servers",
		Long:  `A terminal user interface for managing MCP server configurations for Cursor and Claude desktop.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debug().Msg("Starting UI")
			p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
			_, err := p.Run()
			if err != nil {
				log.Error().Err(err).Msg("Error running UI")
			}
			log.Debug().Msg("UI exited")
			return err
		},
	}

	return cmd
}
