package config

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/spf13/cobra"
)

// NewAddCommand creates the add command for a specific editor
func NewAddCommand(editor string) *cobra.Command {
	var target string
	var env []string
	var overwrite bool
	var url string

	cmd := &cobra.Command{
		Use:   "add NAME COMMAND [ARGS...]",
		Short: "Add MCP server",
		Long: fmt.Sprintf(`Add a new MCP server configuration to %s.

If a server with the same name already exists, the command will fail unless --overwrite is specified.

Transport types:
- Stdio: Provide COMMAND [ARGS...] 
- HTTP/SSE: Provide --url with appropriate URL

URLs containing 'events' or 'sse' in the path are detected as SSE transport.`, editor),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewStoreWithTarget(editor, target)
			if err != nil {
				return err
			}

			name := args[0]
			var command string
			var cmdArgs []string

			// Handle different transport types
			if url != "" {
				// URL-based transport (HTTP/SSE)
				if len(args) > 1 {
					return fmt.Errorf("cannot specify both command and URL")
				}
				command = ""
				cmdArgs = nil
			} else {
				// Stdio transport
				if len(args) < 2 {
					return fmt.Errorf("command required for stdio transport")
				}
				command = args[1]
				cmdArgs = args[2:]
			}

			// Parse environment variables
			envMap := make(map[string]string)
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format: %s (expected KEY=VALUE)", e)
				}
				envMap[parts[0]] = parts[1]
			}

			// Auto-detect transport type
			isSSE := false
			if url != "" {
				// Simple heuristic: URLs containing "events" or "sse" are SSE
				urlLower := strings.ToLower(url)
				isSSE = strings.Contains(urlLower, "events") || strings.Contains(urlLower, "sse")
			}

			commonServer := types.CommonServer{
				Name:    name,
				Command: command,
				Args:    cmdArgs,
				Env:     envMap,
				URL:     url,
				IsSSE:   isSSE,
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
			fmt.Printf("Successfully %s MCP server '%s':\n", action, name)

			if url != "" {
				transportType := "HTTP"
				if isSSE {
					transportType = "SSE"
				}
				fmt.Printf("  Transport: %s\n", transportType)
				fmt.Printf("  URL: %s\n", url)
			} else {
				fmt.Printf("  Transport: stdio\n")
				fmt.Printf("  Command: %s\n", command)
				if len(cmdArgs) > 0 {
					fmt.Printf("  Args: %v\n", cmdArgs)
				}
			}

			if len(envMap) > 0 {
				fmt.Printf("  Environment:\n")
				for k, v := range envMap {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")
	cmd.Flags().StringVar(&url, "url", "", "URL for HTTP/SSE servers")

	return cmd
}

// NewMultiAddCommand creates the add command that supports multiple editors
func NewMultiAddCommand() *cobra.Command {
	var target string
	var env []string
	var overwrite bool
	var url string

	cmd := &cobra.Command{
		Use:   "add EDITORS NAME COMMAND [ARGS...]",
		Short: "Add MCP server to multiple editors",
		Long: `Add a new MCP server configuration to one or more editors.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

If a server with the same name already exists, the command will fail unless --overwrite is specified.

Transport types:
- Stdio: Provide COMMAND [ARGS...] 
- HTTP/SSE: Provide --url with appropriate URL

URLs containing 'events' or 'sse' in the path are detected as SSE transport.

Examples:
  mcp editor config claude,cursor add myserver /path/to/command
  mcp editor config amp add myserver --url http://localhost:8080
  mcp editor config claude,cursor,amp add myserver node script.js`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse editors from first argument
			editors, err := ParseEditors(args[0])
			if err != nil {
				return err
			}

			name := args[1]
			var command string
			var cmdArgs []string

			// Handle different transport types
			if url != "" {
				// URL-based transport (HTTP/SSE)
				if len(args) > 2 {
					return fmt.Errorf("cannot specify both command and URL")
				}
				command = ""
				cmdArgs = nil
			} else {
				// Stdio transport
				if len(args) < 3 {
					return fmt.Errorf("command required for stdio transport")
				}
				command = args[2]
				cmdArgs = args[3:]
			}

			// Parse environment variables
			envMap := make(map[string]string)
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format: %s (expected KEY=VALUE)", e)
				}
				envMap[parts[0]] = parts[1]
			}

			// Auto-detect transport type
			isSSE := false
			if url != "" {
				// Simple heuristic: URLs containing "events" or "sse" are SSE
				urlLower := strings.ToLower(url)
				isSSE = strings.Contains(urlLower, "events") || strings.Contains(urlLower, "sse")
			}

			commonServer := types.CommonServer{
				Name:    name,
				Command: command,
				Args:    cmdArgs,
				Env:     envMap,
				URL:     url,
				IsSSE:   isSSE,
			}

			// Execute for each editor
			stores, err := NewStoresWithTarget(editors, target)
			if err != nil {
				return err
			}

			var successCount, failCount int
			var results []string

			for _, editorStore := range stores {
				err := editorStore.Store.AddServer(commonServer, overwrite)
				if err != nil {
					results = append(results, fmt.Sprintf("❌ %s: %v", editorStore.Editor, err))
					failCount++
					continue
				}

				err = editorStore.Store.Save()
				if err != nil {
					results = append(results, fmt.Sprintf("❌ %s: save failed: %v", editorStore.Editor, err))
					failCount++
					continue
				}

				action := "Added"
				if overwrite {
					action = "Updated"
				}
				results = append(results, fmt.Sprintf("✅ %s: %s server '%s'", editorStore.Editor, action, name))
				successCount++
			}

			// Print results
			if len(editors) == 1 {
				// Single editor: use existing format for backwards compatibility
				if successCount > 0 {
					action := "Added"
					if overwrite {
						action = "Updated"
					}
					fmt.Printf("Successfully %s MCP server '%s':\n", action, name)
					printServerDetails(commonServer, "  ")
				} else {
					return fmt.Errorf("failed to add server: %s", results[0][2:]) // Remove emoji prefix
				}
			} else {
				// Multiple editors: show results for each
				fmt.Printf("Multi-editor add results for server '%s':\n\n", name)
				for _, result := range results {
					fmt.Println(result)
				}
				fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failCount)
				if successCount > 0 {
					fmt.Println("\nServer configuration:")
					printServerDetails(commonServer, "  ")
				}
			}

			// Return error if all failed, but allow partial success
			if successCount == 0 {
				return fmt.Errorf("failed to add server to any editor")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")
	cmd.Flags().StringVar(&url, "url", "", "URL for HTTP/SSE servers")

	return cmd
}
