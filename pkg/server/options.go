package server

import (
	"github.com/go-go-golems/go-go-mcp/pkg"
)

type ServerOption func(*Server)

func WithPromptProvider(pp pkg.PromptProvider) ServerOption {
	return func(s *Server) {
		s.promptProvider = pp
	}
}

func WithResourceProvider(rp pkg.ResourceProvider) ServerOption {
	return func(s *Server) {
		s.resourceProvider = rp
	}
}

func WithToolProvider(tp pkg.ToolProvider) ServerOption {
	return func(s *Server) {
		s.toolProvider = tp
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
