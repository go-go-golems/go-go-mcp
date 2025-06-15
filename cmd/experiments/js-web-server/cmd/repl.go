package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repl"
	pinocchio_cmds "github.com/go-go-golems/pinocchio/pkg/cmds"
	"github.com/pkg/errors"
)

// ReplCmd represents the REPL command
type ReplCmd struct {
	*cmds.CommandDescription
}

// ReplSettings holds the configuration for the REPL command
type ReplSettings struct {
	Multiline bool `glazed.parameter:"multiline"`
}

// Ensure ReplCmd implements BareCommand
var _ cmds.BareCommand = &ReplCmd{}

// NewReplCmd creates a new REPL command
func NewReplCmd() (*ReplCmd, error) {
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

	// Create default layer for REPL specific settings
	defaultLayer, err := layers.NewParameterLayer(
		layers.DefaultSlug,
		"REPL Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"multiline",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Start in multiline mode"),
				parameters.WithShortFlag("m"),
				parameters.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default layer")
	}

	// Combine all layers
	allLayers := append(geppettoLayers, defaultLayer)

	return &ReplCmd{
		CommandDescription: cmds.NewCommandDescription(
			"repl",
			cmds.WithShort("Start an interactive JavaScript REPL with Geppetto AI capabilities"),
			cmds.WithLong(`Start an interactive JavaScript REPL (Read-Eval-Print Loop) for experimenting with JavaScript code.

The REPL provides:
- Interactive JavaScript execution with Goja engine
- Geppetto AI capabilities including Conversation and ChatStepFactory APIs
- Multiline input support (Ctrl+J for additional lines)
- Command history
- Built-in commands (type /help for list)
- AI profile configuration support

The REPL is useful for:
• Interactive testing of JavaScript code
• Experimenting with AI API calls
• Quick prototyping with Geppetto capabilities
• Testing Conversation flows

Examples:
  repl
  repl --multiline
  repl --profile 4o-mini
  repl --profile claude-dev --multiline`),
			cmds.WithLayers(layers.NewParameterLayers(layers.WithLayers(allLayers...))),
		),
	}, nil
}

// Run implements the BareCommand interface
func (c *ReplCmd) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Parse settings from layers
	var replSettings ReplSettings
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, &replSettings)
	if err != nil {
		return errors.Wrap(err, "failed to parse REPL settings")
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

	// Create the REPL model with step settings
	model := repl.NewModelWithSettings(replSettings.Multiline, stepSettings)

	// Create the bubble tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running REPL: %w", err)
	}

	return nil
}
