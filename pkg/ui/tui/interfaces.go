package tui

import (
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// CommonServer is a unified structure to represent server configurations
// from different sources (Cursor, Claude Desktop, etc.) within the UI.
type CommonServer struct {
	Name    string            // Name identifier for the server
	Command string            // Command to execute for the server (Claude)
	Args    []string          // Arguments for the command (Claude)
	Env     map[string]string // Environment variables (Claude)
	URL     string            // URL for SSE connection (Cursor)
	IsSSE   bool              // Type identifier (true for Cursor SSE, false for Claude Command)
}

// ServerConfigEditor defines the interface for managing server configurations
// in a backend-agnostic way for the TUI.
type ServerConfigEditor interface {
	// ListServers retrieves all configured servers (both enabled and disabled).
	// It returns a map where the key is the server name.
	ListServers() (map[string]CommonServer, error)

	// ListDisabledServers returns the names of disabled servers.
	// This might be redundant if ListServers includes enabled status, but kept for now
	// based on the original proposal.
	ListDisabledServers() ([]string, error)

	// EnableMCPServer enables a specific server by name.
	EnableMCPServer(name string) error

	// DisableMCPServer disables a specific server by name.
	DisableMCPServer(name string) error

	// AddMCPServer adds or updates a server configuration.
	AddMCPServer(server CommonServer, overwrite bool) error

	// RemoveMCPServer removes a specific server by name.
	RemoveMCPServer(name string) error

	// Save persists the configuration changes to the underlying storage.
	Save() error

	// GetConfigPath returns the path of the configuration file being managed.
	GetConfigPath() string

	// IsServerDisabled checks if a server with the given name is currently disabled.
	IsServerDisabled(name string) (bool, error)

	// GetServer retrieves a specific server configuration by name.
	GetServer(name string) (CommonServer, bool, error)
}

// Ensure config types implement the interface (compile-time check)
var _ types.ServerConfigEditor = &config.CursorMCPEditor{}
var _ types.ServerConfigEditor = &config.ClaudeDesktopEditor{}
