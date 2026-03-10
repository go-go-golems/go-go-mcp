package cmd

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/tools"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
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
	DOI           string `glazed:"doi"`
	Title         string `glazed:"title"`
	PreferVersion string `glazed:"version"`
}

// RunIntoGlazeProcessor executes the full text search and processes results
func (c *FulltextCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *values.Values,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &FulltextSettings{}
	if err := parsedLayers.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	// Validate that we have at least DOI or title
	if s.DOI == "" && s.Title == "" {
		return fmt.Errorf("either DOI or title must be provided")
	}

	log.Debug().Str("doi", s.DOI).Str("title", s.Title).Str("version", s.PreferVersion).Msg("Finding full text")

	req := common.FindFullTextRequest{
		DOI:           s.DOI,
		Title:         s.Title,
		PreferVersion: s.PreferVersion,
	}

	// Find the full text
	response, err := tools.FindFullText(req)
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
	glazedLayer, err := settings.NewGlazedSection()
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
			fields.New(
				"doi",
				fields.TypeString,
				fields.WithHelp("DOI of the work"),
				fields.WithDefault(""),
				fields.WithShortFlag("i"),
			),
			fields.New(
				"title",
				fields.TypeString,
				fields.WithHelp("Title of the work"),
				fields.WithDefault(""),
				fields.WithShortFlag("t"),
			),
			fields.New(
				"version",
				fields.TypeString,
				fields.WithHelp("Preferred version (published, accepted, submitted)"),
				fields.WithDefault("published"),
				fields.WithShortFlag("v"),
			),
		),

		// Add parameter layers
		cmds.WithSections(
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
	ftCobraCmd, err := cli.BuildCobraCommandFromCommand(fulltextCmd,
		cli.WithCobraShortHelpSections(schema.DefaultSlug),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build fulltext cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(ftCobraCmd)
}
