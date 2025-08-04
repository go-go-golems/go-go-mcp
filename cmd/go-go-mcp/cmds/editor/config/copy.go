package config

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/core/configstore"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/spf13/cobra"
)

// NewCopyCommand creates the copy command for a specific editor
func NewCopyCommand(editor string) *cobra.Command {
	var target string
	var targetEditor string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "copy SOURCE_NAME DEST_NAME",
		Short: "Copy MCP server configuration",
		Long: fmt.Sprintf(`Copy an MCP server configuration to a new name in %s.

The copy preserves all server configuration including command, arguments, 
environment variables, URL, transport type, and enabled/disabled status.

Cross-editor copying:
Use --target-editor to copy from this editor to another editor.
Note: Some transport types may not be compatible between editors.
For example, Claude only supports stdio transport.

If a server with the destination name already exists, the command will fail 
unless --overwrite is specified.`, editor),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceStore, err := NewStoreWithTarget(editor, target)
			if err != nil {
				return err
			}

			sourceName := args[0]
			destName := args[1]

			// Get the source server
			sourceServer, exists, err := sourceStore.GetServer(sourceName)
			if err != nil {
				return fmt.Errorf("failed to get source server: %w", err)
			}
			if !exists {
				return fmt.Errorf("source server '%s' not found", sourceName)
			}

			// Check if source server is disabled
			isDisabled, err := sourceStore.IsServerDisabled(sourceName)
			if err != nil {
				return fmt.Errorf("failed to check source server status: %w", err)
			}

			// Determine destination store
			var destStore configstore.Store
			var destEditorName string
			if targetEditor != "" {
				// Cross-editor copy
				destStore, err = NewStoreWithTarget(targetEditor, target)
				if err != nil {
					return fmt.Errorf("failed to create destination store: %w", err)
				}
				destEditorName = targetEditor

				// Validate transport compatibility for cross-editor copy
				if err := validateTransportCompatibility(editor, targetEditor, sourceServer); err != nil {
					return err
				}
			} else {
				// Same editor copy
				destStore = sourceStore
				destEditorName = editor
			}

			// Create destination server with new name
			destServer := sourceServer
			destServer.Name = destName

			// Add the server to destination
			if err := destStore.AddServer(destServer, overwrite); err != nil {
				return fmt.Errorf("failed to add destination server: %w", err)
			}

			// If source was disabled, disable the destination as well
			if isDisabled {
				if err := destStore.DisableServer(destName); err != nil {
					return fmt.Errorf("failed to disable destination server: %w", err)
				}
			}

			// Save the destination store
			if err := destStore.Save(); err != nil {
				return fmt.Errorf("failed to save destination configuration: %w", err)
			}

			// Print success message
			action := "Copied"
			if overwrite {
				action = "Copied (overwritten)"
			}

			if targetEditor != "" {
				fmt.Printf("Successfully %s MCP server '%s' from %s to %s as '%s':\n",
					action, sourceName, editor, destEditorName, destName)
			} else {
				fmt.Printf("Successfully %s MCP server '%s' to '%s':\n",
					action, sourceName, destName)
			}

			// Show source details
			fmt.Printf("  Source (%s):\n", sourceName)
			printServerDetails(sourceServer, "    ")

			// Show destination details
			fmt.Printf("  Destination (%s):\n", destName)
			printServerDetails(destServer, "    ")

			// Show status
			statusText := "enabled"
			if isDisabled {
				statusText = "disabled"
			}
			fmt.Printf("  Status: %s\n", statusText)

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")
	cmd.Flags().StringVar(&targetEditor, "target-editor", "", "Copy to a different editor (claude, cursor, ampcode, amp, crush)")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")

	return cmd
}

// validateTransportCompatibility checks if the source server's transport is compatible with the target editor
func validateTransportCompatibility(sourceEditor, targetEditor string, server types.CommonServer) error {
	// Claude only supports stdio transport
	if targetEditor == "claude" && server.URL != "" {
		return fmt.Errorf("claude only supports stdio transport, but source server uses URL-based transport (%s)", server.URL)
	}

	// Add more compatibility checks as needed for other editors
	return nil
}

// printServerDetails prints the details of a server configuration
func printServerDetails(server types.CommonServer, indent string) {
	if server.URL != "" {
		transportType := "HTTP"
		if server.IsSSE {
			transportType = "SSE"
		}
		fmt.Printf("%sTransport: %s\n", indent, transportType)
		fmt.Printf("%sURL: %s\n", indent, server.URL)
	} else {
		fmt.Printf("%sTransport: stdio\n", indent)
		fmt.Printf("%sCommand: %s\n", indent, server.Command)
		if len(server.Args) > 0 {
			fmt.Printf("%sArgs: %v\n", indent, server.Args)
		}
	}

	if len(server.Env) > 0 {
		fmt.Printf("%sEnvironment:\n", indent)
		for k, v := range server.Env {
			fmt.Printf("%s  %s: %s\n", indent, k, v)
		}
	}
}
