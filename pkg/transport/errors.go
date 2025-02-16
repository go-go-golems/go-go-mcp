package transport

import (
	"fmt"

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
