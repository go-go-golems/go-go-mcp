package cmds

import (
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/cmd/mcp-client/cmds/helpers"
	"github.com/spf13/cobra"
)

var (
	toolArgs string
)

// ToolsCmd handles the "tools" command group
var ToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Interact with tools",
	Long:  `List available tools and execute specific tools.`,
}

var listToolsCmd = &cobra.Command{
	Use:   "list",
	Short: "List available tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := helpers.CreateClient(cmd)
		if err != nil {
			return err
		}
		defer client.Close(cmd.Context())

		tools, cursor, err := client.ListTools(cmd.Context(), "")
		if err != nil {
			return err
		}

		for _, tool := range tools {
			fmt.Printf("Name: %s\n", tool.Name)
			fmt.Printf("Description: %s\n", tool.Description)
			fmt.Println()
		}

		if cursor != "" {
			fmt.Printf("Next cursor: %s\n", cursor)
		}

		return nil
	},
}

var callToolCmd = &cobra.Command{
	Use:   "call [tool-name]",
	Short: "Call a specific tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := helpers.CreateClient(cmd)
		if err != nil {
			return err
		}
		defer client.Close(cmd.Context())

		// Parse tool arguments
		toolArgMap := make(map[string]interface{})
		if toolArgs != "" {
			if err := json.Unmarshal([]byte(toolArgs), &toolArgMap); err != nil {
				return fmt.Errorf("invalid tool arguments JSON: %w", err)
			}
		}

		result, err := client.CallTool(cmd.Context(), args[0], toolArgMap)
		if err != nil {
			return err
		}

		// Pretty print the result
		for _, content := range result.Content {
			fmt.Printf("Type: %s\n", content.Type)
			if content.Type == "text" {
				fmt.Printf("Content:\n%s\n", content.Text)
			} else if content.Type == "image" {
				fmt.Printf("Image:\n%s\n", content.Data)
			} else if content.Type == "resource" {
				fmt.Printf("URI: %s\n", content.Resource.URI)
				fmt.Printf("MimeType: %s\n", content.Resource.MimeType)
			}
		}
		return nil
	},
}

func init() {
	ToolsCmd.AddCommand(listToolsCmd)
	ToolsCmd.AddCommand(callToolCmd)
	callToolCmd.Flags().StringVarP(&toolArgs, "args", "a", "", "Tool arguments as JSON string")
}
