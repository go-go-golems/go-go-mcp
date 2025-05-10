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

// truncateText helps create more readable error messages by truncating long text
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
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
						},
						"original_query_intent": {
							"type": "string",
							"description": "The human language intent that caused this tool to be called and why"
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
			},
			"original_query_intent": {
				"type": "string",
				"description": "The human language intent that caused this tool to be called and why"
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
					req := common.SuggestKeywordsRequest{
						Text:        text,
						MaxKeywords: maxKeywords,
					}

					// Call the keyword suggestion function
					response, err := tools2.SuggestKeywords(req)
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
			req := common.SuggestKeywordsRequest{
				Text:        text,
				MaxKeywords: maxKeywords,
			}

			// Call the keyword suggestion function
			response, err := tools2.SuggestKeywords(req)
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
