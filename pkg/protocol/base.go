package protocol

import (
	"encoding/json"
	"fmt"
)

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"` // Can be a string or an int
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"` // Can be a string or an int
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// BatchRequest represents an array of JSON-RPC 2.0 requests for batch processing
type BatchRequest []Request

// BatchResponse represents an array of JSON-RPC 2.0 responses for batch processing
type BatchResponse []Response

// Validate checks that the batch request is valid
func (br BatchRequest) Validate() error {
	if len(br) == 0 {
		return fmt.Errorf("batch request cannot be empty")
	}
	// Additional validation: ensure all requests have the correct JSONRPC version
	for i, req := range br {
		if req.JSONRPC != "2.0" {
			return fmt.Errorf("request %d has invalid JSONRPC version: %s", i, req.JSONRPC)
		}
	}
	return nil
}

// GetRequestByID finds a request in the batch by its ID
func (br BatchRequest) GetRequestByID(id json.RawMessage) *Request {
	idStr := string(id)
	for i := range br {
		if string(br[i].ID) == idStr {
			return &br[i]
		}
	}
	return nil
}

// Validate checks that the batch response is valid
func (br BatchResponse) Validate() error {
	// Batch responses can be empty if all requests were notifications
	for i, resp := range br {
		if resp.JSONRPC != "2.0" {
			return fmt.Errorf("response %d has invalid JSONRPC version: %s", i, resp.JSONRPC)
		}
	}
	return nil
}

// Error represents a JSON-RPC 2.0 error.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s, data: %s", e.Code, e.Message, e.Data)
}

// Notification represents a JSON-RPC 2.0 notification.
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RequestMetadata represents common metadata that can be included in requests
type RequestMetadata struct {
	ProgressToken string `json:"progressToken,omitempty"`
}
