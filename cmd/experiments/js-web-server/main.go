package main

import (
	"fmt"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/cmd"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/mcp"
	pinocchio_cmds "github.com/go-go-golems/pinocchio/pkg/cmds"
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

	// Build Cobra command with Geppetto middlewares
	serveCobraCmd, err := pinocchio_cmds.BuildCobraCommandWithGeppettoMiddlewares(
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

	// Add commands to root
	rootCmd.AddCommand(serveCobraCmd, executeCobraCmd, testCobraCmd)

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


