package client

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-go-golems/go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
)

// Transport represents a client transport mechanism
type Transport interface {
	// Send sends a request and returns the response
	Send(request *protocol.Request) (*protocol.Response, error)
	// Close closes the transport connection
	Close() error
}

// Client represents an MCP client that can use different transports
type Client struct {
	mu        sync.Mutex
	logger    zerolog.Logger
	transport Transport
	nextID    int

	// Client capabilities declared during initialization
	capabilities protocol.ClientCapabilities
	// Server capabilities received during initialization
	serverCapabilities protocol.ServerCapabilities
	initialized        bool
}

// NewClient creates a new client instance
func NewClient(logger zerolog.Logger, transport Transport) *Client {
	return &Client{
		logger:    logger,
		transport: transport,
		nextID:    1,
	}
}

// Initialize initializes the connection with the server
func (c *Client) Initialize(capabilities protocol.ClientCapabilities) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return fmt.Errorf("client already initialized")
	}

	params := protocol.InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    capabilities,
		ClientInfo: protocol.ClientInfo{
			Name:    "go-mcp-client",
			Version: "dev",
		},
	}

	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params:  mustMarshal(params),
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	var result protocol.InitializeResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}

	c.capabilities = capabilities
	c.serverCapabilities = result.Capabilities
	c.initialized = true

	// Send initialized notification
	notification := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	_, err = c.transport.Send(notification)
	if err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	return nil
}

// ListPrompts retrieves the list of available prompts from the server
func (c *Client) ListPrompts(cursor string) ([]protocol.Prompt, string, error) {
	if !c.initialized {
		return nil, "", fmt.Errorf("client not initialized")
	}

	params := map[string]string{}
	if cursor != "" {
		params["cursor"] = cursor
	}

	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "prompts/list",
		Params:  mustMarshal(params),
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send prompts/list request: %w", err)
	}

	if response.Error != nil {
		return nil, "", fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	var result struct {
		Prompts    []protocol.Prompt `json:"prompts"`
		NextCursor string            `json:"nextCursor"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal prompts/list result: %w", err)
	}

	return result.Prompts, result.NextCursor, nil
}

// GetPrompt retrieves a specific prompt from the server
func (c *Client) GetPrompt(name string, arguments map[string]string) (*protocol.PromptMessage, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}{
		Name:      name,
		Arguments: arguments,
	}

	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "prompts/get",
		Params:  mustMarshal(params),
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send prompts/get request: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	var result struct {
		Messages []protocol.PromptMessage `json:"messages"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prompts/get result: %w", err)
	}

	if len(result.Messages) == 0 {
		return nil, fmt.Errorf("server returned no messages")
	}

	return &result.Messages[0], nil
}

// ListResources retrieves the list of available resources from the server
func (c *Client) ListResources(cursor string) ([]protocol.Resource, string, error) {
	if !c.initialized {
		return nil, "", fmt.Errorf("client not initialized")
	}

	params := map[string]string{}
	if cursor != "" {
		params["cursor"] = cursor
	}

	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "resources/list",
		Params:  mustMarshal(params),
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send resources/list request: %w", err)
	}

	if response.Error != nil {
		return nil, "", fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	var result struct {
		Resources  []protocol.Resource `json:"resources"`
		NextCursor string              `json:"nextCursor"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal resources/list result: %w", err)
	}

	return result.Resources, result.NextCursor, nil
}

// ReadResource retrieves the content of a specific resource from the server
func (c *Client) ReadResource(uri string) (*protocol.ResourceContent, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := struct {
		URI string `json:"uri"`
	}{
		URI: uri,
	}

	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "resources/read",
		Params:  mustMarshal(params),
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send resources/read request: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	var result struct {
		Contents []protocol.ResourceContent `json:"contents"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resources/read result: %w", err)
	}

	if len(result.Contents) == 0 {
		return nil, fmt.Errorf("server returned no content")
	}

	return &result.Contents[0], nil
}

// ListTools retrieves the list of available tools from the server
func (c *Client) ListTools(cursor string) ([]protocol.Tool, string, error) {
	if !c.initialized {
		return nil, "", fmt.Errorf("client not initialized")
	}

	params := map[string]string{}
	if cursor != "" {
		params["cursor"] = cursor
	}

	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "tools/list",
		Params:  mustMarshal(params),
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send tools/list request: %w", err)
	}

	if response.Error != nil {
		return nil, "", fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	var result struct {
		Tools      []protocol.Tool `json:"tools"`
		NextCursor string          `json:"nextCursor"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal tools/list result: %w", err)
	}

	return result.Tools, result.NextCursor, nil
}

// CallTool calls a specific tool on the server
func (c *Client) CallTool(name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}{
		Name:      name,
		Arguments: arguments,
	}

	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  mustMarshal(params),
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send tools/call request: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	var result protocol.ToolResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools/call result: %w", err)
	}

	return &result, nil
}

// CreateMessage sends a sampling request to create a message
func (c *Client) CreateMessage(messages []protocol.Message, modelPreferences protocol.ModelPreferences, systemPrompt string, maxTokens int) (*protocol.Message, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := struct {
		Messages         []protocol.Message        `json:"messages"`
		ModelPreferences protocol.ModelPreferences `json:"modelPreferences"`
		SystemPrompt     string                    `json:"systemPrompt,omitempty"`
		MaxTokens        int                       `json:"maxTokens,omitempty"`
	}{
		Messages:         messages,
		ModelPreferences: modelPreferences,
		SystemPrompt:     systemPrompt,
		MaxTokens:        maxTokens,
	}

	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "sampling/createMessage",
		Params:  mustMarshal(params),
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send sampling/createMessage request: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	var result protocol.Message
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sampling/createMessage result: %w", err)
	}

	return &result, nil
}

// Ping sends a ping request to the server
func (c *Client) Ping() error {
	request := &protocol.Request{
		JSONRPC: "2.0",
		Method:  "ping",
	}
	c.setRequestID(request)

	response, err := c.transport.Send(request)
	if err != nil {
		return fmt.Errorf("failed to send ping request: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("server returned error: %s", response.Error.Message)
	}

	return nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.transport.Close()
}

// setRequestID sets a unique ID for the request
func (c *Client) setRequestID(request *protocol.Request) {
	if request.Method == "notifications/initialized" {
		return // notifications don't have IDs
	}

	id := json.RawMessage(fmt.Sprintf("%d", c.nextID))
	request.ID = id
	c.nextID++
}

// mustMarshal marshals data to JSON or panics
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return data
}
