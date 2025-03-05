package pkg

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// PromptProvider defines the interface for serving prompts
type PromptProvider interface {
	// ListPrompts returns a list of available prompts with optional pagination
	ListPrompts(ctx context.Context, cursor string) ([]protocol.Prompt, string, error)

	// GetPrompt retrieves a specific prompt with the given arguments
	GetPrompt(ctx context.Context, name string, arguments map[string]string) (*protocol.PromptMessage, error)
}

// ResourceProvider defines the interface for serving resources
type ResourceProvider interface {
	// ListResources returns a list of available resources with optional pagination
	ListResources(ctx context.Context, cursor string) ([]protocol.Resource, string, error)

	// ReadResource retrieves the contents of a specific resource
	ReadResource(ctx context.Context, uri string) ([]protocol.ResourceContent, error)

	// ListResourceTemplates returns a list of available resource templates
	ListResourceTemplates(ctx context.Context) ([]protocol.ResourceTemplate, error)

	// SubscribeToResource registers for notifications about resource changes
	// Returns a channel that will receive notifications and a cleanup function
	SubscribeToResource(ctx context.Context, uri string) (chan struct{}, func(), error)
}

// ToolProvider defines the interface for serving tools
type ToolProvider interface {
	// ListTools returns a list of available tools with optional pagination
	ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error)

	// CallTool invokes a specific tool with the given arguments
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error)
}
