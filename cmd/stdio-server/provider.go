package main

import (
	"fmt"

	"github.com/go-go-golems/go-mcp/pkg"
)

// SimpleProvider implements a basic provider with a single prompt
type SimpleProvider struct{}

// NewSimpleProvider creates a new simple provider
func NewSimpleProvider() *SimpleProvider {
	return &SimpleProvider{}
}

// ListPrompts returns a list of available prompts
func (p *SimpleProvider) ListPrompts(cursor string) ([]pkg.Prompt, string, error) {
	return []pkg.Prompt{
		{
			Name: "simple",
			Description: "A simple prompt that can take optional context and topic " +
				"arguments",
			Arguments: []pkg.PromptArgument{
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
		},
	}, "", nil
}

// GetPrompt retrieves a specific prompt with the given arguments
func (p *SimpleProvider) GetPrompt(name string, arguments map[string]string) (*pkg.PromptMessage, error) {
	if name != "simple" {
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}

	// Create messages based on arguments
	messages := []pkg.PromptMessage{}

	// Add context if provided
	if context, ok := arguments["context"]; ok {
		messages = append(messages, pkg.PromptMessage{
			Role: "user",
			Content: pkg.PromptContent{
				Type: "text",
				Text: fmt.Sprintf("Here is some relevant context: %s", context),
			},
		})
	}

	// Add main prompt
	prompt := "Please help me with "
	if topic, ok := arguments["topic"]; ok {
		prompt += fmt.Sprintf("the following topic: %s", topic)
	} else {
		prompt += "whatever questions I may have."
	}

	return &pkg.PromptMessage{
		Role: "user",
		Content: pkg.PromptContent{
			Type: "text",
			Text: prompt,
		},
	}, nil
}
