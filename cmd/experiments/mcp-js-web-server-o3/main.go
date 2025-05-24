package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/api"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/web"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var (
	port       string
	db         string
	scriptsDir string
	serverURL  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "js-playground",
		Short: "JavaScript playground web server",
		Long:  "A JavaScript playground web server with SQLite integration",
	}

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

	rootCmd.AddCommand(serverCmd, executeCmd, testCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Ensure scripts directory exists
	if err := os.MkdirAll("scripts", 0755); err != nil {
		log.Fatal("Failed to create scripts directory:", err)
	}

	// Initialize JS engine
	jsEngine := engine.NewEngine(db)
	if err := jsEngine.Init("bootstrap.js"); err != nil {
		log.Printf("Warning: failed to load bootstrap.js: %v", err)
	}

	// Load scripts from directory if specified
	if scriptsDir != "" {
		loadScriptsFromDir(jsEngine, scriptsDir)
	}

	// Start dispatcher goroutine
	go jsEngine.StartDispatcher()

	// Setup router
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/v1/execute", api.ExecuteHandler(jsEngine)).Methods("POST")

	// Dynamic routes (handled by JS)
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.HandleDynamicRoute(jsEngine, w, r)
	})

	addr := ":" + port
	fmt.Printf("Starting JavaScript playground server on %s\n", addr)
	fmt.Printf("Database: %s\n", db)
	if scriptsDir != "" {
		fmt.Printf("Scripts directory: %s\n", scriptsDir)
	}
	fmt.Printf("POST /v1/execute to run JavaScript code\n")

	log.Fatal(http.ListenAndServe(addr, r))
}

func runExecute(cmd *cobra.Command, args []string) {
	input := args[0]
	var code string

	// Check if input is a file
	if fileInfo, err := os.Stat(input); err == nil && !fileInfo.IsDir() {
		data, err := os.ReadFile(input)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", input, err)
		}
		code = string(data)
		fmt.Printf("Executing file: %s\n", input)
	} else {
		code = input
		fmt.Printf("Executing code: %s\n", code)
	}

	// Send to server
	resp, err := http.Post(serverURL+"/v1/execute", "application/javascript", strings.NewReader(code))
	if err != nil {
		log.Fatalf("Failed to execute code: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	if len(body) > 0 {
		fmt.Printf("Response: %s\n", string(body))
	}
}

func runTest(cmd *cobra.Command, args []string) {
	fmt.Printf("Testing server at %s\n", serverURL)

	// Test health endpoint
	fmt.Println("\n1. Testing health endpoint...")
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   GET /health -> %s: %s\n", resp.Status, string(body))
	}

	// Test root endpoint
	fmt.Println("\n2. Testing root endpoint...")
	resp, err = http.Get(serverURL + "/")
	if err != nil {
		log.Printf("Root endpoint failed: %v", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   GET / -> %s: %s\n", resp.Status, string(body))
	}

	// Test counter endpoint
	fmt.Println("\n3. Testing counter endpoint...")
	resp, err = http.Post(serverURL+"/counter", "application/json", strings.NewReader("{}"))
	if err != nil {
		log.Printf("Counter endpoint failed: %v", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   POST /counter -> %s: %s\n", resp.Status, string(body))
	}

	// Test execute endpoint
	fmt.Println("\n4. Testing execute endpoint...")
	testCode := `
		console.log("Hello from executed code!");
		registerHandler("GET", "/test", () => ({message: "Test endpoint works!", time: new Date().toISOString()}));
	`
	resp, err = http.Post(serverURL+"/v1/execute", "application/javascript", strings.NewReader(testCode))
	if err != nil {
		log.Printf("Execute endpoint failed: %v", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   POST /v1/execute -> %s: %s\n", resp.Status, string(body))
	}

	// Test newly created endpoint
	fmt.Println("\n5. Testing dynamically created endpoint...")
	resp, err = http.Get(serverURL + "/test")
	if err != nil {
		log.Printf("Dynamic endpoint failed: %v", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   GET /test -> %s: %s\n", resp.Status, string(body))
	}
}

func loadScriptsFromDir(jsEngine *engine.Engine, dir string) {
	fmt.Printf("Loading JavaScript files from %s...\n", dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".js") {
			fmt.Printf("  Loading: %s\n", path)
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Failed to read %s: %v", path, err)
				return nil
			}

			// Submit to engine
			done := make(chan error, 1)
			job := engine.EvalJob{
				Code: string(data),
				Done: done,
			}
			jsEngine.SubmitJob(job)

			// Wait for completion
			if err := <-done; err != nil {
				log.Printf("Failed to execute %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error walking scripts directory: %v", err)
	}
}