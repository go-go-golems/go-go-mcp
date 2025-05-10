package scholarly

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/arxiv"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/crossref"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/openalex"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/querydsl"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// SearchProvider represents a source that can be searched
type SearchProvider string

const (
	ProviderArxiv    SearchProvider = "arxiv"
	ProviderCrossref SearchProvider = "crossref"
	ProviderOpenAlex SearchProvider = "openalex"
)

// SearchOptions configures a multi-provider search
type SearchOptions struct {
	Providers []SearchProvider // which providers to search (default: all)
	Mailto    string           // email for polite pool access
}

// Search performs a unified search across multiple providers using the query DSL
func Search(ctx context.Context, query *querydsl.Query, opts SearchOptions) ([]common.SearchResult, error) {
	if len(opts.Providers) == 0 {
		opts.Providers = []SearchProvider{ProviderArxiv, ProviderCrossref, ProviderOpenAlex}
	}

	// Create clients
	arxivClient := arxiv.NewClient()
	crossrefClient := crossref.NewClient(opts.Mailto)
	openalexClient := openalex.NewClient(opts.Mailto)

	// Use errgroup to handle concurrent searches
	g, ctx := errgroup.WithContext(ctx)

	// Channel for collecting results
	resultsChan := make(chan []common.SearchResult, len(opts.Providers))

	// Convert DSL query to provider-specific params
	for _, provider := range opts.Providers {
		provider := provider // capture for goroutine

		g.Go(func() error {
			var results []common.SearchResult
			var err error

			params := common.SearchParams{
				MaxResults: query.MaxResults,
			}

			switch provider {
			case ProviderArxiv:
				params.Query = query.ToArxiv()
				results, err = arxivClient.Search(params)

			case ProviderCrossref:
				values := query.ToCrossref()
				params.Query = values.Get("query")
				params.Filters = make(map[string]string)
				if filter := values.Get("filter"); filter != "" {
					params.Filters["filter"] = filter
				}
				results, err = crossrefClient.Search(params)

			case ProviderOpenAlex:
				values := query.ToOpenAlex()
				params.Query = values.Get("search")
				params.Filters = make(map[string]string)
				if filter := values.Get("filter"); filter != "" {
					params.Filters["filter"] = filter
				}
				if sort := values.Get("sort"); sort != "" {
					params.Filters["sort"] = sort
				}
				results, err = openalexClient.Search(params)
			}

			if err != nil {
				return errors.Wrapf(err, "search failed for provider %s", provider)
			}

			select {
			case resultsChan <- results:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	// Close results channel when all searches complete
	go func() {
		g.Wait()
		close(resultsChan)
	}()

	// Collect and merge results
	var allResults []common.SearchResult
	seen := make(map[string]struct{}) // track DOIs to avoid duplicates

	for results := range resultsChan {
		for _, result := range results {
			if result.DOI != "" {
				if _, exists := seen[result.DOI]; exists {
					continue
				}
				seen[result.DOI] = struct{}{}
			}
			allResults = append(allResults, result)
		}
	}

	// Check for any search errors
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return allResults, nil
}
