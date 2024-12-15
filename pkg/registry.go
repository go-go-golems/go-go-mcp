package pkg

// Error represents a protocol error with a code and message
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new Error with the given message and code
func NewError(message string, code int) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Common errors
var (
	ErrPromptNotFound   = NewError("prompt not found", -32000)
	ErrResourceNotFound = NewError("resource not found", -32001)
	ErrToolNotFound     = NewError("tool not found", -32002)
	ErrNotImplemented   = NewError("not implemented", -32003)
)
