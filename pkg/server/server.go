package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/rs/zerolog"
)

// Server represents an MCP server that can use different transports
type Server struct {
	mu               sync.Mutex
	logger           zerolog.Logger
	transport        transport.Transport
	promptProvider   pkg.PromptProvider
	resourceProvider pkg.ResourceProvider
	toolProvider     pkg.ToolProvider
	sessionStore     session.SessionStore
	handler          *RequestHandler

	serverName    string
	serverVersion string
}

// NewServer creates a new server instance
func NewServer(logger zerolog.Logger, t transport.Transport, opts ...ServerOption) *Server {
	s := &Server{
		logger:       logger,
		transport:    t,
		sessionStore: session.NewInMemorySessionStore(),
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Create request handler
	s.handler = NewRequestHandler(s)

	return s
}

// Start begins the server with the configured transport
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info().
		Str("transport", s.transport.Info().Type).
		Interface("capabilities", s.transport.Info().Capabilities).
		Msg("Starting MCP server")

	// Pass the session store to the transport
	if s.sessionStore != nil {
		s.transport.SetSessionStore(s.sessionStore)
	}

	return s.transport.Listen(ctx, s.handler)
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info().Msg("Stopping server")

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.transport == nil {
		s.logger.Debug().Msg("No transport to stop")
		return nil
	}

	s.logger.Info().Msg("Stopping server transport")
	err := s.transport.Close(ctx)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("Error stopping transport")
		return fmt.Errorf("error stopping transport: %w", err)
	}

	s.logger.Debug().Msg("Transport stopped successfully")
	return nil
}
