package config

import (
	"fmt"

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
