package transport

import (
	"context"
	"encoding/json"
	"fmt"
)

// Transport handles the low-level communication between client and server
type Transport interface {
	// Listen starts accepting requests and forwards them to the handler
	Listen(ctx context.Context, handler RequestHandler) error

	// Send transmits a response back to the client
	Send(ctx context.Context, response *Response) error

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

// RequestHandler processes incoming requests and notifications
type RequestHandler interface {
	// HandleRequest processes a request and returns a response
	HandleRequest(ctx context.Context, req *Request) (*Response, error)

	// HandleNotification processes a notification (no response expected)
	HandleNotification(ctx context.Context, notif *Notification) error
}

// Request represents an incoming JSON-RPC request
type Request struct {
	ID      string
	Method  string
	Params  json.RawMessage
	Headers map[string]string
}

// Response represents an outgoing JSON-RPC response
type Response struct {
	ID      string
	Result  json.RawMessage
	Error   *ResponseError
	Headers map[string]string
}

// Notification represents an incoming notification
type Notification struct {
	Method  string
	Params  json.RawMessage
	Headers map[string]string
}

// ResponseError represents a JSON-RPC error response
type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (r *ResponseError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, data: %s", r.Code, r.Message, string(r.Data))
}
