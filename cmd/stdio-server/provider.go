package main

import (
	"github.com/go-go-golems/go-mcp/pkg/prompts"
	"github.com/go-go-golems/go-mcp/pkg/protocol"
)

// SimpleProvider implements a basic provider with a single prompt
type SimpleProvider struct {
	registry *prompts.Registry
}

// NewSimpleProvider creates a new simple provider
func NewSimpleProvider() *SimpleProvider {
	registry := prompts.NewRegistry()

	// Register our simple prompt
	registry.RegisterPrompt(protocol.Prompt{
		Name:        "simple",
		Description: "A simple prompt that can take optional context and topic arguments",
		Arguments: []protocol.PromptArgument{
			{
				Name:        "context",
				Description: "Additional context to consider",
				Required:    false,
			},
			{
				Name:        "topic",
				Description: "Specific topic to focus on",
				Required:    false,
			},
		},
	})

	return &SimpleProvider{
		registry: registry,
	}
}

// ListPrompts returns a list of available prompts
func (p *SimpleProvider) ListPrompts(cursor string) ([]protocol.Prompt, string, error) {
	return p.registry.ListPrompts(cursor)
}

// GetPrompt retrieves a specific prompt with the given arguments
func (p *SimpleProvider) GetPrompt(name string, arguments map[string]string) (*protocol.PromptMessage, error) {
	return p.registry.GetPrompt(name, arguments)
}
