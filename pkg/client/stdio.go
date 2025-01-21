package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// StdioTransport implements Transport using standard input/output
type StdioTransport struct {
	mu      sync.Mutex
	scanner *bufio.Scanner
	writer  *json.Encoder
	cmd     *exec.Cmd
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		scanner: bufio.NewScanner(os.Stdin),
		writer:  json.NewEncoder(os.Stdout),
	}
}

// NewCommandStdioTransport creates a new stdio transport that launches a command
func NewCommandStdioTransport(command string, args ...string) (*StdioTransport, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	return &StdioTransport{
		scanner: bufio.NewScanner(stdout),
		writer:  json.NewEncoder(stdin),
		cmd:     cmd,
	}, nil
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
	if t.cmd != nil {
		if err := t.cmd.Process.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("failed to send interrupt signal: %w", err)
		}
		return t.cmd.Wait()
	}
	return nil // Nothing to close for stdio
}
