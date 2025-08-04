package config

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// InitCommand is a Glazed BareCommand for initializing config files
type InitCommand struct {
	*cmds.CommandDescription
}

// InitSettings holds the parameters for the init command
type InitSettings struct {
	Editors string `glazed.parameter:"editors"`
	Target  string `glazed.parameter:"target"`
}

// Ensure interface is implemented
var _ cmds.BareCommand = &InitCommand{}

// Run implements BareCommand interface
func (c *InitCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &InitSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	// Handle multiple editors
	if len(editors) == 1 {
		// Single editor - use original behavior for backwards compatibility
		store, err := NewStoreWithTarget(editors[0], settings.Target)
		if err != nil {
			return err
		}

		if err := store.Save(); err != nil {
			return fmt.Errorf("failed to initialize configuration: %w", err)
		}

		configPath := store.GetConfigPath()
		fmt.Printf("Successfully initialized configuration file: %s\n", configPath)
		return nil
	} else {
		// Multiple editors - initialize each and show results
		fmt.Printf("Initializing config files for %d editors...\n\n", len(editors))
		
		var successCount, failCount int
		var results []string
		
		for _, editor := range editors {
			store, err := NewStoreWithTarget(editor, settings.Target)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %s: %v", editor, err))
				failCount++
				continue
			}

			// Check if file already exists
			configPath := store.GetConfigPath()
			fileExists := true
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fileExists = false
			}

			if err := store.Save(); err != nil {
				results = append(results, fmt.Sprintf("❌ %s: failed to initialize: %v", editor, err))
				failCount++
				continue
			}

			if fileExists {
				results = append(results, fmt.Sprintf("✅ %s: configuration already exists at %s", editor, configPath))
			} else {
				results = append(results, fmt.Sprintf("✅ %s: created configuration at %s", editor, configPath))
			}
			successCount++
		}
		
		// Print results
		for _, result := range results {
			fmt.Println(result)
		}
		
		fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failCount)
		
		// Return error if all failed
		if successCount == 0 {
			return fmt.Errorf("failed to initialize configuration for any editor")
		}
		
		return nil
	}
}

// NewInitCommand creates a new init command
func NewInitCommand() (cmds.BareCommand, error) {
	cmd := &InitCommand{
		CommandDescription: cmds.NewCommandDescription(
			"init",
			cmds.WithShort("Initialize config"),
			cmds.WithLong(`Initialize configuration files for one or more editors.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

If a configuration file already exists, it will not be overwritten.

Examples:
  mcp editor config claude init
  mcp editor config claude,cursor init --target global`),
			
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"editors",
					parameters.ParameterTypeString,
					parameters.WithHelp("Editor(s) to initialize config for (comma-separated)"),
					parameters.WithRequired(true),
				),
			),
			
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"target",
					parameters.ParameterTypeString,
					parameters.WithHelp("Target configuration (e.g., global, cwd)"),
					parameters.WithShortFlag("t"),
					parameters.WithDefault(""),
				),
			),
		),
	}

	return cmd, nil
}
