package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/querydsl"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/tools"
	"github.com/rs/zerolog/log"
)

// SearchCommand is a glazed command for searching across scholarly providers
type SearchCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &SearchCommand{}

// SearchSettings holds the parameters for scholarly search
type SearchSettings struct {
	Query      string   `glazed.parameter:"query"`
	Sources    []string `glazed.parameter:"sources"`
	Limit      int      `glazed.parameter:"limit"`
	Author     string   `glazed.parameter:"author"`
	Title      string   `glazed.parameter:"title"`
	Category   string   `glazed.parameter:"category"`
	WorkType   string   `glazed.parameter:"work-type"`
	FromYear   int      `glazed.parameter:"from-year"`
	ToYear     int      `glazed.parameter:"to-year"`
	SortOrder  string   `glazed.parameter:"sort"`
	OpenAccess string   `glazed.parameter:"open-access"`
	Mailto     string   `glazed.parameter:"mailto"`
}

// RunIntoGlazeProcessor executes the scholarly search and processes results
func (c *SearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &SearchSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate input - either query or specific fields must be provided
	if s.Query == "" && s.Author == "" && s.Title == "" && s.Category == "" && s.WorkType == "" {
		return fmt.Errorf("either query or at least one specific search field must be provided")
	}

	// Validate source
	if !isValidSource(s.Sources) {
		return fmt.Errorf("invalid source: must be one of: arxiv, crossref, openalex, all")
	}

	log.Debug().
		Str("query", s.Query).
		Str("sources", strings.Join(s.Sources, ", ")).
		Int("limit", s.Limit).
		Msg("Search initiated with DSL")

	// Create query using the DSL
	query := querydsl.New()

	// Set basic text fields
	if s.Query != "" {
		query.WithText(s.Query)
	}
	if s.Author != "" {
		query.WithAuthor(s.Author)
	}
	if s.Title != "" {
		query.WithTitle(s.Title)
	}
	if s.Category != "" {
		query.WithCategory(s.Category)
	}
	if s.WorkType != "" {
		query.WithType(s.WorkType)
	}

	// Set year range if provided
	if s.FromYear > 0 || s.ToYear > 0 {
		query.Between(s.FromYear, s.ToYear)
	}

	// Set max results
	query.WithMaxResults(s.Limit)

	// Set sort order if provided
	if s.SortOrder != "" {
		var sortOrder querydsl.SortOrder
		switch strings.ToLower(s.SortOrder) {
		case "newest":
			sortOrder = querydsl.SortNewest
		case "oldest":
			sortOrder = querydsl.SortOldest
		default:
			sortOrder = querydsl.SortRelevance
		}
		query.Order(sortOrder)
	}

	// Set OpenAccess filter if provided
	if s.OpenAccess != "" {
		isOA := strings.ToLower(s.OpenAccess) == "true"
		query.OnlyOA(isOA)
	}

	// Determine which providers to search
	providers := getProvidersFromSources(s.Sources)

	// Create options
	opts := tools.SearchOptions{
		Providers: providers,
		Mailto:    s.Mailto,
	}

	// Create filters map for OpenAlex
	filters := make(map[string]string)
	if s.Author != "" {
		filters["author"] = s.Author
	}
	if s.FromYear > 0 {
		filters["from-publication_year"] = fmt.Sprintf("%d", s.FromYear)
	}

	// Perform search using the DSL
	results, err := tools.Search(ctx, query, opts)
	if err != nil {
		return err
	}

	// If no results, just return without error
	if len(results) == 0 {
		return nil
	}

	// Process results into rows
	for _, result := range results {
		// Parse year from published date (format: YYYY-MM-DD)
		year := 0
		if len(result.Published) >= 4 {
			year, _ = strconv.Atoi(result.Published[:4])
		}

		row := types.NewRow(
			types.MRP("id", result.SourceURL),
			types.MRP("doi", result.DOI),
			types.MRP("title", result.Title),
			types.MRP("authors", result.Authors),
			types.MRP("year", year),
			types.MRP("is_oa", result.OAStatus != ""),
			types.MRP("citation_count", result.Citations),
			types.MRP("abstract", result.Abstract),
			types.MRP("source_name", result.SourceName),
			types.MRP("pdf_url", result.PDFURL),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewSearchCommand creates a new search command
func NewSearchCommand() (*SearchCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"search",
		cmds.WithShort("Search for scholarly works across multiple sources"),
		cmds.WithLong(`Search for scholarly works across Arxiv, Crossref, and OpenAlex using the Query DSL.

Examples:
  # Basic search with just a query string
  scholarly search --query "quantum computing" --source arxiv
  
  # Using specific search fields
  scholarly search --author "Hinton" --title "deep learning" --source crossref
  
  # Filtering by date range and category
  scholarly search --query "neural networks" --from-year 2020 --to-year 2023 --category "cs.AI" --source arxiv
  
  # Sorting and limiting results
  scholarly search --query "climate change" --source openalex --limit 10 --sort newest
  
  # Filtering for open access content
  scholarly search --query "machine learning" --source crossref --open-access true`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("Main search query text"),
				parameters.WithShortFlag("q"),
			),
			parameters.NewParameterDefinition(
				"sources",
				parameters.ParameterTypeChoiceList,
				parameters.WithHelp("Sources to search from"),
				parameters.WithDefault([]string{"all"}),
				parameters.WithShortFlag("s"),
				parameters.WithChoices("arxiv", "crossref", "openalex", "all"),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of results to return"),
				parameters.WithDefault(20),
				parameters.WithShortFlag("l"),
			),
			parameters.NewParameterDefinition(
				"author",
				parameters.ParameterTypeString,
				parameters.WithHelp("Author name to search for"),
			),
			parameters.NewParameterDefinition(
				"title",
				parameters.ParameterTypeString,
				parameters.WithHelp("Title words/phrase to search for"),
			),
			parameters.NewParameterDefinition(
				"category",
				parameters.ParameterTypeString,
				parameters.WithHelp("ArXiv category (e.g., cs.AI)"),
			),
			parameters.NewParameterDefinition(
				"work-type",
				parameters.ParameterTypeString,
				parameters.WithHelp("Publication type (e.g., journal-article)"),
			),
			parameters.NewParameterDefinition(
				"from-year",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Starting year (inclusive)"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"to-year",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Ending year (inclusive)"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"sort",
				parameters.ParameterTypeString,
				parameters.WithHelp("Sort order: relevance, newest, oldest"),
			),
			parameters.NewParameterDefinition(
				"open-access",
				parameters.ParameterTypeString,
				parameters.WithHelp("Open access filter: true, false"),
			),
			parameters.NewParameterDefinition(
				"mailto",
				parameters.ParameterTypeString,
				parameters.WithHelp("Email address for OpenAlex polite pool (highly recommended)"),
				parameters.WithDefault("wesen@ruinwesen.com"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &SearchCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// isValidSource checks if the provided sources are valid
func isValidSource(sources []string) bool {
	validSources := []string{"arxiv", "crossref", "openalex", "all"}

	for _, source := range sources {
		source = strings.ToLower(source)
		isValid := false
		for _, valid := range validSources {
			if source == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return false
		}
	}
	return true
}

// getProvidersFromSources converts source strings to providers
func getProvidersFromSources(sources []string) []tools.SearchProvider {
	var providers []tools.SearchProvider
	for _, source := range sources {
		source = strings.ToLower(source)
		if source == "all" {
			return []tools.SearchProvider{
				tools.ProviderArxiv,
				tools.ProviderCrossref,
				tools.ProviderOpenAlex,
			}
		}
		switch source {
		case "arxiv":
			providers = append(providers, tools.ProviderArxiv)
		case "crossref":
			providers = append(providers, tools.ProviderCrossref)
		case "openalex":
			providers = append(providers, tools.ProviderOpenAlex)
		}
	}
	return providers
}

// init registers the search command
func init() {
	// Create the search command
	searchCmd, err := NewSearchCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create search command")
	}

	// Convert to Cobra command
	sCobraCmd, err := cli.BuildCobraCommandFromCommand(
		searchCmd,
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build search cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(sCobraCmd)
}
