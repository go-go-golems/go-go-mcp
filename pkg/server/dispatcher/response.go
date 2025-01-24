package dispatcher

import (
	"encoding/json"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// NewSuccessResponse creates a new success response with the given result
func NewSuccessResponse(id json.RawMessage, result interface{}) (*protocol.Response, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return NewErrorResponse(id, -32603, "Internal error", err)
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}, nil
}

// NewErrorResponse creates a new error response
func NewErrorResponse(id json.RawMessage, code int, message string, data interface{}) (*protocol.Response, error) {
	var errorData json.RawMessage
	if data != nil {
		if jsonData, err := json.Marshal(data); err == nil {
			errorData = jsonData
		}
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &protocol.Error{
			Code:    code,
			Message: message,
			Data:    errorData,
		},
	}, nil
}

// Common result types
type ListPromptsResult struct {
	Prompts    []protocol.Prompt `json:"prompts"`
	NextCursor string            `json:"nextCursor"`
}

type ListResourcesResult struct {
	Resources  []protocol.Resource `json:"resources"`
	NextCursor string              `json:"nextCursor"`
}

type ListToolsResult struct {
	Tools      []protocol.Tool `json:"tools"`
	NextCursor string          `json:"nextCursor"`
}

type PromptResult struct {
	Description string                   `json:"description"`
	Messages    []protocol.PromptMessage `json:"messages"`
}

type ResourceResult struct {
	Contents []protocol.ResourceContent `json:"contents"`
}
