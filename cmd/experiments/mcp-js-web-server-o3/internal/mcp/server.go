package mcp

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/api"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/web"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed docs/javascript-api.md
var javascriptAPIDoc string

// GlobalJSEngine is the global JavaScript engine instance
var GlobalJSEngine *engine.Engine

// AddMCPCommand adds the MCP command to the root command
func AddMCPCommand(rootCmd *cobra.Command) error {
	// Create the tool description with embedded documentation
	toolDescription := fmt.Sprintf(`Execute JavaScript code in the web server environment.

This tool allows you to execute JavaScript code that can:
- Register HTTP endpoints dynamically
- Access SQLite databases directly
- Maintain persistent state across requests
- Create web applications on the fly

%s`, javascriptAPIDoc)

	// Add MCP command - expose JavaScript execution as MCP tool
	err := embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("JavaScript Web Server MCP"),
		embeddable.WithVersion("1.0.0"),
		embeddable.WithServerDescription("Execute JavaScript code and create dynamic web applications"),
		embeddable.WithTool("executeJS", executeJSHandler,
			embeddable.WithDescription(toolDescription),
			embeddable.WithStringArg("code", "JavaScript code to execute", true),
		),
		embeddable.WithHooks(&embeddable.Hooks{
			OnServerStart: initializeJSEngineForMCP,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to add MCP command: %w", err)
	}

	return nil
}

// initializeJSEngineForMCP initializes the JavaScript engine and HTTP server when MCP starts
func initializeJSEngineForMCP(ctx context.Context) error {
	log.Info().Msg("Initializing JavaScript engine for MCP")

	// Ensure scripts directory exists
	if err := os.MkdirAll("scripts", 0755); err != nil {
		return fmt.Errorf("failed to create scripts directory: %w", err)
	}

	// Initialize JS engine with default database
	GlobalJSEngine = engine.NewEngine("jsserver.db")
	if err := GlobalJSEngine.Init("bootstrap.js"); err != nil {
		log.Warn().Err(err).Msg("Failed to load bootstrap.js")
	}

	// Start dispatcher
	go GlobalJSEngine.StartDispatcher()
	time.Sleep(100 * time.Millisecond)

	// Start HTTP server in background
	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/v1/execute", api.ExecuteHandler(GlobalJSEngine)).Methods("POST")
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			web.HandleDynamicRoute(GlobalJSEngine, w, r)
		})

		log.Info().Str("address", ":8080").Msg("Starting background HTTP server for MCP mode")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Error().Err(err).Msg("Background HTTP server failed")
		}
	}()

	log.Info().Msg("JavaScript engine and HTTP server initialized for MCP")
	return nil
}

// executeJSHandler is the MCP tool handler for executing JavaScript code
func executeJSHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	// Engine should already be initialized by the startup hook
	if GlobalJSEngine == nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("JavaScript engine not initialized")), nil
	}

	// Extract code from arguments
	code, ok := args["code"].(string)
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("code must be a string")), nil
	}

	// Execute the code
	done := make(chan error, 1)
	job := engine.EvalJob{
		Code: code,
		Done: done,
	}

	GlobalJSEngine.SubmitJob(job)

	// Wait for completion with timeout
	select {
	case err := <-done:
		if err != nil {
			return protocol.NewErrorToolResult(protocol.NewTextContent(
				fmt.Sprintf("JavaScript execution failed: %v", err))), nil
		}

		return protocol.NewToolResult(
			protocol.WithText("JavaScript code executed successfully. Check http://localhost:8080 for any web endpoints created."),
		), nil

	case <-time.After(30 * time.Second):
		return protocol.NewErrorToolResult(protocol.NewTextContent("Timeout waiting for JavaScript execution")), nil
	}
}
