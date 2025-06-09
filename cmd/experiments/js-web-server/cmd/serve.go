package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/api"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/web"
	pinocchio_cmds "github.com/go-go-golems/pinocchio/pkg/cmds"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ServeCmd represents the serve command
type ServeCmd struct {
	*cmds.CommandDescription
}

// ServeSettings holds the configuration for the serve command
type ServeSettings struct {
	Port       string `glazed.parameter:"port"`
	AdminPort  string `glazed.parameter:"admin-port"`
	AppDB      string `glazed.parameter:"app-db"`
	SystemDB   string `glazed.parameter:"system-db"`
	ScriptsDir string `glazed.parameter:"scripts"`
}

// Ensure ServeCmd implements BareCommand
var _ cmds.BareCommand = &ServeCmd{}

// NewServeCmd creates a new serve command with Geppetto layers
func NewServeCmd() (*ServeCmd, error) {
	// Create temporary step settings for Geppetto layers
	tempSettings, err := settings.NewStepSettings()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temporary step settings")
	}

	// Create Geppetto layers using pinocchio helper
	geppettoLayers, err := pinocchio_cmds.CreateGeppettoLayers(tempSettings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Geppetto layers")
	}

	return &ServeCmd{
		CommandDescription: cmds.NewCommandDescription(
			"serve",
			cmds.WithShort("Start the JavaScript playground server with Geppetto AI capabilities"),
			cmds.WithLong(`
Start the JavaScript playground server with integrated Geppetto AI capabilities.

The server provides:
- JavaScript runtime with Geppetto APIs (Conversation, ChatStepFactory)
- SQLite integration for application and system data
- Admin interface for monitoring and management
- Script loading from directory on startup
- RESTful API for JavaScript execution

Examples:
  serve --port 8080 --scripts ./scripts
  serve --app-db app.db --system-db system.db --admin-port 9090
			`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"port",
					parameters.ParameterTypeString,
					parameters.WithHelp("HTTP port for JavaScript web server"),
					parameters.WithDefault("8080"),
					parameters.WithShortFlag("p"),
				),
				parameters.NewParameterDefinition(
					"admin-port",
					parameters.ParameterTypeString,
					parameters.WithHelp("HTTP port for admin/system interface"),
					parameters.WithDefault("9090"),
				),
				parameters.NewParameterDefinition(
					"app-db",
					parameters.ParameterTypeString,
					parameters.WithHelp("SQLite database path for application data (accessible via db.* in JavaScript)"),
					parameters.WithDefault("data.sqlite"),
					parameters.WithShortFlag("d"),
				),
				parameters.NewParameterDefinition(
					"system-db",
					parameters.ParameterTypeString,
					parameters.WithHelp("SQLite database path for system operations (execution logs, request logs)"),
					parameters.WithDefault("system.sqlite"),
				),
				parameters.NewParameterDefinition(
					"scripts",
					parameters.ParameterTypeString,
					parameters.WithHelp("Directory containing JavaScript files to load on startup"),
					parameters.WithDefault(""),
					parameters.WithShortFlag("s"),
				),
			),
			// Add Geppetto layers for AI configuration
			cmds.WithLayersList(geppettoLayers...),
		),
	}, nil
}

// Run implements the BareCommand interface
func (c *ServeCmd) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	log.Info().Msg("Starting JavaScript playground server")

	// Parse settings from layers
	s := &ServeSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to parse serve settings")
	}

	// Create StepSettings from parsed Geppetto layers for AI integration
	stepSettings, err := settings.NewStepSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "failed to create step settings from parsed layers")
	}

	log.Debug().Interface("settings", stepSettings).Msg("Loaded AI step settings for JavaScript engine")

	// Find free ports
	requestedPort, err := strconv.Atoi(s.Port)
	if err != nil {
		return errors.Wrapf(err, "invalid port number: %s", s.Port)
	}

	actualPort, err := findFreePort(requestedPort)
	if err != nil {
		return errors.Wrap(err, "failed to find free port")
	}

	if actualPort != requestedPort {
		log.Info().Int("requested_port", requestedPort).Int("actual_port", actualPort).Msg("Requested port was unavailable, using alternative port")
	}

	requestedAdminPort, err := strconv.Atoi(s.AdminPort)
	if err != nil {
		return errors.Wrapf(err, "invalid admin port number: %s", s.AdminPort)
	}

	actualAdminPort, err := findFreePort(requestedAdminPort)
	if err != nil {
		return errors.Wrap(err, "failed to find free admin port")
	}

	if actualAdminPort != requestedAdminPort {
		log.Info().Int("requested_admin_port", requestedAdminPort).Int("actual_admin_port", actualAdminPort).Msg("Requested admin port was unavailable, using alternative port")
	}

	// Ensure scripts directory exists
	if err := os.MkdirAll("scripts", 0755); err != nil {
		return errors.Wrap(err, "failed to create scripts directory")
	}
	log.Debug().Msg("Scripts directory ready")

	// Initialize JavaScript engine with enhanced Geppetto integration
	log.Debug().Str("appDatabase", s.AppDB).Str("systemDatabase", s.SystemDB).Msg("Initializing JavaScript engine")
	jsEngine := engine.NewEngine(s.AppDB, s.SystemDB)
	
	// Enhanced: Pass stepSettings to engine for better AI integration
	// This would require updating the engine to accept stepSettings
	// For now, we'll keep the existing bootstrap initialization
	if err := jsEngine.Init("bootstrap.js"); err != nil {
		log.Warn().Err(err).Msg("Failed to load bootstrap.js")
	}

	// Start dispatcher goroutine
	log.Debug().Msg("Starting JavaScript dispatcher")
	jsEngine.StartDispatcher()

	// Give dispatcher time to start
	time.Sleep(100 * time.Millisecond)

	// Load scripts from directory if specified
	if s.ScriptsDir != "" {
		log.Info().Str("directory", s.ScriptsDir).Msg("Loading scripts from directory")
		if err := loadScriptsFromDir(jsEngine, s.ScriptsDir); err != nil {
			return errors.Wrapf(err, "failed to load scripts from directory: %s", s.ScriptsDir)
		}
		log.Info().Msg("Finished loading scripts")
	}

	// Setup HTTP routers
	log.Debug().Msg("Setting up HTTP routers")

	// JS Server router (user-facing, JavaScript endpoints)
	jsRouter := web.SetupJSRoutes(jsEngine)

	// Admin router (system interface, playground, API)
	adminRouter := web.SetupRoutesWithAPI(jsEngine, api.ExecuteHandler(jsEngine))
	log.Debug().Msg("Registered API endpoint: POST /v1/execute")

	// Configure server addresses
	jsAddr := ":" + strconv.Itoa(actualPort)
	adminAddr := ":" + strconv.Itoa(actualAdminPort)
	jsBaseURL := fmt.Sprintf("http://localhost:%d", actualPort)
	adminBaseURL := fmt.Sprintf("http://localhost:%d", actualAdminPort)

	log.Info().
		Str("js_address", jsAddr).
		Str("admin_address", adminAddr).
		Str("app_database", s.AppDB).
		Str("system_database", s.SystemDB).
		Msg("Server configuration")

	if s.ScriptsDir != "" {
		log.Info().Str("scripts", s.ScriptsDir).Msg("Scripts directory configured")
	}

	log.Info().Str("execute_endpoint", adminBaseURL+"/v1/execute").Msg("API endpoint ready")
	log.Info().Str("js_server", jsBaseURL).Msg("JavaScript web server available")
	log.Info().Str("admin_interface", adminBaseURL).Msg("Admin interface available")
	log.Info().Str("admin_logs", adminBaseURL+"/admin/logs").Msg("Admin logs available")

	// Start servers concurrently
	log.Info().Str("js_address", jsAddr).Msg("Starting JavaScript web server")
	go func() {
		if err := http.ListenAndServe(jsAddr, jsRouter); err != nil {
			log.Fatal().Err(err).Msg("JavaScript web server failed")
		}
	}()

	log.Info().Str("admin_address", adminAddr).Msg("Starting admin interface server")
	if err := http.ListenAndServe(adminAddr, adminRouter); err != nil {
		return errors.Wrap(err, "admin interface server failed")
	}

	return nil
}

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

// loadScriptsFromDir loads JavaScript files from a directory
func loadScriptsFromDir(jsEngine *engine.Engine, dir string) error {
	log.Info().Str("directory", dir).Msg("Loading JavaScript files")

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error accessing file")
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".js") {
			log.Info().Str("file", path).Msg("Loading JavaScript file")
			data, err := os.ReadFile(path)
			if err != nil {
				log.Error().Err(err).Str("file", path).Msg("Failed to read file")
				return nil // Continue with other files
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
}

