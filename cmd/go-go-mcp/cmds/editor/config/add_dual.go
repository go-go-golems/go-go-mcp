package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	mcptypes "github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// AddCommand is a dual Glazed command that implements both BareCommand and GlazeCommand interfaces
type AddCommand struct {
	*cmds.CommandDescription
}

// AddSettings holds the parameters for the add command
type AddSettings struct {
	Editors   string   `glazed.parameter:"editors"`
	Name      string   `glazed.parameter:"name"`
	Command   string   `glazed.parameter:"command"`
	Args      []string `glazed.parameter:"args"`
	Target    string   `glazed.parameter:"target"`
	Env       []string `glazed.parameter:"env"`
	Overwrite bool     `glazed.parameter:"overwrite"`
	URL       string   `glazed.parameter:"url"`
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &AddCommand{}
var _ cmds.GlazeCommand = &AddCommand{}

// Run implements BareCommand interface for human-readable output
func (c *AddCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &AddSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	// Validate arguments
	if settings.Name == "" {
		return fmt.Errorf("server name is required")
	}

	var command string
	var cmdArgs []string

	// Handle different transport types
	if settings.URL != "" {
		// URL-based transport (HTTP/SSE)
		if settings.Command != "" {
			return fmt.Errorf("cannot specify both command and URL")
		}
		command = ""
		cmdArgs = nil
	} else {
		// Stdio transport
		if settings.Command == "" {
			return fmt.Errorf("command required for stdio transport")
		}
		command = settings.Command
		cmdArgs = settings.Args
	}

	// Parse environment variables
	envMap := make(map[string]string)
	for _, e := range settings.Env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid environment variable format: %s (expected KEY=VALUE)", e)
		}
		envMap[parts[0]] = parts[1]
	}

	// Auto-detect transport type
	isSSE := false
	if settings.URL != "" {
		// Simple heuristic: URLs containing "events" or "sse" are SSE
		urlLower := strings.ToLower(settings.URL)
		isSSE = strings.Contains(urlLower, "events") || strings.Contains(urlLower, "sse")
	}

	commonServer := mcptypes.CommonServer{
		Name:    settings.Name,
		Command: command,
		Args:    cmdArgs,
		Env:     envMap,
		URL:     settings.URL,
		IsSSE:   isSSE,
	}

	// Execute for each editor
	stores, err := NewStoresWithTarget(editors, settings.Target)
	if err != nil {
		return err
	}

	var successCount, failCount int
	var results []string

	for _, editorStore := range stores {
		err := editorStore.Store.AddServer(commonServer, settings.Overwrite)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ %s: %v", editorStore.Editor, err))
			failCount++
			continue
		}

		err = editorStore.Store.Save()
		if err != nil {
			results = append(results, fmt.Sprintf("❌ %s: save failed: %v", editorStore.Editor, err))
			failCount++
			continue
		}

		action := "Added"
		if settings.Overwrite {
			action = "Updated"
		}
		results = append(results, fmt.Sprintf("✅ %s: %s server '%s'", editorStore.Editor, action, settings.Name))
		successCount++
	}

	// Print results
	if len(editors) == 1 {
		// Single editor: use existing format for backwards compatibility
		if successCount > 0 {
			action := "Added"
			if settings.Overwrite {
				action = "Updated"
			}
			fmt.Printf("Successfully %s MCP server '%s':\n", action, settings.Name)
			printServerDetails(commonServer, "  ")
		} else {
			return fmt.Errorf("failed to add server: %s", results[0][2:]) // Remove emoji prefix
		}
	} else {
		// Multiple editors: show results for each
		fmt.Printf("Multi-editor add results for server '%s':\n\n", settings.Name)
		for _, result := range results {
			fmt.Println(result)
		}
		fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failCount)
		if successCount > 0 {
			fmt.Println("\nServer configuration:")
			printServerDetails(commonServer, "  ")
		}
	}

	// Return error if all failed, but allow partial success
	if successCount == 0 {
		return fmt.Errorf("failed to add server to any editor")
	}

	return nil
}

// RunIntoGlazeProcessor implements GlazeCommand interface for structured output
func (c *AddCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &AddSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	// Validate arguments
	if settings.Name == "" {
		return fmt.Errorf("server name is required")
	}

	var command string
	var cmdArgs []string

	// Handle different transport types
	if settings.URL != "" {
		// URL-based transport (HTTP/SSE)
		if settings.Command != "" {
			return fmt.Errorf("cannot specify both command and URL")
		}
		command = ""
		cmdArgs = nil
	} else {
		// Stdio transport
		if settings.Command == "" {
			return fmt.Errorf("command required for stdio transport")
		}
		command = settings.Command
		cmdArgs = settings.Args
	}

	// Parse environment variables
	envMap := make(map[string]string)
	for _, e := range settings.Env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid environment variable format: %s (expected KEY=VALUE)", e)
		}
		envMap[parts[0]] = parts[1]
	}

	// Auto-detect transport type
	isSSE := false
	var transport string
	if settings.URL != "" {
		// Simple heuristic: URLs containing "events" or "sse" are SSE
		urlLower := strings.ToLower(settings.URL)
		isSSE = strings.Contains(urlLower, "events") || strings.Contains(urlLower, "sse")
		if isSSE {
			transport = "SSE"
		} else {
			transport = "HTTP"
		}
	} else {
		transport = "stdio"
	}

	commonServer := mcptypes.CommonServer{
		Name:    settings.Name,
		Command: command,
		Args:    cmdArgs,
		Env:     envMap,
		URL:     settings.URL,
		IsSSE:   isSSE,
	}

	// Execute for each editor
	stores, err := NewStoresWithTarget(editors, settings.Target)
	if err != nil {
		return err
	}

	for _, editorStore := range stores {
		var success bool
		var errorMsg string

		err := editorStore.Store.AddServer(commonServer, settings.Overwrite)
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

		// Create row for this editor's result
		row := types.NewRow(
			types.MRP("editor", editorStore.Editor),
			types.MRP("name", settings.Name),
			types.MRP("command", command),
			types.MRP("args", cmdArgs),
			types.MRP("env", envMap),
			types.MRP("url", settings.URL),
			types.MRP("transport", transport),
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

// NewAddCommandDual creates a new dual-mode add command
func NewAddCommandDual() (*AddCommand, error) {
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
		"add",
		cmds.WithShort("Add MCP server to one or more editors"),
		cmds.WithLong(`Add a new MCP server configuration to one or more editors.

EDITORS can be a single editor or comma-separated list: claude,cursor,amp

If a server with the same name already exists, the command will fail unless --overwrite is specified.

Transport types:
- Stdio: Provide COMMAND [ARGS...] 
- HTTP/SSE: Provide --url with appropriate URL

URLs containing 'events' or 'sse' in the path are detected as SSE transport.

Examples:
  # Human-readable output (default)
  `+"```"+`
  mcp editor config claude,cursor add myserver /path/to/cmd --env KEY=value
  `+"```"+`
  
  # Structured output
  `+"```"+`
  mcp editor config claude add myserver /path/to/cmd --with-structured-output --output json
  `+"```"+`
  
  # Multi-editor with structured output
  `+"```"+`
  mcp editor config claude,cursor,amp add shared-server /path/to/cmd --with-structured-output --output table
  `+"```"+``),

		// Define command arguments and flags
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"editors",
				parameters.ParameterTypeString,
				parameters.WithHelp("Editor(s) to add server to (comma-separated)"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the MCP server"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"command",
				parameters.ParameterTypeString,
				parameters.WithHelp("Command to execute for stdio transport"),
				parameters.WithRequired(false),
			),
			parameters.NewParameterDefinition(
				"args",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Command arguments for stdio transport"),
				parameters.WithRequired(false),
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
				"env",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Environment variables in KEY=VALUE format"),
				parameters.WithShortFlag("e"),
				parameters.WithDefault([]string{}),
			),
			parameters.NewParameterDefinition(
				"overwrite",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Overwrite existing server if it exists"),
				parameters.WithShortFlag("w"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"url",
				parameters.ParameterTypeString,
				parameters.WithHelp("URL for HTTP/SSE servers"),
				parameters.WithDefault(""),
			),
		),

		// Add glazed and command settings layers
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)

	return &AddCommand{
		CommandDescription: cmdDesc,
	}, nil
}
