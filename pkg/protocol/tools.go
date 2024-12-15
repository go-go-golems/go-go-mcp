package protocol

import "encoding/json"

// Tool represents a tool that can be invoked
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"` // JSON Schema
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
