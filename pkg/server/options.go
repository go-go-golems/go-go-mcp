package server

import (
	"github.com/go-go-golems/go-go-mcp/pkg/services"
)

type ServerOption func(*Server)

func WithPromptService(ps services.PromptService) ServerOption {
	return func(s *Server) {
		s.promptService = ps
	}
}

func WithResourceService(rs services.ResourceService) ServerOption {
	return func(s *Server) {
		s.resourceService = rs
	}
}

func WithToolService(ts services.ToolService) ServerOption {
	return func(s *Server) {
		s.toolService = ts
	}
}

func WithInitializeService(is services.InitializeService) ServerOption {
	return func(s *Server) {
		s.initializeService = is
	}
}

func WithServerName(name string) ServerOption {
	return func(s *Server) {
		s.serverName = name
	}
}

func WithServerVersion(version string) ServerOption {
	return func(s *Server) {
		s.serverVersion = version
	}
}
