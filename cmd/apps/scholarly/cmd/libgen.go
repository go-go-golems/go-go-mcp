package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/libgen"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
)

// LibgenCommand is a glazed command for searching LibGen
type LibgenCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &LibgenCommand{}

// LibgenSettings holds the parameters for LibGen search
type LibgenSettings struct {
	Query      string `glazed.parameter:"query"`
	MaxResults int    `glazed.parameter:"max_results"`
	Mirror     string `glazed.parameter:"mirror"`
}

// RunIntoGlazeProcessor executes the LibGen search and processes results
func (c *LibgenCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &LibgenSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate query
	if s.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	log.Debug().Str("query", s.Query).Int("max_results", s.MaxResults).Str("mirror", s.Mirror).Msg("LibGen search initiated")

	// add info as a comment row
	infoRow := types.NewRow(
		types.MRP("_type", "info"),
		types.MRP("message", "Note: LibGen search is via web scraping and may be slow or break if the site structure changes."),
	)
	if err := gp.AddRow(ctx, infoRow); err != nil {
		return err
	}

	// Search LibGen
	client := libgen.NewClient(s.Mirror)
	params := common.SearchParams{
		Query:      s.Query,
		MaxResults: s.MaxResults,
	}

	results, err := client.Search(params)
	if err != nil {
		log.Error().Err(err).Msg("LibGen search failed")
		return err
	}

	if len(results) == 0 {
		log.Info().Msg("No entries parsed from LibGen HTML")

		// add warning as a comment row
		warningRow := types.NewRow(
			types.MRP("_type", "warning"),
			types.MRP("message", "No results found, or failed to parse results from the mirror. The structure of LibGen mirrors changes frequently. This scraper might need an update."),
		)
		return gp.AddRow(ctx, warningRow)
	}

	log.Debug().Int("num_entries_parsed", len(results)).Msg("Finished parsing LibGen results")

	// Process results into rows
	maxShow := min(len(results), s.MaxResults)
	for i := 0; i < maxShow; i++ {
		result := results[i]

		// Create a base row with common fields
		row := types.NewRow(
			types.MRP("title", result.Title),
			types.MRP("authors", strings.Join(result.Authors, ", ")),
			types.MRP("source_url", result.SourceURL),
			types.MRP("pdf_url", result.PDFURL),
		)

		// Add optional fields if present
		if result.JournalInfo != "" {
			row.Set("journal_info", result.JournalInfo)
		}

		if result.DOI != "" {
			row.Set("doi", result.DOI)
		}

		if result.FileSize != "" {
			row.Set("file_size", result.FileSize)
		}

		// Add metadata fields
		if editLink, ok := result.Metadata["edit_link"].(string); ok && editLink != "" {
			row.Set("edit_link", editLink)
		}

		if mirrorLinks, ok := result.Metadata["mirror_links"].([]string); ok && len(mirrorLinks) > 0 {
			row.Set("mirror_links", strings.Join(mirrorLinks, ", "))
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewLibgenCommand creates a new LibGen command
func NewLibgenCommand() (*LibgenCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"libgen",
		cmds.WithShort("Search for scientific papers on Library Genesis"),
		cmds.WithLong(`Search for scientific papers on Library Genesis.

Example:
  scholarly libgen --query "artificial intelligence" --max_results 5
  scholarly libgen -q "quantum entanglement" -n 3 --mirror "https://libgen.is"`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("Search query for LibGen (e.g., \"artificial intelligence\", \"ISBN:9783319994912\") (required)"),
				parameters.WithRequired(true),
				parameters.WithShortFlag("q"),
			),
			parameters.NewParameterDefinition(
				"max_results",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of results to display"),
				parameters.WithDefault(10),
				parameters.WithShortFlag("n"),
			),
			parameters.NewParameterDefinition(
				"mirror",
				parameters.ParameterTypeString,
				parameters.WithHelp("LibGen mirror URL (e.g., https://libgen.is, http://libgen.st)"),
				parameters.WithDefault("https://libgen.is"),
				parameters.WithShortFlag("m"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &LibgenCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the libgen command
func init() {
	// Create the libgen command
	libgenCmd, err := NewLibgenCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create libgen command")
	}

	// Convert to Cobra command
	lgCobraCmd, err := cli.BuildCobraCommandFromCommand(libgenCmd,
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build libgen cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(lgCobraCmd)
}
