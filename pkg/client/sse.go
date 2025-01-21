package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/r3labs/sse/v2"
	"github.com/rs/zerolog"
)

// SSETransport implements Transport using Server-Sent Events
type SSETransport struct {
	mu          sync.Mutex
	baseURL     string
	client      *http.Client
	sseClient   *sse.Client
	events      chan *sse.Event
	closeOnce   sync.Once
	logger      zerolog.Logger
	initialized bool
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(baseURL string, logger zerolog.Logger) *SSETransport {
	return &SSETransport{
		baseURL:   baseURL,
		client:    &http.Client{},
		sseClient: sse.NewClient(baseURL + "/sse"),
		events:    make(chan *sse.Event),
		logger:    logger,
	}
}

// Send sends a request and returns the response
func (t *SSETransport) Send(ctx context.Context, request *protocol.Request) (*protocol.Response, error) {
	t.mu.Lock()
	if !t.initialized {
		t.mu.Unlock()
		t.logger.Debug().Msg("Initializing SSE connection")
		if err := t.initializeSSE(ctx); err != nil {
			t.logger.Error().Err(err).Msg("Failed to initialize SSE")
			return nil, fmt.Errorf("failed to initialize SSE: %w", err)
		}
		t.mu.Lock()
	}
	defer t.mu.Unlock()

	t.logger.Debug().
		Str("method", request.Method).
		Interface("params", request.Params).
		Msg("Sending request")

	// Send request via HTTP POST
	reqBody, err := json.Marshal(request)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to marshal request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	t.logger.Debug().
		Str("url", t.baseURL+"/messages").
		RawJSON("request", reqBody).
		Msg("Sending HTTP POST request")

	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+"/messages", bytes.NewReader(reqBody))
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to send HTTP request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		t.logger.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("Server returned error status")
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Wait for response event with context
	t.logger.Debug().Msg("Waiting for response event")
	select {
	case event := <-t.events:
		if string(event.Event) == "error" {
			t.logger.Error().
				Str("error", string(event.Data)).
				Msg("Received error event")
			return nil, fmt.Errorf("server error: %s", string(event.Data))
		}

		t.logger.Debug().
			Str("event", string(event.Event)).
			RawJSON("data", event.Data).
			Msg("Received response event")

		var response protocol.Response
		if err := json.Unmarshal(event.Data, &response); err != nil {
			t.logger.Error().Err(err).Msg("Failed to parse response")
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil

	case <-ctx.Done():
		t.logger.Debug().Msg("Context cancelled while waiting for response")
		return nil, ctx.Err()
	}
}

// initializeSSE sets up the SSE connection
func (t *SSETransport) initializeSSE(ctx context.Context) error {
	t.logger.Debug().Str("url", t.baseURL+"/sse").Msg("Setting up SSE connection")

	// Create a new context with cancellation for the subscription
	subCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()

		t.logger.Debug().Msg("Subscribing to SSE")
		err := t.sseClient.SubscribeWithContext(subCtx, "", func(msg *sse.Event) {
			t.logger.Debug().
				Str("event", string(msg.Event)).
				RawJSON("data", msg.Data).
				Msg("Received SSE event")

			// Forward events to the events channel
			select {
			case t.events <- msg:
				t.logger.Debug().Msg("Forwarded event to channel")
			case <-subCtx.Done():
				t.logger.Debug().Msg("Context cancelled while forwarding event")
			}
		})

		if err != nil {
			t.logger.Error().Err(err).Msg("SSE subscription failed")
			t.mu.Lock()
			t.initialized = false
			t.mu.Unlock()
			return
		}
	}()

	t.mu.Lock()
	t.initialized = true
	t.mu.Unlock()

	t.logger.Debug().Msg("SSE initialization successful")

	return nil
}

// Close closes the transport
func (t *SSETransport) Close(ctx context.Context) error {
	t.logger.Debug().Msg("Closing transport")
	t.closeOnce.Do(func() {
		t.sseClient.Unsubscribe(t.events)
		close(t.events)
		t.logger.Debug().Msg("Transport closed")
	})
	return nil
}
