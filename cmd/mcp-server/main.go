package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/prompts"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/resources"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
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
			zerolog.SetGlobalLevel(level)
			if debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if withCaller {
				zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
					short := file
					for i := len(file) - 1; i > 0; i-- {
						if file[i] == '/' {
							short = file[i+1:]
							break
						}
					}
					return fmt.Sprintf("%s:%d", short, line)
				}
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
			// Use ConsoleWriter for colored output
			consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
			logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

			// Create server
			srv := server.NewServer(logger)
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

			schemaJson := `{
				"type": "object",
				"properties": {
					"message": {
						"type": "string"
					}
				}
			}`
			toolRegistry.RegisterToolWithHandler(
				protocol.Tool{
					Name:        "echo",
					Description: "Echo the input arguments",
					InputSchema: json.RawMessage(schemaJson),
				},
				func(tool protocol.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
					message, ok := arguments["message"].(string)
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("message argument must be a string"),
						), nil
					}
					return protocol.NewToolResult(
						protocol.WithText(message),
					), nil
				})

			// Register registries with the server
			srv.GetRegistry().RegisterPromptProvider(promptRegistry)
			srv.GetRegistry().RegisterResourceProvider(resourceRegistry)
			srv.GetRegistry().RegisterToolProvider(toolRegistry)

			// Start server with selected transport
			switch transport {
			case "stdio":
				logger.Info().Msg("Starting server with stdio transport")
				return srv.StartStdio()
			case "sse":
				logger.Info().Int("port", port).Msg("Starting server with SSE transport")
				return srv.StartSSE(port)
			default:
				return fmt.Errorf("invalid transport type: %s", transport)
			}
		},
	}

	// Add flags to start command
	startCmd.Flags().StringVarP(&transport, "transport", "t", "stdio", "Transport type (stdio or sse)")
	startCmd.Flags().IntVarP(&port, "port", "p", 8000, "Port to listen on for SSE transport")

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
