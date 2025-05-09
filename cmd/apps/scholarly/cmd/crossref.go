package cmd

import (
	"context"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/crossref"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
)

// CrossrefCommand is a glazed command for searching Crossref
type CrossrefCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &CrossrefCommand{}

// CrossrefSettings holds the parameters for Crossref search
type CrossrefSettings struct {
	Query  string `glazed.parameter:"query"`
	Rows   int    `glazed.parameter:"rows"`
	Mailto string `glazed.parameter:"mailto"`
	Filter string `glazed.parameter:"cr_filter"`
}

// RunIntoGlazeProcessor executes the Crossref search and processes results
func (c *CrossrefCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &CrossrefSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate query
	if s.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Warn if mailto is empty
	if s.Mailto == "" {
		log.Warn().Msg("No mailto parameter provided for Crossref API polite pool.")

		// Add warning row for no mailto
		warningRow := types.NewRow(
			types.MRP("_type", "warning"),
			types.MRP("message", "It is recommended to provide an email address using --mailto for the Crossref polite pool."),
		)
		if err := gp.AddRow(ctx, warningRow); err != nil {
			return err
		}
	}

	log.Debug().Str("query", s.Query).Int("rows", s.Rows).Str("mailto", s.Mailto).Str("filter", s.Filter).Msg("Crossref search initiated")

	// Search Crossref
	client := crossref.NewClient(s.Mailto)
	params := common.SearchParams{
		Query:      s.Query,
		MaxResults: s.Rows,
		Filters:    make(map[string]string),
		EmailAddr:  s.Mailto,
	}

	// Only add filter if not empty
	if s.Filter != "" {
		params.Filters["filter"] = s.Filter
	}

	results, err := client.Search(params)
	if err != nil {
		log.Error().Err(err).Msg("Crossref search failed")
		return err
	}

	if len(results) == 0 {
		log.Info().Msg("No results in Crossref response")
		return nil
	}

	// Process results into rows
	for _, result := range results {
		row := types.NewRow(
			types.MRP("title", result.Title),
			types.MRP("doi", result.DOI),
			types.MRP("type", result.Type),
		)

		// Add optional fields if present
		if result.SourceURL != "" {
			row.Set("url", result.SourceURL)
		}

		if len(result.Authors) > 0 {
			row.Set("authors", result.Authors)
		}

		if publisher, ok := result.Metadata["publisher"].(string); ok {
			row.Set("publisher", publisher)
		}

		if result.Published != "" {
			row.Set("published", result.Published)
		}

		if result.Abstract != "" {
			row.Set("abstract", result.Abstract)
		}

		if result.PDFURL != "" {
			row.Set("pdf_url", result.PDFURL)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewCrossrefCommand creates a new Crossref command
func NewCrossrefCommand() (*CrossrefCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"crossref",
		cmds.WithShort("Search for scholarly works on Crossref"),
		cmds.WithLong(`Search for scholarly works (articles, books, datasets, etc.) using the Crossref REST API.

Example:
  arxiv-libgen-cli crossref --query "climate change adaptation" --rows 5 --mailto "your.email@example.com"
  arxiv-libgen-cli crossref -q "consciousness" -n 3 -f "from-pub-date:2022,type:journal-article" -m "user@example.org"`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("Search query for Crossref (required)"),
				parameters.WithRequired(true),
				parameters.WithShortFlag("q"),
			),
			parameters.NewParameterDefinition(
				"rows",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Number of results to return"),
				parameters.WithDefault(10),
				parameters.WithShortFlag("n"),
			),
			parameters.NewParameterDefinition(
				"mailto",
				parameters.ParameterTypeString,
				parameters.WithHelp("Email address for Crossref polite pool (recommended)"),
				parameters.WithDefault(""),
				parameters.WithShortFlag("m"),
			),
			parameters.NewParameterDefinition(
				"cr_filter",
				parameters.ParameterTypeString,
				parameters.WithHelp("Filter parameters for Crossref (e.g., from-pub-date:2020,type:journal-article)"),
				parameters.WithDefault(""),
				parameters.WithShortFlag("f"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &CrossrefCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the crossref command
func init() {
	// Create the crossref command
	crossrefCmd, err := NewCrossrefCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create crossref command")
	}

	// Convert to Cobra command
	cfCobraCmd, err := cli.BuildCobraCommandFromCommand(crossrefCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build crossref cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(cfCobraCmd)
}
