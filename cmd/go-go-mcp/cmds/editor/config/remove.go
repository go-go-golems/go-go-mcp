package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRemoveCommand creates the remove command for a specific editor
func NewRemoveCommand(editor string) *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "remove NAME",
		Short: "Remove MCP server",
		Long:  fmt.Sprintf("Remove an MCP server configuration from %s.", editor),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewStoreWithTarget(editor, target)
			if err != nil {
				return err
			}

			name := args[0]
			
			// Check if server exists before removing
			_, exists, err := store.GetServer(name)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("server '%s' not found", name)
			}

			if err := store.RemoveServer(name); err != nil {
				return err
			}

			if err := store.Save(); err != nil {
				return err
			}

			fmt.Printf("Successfully removed MCP server '%s'\n", name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}
