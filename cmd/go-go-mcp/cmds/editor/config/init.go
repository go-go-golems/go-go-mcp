package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command for a specific editor
func NewInitCommand(editor string) *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize config",
		Long:  fmt.Sprintf("Initialize a new configuration file for %s.", editor),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewStoreWithTarget(editor, target)
			if err != nil {
				return err
			}

			if err := store.Save(); err != nil {
				return fmt.Errorf("failed to initialize configuration: %w", err)
			}

			configPath := store.GetConfigPath()
			fmt.Printf("Successfully initialized configuration file: %s\n", configPath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}
