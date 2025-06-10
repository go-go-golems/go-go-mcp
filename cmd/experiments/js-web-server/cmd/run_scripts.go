package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	pinocchio_cmds "github.com/go-go-golems/pinocchio/pkg/cmds"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// RunScriptsCmd represents the run-scripts command
type RunScriptsCmd struct {
	*cmds.CommandDescription
}

// RunScriptsSettings holds the configuration for the run-scripts command
type RunScriptsSettings struct {
	ScriptsDir string   `glazed.parameter:"scripts"`
	Files      []string `glazed.parameter:"files"`
}

// Ensure RunScriptsCmd implements BareCommand
var _ cmds.BareCommand = &RunScriptsCmd{}

// NewRunScriptsCmd creates a new run-scripts command
func NewRunScriptsCmd() (*RunScriptsCmd, error) {
	// Create temporary step settings for Geppetto layers
	tempSettings, err := settings.NewStepSettings()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create step settings")
	}

	// Create Geppetto layers
	geppettoLayers, err := pinocchio_cmds.CreateGeppettoLayers(tempSettings, pinocchio_cmds.WithHelpersLayer())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Geppetto layers")
	}

	// Create default layer for run-scripts specific settings
	defaultLayer, err := layers.NewParameterLayer(
		layers.DefaultSlug,
		"Run Scripts Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"scripts",
				parameters.ParameterTypeString,
				parameters.WithHelp("Directory containing JavaScript files to execute"),
				parameters.WithShortFlag("s"),
				parameters.WithDefault("./scripts"),
			),
			parameters.NewParameterDefinition(
				"files",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Specific JavaScript files to execute (if not provided, all .js files in scripts directory)"),
				parameters.WithShortFlag("f"),
			),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default layer")
	}

	// Combine all layers
	allLayers := append(geppettoLayers, defaultLayer)

	return &RunScriptsCmd{
		CommandDescription: cmds.NewCommandDescription(
			"run-scripts",
			cmds.WithShort("Execute JavaScript files with Geppetto AI capabilities"),
			cmds.WithLong(`Execute JavaScript files with Geppetto AI capabilities without starting the web server.

This command loads and executes JavaScript files in a Geppetto-enabled environment, 
providing access to Conversation and ChatStepFactory APIs for AI interactions.

The command is useful for:
• Running test scripts
• Executing batch AI operations  
• Testing Geppetto API functionality
• Running standalone JavaScript with AI capabilities

Examples:
  run-scripts --scripts ./tests
  run-scripts --files test1.js,test2.js
  run-scripts --profile 4o-mini --scripts ./ai-tests
  run-scripts --profile claude-dev --files inference_test.js`),
			cmds.WithLayers(layers.NewParameterLayers(layers.WithLayers(allLayers...))),
		),
	}, nil
}

// Run executes the run-scripts command
func (cmd *RunScriptsCmd) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Parse settings
	var runSettings RunScriptsSettings
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, &runSettings)
	if err != nil {
		return errors.Wrap(err, "failed to parse settings")
	}

	// Create step settings from parsed layers
	stepSettings, err := settings.NewStepSettings()
	if err != nil {
		return errors.Wrap(err, "failed to create step settings")
	}

	err = stepSettings.UpdateFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "failed to update step settings from parsed layers")
	}

	log.Info().Str("scripts_dir", runSettings.ScriptsDir).Msg("Starting JavaScript script execution")

	// Initialize JavaScript engine with in-memory databases (since we don't need persistence for script execution)
	jsEngine := engine.NewEngine(":memory:", ":memory:")
	defer func() { _ = jsEngine.Close() }()

	// Update engine with our step settings for AI capabilities
	err = jsEngine.UpdateStepSettings(stepSettings)
	if err != nil {
		return errors.Wrap(err, "failed to update step settings")
	}

	// Determine which files to execute
	var filesToExecute []string

	if len(runSettings.Files) > 0 {
		// Execute specific files
		filesToExecute = runSettings.Files
		log.Info().Strs("files", runSettings.Files).Msg("Executing specific files")
	} else {
		// Execute all .js files in the scripts directory
		if _, err := os.Stat(runSettings.ScriptsDir); os.IsNotExist(err) {
			return errors.Errorf("scripts directory does not exist: %s", runSettings.ScriptsDir)
		}

		err := filepath.Walk(runSettings.ScriptsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".js" {
				filesToExecute = append(filesToExecute, path)
			}
			return nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to scan scripts directory")
		}

		log.Info().Int("file_count", len(filesToExecute)).Str("directory", runSettings.ScriptsDir).Msg("Found JavaScript files")
	}

	if len(filesToExecute) == 0 {
		log.Warn().Msg("No JavaScript files found to execute")
		return nil
	}

	// Execute each file
	for _, filePath := range filesToExecute {
		log.Info().Str("file", filePath).Msg("Executing JavaScript file")

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Err(err).Str("file", filePath).Msg("Failed to read file")
			continue
		}

		// Execute the script and capture results
		result, err := jsEngine.ExecuteScript(string(content))
		if err != nil {
			log.Error().Err(err).Str("file", filePath).Msg("Failed to execute file")
			continue
		}

		log.Info().Str("file", filePath).Msg("Successfully executed JavaScript file")
		if result.Value != nil {
			log.Info().Interface("result", result.Value).Str("file", filePath).Msg("Script result")
		}
		if len(result.ConsoleLog) > 0 {
			for _, logLine := range result.ConsoleLog {
				log.Info().Str("console", logLine).Str("file", filePath).Msg("Script console output")
			}
		}
		if result.Error != nil {
			log.Error().Err(result.Error).Str("file", filePath).Msg("Script execution error")
		}
	}

	log.Info().Msg("JavaScript script execution completed")
	return nil
}
