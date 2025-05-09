package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/client"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
)

// SSEBridgeServer is a stdio server that forwards all requests to an SSE server
type SSEBridgeServer struct {
	scanner    *bufio.Scanner
	writer     *json.Encoder
	logger     zerolog.Logger
	sseClient  *client.SSETransport
	signalChan chan os.Signal
	mu         sync.Mutex
}

// NewSSEBridgeServer creates a new stdio server instance that forwards to SSE
func NewSSEBridgeServer(logger zerolog.Logger, sseURL string) *SSEBridgeServer {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	// Create a ConsoleWriter that writes to stderr with a SERVER tag
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("[STDIO-SSE-BRIDGE] %s", i)
		},
	}

	// Create a new logger that writes to the tagged stderr
	taggedLogger := logger.Output(consoleWriter).With().Caller().Logger()

	// Strip trailing slashes from the SSE URL
	sseURL = strings.TrimRight(sseURL, "/")

	sseClient := client.NewSSETransport(sseURL, taggedLogger)

	s := &SSEBridgeServer{
		scanner:    scanner,
		writer:     json.NewEncoder(os.Stdout),
		logger:     taggedLogger,
		sseClient:  sseClient,
		signalChan: make(chan os.Signal, 1),
	}

	// Set up notification handler to write to stdout
	sseClient.SetNotificationHandler(func(response *protocol.Response) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if err := s.writer.Encode(response); err != nil {
			s.logger.Error().Err(err).Msg("Failed to write notification to stdout")
		}
	})

	return s
}

// Start begins listening for and handling messages on stdio
func (s *SSEBridgeServer) Start(ctx context.Context) error {
	s.logger.Info().Msg("Starting stdio-sse bridge server...")

	// Set up signal handling
	signal.Notify(s.signalChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(s.signalChan)

	// Create a channel for scanner errors
	scanErrChan := make(chan error, 1)

	// Create a cancellable context for the scanner
	scannerCtx, cancelScanner := context.WithCancel(ctx)
	defer cancelScanner()

	// Start scanning in a goroutine
	go func() {
		for s.scanner.Scan() {
			select {
			case <-scannerCtx.Done():
				s.logger.Debug().Msg("Context cancelled, stopping scanner")
				scanErrChan <- scannerCtx.Err()
				return
			default:
				line := s.scanner.Text()
				s.logger.Debug().
					Str("line", line).
					Msg("Received line")
				if err := s.handleMessage(line); err != nil {
					s.logger.Error().Err(err).Msg("Error handling message")
					// Continue processing messages even if one fails
				}
			}
		}

		if err := s.scanner.Err(); err != nil {
			s.logger.Error().
				Err(err).
				Msg("Scanner error")
			scanErrChan <- fmt.Errorf("scanner error: %w", err)
			return
		}

		s.logger.Debug().Msg("Scanner reached EOF")
		scanErrChan <- nil
	}()

	// Wait for either a signal, context cancellation, or scanner error
	select {
	case sig := <-s.signalChan:
		s.logger.Debug().
			Str("signal", sig.String()).
			Msg("Received signal in stdio server")
		cancelScanner()
		return nil
	case <-ctx.Done():
		s.logger.Debug().
			Err(ctx.Err()).
			Msg("Context cancelled in stdio server")
		return ctx.Err()
	case err := <-scanErrChan:
		if err == nil {
			s.logger.Debug().Msg("Scanner completed normally")
		} else {
			s.logger.Error().
				Err(err).
				Msg("Scanner error in stdio server")
		}
		return err
	}
}

// Stop gracefully stops the stdio server
func (s *SSEBridgeServer) Stop(ctx context.Context) error {
	s.logger.Info().Msg("Stopping stdio-sse bridge server")

	// Close SSE client connection
	if err := s.sseClient.Close(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Error closing SSE client")
	}

	// Wait for context to be done or timeout
	select {
	case <-ctx.Done():
		s.logger.Debug().
			Err(ctx.Err()).
			Msg("Stop context cancelled before clean shutdown")
		return ctx.Err()
	case <-time.After(100 * time.Millisecond): // Give a small grace period for cleanup
		s.logger.Debug().Msg("Stop completed successfully")
		return nil
	}
}

// handleMessage processes a single message
func (s *SSEBridgeServer) handleMessage(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug().
		Str("message", message).
		Msg("Processing message")

	// Parse the base message structure
	var request protocol.Request
	if err := json.Unmarshal([]byte(message), &request); err != nil {
		s.logger.Error().
			Err(err).
			Str("message", message).
			Msg("Failed to parse message as JSON-RPC request")
		return s.sendError(nil, -32700, "Parse error", err)
	}

	// Handle cookie management
	switch request.Method {
	case "cookies/set":
		var params struct {
			Cookies []*http.Cookie `json:"cookies"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.sendError(&request.ID, -32602, "Invalid params", err)
		}
		s.sseClient.SetCookies(params.Cookies)
		return s.writer.Encode(&protocol.Response{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result:  json.RawMessage(`{"success":true}`),
		})
	case "cookies/get":
		cookies := s.sseClient.GetCookies()
		cookiesJSON, err := json.Marshal(cookies)
		if err != nil {
			return s.sendError(&request.ID, -32603, "Internal error", err)
		}
		return s.writer.Encode(&protocol.Response{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result:  cookiesJSON,
		})
	default:
		// Forward the request to the SSE server
		response, err := s.sseClient.Send(context.Background(), &request)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("method", request.Method).
				Msg("Error forwarding request to SSE server")
			return s.sendError(&request.ID, -32603, "Internal error", err)
		}

		if response == nil {
			s.logger.Debug().Msg("No response from SSE server")
			return nil
		}

		jsonResponse, _ := json.MarshalIndent(response, "", "  ")
		s.logger.Debug().RawJSON("response", jsonResponse).Msg("Forwarded request to SSE server")

		// Send the response back over stdio
		return s.writer.Encode(response)
	}
}

// sendError sends an error response
func (s *SSEBridgeServer) sendError(id *json.RawMessage, code int, message string, data interface{}) error {
	var errorData json.RawMessage
	if data != nil {
		var err error
		errorData, err = json.Marshal(data)
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
