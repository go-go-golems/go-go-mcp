package embeddable

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-go-golems/go-go-mcp/pkg/server"
	transportpkg "github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/sse"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/stdio"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/streamable_http"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewMCPCommand creates a new 'mcp' command with the given options
func NewMCPCommand(opts ...ServerOption) *cobra.Command {
	config := NewServerConfig()

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			log.Fatal().Err(err).Msg("Failed to configure server")
		}
	}

	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
		Long:  fmt.Sprintf("%s - %s", config.Name, config.Description),
	}

	// Add start command
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the MCP server",
		Long:  "Start the MCP server with the specified transport",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServer(cmd, config)
		},
	}

	// Add flags
	startCmd.Flags().String("transport", config.defaultTransport, "Transport type (stdio, sse, streamable_http)")
	startCmd.Flags().Int("port", config.defaultPort, "Port for SSE and streamable HTTP transport")
	startCmd.Flags().StringSlice("internal-servers", config.internalServers, "Built-in tools to enable")
	if config.enableConfig {
		startCmd.Flags().String("config", config.configFile, "Configuration file path")
	}

	mcpCmd.AddCommand(startCmd)

	// Add list-tools command
	listToolsCmd := &cobra.Command{
		Use:   "list-tools",
		Short: "List available tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTools(cmd, config)
		},
	}
	mcpCmd.AddCommand(listToolsCmd)

	// Add test-tool command
	testToolCmd := &cobra.Command{
		Use:   "test-tool [tool-name]",
		Short: "Test a specific tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return testTool(cmd, config, args[0])
		},
	}
	mcpCmd.AddCommand(testToolCmd)

	if config.enableConfig {
		// Add config command
		configCmd := &cobra.Command{
			Use:   "config",
			Short: "Configuration management",
		}

		configCmd.AddCommand(&cobra.Command{
			Use:   "init",
			Short: "Initialize configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				return initConfig(cmd, config)
			},
		})

		configCmd.AddCommand(&cobra.Command{
			Use:   "show",
			Short: "Show current configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				return showConfig(cmd, config)
			},
		})

		mcpCmd.AddCommand(configCmd)
	}

	return mcpCmd
}

// AddMCPCommand adds a standard 'mcp' subcommand to an existing cobra application
func AddMCPCommand(rootCmd *cobra.Command, opts ...ServerOption) error {
	mcpCmd := NewMCPCommand(opts...)
	rootCmd.AddCommand(mcpCmd)
	return nil
}

func startServer(cmd *cobra.Command, config *ServerConfig) error {
	// Set up logger
	logger := log.Logger

	// Get transport type
	transportType, err := cmd.Flags().GetString("transport")
	if err != nil {
		return fmt.Errorf("failed to get transport flag: %w", err)
	}

	// Get port
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("failed to get port flag: %w", err)
	}

	// Create transport
	var transport transportpkg.Transport
	switch transportType {
	case "stdio":
		transport, err = stdio.NewStdioTransport(transportpkg.WithLogger(logger))
		if err != nil {
			return fmt.Errorf("failed to create stdio transport: %w", err)
		}
	case "sse":
		addr := ":" + strconv.Itoa(port)
		transport, err = sse.NewSSETransport(
			transportpkg.WithLogger(logger),
			transportpkg.WithSSEOptions(transportpkg.SSEOptions{
				Addr: addr,
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to create SSE transport: %w", err)
		}
	case "streamable_http":
		addr := ":" + strconv.Itoa(port)
		transport, err = streamable_http.NewStreamableHTTPTransport(
			transportpkg.WithLogger(logger),
			transportpkg.WithStreamableHTTPOptions(transportpkg.StreamableHTTPOptions{
				Addr:            addr,
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to create streamable HTTP transport: %w", err)
		}
	default:
		return fmt.Errorf("unsupported transport type: %s", transportType)
	}

	// Create server
	srv := server.NewServer(logger, transport,
		server.WithServerName(config.Name),
		server.WithServerVersion(config.Version),
		server.WithToolProvider(config.GetToolProvider()),
		server.WithSessionStore(config.sessionStore),
	)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info().Msg("Received shutdown signal")
		cancel()
	}()

	// Start server
	logger.Info().
		Str("transport", transportType).
		Int("port", port).
		Msg("Starting MCP server")

	err = srv.Start(ctx)
	if err != nil && err != context.Canceled {
		return fmt.Errorf("server error: %w", err)
	}

	logger.Info().Msg("Server stopped")
	return nil
}

func listTools(cmd *cobra.Command, config *ServerConfig) error {
	tools, _, err := config.toolRegistry.ListTools(context.Background(), "")
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	if len(tools) == 0 {
		fmt.Println("No tools registered")
		return nil
	}

	fmt.Printf("Available tools (%d):\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("  %s - %s\n", tool.Name, tool.Description)
	}

	return nil
}

func testTool(cmd *cobra.Command, config *ServerConfig, toolName string) error {
	// This is a simple implementation - in a full version we'd want to
	// allow interactive input of arguments
	result, err := config.toolRegistry.CallTool(context.Background(), toolName, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to call tool %s: %w", toolName, err)
	}

	fmt.Printf("Tool %s result:\n", toolName)
	if len(result.Content) > 0 {
		for _, content := range result.Content {
			if content.Type == "text" {
				fmt.Printf("  Text: %s\n", content.Text)
			}
		}
	}

	return nil
}

func initConfig(cmd *cobra.Command, config *ServerConfig) error {
	fmt.Println("Configuration initialization not implemented yet")
	return nil
}

func showConfig(cmd *cobra.Command, config *ServerConfig) error {
	fmt.Printf("Server Name: %s\n", config.Name)
	fmt.Printf("Server Version: %s\n", config.Version)
	fmt.Printf("Description: %s\n", config.Description)
	fmt.Printf("Default Transport: %s\n", config.defaultTransport)
	fmt.Printf("Default Port: %d\n", config.defaultPort)
	fmt.Printf("Config Enabled: %t\n", config.enableConfig)
	if config.configFile != "" {
		fmt.Printf("Config File: %s\n", config.configFile)
	}
	if len(config.internalServers) > 0 {
		fmt.Printf("Internal Servers: %v\n", config.internalServers)
	}
	return nil
}
