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

// NewMultiDisableCommand creates the disable command that supports multiple editors
func NewMultiDisableCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "disable EDITORS NAME",
		Short: "Disable server in multiple editors",
		Long: `Disable an MCP server in one or more editors without removing it.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

Examples:
  mcp editor config claude,cursor disable myserver
  mcp editor config amp disable myserver --target global`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse editors from first argument
			editors, err := ParseEditors(args[0])
			if err != nil {
				return err
			}

			name := args[1]

			// Execute for each editor
			stores, err := NewStoresWithTarget(editors, target)
			if err != nil {
				return err
			}

			var successCount, failCount int
			var results []string

			for _, editorStore := range stores {
				// Check if server exists
				_, exists, err := editorStore.Store.GetServer(name)
				if err != nil {
					results = append(results, fmt.Sprintf("❌ %s: error checking server: %v", editorStore.Editor, err))
					failCount++
					continue
				}
				if !exists {
					results = append(results, fmt.Sprintf("⚠️  %s: server '%s' not found", editorStore.Editor, name))
					failCount++
					continue
				}

				// Check if server is already disabled
				disabled, err := editorStore.Store.IsServerDisabled(name)
				if err != nil {
					results = append(results, fmt.Sprintf("❌ %s: error checking status: %v", editorStore.Editor, err))
					failCount++
					continue
				}
				if disabled {
					results = append(results, fmt.Sprintf("ℹ️  %s: server '%s' already disabled", editorStore.Editor, name))
					successCount++ // Count as success since desired state is achieved
					continue
				}

				if err := editorStore.Store.DisableServer(name); err != nil {
					results = append(results, fmt.Sprintf("❌ %s: %v", editorStore.Editor, err))
					failCount++
					continue
				}

				if err := editorStore.Store.Save(); err != nil {
					results = append(results, fmt.Sprintf("❌ %s: save failed: %v", editorStore.Editor, err))
					failCount++
					continue
				}

				results = append(results, fmt.Sprintf("✅ %s: disabled server '%s'", editorStore.Editor, name))
				successCount++
			}

			// Print results
			if len(editors) == 1 {
				// Single editor: use existing format for backwards compatibility
				if successCount > 0 {
					fmt.Printf("Successfully disabled MCP server '%s'\n", name)
				} else {
					return fmt.Errorf("failed to disable server: %s", results[0][2:]) // Remove emoji prefix
				}
			} else {
				// Multiple editors: show results for each
				fmt.Printf("Multi-editor disable results for server '%s':\n\n", name)
				for _, result := range results {
					fmt.Println(result)
				}
				fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failCount)
			}

			// Return error if all failed, but allow partial success
			if successCount == 0 {
				return fmt.Errorf("failed to disable server in any editor")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}
