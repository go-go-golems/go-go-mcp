package defaults

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
)

type DefaultToolService struct {
	registry *pkg.ProviderRegistry
	logger   zerolog.Logger
}

func NewToolService(registry *pkg.ProviderRegistry, logger zerolog.Logger) *DefaultToolService {
	return &DefaultToolService{
		registry: registry,
		logger:   logger,
	}
}

func (s *DefaultToolService) ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	var allTools []protocol.Tool
	var lastCursor string

	for _, provider := range s.registry.GetToolProviders() {
		tools, nextCursor, err := provider.ListTools(cursor)
		if err != nil {
			s.logger.Error().Err(err).Msg("Error listing tools from provider")
			continue
		}
		allTools = append(allTools, tools...)
		if nextCursor != "" {
			lastCursor = nextCursor
		}
	}

	return allTools, lastCursor, nil
}

func (s *DefaultToolService) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error) {
	for _, provider := range s.registry.GetToolProviders() {
		result, err := provider.CallTool(ctx, name, arguments)
		if err == nil {
			return result, nil
		}
	}
	return nil, pkg.ErrToolNotFound
}
