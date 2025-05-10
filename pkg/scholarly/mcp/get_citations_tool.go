package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	tools2 "github.com/go-go-golems/go-go-mcp/pkg/scholarly/tools"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/pkg/errors"
)

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
						},
						"original_query_intent": {
							"type": "string",
							"description": "The human language intent that caused this tool to be called and why"
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
			},
			"original_query_intent": {
				"type": "string",
				"description": "The human language intent that caused this tool to be called and why"
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
					req := common.GetCitationsRequest{
						WorkID:    workID,
						Direction: direction,
						Limit:     limit,
					}

					// Call the citations function
					response, err := tools2.GetCitations(req)
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
			req := common.GetCitationsRequest{
				WorkID:    workID,
				Direction: direction,
				Limit:     limit,
			}

			// Call the citations function
			response, err := tools2.GetCitations(req)
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
