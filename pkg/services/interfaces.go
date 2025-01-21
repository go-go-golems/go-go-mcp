package services

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// PromptService handles prompt-related operations
type PromptService interface {
	ListPrompts(ctx context.Context, cursor string) ([]protocol.Prompt, string, error)
	GetPrompt(ctx context.Context, name string, arguments map[string]string) (*protocol.PromptMessage, error)
}

// ResourceService handles resource-related operations
type ResourceService interface {
	ListResources(ctx context.Context, cursor string) ([]protocol.Resource, string, error)
	ReadResource(ctx context.Context, uri string) (*protocol.ResourceContent, error)
}

// ToolService handles tool-related operations
type ToolService interface {
	ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error)
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error)
}

// InitializeService handles initialization operations
type InitializeService interface {
	Initialize(ctx context.Context, params protocol.InitializeParams) (protocol.InitializeResult, error)
}
