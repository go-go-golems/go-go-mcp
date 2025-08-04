package config

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/spf13/cobra"
)

// NewAddGoGoCommand creates the add-go-go command for a specific editor
func NewAddGoGoCommand(editor string) *cobra.Command {
	var target string
	var env []string
	var overwrite bool
	var additionalArgs []string

	cmd := &cobra.Command{
		Use:   "add-go-go NAME PROFILE [ARGS...]",
		Short: "Add go-go-mcp server",
		Long: fmt.Sprintf("Add a new MCP server configuration that uses go-go-mcp server for %s.", editor),
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewStoreWithTarget(editor, target)
			if err != nil {
				return err
			}

			name := args[0]
			profile := args[1]
			cmdArgs := append([]string{"server", "start", "--profile", profile}, args[2:]...)
			cmdArgs = append(cmdArgs, additionalArgs...)

			// Parse environment variables
			envMap := make(map[string]string)
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format: %s (expected KEY=VALUE)", e)
				}
				envMap[parts[0]] = parts[1]
			}

			// Find the path to the mcp binary
			mcpPath, err := exec.LookPath("mcp")
			if err != nil {
				return fmt.Errorf("could not find mcp executable in PATH: %w", err)
			}

			commonServer := types.CommonServer{
				Name:    name,
				Command: mcpPath,
				Args:    cmdArgs,
				Env:     envMap,
				IsSSE:   false,
			}

			if err := store.AddServer(commonServer, overwrite); err != nil {
				return err
			}
			if err := store.Save(); err != nil {
				return err
			}

			// Print success message with configuration details
			action := "Added"
			if overwrite {
				action = "Updated"
			}
			fmt.Printf("Successfully %s go-go-mcp server '%s':\n", action, name)
			fmt.Printf("  Command: %s\n", mcpPath)
			fmt.Printf("  Profile: %s\n", profile)
			if len(cmdArgs) > 4 {
				fmt.Printf("  Additional Args: %v\n", cmdArgs[4:])
			}
			if len(envMap) > 0 {
				fmt.Printf("  Environment:\n")
				for k, v := range envMap {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			fmt.Printf("\nConfiguration saved to: %s\n", store.GetConfigPath())

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")
	cmd.Flags().StringArrayVar(&additionalArgs, "args", []string{}, "Additional arguments to pass to the server command")

	return cmd
}
