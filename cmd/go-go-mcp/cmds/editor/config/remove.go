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

// NewMultiRemoveCommand creates the remove command that supports multiple editors
func NewMultiRemoveCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "remove EDITORS NAME",
		Short: "Remove MCP server from multiple editors",
		Long: `Remove an MCP server configuration from one or more editors.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

Examples:
  mcp editor config claude,cursor remove myserver
  mcp editor config amp remove myserver --target global`,
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
				// Check if server exists before removing
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

				if err := editorStore.Store.RemoveServer(name); err != nil {
					results = append(results, fmt.Sprintf("❌ %s: %v", editorStore.Editor, err))
					failCount++
					continue
				}

				if err := editorStore.Store.Save(); err != nil {
					results = append(results, fmt.Sprintf("❌ %s: save failed: %v", editorStore.Editor, err))
					failCount++
					continue
				}

				results = append(results, fmt.Sprintf("✅ %s: removed server '%s'", editorStore.Editor, name))
				successCount++
			}

			// Print results
			if len(editors) == 1 {
				// Single editor: use existing format for backwards compatibility
				if successCount > 0 {
					fmt.Printf("Successfully removed MCP server '%s'\n", name)
				} else {
					return fmt.Errorf("failed to remove server: %s", results[0][2:]) // Remove emoji prefix
				}
			} else {
				// Multiple editors: show results for each
				fmt.Printf("Multi-editor remove results for server '%s':\n\n", name)
				for _, result := range results {
					fmt.Println(result)
				}
				fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failCount)
			}

			// Return error if all failed, but allow partial success
			if successCount == 0 {
				return fmt.Errorf("failed to remove server from any editor")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}
