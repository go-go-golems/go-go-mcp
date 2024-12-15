package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/go-go-golems/go-mcp/pkg/protocol"
	"github.com/r3labs/sse/v2"
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
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(baseURL string) *SSETransport {
	return &SSETransport{
		baseURL:   baseURL,
		client:    &http.Client{},
		sseClient: sse.NewClient(baseURL + "/sse"),
		events:    make(chan *sse.Event),
		closeChan: make(chan struct{}),
	}
}

// Send sends a request and returns the response
func (t *SSETransport) Send(request *protocol.Request) (*protocol.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Initialize SSE connection if not already done
	if !t.initialized {
		if err := t.initializeSSE(); err != nil {
			return nil, fmt.Errorf("failed to initialize SSE: %w", err)
		}
	}

	// Send request via HTTP POST
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := t.client.Post(t.baseURL+"/messages", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Wait for response event
	select {
	case event := <-t.events:
		if string(event.Event) == "error" {
			return nil, fmt.Errorf("server error: %s", string(event.Data))
		}

		var response protocol.Response
		if err := json.Unmarshal(event.Data, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil

	case <-t.closeChan:
		return nil, fmt.Errorf("transport closed")
	}
}

// initializeSSE sets up the SSE connection
func (t *SSETransport) initializeSSE() error {
	// Subscribe to SSE events
	if err := t.sseClient.SubscribeRaw(func(msg *sse.Event) {
		// Handle session ID event
		if string(msg.Event) == "session" {
			t.sessionID = string(msg.Data)
			return
		}

		// Forward other events to the events channel
		select {
		case t.events <- msg:
		case <-t.closeChan:
		}
	}); err != nil {
		return fmt.Errorf("failed to subscribe to SSE: %w", err)
	}

	t.initialized = true
	return nil
}

// Close closes the transport
func (t *SSETransport) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeChan)
		t.sseClient.Unsubscribe(t.events)
	})
	return nil
}
