package mcp

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/api"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/web"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/google/uuid"
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
		// embeddable.WithTool("executeJSFile", executeJSFileHandler,
		// 	embeddable.WithDescription("Execute JavaScript code from a file on the filesystem"),
		// 	embeddable.WithStringArg("absolutePath", "Absolute path to the JavaScript file to execute", true),
		// ),
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

		// API routes
		r.HandleFunc("/v1/execute", api.ExecuteHandler(GlobalJSEngine)).Methods("POST")
		log.Debug().Msg("Registered API endpoint: POST /v1/execute (MCP mode)")

		// Setup all admin routes (including logs console)
		web.SetupAdminRoutes(r, GlobalJSEngine)

		// Setup dynamic routes with request logging
		web.SetupDynamicRoutes(r, GlobalJSEngine)

		log.Info().Str("address", ":8080").Msg("Starting background HTTP server for MCP mode with admin console")
		log.Info().Msg("Admin console available at http://localhost:8080/admin/logs")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Error().Err(err).Msg("Background HTTP server failed")
		}
	}()

	log.Info().Msg("JavaScript engine and HTTP server initialized for MCP")
	return nil
}

// executeJSHandler is the MCP tool handler for executing JavaScript code
func executeJSHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	// Initialize engine if not already done (for test-tool command)
	if GlobalJSEngine == nil {
		log.Info().Msg("JavaScript engine not initialized, initializing now")
		if err := initializeJSEngineForMCP(ctx); err != nil {
			return protocol.NewErrorToolResult(protocol.NewTextContent(
				fmt.Sprintf("Failed to initialize JavaScript engine: %v", err))), nil
		}
	}

	// Extract code from arguments
	code, ok := args["code"].(string)
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("code must be a string")), nil
	}

	// Generate session ID for tracking
	sessionID := uuid.New().String()

	// Save the code to a file with timestamp
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	filename := fmt.Sprintf("scripts/mcp-exec-%s.js", timestamp)

	// Ensure scripts directory exists
	if err := os.MkdirAll("scripts", 0755); err != nil {
		log.Warn().Err(err).Msg("Failed to create scripts directory")
	} else {
		// Save the code to file
		if err := os.WriteFile(filename, []byte(code), 0644); err != nil {
			log.Warn().Err(err).Str("filename", filename).Msg("Failed to save code to file")
		} else {
			log.Info().Str("filename", filename).Msg("Saved executed code to file")
		}
	}

	// Execute the code with result capture
	done := make(chan error, 1)
	resultChan := make(chan *engine.EvalResult, 1)
	job := engine.EvalJob{
		Code:      code,
		Done:      done,
		Result:    resultChan,
		SessionID: sessionID,
		Source:    "mcp",
	}

	GlobalJSEngine.SubmitJob(job)

	// Wait for completion with timeout
	select {
	case result := <-resultChan:
		// Also wait for done signal to ensure completion
		select {
		case err := <-done:
			if err != nil {
				return protocol.NewErrorToolResult(protocol.NewTextContent(
					fmt.Sprintf("JavaScript execution failed: %v", err))), nil
			}
		case <-time.After(5 * time.Second):
			// Continue even if done signal is delayed
		}

		// Create response with result and console output
		responseData := map[string]interface{}{
			"success":    true,
			"result":     result.Value,
			"consoleLog": result.ConsoleLog,
			"savedAs":    filename,
			"message":    "JavaScript code executed successfully. Check http://localhost:8080 for any web endpoints created. Monitor execution at http://localhost:8080/admin/logs",
		}

		// Convert to JSON
		jsonData, err := json.Marshal(responseData)
		if err != nil {
			return protocol.NewErrorToolResult(protocol.NewTextContent(
				fmt.Sprintf("Failed to marshal result: %v", err))), nil
		}

		return protocol.NewToolResult(
			protocol.WithText(string(jsonData)),
		), nil

	case <-time.After(30 * time.Second):
		return protocol.NewErrorToolResult(protocol.NewTextContent("Timeout waiting for JavaScript execution")), nil
	}
}

// executeJSFileHandler is the MCP tool handler for executing JavaScript files
func executeJSFileHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	// Initialize engine if not already done (for test-tool command)
	if GlobalJSEngine == nil {
		log.Info().Msg("JavaScript engine not initialized, initializing now")
		if err := initializeJSEngineForMCP(ctx); err != nil {
			return protocol.NewErrorToolResult(protocol.NewTextContent(
				fmt.Sprintf("Failed to initialize JavaScript engine: %v", err))), nil
		}
	}

	// Extract file path from arguments
	filePath, ok := args["absolutePath"].(string)
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("absolutePath must be a string")), nil
	}

	// Validate that the path is absolute
	if !filepath.IsAbs(filePath) {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Path must be absolute")), nil
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return protocol.NewErrorToolResult(protocol.NewTextContent(
			fmt.Sprintf("File does not exist: %s", filePath))), nil
	}

	// Read the file
	codeBytes, err := os.ReadFile(filePath)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(
			fmt.Sprintf("Failed to read file: %v", err))), nil
	}

	code := string(codeBytes)
	log.Info().Str("file", filePath).Int("bytes", len(codeBytes)).Msg("Executing JavaScript file")

	// Generate session ID for tracking
	sessionID := uuid.New().String()

	// Execute the code with result capture
	done := make(chan error, 1)
	resultChan := make(chan *engine.EvalResult, 1)
	job := engine.EvalJob{
		Code:      code,
		Done:      done,
		Result:    resultChan,
		SessionID: sessionID,
		Source:    "mcp-file",
	}

	GlobalJSEngine.SubmitJob(job)

	// Wait for completion with timeout
	select {
	case result := <-resultChan:
		// Also wait for done signal to ensure completion
		select {
		case err := <-done:
			if err != nil {
				return protocol.NewErrorToolResult(protocol.NewTextContent(
					fmt.Sprintf("JavaScript execution failed: %v", err))), nil
			}
		case <-time.After(5 * time.Second):
			// Continue even if done signal is delayed
		}

		// Create response with result and console output
		responseData := map[string]interface{}{
			"success":      true,
			"result":       result.Value,
			"consoleLog":   result.ConsoleLog,
			"executedFile": filePath,
			"message":      fmt.Sprintf("JavaScript file executed successfully: %s. Check http://localhost:8080 for any web endpoints created. Monitor execution at http://localhost:8080/admin/logs", filepath.Base(filePath)),
		}

		// Convert to JSON
		jsonData, err := json.Marshal(responseData)
		if err != nil {
			return protocol.NewErrorToolResult(protocol.NewTextContent(
				fmt.Sprintf("Failed to marshal result: %v", err))), nil
		}

		return protocol.NewToolResult(
			protocol.WithText(string(jsonData)),
		), nil

	case <-time.After(30 * time.Second):
		return protocol.NewErrorToolResult(protocol.NewTextContent("Timeout waiting for JavaScript execution")), nil
	}
}
