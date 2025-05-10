package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/querydsl"
	scholarly_tools "github.com/go-go-golems/go-go-mcp/pkg/scholarly/tools"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/pkg/errors"
)

// registerSearchWorksTool registers the search_works tool with the new querydsl implementation
func registerSearchWorksTool(registry *tool_registry.Registry) error {
	schemaJSON := `{
 		"type": "object",
 		"properties": {
 			"queries": {
 				"type": "array",
 				"items": {
 					"type": "object",
 					"properties": {
 						"text": {
 							"type": "string",
 							"description": "Free-text search query (e.g. 'quantum computing applications')"
 						},
 						"author": {
 							"type": "string",
 							"description": "Author name to search for"
 						},
 						"title": {
 							"type": "string",
 							"description": "Search for words or phrases in the title"
 						},
 						"category": {
 							"type": "string",
 							"description": "ArXiv category (e.g., 'cs.AI', 'physics.comp-ph')"
 						},
 						"work_type": {
 							"type": "string",
 							"description": "Document type (e.g., 'journal-article', 'preprint')"
 						},
 						"from_year": {
 							"type": "integer",
 							"description": "Start year for date range filter (inclusive)"
 						},
 						"to_year": {
 							"type": "integer",
 							"description": "End year for date range filter (inclusive)"
 						},
 						"open_access": {
 							"type": "boolean",
 							"description": "Filter for open access content only"
 						},
 						"sort": {
 							"type": "string",
 							"enum": ["relevance", "newest", "oldest"],
 							"default": "relevance",
 							"description": "Sort order for results"
 						},
 						"providers": {
 							"type": "array",
 							"items": {
 								"type": "string",
 								"enum": ["arxiv", "openalex", "crossref"]
 							},
 							"description": "List of providers to search (default: all)"
 						},
 						"max_results": {
 							"type": "integer",
 							"minimum": 1,
 							"maximum": 100,
 							"default": 20,
 							"description": "Maximum number of results to return (1-100)"
 						},
 						"original_query_intent": {
 							"type": "string",
 							"description": "The human language intent that caused this tool to be called and why"
 						}
 					},
 					"required": ["text"]
 				},
 				"description": "Array of search queries to process in batch"
 			},
 			"text": {
 				"type": "string",
 				"description": "Free-text search query (single query mode)"
 			},
 			"author": {
 				"type": "string",
 				"description": "Author name to search for (single query mode)"
 			},
 			"title": {
 				"type": "string",
 				"description": "Search for words or phrases in the title (single query mode)"
 			},
 			"category": {
 				"type": "string",
 				"description": "ArXiv category (e.g., 'cs.AI', 'physics.comp-ph') (single query mode)"
 			},
 			"work_type": {
 				"type": "string",
 				"description": "Document type (e.g., 'journal-article', 'preprint') (single query mode)"
 			},
 			"from_year": {
 				"type": "integer",
 				"description": "Start year for date range filter (inclusive) (single query mode)"
 			},
 			"to_year": {
 				"type": "integer",
 				"description": "End year for date range filter (inclusive) (single query mode)"
 			},
 			"open_access": {
 				"type": "boolean",
 				"description": "Filter for open access content only (single query mode)"
 			},
 			"sort": {
 				"type": "string",
 				"enum": ["relevance", "newest", "oldest"],
 				"default": "relevance",
 				"description": "Sort order for results (single query mode)"
 			},
 			"providers": {
 				"type": "array",
 				"items": {
 					"type": "string",
 					"enum": ["arxiv", "openalex", "crossref"]
 				},
 				"description": "List of providers to search (default: all) (single query mode)"
 			},
 			"max_results": {
 				"type": "integer",
 				"minimum": 1,
 				"maximum": 100,
 				"default": 20,
 				"description": "Maximum number of results to return (1-100) (single query mode)"
 			},
 			"original_query_intent": {
 				"type": "string",
 				"description": "The human language intent that caused this tool to be called and why"
 			}
 		}
 	}`

	tool, err := tools.NewToolImpl(
		"scholarly_search_works",
		"Search for scholarly works across academic databases including arXiv (pre-prints), OpenAlex (open access research), and Crossref (comprehensive scholarly records). Returns metadata about matching papers including titles, authors, and publication details. Don't use this tool to find citations or cited works, use the citations tool.",
		json.RawMessage(schemaJSON),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create search_works tool")
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Check if we have batch queries
			if queriesVal, ok := arguments["queries"].([]interface{}); ok {
				// Process batch queries
				batchResponses := make([]interface{}, 0, len(queriesVal))

				for _, queryObj := range queriesVal {
					queryMap, ok := queryObj.(map[string]interface{})
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("each query in queries must be an object"),
						), nil
					}

					// Build query from parameters
					query, err := buildQuery(queryMap)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error building query: %v", err)),
						), nil
					}

					// Get providers
					opts, err := buildSearchOptions(queryMap)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error building search options: %v", err)),
						), nil
					}

					// Execute search
					results, err := scholarly_tools.Search(ctx, query, opts)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error searching works: %v", err)),
						), nil
					}

					response := &SearchResponse{Results: results}
					batchResponses = append(batchResponses, response)
				}

				return protocol.NewToolResult(
					protocol.WithJSON(batchResponses),
				), nil
			}

			// Handle single query case
			query, err := buildQuery(arguments)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error building query: %v", err)),
				), nil
			}

			// Get providers
			opts, err := buildSearchOptions(arguments)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error building search options: %v", err)),
				), nil
			}

			// Execute search
			results, err := scholarly_tools.Search(ctx, query, opts)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error searching works: %v", err)),
				), nil
			}

			response := &SearchResponse{Results: results}
			return protocol.NewToolResult(
				protocol.WithJSON(response),
			), nil
		})

	return nil
}

// buildQuery creates a querydsl.Query from a map of parameters
func buildQuery(params map[string]interface{}) (*querydsl.Query, error) {
	query := querydsl.New()

	// Process text search
	if text, ok := params["text"].(string); ok && text != "" {
		query.WithText(text)
	}

	// Process author
	if author, ok := params["author"].(string); ok && author != "" {
		query.WithAuthor(author)
	}

	// Process title
	if title, ok := params["title"].(string); ok && title != "" {
		query.WithTitle(title)
	}

	// Process category
	if category, ok := params["category"].(string); ok && category != "" {
		query.WithCategory(category)
	}

	// Process work type
	if workType, ok := params["work_type"].(string); ok && workType != "" {
		query.WithType(workType)
	}

	// Process year range
	fromYear := 0
	toYear := 0

	if fromYearVal, ok := params["from_year"].(float64); ok {
		fromYear = int(fromYearVal)
	}

	if toYearVal, ok := params["to_year"].(float64); ok {
		toYear = int(toYearVal)
	}

	if fromYear > 0 || toYear > 0 {
		query.Between(fromYear, toYear)
	}

	// Process open access flag
	if openAccess, ok := params["open_access"].(bool); ok {
		query.OnlyOA(openAccess)
	}

	// Process sort order
	if sortVal, ok := params["sort"].(string); ok {
		switch sortVal {
		case "newest":
			query.Order(querydsl.SortNewest)
		case "oldest":
			query.Order(querydsl.SortOldest)
		default: // relevance is default
			query.Order(querydsl.SortRelevance)
		}
	}

	// Process max results
	if maxResults, ok := params["max_results"].(float64); ok && maxResults > 0 {
		query.WithMaxResults(int(maxResults))
	}

	return query, nil
}

// buildSearchOptions creates a SearchOptions from a map of parameters
func buildSearchOptions(params map[string]interface{}) (scholarly_tools.SearchOptions, error) {
	opts := scholarly_tools.SearchOptions{}

	// Process providers
	if providersVal, ok := params["providers"].([]interface{}); ok && len(providersVal) > 0 {
		providers := make([]scholarly_tools.SearchProvider, 0, len(providersVal))

		for _, p := range providersVal {
			if providerStr, ok := p.(string); ok {
				switch providerStr {
				case "arxiv":
					providers = append(providers, scholarly_tools.ProviderArxiv)
				case "crossref":
					providers = append(providers, scholarly_tools.ProviderCrossref)
				case "openalex":
					providers = append(providers, scholarly_tools.ProviderOpenAlex)
				default:
					return opts, fmt.Errorf("unknown provider: %s", providerStr)
				}
			}
		}

		if len(providers) > 0 {
			opts.Providers = providers
		}
	}

	return opts, nil
}
