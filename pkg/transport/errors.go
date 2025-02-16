package transport

import "fmt"

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
func NewParseError(msg string) *ResponseError {
	return &ResponseError{
		Code:    ErrCodeParse,
		Message: fmt.Sprintf("Parse error: %s", msg),
	}
}

func NewInvalidRequestError(msg string) *ResponseError {
	return &ResponseError{
		Code:    ErrCodeInvalidRequest,
		Message: fmt.Sprintf("Invalid request: %s", msg),
	}
}

func NewMethodNotFoundError(msg string) *ResponseError {
	return &ResponseError{
		Code:    ErrCodeMethodNotFound,
		Message: fmt.Sprintf("Method not found: %s", msg),
	}
}

func NewInvalidParamsError(msg string) *ResponseError {
	return &ResponseError{
		Code:    ErrCodeInvalidParams,
		Message: fmt.Sprintf("Invalid params: %s", msg),
	}
}

func NewInternalError(msg string) *ResponseError {
	return &ResponseError{
		Code:    ErrCodeInternal,
		Message: fmt.Sprintf("Internal error: %s", msg),
	}
}

func NewTransportError(msg string) *ResponseError {
	return &ResponseError{
		Code:    ErrCodeTransport,
		Message: fmt.Sprintf("Transport error: %s", msg),
	}
}

func NewTimeoutError(msg string) *ResponseError {
	return &ResponseError{
		Code:    ErrCodeTimeout,
		Message: fmt.Sprintf("Timeout error: %s", msg),
	}
}
