package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewDisableCommand creates the disable command for a specific editor
func NewDisableCommand(editor string) *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "disable NAME",
		Short: "Disable server",
		Long:  fmt.Sprintf("Disable an MCP server for %s without removing it.", editor),
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

			// Check if server is already disabled
			disabled, err := store.IsServerDisabled(name)
			if err != nil {
				return err
			}
			if disabled {
				return fmt.Errorf("server '%s' is already disabled", name)
			}

			if err := store.DisableServer(name); err != nil {
				return err
			}

			if err := store.Save(); err != nil {
				return err
			}

			fmt.Printf("Successfully disabled MCP server '%s'\n", name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}
