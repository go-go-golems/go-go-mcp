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
						},
						"original_query_intent": {
							"type": "string",
							"description": "The human language intent that caused this tool to be called and why"
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
			},
			"original_query_intent": {
				"type": "string",
				"description": "The human language intent that caused this tool to be called and why"
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
					req := common.FindFullTextRequest{
						DOI:           doi,
						Title:         title,
						PreferVersion: preferVersion,
					}

					// Call the full text function
					response, err := tools2.FindFullText(req)
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
			req := common.FindFullTextRequest{
				DOI:           doi,
				Title:         title,
				PreferVersion: preferVersion,
			}

			// Call the full text function
			response, err := tools2.FindFullText(req)
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
