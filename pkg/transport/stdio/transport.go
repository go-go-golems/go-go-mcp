package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type StdioTransport struct {
	mu             sync.Mutex
	logger         zerolog.Logger
	scanner        *bufio.Scanner
	writer         *json.Encoder
	handler        transport.RequestHandler
	sessionStore   session.SessionStore
	currentSession *session.Session
	signalChan     chan os.Signal
	wg             sync.WaitGroup
	cancel         context.CancelFunc
}

func NewStdioTransport(opts ...transport.TransportOption) (*StdioTransport, error) {
	options := &transport.TransportOptions{
		MaxMessageSize: 1024 * 1024, // 1MB default
		Logger:         zerolog.Nop(),
	}

	for _, opt := range opts {
		opt(options)
	}

	scanner := bufio.NewScanner(os.Stdin)
	if options.Stdio != nil && options.Stdio.ReadBufferSize > 0 {
		scanner.Buffer(make([]byte, options.Stdio.ReadBufferSize), options.Stdio.ReadBufferSize)
	} else {
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB default
	}

	pid := os.Getpid()
	baseLogger := log.Logger.With().Int("pid", pid).Logger()

	return &StdioTransport{
		scanner:    scanner,
		writer:     json.NewEncoder(os.Stdout),
		logger:     baseLogger.With().Str("component", "stdio_transport").Logger(),
		signalChan: make(chan os.Signal, 1),
	}, nil
}

func (s *StdioTransport) Listen(ctx context.Context, handler transport.RequestHandler) error {
	s.logger.Info().Msg("Starting stdio transport...")

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.handler = handler

	// Set up signal handling
	signal.Notify(s.signalChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(s.signalChan)

	// Create a channel for scanner errors
	scanErrChan := make(chan error, 1)

	// Create a cancellable context for the scanner
	scannerCtx, cancelScanner := context.WithCancel(ctx)
	defer cancelScanner()

	// Start scanning in a goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		scanLogger := s.logger.With().Str("routine", "scanner").Logger()

		for s.scanner.Scan() {
			select {
			case <-scannerCtx.Done():
				scanLogger.Debug().Msg("Context cancelled, stopping scanner")
				scanErrChan <- scannerCtx.Err()
				return
			default:
				line := s.scanner.Text()
				scanLogger.Debug().Msg("Received line")

				// Try to parse the message as either single request or batch
				message, err := transport.ParseMessage([]byte(line))
				if err != nil {
					scanLogger.Error().Err(err).Msg("Failed to parse JSON-RPC message")
					continue // Skip this message
				}

				if err := s.handleParsedMessage(scannerCtx, message); err != nil {
					scanLogger.Error().Err(err).Msg("Error handling parsed message")
					// Continue processing messages even if one fails
				}
			}
		}

		if err := s.scanner.Err(); err != nil {
			scanLogger.Error().Err(err).Msg("Scanner error")
			scanErrChan <- fmt.Errorf("scanner error: %w", err)
			return
		}

		scanLogger.Debug().Msg("Scanner reached EOF")
		scanErrChan <- io.EOF
	}()

	// Wait for either a signal, context cancellation, or scanner error
	select {
	case sig := <-s.signalChan:
		s.logger.Debug().Str("signal", sig.String()).Msg("Received signal in stdio transport")
		cancelScanner()
		return nil
	case <-ctx.Done():
		s.logger.Debug().Err(ctx.Err()).Msg("Context cancelled in stdio transport")
		return ctx.Err()
	case err := <-scanErrChan:
		if err == io.EOF {
			s.logger.Debug().Msg("Scanner completed normally")
			return nil
		}
		s.logger.Error().Err(err).Msg("Scanner error in stdio transport")
		return err
	}
}

func (s *StdioTransport) Send(ctx context.Context, response *protocol.Response) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a scoped logger for this response
	respLogger := s.logger.With().Logger()
	if session_, ok := session.GetSessionFromContext(ctx); ok {
		respLogger = respLogger.With().Str("session_id", string(session_.ID)).Logger()
	}

	respLogger.Debug().Interface("response", response).Msg("Sending response")
	return s.writer.Encode(response)
}

func (s *StdioTransport) Close(ctx context.Context) error {
	closeLogger := s.logger.With().Str("operation", "close").Logger()
	closeLogger.Info().Msg("Stopping stdio transport")

	if s.cancel != nil {
		closeLogger.Debug().Msg("Cancelling context")
		s.cancel()
	}

	// Wait for context to be done or timeout
	done := make(chan struct{})
	go func() {
		closeLogger.Debug().Msg("Waiting for goroutines to finish")
		s.wg.Wait()
		closeLogger.Debug().Msg("Goroutines finished")
		close(done)
	}()

	select {
	case <-done:
		closeLogger.Debug().Msg("All goroutines finished")
		return nil
	case <-ctx.Done():
		closeLogger.Debug().Err(ctx.Err()).Msg("Stop context cancelled before clean shutdown")
		return ctx.Err()
	case <-time.After(100 * time.Millisecond): // Give a small grace period for cleanup
		closeLogger.Debug().Msg("Stop completed with timeout")
		return nil
	}
}

func (s *StdioTransport) Info() transport.TransportInfo {
	return transport.TransportInfo{
		Type: "stdio",
		Capabilities: map[string]bool{
			"bidirectional": true,
			"persistent":    false,
		},
		Metadata: map[string]string{
			"pid": fmt.Sprintf("%d", os.Getpid()),
		},
	}
}

// SetSessionStore implements the transport.Transport interface
func (s *StdioTransport) SetSessionStore(store session.SessionStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionStore = store
	s.logger.Info().Msg("Session store set for stdio transport")
}

// manageSession handles session logic for incoming requests.
// It creates a new session for `initialize` requests and uses the existing one otherwise.
// It returns a context enhanced with the appropriate session.
func (s *StdioTransport) manageSession(baseCtx context.Context, req *protocol.Request) context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sessionStore == nil {
		s.logger.Error().Msg("Session store not set in stdio transport, cannot manage sessions")
		return baseCtx // Return base context if store is missing
	}

	// Create a scoped logger for session operations
	var sessionLogger zerolog.Logger
	if s.currentSession != nil {
		sessionLogger = s.logger.With().Str("session_id", string(s.currentSession.ID)).Logger()
	} else {
		sessionLogger = s.logger
	}

	// Create a new session specifically for the 'initialize' method
	if req.Method == "initialize" {
		// If there was an old session, maybe log its closure?
		if s.currentSession != nil {
			sessionLogger.Info().Msg("Replacing existing stdio session due to initialize request")
			// Optionally delete the old session from the store if it shouldn't persist
			// s.sessionStore.Delete(s.currentSession.ID)
		}
		s.currentSession = s.sessionStore.Create()
		sessionLogger = s.logger.With().Str("session_id", string(s.currentSession.ID)).Logger()
		sessionLogger.Info().Msg("Created new session for stdio connection (initialize)")
	} else if s.currentSession == nil {
		// If it's not 'initialize' and no session exists, create one.
		// This handles the case where the first request isn't 'initialize'.
		s.currentSession = s.sessionStore.Create()
		sessionLogger = s.logger.With().Str("session_id", string(s.currentSession.ID)).Logger()
		sessionLogger.Info().Msg("Created implicit session for stdio connection (first request was not initialize)")
	}

	// Enhance context with the current session
	return session.WithSession(baseCtx, s.currentSession)
}

// handleParsedMessage handles either single requests or batch requests
func (s *StdioTransport) handleParsedMessage(ctx context.Context, message interface{}) error {
	switch msg := message.(type) {
	case *protocol.Request:
		// Handle session creation/update based on method
		sessionCtx := s.manageSession(ctx, msg)
		return s.handleMessage(sessionCtx, msg)
	case protocol.BatchRequest:
		return s.handleBatchMessage(ctx, msg)
	default:
		return fmt.Errorf("unknown message type: %T", message)
	}
}

// handleBatchMessage handles batch requests
func (s *StdioTransport) handleBatchMessage(ctx context.Context, batch protocol.BatchRequest) error {
	batchLogger := s.logger.With().Int("batch_size", len(batch)).Logger()
	batchLogger.Debug().Msg("Handling batch request")

	// Process each request in the batch and collect responses
	responses := make(protocol.BatchResponse, 0, len(batch))

	for i, request := range batch {
		reqLogger := batchLogger.With().
			Int("batch_index", i).
			Str("method", request.Method).
			Logger()

		// Handle session creation/update for each request
		sessionCtx := s.manageSession(ctx, &request)

		// Handle the individual request
		if !transport.IsNotification(&request) {
			// This is a request that expects a response
			response, err := s.handler.HandleRequest(sessionCtx, &request)
			if err != nil {
				reqLogger.Error().Err(err).Msg("Error handling batch request")
				jsonErr := transport.ProcessError(err)
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    jsonErr.Code,
						Message: jsonErr.Message,
						Data:    jsonErr.Data,
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
			if err := s.handler.HandleNotification(sessionCtx, &notif); err != nil {
				reqLogger.Error().Err(err).Msg("Error handling batch notification")
			}
		}
	}

	// Send batch response if there are any responses
	if len(responses) > 0 {
		return s.sendBatchResponse(ctx, responses)
	}

	return nil
}

func (s *StdioTransport) handleMessage(ctx context.Context, request *protocol.Request) error {
	// Create a scoped logger for this message
	msgLogger := s.logger.With().Str("method", request.Method).Logger()
	if session_, ok := session.GetSessionFromContext(ctx); ok {
		msgLogger = msgLogger.With().Str("session_id", string(session_.ID)).Logger()
	}

	// Handle requests vs notifications based on ID presence
	if !transport.IsNotification(request) {
		msgLogger.Debug().
			RawJSON("id", request.ID).
			Msg("Handling request")
		return s.handleRequest(ctx, request)
	}

	msgLogger.Debug().Msg("Handling notification")
	return s.handleNotification(ctx, protocol.Notification{
		Method: request.Method,
		Params: request.Params,
	})
}

func (s *StdioTransport) handleRequest(ctx context.Context, request *protocol.Request) error {
	// Create a scoped logger for this request
	reqLogger := s.logger.With().Str("method", request.Method).Logger()
	if session_, ok := session.GetSessionFromContext(ctx); ok {
		reqLogger = reqLogger.With().Str("session_id", string(session_.ID)).Logger()
	}

	response, err := s.handler.HandleRequest(ctx, request) // Pass context
	if err != nil {
		reqLogger.Error().Err(err).Msg("Error handling request")
		jsonErr := transport.ProcessError(err) // Convert to JSON-RPC error
		return s.sendError(request.ID, jsonErr.Code, jsonErr.Message, jsonErr.Data)
	}

	if response != nil {
		return s.Send(ctx, response) // Pass context
	}

	return nil
}

func (s *StdioTransport) handleNotification(ctx context.Context, notification protocol.Notification) error {
	// Create a scoped logger for this notification
	notifLogger := s.logger.With().Str("method", notification.Method).Logger()
	if session_, ok := session.GetSessionFromContext(ctx); ok {
		notifLogger = notifLogger.With().Str("session_id", string(session_.ID)).Logger()
	}

	if err := s.handler.HandleNotification(ctx, &notification); err != nil { // Pass context
		notifLogger.Error().Err(err).Msg("Error handling notification")
		// Don't send error responses for notifications
	}

	return nil
}

func (s *StdioTransport) sendError(id json.RawMessage, code int, message string, data interface{}) error {
	var errorData json.RawMessage
	if data != nil {
		// Check if data is already json.RawMessage
		if rawData, ok := data.(json.RawMessage); ok {
			errorData = rawData
		} else {
			var err error
			errorData, err = json.Marshal(data)
			if err != nil {
				// If we can't marshal the error data, log it and send a simpler error
				s.logger.Error().Err(err).Interface("data", data).Msg("Failed to marshal error data")
				// Avoid recursion by creating the error struct directly
				errorResponse := &protocol.Response{
					Error: &protocol.Error{
						Code:    transport.ErrCodeInternal,
						Message: "Internal error marshaling error data",
					},
					ID: id,
				}
				return s.Send(context.Background(), errorResponse)
			}
		}
	}

	response := &protocol.Response{
		JSONRPC: "2.0", // Ensure JSONRPC version is set
		Error: &protocol.Error{
			Code:    code,
			Message: message,
			Data:    errorData,
		},
		ID: id,
	}

	errLogger := s.logger.With().Int("error_code", code).Logger()
	errLogger.Debug().Interface("response", response).Msg("Sending error response")
	return s.Send(context.Background(), response)
}

// sendBatchResponse sends a batch response to the client
func (s *StdioTransport) sendBatchResponse(ctx context.Context, responses protocol.BatchResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a scoped logger for this batch response
	batchLogger := s.logger.With().Int("response_count", len(responses)).Logger()
	if session_, ok := session.GetSessionFromContext(ctx); ok {
		batchLogger = batchLogger.With().Str("session_id", string(session_.ID)).Logger()
	}

	batchLogger.Debug().Interface("responses", responses).Msg("Sending batch response")
	return s.writer.Encode(responses)
}
