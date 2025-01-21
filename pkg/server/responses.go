package server

import "github.com/go-go-golems/go-go-mcp/pkg/protocol"

// ListPromptsResult is the response type for prompts/list method
type ListPromptsResult struct {
	Prompts    []protocol.Prompt `json:"prompts"`
	NextCursor string            `json:"nextCursor"`
}

// ListResourcesResult is the response type for resources/list method
type ListResourcesResult struct {
	Resources  []protocol.Resource `json:"resources"`
	NextCursor string              `json:"nextCursor"`
}

// ListToolsResult is the response type for tools/list method
type ListToolsResult struct {
	Tools      []protocol.Tool `json:"tools"`
	NextCursor string          `json:"nextCursor"`
}
