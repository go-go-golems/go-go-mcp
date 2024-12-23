package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/go-mcp/pkg/client"
	"github.com/go-go-golems/go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	// Version information
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "none"

	// Command flags
	transport string
	serverURL string
	debug     bool

	// Operation flags
	promptArgs string
	toolArgs   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mcp-client",
		Short: "MCP client CLI",
		Long: `A Model Context Protocol (MCP) client CLI implementation.
Supports both stdio and SSE transports for client-server communication.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVarP(&transport, "transport", "t", "stdio", "Transport type (stdio or sse)")
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "http://localhost:8000", "Server URL for SSE transport")

	// Prompts command group
	promptsCmd := &cobra.Command{
		Use:   "prompts",
		Short: "Interact with prompts",
		Long:  `List available prompts and execute specific prompts.`,
	}

	listPromptsCmd := &cobra.Command{
		Use:   "list",
		Short: "List available prompts",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			defer client.Close()

			prompts, cursor, err := client.ListPrompts("")
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

	executePromptCmd := &cobra.Command{
		Use:   "execute [prompt-name]",
		Short: "Execute a specific prompt",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			defer client.Close()

			// Parse prompt arguments
			promptArgMap := make(map[string]string)
			if promptArgs != "" {
				if err := json.Unmarshal([]byte(promptArgs), &promptArgMap); err != nil {
					return fmt.Errorf("invalid prompt arguments JSON: %w", err)
				}
			}

			message, err := client.GetPrompt(args[0], promptArgMap)
			if err != nil {
				return err
			}

			// Pretty print the response
			fmt.Printf("Role: %s\n", message.Role)
			fmt.Printf("Content: %s\n", message.Content.Text)
			return nil
		},
	}
	executePromptCmd.Flags().StringVarP(&promptArgs, "args", "a", "", "Prompt arguments as JSON string")

	// Tools command group
	toolsCmd := &cobra.Command{
		Use:   "tools",
		Short: "Interact with tools",
		Long:  `List available tools and execute specific tools.`,
	}

	listToolsCmd := &cobra.Command{
		Use:   "list",
		Short: "List available tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tools, cursor, err := client.ListTools("")
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

	callToolCmd := &cobra.Command{
		Use:   "call [tool-name]",
		Short: "Call a specific tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			defer client.Close()

			// Parse tool arguments
			toolArgMap := make(map[string]interface{})
			if toolArgs != "" {
				if err := json.Unmarshal([]byte(toolArgs), &toolArgMap); err != nil {
					return fmt.Errorf("invalid tool arguments JSON: %w", err)
				}
			}

			result, err := client.CallTool(args[0], toolArgMap)
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
	callToolCmd.Flags().StringVarP(&toolArgs, "args", "a", "", "Tool arguments as JSON string")

	// Resources command group
	resourcesCmd := &cobra.Command{
		Use:   "resources",
		Short: "Interact with resources",
		Long:  `List available resources and read specific resources.`,
	}

	listResourcesCmd := &cobra.Command{
		Use:   "list",
		Short: "List available resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			defer client.Close()

			resources, cursor, err := client.ListResources("")
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

	readResourceCmd := &cobra.Command{
		Use:   "read [resource-uri]",
		Short: "Read a specific resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			defer client.Close()

			content, err := client.ReadResource(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("URI: %s\n", content.URI)
			fmt.Printf("MimeType: %s\n", content.MimeType)
			fmt.Printf("Content:\n%s\n", content.Text)
			return nil
		},
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mcp-client version %s\n", Version)
			fmt.Printf("  Build time: %s\n", BuildTime)
			fmt.Printf("  Git commit: %s\n", GitCommit)
		},
	}

	// Add subcommands to prompts command
	promptsCmd.AddCommand(listPromptsCmd)
	promptsCmd.AddCommand(executePromptCmd)

	// Add subcommands to tools command
	toolsCmd.AddCommand(listToolsCmd)
	toolsCmd.AddCommand(callToolCmd)

	// Add subcommands to resources command
	resourcesCmd.AddCommand(listResourcesCmd)
	resourcesCmd.AddCommand(readResourceCmd)

	// Add all command groups to root command
	rootCmd.AddCommand(promptsCmd)
	rootCmd.AddCommand(toolsCmd)
	rootCmd.AddCommand(resourcesCmd)
	rootCmd.AddCommand(versionCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createClient() (*client.Client, error) {
	// Use ConsoleWriter for colored output
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	var t client.Transport
	var err error

	switch transport {
	case "stdio":
		t = client.NewStdioTransport()
	case "sse":
		t = client.NewSSETransport(serverURL)
	default:
		return nil, fmt.Errorf("invalid transport type: %s", transport)
	}

	// Create and initialize client
	c := client.NewClient(logger, t)
	err = c.Initialize(protocol.ClientCapabilities{
		Sampling: &protocol.SamplingCapability{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	return c, nil
}
