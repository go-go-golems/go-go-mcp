package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/client/layers"

	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-mcp/cmd/go-go-mcp/cmds/client/helpers"
	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// PromptsCmd handles the "prompts" command group
var PromptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "Interact with prompts",
	Long:  `List available prompts and execute specific prompts.`,
}

type ListPromptsCommand struct {
	*cmds.CommandDescription
}

type ExecutePromptCommand struct {
	*cmds.CommandDescription
}

type ExecutePromptSettings struct {
	Args       string `glazed.parameter:"args"`
	PromptName string `glazed.parameter:"prompt-name"`
}

func NewListPromptsCommand() (*ListPromptsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	clientLayer, err := layers.NewClientParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create client parameter layer")
	}

	return &ListPromptsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List available prompts"),
			cmds.WithLayersList(
				glazedParameterLayer,
				clientLayer,
			),
		),
	}, nil
}

func NewExecutePromptCommand() (*ExecutePromptCommand, error) {
	clientLayer, err := layers.NewClientParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create client parameter layer")
	}

	return &ExecutePromptCommand{
		CommandDescription: cmds.NewCommandDescription(
			"execute",
			cmds.WithShort("Execute a specific prompt"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"args",
					parameters.ParameterTypeString,
					parameters.WithHelp("Prompt arguments as JSON string"),
					parameters.WithDefault(""),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"prompt-name",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the prompt to execute"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				clientLayer,
			),
		),
	}, nil
}

func (c *ListPromptsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("failed to close client")
		}
	}()

	res, err := client.ListPrompts(ctx, mcp.ListPromptsRequest{})
	if err != nil {
		return err
	}

	for _, prompt := range res.Prompts {
		row := types.NewRow(
			types.MRP("name", prompt.Name),
			types.MRP("description", prompt.Description),
		)

		args := make([]map[string]interface{}, len(prompt.Arguments))
		for i, arg := range prompt.Arguments {
			args[i] = map[string]interface{}{
				"name":        arg.Name,
				"required":    arg.Required,
				"description": arg.Description,
			}
		}
		row.Set("arguments", args)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	if res.NextCursor != "" {
		cursorRow := types.NewRow(types.MRP("cursor", res.NextCursor))
		if err := gp.AddRow(ctx, cursorRow); err != nil {
			return err
		}
	}

	return nil
}

func (c *ListPromptsCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	w io.Writer,
) error {
	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("failed to close client")
		}
	}()

	res, err := client.ListPrompts(ctx, mcp.ListPromptsRequest{})
	if err != nil {
		return err
	}
	if len(res.Prompts) == 0 {
		_, _ = fmt.Fprintln(w, "No prompts available")
		return nil
	}
	for _, p := range res.Prompts {
		_, _ = fmt.Fprintf(w, "- %s: %s\n", p.Name, p.Description)
	}
	return nil
}

func (c *ExecutePromptCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	w io.Writer,
) error {
	s := &ExecutePromptSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("failed to close client")
		}
	}()

	promptArgMap := make(map[string]string)
	if s.Args != "" {
		if err := json.Unmarshal([]byte(s.Args), &promptArgMap); err != nil {
			return fmt.Errorf("invalid prompt arguments JSON: %w", err)
		}
	}

	res, err := client.GetPrompt(ctx, mcp.GetPromptRequest{Params: mcp.GetPromptParams{Name: s.PromptName, Arguments: promptArgMap}})
	if err != nil {
		return err
	}

	for _, message := range res.Messages {
		if tc, ok := message.Content.(mcp.TextContent); ok {
			_, _ = fmt.Fprintf(w, "%s: %s\n", message.Role, tc.Text)
		}
	}
	return nil
}

func (c *ExecutePromptCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ExecutePromptSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("failed to close client")
		}
	}()

	promptArgMap := make(map[string]string)
	if s.Args != "" {
		if err := json.Unmarshal([]byte(s.Args), &promptArgMap); err != nil {
			return fmt.Errorf("invalid prompt arguments JSON: %w", err)
		}
	}

	res, err := client.GetPrompt(ctx, mcp.GetPromptRequest{Params: mcp.GetPromptParams{Name: s.PromptName, Arguments: promptArgMap}})
	if err != nil {
		return err
	}

	for _, message := range res.Messages {
		if tc, ok := message.Content.(mcp.TextContent); ok {
			row := types.NewRow(
				types.MRP("role", message.Role),
				types.MRP("text", tc.Text),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	listCmd, err := NewListPromptsCommand()
	cobra.CheckErr(err)

	listCobraCmd, err := cli.BuildCobraCommand(listCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	cobra.CheckErr(err)

	executeCmd, err := NewExecutePromptCommand()
	cobra.CheckErr(err)

	executeCobraCmd, err := cli.BuildCobraCommand(executeCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	cobra.CheckErr(err)

	PromptsCmd.AddCommand(listCobraCmd)
	PromptsCmd.AddCommand(executeCobraCmd)
}
