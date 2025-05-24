package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/stdio"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// DemoHandler implements a handler that demonstrates batch processing and cancellation
type DemoHandler struct {
	logger          zerolog.Logger
	pendingRequests map[string]context.CancelFunc
	requestContexts map[string]context.Context
}

func NewDemoHandler() *DemoHandler {
	return &DemoHandler{
		logger:          log.Logger.With().Str("component", "demo_handler").Logger(),
		pendingRequests: make(map[string]context.CancelFunc),
		requestContexts: make(map[string]context.Context),
	}
}

func (h *DemoHandler) HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	reqID := string(req.ID)

	// Store the request context and cancellation function for potential cancellation
	requestCtx, cancel := context.WithCancel(ctx)
	h.pendingRequests[reqID] = cancel
	h.requestContexts[reqID] = requestCtx

	// Clean up when done
	defer func() {
		delete(h.pendingRequests, reqID)
		delete(h.requestContexts, reqID)
	}()

	h.logger.Info().
		Str("method", req.Method).
		Str("id", reqID).
		Msg("Handling request")

	switch req.Method {
	case "echo":
		return h.handleEcho(requestCtx, req)
	case "slow_task":
		return h.handleSlowTask(requestCtx, req)
	case "ping":
		return h.handlePing(requestCtx, req)
	default:
		return nil, fmt.Errorf("unknown method: %s", req.Method)
	}
}

func (h *DemoHandler) HandleNotification(ctx context.Context, notif *protocol.Notification) error {
	h.logger.Info().
		Str("method", notif.Method).
		Msg("Handling notification")

	switch notif.Method {
	case "notifications/cancelled":
		return h.handleCancellation(ctx, notif)
	case "notifications/initialized":
		h.logger.Info().Msg("Client initialized")
		return nil
	default:
		h.logger.Warn().Str("method", notif.Method).Msg("Unknown notification method")
		return nil
	}
}

func (h *DemoHandler) HandleBatchRequest(ctx context.Context, batch protocol.BatchRequest) (protocol.BatchResponse, error) {
	h.logger.Info().Int("batch_size", len(batch)).Msg("Handling batch request")

	responses := make(protocol.BatchResponse, 0, len(batch))

	for i, request := range batch {
		reqLogger := h.logger.With().
			Int("batch_index", i).
			Str("method", request.Method).
			Logger()

		// Handle individual request
		if !transport.IsNotification(&request) {
			// This is a request that expects a response
			response, err := h.HandleRequest(ctx, &request)
			if err != nil {
				reqLogger.Error().Err(err).Msg("Error handling batch request")
				// Convert error to JSON-RPC error response
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    -32603, // Internal error
						Message: err.Error(),
					},
				}
			}
			if response != nil {
				responses = append(responses, *response)
			}
		} else {
			// This is a notification - handle it but don't add to responses
			notif := protocol.Notification{
				JSONRPC: request.JSONRPC,
				Method:  request.Method,
				Params:  request.Params,
			}
			if err := h.HandleNotification(ctx, &notif); err != nil {
				reqLogger.Error().Err(err).Msg("Error handling batch notification")
			}
		}
	}

	return responses, nil
}

func (h *DemoHandler) handleEcho(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
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

func (h *DemoHandler) handleSlowTask(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	h.logger.Info().Msg("Starting slow task")

	// Simulate a slow task that can be cancelled
	select {
	case <-time.After(5 * time.Second):
		// Task completed normally
		result := map[string]interface{}{
			"status":    "completed",
			"message":   "Slow task finished successfully",
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

	case <-ctx.Done():
		// Task was cancelled
		h.logger.Info().Msg("Slow task was cancelled")
		return nil, fmt.Errorf("request was cancelled: %w", ctx.Err())
	}
}

func (h *DemoHandler) handlePing(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	result := map[string]interface{}{
		"pong":      true,
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

func (h *DemoHandler) handleCancellation(ctx context.Context, notif *protocol.Notification) error {
	var params protocol.CancellationParams
	if err := json.Unmarshal(notif.Params, &params); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal cancellation parameters")
		return nil // Don't return error for notifications
	}

	h.logger.Info().
		Str("request_id", params.RequestID).
		Str("reason", params.Reason).
		Msg("Processing cancellation request")

	// Find and cancel the pending request
	if cancel, exists := h.pendingRequests[params.RequestID]; exists {
		h.logger.Info().
			Str("request_id", params.RequestID).
			Msg("Cancelling pending request")
		cancel()

		// Clean up immediately
		delete(h.pendingRequests, params.RequestID)
		delete(h.requestContexts, params.RequestID)
	} else {
		h.logger.Warn().
			Str("request_id", params.RequestID).
			Msg("No pending request found for cancellation")
	}

	return nil
}

func main() {
	// Configure logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: log.Logger})

	log.Info().Msg("Starting JSON-RPC Batch and Cancellation Demo")
	log.Info().Msg("This demo shows:")
	log.Info().Msg("  - JSON-RPC batch request processing")
	log.Info().Msg("  - Request cancellation via notifications")
	log.Info().Msg("  - Context-aware long-running operations")
	log.Info().Msg("")
	log.Info().Msg("Try sending these sample messages via stdin:")
	log.Info().Msg("")
	log.Info().Msg("Single request:")
	log.Info().Msg(`  {"jsonrpc":"2.0","id":"1","method":"echo","params":{"message":"hello"}}`)
	log.Info().Msg("")
	log.Info().Msg("Batch request:")
	log.Info().Msg(`  [{"jsonrpc":"2.0","id":"1","method":"ping"},{"jsonrpc":"2.0","id":"2","method":"echo","params":{"test":true}}]`)
	log.Info().Msg("")
	log.Info().Msg("Long-running task (try cancelling with cancellation notification):")
	log.Info().Msg(`  {"jsonrpc":"2.0","id":"slow-1","method":"slow_task"}`)
	log.Info().Msg("")
	log.Info().Msg("Cancellation notification:")
	log.Info().Msg(`  {"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"slow-1","reason":"User requested cancellation"}}`)
	log.Info().Msg("")

	// Create demo handler
	handler := NewDemoHandler()

	// Create stdio transport
	stdioTransport, err := stdio.NewStdioTransport()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create transport")
	}

	// Create session store (not really needed for this demo but required by interface)
	sessionStore := session.NewInMemorySessionStore()
	stdioTransport.SetSessionStore(sessionStore)

	// Start the transport
	ctx := context.Background()
	if err := stdioTransport.Listen(ctx, handler); err != nil {
		log.Fatal().Err(err).Msg("Transport failed")
	}
}
