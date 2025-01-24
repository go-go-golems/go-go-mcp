package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/pkg/cmds"
	"github.com/go-go-golems/go-go-mcp/pkg/cmds/shell"
	"github.com/go-go-golems/go-go-mcp/pkg/prompts"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/resources"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/examples"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/examples/cursor"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"

	clay "github.com/go-go-golems/clay/pkg"
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

// Add after the runCommandCmd definition:

var schemaCmd = &cobra.Command{
	Use:   "schema [file]",
	Short: "Output JSON schema for a command YAML file",
	Long: `Generate and output a JSON schema representation of a shell command YAML file.
This schema can be used for LLM tool calling definitions or command validation.

Example:
  mcp-server schema ./commands/my-command.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the command
		loader := &shell.ShellCommandLoader{}
		fs_, filePath, err := loaders.FileNameToFsFilePath(args[0])
		if err != nil {
			return fmt.Errorf("could not get absolute path: %w", err)
		}
		
		cmds_, err := loader.LoadCommands(fs_, filePath, []glazed_cmds.CommandDescriptionOption{}, []alias.Option{})
		if err != nil {
			return fmt.Errorf("could not load command: %w", err)
		}
		if len(cmds_) != 1 {
			return fmt.Errorf("expected exactly one command, got %d", len(cmds_))
		}

		shellCmd, ok := cmds_[0].(*cmds.ShellCommand)
		if !ok {
			return fmt.Errorf("expected ShellCommand, got %T", cmds_[0])
		}

		// Convert to JSON schema
		schema, err := shellCmd.ToJsonSchema()
		if err != nil {
			return fmt.Errorf("could not convert to JSON schema: %w", err)
		}

		// Output as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(schema); err != nil {
			return fmt.Errorf("could not encode JSON schema: %w", err)
		}

		return nil
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
	rootCmd.AddCommand(schemaCmd)

	return helpSystem, nil
}

func main() {
	// first, check if the args are "run-command file.yaml",
	// because we need to load the file and then run the command itself.
	// we need to do this before cobra, because we don't know which flags to load yet
	if len(os.Args) >= 3 && os.Args[1] == "run-command" && os.Args[2] != "--help" {
		// load the command
		loader := &shell.ShellCommandLoader{}
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

		// Start command
		startCmd := &cobra.Command{
			Use:   "start",
			Short: "Start the MCP server",
			Long: `Start the MCP server using the specified transport.
		
Available transports:
- stdio: Standard input/output transport (default)
- sse: Server-Sent Events transport over HTTP`,
			RunE: func(cmd *cobra.Command, args []string) error {
				// Create server
				srv := server.NewServer(log.Logger)
				promptRegistry := prompts.NewRegistry()
				resourceRegistry := resources.NewRegistry()
				toolRegistry := tools.NewRegistry()

				// Register a simple prompt directly
				promptRegistry.RegisterPrompt(protocol.Prompt{
					Name:        "simple",
					Description: "A simple prompt that can take optional context and topic arguments",
					Arguments: []protocol.PromptArgument{
						{
							Name:        "context",
							Description: "Additional context to consider",
							Required:    false,
						},
						{
							Name:        "topic",
							Description: "Specific topic to focus on",
							Required:    false,
						},
					},
				})

				// Register registries with the server
				srv.GetRegistry().RegisterPromptProvider(promptRegistry)
				srv.GetRegistry().RegisterResourceProvider(resourceRegistry)
				srv.GetRegistry().RegisterToolProvider(toolRegistry)

				// Register tools
				if err := examples.RegisterEchoTool(toolRegistry); err != nil {
					log.Error().Err(err).Msg("Error registering echo tool")
					return err
				}
				if err := examples.RegisterFetchTool(toolRegistry); err != nil {
					log.Error().Err(err).Msg("Error registering fetch tool")
					return err
				}
				if err := examples.RegisterSQLiteTool(toolRegistry); err != nil {
					log.Error().Err(err).Msg("Error registering sqlite tool")
					return err
				}
				if err := cursor.RegisterCursorTools(toolRegistry); err != nil {
					log.Error().Err(err).Msg("Error registering cursor tools")
					return err
				}

				// Create root context with cancellation
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Set up signal handling
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

				// Start server in a goroutine
				errChan := make(chan error, 1)
				go func() {
					// Start server with selected transport
					var err error
					switch transport {
					case "stdio":
						log.Info().Msg("Starting server with stdio transport")
						err = srv.StartStdio(ctx)
					case "sse":
						log.Info().Int("port", port).Msg("Starting server with SSE transport")
						err = srv.StartSSE(ctx, port)
					default:
						err = fmt.Errorf("invalid transport type: %s", transport)
					}
					errChan <- err
				}()

				// Wait for either server error or interrupt signal
				select {
				case err := <-errChan:
					if err != nil && err != io.EOF {
						log.Error().Err(err).Msg("Server error")
						return err
					}
					return nil
				case sig := <-sigChan:
					log.Info().Str("signal", sig.String()).Msg("Received signal, initiating graceful shutdown")
					// Cancel context to initiate shutdown
					cancel()
					// Create a timeout context for shutdown
					shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer shutdownCancel()
					if err := srv.Stop(shutdownCtx); err != nil {
						log.Error().Err(err).Msg("Error during shutdown")
						return err

					}
					log.Info().Msg("Server stopped gracefully")
					return nil
				}
			},
		}

		// Add flags to start command
		startCmd.Flags().StringVarP(&transport, "transport", "t", "stdio", "Transport type (stdio or sse)")
		startCmd.Flags().IntVarP(&port, "port", "p", 3001, "Port to listen on for SSE transport")

		// Version command
		versionCmd := &cobra.Command{
			Use:   "version",
			Short: "Print version information",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("mcp-server version %s\n", Version)
				fmt.Printf("  Build time: %s\n", BuildTime)
				fmt.Printf("  Git commit: %s\n", GitCommit)
			},
		}

		// Add commands to root command
		rootCmd.AddCommand(startCmd)
		rootCmd.AddCommand(versionCmd)
	}

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
