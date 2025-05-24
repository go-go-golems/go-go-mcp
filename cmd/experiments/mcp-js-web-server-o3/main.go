package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/api"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/mcp"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/web"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	port       string
	db         string
	scriptsDir string
	serverURL  string
	logLevel   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "js-playground",
		Short: "JavaScript playground web server",
		Long:  "A JavaScript playground web server with SQLite integration",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setupLogging()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "debug", "Log level (trace, debug, info, warn, error, fatal, panic)")

	// Server command
	serverCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the JavaScript playground server",
		Run:   runServer,
	}
	serverCmd.Flags().StringVarP(&port, "port", "p", "8080", "HTTP port to listen on")
	serverCmd.Flags().StringVarP(&db, "db", "d", "data.sqlite", "SQLite database path")
	serverCmd.Flags().StringVarP(&scriptsDir, "scripts", "s", "", "Directory containing JavaScript files to load on startup")

	// Execute command
	executeCmd := &cobra.Command{
		Use:   "execute [file or code]",
		Short: "Execute JavaScript code",
		Args:  cobra.ExactArgs(1),
		Run:   runExecute,
	}
	executeCmd.Flags().StringVarP(&serverURL, "url", "u", "http://localhost:8080", "Server URL")

	// Test command
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test the server endpoints",
		Run:   runTest,
	}
	testCmd.Flags().StringVarP(&serverURL, "url", "u", "http://localhost:8080", "Server URL")

	// MCP command - expose JavaScript execution as MCP tool
	if err := mcp.AddMCPCommand(rootCmd); err != nil {
		log.Fatal().Err(err).Msg("Failed to add MCP command")
	}

	rootCmd.AddCommand(serverCmd, executeCmd, testCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

func setupLogging() {
	// Configure zerolog - always log to stderr to avoid interfering with MCP protocol
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	// Set log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Str("level", logLevel).Msg("Invalid log level, using debug")
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Debug().Str("level", level.String()).Msg("Logging initialized")
}

func runServer(cmd *cobra.Command, args []string) {
	log.Info().Msg("Starting JavaScript playground server")

	// Ensure scripts directory exists
	if err := os.MkdirAll("scripts", 0755); err != nil {
		log.Fatal().Err(err).Msg("Failed to create scripts directory")
	}
	log.Debug().Msg("Scripts directory ready")

	// Initialize JS engine
	log.Debug().Str("database", db).Msg("Initializing JavaScript engine")
	jsEngine := engine.NewEngine(db)
	if err := jsEngine.Init("bootstrap.js"); err != nil {
		log.Warn().Err(err).Msg("Failed to load bootstrap.js")
	}

	// Start dispatcher goroutine first
	log.Debug().Msg("Starting JavaScript dispatcher")
	jsEngine.StartDispatcher()

	// Give dispatcher time to start
	time.Sleep(100 * time.Millisecond)

	// Load scripts from directory if specified
	if scriptsDir != "" {
		log.Info().Str("directory", scriptsDir).Msg("Loading scripts from directory")
		loadScriptsFromDir(jsEngine, scriptsDir)
		log.Info().Msg("Finished loading scripts")
	}

	// Setup router
	log.Debug().Msg("Setting up HTTP router")
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/v1/execute", api.ExecuteHandler(jsEngine)).Methods("POST")
	log.Debug().Msg("Registered API endpoint: POST /v1/execute")

	// Setup admin routes (consolidated)
	web.SetupAdminRoutes(r, jsEngine)

	// Setup dynamic routes with logging (consolidated)
	web.SetupDynamicRoutes(r, jsEngine)

	addr := ":" + port
	log.Info().Str("address", addr).Str("database", db).Msg("Server configuration")
	if scriptsDir != "" {
		log.Info().Str("scripts", scriptsDir).Msg("Scripts directory configured")
	}
	log.Info().Msg("POST /v1/execute to run JavaScript code")

	log.Info().Str("address", addr).Msg("Starting HTTP server")
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal().Err(err).Msg("HTTP server failed")
	}
}

func runExecute(cmd *cobra.Command, args []string) {
	input := args[0]
	var code string

	// Check if input is a file
	if fileInfo, err := os.Stat(input); err == nil && !fileInfo.IsDir() {
		data, err := os.ReadFile(input)
		if err != nil {
			log.Fatal().Err(err).Str("file", input).Msg("Failed to read file")
		}
		code = string(data)
		log.Info().Str("file", input).Msg("Executing file")
	} else {
		code = input
		log.Info().Str("code", code).Msg("Executing code")
	}

	// Send to server
	log.Debug().Str("url", serverURL+"/v1/execute").Msg("Sending request to server")
	resp, err := http.Post(serverURL+"/v1/execute", "application/javascript", strings.NewReader(code))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to execute code")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read response")
	}

	fmt.Printf("Status: %s\n", resp.Status)
	if len(body) > 0 {
		fmt.Printf("Response: %s\n", string(body))
	}
}

func runTest(cmd *cobra.Command, args []string) {
	log.Info().Str("url", serverURL).Msg("Testing server")

	// Test health endpoint
	log.Info().Msg("Testing health endpoint")
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		log.Error().Err(err).Msg("Health check failed")
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		log.Info().Str("status", resp.Status).Str("body", string(body)).Msg("GET /health")
	}

	// Test root endpoint
	log.Info().Msg("Testing root endpoint")
	resp, err = http.Get(serverURL + "/")
	if err != nil {
		log.Error().Err(err).Msg("Root endpoint failed")
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		log.Info().Str("status", resp.Status).Str("body", string(body)).Msg("GET /")
	}

	// Test counter endpoint
	log.Info().Msg("Testing counter endpoint")
	resp, err = http.Post(serverURL+"/counter", "application/json", strings.NewReader("{}"))
	if err != nil {
		log.Error().Err(err).Msg("Counter endpoint failed")
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		log.Info().Str("status", resp.Status).Str("body", string(body)).Msg("POST /counter")
	}

	// Test execute endpoint
	log.Info().Msg("Testing execute endpoint")
	testCode := `
		console.log("Hello from executed code!");
		registerHandler("GET", "/test", () => ({message: "Test endpoint works!", time: new Date().toISOString()}));
	`
	resp, err = http.Post(serverURL+"/v1/execute", "application/javascript", strings.NewReader(testCode))
	if err != nil {
		log.Error().Err(err).Msg("Execute endpoint failed")
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		log.Info().Str("status", resp.Status).Str("body", string(body)).Msg("POST /v1/execute")
	}

	// Test newly created endpoint
	log.Info().Msg("Testing dynamically created endpoint")
	resp, err = http.Get(serverURL + "/test")
	if err != nil {
		log.Error().Err(err).Msg("Dynamic endpoint failed")
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		log.Info().Str("status", resp.Status).Str("body", string(body)).Msg("GET /test")
	}
}

func loadScriptsFromDir(jsEngine *engine.Engine, dir string) {
	log.Info().Str("directory", dir).Msg("Loading JavaScript files")

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error accessing file")
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".js") {
			log.Info().Str("file", path).Msg("Loading JavaScript file")
			data, err := os.ReadFile(path)
			if err != nil {
				log.Error().Err(err).Str("file", path).Msg("Failed to read file")
				return nil
			}

			log.Debug().Str("file", path).Int("bytes", len(data)).Msg("Read JavaScript file")

			// Submit to engine with timeout
			done := make(chan error, 1)
			job := engine.EvalJob{
				Code:      string(data),
				Done:      done,
				SessionID: "startup-" + filepath.Base(path),
				Source:    "file",
			}

			log.Debug().Str("file", path).Msg("Submitting job to engine")
			jsEngine.SubmitJob(job)

			// Wait for completion with timeout
			select {
			case err := <-done:
				if err != nil {
					log.Error().Err(err).Str("file", path).Msg("Failed to execute file")
				} else {
					log.Info().Str("file", path).Msg("Successfully loaded JavaScript file")
				}
			case <-time.After(10 * time.Second):
				log.Error().Str("file", path).Msg("Timeout waiting for file execution")
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("directory", dir).Msg("Error walking scripts directory")
	}
}
