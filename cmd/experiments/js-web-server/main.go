package main

import (
	"fmt"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	clay_profiles "github.com/go-go-golems/clay/pkg/cmds/profiles"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/cmd"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/mcp"
	"github.com/spf13/cobra"
)

func main() {
	// Initialize help system
	helpSystem := help.NewHelpSystem()

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "js-playground",
		Short: "JavaScript playground web server with Geppetto AI integration",
		Long:  "A JavaScript playground web server with SQLite integration and Geppetto AI capabilities",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromViper()
		},
	}

	// Set up help system for the root command
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Initialize Viper for configuration management
	if err := clay.InitViper("js-web-server", rootCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize viper: %v\n", err)
		os.Exit(1)
	}

	// Create Glazed commands with Geppetto integration

	// Serve command with Geppetto layers
	serveCmd, err := cmd.NewServeCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating serve command: %v\n", err)
		os.Exit(1)
	}

	// Build Cobra command with custom js-web-server middlewares and profile support
	serveCobraCmd, err := cmd.BuildCobraCommandWithServeMiddlewares(
		serveCmd,
		cli.WithProfileSettingsLayer(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building serve command: %v\n", err)
		os.Exit(1)
	}

	// Execute command
	executeCmd, err := cmd.NewExecuteCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating execute command: %v\n", err)
		os.Exit(1)
	}

	executeCobraCmd, err := cli.BuildCobraCommandFromCommand(executeCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building execute command: %v\n", err)
		os.Exit(1)
	}

	// Test command
	testCmd, err := cmd.NewTestCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating test command: %v\n", err)
		os.Exit(1)
	}

	testCobraCmd, err := cli.BuildCobraCommandFromCommand(testCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building test command: %v\n", err)
		os.Exit(1)
	}

	// Run Scripts command with Geppetto integration
	runScriptsCmd, err := cmd.NewRunScriptsCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating run-scripts command: %v\n", err)
		os.Exit(1)
	}

	runScriptsCobraCmd, err := cmd.BuildCobraCommandWithServeMiddlewares(
		runScriptsCmd,
		cli.WithProfileSettingsLayer(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building run-scripts command: %v\n", err)
		os.Exit(1)
	}

	// Add commands to root
	rootCmd.AddCommand(serveCobraCmd, executeCobraCmd, testCobraCmd, runScriptsCobraCmd)

	// Add profiles command for configuration management
	profilesCmd, err := clay_profiles.NewProfilesCommand("js-web-server", jsWebServerInitialProfilesContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing profiles command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(profilesCmd)

	// MCP command - expose JavaScript execution as MCP tool
	if err := mcp.AddMCPCommand(rootCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add MCP command: %v\n", err)
		os.Exit(1)
	}

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

// jsWebServerInitialProfilesContent provides the default YAML content for a new js-web-server profiles file.
func jsWebServerInitialProfilesContent() string {
	return `# JavaScript Web Server Profiles Configuration
#
# This file contains profile configurations for the JavaScript Web Server with Geppetto AI integration.
# Each profile can override layer parameters for different components (like AI models, databases, server settings).
# Profiles allow you to easily switch between different environments, model providers, or configurations.
#
# Profiles are selected using the --profile <profile-name> flag.
#
# Example profiles:

# Development profile with local settings
development:
  # AI Chat settings for development
  ai-chat:
    ai-engine: gpt-4o-mini
    ai-api-type: openai
    ai-temperature: 0.8
    ai-stream: true
  # OpenAI configuration for development
  openai-chat:
    openai-api-key: "[REDACTED:api-key]" # Replace with your key or use environment variable
  # Server settings for development
  default:
    port: "8080"
    admin-port: "9090"
    app-db: "dev-data.sqlite"
    system-db: "dev-system.sqlite"
    scripts: "./scripts"

# Production profile with optimized settings
production:
  # AI Chat settings for production
  ai-chat:
    ai-engine: gpt-4
    ai-api-type: openai
    ai-temperature: 0.7
    ai-max-response-tokens: 2000
    ai-cache-type: disk
    ai-cache-directory: "/var/cache/js-web-server"
  # OpenAI configuration for production
  openai-chat:
    openai-api-key: "[REDACTED:api-key]" # Use environment variable in production
  # Server settings for production
  default:
    port: "8080"
    admin-port: "9090"
    app-db: "/var/lib/js-web-server/data.sqlite"
    system-db: "/var/lib/js-web-server/system.sqlite"
    scripts: "/etc/js-web-server/scripts"

# Claude-based profile
claude-dev:
  # AI Chat settings using Claude
  ai-chat:
    ai-engine: claude-3-sonnet-20240229
    ai-api-type: claude
    ai-temperature: 0.6
    ai-stream: true
  # Claude configuration
  claude-chat:
    claude-api-key: "[REDACTED:api-key]" # Replace with your Anthropic API key
    claude-base-url: "https://api.anthropic.com/"
  # Server settings
  default:
    port: "8081"
    admin-port: "9091"
    app-db: "claude-data.sqlite"
    system-db: "claude-system.sqlite"

# Local LLM profile using Ollama
local-llm:
  # AI Chat settings for local models
  ai-chat:
    ai-engine: llama3:8b
    ai-api-type: ollama
    ai-temperature: 0.5
    ai-stream: true
  # Embeddings using local models
  embeddings:
    embeddings-type: ollama
    embeddings-engine: nomic-embed-text
    embeddings-dimensions: 768
  # Server settings for local development
  default:
    port: "8082"
    admin-port: "9092"
    app-db: "local-data.sqlite"
    system-db: "local-system.sqlite"
    scripts: "./scripts"

# Testing profile with minimal settings
testing:
  # AI Chat settings for testing
  ai-chat:
    ai-engine: gpt-3.5-turbo
    ai-api-type: openai
    ai-temperature: 0.0
    ai-max-response-tokens: 100
  # Server settings for testing
  default:
    port: "18080"
    admin-port: "19090"
    app-db: ":memory:"
    system-db: ":memory:"

#
# You can manage this file using the 'js-web-server profiles' commands:
# - list: List all profiles
# - get <profile> [layer] [key]: Get profile settings
# - set <profile> <layer> <key> <value>: Set a profile setting
# - delete <profile> [layer] [key]: Delete a profile, layer, or setting
# - edit: Open this file in your editor
# - init: Create this file if it doesn't exist
# - duplicate <source> <new>: Copy an existing profile
#
# Examples:
#   js-web-server --profile development serve
#   js-web-server --profile production serve --port 80
#   js-web-server --profile claude-dev serve
#   js-web-server --profile local-llm serve
#   js-web-server profiles list
#   js-web-server profiles get development ai-chat ai-engine
#   js-web-server profiles set testing ai-chat ai-temperature 0.1
`
}
