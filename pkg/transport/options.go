package transport

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// TransportOptions contains common options for all transports
type TransportOptions struct {
	// Common options
	MaxMessageSize int64
	Logger         zerolog.Logger

	// Transport specific options
	SSE   *SSEOptions
	Stdio *StdioOptions
}

// SSEOptions contains SSE-specific transport options
type SSEOptions struct {
	// HTTP server configuration
	Addr      string
	TLSConfig *tls.Config

	// Middleware
	Middleware []func(http.Handler) http.Handler

	// Router configuration
	Router     *mux.Router // Optional: existing router to use
	PathPrefix string      // Optional: prefix for all SSE endpoints
}

// StdioOptions contains stdio-specific transport options
type StdioOptions struct {
	// Buffer sizes
	ReadBufferSize  int
	WriteBufferSize int

	// Process management
	Command     string
	Args        []string
	WorkingDir  string
	Environment map[string]string

	// Signal handling
	SignalHandlers map[os.Signal]func()
}

// TransportOption is a function that modifies TransportOptions
type TransportOption func(*TransportOptions)

// Common option constructors
func WithLogger(logger zerolog.Logger) TransportOption {
	return func(o *TransportOptions) {
		o.Logger = logger
	}
}

func WithMaxMessageSize(size int64) TransportOption {
	return func(o *TransportOptions) {
		o.MaxMessageSize = size
	}
}

// SSE-specific options
func WithSSEOptions(opts SSEOptions) TransportOption {
	return func(o *TransportOptions) {
		o.SSE = &opts
	}
}

// Stdio-specific options
func WithStdioOptions(opts StdioOptions) TransportOption {
	return func(o *TransportOptions) {
		o.Stdio = &opts
	}
}
