package defaults

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
)

type DefaultPromptService struct {
	registry *pkg.ProviderRegistry
	logger   zerolog.Logger
}

func NewPromptService(registry *pkg.ProviderRegistry, logger zerolog.Logger) *DefaultPromptService {
	return &DefaultPromptService{
		registry: registry,
		logger:   logger,
	}
}

func (s *DefaultPromptService) ListPrompts(ctx context.Context, cursor string) ([]protocol.Prompt, string, error) {
	var allPrompts []protocol.Prompt
	var lastCursor string

	for _, provider := range s.registry.GetPromptProviders() {
		prompts, nextCursor, err := provider.ListPrompts(cursor)
		if err != nil {
			s.logger.Error().Err(err).Msg("Error listing prompts from provider")
			continue
		}
		allPrompts = append(allPrompts, prompts...)
		if nextCursor != "" {
			lastCursor = nextCursor
		}
	}

	return allPrompts, lastCursor, nil
}

func (s *DefaultPromptService) GetPrompt(ctx context.Context, name string, arguments map[string]string) (*protocol.PromptMessage, error) {
	for _, provider := range s.registry.GetPromptProviders() {
		message, err := provider.GetPrompt(name, arguments)
		if err == nil {
			return message, nil
		}
	}
	return nil, pkg.ErrPromptNotFound
}
