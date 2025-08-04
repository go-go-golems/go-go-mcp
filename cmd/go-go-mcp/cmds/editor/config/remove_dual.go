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

// RemoveCommand is a dual Glazed command that implements both BareCommand and GlazeCommand interfaces
type RemoveCommand struct {
	*cmds.CommandDescription
}

// RemoveSettings holds the parameters for the remove command
type RemoveSettings struct {
	Editors    string `glazed.parameter:"editors"`
	ServerName string `glazed.parameter:"server-name"`
	Target     string `glazed.parameter:"target"`
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &RemoveCommand{}
var _ cmds.GlazeCommand = &RemoveCommand{}

// Run implements BareCommand interface for human-readable output
func (c *RemoveCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &RemoveSettings{}
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
		// Check if server exists before removing
		server, exists, err := editorStore.Store.GetServer(settings.ServerName)
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

		// Show server details before removal (for confirmation in logs)
		if len(editors) > 1 {
			fmt.Printf("📋 %s: removing server '%s' (command: %s)\n",
				editorStore.Editor, settings.ServerName, server.Command)
		}

		if err := editorStore.Store.RemoveServer(settings.ServerName); err != nil {
			results = append(results, fmt.Sprintf("❌ %s: %v", editorStore.Editor, err))
			failCount++
			continue
		}

		if err := editorStore.Store.Save(); err != nil {
			results = append(results, fmt.Sprintf("❌ %s: save failed: %v", editorStore.Editor, err))
			failCount++
			continue
		}

		results = append(results, fmt.Sprintf("✅ %s: removed server '%s'", editorStore.Editor, settings.ServerName))
		successCount++
	}

	// Print results
	if len(editors) == 1 {
		// Single editor: use existing format for backwards compatibility
		if successCount > 0 {
			fmt.Printf("Successfully removed MCP server '%s'\n", settings.ServerName)
		} else {
			return fmt.Errorf("failed to remove server: %s", results[0][2:]) // Remove emoji prefix
		}
	} else {
		// Multiple editors: show results for each
		fmt.Printf("Multi-editor remove results for server '%s':\n\n", settings.ServerName)
		for _, result := range results {
			fmt.Println(result)
		}
		fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failCount)
	}

	// Return error if all failed, but allow partial success
	if successCount == 0 {
		return fmt.Errorf("failed to remove server from any editor")
	}

	return nil
}

// RunIntoGlazeProcessor implements GlazeCommand interface for structured output
func (c *RemoveCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &RemoveSettings{}
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
		var serverDetails map[string]interface{}

		// Check if server exists before removing
		server, exists, err := editorStore.Store.GetServer(settings.ServerName)
		if err != nil {
			success = false
			errorMsg = fmt.Sprintf("error checking server: %v", err)
			serverDetails = map[string]interface{}{
				"command":     "",
				"args":        []string{},
				"env":         map[string]string{},
				"url":         "",
				"transport":   "",
				"was_enabled": false,
			}
		} else if !exists {
			success = false
			errorMsg = fmt.Sprintf("server '%s' not found", settings.ServerName)
			serverDetails = map[string]interface{}{
				"command":     "",
				"args":        []string{},
				"env":         map[string]string{},
				"url":         "",
				"transport":   "",
				"was_enabled": false,
			}
		} else {
			// Capture server details before removal
			wasDisabled, _ := editorStore.Store.IsServerDisabled(settings.ServerName)

			// Determine transport type
			transport := "stdio"
			if server.URL != "" {
				if server.IsSSE {
					transport = "sse"
				} else {
					transport = "http"
				}
			}

			serverDetails = map[string]interface{}{
				"command":     server.Command,
				"args":        server.Args,
				"env":         server.Env,
				"url":         server.URL,
				"transport":   transport,
				"was_enabled": !wasDisabled,
			}

			// Try to remove
			err = editorStore.Store.RemoveServer(settings.ServerName)
			if err != nil {
				success = false
				errorMsg = err.Error()
			} else {
				err = editorStore.Store.Save()
				if err != nil {
					success = false
					errorMsg = fmt.Sprintf("save failed: %v", err)
				} else {
					success = true
				}
			}
		}

		// Create row for this operation
		row := types.NewRow(
			types.MRP("editor", editorStore.Editor),
			types.MRP("server_name", settings.ServerName),
			types.MRP("command", serverDetails["command"]),
			types.MRP("args", serverDetails["args"]),
			types.MRP("env", serverDetails["env"]),
			types.MRP("url", serverDetails["url"]),
			types.MRP("transport", serverDetails["transport"]),
			types.MRP("was_enabled", serverDetails["was_enabled"]),
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

// NewRemoveCommandDual creates a new dual-mode remove command
func NewRemoveCommandDual() (*RemoveCommand, error) {
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
		"remove",
		cmds.WithShort("Remove MCP server from one or more editors"),
		cmds.WithLong(`Remove an MCP server configuration from one or more editors.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

The remove operation will:
- Check if the server exists
- Display server configuration details (for multi-editor operations)
- Remove the server configuration completely
- Save the updated configuration

Examples:
  # Human-readable output (default)
  `+"```"+`
  mcp editor config claude remove myserver
  mcp editor config claude,cursor remove myserver --target global
  `+"```"+`
  
  # Structured output
  `+"```"+`
  mcp editor config claude remove myserver --with-structured-output --output json
  mcp editor config claude,cursor,amp remove myserver --with-structured-output --output table
  `+"```"+``),

		// Define command arguments
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"editors",
				parameters.ParameterTypeString,
				parameters.WithHelp("Editor(s) to remove server from (comma-separated)"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"server-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the MCP server to remove"),
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

	return &RemoveCommand{
		CommandDescription: cmdDesc,
	}, nil
}
