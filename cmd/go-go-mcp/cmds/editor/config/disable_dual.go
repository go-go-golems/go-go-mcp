package config

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

// DisableCommand is a dual Glazed command that implements both BareCommand and GlazeCommand interfaces
type DisableCommand struct {
	*cmds.CommandDescription
}

// DisableSettings holds the parameters for the disable command
type DisableSettings struct {
	Editors    string `glazed.parameter:"editors"`
	ServerName string `glazed.parameter:"server-name"`
	Target     string `glazed.parameter:"target"`
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &DisableCommand{}
var _ cmds.GlazeCommand = &DisableCommand{}

// Run implements BareCommand interface for human-readable output
func (c *DisableCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &DisableSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	// Execute for each editor
	stores, err := NewStoresWithTarget(editors, settings.Target)
	if err != nil {
		return err
	}

	var successCount, failCount int
	var results []string

	for _, editorStore := range stores {
		// Check if server exists
		_, exists, err := editorStore.Store.GetServer(settings.ServerName)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ %s: error checking server: %v", editorStore.Editor, err))
			failCount++
			continue
		}
		if !exists {
			results = append(results, fmt.Sprintf("⚠️  %s: server '%s' not found", editorStore.Editor, settings.ServerName))
			failCount++
			continue
		}

		// Check if server is already disabled
		disabled, err := editorStore.Store.IsServerDisabled(settings.ServerName)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ %s: error checking status: %v", editorStore.Editor, err))
			failCount++
			continue
		}
		if disabled {
			results = append(results, fmt.Sprintf("ℹ️  %s: server '%s' already disabled", editorStore.Editor, settings.ServerName))
			successCount++ // Count as success since desired state is achieved
			continue
		}

		if err := editorStore.Store.DisableServer(settings.ServerName); err != nil {
			results = append(results, fmt.Sprintf("❌ %s: %v", editorStore.Editor, err))
			failCount++
			continue
		}

		if err := editorStore.Store.Save(); err != nil {
			results = append(results, fmt.Sprintf("❌ %s: save failed: %v", editorStore.Editor, err))
			failCount++
			continue
		}

		results = append(results, fmt.Sprintf("✅ %s: disabled server '%s'", editorStore.Editor, settings.ServerName))
		successCount++
	}

	// Print results
	if len(editors) == 1 {
		// Single editor: use existing format for backwards compatibility
		if successCount > 0 {
			fmt.Printf("Successfully disabled MCP server '%s'\n", settings.ServerName)
		} else {
			return fmt.Errorf("failed to disable server: %s", results[0][2:]) // Remove emoji prefix
		}
	} else {
		// Multiple editors: show results for each
		fmt.Printf("Multi-editor disable results for server '%s':\n\n", settings.ServerName)
		for _, result := range results {
			fmt.Println(result)
		}
		fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failCount)
	}

	// Return error if all failed, but allow partial success
	if successCount == 0 {
		return fmt.Errorf("failed to disable server in any editor")
	}

	return nil
}

// RunIntoGlazeProcessor implements GlazeCommand interface for structured output
func (c *DisableCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &DisableSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	// Execute for each editor
	stores, err := NewStoresWithTarget(editors, settings.Target)
	if err != nil {
		return err
	}

	for _, editorStore := range stores {
		var success bool
		var errorMsg string
		var previousStatus, newStatus string

		// Check if server exists
		_, exists, err := editorStore.Store.GetServer(settings.ServerName)
		if err != nil {
			success = false
			errorMsg = fmt.Sprintf("error checking server: %v", err)
			previousStatus = "unknown"
			newStatus = "unknown"
		} else if !exists {
			success = false
			errorMsg = fmt.Sprintf("server '%s' not found", settings.ServerName)
			previousStatus = "not_found"
			newStatus = "not_found"
		} else {
			// Check if server is already disabled
			disabled, err := editorStore.Store.IsServerDisabled(settings.ServerName)
			if err != nil {
				success = false
				errorMsg = fmt.Sprintf("error checking status: %v", err)
				previousStatus = "unknown"
				newStatus = "unknown"
			} else {
				previousStatus = "enabled"
				if disabled {
					previousStatus = "disabled"
				}

				if disabled {
					// Already disabled
					success = true
					newStatus = "disabled"
					errorMsg = "already disabled"
				} else {
					// Try to disable
					err = editorStore.Store.DisableServer(settings.ServerName)
					if err != nil {
						success = false
						errorMsg = err.Error()
						newStatus = "enabled"
					} else {
						err = editorStore.Store.Save()
						if err != nil {
							success = false
							errorMsg = fmt.Sprintf("save failed: %v", err)
							newStatus = "enabled"
						} else {
							success = true
							newStatus = "disabled"
						}
					}
				}
			}
		}

		// Create row for this operation
		row := types.NewRow(
			types.MRP("editor", editorStore.Editor),
			types.MRP("server_name", settings.ServerName),
			types.MRP("operation", "disable"),
			types.MRP("previous_status", previousStatus),
			types.MRP("new_status", newStatus),
			types.MRP("target", settings.Target),
			types.MRP("success", success),
			types.MRP("error", errorMsg),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add row for editor %s: %w", editorStore.Editor, err)
		}
	}

	return nil
}

// NewDisableCommandDual creates a new dual-mode disable command
func NewDisableCommandDual() (*DisableCommand, error) {
	// Create glazed layer for output formatting options
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command settings layer for debugging features
	commandSettingsLayer, err := cli.NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	// Define command with parameters
	cmdDesc := cmds.NewCommandDescription(
		"disable",
		cmds.WithShort("Disable MCP server in one or more editors"),
		cmds.WithLong(`Disable an MCP server in one or more editors without removing it.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

The disable operation will:
- Check if the server exists
- Check its current status (enabled/disabled)
- Disable the server if currently enabled
- Report success if already disabled

Examples:
  # Human-readable output (default)
  `+"```"+`
  mcp editor config claude disable myserver
  mcp editor config claude,cursor disable myserver --target global
  `+"```"+`
  
  # Structured output
  `+"```"+`
  mcp editor config claude disable myserver --with-structured-output --output json
  mcp editor config claude,cursor,amp disable myserver --with-structured-output --output table
  `+"```"+``),

		// Define command arguments
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"editors",
				parameters.ParameterTypeString,
				parameters.WithHelp("Editor(s) to disable server in (comma-separated)"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"server-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the MCP server to disable"),
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

		// Add glazed and command settings layers
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)

	return &DisableCommand{
		CommandDescription: cmdDesc,
	}, nil
}
