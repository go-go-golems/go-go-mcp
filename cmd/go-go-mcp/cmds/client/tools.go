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
	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
					"json-file",
					parameters.ParameterTypeStringFromFile,
					parameters.WithHelp("Tool arguments as JSON file"),
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

// Glaze output (structured)
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
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			if err == nil {
				err = errors.Wrap(closeErr, "failed to close client")
			}
		}
	}()

	res, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return err
	}

	for _, tool := range res.Tools {
		// Prepare raw schema JSON from either RawInputSchema or structured InputSchema
		var schemaBytes []byte
		if tool.RawInputSchema != nil {
			schemaBytes = tool.RawInputSchema
		} else {
			var err error
			schemaBytes, err = json.Marshal(tool.InputSchema)
			if err != nil {
				return fmt.Errorf("failed to marshal structured schema: %w", err)
			}
		}

		// Unmarshal into interface for display
		var schemaObj interface{}
		if err := json.Unmarshal(schemaBytes, &schemaObj); err != nil {
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

	if res.NextCursor != "" {
		row := types.NewRow(
			types.MRP("cursor", res.NextCursor),
			types.MRP("type", "cursor"),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// Human-readable default output
func (c *ListToolsCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	w io.Writer,
) error {
	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	defer client.Close()

	res, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return err
	}
	if len(res.Tools) == 0 {
		_, _ = fmt.Fprintln(w, "No tools available")
		return nil
	}
	for _, tool := range res.Tools {
		_, _ = fmt.Fprintf(w, "- %s: %s\n", tool.Name, tool.Description)
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
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			if err == nil {
				err = errors.Wrap(closeErr, "failed to close client")
			}
		}
	}()

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

	res, err := client.CallTool(ctx, mcp.CallToolRequest{
		Request: mcp.Request{Method: string(mcp.MethodToolsCall)},
		Params: mcp.CallToolParams{
			Name:      s.ToolName,
			Arguments: toolArgMap,
		},
	})
	if err != nil {
		return err
	}

	// Pretty print the result
	for _, content := range res.Content {
		switch c := content.(type) {
		case mcp.TextContent:
			_, err = fmt.Fprintf(w, "%s\n", c.Text)
		case mcp.ImageContent:
			_, err = fmt.Fprintf(w, "[image %s, %d bytes base64]\n", c.MIMEType, len(c.Data))
		case mcp.EmbeddedResource:
			_, err = fmt.Fprintf(w, "[embedded resource]\n")
		default:
			_, err = fmt.Fprintf(w, "[unknown content]\n")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Structured output for call tool
func (c *CallToolCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &CallToolSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	defer client.Close()

	toolArgMap := make(map[string]interface{})
	if s.JSON != "" {
		if err := json.Unmarshal([]byte(s.JSON), &toolArgMap); err != nil {
			return fmt.Errorf("invalid tool arguments JSON: %w", err)
		}
	} else if len(s.Args) > 0 {
		toolArgMap = s.Args
	}

	res, err := client.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: s.ToolName, Arguments: toolArgMap}})
	if err != nil {
		return err
	}

	for _, content := range res.Content {
		switch c := content.(type) {
		case mcp.TextContent:
			if err := gp.AddRow(ctx, types.NewRow(types.MRP("type", "text"), types.MRP("text", c.Text))); err != nil {
				return err
			}
		case mcp.ImageContent:
			if err := gp.AddRow(ctx, types.NewRow(types.MRP("type", "image"), types.MRP("mime", c.MIMEType), types.MRP("data", c.Data))); err != nil {
				return err
			}
		case mcp.EmbeddedResource:
			if err := gp.AddRow(ctx, types.NewRow(types.MRP("type", "resource"))); err != nil {
				return err
			}
		default:
			if err := gp.AddRow(ctx, types.NewRow(types.MRP("type", "unknown"))); err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	listCmd, err := NewListToolsCommand()
	cobra.CheckErr(err)
	cobraListCmd, err := cli.BuildCobraCommand(listCmd)
	cobra.CheckErr(err)

	callCmd, err := NewCallToolCommand()
	cobra.CheckErr(err)
	cobraCallCmd, err := cli.BuildCobraCommand(callCmd)
	cobra.CheckErr(err)

	ToolsCmd.AddCommand(cobraListCmd)
	ToolsCmd.AddCommand(cobraCallCmd)
}
