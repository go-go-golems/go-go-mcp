package scholarly

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/scholarly"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/pkg/errors"
)

// truncateText helps create more readable error messages by truncating long text
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

// RegisterScholarlyTools registers all scholarly tools with the provided registry
func RegisterScholarlyTools(registry *tool_registry.Registry) error {
	// Register all scholarly tools
	if err := registerSearchWorksTool(registry); err != nil {
		return errors.Wrap(err, "failed to register search_works tool")
	}

	if err := registerResolveDOITool(registry); err != nil {
		return errors.Wrap(err, "failed to register resolve_doi tool")
	}

	// if err := registerSuggestKeywordsTool(registry); err != nil {
	// 	return errors.Wrap(err, "failed to register suggest_keywords tool")
	// }

	// if err := registerGetMetricsTool(registry); err != nil {
	// 	return errors.Wrap(err, "failed to register get_metrics tool")
	// }

	// if err := registerGetCitationsTool(registry); err != nil {
	// 	return errors.Wrap(err, "failed to register get_citations tool")
	// }

	// if err := registerFindFullTextTool(registry); err != nil {
	// 	return errors.Wrap(err, "failed to register find_full_text tool")
	// }

	return nil
}

// registerSearchWorksTool registers the search_works tool
func registerSearchWorksTool(registry *tool_registry.Registry) error {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"queries": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"query": {
							"type": "string",
							"description": "Search query text (e.g. 'quantum computing applications')"
						},
						"source": {
							"type": "string",
							"enum": ["arxiv", "openalex", "crossref"],
							"description": "Source database to search: 'arxiv' for pre-prints, 'openalex' for open access research, 'crossref' for comprehensive scholarly records"
						},
						"limit": {
							"type": "integer",
							"minimum": 1,
							"maximum": 100,
							"default": 20,
							"description": "Maximum number of results to return (1-100)"
						},
						"filter": {
							"type": "object",
							"description": "Optional source-specific filters as key-value pairs. For arxiv: category, author, title. For openalex: is_oa, publication_year, has_doi. For crossref: type, from-pub-date, until-pub-date"
						}
					},
					"required": ["query", "source"]
				},
				"description": "Array of search queries to process in batch"
			},
			"query": {
				"type": "string",
				"description": "Single search query text (legacy parameter, use queries for batch processing)"
			},
			"source": {
				"type": "string",
				"enum": ["arxiv", "openalex", "crossref"],
				"description": "Source database for single query (legacy parameter, use queries for batch processing)"
			},
			"limit": {
				"type": "integer",
				"minimum": 1,
				"maximum": 100,
				"default": 20,
				"description": "Maximum results for single query (legacy parameter, use queries for batch processing)"
			},
			"filter": {
				"type": "object",
				"description": "Filters for single query (legacy parameter, use queries for batch processing)"
			}
		}
	}`

	tool, err := tools.NewToolImpl(
		"scholarly_search_works",
		"Search for scholarly works across academic databases including arXiv (pre-prints), OpenAlex (open access research), and Crossref (comprehensive scholarly records). Returns metadata about matching papers including titles, authors, and publication details.",
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

					// Extract parameters for this query
					query, ok := queryMap["query"].(string)
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("query field must be a string in each query object"),
						), nil
					}

					source, ok := queryMap["source"].(string)
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("source field must be a string in each query object"),
						), nil
					}

					// Get optional limit parameter with default value
					limit := 20
					if limitVal, ok := queryMap["limit"].(float64); ok {
						limit = int(limitVal)
					}

					// Get optional filter parameter and convert to expected format
					filterMap := make(map[string]string)
					if filterVal, ok := queryMap["filter"].(map[string]interface{}); ok {
						for k, v := range filterVal {
							switch val := v.(type) {
							case string:
								filterMap[k] = val
							case float64:
								filterMap[k] = fmt.Sprintf("%v", val)
							case bool:
								filterMap[k] = fmt.Sprintf("%v", val)
							}
						}
					}

					// Prepare the request
					req := scholarly.SearchWorksRequest{
						Query:  query,
						Source: source,
						Limit:  limit,
						Filter: filterMap,
					}

					// Call the scholarly search function
					response, err := scholarly.SearchWorks(req)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error searching works: %v", err)),
						), nil
					}

					batchResponses = append(batchResponses, response)
				}

				return protocol.NewToolResult(
					protocol.WithJSON(batchResponses),
				), nil
			}

			// Handle legacy single query case
			query, ok := arguments["query"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("query argument must be a string"),
				), nil
			}

			source, ok := arguments["source"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("source argument must be a string"),
				), nil
			}

			// Get optional limit parameter with default value
			limit := 20
			if limitVal, ok := arguments["limit"].(float64); ok {
				limit = int(limitVal)
			}

			// Get optional filter parameter and convert to expected format
			filterMap := make(map[string]string)
			if filterVal, ok := arguments["filter"].(map[string]interface{}); ok {
				for k, v := range filterVal {
					switch val := v.(type) {
					case string:
						filterMap[k] = val
					case float64:
						filterMap[k] = fmt.Sprintf("%v", val)
					case bool:
						filterMap[k] = fmt.Sprintf("%v", val)
					}
				}
			}

			// Prepare the request
			req := scholarly.SearchWorksRequest{
				Query:  query,
				Source: source,
				Limit:  limit,
				Filter: filterMap,
			}

			// Call the scholarly search function
			response, err := scholarly.SearchWorks(req)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error searching works: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithJSON(response),
			), nil
		})

	return nil
}

// registerResolveDOITool registers the resolve_doi tool
func registerResolveDOITool(registry *tool_registry.Registry) error {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"dois": {
				"type": "array",
				"items": {
					"type": "string",
					"pattern": "^10\\..+/.+"
				},
				"description": "Array of Digital Object Identifiers (DOIs) to resolve in batch (e.g., ['10.1038/nphys1170', '10.1109/5.771073'])"
			},
			"doi": {
				"type": "string",
				"pattern": "^10\\..+/.+",
				"description": "Digital Object Identifier (DOI) for a single scholarly work (legacy parameter, use dois for batch processing)"
			}
		}
	}`

	tool, err := tools.NewToolImpl(
		"scholarly_resolve_doi",
		"Retrieve comprehensive metadata for a scholarly work using its DOI (Digital Object Identifier). This tool combines data from multiple sources to provide detailed information about the publication, including authors, abstract, journal, citations, and open access status.",
		json.RawMessage(schemaJSON),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create resolve_doi tool")
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Check if we have batch DOIs
			if doisVal, ok := arguments["dois"].([]interface{}); ok {
				// Process batch DOIs
				batchResults := make([]interface{}, 0, len(doisVal))

				for _, doiObj := range doisVal {
					doi, ok := doiObj.(string)
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("each DOI in dois array must be a string"),
						), nil
					}

					// Create request for this DOI
					req := scholarly.ResolveDOIRequest{
						DOI: doi,
					}

					// Call the scholarly DOI resolution function
					work, err := scholarly.ResolveDOI(req)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error resolving DOI %s: %v", doi, err)),
						), nil
					}

					batchResults = append(batchResults, work)
				}

				return protocol.NewToolResult(
					protocol.WithJSON(batchResults),
				), nil
			}

			// Handle legacy single DOI case
			doi, ok := arguments["doi"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("doi argument must be a string"),
				), nil
			}

			// Create request
			req := scholarly.ResolveDOIRequest{
				DOI: doi,
			}

			// Call the scholarly DOI resolution function
			work, err := scholarly.ResolveDOI(req)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error resolving DOI: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithJSON(work),
			), nil
		})

	return nil
}

// registerSuggestKeywordsTool registers the suggest_keywords tool
func registerSuggestKeywordsTool(registry *tool_registry.Registry) error {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"texts": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"text": {
							"type": "string",
							"description": "Title, abstract or any text to analyze for keywords"
						},
						"max_keywords": {
							"type": "integer",
							"default": 10,
							"minimum": 1,
							"maximum": 50,
							"description": "Maximum number of keywords to return for this text (1-50)"
						}
					},
					"required": ["text"]
				},
				"description": "Array of text items to analyze for keywords in batch"
			},
			"text": {
				"type": "string",
				"description": "Title, abstract or any text to analyze for keywords (legacy parameter, use texts for batch processing)"
			},
			"max_keywords": {
				"type": "integer",
				"default": 10,
				"minimum": 1,
				"maximum": 50,
				"description": "Maximum number of keywords to return for single text (legacy parameter, use texts for batch processing)"
			}
		}
	}`

	tool, err := tools.NewToolImpl(
		"scholarly_suggest_keywords",
		"Extract relevant academic concepts and standardized keywords from text. Useful for finding controlled vocabulary terms to use in academic searches or for suggesting related research areas for a given title or abstract.",
		json.RawMessage(schemaJSON),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create suggest_keywords tool")
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Check if we have batch texts
			if textsVal, ok := arguments["texts"].([]interface{}); ok {
				// Process batch texts
				batchResults := make([]interface{}, 0, len(textsVal))

				for _, textObj := range textsVal {
					textMap, ok := textObj.(map[string]interface{})
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("each item in texts array must be an object"),
						), nil
					}

					// Extract text parameter for this item
					text, ok := textMap["text"].(string)
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("text field must be a string in each item"),
						), nil
					}

					// Get optional max_keywords parameter with default value
					maxKeywords := 10
					if maxVal, ok := textMap["max_keywords"].(float64); ok {
						maxKeywords = int(maxVal)
					}

					// Create request for this text
					req := scholarly.SuggestKeywordsRequest{
						Text:        text,
						MaxKeywords: maxKeywords,
					}

					// Call the keyword suggestion function
					response, err := scholarly.SuggestKeywords(req)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error suggesting keywords for text '%s': %v",
								truncateText(text, 30), err)),
						), nil
					}

					batchResults = append(batchResults, response)
				}

				return protocol.NewToolResult(
					protocol.WithJSON(batchResults),
				), nil
			}

			// Handle legacy single text case
			text, ok := arguments["text"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("text argument must be a string"),
				), nil
			}

			// Get optional max_keywords parameter with default value
			maxKeywords := 10
			if maxVal, ok := arguments["max_keywords"].(float64); ok {
				maxKeywords = int(maxVal)
			}

			// Create request
			req := scholarly.SuggestKeywordsRequest{
				Text:        text,
				MaxKeywords: maxKeywords,
			}

			// Call the keyword suggestion function
			response, err := scholarly.SuggestKeywords(req)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error suggesting keywords: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithJSON(response),
			), nil
		})

	return nil
}

// registerGetMetricsTool registers the get_metrics tool
func registerGetMetricsTool(registry *tool_registry.Registry) error {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"work_ids": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "Array of work identifiers to retrieve metrics for in batch. Can be DOIs (10.xxxx/yyyy), OpenAlex IDs (W12345678), or arXiv IDs (2101.12345)"
			},
			"work_id": {
				"type": "string",
				"description": "Identifier for a single work (legacy parameter, use work_ids for batch processing). Can be a DOI (10.xxxx/yyyy), OpenAlex ID (W12345678), or arXiv ID (2101.12345)"
			}
		}
	}`

	tool, err := tools.NewToolImpl(
		"scholarly_get_metrics",
		"Retrieve impact metrics for a scholarly work, including citation counts, reference counts, and open access status. Helps assess the scholarly impact and visibility of research publications.",
		json.RawMessage(schemaJSON),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create get_metrics tool")
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Check if we have batch work IDs
			if workIDsVal, ok := arguments["work_ids"].([]interface{}); ok {
				// Process batch work IDs
				batchResults := make([]interface{}, 0, len(workIDsVal))

				for _, workIDObj := range workIDsVal {
					workID, ok := workIDObj.(string)
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("each work ID in work_ids array must be a string"),
						), nil
					}

					// Create request for this work ID
					req := scholarly.GetMetricsRequest{
						WorkID: workID,
					}

					// Call the metrics function
					metrics, err := scholarly.GetMetrics(req)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error getting metrics for work %s: %v", workID, err)),
						), nil
					}

					batchResults = append(batchResults, metrics)
				}

				return protocol.NewToolResult(
					protocol.WithJSON(batchResults),
				), nil
			}

			// Handle legacy single work ID case
			workID, ok := arguments["work_id"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("work_id argument must be a string"),
				), nil
			}

			// Create request
			req := scholarly.GetMetricsRequest{
				WorkID: workID,
			}

			// Call the metrics function
			metrics, err := scholarly.GetMetrics(req)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error getting metrics: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithJSON(metrics),
			), nil
		})

	return nil
}

// registerGetCitationsTool registers the get_citations tool
func registerGetCitationsTool(registry *tool_registry.Registry) error {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"work_ids": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"work_id": {
							"type": "string",
							"description": "Identifier for the work. Can be a DOI (10.xxxx/yyyy), OpenAlex ID (W12345678), or arXiv ID (2101.12345)"
						},
						"direction": {
							"type": "string",
							"enum": ["cited_by", "references"],
							"default": "cited_by",
							"description": "Direction of citations for this work"
						},
						"limit": {
							"type": "integer",
							"default": 50,
							"minimum": 1,
							"maximum": 200,
							"description": "Maximum number of citations to return for this work"
						}
					},
					"required": ["work_id"]
				},
				"description": "Array of work identifiers to retrieve citations for in batch"
			},
			"work_id": {
				"type": "string",
				"description": "Identifier for a single work (legacy parameter, use work_ids for batch processing). Can be a DOI (10.xxxx/yyyy), OpenAlex ID (W12345678), or arXiv ID (2101.12345)"
			},
			"direction": {
				"type": "string",
				"enum": ["cited_by", "references"],
				"default": "cited_by",
				"description": "Direction of citations for single work (legacy parameter): 'cited_by' for works that cite this one, 'references' for works this one cites"
			},
			"limit": {
				"type": "integer",
				"default": 50,
				"minimum": 1,
				"maximum": 200,
				"description": "Maximum number of citations to return for single work (legacy parameter)"
			}
		}
	}`

	tool, err := tools.NewToolImpl(
		"scholarly_get_citations",
		"Retrieve citation relationships for a scholarly work. Can fetch either works that cite the specified publication ('cited_by') or works that are cited by the publication ('references'). Useful for literature reviews and understanding research influence.",
		json.RawMessage(schemaJSON),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create get_citations tool")
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Check if we have batch work IDs
			if workIDsVal, ok := arguments["work_ids"].([]interface{}); ok {
				// Process batch work IDs
				batchResults := make([]interface{}, 0, len(workIDsVal))

				for _, workIDObj := range workIDsVal {
					workIDMap, ok := workIDObj.(map[string]interface{})
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("each item in work_ids array must be an object"),
						), nil
					}

					// Extract work ID
					workID, ok := workIDMap["work_id"].(string)
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("work_id field must be a string in each item"),
						), nil
					}

					// Get optional direction parameter with default value
					direction := "cited_by"
					if directionVal, ok := workIDMap["direction"].(string); ok {
						direction = directionVal
					}

					// Get optional limit parameter with default value
					limit := 50
					if limitVal, ok := workIDMap["limit"].(float64); ok {
						limit = int(limitVal)
					}

					// Create request for this work ID
					req := scholarly.GetCitationsRequest{
						WorkID:    workID,
						Direction: direction,
						Limit:     limit,
					}

					// Call the citations function
					response, err := scholarly.GetCitations(req)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error getting citations for work %s: %v", workID, err)),
						), nil
					}

					batchResults = append(batchResults, response)
				}

				return protocol.NewToolResult(
					protocol.WithJSON(batchResults),
				), nil
			}

			// Handle legacy single work ID case
			workID, ok := arguments["work_id"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("work_id argument must be a string"),
				), nil
			}

			// Get optional direction parameter with default value
			direction := "cited_by"
			if directionVal, ok := arguments["direction"].(string); ok {
				direction = directionVal
			}

			// Get optional limit parameter with default value
			limit := 50
			if limitVal, ok := arguments["limit"].(float64); ok {
				limit = int(limitVal)
			}

			// Create request
			req := scholarly.GetCitationsRequest{
				WorkID:    workID,
				Direction: direction,
				Limit:     limit,
			}

			// Call the citations function
			response, err := scholarly.GetCitations(req)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error getting citations: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithJSON(response),
			), nil
		})

	return nil
}

// registerFindFullTextTool registers the find_full_text tool
func registerFindFullTextTool(registry *tool_registry.Registry) error {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"papers": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"doi": {
							"type": "string",
							"description": "Digital Object Identifier for the paper (preferred over title when available)"
						},
						"title": {
							"type": "string", 
							"description": "Title of the paper (used when DOI is not available or not found)"
						},
						"prefer_version": {
							"type": "string",
							"enum": ["published", "preprint", "any"],
							"default": "any",
							"description": "Preferred version for this paper"
						}
					}
				},
				"description": "Array of papers to find full text for in batch"
			},
			"doi": {
				"type": "string",
				"description": "Digital Object Identifier for a single paper (legacy parameter, use papers for batch processing)"
			},
			"title": {
				"type": "string", 
				"description": "Title of a single paper (legacy parameter, use papers for batch processing)"
			},
			"prefer_version": {
				"type": "string",
				"enum": ["published", "preprint", "any"],
				"default": "any",
				"description": "Preferred version for single paper (legacy parameter): 'published' for final version, 'preprint' for author manuscript, 'any' for either"
			}
		}
	}`

	tool, err := tools.NewToolImpl(
		"scholarly_find_full_text",
		"Locate the full text of a scholarly article by searching across various repositories including open access journals, pre-print servers, and academic databases. Returns direct links to PDF or HTML when available.",
		json.RawMessage(schemaJSON),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create find_full_text tool")
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Check if we have batch papers
			if papersVal, ok := arguments["papers"].([]interface{}); ok {
				// Process batch papers
				batchResults := make([]interface{}, 0, len(papersVal))

				for _, paperObj := range papersVal {
					paperMap, ok := paperObj.(map[string]interface{})
					if !ok {
						return protocol.NewToolResult(
							protocol.WithError("each item in papers array must be an object"),
						), nil
					}

					// Extract parameters for this paper
					doi := ""
					if doiVal, ok := paperMap["doi"].(string); ok {
						doi = doiVal
					}

					title := ""
					if titleVal, ok := paperMap["title"].(string); ok {
						title = titleVal
					}

					// Require at least one of DOI or title
					if doi == "" && title == "" {
						return protocol.NewToolResult(
							protocol.WithError("at least one of doi or title must be provided for each paper"),
						), nil
					}

					// Get optional prefer_version parameter with default value
					preferVersion := "any"
					if versionVal, ok := paperMap["prefer_version"].(string); ok {
						preferVersion = versionVal
					}

					// Create request for this paper
					req := scholarly.FindFullTextRequest{
						DOI:           doi,
						Title:         title,
						PreferVersion: preferVersion,
					}

					// Call the full text function
					response, err := scholarly.FindFullText(req)
					if err != nil {
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error finding full text for paper: %v", err)),
						), nil
					}

					batchResults = append(batchResults, response)
				}

				return protocol.NewToolResult(
					protocol.WithJSON(batchResults),
				), nil
			}

			// Handle legacy single paper case
			doi := ""
			if doiVal, ok := arguments["doi"].(string); ok {
				doi = doiVal
			}

			title := ""
			if titleVal, ok := arguments["title"].(string); ok {
				title = titleVal
			}

			// Require at least one of DOI or title
			if doi == "" && title == "" {
				return protocol.NewToolResult(
					protocol.WithError("at least one of doi or title must be provided"),
				), nil
			}

			// Get optional prefer_version parameter with default value
			preferVersion := "any"
			if versionVal, ok := arguments["prefer_version"].(string); ok {
				preferVersion = versionVal
			}

			// Create request
			req := scholarly.FindFullTextRequest{
				DOI:           doi,
				Title:         title,
				PreferVersion: preferVersion,
			}

			// Call the full text function
			response, err := scholarly.FindFullText(req)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error finding full text: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithJSON(response),
			), nil
		})

	return nil
}
