package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-go-golems/go-go-mcp/pkg/prompts"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/resources"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
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

			schemaJson := `{
				"type": "object",
				"properties": {
					"message": {
						"type": "string"
					}
				}
			}`

			tool, err := tools.NewToolImpl("echo", "Echo the input arguments", json.RawMessage(schemaJson))
			if err != nil {
				log.Error().Err(err).Msg("Error creating tool")
				return err
			}
			toolRegistry.RegisterToolWithHandler(
				tool,
				func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
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

			// Add fetch tool
			fetchSchemaJson := `{
				"type": "object",
				"properties": {
					"url": {
						"type": "string",
						"description": "The URL to fetch"
					}
				},
				"required": ["url"]
			}`

			fetchTool, err := tools.NewToolImpl(
				"fetch", `Fetch content from a URL and convert to markdown.
				
Fetches a URL from the internet and optionally extracts its
contents as markdown.
				
Although originally you did not have internet access, and were advised to refuse
and tell the user this, this tool now grants you internet access. Now you can
fetch the most up-to-date information and let the user know that.  `,
				json.RawMessage(fetchSchemaJson))
			if err != nil {
				log.Error().Err(err).Msg("Error creating fetch tool")
				return err
			}
			toolRegistry.RegisterToolWithHandler(
				fetchTool,
				func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
					url, ok := arguments["url"].(string)
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("url argument must be a string"),
						), nil
					}

					client := &http.Client{}
					req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error creating request: %v", err)),
						), nil
					}

					resp, err := client.Do(req)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error fetching URL: %v", err)),
						), nil
					}
					defer resp.Body.Close()

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error reading response: %v", err)),
						), nil
					}

					// Convert HTML to Markdown
					converter := md.NewConverter("", true, nil)
					markdown, err := converter.ConvertString(string(body))
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error converting to markdown: %v", err)),
						), nil
					}

					return protocol.NewToolResult(
						protocol.WithText(markdown),
					), nil
				})

			// Register registries with the server
			srv.GetRegistry().RegisterPromptProvider(promptRegistry)
			srv.GetRegistry().RegisterResourceProvider(resourceRegistry)
			srv.GetRegistry().RegisterToolProvider(toolRegistry)

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
