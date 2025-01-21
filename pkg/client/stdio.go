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
	"github.com/rs/zerolog"
)

// StdioTransport implements Transport using standard input/output
type StdioTransport struct {
	mu      sync.Mutex
	scanner *bufio.Scanner
	writer  *json.Encoder
	cmd     *exec.Cmd
	logger  zerolog.Logger
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		scanner: bufio.NewScanner(os.Stdin),
		writer:  json.NewEncoder(os.Stdout),
		logger:  zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger(),
	}
}

// NewCommandStdioTransport creates a new stdio transport that launches a command
func NewCommandStdioTransport(command string, args ...string) (*StdioTransport, error) {
	cmd := exec.Command(command, args...)

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	logger.Debug().
		Str("command", command).
		Strs("args", args).
		Msg("Creating command stdio transport")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create stdin pipe")
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create stdout pipe")
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Forward stderr to client's stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logger.Error().Err(err).Msg("Failed to start command")
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	logger.Debug().Msg("Command started successfully")

	return &StdioTransport{
		scanner: bufio.NewScanner(stdout),
		writer:  json.NewEncoder(stdin),
		cmd:     cmd,
		logger:  logger,
	}, nil
}

// Send sends a request and returns the response
func (t *StdioTransport) Send(request *protocol.Request) (*protocol.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.logger.Debug().
		Str("method", request.Method).
		Interface("params", request.Params).
		Msg("Sending request")

	// Write request
	if err := t.writer.Encode(request); err != nil {
		t.logger.Error().Err(err).Msg("Failed to write request")
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	t.logger.Debug().Msg("Waiting for response")

	// Read response
	if !t.scanner.Scan() {
		if err := t.scanner.Err(); err != nil {
			t.logger.Error().Err(err).Msg("Failed to read response")
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		t.logger.Debug().Msg("EOF while reading response")
		return nil, io.EOF
	}

	t.logger.Debug().
		RawJSON("response", t.scanner.Bytes()).
		Msg("Received response")

	var response protocol.Response
	if err := json.Unmarshal(t.scanner.Bytes(), &response); err != nil {
		t.logger.Error().Err(err).Msg("Failed to parse response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.logger.Debug().Msg("Closing transport")
	if t.cmd != nil {
		t.logger.Debug().
			Int("pid", t.cmd.Process.Pid).
			Msg("Attempting to send interrupt signal to command")

		// First try to send an interrupt signal
		if err := t.cmd.Process.Signal(os.Interrupt); err != nil {
			t.logger.Warn().
				Err(err).
				Int("pid", t.cmd.Process.Pid).
				Msg("Failed to send interrupt signal, falling back to Kill")

			// If interrupt fails, try to kill the process
			if err := t.cmd.Process.Kill(); err != nil {
				t.logger.Error().
					Err(err).
					Int("pid", t.cmd.Process.Pid).
					Msg("Failed to kill process")
				return fmt.Errorf("failed to kill process: %w", err)
			}
		}

		// Wait for the process to exit
		t.logger.Debug().Msg("Waiting for command to exit")
		err := t.cmd.Wait()
		if err != nil {
			// Check if it's an expected exit error (like signal kill)
			if exitErr, ok := err.(*exec.ExitError); ok {
				t.logger.Debug().
					Err(err).
					Int("exit_code", exitErr.ExitCode()).
					Msg("Command exited with error (expected for signal termination)")
				return nil
			}
			t.logger.Error().
				Err(err).
				Msg("Command exited with unexpected error")
			return err
		}

		t.logger.Debug().Msg("Command exited successfully")
		return nil
	}
	t.logger.Debug().Msg("No command to close")
	return nil
}
