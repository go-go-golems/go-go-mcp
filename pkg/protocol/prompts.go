package protocol

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// Prompt represents a prompt template
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptMessage represents a message in a prompt
type PromptMessage struct {
	Role    string        `json:"role"` // "user" or "assistant"
	Content PromptContent `json:"content"`
}

// PromptContent represents different types of content in a prompt message
type PromptContent struct {
	Type     string           `json:"type"`           // "text", "image", or "resource"
	Text     string           `json:"text,omitempty"` // For text content
	Data     string           `json:"data,omitempty"` // Base64 encoded for image content
	MimeType string           `json:"mimeType,omitempty"`
	Resource *ResourceContent `json:"resource,omitempty"` // For resource content
}
