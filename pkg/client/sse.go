package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	sessionID   string
	closeOnce   sync.Once
	closeChan   chan struct{}
	initialized bool
	initChan    chan struct{} // Channel to signal initialization completion
	logger      zerolog.Logger
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(baseURL string) *SSETransport {
	return &SSETransport{
		baseURL:   baseURL,
		client:    &http.Client{},
		sseClient: sse.NewClient(baseURL + "/sse"),
		events:    make(chan *sse.Event),
		closeChan: make(chan struct{}),
		initChan:  make(chan struct{}),
		logger:    zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger(),
	}
}

// Send sends a request and returns the response
func (t *SSETransport) Send(request *protocol.Request) (*protocol.Response, error) {
	t.mu.Lock()
	if !t.initialized {
		t.mu.Unlock()
		t.logger.Debug().Msg("Initializing SSE connection")
		if err := t.initializeSSE(); err != nil {
			t.logger.Error().Err(err).Msg("Failed to initialize SSE")
			return nil, fmt.Errorf("failed to initialize SSE: %w", err)
		}
		// Wait for initialization to complete
		select {
		case <-t.initChan:
			t.logger.Debug().Msg("SSE initialization completed")
		case <-t.closeChan:
			return nil, fmt.Errorf("transport closed during initialization")
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

	resp, err := t.client.Post(t.baseURL+"/messages", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to send HTTP request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.logger.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("Server returned error status")
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Wait for response event
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

	case <-t.closeChan:
		t.logger.Debug().Msg("Transport closed while waiting for response")
		return nil, fmt.Errorf("transport closed")
	}
}

// initializeSSE sets up the SSE connection
func (t *SSETransport) initializeSSE() error {
	t.logger.Debug().Str("url", t.baseURL+"/sse").Msg("Setting up SSE connection")

	// Start SSE subscription in a goroutine
	go func() {
		defer close(t.initChan)

		err := t.sseClient.SubscribeRaw(func(msg *sse.Event) {
			// Handle session ID event
			if string(msg.Event) == "session" {
				t.mu.Lock()
				t.sessionID = string(msg.Data)
				t.initialized = true
				t.mu.Unlock()
				t.logger.Debug().Str("sessionID", t.sessionID).Msg("Received session ID")
				return
			}

			t.logger.Debug().
				Str("event", string(msg.Event)).
				RawJSON("data", msg.Data).
				Msg("Received SSE event")

			// Forward other events to the events channel
			select {
			case t.events <- msg:
				t.logger.Debug().Msg("Forwarded event to channel")
			case <-t.closeChan:
				t.logger.Debug().Msg("Transport closed while forwarding event")
			}
		})

		if err != nil {
			t.logger.Error().Err(err).Msg("SSE subscription failed")
			// Signal initialization failure
			t.mu.Lock()
			t.initialized = false
			t.mu.Unlock()
		}
	}()

	return nil
}

// Close closes the transport
func (t *SSETransport) Close() error {
	t.logger.Debug().Msg("Closing transport")
	t.closeOnce.Do(func() {
		close(t.closeChan)
		t.sseClient.Unsubscribe(t.events)
		t.logger.Debug().Msg("Transport closed")
	})
	return nil
}
