package client

import (
	"context"
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

// ResourcesCmd handles the "resources" command group
var ResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Interact with resources",
	Long:  `List available resources and read specific resources.`,
}

type ListResourcesCommand struct {
	*cmds.CommandDescription
}

type ListResourcesSettings struct {
}

type ReadResourceCommand struct {
	*cmds.CommandDescription
}

type ReadResourceSettings struct {
	URI string `glazed.parameter:"uri"`
}

func NewListResourcesCommand() (*ListResourcesCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	clientLayer, err := layers.NewClientParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create client parameter layer")
	}

	return &ListResourcesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List available resources"),
			cmds.WithLayersList(
				glazedParameterLayer,
				clientLayer,
			),
		),
	}, nil
}

func NewReadResourceCommand() (*ReadResourceCommand, error) {
	clientLayer, err := layers.NewClientParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create client parameter layer")
	}

	return &ReadResourceCommand{
		CommandDescription: cmds.NewCommandDescription(
			"read",
			cmds.WithShort("Read a specific resource"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"uri",
					parameters.ParameterTypeString,
					parameters.WithHelp("URI of the resource to read"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(clientLayer),
		),
	}, nil
}

func (c *ListResourcesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListResourcesSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	var closeErr error
	defer func() {
		if cerr := client.Close(); cerr != nil {
			closeErr = errors.Wrap(cerr, "failed to close client")
		}
		if err == nil {
			err = closeErr
		}
	}()

	res, err := client.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return err
	}

	for _, resource := range res.Resources {
		row := types.NewRow(
			types.MRP("uri", resource.URI),
			types.MRP("name", resource.Name),
			types.MRP("description", resource.Description),
			types.MRP("mime_type", resource.MIMEType),
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

func (c *ListResourcesCommand) RunIntoWriter(
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
			// log and continue
			_ = cerr
		}
	}()

	res, err := client.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return err
	}
	if len(res.Resources) == 0 {
		_, _ = fmt.Fprintln(w, "No resources available")
		return nil
	}
	for _, r := range res.Resources {
		_, _ = fmt.Fprintf(w, "- %s (%s)\n", r.Name, r.URI)
	}
	return nil
}

func (c *ReadResourceCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	w io.Writer,
) error {
	s := &ReadResourceSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	var closeErr error
	defer func() {
		if cerr := client.Close(); cerr != nil {
			closeErr = errors.Wrap(cerr, "failed to close client")
		}
		if err == nil {
			err = closeErr
		}
	}()

	res, err := client.ReadResource(ctx, mcp.ReadResourceRequest{Params: mcp.ReadResourceParams{URI: s.URI}})
	if err != nil {
		return err
	}

	for _, c := range res.Contents {
		if tc, ok := c.(mcp.TextResourceContents); ok {
			_, _ = fmt.Fprintf(w, "%s\n", tc.Text)
			return nil
		}
	}
	_, _ = fmt.Fprintf(w, "Resource %s has non-text contents\n", s.URI)
	return nil
}

func (c *ReadResourceCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ReadResourceSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
		return err
	}

	client, err := helpers.CreateClientFromSettings(parsedLayers)
	if err != nil {
		return err
	}
	var closeErr error
	defer func() {
		if cerr := client.Close(); cerr != nil {
			closeErr = errors.Wrap(cerr, "failed to close client")
		}
		if err == nil {
			err = closeErr
		}
	}()

	res, err := client.ReadResource(ctx, mcp.ReadResourceRequest{Params: mcp.ReadResourceParams{URI: s.URI}})
	if err != nil {
		return err
	}

	for _, c := range res.Contents {
		if tc, ok := c.(mcp.TextResourceContents); ok {
			if err := gp.AddRow(ctx, types.NewRow(types.MRP("uri", tc.URI), types.MRP("mime", tc.MIMEType), types.MRP("text", tc.Text))); err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	listCmd, err := NewListResourcesCommand()
	cobra.CheckErr(err)

	cobralistCmd, err := cli.BuildCobraCommand(listCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	cobra.CheckErr(err)

	readCmd, err := NewReadResourceCommand()
	cobra.CheckErr(err)

	cobraReadCmd, err := cli.BuildCobraCommand(readCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	cobra.CheckErr(err)

	ResourcesCmd.AddCommand(cobralistCmd)
	ResourcesCmd.AddCommand(cobraReadCmd)
}
