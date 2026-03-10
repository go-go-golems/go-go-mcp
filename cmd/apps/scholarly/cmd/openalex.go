package cmd

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/openalex"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/rs/zerolog/log"
)

// OpenAlexCommand is a glazed command for searching OpenAlex
type OpenAlexCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &OpenAlexCommand{}

// OpenAlexSettings holds the parameters for OpenAlex search
type OpenAlexSettings struct {
	Query   string `glazed:"query"`
	PerPage int    `glazed:"per_page"`
	Mailto  string `glazed:"mailto"`
	Filter  string `glazed:"oa_filter"`
	Sort    string `glazed:"oa_sort"`
}

// RunIntoGlazeProcessor executes the OpenAlex search and processes results
func (c *OpenAlexCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *values.Values,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &OpenAlexSettings{}
	if err := parsedLayers.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	// Validate query or filter
	if s.Query == "" && s.Filter == "" {
		return fmt.Errorf("query or filter must be provided")
	}

	// Warn if mailto is empty
	if s.Mailto == "" {
		log.Warn().Msg("No mailto parameter provided for OpenAlex API polite pool.")

		// Add warning row
		warningRow := types.NewRow(
			types.MRP("_type", "warning"),
			types.MRP("message", "It is highly recommended to provide an email address using --mailto for the OpenAlex polite pool."),
		)
		if err := gp.AddRow(ctx, warningRow); err != nil {
			return err
		}
	}

	log.Debug().Str("query", s.Query).Int("per_page", s.PerPage).Str("mailto", s.Mailto).Str("filter", s.Filter).Str("sort", s.Sort).Msg("OpenAlex search initiated")

	// Search OpenAlex
	client := openalex.NewClient(s.Mailto)
	params := common.SearchParams{
		Query:      s.Query,
		MaxResults: s.PerPage,
		Filters:    make(map[string]string),
		EmailAddr:  s.Mailto,
	}

	// Only add filter and sort if not empty
	if s.Filter != "" {
		params.Filters["filter"] = s.Filter
	}

	if s.Sort != "" {
		params.Filters["sort"] = s.Sort
	}

	results, err := client.Search(params)
	if err != nil {
		log.Error().Err(err).Msg("OpenAlex search failed")
		return err
	}

	if len(results) == 0 {
		log.Info().Msg("No results in OpenAlex response")
		return nil
	}

	// Process results into rows
	for _, result := range results {
		row := types.NewRow(
			types.MRP("title", result.Title),
			types.MRP("openalex_id", result.SourceURL),
			types.MRP("authors", result.Authors),
			types.MRP("publication_date", result.Published),
			types.MRP("type", result.Type),
			types.MRP("citations", result.Citations),
		)

		// Add optional fields if present
		if result.DOI != "" {
			row.Set("doi", result.DOI)
		}

		if result.JournalInfo != "" {
			row.Set("venue", result.JournalInfo)
		}

		if result.OAStatus != "" {
			row.Set("open_access_status", result.OAStatus)
		}

		if result.PDFURL != "" {
			row.Set("pdf_url", result.PDFURL)
		}

		if result.License != "" {
			row.Set("license", result.License)
		}

		if result.Abstract != "" {
			row.Set("abstract", result.Abstract)
		}

		if relevance, ok := result.Metadata["relevance_score"].(float64); ok && relevance > 0 {
			row.Set("relevance_score", relevance)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewOpenAlexCommand creates a new OpenAlex command
func NewOpenAlexCommand() (*OpenAlexCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedSection()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"openalex",
		cmds.WithShort("Search for scholarly works on OpenAlex"),
		cmds.WithLong(`Search for scholarly works (articles, books, datasets, etc.) using the OpenAlex API.

Example:
  arxiv-libgen-cli openalex --query "machine learning applications" --per_page 5 --mailto "your.email@example.com"
  arxiv-libgen-cli openalex -q "bioinformatics" -n 3 -f "publication_year:2022,type:journal-article" -s "cited_by_count:desc" -m "user@example.org"`),

		// Define command flags
		cmds.WithFlags(
			fields.New(
				"query",
				fields.TypeString,
				fields.WithHelp("Search query for OpenAlex (searches title, abstract, fulltext)"),
				fields.WithDefault(""),
				fields.WithShortFlag("q"),
			),
			fields.New(
				"per_page",
				fields.TypeInteger,
				fields.WithHelp("Number of results per page"),
				fields.WithDefault(10),
				fields.WithShortFlag("n"),
			),
			fields.New(
				"mailto",
				fields.TypeString,
				fields.WithHelp("Email address for OpenAlex polite pool (highly recommended)"),
				fields.WithDefault(""),
				fields.WithShortFlag("m"),
			),
			fields.New(
				"oa_filter",
				fields.TypeString,
				fields.WithHelp("Filter parameters for OpenAlex (e.g., publication_year:2022,type:journal-article)"),
				fields.WithDefault(""),
				fields.WithShortFlag("f"),
			),
			fields.New(
				"oa_sort",
				fields.TypeString,
				fields.WithHelp("Sort order (e.g., cited_by_count:desc, publication_date:asc)"),
				fields.WithDefault("relevance_score:desc"),
				fields.WithShortFlag("s"),
			),
		),

		// Add parameter layers
		cmds.WithSections(
			glazedLayer,
		),
	)

	return &OpenAlexCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the openalex command
func init() {
	// Create the openalex command
	openalexCmd, err := NewOpenAlexCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create openalex command")
	}

	// Convert to Cobra command
	oaCobraCmd, err := cli.BuildCobraCommandFromCommand(openalexCmd,
		cli.WithCobraShortHelpSections(schema.DefaultSlug),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build openalex cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(oaCobraCmd)
}
