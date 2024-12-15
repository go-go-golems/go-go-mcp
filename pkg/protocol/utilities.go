package protocol

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

// LogMessage represents a log message notification
type LogMessage struct {
	Level  string         `json:"level"`
	Logger string         `json:"logger,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
}

// Root represents a filesystem root exposed by the client
type Root struct {
	URI  string `json:"uri"` // Must be a file:// URI
	Name string `json:"name,omitempty"`
}
