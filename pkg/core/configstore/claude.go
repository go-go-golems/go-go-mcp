package configstore

import (
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// ClaudeStore implements the Store interface for Claude Desktop configuration
type ClaudeStore struct {
	editor types.ServerConfigEditor
	path   string
}

// NewClaudeStore creates a new Claude configuration store
func NewClaudeStore() (*ClaudeStore, error) {
	configPath, err := config.GetDefaultClaudeDesktopConfigPath()
	if err != nil {
		return nil, err
	}

	editor, err := config.NewClaudeDesktopEditor(configPath)
	if err != nil {
		return nil, err
	}

	return &ClaudeStore{
		editor: editor,
		path:   configPath,
	}, nil
}

// Load loads the configuration
func (s *ClaudeStore) Load() error {
	// The editor automatically loads on creation, so this is a no-op
	return nil
}

// Save saves the configuration
func (s *ClaudeStore) Save() error {
	return s.editor.Save()
}

// ListServers returns all servers in the configuration
func (s *ClaudeStore) ListServers() (map[string]types.CommonServer, error) {
	return s.editor.ListServers()
}

// GetServer retrieves a specific server by name
func (s *ClaudeStore) GetServer(name string) (types.CommonServer, bool, error) {
	servers, err := s.editor.ListServers()
	if err != nil {
		return types.CommonServer{}, false, err
	}

	server, exists := servers[name]
	return server, exists, nil
}

// AddServer adds or updates a server
func (s *ClaudeStore) AddServer(server types.CommonServer, overwrite bool) error {
	return s.editor.AddMCPServer(server, overwrite)
}

// RemoveServer removes a server by name
func (s *ClaudeStore) RemoveServer(name string) error {
	return s.editor.RemoveMCPServer(name)
}

// IsServerDisabled checks if a server is disabled
func (s *ClaudeStore) IsServerDisabled(name string) (bool, error) {
	return s.editor.IsServerDisabled(name)
}

// EnableServer enables a server
func (s *ClaudeStore) EnableServer(name string) error {
	return s.editor.EnableMCPServer(name)
}

// DisableServer disables a server
func (s *ClaudeStore) DisableServer(name string) error {
	return s.editor.DisableMCPServer(name)
}
