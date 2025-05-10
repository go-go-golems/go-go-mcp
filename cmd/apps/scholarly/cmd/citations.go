package cmd

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/scholarly"
	"github.com/rs/zerolog/log"
)

// CitationsCommand is a glazed command for getting citations
type CitationsCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &CitationsCommand{}

// CitationsSettings holds the parameters for citation retrieval
type CitationsSettings struct {
	WorkID    string `glazed.parameter:"id"`
	Direction string `glazed.parameter:"direction"`
	Limit     int    `glazed.parameter:"limit"`
}

// RunIntoGlazeProcessor executes the citation retrieval and processes results
func (c *CitationsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &CitationsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate work ID
	if s.WorkID == "" {
		return fmt.Errorf("work_id cannot be empty")
	}

	// Validate direction
	if s.Direction != "refs" && s.Direction != "cited_by" {
		return fmt.Errorf("direction must be either 'refs' or 'cited_by', got '%s'", s.Direction)
	}

	log.Debug().Str("work_id", s.WorkID).Str("direction", s.Direction).Int("limit", s.Limit).Msg("Getting citations")

	req := scholarly.GetCitationsRequest{
		WorkID:    s.WorkID,
		Direction: s.Direction,
		Limit:     s.Limit,
	}

	// Get citations
	response, err := scholarly.GetCitations(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get citations")
		return err
	}

	if len(response.Citations) == 0 {
		// Add an info row for no citations
		directionName := getDirectionDisplayName(s.Direction)
		infoRow := types.NewRow(
			types.MRP("_type", "info"),
			types.MRP("message", fmt.Sprintf("No %s found for the work.", directionName)),
		)
		return gp.AddRow(ctx, infoRow)
	}

	// Process citations into rows
	for _, citation := range response.Citations {
		row := types.NewRow(
			types.MRP("title", citation.Title),
		)

		// Add optional fields if present
		if citation.DOI != "" {
			row.Set("doi", citation.DOI)
		}

		if citation.Year > 0 {
			row.Set("year", citation.Year)
		}

		if citation.ID != "" {
			row.Set("id", citation.ID)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Add a pagination notice if there's a next cursor
	if response.NextCursor != "" {
		paginationRow := types.NewRow(
			types.MRP("_type", "pagination"),
			types.MRP("next_cursor", response.NextCursor),
			types.MRP("message", "More results available. Use cursor token for pagination."),
		)
		if err := gp.AddRow(ctx, paginationRow); err != nil {
			return err
		}
	}

	return nil
}

// getDirectionDisplayName returns a user-friendly name for the citation direction
func getDirectionDisplayName(direction string) string {
	if direction == "refs" {
		return "references"
	}
	return "citations"
}

// NewCitationsCommand creates a new citations command
func NewCitationsCommand() (*CitationsCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"citations",
		cmds.WithShort("Get citations for a scholarly work"),
		cmds.WithLong(`Retrieve one hop of the citation graph - either works this paper cites
(outgoing references) or works that cite it (incoming citations).

Example:
  arxiv-libgen-cli citations --id "10.1038/nphys1170" --direction refs
  arxiv-libgen-cli citations --id "W2741809809" --direction cited_by --limit 20`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Work ID (DOI or OpenAlex ID) (required)"),
				parameters.WithRequired(true),
				parameters.WithShortFlag("i"),
			),
			parameters.NewParameterDefinition(
				"direction",
				parameters.ParameterTypeString,
				parameters.WithHelp("Citation direction: 'refs' (outgoing) or 'cited_by' (incoming)"),
				parameters.WithDefault("cited_by"),
				parameters.WithShortFlag("r"),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of citations to return"),
				parameters.WithDefault(20),
				parameters.WithShortFlag("l"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &CitationsCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the citations command
func init() {
	// Create the citations command
	citationsCmd, err := NewCitationsCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create citations command")
	}

	// Convert to Cobra command
	ctCobraCmd, err := cli.BuildCobraCommandFromCommand(citationsCmd,
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build citations cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(ctCobraCmd)
}
