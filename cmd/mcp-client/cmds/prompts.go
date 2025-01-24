package cmds

import (
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/cmd/mcp-client/cmds/helpers"
	"github.com/spf13/cobra"
)

var (
	promptArgs string
)

// PromptsCmd handles the "prompts" command group
var PromptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "Interact with prompts",
	Long:  `List available prompts and execute specific prompts.`,
}

var listPromptsCmd = &cobra.Command{
	Use:   "list",
	Short: "List available prompts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := helpers.CreateClient(cmd)
		if err != nil {
			return err
		}
		defer client.Close(cmd.Context())

		prompts, cursor, err := client.ListPrompts(cmd.Context(), "")
		if err != nil {
			return err
		}

		for _, prompt := range prompts {
			fmt.Printf("Name: %s\n", prompt.Name)
			fmt.Printf("Description: %s\n", prompt.Description)
			fmt.Printf("Arguments:\n")
			for _, arg := range prompt.Arguments {
				fmt.Printf("  - %s (required: %v): %s\n",
					arg.Name, arg.Required, arg.Description)
			}
			fmt.Println()
		}

		if cursor != "" {
			fmt.Printf("Next cursor: %s\n", cursor)
		}

		return nil
	},
}

var executePromptCmd = &cobra.Command{
	Use:   "execute [prompt-name]",
	Short: "Execute a specific prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := helpers.CreateClient(cmd)
		if err != nil {
			return err
		}
		defer client.Close(cmd.Context())

		// Parse prompt arguments
		promptArgMap := make(map[string]string)
		if promptArgs != "" {
			if err := json.Unmarshal([]byte(promptArgs), &promptArgMap); err != nil {
				return fmt.Errorf("invalid prompt arguments JSON: %w", err)
			}
		}

		message, err := client.GetPrompt(cmd.Context(), args[0], promptArgMap)
		if err != nil {
			return err
		}

		// Pretty print the response
		fmt.Printf("Role: %s\n", message.Role)
		fmt.Printf("Content: %s\n", message.Content.Text)
		return nil
	},
}

func init() {
	PromptsCmd.AddCommand(listPromptsCmd)
	PromptsCmd.AddCommand(executePromptCmd)
	executePromptCmd.Flags().StringVarP(&promptArgs, "args", "a", "", "Prompt arguments as JSON string")
}
