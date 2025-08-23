package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/go-go-mcp/pkg/cmds"
	"github.com/go-go-golems/go-go-mcp/pkg/doc"

	clay "github.com/go-go-golems/clay/pkg"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	helpCmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	mcp_cmds "github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds"
	server_cmds "github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/server"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
)

var (
	// Version information
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP client and server implementation in Go",
	Long: `A Model Context Protocol (MCP) client and server implementation in Go.
Supports both stdio and SSE transports for client-server communication.

The server implements the Model Context Protocol (MCP) specification,
providing a framework for building MCP servers and clients.`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := logging.InitLoggerFromViper()
		if err != nil {
			return err
		}
		return nil
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

	err := doc.AddDocToHelpSystem(helpSystem)
	if err != nil {
		return nil, err
	}

	// Set up help system with UI support
	helpCmd.SetupCobraRootCommand(helpSystem, rootCmd)

	err = clay.InitViper("mcp", rootCmd)
	if err != nil {
		return nil, err
	}

	// Initialize commands
	rootCmd.AddCommand(mcp_cmds.ClientCmd)

	err = mcp_cmds.InitClientCommand(helpSystem)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize client command")
	}

	rootCmd.AddCommand(server_cmds.ServerCmd)

	rootCmd.AddCommand(runCommandCmd)

	// Create and add schema command
	schemaCmd, err := mcp_cmds.NewSchemaCommand()
	cobra.CheckErr(err)
	cobraSchemaCmd, err := cli.BuildCobraCommandFromWriterCommand(schemaCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraSchemaCmd)

	// bridge command removed

	// Add config command group
	configCmd := mcp_cmds.NewConfigGroupCommand()
	rootCmd.AddCommand(configCmd)

	// Add Claude config command group
	claudeConfigCmd := mcp_cmds.NewClaudeConfigCommand()
	rootCmd.AddCommand(claudeConfigCmd)

	// Add Cursor config command group
	cursorConfigCmd := mcp_cmds.NewCursorConfigCommand()
	rootCmd.AddCommand(cursorConfigCmd)

	// Add UI command
	rootCmd.AddCommand(mcp_cmds.NewUICommand())

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

		// Set up help system for the dynamically created command
		helpCmd.SetupCobraRootCommand(helpSystem, cobraCommand)

		rootCmd.AddCommand(cobraCommand)
		restArgs := os.Args[3:]
		os.Args = append([]string{os.Args[0], cobraCommand.Use}, restArgs...)
	} else {
		_, err := initRootCmd()
		cobra.CheckErr(err)
	}

	// Handle interrupts
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Execute
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("Error executing root command")
	}
}
