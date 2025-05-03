package transport

import (
	"fmt"
	"net/http"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// Common error codes
const (
	ErrCodeParse          = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternal       = -32603
	ErrCodeTransport      = -32500
	ErrCodeTimeout        = -32501
)

// Error constructors
func NewParseError(msg string) *protocol.Error {
	return &protocol.Error{
		Code:    ErrCodeParse,
		Message: fmt.Sprintf("Parse error: %s", msg),
	}
}

func NewInvalidRequestError(msg string) *protocol.Error {
	return &protocol.Error{
		Code:    ErrCodeInvalidRequest,
		Message: fmt.Sprintf("Invalid request: %s", msg),
	}
}

func NewMethodNotFoundError(msg string) *protocol.Error {
	return &protocol.Error{
		Code:    ErrCodeMethodNotFound,
		Message: fmt.Sprintf("Method not found: %s", msg),
	}
}

func NewInvalidParamsError(msg string) *protocol.Error {
	return &protocol.Error{
		Code:    ErrCodeInvalidParams,
		Message: fmt.Sprintf("Invalid params: %s", msg),
	}
}

func NewInternalError(msg string) *protocol.Error {
	return &protocol.Error{
		Code:    ErrCodeInternal,
		Message: fmt.Sprintf("Internal error: %s", msg),
	}
}

func NewTransportError(msg string) *protocol.Error {
	return &protocol.Error{
		Code:    ErrCodeTransport,
		Message: fmt.Sprintf("Transport error: %s", msg),
	}
}

func NewTimeoutError(msg string) *protocol.Error {
	return &protocol.Error{
		Code:    ErrCodeTimeout,
		Message: fmt.Sprintf("Timeout error: %s", msg),
	}
}

// ProcessError converts a Go error into a standard JSON-RPC Error object.
// It tries to map known transport errors to specific codes, otherwise defaults to Internal Error.
func ProcessError(err error) *protocol.Error {
	// Check if the error is already a *protocol.Error
	if jsonRPCErr, ok := err.(*protocol.Error); ok {
		return jsonRPCErr
	}

	// TODO(manuel): Add more specific error type checks if needed, e.g., os.IsNotExist

	// Default to internal error
	return NewInternalError(err.Error())
}

// ErrorToHTTPStatus maps JSON-RPC error codes to appropriate HTTP status codes.
func ErrorToHTTPStatus(code int) int {
	switch code {
	case ErrCodeParse:
		return http.StatusBadRequest // 400
	case ErrCodeInvalidRequest:
		return http.StatusBadRequest // 400
	case ErrCodeMethodNotFound:
		return http.StatusNotFound // 404
	case ErrCodeInvalidParams:
		return http.StatusBadRequest // 400
	case ErrCodeInternal:
		return http.StatusInternalServerError // 500
	case ErrCodeTransport:
		return http.StatusInternalServerError // 500 (or specific transport issue code?)
	case ErrCodeTimeout:
		return http.StatusGatewayTimeout // 504
	default:
		// Handle custom server errors (-32000 to -32099)
		if code >= -32099 && code <= -32000 {
			// You might map specific custom codes to HTTP statuses here
			return http.StatusInternalServerError // Default for custom errors
		}
		// For any other codes, default to Internal Server Error
		return http.StatusInternalServerError // 500
	}
}
