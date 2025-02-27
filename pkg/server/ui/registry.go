package ui

import (
	"sync"
	"time"
)

// UIActionResponse represents the response from a UI action
type UIActionResponse struct {
	Action        string
	ComponentID   string
	Data          map[string]interface{}
	Error         error
	Timestamp     time.Time
	RelatedEvents []UIActionEvent // Additional events related to this response
}

// UIActionEvent represents a single UI action event
type UIActionEvent struct {
	Action      string
	ComponentID string
	Data        map[string]interface{}
	Timestamp   time.Time
}

// WaitRegistry tracks pending UI update requests waiting for user actions
type WaitRegistry struct {
	pending map[string]chan UIActionResponse
	mu      sync.RWMutex
	timeout time.Duration
}

// NewWaitRegistry creates a new registry with the specified timeout
func NewWaitRegistry(timeout time.Duration) *WaitRegistry {
	return &WaitRegistry{
		pending: make(map[string]chan UIActionResponse),
		timeout: timeout,
	}
}

// Register adds a new request to the registry and returns a channel for the response
func (wr *WaitRegistry) Register(requestID string) chan UIActionResponse {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	responseChan := make(chan UIActionResponse, 1)
	wr.pending[requestID] = responseChan

	return responseChan
}

// Resolve completes a pending request with the given action response
func (wr *WaitRegistry) Resolve(requestID string, response UIActionResponse) bool {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	if ch, exists := wr.pending[requestID]; exists {
		// Send the response to the waiting handler
		ch <- response

		// Only remove the request from the registry if it's not a click event
		// This allows click events to be potentially followed by submit events
		if response.Action != "clicked" {
			delete(wr.pending, requestID)
		}

		return true
	}
	return false
}

// Cleanup removes a request from the registry
func (wr *WaitRegistry) Cleanup(requestID string) {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	if ch, exists := wr.pending[requestID]; exists {
		close(ch)
		delete(wr.pending, requestID)
	}
}

// CleanupStale removes requests that have been in the registry longer than the timeout
// This is a placeholder implementation that would need to be enhanced with timestamp tracking
func (wr *WaitRegistry) CleanupStale() int {
	// This would require tracking timestamps for each request
	// For now, we'll just return 0 as we're not implementing the full functionality yet
	return 0
}
