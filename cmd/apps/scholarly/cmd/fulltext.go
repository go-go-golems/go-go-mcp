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

// FulltextCommand is a glazed command for finding full text URLs
type FulltextCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &FulltextCommand{}

// FulltextSettings holds the parameters for full text search
type FulltextSettings struct {
	DOI           string `glazed.parameter:"doi"`
	Title         string `glazed.parameter:"title"`
	PreferVersion string `glazed.parameter:"version"`
}

// RunIntoGlazeProcessor executes the full text search and processes results
func (c *FulltextCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &FulltextSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate that we have at least DOI or title
	if s.DOI == "" && s.Title == "" {
		return fmt.Errorf("either DOI or title must be provided")
	}

	log.Debug().Str("doi", s.DOI).Str("title", s.Title).Str("version", s.PreferVersion).Msg("Finding full text")

	req := scholarly.FindFullTextRequest{
		DOI:           s.DOI,
		Title:         s.Title,
		PreferVersion: s.PreferVersion,
	}

	// Find the full text
	response, err := scholarly.FindFullText(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to find full text")
		return err
	}

	// Create a row with the response data
	row := types.NewRow(
		types.MRP("url", response.PDFURL),
		types.MRP("source", response.Source),
		types.MRP("is_pdf", response.IsPDF),
	)

	// Add optional fields if present
	if response.OAStatus != "" {
		row.Set("open_access_status", response.OAStatus)
	}

	if response.License != "" {
		row.Set("license", response.License)
	}

	if response.MD5 != "" {
		row.Set("md5", response.MD5)
	}

	// Add a note if the source is LibGen
	if response.Source == "libgen" {
		// Add an info row with a cautionary note
		infoRow := types.NewRow(
			types.MRP("_type", "info"),
			types.MRP("message", "This URL was obtained from LibGen. Please respect copyright laws and the terms of use for the content you access."),
		)
		if err := gp.AddRow(ctx, infoRow); err != nil {
			return err
		}
	}

	return gp.AddRow(ctx, row)
}

// NewFulltextCommand creates a new fulltext command
func NewFulltextCommand() (*FulltextCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"fulltext",
		cmds.WithShort("Find full text URL for a scholarly work"),
		cmds.WithLong(`Find the best PDF or HTML URL for a scholarly work, checking
open access sources first and falling back to LibGen if necessary.

Provide either a DOI or a title to search for. By default, the
published version is preferred over accepted or submitted versions.

Example:
  arxiv-libgen-cli fulltext --doi "10.1038/nphys1170"
  arxiv-libgen-cli fulltext --title "The rise of quantum biology" --version accepted`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"doi",
				parameters.ParameterTypeString,
				parameters.WithHelp("DOI of the work"),
				parameters.WithDefault(""),
				parameters.WithShortFlag("i"),
			),
			parameters.NewParameterDefinition(
				"title",
				parameters.ParameterTypeString,
				parameters.WithHelp("Title of the work"),
				parameters.WithDefault(""),
				parameters.WithShortFlag("t"),
			),
			parameters.NewParameterDefinition(
				"version",
				parameters.ParameterTypeString,
				parameters.WithHelp("Preferred version (published, accepted, submitted)"),
				parameters.WithDefault("published"),
				parameters.WithShortFlag("v"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &FulltextCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the fulltext command
func init() {
	// Create the fulltext command
	fulltextCmd, err := NewFulltextCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create fulltext command")
	}

	// Convert to Cobra command
	ftCobraCmd, err := cli.BuildCobraCommandFromCommand(fulltextCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build fulltext cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(ftCobraCmd)
}
