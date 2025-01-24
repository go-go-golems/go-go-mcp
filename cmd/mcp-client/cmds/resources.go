package cmds

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/cmd/mcp-client/cmds/helpers"
	"github.com/spf13/cobra"
)

// ResourcesCmd handles the "resources" command group
var ResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Interact with resources",
	Long:  `List available resources and read specific resources.`,
}

var listResourcesCmd = &cobra.Command{
	Use:   "list",
	Short: "List available resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := helpers.CreateClient(cmd)
		if err != nil {
			return err
		}
		defer client.Close(cmd.Context())

		resources, cursor, err := client.ListResources(cmd.Context(), "")
		if err != nil {
			return err
		}

		for _, resource := range resources {
			fmt.Printf("URI: %s\n", resource.URI)
			fmt.Printf("Name: %s\n", resource.Name)
			fmt.Printf("Description: %s\n", resource.Description)
			fmt.Printf("MimeType: %s\n", resource.MimeType)
			fmt.Println()
		}

		if cursor != "" {
			fmt.Printf("Next cursor: %s\n", cursor)
		}

		return nil
	},
}

var readResourceCmd = &cobra.Command{
	Use:   "read [resource-uri]",
	Short: "Read a specific resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := helpers.CreateClient(cmd)
		if err != nil {
			return err
		}
		defer client.Close(cmd.Context())

		content, err := client.ReadResource(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		fmt.Printf("URI: %s\n", content.URI)
		fmt.Printf("MimeType: %s\n", content.MimeType)
		fmt.Printf("Content:\n%s\n", content.Text)
		return nil
	},
}

func init() {
	ResourcesCmd.AddCommand(listResourcesCmd)
	ResourcesCmd.AddCommand(readResourceCmd)
}
