package tools

import (
	"sort"
	"sync"

	"github.com/go-go-golems/go-mcp/pkg"
	"github.com/go-go-golems/go-mcp/pkg/protocol"
)

// Registry provides a simple way to register individual tools
type Registry struct {
	mu       sync.RWMutex
	tools    map[string]protocol.Tool
	handlers map[string]Handler
}

// Handler is a function that executes a tool with given arguments
type Handler func(tool protocol.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error)

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools:    make(map[string]protocol.Tool),
		handlers: make(map[string]Handler),
	}
}

// RegisterTool adds a tool to the registry
func (r *Registry) RegisterTool(tool protocol.Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

// RegisterToolWithHandler adds a tool with a custom handler
func (r *Registry) RegisterToolWithHandler(tool protocol.Tool, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
	r.handlers[tool.Name] = handler
}

// UnregisterTool removes a tool from the registry
func (r *Registry) UnregisterTool(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
	delete(r.handlers, name)
}

// ListTools implements ToolProvider interface
func (r *Registry) ListTools(cursor string) ([]protocol.Tool, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]protocol.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}

	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})

	if cursor == "" {
		return tools, "", nil
	}

	pos := -1
	for i, t := range tools {
		if t.Name == cursor {
			pos = i
			break
		}
	}

	if pos == -1 {
		return tools, "", nil
	}

	return tools[pos+1:], "", nil
}

// CallTool implements ToolProvider interface
func (r *Registry) CallTool(name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	if !ok {
		return nil, pkg.ErrToolNotFound
	}

	if handler, ok := r.handlers[name]; ok {
		return handler(tool, arguments)
	}

	// Return empty result if no handler is registered
	return &protocol.ToolResult{}, nil
}
