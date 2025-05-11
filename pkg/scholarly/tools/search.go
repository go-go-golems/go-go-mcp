package tools

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/arxiv"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/crossref"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/openalex"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
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
	Providers   []SearchProvider // which providers to search (default: all)
	Mailto      string           // email for polite pool access
	UseReranker bool             // whether to use the reranker (default: true)
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
				values := query.ToArxiv()
				// Don't URL-encode the search_query since it's already properly formatted
				// The raw query string has the correct syntax for ArXiv API
				params.Query = values.Get("search_query")
				params.Filters = make(map[string]string)

				// Extract sort parameters
				if sortBy := values.Get("sortBy"); sortBy != "" {
					params.Filters["sortBy"] = sortBy
				}
				if sortOrder := values.Get("sortOrder"); sortOrder != "" {
					params.Filters["sortOrder"] = sortOrder
				}
				results, err = arxivClient.Search(params)

			case ProviderCrossref:
				values := query.ToCrossref()
				params.Query = values.Get("query")
				params.Filters = make(map[string]string)

				// Extract filter
				if filter := values.Get("filter"); filter != "" {
					params.Filters["filter"] = filter
				}

				// Extract author and title specific query params
				if author := values.Get("query.author"); author != "" {
					params.Filters["crossref_query.author"] = author
				}
				if title := values.Get("query.title"); title != "" {
					params.Filters["crossref_query.title"] = title
				}

				// Extract sort parameters
				if sort := values.Get("sort"); sort != "" {
					params.Filters["crossref_sort"] = sort
				}
				if order := values.Get("order"); order != "" {
					params.Filters["crossref_order"] = order
				}

				results, err = crossrefClient.Search(params)

			case ProviderOpenAlex:
				values := query.ToOpenAlex()
				params.Query = values.Get("search")
				params.Filters = make(map[string]string)

				// Build filter string
				var filters []string

				// Let the client handle author lookup and filtering
				if query.Author != "" {
					params.Filters["author"] = query.Author
				}

				// Handle year filter
				if query.FromYear > 0 {
					// XXX: This was publication_year:%d before, ensure this is what OpenAlex client expects
					// The client.go Search function looks for "from-publication_year"
					params.Filters["from-publication_year"] = fmt.Sprintf("%d", query.FromYear)
				}

				// Handle work type
				if query.WorkType != "" {
					filters = append(filters, fmt.Sprintf("type:%s", query.WorkType))
				}

				// Handle open access filter
				if query.OpenAccess != nil {
					filters = append(filters, fmt.Sprintf("is_oa:%v", *query.OpenAccess))
				}

				// Combine additional filters (work type, OA status)
				if len(filters) > 0 {
					// Check if a base filter string already exists from ToOpenAlex() (e.g. from raw query)
					if existingFilter := values.Get("filter"); existingFilter != "" {
						params.Filters["filter"] = existingFilter + "," + strings.Join(filters, ",")
					} else {
						params.Filters["filter"] = strings.Join(filters, ",")
					}
				}

				// Handle sort order
				switch query.Sort {
				case querydsl.SortNewest:
					params.Filters["sort"] = "publication_date:desc"
				case querydsl.SortOldest:
					params.Filters["sort"] = "publication_date:asc"
				case querydsl.SortRelevance:
					fallthrough
				default:
					// If sort was already set by ToOpenAlex (e.g. from raw query), don't override
					if _, ok := params.Filters["sort"]; !ok {
						params.Filters["sort"] = "relevance_score:desc"
					}
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
		_ = g.Wait()
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

	// Apply reranking if enabled (default is true)
	if opts.UseReranker || len(opts.Providers) > 0 { // Default to true if not explicitly set
		// Create reranker client
		rerankerClient := NewRerankerClient("", 0)

		// Check if reranker is available
		if rerankerClient.IsRerankerAvailable(ctx) {
			// Only rerank if we have a text query
			if query.Text != "" {
				rerankedResults, err := rerankerClient.Rerank(ctx, query.Text, allResults, query.MaxResults)
				if err == nil {
					return rerankedResults, nil
				}
				// If reranking fails, fall back to original results
				// Log the error but continue with original results
				fmt.Printf("Warning: reranking failed: %v, falling back to original results\n", err)
			}
		}
	}

	return allResults, nil
}
