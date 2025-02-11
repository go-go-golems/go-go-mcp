package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog"
)

// StdioTransport implements Transport using standard input/output
type StdioTransport struct {
	mu                  sync.Mutex
	scanner             *bufio.Scanner
	writer              *json.Encoder
	cmd                 *exec.Cmd
	logger              zerolog.Logger
	notificationHandler func(*protocol.Response)
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(logger zerolog.Logger) *StdioTransport {
	scanner := bufio.NewScanner(os.Stdin)
	// Set 1MB buffer size to avoid "token too long" errors
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, len(buf))

	return &StdioTransport{
		scanner: scanner,
		writer:  json.NewEncoder(os.Stdout),
		logger:  logger,
	}
}

// SetNotificationHandler sets the handler for notifications
func (t *StdioTransport) SetNotificationHandler(handler func(*protocol.Response)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.notificationHandler = handler
}

var _ Transport = &StdioTransport{}

// NewCommandStdioTransport creates a new stdio transport that launches a command
func NewCommandStdioTransport(logger zerolog.Logger, command string, args ...string) (*StdioTransport, error) {
	cmd := exec.Command(command, args...)

	logger.Debug().
		Str("command", command).
		Strs("args", args).
		Msg("Creating command stdio transport")

	// Set up process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

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

	logger.Debug().
		Int("pid", cmd.Process.Pid).
		Msg("Command started successfully in new process group")

	return &StdioTransport{
		scanner: bufio.NewScanner(stdout),
		writer:  json.NewEncoder(stdin),
		cmd:     cmd,
		logger:  logger,
	}, nil
}

// Send sends a request and returns the response
func (t *StdioTransport) Send(ctx context.Context, request *protocol.Request) (*protocol.Response, error) {
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

	// If this is a notification, don't wait for response
	if len(request.ID) == 0 || string(request.ID) == "null" || strings.HasPrefix(request.Method, "notifications/") {
		t.logger.Debug().Msg("Request is a notification, not waiting for response")
		return nil, nil
	}

	t.logger.Debug().Msg("Waiting for response")

	// Create a channel for the response
	responseCh := make(chan struct {
		response *protocol.Response
		err      error
	}, 1)

	// Read response in a goroutine
	go func() {
		for t.scanner.Scan() {
			data := t.scanner.Bytes()
			t.logger.Debug().RawJSON("data", data).Msg("Received data")

			var response protocol.Response
			if err := json.Unmarshal(data, &response); err != nil {
				t.logger.Error().Err(err).Msg("Failed to parse response")
				continue
			}

			// If this is a notification, handle it and continue scanning
			if len(response.ID) == 0 || string(response.ID) == "null" {
				t.mu.Lock()
				if t.notificationHandler != nil {
					t.notificationHandler(&response)
				}
				t.mu.Unlock()
				continue
			}

			// This is a response, send it to the channel
			responseCh <- struct {
				response *protocol.Response
				err      error
			}{&response, nil}
			return
		}

		// Handle scanner errors
		if err := t.scanner.Err(); err != nil {
			responseCh <- struct {
				response *protocol.Response
				err      error
			}{nil, fmt.Errorf("failed to read response: %w", err)}
		} else {
			responseCh <- struct {
				response *protocol.Response
				err      error
			}{nil, io.EOF}
		}
	}()

	// Wait for either response or context cancellation
	select {
	case result := <-responseCh:
		return result.response, result.err
	case <-ctx.Done():
		t.logger.Debug().Msg("Context cancelled while waiting for response")
		return nil, ctx.Err()
	}
}

// Close closes the transport
func (t *StdioTransport) Close(ctx context.Context) error {
	t.logger.Debug().Msg("Closing transport")
	if t.cmd != nil {
		t.logger.Debug().
			Int("pid", t.cmd.Process.Pid).
			Msg("Attempting to send interrupt signal to process group")

		// Send interrupt signal to the process group
		pgid, err := syscall.Getpgid(t.cmd.Process.Pid)
		if err != nil {
			t.logger.Warn().
				Err(err).
				Int("pid", t.cmd.Process.Pid).
				Msg("Failed to get process group ID, falling back to direct process signal")

			// Fall back to sending signal directly to process
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
		} else {
			// Send interrupt to the process group
			if err := syscall.Kill(-pgid, syscall.SIGINT); err != nil {
				t.logger.Warn().
					Err(err).
					Int("pgid", pgid).
					Msg("Failed to send interrupt signal to process group, falling back to Kill")

				// If interrupt fails, try to kill the process group
				if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
					t.logger.Error().
						Err(err).
						Int("pgid", pgid).
						Msg("Failed to kill process group")
					return fmt.Errorf("failed to kill process group: %w", err)
				}
			}
		}

		// Create a channel to receive the wait result
		waitCh := make(chan error, 1)
		go func() {
			waitCh <- t.cmd.Wait()
		}()

		// Wait for either the process to exit or context to be cancelled
		select {
		case err := <-waitCh:
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
		case <-ctx.Done():
			t.logger.Debug().Msg("Context cancelled while waiting for command to exit")
			return ctx.Err()
		}
	}
	t.logger.Debug().Msg("No command to close")
	return nil
}
