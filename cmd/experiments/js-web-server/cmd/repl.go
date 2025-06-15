package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repl"
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
	return &ReplCmd{
		CommandDescription: cmds.NewCommandDescription(
			"repl",
			cmds.WithShort("Start an interactive JavaScript REPL"),
			cmds.WithLong(`Start an interactive JavaScript REPL (Read-Eval-Print Loop) for experimenting with JavaScript code.

The REPL provides:
- Interactive JavaScript execution with Goja engine
- Multiline input support (Ctrl+J for additional lines)
- Command history
- Built-in commands (type /help for list)
- Integration with existing js-web-server configurations`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"multiline",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Start in multiline mode"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

// Run implements the BareCommand interface
func (c *ReplCmd) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Parse settings from layers
	s := &ReplSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to parse REPL settings")
	}

	// Create the REPL model
	model := repl.NewModel(s.Multiline)

	// Create the bubble tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running REPL: %w", err)
	}

	return nil
}
