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
			},
			"original_query_intent": {
				"type": "string",
				"description": "The human language intent that caused this tool to be called and why"
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
					req := common.GetMetricsRequest{
						WorkID: workID,
					}

					// Call the metrics function
					metrics, err := tools2.GetMetrics(req)
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
			req := common.GetMetricsRequest{
				WorkID: workID,
			}

			// Call the metrics function
			metrics, err := tools2.GetMetrics(req)
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
