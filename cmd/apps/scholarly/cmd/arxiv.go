package cmd

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/arxiv"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"

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

// ArxivCommand is a glazed command for searching arxiv
type ArxivCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ArxivCommand{}

// ArxivSettings holds the parameters for arxiv search
type ArxivSettings struct {
	Query      string `glazed:"query"`
	MaxResults int    `glazed:"max_results"`
}

// RunIntoGlazeProcessor executes the arxiv search and processes results
func (c *ArxivCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *values.Values,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &ArxivSettings{}
	if err := parsedLayers.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	// Validate query
	if s.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	log.Debug().Str("query", s.Query).Int("max_results", s.MaxResults).Msg("Arxiv search initiated")

	// Search Arxiv
	client := arxiv.NewClient()
	params := common.SearchParams{
		Query:      s.Query,
		MaxResults: s.MaxResults,
	}

	results, err := client.Search(params)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		log.Info().Msg("No results found in Arxiv response")
		return nil
	}

	// Process results into rows
	for _, result := range results {
		row := types.NewRow(
			types.MRP("title", result.Title),
			types.MRP("authors", strings.Join(result.Authors, ", ")),
			types.MRP("published", result.Published),
			types.MRP("source_url", result.SourceURL),
			types.MRP("pdf_url", result.PDFURL),
			types.MRP("abstract", result.Abstract),
		)

		// Add any additional metadata fields
		if updated, ok := result.Metadata["updated"].(string); ok {
			row.Set("updated", updated)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewArxivCommand creates a new arxiv command
func NewArxivCommand() (*ArxivCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedSection()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"arxiv",
		cmds.WithShort("Search for scientific papers on Arxiv"),
		cmds.WithLong(`Search for scientific papers on Arxiv using its public API.

Example:
  arxiv-libgen-cli arxiv --query "all:electron" --max_results 5
  arxiv-libgen-cli arxiv -q "ti:large language models AND au:Hinton" -n 3

Thank you to arXiv for use of its open access interoperability.`),

		// Define command flags
		cmds.WithFlags(
			fields.New(
				"query",
				fields.TypeString,
				fields.WithHelp("Search query for Arxiv (e.g., 'all:electron', 'ti:\"quantum computing\" AND au:\"John Preskill\"') (required)"),
				fields.WithRequired(true),
				fields.WithShortFlag("q"),
			),
			fields.New(
				"max_results",
				fields.TypeInteger,
				fields.WithHelp("Maximum number of results to return"),
				fields.WithDefault(10),
				fields.WithShortFlag("n"),
			),
		),

		// Add parameter layers
		cmds.WithSections(
			glazedLayer,
		),
	)

	return &ArxivCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the arxiv command
func init() {
	// Create the arxiv command
	arxivCmd, err := NewArxivCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create arxiv command")
	}

	// Convert to Cobra command
	axCobraCmd, err := cli.BuildCobraCommandFromCommand(
		arxivCmd,
		cli.WithCobraShortHelpSections(schema.DefaultSlug),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build arxiv cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(axCobraCmd)
}
