package config

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// EditCommand is a Glazed BareCommand for editing config files
type EditCommand struct {
	*cmds.CommandDescription
}

// EditSettings holds the parameters for the edit command
type EditSettings struct {
	Editors string `glazed.parameter:"editors"`
	Target  string `glazed.parameter:"target"`
}

// Ensure interface is implemented
var _ cmds.BareCommand = &EditCommand{}

// Run implements BareCommand interface
func (c *EditCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &EditSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	// Get editor command
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vi"
	}

	// Handle multiple editors
	if len(editors) == 1 {
		// Single editor - use original behavior
		store, err := NewStoreWithTarget(editors[0], settings.Target)
		if err != nil {
			return err
		}

		configPath := store.GetConfigPath()

		c := exec.Command(editorCmd, configPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	} else {
		// Multiple editors - edit sequentially with confirmation
		fmt.Printf("Editing config files for %d editors...\n", len(editors))

		for i, editor := range editors {
			fmt.Printf("\n[%d/%d] Editing %s configuration:\n", i+1, len(editors), editor)

			store, err := NewStoreWithTarget(editor, settings.Target)
			if err != nil {
				fmt.Printf("❌ Error opening %s config: %v\n", editor, err)
				continue
			}

			configPath := store.GetConfigPath()
			fmt.Printf("Opening: %s\n", configPath)

			c := exec.Command(editorCmd, configPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				fmt.Printf("❌ Error editing %s config: %v\n", editor, err)
				continue
			}

			fmt.Printf("✅ %s config edited successfully\n", editor)
		}

		return nil
	}
}

// NewEditCommand creates a new edit command
func NewEditCommand() (cmds.BareCommand, error) {
	cmd := &EditCommand{
		CommandDescription: cmds.NewCommandDescription(
			"edit",
			cmds.WithShort("Edit config file"),
			cmds.WithLong("Edit the configuration file for one or more editors in your default editor.\n\nThe command respects the $EDITOR environment variable, defaulting to 'vi'.\n\nFor multiple editors, files are edited sequentially with confirmation prompts.\n\nExamples:\n```\n# Edit configuration for a single editor\nmcp editor config claude edit\n\n# Edit configurations for multiple editors\nmcp editor config claude,cursor edit\n\n# Edit with specific target configuration\nmcp editor config claude edit --target global\n```"),

			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"editors",
					parameters.ParameterTypeString,
					parameters.WithHelp("Editor(s) to edit config for (comma-separated)"),
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
