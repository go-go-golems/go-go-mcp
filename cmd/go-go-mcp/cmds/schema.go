package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"
	mcp_cmds "github.com/go-go-golems/go-go-mcp/pkg/cmds"
	"github.com/pkg/errors"
)

type SchemaCommandSettings struct {
	File string `glazed.parameter:"file"`
}

type SchemaCommand struct {
	*cmds.CommandDescription
}

func NewSchemaCommand() (*SchemaCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &SchemaCommand{
		CommandDescription: cmds.NewCommandDescription(
			"schema",
			cmds.WithShort("Output JSON schema for a command YAML file"),
			cmds.WithLong(`Generate and output a JSON schema representation of a shell command YAML file.
This schema can be used for LLM tool calling definitions or command validation.

Example:
  mcp-server schema ./commands/my-command.yaml`),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML command file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *SchemaCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &SchemaCommandSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Load the command
	loader := &mcp_cmds.ShellCommandLoader{}
	fs_, filePath, err := loaders.FileNameToFsFilePath(s.File)
	if err != nil {
		return fmt.Errorf("could not get absolute path: %w", err)
	}

	cmds_, err := loader.LoadCommands(fs_, filePath, []cmds.CommandDescriptionOption{}, []alias.Option{})
	if err != nil {
		return fmt.Errorf("could not load command: %w", err)
	}
	if len(cmds_) != 1 {
		return fmt.Errorf("expected exactly one command, got %d", len(cmds_))
	}

	shellCmd, ok := cmds_[0].(*mcp_cmds.ShellCommand)
	if !ok {
		return fmt.Errorf("expected ShellCommand, got %T", cmds_[0])
	}

	// Convert to JSON schema
	schema, err := shellCmd.Description().ToJsonSchema()
	if err != nil {
		return fmt.Errorf("could not convert to JSON schema: %w", err)
	}

	// Output as JSON
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(schema); err != nil {
		return fmt.Errorf("could not encode JSON schema: %w", err)
	}

	return nil
}
