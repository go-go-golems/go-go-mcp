package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/go-go-golems/go-mcp/pkg"
	"github.com/go-go-golems/go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
)

// Server handles MCP protocol communication
type Server struct {
	mu       sync.Mutex
	scanner  *bufio.Scanner
	writer   *json.Encoder
	logger   zerolog.Logger
	registry *pkg.ProviderRegistry
}

// NewServer creates a new stdio server
func NewServer() *Server {
	// Configure scanner for line-based input
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	// Configure JSON encoder for stdout
	writer := json.NewEncoder(os.Stdout)

	// Set the global log level to debug
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Use ConsoleWriter for colored output
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	return &Server{
		scanner:  scanner,
		writer:   writer,
		logger:   logger,
		registry: pkg.NewProviderRegistry(),
	}
}

// Start begins listening for and handling messages
func (s *Server) Start() error {
	s.logger.Info().Msg("Server starting...")

	// Process messages until stdin is closed
	for s.scanner.Scan() {
		line := s.scanner.Text()
		s.logger.Debug().Str("line", line).Msg("Received line")
		if err := s.handleMessage(line); err != nil {
			s.logger.Error().Err(err).Msg("Error handling message")
			// Continue processing messages even if one fails
		}
	}

	if err := s.scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return io.EOF
}

// handleMessage processes a single message
func (s *Server) handleMessage(message string) error {
	s.logger.Debug().Str("message", message).Msg("Received message")

	// Parse the base message structure
	var request protocol.Request
	if err := json.Unmarshal([]byte(message), &request); err != nil {
		return s.sendError(nil, -32700, "Parse error", err)
	}

	// Validate JSON-RPC version
	if request.JSONRPC != "2.0" {
		return s.sendError(&request.ID, -32600, "Invalid Request", fmt.Errorf("invalid JSON-RPC version"))
	}

	// Handle requests vs notifications based on ID presence
	if len(request.ID) > 0 {
		return s.handleRequest(request)
	}
	return s.handleNotification(request)
}

// handleRequest processes a request message
func (s *Server) handleRequest(request protocol.Request) error {
	switch request.Method {
	case "initialize":
		var params protocol.InitializeParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}
		s.logger.Info().Interface("params", params).Msg("Handling initialize request")
		return s.handleInitialize(&request.ID, &params)

	case "ping":
		return s.sendResult(&request.ID, struct{}{})

	case "prompts/list":
		var cursor string
		if request.Params != nil {
			var params struct {
				Cursor string `json:"cursor"`
			}
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return s.sendError(&request.ID, -32602, "Invalid params", err)
			}
			cursor = params.Cursor
		}

		var allPrompts []protocol.Prompt
		var lastCursor string
		for _, provider := range s.registry.GetPromptProviders() {
			prompts, nextCursor, err := provider.ListPrompts(cursor)
			if err != nil {
				return s.sendError(&request.ID, -32603, "Internal error", err)
			}
			allPrompts = append(allPrompts, prompts...)
			if nextCursor != "" {
				lastCursor = nextCursor
			}
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"prompts":    allPrompts,
			"nextCursor": lastCursor,
		})

	case "prompts/get":
		var params struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}

		for _, provider := range s.registry.GetPromptProviders() {
			message, err := provider.GetPrompt(params.Name, params.Arguments)
			if err == nil {
				return s.sendResult(&request.ID, map[string]interface{}{
					"description": "Prompt from provider",
					"messages":    []protocol.PromptMessage{*message},
				})
			}
		}
		return s.sendError(&request.ID, -32602, "Prompt not found", nil)

	case "resources/list":
		var cursor string
		if request.Params != nil {
			var params struct {
				Cursor string `json:"cursor"`
			}
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return s.sendError(&request.ID, -32602, "Invalid params", err)
			}
			cursor = params.Cursor
		}

		var allResources []protocol.Resource
		var lastCursor string
		for _, provider := range s.registry.GetResourceProviders() {
			resources, nextCursor, err := provider.ListResources(cursor)
			if err != nil {
				return s.sendError(&request.ID, -32603, "Internal error", err)
			}
			allResources = append(allResources, resources...)
			if nextCursor != "" {
				lastCursor = nextCursor
			}
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"resources":  allResources,
			"nextCursor": lastCursor,
		})

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}

		for _, provider := range s.registry.GetResourceProviders() {
			content, err := provider.ReadResource(params.URI)
			if err == nil {
				return s.sendResult(&request.ID, map[string]interface{}{
					"contents": []protocol.ResourceContent{*content},
				})
			}
		}
		return s.sendError(&request.ID, -32002, "Resource not found", nil)

	case "tools/list":
		var cursor string
		if request.Params != nil {
			var params struct {
				Cursor string `json:"cursor"`
			}
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return s.sendError(&request.ID, -32602, "Invalid params", err)
			}
			cursor = params.Cursor
		}

		var allTools []protocol.Tool
		var lastCursor string
		for _, provider := range s.registry.GetToolProviders() {
			tools, nextCursor, err := provider.ListTools(cursor)
			if err != nil {
				return s.sendError(&request.ID, -32603, "Internal error", err)
			}
			allTools = append(allTools, tools...)
			if nextCursor != "" {
				lastCursor = nextCursor
			}
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"tools":      allTools,
			"nextCursor": lastCursor,
		})

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}

		for _, provider := range s.registry.GetToolProviders() {
			result, err := provider.CallTool(params.Name, params.Arguments)
			if err == nil {
				return s.sendResult(&request.ID, result)
			}
		}
		return s.sendError(&request.ID, -32602, "Tool not found", nil)

	default:
		return s.sendError(&request.ID, -32601, "Method not found", nil)
	}
}

// handleNotification processes a notification message
func (s *Server) handleNotification(request protocol.Request) error {
	var notification protocol.Notification
	notification.JSONRPC = request.JSONRPC
	notification.Method = request.Method
	notification.Params = request.Params

	switch notification.Method {
	case "notifications/initialized":
		s.logger.Info().Msg("Client initialized")
		return nil

	default:
		s.logger.Warn().Str("method", notification.Method).Msg("Unknown notification method")
		return nil
	}
}

// handleInitialize processes an initialize request
func (s *Server) handleInitialize(id *json.RawMessage, params *protocol.InitializeParams) error {
	// Validate protocol version
	supportedVersions := []string{"2024-11-05"}
	isSupported := false
	for _, version := range supportedVersions {
		if params.ProtocolVersion == version {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return s.sendError(id, -32602, "Unsupported protocol version", &struct {
			Supported []string `json:"supported"`
			Requested string   `json:"requested"`
		}{
			Supported: supportedVersions,
			Requested: params.ProtocolVersion,
		})
	}

	// Accept the protocol version and declare capabilities
	result := protocol.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: protocol.ServerCapabilities{
			Logging: &protocol.LoggingCapability{},
			Prompts: &protocol.PromptsCapability{
				ListChanged: true,
			},
			Resources: &protocol.ResourcesCapability{
				Subscribe:   true,
				ListChanged: true,
			},
			Tools: &protocol.ToolsCapability{
				ListChanged: true,
			},
		},
		ServerInfo: protocol.ServerInfo{
			Name:    "example-stdio-server",
			Version: "2024-11-05",
		},
	}

	return s.sendResult(id, result)
}

// sendResult sends a successful response
func (s *Server) sendResult(id *json.RawMessage, result interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := protocol.Response{
		JSONRPC: "2.0",
		Result:  mustMarshal(result),
	}
	if id != nil {
		response.ID = *id
	}

	s.logger.Debug().Interface("response", response).Msg("Sending response")
	return s.writer.Encode(response)
}

// sendError sends an error response
func (s *Server) sendError(id *json.RawMessage, code int, message string, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errorData json.RawMessage
	if data != nil {
		errorData = mustMarshal(data)
	}

	response := protocol.Response{
		JSONRPC: "2.0",
		Error: &protocol.Error{
			Code:    code,
			Message: message,
			Data:    errorData,
		},
	}
	if id != nil {
		response.ID = *id
	}

	s.logger.Debug().Interface("response", response).Msg("Sending error response")
	return s.writer.Encode(response)
}

// mustMarshal marshals data to JSON or panics
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return data
}

// GetRegistry returns the server's provider registry
func (s *Server) GetRegistry() *pkg.ProviderRegistry {
	return s.registry
}
