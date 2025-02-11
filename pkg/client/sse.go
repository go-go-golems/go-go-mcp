package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/r3labs/sse/v2"
	"github.com/rs/zerolog"
)

// SSETransport implements Transport using Server-Sent Events
type SSETransport struct {
	mu                  sync.Mutex
	baseURL             string
	client              *http.Client
	sseClient           *sse.Client
	events              chan *sse.Event
	responses           chan *sse.Event
	notifications       chan *sse.Event
	closeOnce           sync.Once
	logger              zerolog.Logger
	initialized         bool
	sessionID           string
	endpoint            string
	notificationHandler func(*protocol.Response)
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(baseURL string, logger zerolog.Logger) *SSETransport {
	return &SSETransport{
		baseURL:       baseURL,
		client:        &http.Client{},
		sseClient:     sse.NewClient(baseURL + "/sse"),
		events:        make(chan *sse.Event),
		responses:     make(chan *sse.Event),
		notifications: make(chan *sse.Event),
		logger:        logger,
	}
}

// SetNotificationHandler sets the handler for notifications
func (t *SSETransport) SetNotificationHandler(handler func(*protocol.Response)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.notificationHandler = handler
}

// isNotification checks if an event is a notification
func isNotification(event *sse.Event) bool {
	var response protocol.Response
	if err := json.Unmarshal(event.Data, &response); err != nil {
		return false
	}
	return len(response.ID) == 0 || string(response.ID) == "null"
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
		Str("url", t.endpoint).
		RawJSON("request", reqBody).
		Msg("Sending HTTP POST request")

	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, "POST", t.endpoint, bytes.NewReader(reqBody))
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

	// If this is a notification, don't wait for response
	if len(request.ID) == 0 || string(request.ID) == "null" || strings.HasPrefix(request.Method, "notifications/") {
		t.logger.Debug().Msg("Request is a notification, not waiting for response")
		return nil, nil
	}

	// Wait for response event with context
	t.logger.Debug().Msg("Waiting for response event")
	select {
	case event := <-t.responses:
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

	// Channel to wait for endpoint event
	endpointCh := make(chan string, 1)

	// Start notification handler goroutine
	go func() {
		for {
			select {
			case event := <-t.notifications:
				if t.notificationHandler != nil {
					var response protocol.Response
					if err := json.Unmarshal(event.Data, &response); err != nil {
						t.logger.Error().Err(err).Msg("Failed to parse notification")
						continue
					}
					t.notificationHandler(&response)
				}
			case <-subCtx.Done():
				return
			}
		}
	}()

	go func() {
		defer cancel()

		t.logger.Debug().Msg("Subscribing to SSE")
		err := t.sseClient.SubscribeWithContext(subCtx, "", func(msg *sse.Event) {
			t.logger.Debug().
				Str("event", string(msg.Event)).
				Str("data", string(msg.Data)).
				Str("retry", string(msg.Retry)).
				Str("id", string(msg.ID)).
				Str("comment", string(msg.Comment)).
				Msg("Received SSE event")

			// Handle endpoint event
			if string(msg.Event) == "endpoint" {
				endpoint := string(msg.Data)
				t.logger.Debug().
					Str("endpoint", endpoint).
					Msg("Received endpoint event")

				// Parse endpoint URL and extract session ID
				if strings.Contains(endpoint, "sessionId=") {
					t.mu.Lock()
					t.endpoint = t.baseURL + endpoint
					t.sessionID = strings.Split(strings.Split(endpoint, "sessionId=")[1], "&")[0]
					t.mu.Unlock()
					endpointCh <- endpoint
					return
				}
			}

			// Route event to appropriate channel
			if isNotification(msg) {
				select {
				case t.notifications <- msg:
					t.logger.Debug().Msg("Forwarded notification event")
				case <-subCtx.Done():
					t.logger.Debug().Msg("Context cancelled while forwarding notification")
				}
			} else {
				select {
				case t.responses <- msg:
					t.logger.Debug().Msg("Forwarded response event")
				case <-subCtx.Done():
					t.logger.Debug().Msg("Context cancelled while forwarding response")
				}
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

	// Wait for endpoint event or context cancellation
	select {
	case <-endpointCh:
		t.mu.Lock()
		t.initialized = true
		t.mu.Unlock()
		t.logger.Debug().
			Str("endpoint", t.endpoint).
			Str("sessionId", t.sessionID).
			Msg("SSE initialization successful")
		return nil
	case <-ctx.Done():
		t.logger.Error().Msg("Context cancelled while waiting for endpoint event")
		return ctx.Err()
	}
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
