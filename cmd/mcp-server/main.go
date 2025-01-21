package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/prompts"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/resources"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/cursor"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

func main() {
	rootCmd := &cobra.Command{
		Use:   "mcp-server",
		Short: "MCP server implementation in Go",
		Long: `A Model Context Protocol (MCP) server implementation in Go.
Supports both stdio and SSE transports for client-server communication.

The server implements the Model Context Protocol (MCP) specification,
providing a framework for building MCP servers and clients.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid log level %s, defaulting to info\n", logLevel)
				level = zerolog.InfoLevel
			}

			var writer zerolog.ConsoleWriter
			if term.IsTerminal(int(os.Stderr.Fd())) {
				writer = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
			} else {
				writer = zerolog.ConsoleWriter{
					Out:        os.Stderr,
					TimeFormat: time.RFC3339,
					NoColor:    true,
				}
			}

			logger := zerolog.New(writer).With().Timestamp()
			if withCaller {
				logger = logger.Caller()
			}
			log.Logger = logger.Logger()

			zerolog.SetGlobalLevel(level)
			if debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
		},
	}

	// Add persistent flags available to all commands
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().BoolVar(&withCaller, "with-caller", true, "Show caller information in logs")

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
			if err := tools.RegisterEchoTool(toolRegistry); err != nil {
				log.Error().Err(err).Msg("Error registering echo tool")
				return err
			}
			if err := tools.RegisterFetchTool(toolRegistry); err != nil {
				log.Error().Err(err).Msg("Error registering fetch tool")
				return err
			}
			if err := tools.RegisterSQLiteTool(toolRegistry); err != nil {
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

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
