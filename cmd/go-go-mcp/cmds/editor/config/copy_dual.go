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
	"github.com/go-go-golems/go-go-mcp/pkg/core/configstore"
)

// CopyCommand is a dual Glazed command that implements both BareCommand and GlazeCommand interfaces
type CopyCommand struct {
	*cmds.CommandDescription
}

// CopySettings holds the parameters for the copy command
type CopySettings struct {
	Editors      string `glazed.parameter:"editors"`
	SourceName   string `glazed.parameter:"source-name"`
	DestName     string `glazed.parameter:"dest-name"`
	Target       string `glazed.parameter:"target"`
	TargetEditor string `glazed.parameter:"target-editor"`
	Overwrite    bool   `glazed.parameter:"overwrite"`
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &CopyCommand{}
var _ cmds.GlazeCommand = &CopyCommand{}

// Run implements BareCommand interface for human-readable output
func (c *CopyCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &CopySettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	// For single editor, handle simple copy
	if len(editors) == 1 && settings.TargetEditor == "" {
		return c.runSingleEditorCopy(editors[0], settings)
	}

	// For cross-editor or multi-editor copies
	return c.runMultiEditorCopy(editors, settings)
}

// runSingleEditorCopy handles copying within a single editor
func (c *CopyCommand) runSingleEditorCopy(editor string, settings *CopySettings) error {
	sourceStore, err := NewStoreWithTarget(editor, settings.Target)
	if err != nil {
		return err
	}

	// Get the source server
	sourceServer, exists, err := sourceStore.GetServer(settings.SourceName)
	if err != nil {
		return fmt.Errorf("failed to get source server: %w", err)
	}
	if !exists {
		return fmt.Errorf("source server '%s' not found", settings.SourceName)
	}

	// Check if source server is disabled
	isDisabled, err := sourceStore.IsServerDisabled(settings.SourceName)
	if err != nil {
		return fmt.Errorf("failed to check source server status: %w", err)
	}

	// Create destination server with new name
	destServer := sourceServer
	destServer.Name = settings.DestName

	// Add the server to destination
	if err := sourceStore.AddServer(destServer, settings.Overwrite); err != nil {
		return fmt.Errorf("failed to add destination server: %w", err)
	}

	// If source was disabled, disable the destination as well
	if isDisabled {
		if err := sourceStore.DisableServer(settings.DestName); err != nil {
			return fmt.Errorf("failed to disable destination server: %w", err)
		}
	}

	// Save the store
	if err := sourceStore.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Print success message
	action := "Copied"
	if settings.Overwrite {
		action = "Copied (overwritten)"
	}

	fmt.Printf("Successfully %s MCP server '%s' to '%s':\n", action, settings.SourceName, settings.DestName)

	// Show source details
	fmt.Printf("  Source (%s):\n", settings.SourceName)
	printServerDetails(sourceServer, "    ")

	// Show destination details
	fmt.Printf("  Destination (%s):\n", settings.DestName)
	printServerDetails(destServer, "    ")

	// Show status
	statusText := "enabled"
	if isDisabled {
		statusText = "disabled"
	}
	fmt.Printf("  Status: %s\n", statusText)

	return nil
}

// runMultiEditorCopy handles cross-editor or multi-editor copies
func (c *CopyCommand) runMultiEditorCopy(editors []string, settings *CopySettings) error {
	var results []string
	var successCount, failCount int

	for _, editor := range editors {
		sourceStore, err := NewStoreWithTarget(editor, settings.Target)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ %s: failed to create store: %v", editor, err))
			failCount++
			continue
		}

		// Get the source server
		sourceServer, exists, err := sourceStore.GetServer(settings.SourceName)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ %s: failed to get source server: %v", editor, err))
			failCount++
			continue
		}
		if !exists {
			results = append(results, fmt.Sprintf("❌ %s: source server '%s' not found", editor, settings.SourceName))
			failCount++
			continue
		}

		// Check if source server is disabled
		isDisabled, err := sourceStore.IsServerDisabled(settings.SourceName)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ %s: failed to check source server status: %v", editor, err))
			failCount++
			continue
		}

		// Determine destination store
		var destStore configstore.Store
		var destEditorName string
		if settings.TargetEditor != "" {
			// Cross-editor copy
			destStore, err = NewStoreWithTarget(settings.TargetEditor, settings.Target)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %s: failed to create destination store: %v", editor, err))
				failCount++
				continue
			}
			destEditorName = settings.TargetEditor

			// Validate transport compatibility for cross-editor copy
			if err := validateTransportCompatibility(editor, settings.TargetEditor, sourceServer); err != nil {
				results = append(results, fmt.Sprintf("❌ %s: transport incompatible: %v", editor, err))
				failCount++
				continue
			}
		} else {
			// Same editor copy
			destStore = sourceStore
			destEditorName = editor
		}

		// Create destination server with new name
		destServer := sourceServer
		destServer.Name = settings.DestName

		// Add the server to destination
		if err := destStore.AddServer(destServer, settings.Overwrite); err != nil {
			results = append(results, fmt.Sprintf("❌ %s: failed to add destination server: %v", destEditorName, err))
			failCount++
			continue
		}

		// If source was disabled, disable the destination as well
		if isDisabled {
			if err := destStore.DisableServer(settings.DestName); err != nil {
				results = append(results, fmt.Sprintf("❌ %s: failed to disable destination server: %v", destEditorName, err))
				failCount++
				continue
			}
		}

		// Save the destination store
		if err := destStore.Save(); err != nil {
			results = append(results, fmt.Sprintf("❌ %s: failed to save configuration: %v", destEditorName, err))
			failCount++
			continue
		}

		// Success
		action := "Copied"
		if settings.Overwrite {
			action = "Copied (overwritten)"
		}

		if settings.TargetEditor != "" {
			results = append(results, fmt.Sprintf("✅ %s → %s: %s server '%s' to '%s'", editor, destEditorName, action, settings.SourceName, settings.DestName))
		} else {
			results = append(results, fmt.Sprintf("✅ %s: %s server '%s' to '%s'", editor, action, settings.SourceName, settings.DestName))
		}
		successCount++
	}

	// Print results
	if settings.TargetEditor != "" {
		fmt.Printf("Cross-editor copy results for server '%s' → '%s':\n\n", settings.SourceName, settings.DestName)
	} else {
		fmt.Printf("Multi-editor copy results for server '%s' → '%s':\n\n", settings.SourceName, settings.DestName)
	}

	for _, result := range results {
		fmt.Println(result)
	}
	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failCount)

	// Return error if all failed
	if successCount == 0 {
		return fmt.Errorf("failed to copy server to any editor")
	}

	return nil
}

// RunIntoGlazeProcessor implements GlazeCommand interface for structured output
func (c *CopyCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &CopySettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	for _, editor := range editors {
		sourceStore, err := NewStoreWithTarget(editor, settings.Target)
		if err != nil {
			// Add error row
			row := c.createErrorRow(editor, "", settings.SourceName, settings.DestName, "", "", nil, false, err.Error())
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add error row for editor %s: %w", editor, err)
			}
			continue
		}

		// Get the source server
		sourceServer, exists, err := sourceStore.GetServer(settings.SourceName)
		if err != nil || !exists {
			errorMsg := err.Error()
			if !exists {
				errorMsg = fmt.Sprintf("source server '%s' not found", settings.SourceName)
			}
			row := c.createErrorRow(editor, "", settings.SourceName, settings.DestName, "", "", nil, false, errorMsg)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add error row for editor %s: %w", editor, err)
			}
			continue
		}

		// Check if source server is disabled
		isDisabled, err := sourceStore.IsServerDisabled(settings.SourceName)
		if err != nil {
			row := c.createErrorRow(editor, "", settings.SourceName, settings.DestName, "", "", nil, false, fmt.Sprintf("failed to check source server status: %v", err))
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add error row for editor %s: %w", editor, err)
			}
			continue
		}

		// Determine destination store and transport
		var destStore configstore.Store
		var destEditorName string
		var transport string

		if sourceServer.URL != "" {
			if sourceServer.IsSSE {
				transport = "SSE"
			} else {
				transport = "HTTP"
			}
		} else {
			transport = "stdio"
		}

		if settings.TargetEditor != "" {
			// Cross-editor copy
			destStore, err = NewStoreWithTarget(settings.TargetEditor, settings.Target)
			if err != nil {
				row := c.createErrorRow(editor, settings.TargetEditor, settings.SourceName, settings.DestName, transport, sourceServer.Command, sourceServer.Args, !isDisabled, fmt.Sprintf("failed to create destination store: %v", err))
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to add error row for editor %s: %w", editor, err)
				}
				continue
			}
			destEditorName = settings.TargetEditor

			// Validate transport compatibility for cross-editor copy
			if err := validateTransportCompatibility(editor, settings.TargetEditor, sourceServer); err != nil {
				row := c.createErrorRow(editor, settings.TargetEditor, settings.SourceName, settings.DestName, transport, sourceServer.Command, sourceServer.Args, !isDisabled, fmt.Sprintf("transport incompatible: %v", err))
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to add error row for editor %s: %w", editor, err)
				}
				continue
			}
		} else {
			// Same editor copy
			destStore = sourceStore
			destEditorName = editor
		}

		// Create destination server with new name
		destServer := sourceServer
		destServer.Name = settings.DestName

		var success bool
		var errorMsg string

		// Add the server to destination
		err = destStore.AddServer(destServer, settings.Overwrite)
		if err != nil {
			success = false
			errorMsg = fmt.Sprintf("failed to add destination server: %v", err)
		} else {
			// If source was disabled, disable the destination as well
			if isDisabled {
				err = destStore.DisableServer(settings.DestName)
				if err != nil {
					success = false
					errorMsg = fmt.Sprintf("failed to disable destination server: %v", err)
				}
			}

			if err == nil {
				// Save the destination store
				err = destStore.Save()
				if err != nil {
					success = false
					errorMsg = fmt.Sprintf("failed to save configuration: %v", err)
				} else {
					success = true
				}
			}
		}

		// Create row for this operation
		row := types.NewRow(
			types.MRP("source_editor", editor),
			types.MRP("dest_editor", destEditorName),
			types.MRP("source_name", settings.SourceName),
			types.MRP("dest_name", settings.DestName),
			types.MRP("command", sourceServer.Command),
			types.MRP("args", sourceServer.Args),
			types.MRP("env", sourceServer.Env),
			types.MRP("url", sourceServer.URL),
			types.MRP("transport", transport),
			types.MRP("enabled", !isDisabled),
			types.MRP("success", success),
			types.MRP("error", errorMsg),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add row for editor %s: %w", editor, err)
		}
	}

	return nil
}

// createErrorRow creates a row for error cases
func (c *CopyCommand) createErrorRow(sourceEditor, destEditor, sourceName, destName, transport, command string, args []string, enabled bool, errorMsg string) types.Row {
	if destEditor == "" {
		destEditor = sourceEditor
	}

	return types.NewRow(
		types.MRP("source_editor", sourceEditor),
		types.MRP("dest_editor", destEditor),
		types.MRP("source_name", sourceName),
		types.MRP("dest_name", destName),
		types.MRP("command", command),
		types.MRP("args", args),
		types.MRP("env", map[string]string{}),
		types.MRP("url", ""),
		types.MRP("transport", transport),
		types.MRP("enabled", enabled),
		types.MRP("success", false),
		types.MRP("error", errorMsg),
	)
}

// NewCopyCommandDual creates a new dual-mode copy command
func NewCopyCommandDual() (*CopyCommand, error) {
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
		"copy",
		cmds.WithShort("Copy MCP server configuration within or across editors"),
		cmds.WithLong(`Copy an MCP server configuration to a new name within the same editor or across different editors.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

The copy preserves all server configuration including command, arguments, 
environment variables, URL, transport type, and enabled/disabled status.

Cross-editor copying:
Use --target-editor to copy from source editor(s) to another editor.
Note: Some transport types may not be compatible between editors.
For example, Claude only supports stdio transport.

If a server with the destination name already exists, the command will fail 
unless --overwrite is specified.

Examples:
  # Human-readable output (default)
  mcp editor config claude copy server1 server2
  mcp editor config claude copy server1 server2 --target-editor cursor
  
  # Structured output
  mcp editor config claude copy server1 server2 --with-structured-output --output json
  
  # Cross-editor with structured output
  mcp editor config claude copy server1 server2 --target-editor cursor --with-structured-output --output table`),

		// Define command arguments
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"editors",
				parameters.ParameterTypeString,
				parameters.WithHelp("Editor(s) containing source server (comma-separated)"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"source-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the source MCP server"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"dest-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name for the destination MCP server"),
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
			parameters.NewParameterDefinition(
				"target-editor",
				parameters.ParameterTypeString,
				parameters.WithHelp("Copy to a different editor (claude, cursor, ampcode, amp, crush)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"overwrite",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Overwrite existing server if it exists"),
				parameters.WithShortFlag("w"),
				parameters.WithDefault(false),
			),
		),

		// Add glazed and command settings layers
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)

	return &CopyCommand{
		CommandDescription: cmdDesc,
	}, nil
}
