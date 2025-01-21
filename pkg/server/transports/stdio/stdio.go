package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/services"
	"github.com/rs/zerolog"
)

// Server handles stdio transport for MCP protocol
type Server struct {
	scanner           *bufio.Scanner
	writer            *json.Encoder
	logger            zerolog.Logger
	promptService     services.PromptService
	resourceService   services.ResourceService
	toolService       services.ToolService
	initializeService services.InitializeService
}

// NewServer creates a new stdio server instance
func NewServer(logger zerolog.Logger, ps services.PromptService, rs services.ResourceService, ts services.ToolService, is services.InitializeService) *Server {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	return &Server{
		scanner:           scanner,
		writer:            json.NewEncoder(os.Stdout),
		logger:            logger,
		promptService:     ps,
		resourceService:   rs,
		toolService:       ts,
		initializeService: is,
	}
}

// Start begins listening for and handling messages on stdio
func (s *Server) Start() error {
	s.logger.Info().Msg("Starting stdio server...")

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
	ctx := context.Background()

	switch request.Method {
	case "initialize":
		var params protocol.InitializeParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}
		s.logger.Info().Interface("params", params).Msg("Handling initialize request")

		result, err := s.initializeService.Initialize(ctx, params)
		if err != nil {
			return s.sendError(&request.ID, -32603, "Initialize failed", err)
		}
		return s.sendResult(&request.ID, result)

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

		prompts, nextCursor, err := s.promptService.ListPrompts(ctx, cursor)
		if err != nil {
			return s.sendError(&request.ID, -32603, "Internal error", err)
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"prompts":    prompts,
			"nextCursor": nextCursor,
		})

	case "prompts/get":
		var params struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}

		message, err := s.promptService.GetPrompt(ctx, params.Name, params.Arguments)
		if err != nil {
			return s.sendError(&request.ID, -32602, "Prompt not found", err)
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"description": "Prompt from provider",
			"messages":    []protocol.PromptMessage{*message},
		})

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

		resources, nextCursor, err := s.resourceService.ListResources(ctx, cursor)
		if err != nil {
			return s.sendError(&request.ID, -32603, "Internal error", err)
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"resources":  resources,
			"nextCursor": nextCursor,
		})

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}

		content, err := s.resourceService.ReadResource(ctx, params.URI)
		if err != nil {
			return s.sendError(&request.ID, -32002, "Resource not found", err)
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"contents": []protocol.ResourceContent{*content},
		})

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

		tools, nextCursor, err := s.toolService.ListTools(ctx, cursor)
		if err != nil {
			return s.sendError(&request.ID, -32603, "Internal error", err)
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"tools":      tools,
			"nextCursor": nextCursor,
		})

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}

		result, err := s.toolService.CallTool(ctx, params.Name, params.Arguments)
		if err != nil {
			return s.sendError(&request.ID, -32602, "Tool not found", err)
		}

		return s.sendResult(&request.ID, result)

	default:
		return s.sendError(&request.ID, -32601, "Method not found", nil)
	}
}

// handleNotification processes a notification message
func (s *Server) handleNotification(request protocol.Request) error {
	switch request.Method {
	case "notifications/initialized":
		s.logger.Info().Msg("Client initialized")
		return nil

	default:
		s.logger.Warn().Str("method", request.Method).Msg("Unknown notification method")
		return nil
	}
}

// marshalJSON marshals data to JSON and returns any error
func (s *Server) marshalJSON(v interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		s.logger.Error().Err(err).Interface("value", v).Msg("Failed to marshal JSON")
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// sendResult sends a successful response
func (s *Server) sendResult(id *json.RawMessage, result interface{}) error {
	resultJSON, err := s.marshalJSON(result)
	if err != nil {
		return s.sendError(id, -32603, "Internal error", err)
	}

	response := protocol.Response{
		JSONRPC: "2.0",
		Result:  resultJSON,
	}
	if id != nil {
		response.ID = *id
	}

	s.logger.Debug().Interface("response", response).Msg("Sending response")
	return s.writer.Encode(response)
}

// sendError sends an error response
func (s *Server) sendError(id *json.RawMessage, code int, message string, data interface{}) error {
	var errorData json.RawMessage
	if data != nil {
		var err error
		errorData, err = s.marshalJSON(data)
		if err != nil {
			// If we can't marshal the error data, log it and send a simpler error
			s.logger.Error().Err(err).Interface("data", data).Msg("Failed to marshal error data")
			return s.sendError(id, -32603, "Internal error marshaling error data", nil)
		}
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
