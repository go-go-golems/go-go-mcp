package transport

import (
	"context"
	"encoding/json"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
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

// RequestHandler processes incoming requests and notifications
type RequestHandler interface {
	// HandleRequest processes a request and returns a response
	HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error)

	// HandleNotification processes a notification (no response expected)
	HandleNotification(ctx context.Context, notif *protocol.Notification) error
}
