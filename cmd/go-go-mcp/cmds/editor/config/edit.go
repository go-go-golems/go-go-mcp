package config

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// NewEditCommand creates the edit command for a specific editor
func NewEditCommand(editor string) *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit config file",
		Long:  fmt.Sprintf("Edit the configuration file for %s in your default editor.", editor),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewStoreWithTarget(editor, target)
			if err != nil {
				return err
			}

			configPath := store.GetConfigPath()

			editorCmd := os.Getenv("EDITOR")
			if editorCmd == "" {
				editorCmd = "vi"
			}

			c := exec.Command(editorCmd, configPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}
