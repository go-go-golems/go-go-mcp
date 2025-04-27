package tools

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// CombinedProvider implements the ToolProvider interface by combining multiple providers
type CombinedProvider struct {
	providers []pkg.ToolProvider
}

// ListTools implements the ToolProvider interface by combining results from all providers
func (cp *CombinedProvider) ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	var tools []protocol.Tool
	for _, provider := range cp.providers {
		providerTools, _, err := provider.ListTools(ctx, "")
		if err != nil {
			return nil, "", err
		}
		tools = append(tools, providerTools...)
	}
	return tools, "", nil
}

// CallTool implements the ToolProvider interface by trying each provider in order
func (cp *CombinedProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	for _, provider := range cp.providers {
		result, err := provider.CallTool(ctx, name, arguments)
		if err == nil {
			return result, nil
		}
		// If tool not found, try the next provider
		if err == pkg.ErrToolNotFound {
			continue
		}
		// For other errors, return them
		return nil, err
	}
	// If no provider can handle the tool, return not found
	return nil, pkg.ErrToolNotFound
}

// CombineProviders creates a new provider that combines multiple providers
func CombineProviders(providers ...pkg.ToolProvider) pkg.ToolProvider {
	return &CombinedProvider{
		providers: providers,
	}
}
