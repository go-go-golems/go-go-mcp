package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/go-go-golems/go-mcp/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Server handles MCP protocol communication
type Server struct {
	mu      sync.Mutex
	scanner *bufio.Scanner
	writer  *json.Encoder
	logger  zerolog.Logger
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
		scanner: scanner,
		writer:  writer,
		logger:  logger,
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
	var request pkg.Request
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
func (s *Server) handleRequest(request pkg.Request) error {
	switch request.Method {
	case "initialize":
		var params pkg.InitializeParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}
		s.logger.Info().Interface("params", params).Msg("Handling initialize request")
		return s.handleInitialize(&request.ID, &params)

	case "ping":
		return s.sendResult(&request.ID, struct{}{})

	case "prompts/list":
		prompts := s.listPrompts()
		return s.sendResult(&request.ID, map[string]interface{}{
			"prompts": prompts,
		})

	case "prompts/get":
		// Parse parameters
		var params struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments,omitempty"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}

		message, err := s.getPrompt(params.Name, params.Arguments)
		if err != nil {
			return s.sendError(&request.ID, -32602, "Invalid prompt", err)
		}

		return s.sendResult(&request.ID, map[string]interface{}{
			"description": "A simple prompt with optional context and topic arguments",
			"messages":    []pkg.PromptMessage{*message},
		})

	default:
		return s.sendError(&request.ID, -32601, "Method not found", nil)
	}
}

// handleNotification processes a notification message
func (s *Server) handleNotification(request pkg.Request) error {
	var notification pkg.Notification
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
func (s *Server) handleInitialize(id *json.RawMessage, params *pkg.InitializeParams) error {
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
	result := pkg.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: pkg.ServerCapabilities{
			Logging: &pkg.LoggingCapability{},
			Prompts: &pkg.PromptsCapability{
				ListChanged: true,
			},
		},
		ServerInfo: pkg.ServerInfo{
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

	response := pkg.Response{
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

	response := pkg.Response{
		JSONRPC: "2.0",
		Error: &pkg.Error{
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

func (s *Server) listPrompts() []pkg.Prompt {
	return []pkg.Prompt{
		{
			Name: "simple",
			Description: "A simple prompt that can take optional context and topic " +
				"arguments",
			Arguments: []pkg.PromptArgument{
				{
					Name:        "context",
					Description: "Additional context to consider",
					Required:    false,
				},
				{
					Name:        "topic",
					Description: "Specific topic to focus on",
					Required:    false,
				},
			},
		},
	}
}

func (s *Server) getPrompt(name string, arguments map[string]string) (*pkg.PromptMessage, error) {
	if name != "simple" {
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}

	// Create messages based on arguments
	messages := []pkg.PromptMessage{}

	// Add context if provided
	if context, ok := arguments["context"]; ok {
		messages = append(messages, pkg.PromptMessage{
			Role: "user",
			Content: pkg.PromptContent{
				Type: "text",
				Text: fmt.Sprintf("Here is some relevant context: %s", context),
			},
		})
	}

	// Add main prompt
	prompt := "Please help me with "
	if topic, ok := arguments["topic"]; ok {
		prompt += fmt.Sprintf("the following topic: %s", topic)
	} else {
		prompt += "whatever questions I may have."
	}

	messages = append(messages, pkg.PromptMessage{
		Role: "user",
		Content: pkg.PromptContent{
			Type: "text",
			Text: prompt,
		},
	})

	return &pkg.PromptMessage{
		Role: "user",
		Content: pkg.PromptContent{
			Type: "text",
			Text: prompt,
		},
	}, nil
}

func main() {
	server := NewServer()
	if err := server.Start(); err != nil && err != io.EOF {
		log.Fatal().Err(err).Msg("Server error")
	}
}