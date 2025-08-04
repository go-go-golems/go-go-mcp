package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewEnableCommand creates the enable command for a specific editor
func NewEnableCommand(editor string) *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "enable NAME",
		Short: "Enable server",
		Long:  fmt.Sprintf("Enable a disabled MCP server for %s.", editor),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewStoreWithTarget(editor, target)
			if err != nil {
				return err
			}

			name := args[0]
			
			// Check if server exists
			_, exists, err := store.GetServer(name)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("server '%s' not found", name)
			}

			// Check if server is already enabled
			disabled, err := store.IsServerDisabled(name)
			if err != nil {
				return err
			}
			if !disabled {
				return fmt.Errorf("server '%s' is already enabled", name)
			}

			if err := store.EnableServer(name); err != nil {
				return err
			}

			if err := store.Save(); err != nil {
				return err
			}

			fmt.Printf("Successfully enabled MCP server '%s'\n", name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}
