package server

import (
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/server/transports/stdio"
	"github.com/go-go-golems/go-go-mcp/pkg/services"
	"github.com/go-go-golems/go-go-mcp/pkg/services/defaults"
	"github.com/rs/zerolog"
)

// Transport represents a server transport mechanism
type Transport interface {
	// Start starts the transport
	Start() error
	// Stop gracefully stops the transport
	Stop() error
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
func (s *Server) StartStdio() error {
	s.mu.Lock()
	stdioServer := stdio.NewServer(s.logger, s.promptService, s.resourceService, s.toolService, s.initializeService)
	s.transport = stdioServer
	s.mu.Unlock()
	return stdioServer.Start()
}

// StartSSE starts the server with SSE transport on the specified port
func (s *Server) StartSSE(port int) error {
	s.mu.Lock()
	sseServer := NewSSEServer(s.logger, s.registry, port)
	s.transport = sseServer
	s.mu.Unlock()
	return sseServer.Start()
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.transport == nil {
		return nil
	}

	s.logger.Info().Msg("Stopping server")
	return s.transport.Stop()
}
