package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/server/dispatcher"
	"github.com/go-go-golems/go-go-mcp/pkg/services"
	"github.com/rs/zerolog"
)

// StdioServer handles stdio transport for MCP protocol
type StdioServer struct {
	scanner    *bufio.Scanner
	writer     *json.Encoder
	logger     zerolog.Logger
	dispatcher *dispatcher.Dispatcher
}

// NewStdioServer creates a new stdio server instance
func NewStdioServer(logger zerolog.Logger, ps services.PromptService, rs services.ResourceService, ts services.ToolService, is services.InitializeService) *StdioServer {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	return &StdioServer{
		scanner:    scanner,
		writer:     json.NewEncoder(os.Stdout),
		logger:     logger,
		dispatcher: dispatcher.NewDispatcher(logger, ps, rs, ts, is),
	}
}

// Start begins listening for and handling messages on stdio
func (s *StdioServer) Start() error {
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
func (s *StdioServer) handleMessage(message string) error {
	s.logger.Debug().Str("message", message).Msg("Received message")

	// Parse the base message structure
	var request protocol.Request
	if err := json.Unmarshal([]byte(message), &request); err != nil {
		response, _ := dispatcher.NewErrorResponse(nil, -32700, "Parse error", err)
		return s.writer.Encode(response)
	}

	// Use the dispatcher to handle the request
	ctx := context.Background()
	response, err := s.dispatcher.Dispatch(ctx, request)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error dispatching request")
		response, _ = dispatcher.NewErrorResponse(request.ID, -32603, "Internal error", err)
	}

	// Send response if it's not nil (notifications don't have responses)
	if response != nil {
		s.logger.Debug().Interface("response", response).Msg("Sending response")
		return s.writer.Encode(response)
	}

	return nil
}
