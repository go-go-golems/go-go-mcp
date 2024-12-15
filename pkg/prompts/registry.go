package prompts

import (
	"fmt"
	"sort"
	"sync"

	"github.com/go-go-golems/go-mcp/pkg"
	"github.com/go-go-golems/go-mcp/pkg/protocol"
)

// Registry provides a simple way to register individual prompts
type Registry struct {
	mu      sync.RWMutex
	prompts map[string]protocol.Prompt
	// handlers map prompt names to custom handlers for generating messages
	handlers map[string]Handler
}

// Handler is a function that generates a prompt message based on arguments
type Handler func(prompt protocol.Prompt, arguments map[string]string) (*protocol.PromptMessage, error)

// NewRegistry creates a new prompt registry
func NewRegistry() *Registry {
	return &Registry{
		prompts:  make(map[string]protocol.Prompt),
		handlers: make(map[string]Handler),
	}
}

// RegisterPrompt adds a prompt to the registry
func (r *Registry) RegisterPrompt(prompt protocol.Prompt) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prompts[prompt.Name] = prompt
}

// RegisterPromptWithHandler adds a prompt with a custom handler to the registry
func (r *Registry) RegisterPromptWithHandler(prompt protocol.Prompt, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prompts[prompt.Name] = prompt
	r.handlers[prompt.Name] = handler
}

// UnregisterPrompt removes a prompt from the registry
func (r *Registry) UnregisterPrompt(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.prompts, name)
	delete(r.handlers, name)
}

// ListPrompts implements PromptProvider interface
func (r *Registry) ListPrompts(cursor string) ([]protocol.Prompt, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get all prompts
	prompts := make([]protocol.Prompt, 0, len(r.prompts))
	for _, p := range r.prompts {
		prompts = append(prompts, p)
	}

	// Sort prompts by name for consistent ordering
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})

	// If cursor is empty, return all prompts
	if cursor == "" {
		return prompts, "", nil
	}

	// Find the cursor position
	pos := -1
	for i, p := range prompts {
		if p.Name == cursor {
			pos = i
			break
		}
	}

	// If cursor not found, return all prompts
	if pos == -1 {
		return prompts, "", nil
	}

	// Return prompts after cursor
	return prompts[pos+1:], "", nil
}

// GetPrompt implements PromptProvider interface
func (r *Registry) GetPrompt(name string, arguments map[string]string) (*protocol.PromptMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prompt, ok := r.prompts[name]
	if !ok {
		return nil, pkg.ErrPromptNotFound
	}

	// Validate required arguments
	for _, arg := range prompt.Arguments {
		if arg.Required {
			if _, ok := arguments[arg.Name]; !ok {
				return nil, fmt.Errorf("missing required argument: %s", arg.Name)
			}
		}
	}

	// Use custom handler if available
	if handler, ok := r.handlers[name]; ok {
		return handler(prompt, arguments)
	}

	// Build message content from arguments
	content := "Please help me with the following task:\n\n"
	for key, value := range arguments {
		content += fmt.Sprintf("%s: %s\n", key, value)
	}

	return &protocol.PromptMessage{
		Role: "user",
		Content: protocol.PromptContent{
			Type: "text",
			Text: content,
		},
	}, nil
}
