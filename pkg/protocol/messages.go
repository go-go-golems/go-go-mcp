package protocol

import "encoding/json"

// Common message types

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"` // Can be a string or an int
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"` // Can be a string or an int
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Notification represents a JSON-RPC 2.0 notification.
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Initialization types

// InitializeParams represents the parameters for an initialize request
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// InitializeResult represents the response to an initialize request
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ClientInfo contains information about the client implementation
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo contains information about the server implementation
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Capability types

// ClientCapabilities describes the features supported by the client
type ClientCapabilities struct {
	Roots        *RootsCapability       `json:"roots,omitempty"`
	Sampling     *SamplingCapability    `json:"sampling,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// ServerCapabilities describes the features supported by the server
type ServerCapabilities struct {
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Tools        *ToolsCapability       `json:"tools,omitempty"`
	Logging      *LoggingCapability     `json:"logging,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// RootsCapability describes filesystem root capabilities
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability describes LLM sampling capabilities
type SamplingCapability struct{}

// PromptsCapability describes prompt template capabilities
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability describes resource capabilities
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability describes tool capabilities
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability describes logging capabilities
type LoggingCapability struct{}

// Utility message types

// CancellationParams represents parameters for a cancellation notification
type CancellationParams struct {
	RequestID string `json:"requestId"`
	Reason    string `json:"reason,omitempty"`
}

// ProgressParams represents parameters for a progress notification
type ProgressParams struct {
	ProgressToken string  `json:"progressToken"`
	Progress      float64 `json:"progress"`
	Total         float64 `json:"total,omitempty"`
}

// RequestMetadata represents common metadata that can be included in requests
type RequestMetadata struct {
	ProgressToken string `json:"progressToken,omitempty"`
}

// Prompt types

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

// Resource types

// Resource represents a resource exposed by the server
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"` // Base64 encoded
}

// ResourceTemplate represents a parameterized resource template
type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Tool types

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

// Completion types

// CompletionReference represents what is being completed
type CompletionReference struct {
	Type string `json:"type"` // "ref/prompt" or "ref/resource"
	Name string `json:"name,omitempty"`
	URI  string `json:"uri,omitempty"`
}

// CompletionArgument represents the argument being completed
type CompletionArgument struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CompletionResult represents completion suggestions
type CompletionResult struct {
	Values  []string `json:"values"`
	Total   *int     `json:"total,omitempty"`
	HasMore bool     `json:"hasMore"`
}

// Logging types

// LogMessage represents a log message notification
type LogMessage struct {
	Level  string         `json:"level"`
	Logger string         `json:"logger,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
}

// Root types

// Root represents a filesystem root exposed by the client
type Root struct {
	URI  string `json:"uri"` // Must be a file:// URI
	Name string `json:"name,omitempty"`
}

// Sampling types

// ModelPreferences represents preferences for model selection
type ModelPreferences struct {
	Hints                []ModelHint `json:"hints,omitempty"`
	CostPriority         float64     `json:"costPriority,omitempty"`
	SpeedPriority        float64     `json:"speedPriority,omitempty"`
	IntelligencePriority float64     `json:"intelligencePriority,omitempty"`
}

// ModelHint represents a suggested model name
type ModelHint struct {
	Name string `json:"name"`
}

// SamplingMessage represents a message in a sampling conversation
type SamplingMessage struct {
	Role    string         `json:"role"` // "user" or "assistant"
	Content MessageContent `json:"content"`
}

// MessageContent represents different types of content in a sampling message
type MessageContent struct {
	Type     string `json:"type"`           // "text" or "image"
	Text     string `json:"text,omitempty"` // For text content
	Data     string `json:"data,omitempty"` // Base64 encoded for image content
	MimeType string `json:"mimeType,omitempty"`
}

// SamplingCreateRequest represents a request to create a sampling message
type SamplingCreateRequest struct {
	Messages         []SamplingMessage `json:"messages"`
	ModelPreferences ModelPreferences  `json:"modelPreferences,omitempty"`
	SystemPrompt     string            `json:"systemPrompt,omitempty"`
	MaxTokens        int               `json:"maxTokens,omitempty"`
}

// SamplingCreateResponse represents the response to a sampling create request
type SamplingCreateResponse struct {
	Role       string         `json:"role"`
	Content    MessageContent `json:"content"`
	Model      string         `json:"model,omitempty"`
	StopReason string         `json:"stopReason,omitempty"`
}
