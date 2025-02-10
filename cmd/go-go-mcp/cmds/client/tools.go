package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/client/helpers"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/client/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	toolArgs string
)

// ToolsCmd handles the "tools" command group
var ToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Interact with tools",
	Long:  `List available tools and execute specific tools.`,
}

type ListToolsCommand struct {
	*cmds.CommandDescription
}

type ListToolsSettings struct {
}

type CallToolCommand struct {
	*cmds.CommandDescription
}

type CallToolSettings struct {
	ToolName string                 `glazed.parameter:"tool-name"`
	JSON     string                 `glazed.parameter:"json"`
	Args     map[string]interface{} `glazed.parameter:"args"`
}

func NewListToolsCommand() (*ListToolsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	clientLayer, err := layers.NewClientParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create client parameter layer")
	}

	return &ListToolsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List available tools"),
			cmds.WithLayersList(
				glazedParameterLayer,
				clientLayer,
			),
		),
	}, nil
}

func NewCallToolCommand() (*CallToolCommand, error) {
	clientLayer, err := layers.NewClientParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create client parameter layer")
	}

	return &CallToolCommand{
		CommandDescription: cmds.NewCommandDescription(
			"call",
			cmds.WithShort("Call a specific tool"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"tool-name",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the tool to call"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"json",
					parameters.ParameterTypeString,
					parameters.WithHelp("Tool arguments as JSON string"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"args",
					parameters.ParameterTypeKeyValue,
					parameters.WithHelp("Tool arguments as key=value pairs"),
					parameters.WithDefault(map[string]interface{}{}),
				),
			),
			cmds.WithLayersList(clientLayer),
		),
	}, nil
}

func (c *ListToolsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListToolsSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	defer client.Close(ctx)

	tools, cursor, err := client.ListTools(ctx, "")
	if err != nil {
		return err
	}

	for _, tool := range tools {
		// First unmarshal the schema into an interface{} to ensure it's valid JSON
		var schemaObj interface{}
		if err := json.Unmarshal(tool.InputSchema, &schemaObj); err != nil {
			return fmt.Errorf("failed to parse schema JSON: %w", err)
		}

		row := types.NewRow(
			types.MRP("name", tool.Name),
			types.MRP("description", tool.Description),
			types.MRP("schema", schemaObj),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	if cursor != "" {
		row := types.NewRow(
			types.MRP("cursor", cursor),
			types.MRP("type", "cursor"),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *CallToolCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	w io.Writer,
) error {
	s := &CallToolSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	defer client.Close(ctx)

	// Parse tool arguments - first try JSON string, then key-value pairs
	toolArgMap := make(map[string]interface{})

	// If JSON args are provided, they take precedence
	if s.JSON != "" {
		if err := json.Unmarshal([]byte(s.JSON), &toolArgMap); err != nil {
			return fmt.Errorf("invalid tool arguments JSON: %w", err)
		}
	} else if len(s.Args) > 0 {
		// Otherwise use key-value pairs if provided
		toolArgMap = s.Args
	}

	result, err := client.CallTool(ctx, s.ToolName, toolArgMap)
	if err != nil {
		return err
	}

	// Pretty print the result
	for _, content := range result.Content {
		_, err = fmt.Fprintf(w, "Type: %s\n", content.Type)
		if err != nil {
			return err
		}

		switch content.Type {
		case "text":
			_, err = fmt.Fprintf(w, "Content:\n%s\n", content.Text)
		case "image":
			_, err = fmt.Fprintf(w, "Image:\n%s\n", content.Data)
		case "resource":
			_, err = fmt.Fprintf(w, "URI: %s\nMimeType: %s\n",
				content.Resource.URI, content.Resource.MimeType)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	listCmd, err := NewListToolsCommand()
	cobra.CheckErr(err)

	cobraListCmd, err := cli.BuildCobraCommandFromGlazeCommand(listCmd)
	cobra.CheckErr(err)

	callCmd, err := NewCallToolCommand()
	cobra.CheckErr(err)

	cobraCallCmd, err := cli.BuildCobraCommandFromWriterCommand(callCmd)
	cobra.CheckErr(err)

	ToolsCmd.AddCommand(cobraListCmd)
	ToolsCmd.AddCommand(cobraCallCmd)
}
