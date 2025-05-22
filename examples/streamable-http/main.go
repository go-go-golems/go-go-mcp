package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/streamable_http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SimpleHandler implements a basic request handler for demonstration
type SimpleHandler struct {
	logger zerolog.Logger
}

func NewSimpleHandler() *SimpleHandler {
	return &SimpleHandler{
		logger: log.Logger.With().Str("component", "simple_handler").Logger(),
	}
}

func (h *SimpleHandler) HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	h.logger.Info().
		Str("method", req.Method).
		RawJSON("params", req.Params).
		Msg("Handling request")

	// Simple echo response
	result := map[string]interface{}{
		"echo":      req.Method,
		"params":    json.RawMessage(req.Params),
		"timestamp": time.Now().Unix(),
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  resultBytes,
	}, nil
}

func (h *SimpleHandler) HandleNotification(ctx context.Context, notif *protocol.Notification) error {
	h.logger.Info().
		Str("method", notif.Method).
		RawJSON("params", notif.Params).
		Msg("Handling notification")
	return nil
}

func main() {
	// Configure logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: log.Logger})

	ctx := context.Background()

	// Create transport with streamable HTTP options
	transport, err := streamable_http.NewStreamableHTTPTransport(
		transport.WithStreamableHTTPOptions(transport.StreamableHTTPOptions{
			Addr:            ":8080",
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for demo purposes
				// In production, implement proper origin checking
				return true
			},
		}),
		transport.WithLogger(log.Logger),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create transport")
	}

	// Create session store (in-memory for demo)
	sessionStore := session.NewInMemorySessionStore()
	transport.SetSessionStore(sessionStore)

	// Create handler
	handler := NewSimpleHandler()

	log.Info().Msg("Starting streamable HTTP MCP server on :8080")
	log.Info().Msg("WebSocket endpoint: ws://localhost:8080/stream")
	log.Info().Msg("HTTP endpoint: http://localhost:8080/messages")

	// Start the transport
	if err := transport.Listen(ctx, handler); err != nil {
		log.Fatal().Err(err).Msg("Transport failed")
	}
}
