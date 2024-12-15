package pkg

// PromptProvider defines the interface for serving prompts
type PromptProvider interface {
	// ListPrompts returns a list of available prompts with optional pagination
	ListPrompts(cursor string) ([]Prompt, string, error)

	// GetPrompt retrieves a specific prompt with the given arguments
	GetPrompt(name string, arguments map[string]string) (*PromptMessage, error)
}

// ResourceProvider defines the interface for serving resources
type ResourceProvider interface {
	// ListResources returns a list of available resources with optional pagination
	ListResources(cursor string) ([]Resource, string, error)

	// ReadResource retrieves the contents of a specific resource
	ReadResource(uri string) (*ResourceContent, error)

	// ListResourceTemplates returns a list of available resource templates
	ListResourceTemplates() ([]ResourceTemplate, error)

	// SubscribeToResource registers for notifications about resource changes
	// Returns a channel that will receive notifications and a cleanup function
	SubscribeToResource(uri string) (chan struct{}, func(), error)
}

// ToolProvider defines the interface for serving tools
type ToolProvider interface {
	// ListTools returns a list of available tools with optional pagination
	ListTools(cursor string) ([]Tool, string, error)

	// CallTool invokes a specific tool with the given arguments
	CallTool(name string, arguments map[string]interface{}) (*ToolResult, error)
}

// Provider combines all provider interfaces
type Provider interface {
	PromptProvider
	ResourceProvider
	ToolProvider
}

// ProviderRegistry manages a list of providers
type ProviderRegistry struct {
	promptProviders   []PromptProvider
	resourceProviders []ResourceProvider
	toolProviders     []ToolProvider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		promptProviders:   make([]PromptProvider, 0),
		resourceProviders: make([]ResourceProvider, 0),
		toolProviders:     make([]ToolProvider, 0),
	}
}

// RegisterPromptProvider adds a prompt provider to the registry
func (r *ProviderRegistry) RegisterPromptProvider(provider PromptProvider) {
	r.promptProviders = append(r.promptProviders, provider)
}

// RegisterResourceProvider adds a resource provider to the registry
func (r *ProviderRegistry) RegisterResourceProvider(provider ResourceProvider) {
	r.resourceProviders = append(r.resourceProviders, provider)
}

// RegisterToolProvider adds a tool provider to the registry
func (r *ProviderRegistry) RegisterToolProvider(provider ToolProvider) {
	r.toolProviders = append(r.toolProviders, provider)
}

// RegisterProvider adds a combined provider to the registry
func (r *ProviderRegistry) RegisterProvider(provider Provider) {
	r.RegisterPromptProvider(provider)
	r.RegisterResourceProvider(provider)
	r.RegisterToolProvider(provider)
}

// GetPromptProviders returns all registered prompt providers
func (r *ProviderRegistry) GetPromptProviders() []PromptProvider {
	return r.promptProviders
}

// GetResourceProviders returns all registered resource providers
func (r *ProviderRegistry) GetResourceProviders() []ResourceProvider {
	return r.resourceProviders
}

// GetToolProviders returns all registered tool providers
func (r *ProviderRegistry) GetToolProviders() []ToolProvider {
	return r.toolProviders
}
