package defaults

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
)

type DefaultResourceService struct {
	registry *pkg.ProviderRegistry
	logger   zerolog.Logger
}

func NewResourceService(registry *pkg.ProviderRegistry, logger zerolog.Logger) *DefaultResourceService {
	return &DefaultResourceService{
		registry: registry,
		logger:   logger,
	}
}

func (s *DefaultResourceService) ListResources(ctx context.Context, cursor string) ([]protocol.Resource, string, error) {
	var allResources []protocol.Resource
	var lastCursor string

	for _, provider := range s.registry.GetResourceProviders() {
		resources, nextCursor, err := provider.ListResources(cursor)
		if err != nil {
			s.logger.Error().Err(err).Msg("Error listing resources from provider")
			continue
		}
		allResources = append(allResources, resources...)
		if nextCursor != "" {
			lastCursor = nextCursor
		}
	}

	return allResources, lastCursor, nil
}

func (s *DefaultResourceService) ReadResource(ctx context.Context, uri string) (*protocol.ResourceContent, error) {
	for _, provider := range s.registry.GetResourceProviders() {
		content, err := provider.ReadResource(uri)
		if err == nil {
			return content, nil
		}
	}
	return nil, pkg.ErrResourceNotFound
}
