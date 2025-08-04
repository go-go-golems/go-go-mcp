package config

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/core/configstore"
	"github.com/spf13/cobra"
)

// NewListCommand creates the list command for a specific editor
func NewListCommand(editor string) *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List servers",
		Long:  fmt.Sprintf("List all configured MCP servers for %s.", editor),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewStoreWithTarget(editor, target)
			if err != nil {
				return err
			}

			servers, err := store.ListServers()
			if err != nil {
				return err
			}

			if len(servers) == 0 {
				fmt.Printf("No MCP servers configured for %s.\n", editor)
				return nil
			}

			fmt.Printf("Configured MCP servers for %s:\n\n", editor)
			for name, server := range servers {
				// Check if server is disabled
				disabled, err := store.IsServerDisabled(name)
				if err != nil {
					return err
				}

				disabledStatus := ""
				if disabled {
					disabledStatus = " (disabled)"
				}

				fmt.Printf("%s%s:\n", name, disabledStatus)

				if server.URL != "" {
					// URL-based transport
					transportType := "HTTP"
					if server.IsSSE {
						transportType = "SSE"
					}
					fmt.Printf("  Transport: %s\n", transportType)
					fmt.Printf("  URL: %s\n", server.URL)
				} else {
					// Stdio transport
					fmt.Printf("  Transport: stdio\n")
					fmt.Printf("  Command: %s\n", server.Command)
					if len(server.Args) > 0 {
						fmt.Printf("  Args: %v\n", server.Args)
					}
				}

				if len(server.Env) > 0 {
					fmt.Printf("  Environment:\n")
					for k, v := range server.Env {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}

// NewMultiListCommand creates the list command that supports multiple editors
func NewMultiListCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "list EDITORS",
		Short: "List servers from multiple editors",
		Long: `List all configured MCP servers for one or more editors.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

Examples:
  mcp editor config claude,cursor list
  mcp editor config amp list --target global`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse editors from first argument
			editors, err := ParseEditors(args[0])
			if err != nil {
				return err
			}

			stores, err := NewStoresWithTarget(editors, target)
			if err != nil {
				return err
			}

			if len(editors) == 1 {
				// Single editor: use existing format for backwards compatibility
				return runSingleEditorList(stores[0].Editor, stores[0].Store)
			}

			// Multiple editors: show grouped results
			var hasAnyServers bool
			for _, editorStore := range stores {
				servers, err := editorStore.Store.ListServers()
				if err != nil {
					fmt.Printf("❌ %s: error listing servers: %v\n\n", editorStore.Editor, err)
					continue
				}

				fmt.Printf("📝 %s:\n", editorStore.Editor)
				if len(servers) == 0 {
					fmt.Printf("  No MCP servers configured.\n\n")
					continue
				}

				hasAnyServers = true
				for name, server := range servers {
					// Check if server is disabled
					disabled, err := editorStore.Store.IsServerDisabled(name)
					if err != nil {
						fmt.Printf("  ❌ %s: error checking status: %v\n", name, err)
						continue
					}

					disabledStatus := ""
					if disabled {
						disabledStatus = " (disabled)"
					}

					fmt.Printf("  %s%s:\n", name, disabledStatus)

					if server.URL != "" {
						// URL-based transport
						transportType := "HTTP"
						if server.IsSSE {
							transportType = "SSE"
						}
						fmt.Printf("    Transport: %s\n", transportType)
						fmt.Printf("    URL: %s\n", server.URL)
					} else {
						// Stdio transport
						fmt.Printf("    Transport: stdio\n")
						fmt.Printf("    Command: %s\n", server.Command)
						if len(server.Args) > 0 {
							fmt.Printf("    Args: %v\n", server.Args)
						}
					}

					if len(server.Env) > 0 {
						fmt.Printf("    Environment:\n")
						for k, v := range server.Env {
							fmt.Printf("      %s: %s\n", k, v)
						}
					}
					fmt.Println()
				}
			}

			if !hasAnyServers {
				fmt.Println("No MCP servers configured in any of the specified editors.")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")

	return cmd
}

// runSingleEditorList runs the list command for a single editor (backwards compatibility)
func runSingleEditorList(editor string, store configstore.Store) error {
	servers, err := store.ListServers()
	if err != nil {
		return err
	}

	if len(servers) == 0 {
		fmt.Printf("No MCP servers configured for %s.\n", editor)
		return nil
	}

	fmt.Printf("Configured MCP servers for %s:\n\n", editor)
	for name, server := range servers {
		// Check if server is disabled
		disabled, err := store.IsServerDisabled(name)
		if err != nil {
			return err
		}

		disabledStatus := ""
		if disabled {
			disabledStatus = " (disabled)"
		}

		fmt.Printf("%s%s:\n", name, disabledStatus)

		if server.URL != "" {
			// URL-based transport
			transportType := "HTTP"
			if server.IsSSE {
				transportType = "SSE"
			}
			fmt.Printf("  Transport: %s\n", transportType)
			fmt.Printf("  URL: %s\n", server.URL)
		} else {
			// Stdio transport
			fmt.Printf("  Transport: stdio\n")
			fmt.Printf("  Command: %s\n", server.Command)
			if len(server.Args) > 0 {
				fmt.Printf("  Args: %v\n", server.Args)
			}
		}

		if len(server.Env) > 0 {
			fmt.Printf("  Environment:\n")
			for k, v := range server.Env {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
		fmt.Println()
	}

	return nil
}
