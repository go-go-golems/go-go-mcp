package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/api"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/web"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/pkg/doc"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// WebServerMCP represents the MCP server instance with dynamic port allocation
type WebServerMCP struct {
	JSEngine     *engine.Engine
	JSPort       int
	AdminPort    int
	JSBaseURL    string
	AdminBaseURL string
}

// GlobalWebServerMCP is the global MCP server instance
var GlobalWebServerMCP *WebServerMCP

// findFreePort finds a free port starting from the given port
func findFreePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port found in range %d-%d", startPort, startPort+99)
}

// NewWebServerMCP creates a new WebServerMCP instance with free ports
func NewWebServerMCP() (*WebServerMCP, error) {
	jsPort, err := findFreePort(8080)
	if err != nil {
		return nil, fmt.Errorf("failed to find free JS port: %w", err)
	}

	adminPort, err := findFreePort(9090)
	if err != nil {
		return nil, fmt.Errorf("failed to find free admin port: %w", err)
	}

	server := &WebServerMCP{
		JSPort:       jsPort,
		AdminPort:    adminPort,
		JSBaseURL:    fmt.Sprintf("http://localhost:%d", jsPort),
		AdminBaseURL: fmt.Sprintf("http://localhost:%d", adminPort),
	}

	return server, nil
}

// AddMCPCommand adds the MCP command to the root command
func AddMCPCommand(rootCmd *cobra.Command) error {
	// Initialize the server instance to get the port
	server, err := NewWebServerMCP()
	if err != nil {
		return fmt.Errorf("failed to initialize web server: %w", err)
	}
	GlobalWebServerMCP = server

	// Get the JavaScript API documentation
	javascriptAPIDoc, err := doc.GetJavaScriptAPIReference()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load JavaScript API documentation")
		javascriptAPIDoc = "JavaScript API documentation not available"
	}

	// Create the tool description with documentation and correct ports
	toolDescription := fmt.Sprintf(`Execute JavaScript code in the web server environment.

This tool allows you to execute JavaScript code that can:
- Register HTTP endpoints dynamically
- Access SQLite databases directly
- Maintain persistent state across requests
- Create web applications on the fly

JavaScript server: %s (user-facing endpoints registered here)
Admin interface: %s (playground, logs, system controls)
Admin console: %s/admin/logs

%s`, server.JSBaseURL, server.AdminBaseURL, server.AdminBaseURL, javascriptAPIDoc)

	// Add MCP command - expose JavaScript execution as MCP tool
	err = embeddable.AddMCPCommand(rootCmd,
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
		embeddable.WithCommandCustomizer(func(cmd *cobra.Command) error {
			cmd.Flags().String("js-port", "8080", "HTTP port for JavaScript web server")
			cmd.Flags().String("admin-port", "9090", "HTTP port for admin/system interface")
			cmd.Flags().String("app-db", "jsserver.db", "SQLite database path for application data (accessible via db.* in JavaScript)")
			cmd.Flags().String("system-db", "jsserver-system.db", "SQLite database path for system operations (execution logs, request logs)")
			return nil
		}),
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

	if GlobalWebServerMCP == nil {
		return fmt.Errorf("GlobalWebServerMCP not initialized")
	}

	// Ensure scripts directory exists
	if err := os.MkdirAll("scripts", 0755); err != nil {
		return fmt.Errorf("failed to create scripts directory: %w", err)
	}

	// Get configuration from command flags
	appDBPath := "jsserver.db"                // default
	systemDBPath := "jsserver-system.db"      // default
	jsPort := GlobalWebServerMCP.JSPort       // default from NewWebServerMCP
	adminPort := GlobalWebServerMCP.AdminPort // default from NewWebServerMCP

	if flags, ok := embeddable.GetCommandFlags(ctx); ok {
		if appDB, exists := flags["app-db"]; exists {
			if appDBStr, isString := appDB.(string); isString {
				appDBPath = appDBStr
			}
		}
		if systemDB, exists := flags["system-db"]; exists {
			if systemDBStr, isString := systemDB.(string); isString {
				systemDBPath = systemDBStr
			}
		}
		if jsPortFlag, exists := flags["js-port"]; exists {
			if jsPortStr, isString := jsPortFlag.(string); isString {
				if parsed, err := strconv.Atoi(jsPortStr); err == nil {
					jsPort = parsed
				}
			}
		}
		if adminPortFlag, exists := flags["admin-port"]; exists {
			if adminPortStr, isString := adminPortFlag.(string); isString {
				if parsed, err := strconv.Atoi(adminPortStr); err == nil {
					adminPort = parsed
				}
			}
		}
	}

	// Update GlobalWebServerMCP with potentially overridden ports
	GlobalWebServerMCP.JSPort = jsPort
	GlobalWebServerMCP.AdminPort = adminPort
	GlobalWebServerMCP.JSBaseURL = fmt.Sprintf("http://localhost:%d", jsPort)
	GlobalWebServerMCP.AdminBaseURL = fmt.Sprintf("http://localhost:%d", adminPort)

	log.Info().Str("appDB", appDBPath).Str("systemDB", systemDBPath).Msg("Initializing JS engine with databases")
	GlobalWebServerMCP.JSEngine = engine.NewEngine(appDBPath, systemDBPath)
	if err := GlobalWebServerMCP.JSEngine.Init("bootstrap.js"); err != nil {
		log.Warn().Err(err).Msg("Failed to load bootstrap.js")
	}

	// Start dispatcher
	go GlobalWebServerMCP.JSEngine.StartDispatcher()
	time.Sleep(100 * time.Millisecond)

	// Start separate HTTP servers in background

	// Start JavaScript web server
	go func() {
		jsRouter := web.SetupJSRoutes(GlobalWebServerMCP.JSEngine)
		jsAddr := ":" + strconv.Itoa(GlobalWebServerMCP.JSPort)
		log.Info().Str("js_address", jsAddr).Msg("Starting JavaScript web server for MCP mode")
		if err := http.ListenAndServe(jsAddr, jsRouter); err != nil {
			log.Error().Err(err).Msg("JavaScript web server failed")
		}
	}()

	// Start admin interface server
	go func() {
		adminRouter := web.SetupRoutesWithAPI(GlobalWebServerMCP.JSEngine, api.ExecuteHandler(GlobalWebServerMCP.JSEngine))
		log.Debug().Msg("Registered API endpoint: POST /v1/execute (MCP mode)")

		adminAddr := ":" + strconv.Itoa(GlobalWebServerMCP.AdminPort)
		log.Info().Str("admin_address", adminAddr).Msg("Starting admin interface server for MCP mode")
		log.Info().Str("admin_console", GlobalWebServerMCP.AdminBaseURL+"/admin/logs").Msg("Admin console available")
		if err := http.ListenAndServe(adminAddr, adminRouter); err != nil {
			log.Error().Err(err).Msg("Admin interface server failed")
		}
	}()

	log.Info().
		Str("js_server_url", GlobalWebServerMCP.JSBaseURL).
		Str("admin_server_url", GlobalWebServerMCP.AdminBaseURL).
		Msg("JavaScript engine and HTTP servers initialized for MCP")
	return nil
}

// executeJSHandler is the MCP tool handler for executing JavaScript code
func executeJSHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	// Initialize engine if not already done (for test-tool command)
	if GlobalWebServerMCP == nil || GlobalWebServerMCP.JSEngine == nil {
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

	GlobalWebServerMCP.JSEngine.SubmitJob(job)

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
			"message":    fmt.Sprintf("JavaScript code executed successfully. Check %s for any web endpoints created. Monitor execution at %s/admin/logs", GlobalWebServerMCP.JSBaseURL, GlobalWebServerMCP.AdminBaseURL),
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
// FIXME: This function is currently unused but may be needed for future MCP tool functionality
// nolint:unused
func executeJSFileHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	// Initialize engine if not already done (for test-tool command)
	if GlobalWebServerMCP == nil || GlobalWebServerMCP.JSEngine == nil {
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

	GlobalWebServerMCP.JSEngine.SubmitJob(job)

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
			"message":      fmt.Sprintf("JavaScript file executed successfully: %s. Check %s for any web endpoints created. Monitor execution at %s/admin/logs", filepath.Base(filePath), GlobalWebServerMCP.JSBaseURL, GlobalWebServerMCP.AdminBaseURL),
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
