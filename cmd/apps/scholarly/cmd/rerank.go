package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/tools"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// RerankerCommand is a glazed command for reranking documents
type RerankerCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &RerankerCommand{}

// RerankerSettings holds the parameters for document reranking
type RerankerSettings struct {
	Query     string   `glazed.parameter:"query"`
	Documents []string `glazed.parameter:"documents"`
	Limit     int      `glazed.parameter:"limit"`
	URL       string   `glazed.parameter:"url"`
	Timeout   int      `glazed.parameter:"timeout"`
}

// RunIntoGlazeProcessor executes the document reranking and processes results
func (c *RerankerCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &RerankerSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate input - query and at least one document must be provided
	if s.Query == "" {
		return fmt.Errorf("query is required for reranking")
	}

	if len(s.Documents) == 0 {
		return fmt.Errorf("at least one document must be provided for reranking")
	}

	// Create reranker client
	timeout := time.Duration(s.Timeout) * time.Second
	rerankerClient := tools.NewRerankerClient(s.URL, timeout)

	// Convert string documents to search results
	results := make([]common.SearchResult, len(s.Documents))
	for i, doc := range s.Documents {
		results[i] = common.SearchResult{
			Title:    fmt.Sprintf("Document %d", i+1),
			Abstract: doc,
		}
	}

	// Check if reranker is available
	if !rerankerClient.IsRerankerAvailable(ctx) {
		return errors.New("reranker service is not available at " + s.URL)
	}

	// Perform reranking
	limit := s.Limit
	if limit <= 0 || limit > len(results) {
		limit = len(results)
	}

	rerankedResults, err := rerankerClient.Rerank(ctx, s.Query, results, limit)
	if err != nil {
		return errors.Wrap(err, "failed to rerank documents")
	}

	// Process results into rows
	for i, result := range rerankedResults {
		row := types.NewRow(
			types.MRP("rank", i+1),
			types.MRP("document", result.Abstract),
			types.MRP("score", result.RerankerScore),
			types.MRP("original_index", result.OriginalIndex),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewRerankerCommand creates a new reranker command
func NewRerankerCommand() (*RerankerCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"rerank",
		cmds.WithShort("Rerank documents based on relevance to a query"),
		cmds.WithLong(`Rerank a list of documents based on their relevance to a query using a neural reranker.

The reranker uses a cross-encoder model to compute relevance scores between the query and each document.

Examples:
  # Rerank a list of documents provided as separate arguments
  scholarly rerank --query "quantum computing" --documents "Document about quantum physics" --documents "Document about classical computing" --documents "Document about quantum algorithms"
  
  # The --documents flag must be repeated for each document (don't use commas between documents)
  scholarly rerank --query "climate change" --documents "Text about global warming" --documents "Text about carbon emissions" --limit 2
  
  # Use a custom reranker endpoint
  scholarly rerank --query "machine learning" --url "http://localhost:9000/rerank" --documents "Document about neural networks"`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("The query to rerank documents against"),
				parameters.WithShortFlag("q"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"documents",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("List of documents to rerank"),
				parameters.WithShortFlag("d"),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of results to return"),
				parameters.WithDefault(10),
				parameters.WithShortFlag("l"),
			),
			parameters.NewParameterDefinition(
				"url",
				parameters.ParameterTypeString,
				parameters.WithHelp("URL of the reranker service"),
				parameters.WithDefault("http://localhost:8000/rerank"),
			),
			parameters.NewParameterDefinition(
				"timeout",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Timeout in seconds for the reranker request"),
				parameters.WithDefault(10),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &RerankerCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the reranker command
func init() {
	// Create the reranker command
	rerankerCmd, err := NewRerankerCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create reranker command")
	}

	// Convert to Cobra command
	rCobraCmd, err := cli.BuildCobraCommandFromCommand(
		rerankerCmd,
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build reranker cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(rCobraCmd)
}
