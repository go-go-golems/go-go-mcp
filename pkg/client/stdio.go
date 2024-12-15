package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/go-go-golems/go-mcp/pkg/protocol"
)

// StdioTransport implements Transport using standard input/output
type StdioTransport struct {
	mu      sync.Mutex
	scanner *bufio.Scanner
	writer  *json.Encoder
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		scanner: bufio.NewScanner(os.Stdin),
		writer:  json.NewEncoder(os.Stdout),
	}
}

// Send sends a request and returns the response
func (t *StdioTransport) Send(request *protocol.Request) (*protocol.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Write request
	if err := t.writer.Encode(request); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	if !t.scanner.Scan() {
		if err := t.scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		return nil, io.EOF
	}

	var response protocol.Response
	if err := json.Unmarshal(t.scanner.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	return nil // Nothing to close for stdio
}
