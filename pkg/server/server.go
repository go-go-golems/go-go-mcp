package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/server/transports/stdio"
	"github.com/go-go-golems/go-go-mcp/pkg/services"
	"github.com/go-go-golems/go-go-mcp/pkg/services/defaults"
	"github.com/rs/zerolog"
)

// Transport represents a server transport mechanism
type Transport interface {
	// Start starts the transport with the given context
	Start(ctx context.Context) error
	// Stop gracefully stops the transport with the given context
	Stop(ctx context.Context) error
}

// Server represents an MCP server that can use different transports
type Server struct {
	mu                sync.Mutex
	logger            zerolog.Logger
	registry          *pkg.ProviderRegistry
	promptService     services.PromptService
	resourceService   services.ResourceService
	toolService       services.ToolService
	initializeService services.InitializeService
	serverName        string
	serverVersion     string
	transport         Transport
}

// NewServer creates a new server instance
func NewServer(logger zerolog.Logger, options ...ServerOption) *Server {
	registry := pkg.NewProviderRegistry()
	s := &Server{
		logger:            logger,
		registry:          registry,
		serverName:        "go-mcp-server",
		serverVersion:     "1.0.0",
		promptService:     defaults.NewPromptService(registry, logger),
		resourceService:   defaults.NewResourceService(registry, logger),
		toolService:       defaults.NewToolService(registry, logger),
		initializeService: defaults.NewInitializeService("go-mcp-server", "1.0.0"),
	}

	for _, opt := range options {
		opt(s)
	}
	return s
}

// GetRegistry returns the server's provider registry
func (s *Server) GetRegistry() *pkg.ProviderRegistry {
	return s.registry
}

// StartStdio starts the server with stdio transport
func (s *Server) StartStdio(ctx context.Context) error {
	s.mu.Lock()
	s.logger.Debug().Msg("Creating stdio transport")
	// Create a new logger for the stdio server that preserves the log level and other settings
	stdioLogger := s.logger.With().Logger()
	stdioServer := stdio.NewServer(stdioLogger, s.promptService, s.resourceService, s.toolService, s.initializeService)
	s.transport = stdioServer
	s.mu.Unlock()

	s.logger.Debug().Msg("Starting stdio transport")
	err := stdioServer.Start(ctx)
	if err != nil {
		s.logger.Debug().
			Err(err).
			Msg("Stdio transport stopped with error")
		return err
	}
	s.logger.Debug().Msg("Stdio transport stopped cleanly")
	return nil
}

// StartSSE starts the server with SSE transport on the specified port
func (s *Server) StartSSE(ctx context.Context, port int) error {
	s.mu.Lock()
	s.logger.Debug().Int("port", port).Msg("Creating SSE transport")
	sseServer := NewSSEServer(s.logger, s.registry, port)
	s.transport = sseServer
	s.mu.Unlock()

	s.logger.Debug().Int("port", port).Msg("Starting SSE transport")
	err := sseServer.Start(ctx)
	if err != nil {
		s.logger.Debug().
			Err(err).
			Msg("SSE transport stopped with error")
		return err
	}
	s.logger.Debug().Msg("SSE transport stopped cleanly")
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.transport == nil {
		s.logger.Debug().Msg("No transport to stop")
		return nil
	}

	s.logger.Info().Msg("Stopping server transport")
	err := s.transport.Stop(ctx)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("Error stopping transport")
		return fmt.Errorf("error stopping transport: %w", err)
	}

	s.logger.Debug().Msg("Transport stopped successfully")
	return nil
}
