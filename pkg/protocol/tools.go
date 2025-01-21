package protocol

import (
	"encoding/json"
	"fmt"
)

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolResult represents the result of a tool invocation
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError"`
}

// ToolContent represents different types of content in a tool result
type ToolContent struct {
	Type     string           `json:"type"`           // "text", "image", or "resource"
	Text     string           `json:"text,omitempty"` // For text content
	Data     string           `json:"data,omitempty"` // Base64 encoded for image content
	MimeType string           `json:"mimeType,omitempty"`
	Resource *ResourceContent `json:"resource,omitempty"` // For resource content
}

// ToolResultOption is a function that modifies a ToolResult
type ToolResultOption func(*ToolResult)

// NewToolResult creates a new ToolResult with the given options
func NewToolResult(opts ...ToolResultOption) *ToolResult {
	tr := &ToolResult{
		Content: []ToolContent{},
		IsError: false,
	}

	for _, opt := range opts {
		opt(tr)
	}

	return tr
}

// WithText adds a text content to the ToolResult
func WithText(text string) ToolResultOption {
	return func(tr *ToolResult) {
		tr.Content = append(tr.Content, NewTextContent(text))
	}
}

// WithJSON adds JSON-serialized content to the ToolResult
// If marshaling fails, it adds an error message instead
func WithJSON(data interface{}) ToolResultOption {
	return func(tr *ToolResult) {
		content, err := NewJSONContent(data)
		if err != nil {
			tr.Content = append(tr.Content, NewTextContent(fmt.Sprintf("Error marshaling JSON: %v", err)))
			tr.IsError = true
			return
		}
		tr.Content = append(tr.Content, content)
	}
}

// WithImage adds an image content to the ToolResult
func WithImage(base64Data, mimeType string) ToolResultOption {
	return func(tr *ToolResult) {
		tr.Content = append(tr.Content, NewImageContent(base64Data, mimeType))
	}
}

// WithResource adds a resource content to the ToolResult
func WithResource(resource *ResourceContent) ToolResultOption {
	return func(tr *ToolResult) {
		tr.Content = append(tr.Content, NewResourceContent(resource))
	}
}

// WithError marks the ToolResult as an error and optionally adds an error message
func WithError(errorMsg string) ToolResultOption {
	return func(tr *ToolResult) {
		tr.IsError = true
		if errorMsg != "" {
			tr.Content = append(tr.Content, NewTextContent(errorMsg))
		}
	}
}

// WithContent adds raw ToolContent to the ToolResult
func WithContent(content ToolContent) ToolResultOption {
	return func(tr *ToolResult) {
		tr.Content = append(tr.Content, content)
	}
}

// NewErrorToolResult creates a new ToolResult marked as error with the given content
func NewErrorToolResult(content ...ToolContent) *ToolResult {
	return &ToolResult{
		Content: content,
		IsError: true,
	}
}

// NewTextContent creates a new ToolContent with text type
func NewTextContent(text string) ToolContent {
	return ToolContent{
		Type: "text",
		Text: text,
	}
}

// NewJSONContent creates a new ToolContent with JSON-serialized data
func NewJSONContent(data interface{}) (ToolContent, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ToolContent{}, err
	}

	return ToolContent{
		Type:     "text",
		Text:     string(jsonBytes),
		MimeType: "application/json",
	}, nil
}

// MustNewJSONContent creates a new ToolContent with JSON-serialized data
// Panics if marshaling fails
func MustNewJSONContent(data interface{}) ToolContent {
	content, err := NewJSONContent(data)
	if err != nil {
		panic(err)
	}
	return content
}

// NewImageContent creates a new ToolContent with base64-encoded image data
func NewImageContent(base64Data, mimeType string) ToolContent {
	return ToolContent{
		Type:     "image",
		Data:     base64Data,
		MimeType: mimeType,
	}
}

// NewResourceContent creates a new ToolContent with resource data
func NewResourceContent(resource *ResourceContent) ToolContent {
	return ToolContent{
		Type:     "resource",
		Resource: resource,
	}
}
