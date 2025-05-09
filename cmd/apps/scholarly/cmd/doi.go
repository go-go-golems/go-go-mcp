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

// DOICommand is a glazed command for resolving DOIs
type DOICommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &DOICommand{}

// DOISettings holds the parameters for DOI resolution
type DOISettings struct {
	DOI string `glazed.parameter:"doi"`
}

// RunIntoGlazeProcessor executes the DOI resolution and processes results
func (c *DOICommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &DOISettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate DOI
	if s.DOI == "" {
		return fmt.Errorf("DOI cannot be empty")
	}

	log.Debug().Str("doi", s.DOI).Msg("Resolving DOI")

	req := scholarly.ResolveDOIRequest{
		DOI: s.DOI,
	}

	// Resolve DOI
	work, err := scholarly.ResolveDOI(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to resolve DOI")
		return err
	}

	// Create a row from the work
	row := types.NewRow()

	// Add all available fields to the row
	row.Set("title", work.Title)
	row.Set("doi", work.DOI)
	row.Set("id", work.ID)
	row.Set("year", work.Year)
	row.Set("authors", work.Authors)
	row.Set("citation_count", work.CitationCount)
	row.Set("is_open_access", work.IsOA)
	row.Set("pdf_url", work.PDFURL)
	row.Set("abstract", work.Abstract)
	row.Set("source_name", work.SourceName)

	return gp.AddRow(ctx, row)
}

// NewDOICommand creates a new DOI command
func NewDOICommand() (*DOICommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"doi",
		cmds.WithShort("Resolve a DOI to get full metadata"),
		cmds.WithLong(`Resolve a DOI to get complete metadata from both Crossref and OpenAlex.

This command fetches metadata from multiple sources and merges them into a single
rich record with information like authors, citations, concepts, and more.

Example:
  arxiv-libgen-cli doi --doi "10.1038/nphys1170"
  arxiv-libgen-cli doi --doi "10.1103/PhysRevLett.116.061102" --output json`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"doi",
				parameters.ParameterTypeString,
				parameters.WithHelp("DOI to resolve (required)"),
				parameters.WithRequired(true),
				parameters.WithShortFlag("i"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &DOICommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the DOI command
func init() {
	// Create the DOI command
	doiCmd, err := NewDOICommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create DOI command")
	}

	// Convert to Cobra command
	doiCobraCmd, err := cli.BuildCobraCommandFromCommand(doiCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build DOI cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(doiCobraCmd)
}
