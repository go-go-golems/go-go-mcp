package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/go-go-mcp/pkg/cmds"

	clay "github.com/go-go-golems/clay/pkg"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/help"
	server_cmds "github.com/go-go-golems/go-go-mcp/cmd/mcp-server/cmds"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
)

var (
	// Version information
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "none"

	// Command flags
	transport  string
	port       int
	debug      bool
	logLevel   string
	withCaller bool
)

var rootCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "MCP server implementation in Go",
	Long: `A Model Context Protocol (MCP) server implementation in Go.
Supports both stdio and SSE transports for client-server communication.

The server implements the Model Context Protocol (MCP) specification,
providing a framework for building MCP servers and clients.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// reinitialize the logger because we can now parse --log-level and co
		// from the command line flag
		err := clay.InitLogger()
		cobra.CheckErr(err)
	},
}

var runCommandCmd = &cobra.Command{
	Use:   "run-command [file]",
	Short: "Run a command from a YAML file",
	Long: `Run a shell command or script defined in a YAML file.
The file should contain a shell command definition with flags, arguments,
and either a command list or shell script to execute.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.Errorf("run-command must be called with a YAML file as first argument")
	},
}

func initRootCmd() (*help.HelpSystem, error) {
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("mcp", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	rootCmd.AddCommand(runCommandCmd)

	// Create and add start command
	startCmd, err := server_cmds.NewStartCommand()
	cobra.CheckErr(err)
	cobraStartCmd, err := cli.BuildCobraCommandFromBareCommand(startCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraStartCmd)

	// Create and add schema command
	schemaCmd, err := server_cmds.NewSchemaCommand()
	cobra.CheckErr(err)
	cobraSchemaCmd, err := cli.BuildCobraCommandFromWriterCommand(schemaCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraSchemaCmd)

	bridgeCmd := server_cmds.NewBridgeCommand(log.Logger)
	rootCmd.AddCommand(bridgeCmd)

	// Add version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mcp-server version %s\n", Version)
			fmt.Printf("  Build time: %s\n", BuildTime)
			fmt.Printf("  Git commit: %s\n", GitCommit)
		},
	}
	rootCmd.AddCommand(versionCmd)

	return helpSystem, nil
}

func main() {
	// first, check if the args are "run-command file.yaml",
	// because we need to load the file and then run the command itself.
	// we need to do this before cobra, because we don't know which flags to load yet
	if len(os.Args) >= 3 && os.Args[1] == "run-command" && os.Args[2] != "--help" {
		// load the command
		loader := &cmds.ShellCommandLoader{}
		fs_, filePath, err := loaders.FileNameToFsFilePath(os.Args[2])
		if err != nil {
			fmt.Printf("Could not get absolute path: %v\n", err)
			os.Exit(1)
		}
		cmds_, err := loader.LoadCommands(fs_, filePath, []glazed_cmds.CommandDescriptionOption{}, []alias.Option{})
		if err != nil {
			fmt.Printf("Could not load command: %v\n", err)
			os.Exit(1)
		}
		if len(cmds_) != 1 {
			fmt.Printf("Expected exactly one command, got %d", len(cmds_))
			os.Exit(1)
		}

		writerCommand, ok := cmds_[0].(glazed_cmds.WriterCommand)
		if !ok {
			fmt.Printf("Expected WriterCommand, got %T", cmds_[0])
			os.Exit(1)
		}

		cobraCommand, err := cli.BuildCobraCommandFromWriterCommand(writerCommand)
		if err != nil {
			fmt.Printf("Could not build cobra command: %v\n", err)
			os.Exit(1)
		}

		helpSystem, err := initRootCmd()
		cobra.CheckErr(err)

		helpSystem.SetupCobraRootCommand(cobraCommand)

		rootCmd.AddCommand(cobraCommand)
		restArgs := os.Args[3:]
		os.Args = append([]string{os.Args[0], cobraCommand.Use}, restArgs...)
	} else {
		_, err := initRootCmd()
		cobra.CheckErr(err)

	}

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
