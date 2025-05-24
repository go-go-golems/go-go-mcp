package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/jsserver"
	"github.com/spf13/cobra"
)

var (
	port       string
	dbPath     string
	archiveDir string
	staticDir  string
	loadDir    string
	serverURL  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "js-server",
		Short: "JavaScript Web Server with embedded runtime",
		Long: `A Go web server that embeds a JavaScript runtime (Goja) allowing 
dynamic REST endpoint registration and file serving through JavaScript code execution.`,
	}

	// Server command
	serverCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the JavaScript web server",
		RunE:  runServer,
	}

	serverCmd.Flags().StringVar(&port, "port", "8080", "Server port")
	serverCmd.Flags().StringVar(&dbPath, "db", "jsserver.db", "SQLite database path")
	serverCmd.Flags().StringVar(&archiveDir, "archive", "code-archive", "Directory to store executed code")
	serverCmd.Flags().StringVar(&staticDir, "static", "static", "Static files directory")
	serverCmd.Flags().StringVar(&loadDir, "load-dir", "", "Directory containing JavaScript files to load on startup")

	// API commands
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Interact with the JavaScript server API",
	}

	apiCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")

	// Execute command
	executeCmd := &cobra.Command{
		Use:   "execute [file]",
		Short: "Execute JavaScript code from file or stdin",
		Args:  cobra.MaximumNArgs(1),
		RunE:  executeCode,
	}
	executeCmd.Flags().Bool("persist", false, "Archive the executed code")
	executeCmd.Flags().String("name", "", "Name for the archived code")

	// Routes commands
	routesCmd := &cobra.Command{
		Use:   "routes",
		Short: "Manage registered routes",
	}

	listRoutesCmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered routes",
		RunE:  listRoutes,
	}

	deleteRouteCmd := &cobra.Command{
		Use:   "delete [path]",
		Short: "Delete a registered route",
		Args:  cobra.ExactArgs(1),
		RunE:  deleteRoute,
	}

	routesCmd.AddCommand(listRoutesCmd, deleteRouteCmd)

	// Files commands
	filesCmd := &cobra.Command{
		Use:   "files",
		Short: "Manage registered files",
	}

	listFilesCmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered files",
		RunE:  listFiles,
	}

	deleteFileCmd := &cobra.Command{
		Use:   "delete [path]",
		Short: "Delete a registered file",
		Args:  cobra.ExactArgs(1),
		RunE:  deleteFile,
	}

	filesCmd.AddCommand(listFilesCmd, deleteFileCmd)

	// State commands
	stateCmd := &cobra.Command{
		Use:   "state",
		Short: "Manage global state",
	}

	getStateCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get global state (all if no key specified)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  getState,
	}

	setStateCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set global state value",
		Args:  cobra.ExactArgs(2),
		RunE:  setState,
	}

	deleteStateCmd := &cobra.Command{
		Use:   "delete [key]",
		Short: "Delete global state key",
		Args:  cobra.ExactArgs(1),
		RunE:  deleteState,
	}

	stateCmd.AddCommand(getStateCmd, setStateCmd, deleteStateCmd)

	// Archive commands
	archiveCmd := &cobra.Command{
		Use:   "archive",
		Short: "Manage archived code",
	}

	listArchiveCmd := &cobra.Command{
		Use:   "list",
		Short: "List archived files",
		RunE:  listArchive,
	}

	getArchiveCmd := &cobra.Command{
		Use:   "get [filename]",
		Short: "Get archived file content",
		Args:  cobra.ExactArgs(1),
		RunE:  getArchive,
	}

	archiveCmd.AddCommand(listArchiveCmd, getArchiveCmd)

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get server status",
		RunE:  getStatus,
	}

	// Load directory command
	loadDirCmd := &cobra.Command{
		Use:   "load-dir [directory]",
		Short: "Load all JavaScript files from a directory",
		Args:  cobra.ExactArgs(1),
		RunE:  loadDirectory,
	}

	apiCmd.AddCommand(executeCmd, routesCmd, filesCmd, stateCmd, archiveCmd, statusCmd, loadDirCmd)
	rootCmd.AddCommand(serverCmd, apiCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Create archive directory if it doesn't exist
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Initialize the JavaScript web server
	server, err := jsserver.New(&jsserver.Config{
		DatabasePath: dbPath,
		ArchiveDir:   archiveDir,
		StaticDir:    staticDir,
	})
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	defer server.Close()

	// Load JavaScript files from directory if specified
	if loadDir != "" {
		if err := loadJavaScriptFiles(server, loadDir); err != nil {
			log.Printf("Warning: Failed to load JavaScript files from %s: %v", loadDir, err)
		}
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: server,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting JavaScript web server on port %s", port)
		log.Printf("Database: %s", dbPath)
		log.Printf("Archive directory: %s", archiveDir)
		log.Printf("Static directory: %s", staticDir)
		if loadDir != "" {
			log.Printf("Will load JavaScript files from: %s", loadDir)
		}

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait a moment for server to start, then load JavaScript files
	if loadDir != "" {
		go func() {
			time.Sleep(1 * time.Second) // Give server time to start
			log.Printf("Loading JavaScript files from: %s", loadDir)
			if err := loadJavaScriptFiles(nil, loadDir); err != nil {
				log.Printf("Warning: Failed to load JavaScript files from %s: %v", loadDir, err)
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
	return nil
}
