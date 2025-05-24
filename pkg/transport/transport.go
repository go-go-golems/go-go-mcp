package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
)

// Transport handles the low-level communication between client and server
type Transport interface {
	// Listen starts accepting requests and forwards them to the handler
	Listen(ctx context.Context, handler RequestHandler) error

	// Send transmits a response back to the client
	Send(ctx context.Context, response *protocol.Response) error

	// Close cleanly shuts down the transport
	Close(ctx context.Context) error

	// Info returns metadata about the transport
	Info() TransportInfo

	// SetSessionStore provides the session store to the transport layer
	SetSessionStore(store session.SessionStore)
}

// TransportInfo provides metadata about the transport
type TransportInfo struct {
	Type         string            // "sse", "stdio", etc.
	RemoteAddr   string            // Remote address if applicable
	Capabilities map[string]bool   // Transport capabilities
	Metadata     map[string]string // Additional transport metadata
}

// IsNotification checks if a request is a notification (no ID)
func IsNotification(req *protocol.Request) bool {
	return req.ID == nil || string(req.ID) == "null" || len(req.ID) == 0
}

// StringToID converts a string to a JSON-RPC ID (json.RawMessage)
func StringToID(s string) json.RawMessage {
	if s == "" {
		return nil
	}
	// Quote the string to make it a valid JSON string
	return json.RawMessage(`"` + s + `"`)
}

// IDToString converts a JSON-RPC ID to a string
func IDToString(id json.RawMessage) string {
	if id == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(id, &s); err != nil {
		return string(id)
	}
	return s
}

// ParseMessage attempts to parse a JSON message as either a single request or a batch request
func ParseMessage(data []byte) (interface{}, error) {
	// First try to parse as a single request
	var singleReq protocol.Request
	if err := json.Unmarshal(data, &singleReq); err == nil {
		// Check if it looks like a valid single request
		if singleReq.JSONRPC == "2.0" && singleReq.Method != "" {
			return &singleReq, nil
		}
	}

	// Try to parse as a batch request
	var batchReq protocol.BatchRequest
	if err := json.Unmarshal(data, &batchReq); err == nil {
		// Validate the batch
		if err := batchReq.Validate(); err == nil {
			return batchReq, nil
		}
	}

	return nil, fmt.Errorf("invalid JSON-RPC message: neither single request nor valid batch")
}

// IsBatchMessage checks if the raw JSON data represents a batch request (starts with '[')
func IsBatchMessage(data []byte) bool {
	// Trim whitespace and check if it starts with '['
	trimmed := bytes.TrimSpace(data)
	return len(trimmed) > 0 && trimmed[0] == '['
}

// RequestHandler processes incoming requests and notifications
type RequestHandler interface {
	// HandleRequest processes a request and returns a response
	HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error)

	// HandleNotification processes a notification (no response expected)
	HandleNotification(ctx context.Context, notif *protocol.Notification) error

	// HandleBatchRequest processes a batch of requests and returns batch responses
	HandleBatchRequest(ctx context.Context, batch protocol.BatchRequest) (protocol.BatchResponse, error)
}
