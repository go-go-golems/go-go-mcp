package main

import (
	"fmt"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"

	// Import the cmds package
	"github.com/go-go-golems/go-go-mcp/cmd/mcp-client/cmds"
)

var (
	// Version information
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "none"

	// Command flags
	debug      bool
	logLevel   string
	withCaller bool

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
			// reinitialize the logger because we can now parse --log-level and co
			// from the command line flag
			err := clay.InitLogger()
			cobra.CheckErr(err)

		},
	}

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("mcp", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	// Add persistent flags
	rootCmd.PersistentFlags().StringP("transport", "t", "command", "Transport type (command or sse)")
	rootCmd.PersistentFlags().StringP("server", "s", "http://localhost:3001", "Server URL for SSE transport")
	rootCmd.PersistentFlags().StringSliceP("command", "c", []string{"mcp-server", "start", "--transport", "stdio"}, "Command and arguments for command transport (first argument is the command)")

	// Add all command groups to root command
	// Remove existing subcommand additions
	// rootCmd.AddCommand(promptsCmd)
	// rootCmd.AddCommand(toolsCmd)
	// rootCmd.AddCommand(resourcesCmd)
	// rootCmd.AddCommand(versionCmd)

	// Add subcommands from cmds package
	rootCmd.AddCommand(cmds.PromptsCmd)
	rootCmd.AddCommand(cmds.ToolsCmd)
	rootCmd.AddCommand(cmds.ResourcesCmd)
	rootCmd.AddCommand(cmds.VersionCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
