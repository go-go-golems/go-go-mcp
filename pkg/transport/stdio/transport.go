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

	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/rs/zerolog"
)

type StdioTransport struct {
	mu         sync.Mutex
	logger     zerolog.Logger
	scanner    *bufio.Scanner
	writer     *json.Encoder
	handler    transport.RequestHandler
	signalChan chan os.Signal
	wg         sync.WaitGroup
	cancel     context.CancelFunc
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

	// Create a ConsoleWriter that writes to stderr with a SERVER tag
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("[STDIO] %s", i)
		},
	}

	// Create a new logger that writes to the tagged stderr
	taggedLogger := options.Logger.Output(consoleWriter)

	return &StdioTransport{
		scanner:    scanner,
		writer:     json.NewEncoder(os.Stdout),
		logger:     taggedLogger,
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
		scanErrChan <- io.EOF
	}()

	// Wait for either a signal, context cancellation, or scanner error
	select {
	case sig := <-s.signalChan:
		s.logger.Debug().
			Str("signal", sig.String()).
			Msg("Received signal in stdio transport")
		cancelScanner()
		return nil
	case <-ctx.Done():
		s.logger.Debug().
			Err(ctx.Err()).
			Msg("Context cancelled in stdio transport")
		return ctx.Err()
	case err := <-scanErrChan:
		if err == io.EOF {
			s.logger.Debug().Msg("Scanner completed normally")
			return nil
		}
		s.logger.Error().
			Err(err).
			Msg("Scanner error in stdio transport")
		return err
	}
}

func (s *StdioTransport) Send(ctx context.Context, response *transport.Response) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug().Interface("response", response).Msg("Sending response")
	return s.writer.Encode(response)
}

func (s *StdioTransport) Close(ctx context.Context) error {
	s.logger.Info().Msg("Stopping stdio transport")

	if s.cancel != nil {
		s.cancel()
	}

	// Wait for context to be done or timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Debug().Msg("All goroutines finished")
		return nil
	case <-ctx.Done():
		s.logger.Debug().
			Err(ctx.Err()).
			Msg("Stop context cancelled before clean shutdown")
		return ctx.Err()
	case <-time.After(100 * time.Millisecond): // Give a small grace period for cleanup
		s.logger.Debug().Msg("Stop completed with timeout")
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

func (s *StdioTransport) handleMessage(message string) error {
	s.logger.Debug().
		Str("message", message).
		Msg("Processing message")

	// Parse the base message structure
	var request transport.Request
	if err := json.Unmarshal([]byte(message), &request); err != nil {
		s.logger.Error().
			Err(err).
			Str("message", message).
			Msg("Failed to parse message as JSON-RPC request")
		return s.sendError(nil, transport.ErrCodeParse, "Parse error", err)
	}

	// Handle requests vs notifications based on ID presence
	if request.ID != "" {
		s.logger.Debug().
			Str("id", request.ID).
			Str("method", request.Method).
			Msg("Handling request")
		return s.handleRequest(request)
	}

	s.logger.Debug().
		Str("method", request.Method).
		Msg("Handling notification")
	return s.handleNotification(request)
}

func (s *StdioTransport) handleRequest(request transport.Request) error {
	response, err := s.handler.HandleRequest(context.Background(), &request)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("method", request.Method).
			Msg("Error handling request")
		return s.sendError(&request.ID, transport.ErrCodeInternal, "Internal error", err)
	}

	if response != nil {
		return s.Send(context.Background(), response)
	}

	return nil
}

func (s *StdioTransport) handleNotification(request transport.Request) error {
	notification := &transport.Notification{
		Method:  request.Method,
		Params:  request.Params,
		Headers: request.Headers,
	}

	if err := s.handler.HandleNotification(context.Background(), notification); err != nil {
		s.logger.Error().
			Err(err).
			Str("method", request.Method).
			Msg("Error handling notification")
		// Don't send error responses for notifications
	}

	return nil
}

func (s *StdioTransport) sendError(id *string, code int, message string, data interface{}) error {
	var errorData json.RawMessage
	if data != nil {
		var err error
		errorData, err = json.Marshal(data)
		if err != nil {
			// If we can't marshal the error data, log it and send a simpler error
			s.logger.Error().Err(err).Interface("data", data).Msg("Failed to marshal error data")
			return s.sendError(id, transport.ErrCodeInternal, "Internal error marshaling error data", nil)
		}
	}

	response := &transport.Response{
		Error: &transport.ResponseError{
			Code:    code,
			Message: message,
			Data:    errorData,
		},
	}
	if id != nil {
		response.ID = *id
	}

	s.logger.Debug().Interface("response", response).Msg("Sending error response")
	return s.Send(context.Background(), response)
}
