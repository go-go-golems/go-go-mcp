package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/arxiv"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
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
	Query      string `glazed.parameter:"query"`
	MaxResults int    `glazed.parameter:"max_results"`
}

// RunIntoGlazeProcessor executes the arxiv search and processes results
func (c *ArxivCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &ArxivSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
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
	glazedLayer, err := settings.NewGlazedParameterLayers()
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
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("Search query for Arxiv (e.g., 'all:electron', 'ti:\"quantum computing\" AND au:\"John Preskill\"') (required)"),
				parameters.WithRequired(true),
				parameters.WithShortFlag("q"),
			),
			parameters.NewParameterDefinition(
				"max_results",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of results to return"),
				parameters.WithDefault(10),
				parameters.WithShortFlag("n"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
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
	axCobraCmd, err := cli.BuildCobraCommandFromCommand(arxivCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build arxiv cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(axCobraCmd)
}
