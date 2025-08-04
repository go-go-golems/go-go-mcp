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
	"github.com/spf13/cobra"
)

// ListCommand is a glazed command for listing MCP servers across editors
type ListCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ListCommand{}
var _ cmds.GlazeCommand = &SingleEditorListCommand{}

// ListSettings holds the parameters for the list command
type ListSettings struct {
	Editors string `glazed.parameter:"editors"`
	Target  string `glazed.parameter:"target"`
}

// RunIntoGlazeProcessor executes the list command and outputs structured data
func (c *ListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &ListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Parse editors from the editors parameter
	if s.Editors == "" {
		return fmt.Errorf("editors parameter is required")
	}

	editors, err := ParseEditors(s.Editors)
	if err != nil {
		return err
	}

	// Create stores for each editor
	stores, err := NewStoresWithTarget(editors, s.Target)
	if err != nil {
		return err
	}

	// Iterate through each editor and collect server data
	for _, editorStore := range stores {
		// Get the target path information
		targetPath := ""
		if s.Target != "" {
			targetPath = s.Target
		} else {
			targetPath = "default"
		}

		// List servers for this editor
		servers, err := editorStore.Store.ListServers()
		if err != nil {
			// Create an error row for this editor
			row := types.NewRow(
				types.MRP("editor", editorStore.Editor),
				types.MRP("name", "ERROR"),
				types.MRP("command", ""),
				types.MRP("args", []string{}),
				types.MRP("env", map[string]string{}),
				types.MRP("url", ""),
				types.MRP("transport", ""),
				types.MRP("enabled", false),
				types.MRP("target", targetPath),
				types.MRP("error", err.Error()),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		// If no servers for this editor, add a placeholder row
		if len(servers) == 0 {
			row := types.NewRow(
				types.MRP("editor", editorStore.Editor),
				types.MRP("name", ""),
				types.MRP("command", ""),
				types.MRP("args", []string{}),
				types.MRP("env", map[string]string{}),
				types.MRP("url", ""),
				types.MRP("transport", ""),
				types.MRP("enabled", false),
				types.MRP("target", targetPath),
				types.MRP("error", "No servers configured"),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		// Add a row for each server
		for name, server := range servers {
			// Check if server is disabled
			disabled, err := editorStore.Store.IsServerDisabled(name)
			if err != nil {
				// If we can't check status, assume enabled but note the error
				disabled = false
			}

			// Determine transport type
			transport := "stdio"
			if server.URL != "" {
				if server.IsSSE {
					transport = "sse"
				} else {
					transport = "http"
				}
			}

			// Create row with all server information
			row := types.NewRow(
				types.MRP("editor", editorStore.Editor),
				types.MRP("name", name),
				types.MRP("command", server.Command),
				types.MRP("args", server.Args),
				types.MRP("env", server.Env),
				types.MRP("url", server.URL),
				types.MRP("transport", transport),
				types.MRP("enabled", !disabled),
				types.MRP("target", targetPath),
			)

			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewGlazedListCommand creates a new Glazed list command
func NewGlazedListCommand() (*ListCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Get supported editor names for help text
	editorNames := getSupportedEditorNames()
	editorList := strings.Join(editorNames, ", ")

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List MCP servers from one or more editors"),
		cmds.WithLong(fmt.Sprintf(`List all configured MCP servers for one or more editors.

Supported editors: %s

The command provides structured output showing all server configurations including:
- Server name and enabled/disabled status  
- Transport type (stdio, http, sse)
- Command and arguments for stdio servers
- URL for HTTP/SSE servers
- Environment variables and configuration
- Target location information

Examples:
  # List servers from a single editor
  mcp editor config list --editors claude
  
  # List servers from multiple editors  
  mcp editor config list --editors claude,cursor,amp
  
  # Use specific target configuration
  mcp editor config list --editors claude --target global
  
  # Output as JSON
  mcp editor config list --editors claude,cursor --output json
  
  # Output as CSV with specific fields
  mcp editor config list --editors claude --output csv --fields name,transport,enabled`, editorList)),

		// Define command arguments
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"editors",
				parameters.ParameterTypeString,
				parameters.WithHelp(fmt.Sprintf("Comma-separated list of editors to query. Supported: %s", editorList)),
				parameters.WithShortFlag("e"),
			),
			parameters.NewParameterDefinition(
				"target",
				parameters.ParameterTypeString,
				parameters.WithHelp("Target configuration (e.g., global, cwd)"),
				parameters.WithShortFlag("t"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &ListCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// NewMultiListCommandGlazed creates the new Glazed-based multi-editor list command
func NewMultiListCommandGlazed() *cobra.Command {
	// Create the Glazed command
	listCmd, err := NewGlazedListCommand()
	if err != nil {
		panic(fmt.Sprintf("Failed to create list command: %v", err))
	}

	// Convert to Cobra command
	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		listCmd,
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to build cobra command: %v", err))
	}

	// Override the Use and Args to match the existing interface
	cobraCmd.Use = "list [EDITORS]"
	cobraCmd.Args = cobra.ArbitraryArgs // Allow any arguments to support mixed positional and flags

	// Wrap the run function to handle the positional argument
	originalRunE := cobraCmd.RunE
	if originalRunE != nil {
		cobraCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Find the first non-flag argument as the editors parameter
			for _, arg := range args {
				if !strings.HasPrefix(arg, "-") {
					if err := cmd.Flags().Set("editors", arg); err != nil {
						return err
					}
					break
				}
			}
			// Call the original run function
			return originalRunE(cmd, []string{})
		}
	}

	return cobraCmd
}

// SingleEditorListCommand is a specialized command for single editor listing
type SingleEditorListCommand struct {
	*cmds.CommandDescription
	Editor string
}

// SingleEditorListSettings holds the parameters for the single editor list command
type SingleEditorListSettings struct {
	Target string `glazed.parameter:"target"`
}

// RunIntoGlazeProcessor executes the single editor list command
func (c *SingleEditorListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &SingleEditorListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Create store for the specific editor
	store, err := NewStoreWithTarget(c.Editor, s.Target)
	if err != nil {
		return err
	}

	// Get the target path information
	targetPath := ""
	if s.Target != "" {
		targetPath = s.Target
	} else {
		targetPath = "default"
	}

	// List servers for this editor
	servers, err := store.ListServers()
	if err != nil {
		// Create an error row for this editor
		row := types.NewRow(
			types.MRP("editor", c.Editor),
			types.MRP("name", "ERROR"),
			types.MRP("command", ""),
			types.MRP("args", []string{}),
			types.MRP("env", map[string]string{}),
			types.MRP("url", ""),
			types.MRP("transport", ""),
			types.MRP("enabled", false),
			types.MRP("target", targetPath),
			types.MRP("error", err.Error()),
		)
		return gp.AddRow(ctx, row)
	}

	// If no servers for this editor, add a placeholder row
	if len(servers) == 0 {
		row := types.NewRow(
			types.MRP("editor", c.Editor),
			types.MRP("name", ""),
			types.MRP("command", ""),
			types.MRP("args", []string{}),
			types.MRP("env", map[string]string{}),
			types.MRP("url", ""),
			types.MRP("transport", ""),
			types.MRP("enabled", false),
			types.MRP("target", targetPath),
			types.MRP("error", "No servers configured"),
		)
		return gp.AddRow(ctx, row)
	}

	// Add a row for each server
	for name, server := range servers {
		// Check if server is disabled
		disabled, err := store.IsServerDisabled(name)
		if err != nil {
			// If we can't check status, assume enabled but note the error
			disabled = false
		}

		// Determine transport type
		transport := "stdio"
		if server.URL != "" {
			if server.IsSSE {
				transport = "sse"
			} else {
				transport = "http"
			}
		}

		// Create row with all server information
		row := types.NewRow(
			types.MRP("editor", c.Editor),
			types.MRP("name", name),
			types.MRP("command", server.Command),
			types.MRP("args", server.Args),
			types.MRP("env", server.Env),
			types.MRP("url", server.URL),
			types.MRP("transport", transport),
			types.MRP("enabled", !disabled),
			types.MRP("target", targetPath),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewSingleEditorListCommand creates a new single editor list command
func NewSingleEditorListCommand(editor string) (*SingleEditorListCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort(fmt.Sprintf("List MCP servers for %s", editor)),
		cmds.WithLong(fmt.Sprintf(`List all configured MCP servers for %s.

The command provides structured output showing all server configurations including:
- Server name and enabled/disabled status  
- Transport type (stdio, http, sse)
- Command and arguments for stdio servers
- URL for HTTP/SSE servers
- Environment variables and configuration
- Target location information

Examples:
  # List servers with default output
  mcp editor config %s list
  
  # Use specific target configuration
  mcp editor config %s list --target global
  
  # Output as JSON
  mcp editor config %s list --output json
  
  # Output as CSV with specific fields
  mcp editor config %s list --output csv --fields name,transport,enabled`, editor, editor, editor, editor, editor)),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"target",
				parameters.ParameterTypeString,
				parameters.WithHelp("Target configuration (e.g., global, cwd)"),
				parameters.WithShortFlag("t"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &SingleEditorListCommand{
		CommandDescription: cmdDesc,
		Editor:             editor,
	}, nil
}

// NewListCommandGlazed creates a new Glazed-based list command for a specific editor
func NewListCommandGlazed(editor string) *cobra.Command {
	// Create the single editor Glazed command
	listCmd, err := NewSingleEditorListCommand(editor)
	if err != nil {
		panic(fmt.Sprintf("Failed to create list command: %v", err))
	}

	// Convert to Cobra command
	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		listCmd,
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to build cobra command: %v", err))
	}

	return cobraCmd
}
