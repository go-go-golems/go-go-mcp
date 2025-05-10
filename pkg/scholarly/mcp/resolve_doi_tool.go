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
			},
			"original_query_intent": {
				"type": "string",
				"description": "The human language intent that caused this tool to be called and why"
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
					req := common.ResolveDOIRequest{
						DOI: doi,
					}

					// Call the scholarly DOI resolution function
					work, err := tools2.ResolveDOI(req)
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
			req := common.ResolveDOIRequest{
				DOI: doi,
			}

			// Call the scholarly DOI resolution function
			work, err := tools2.ResolveDOI(req)
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
