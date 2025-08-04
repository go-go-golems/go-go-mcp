package configstore

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// AmpStore implements the Store interface for Amp configuration
type AmpStore struct {
	editor types.ServerConfigEditor
	path   string
}

// NewAmpStore creates a new Amp configuration store
func NewAmpStore() (*AmpStore, error) {
	configPath, err := config.GetAmpConfigPath()
	if err != nil {
		return nil, err
	}

	editor, err := config.NewAmpCodeEditor(configPath)
	if err != nil {
		return nil, err
	}

	return &AmpStore{
		editor: editor,
		path:   configPath,
	}, nil
}

// Load loads the configuration
func (s *AmpStore) Load() error {
	// The editor automatically loads on creation, so this is a no-op
	return nil
}

// Save saves the configuration
func (s *AmpStore) Save() error {
	return s.editor.Save()
}

// ListServers returns all servers in the configuration
func (s *AmpStore) ListServers() (map[string]types.CommonServer, error) {
	return s.editor.ListServers()
}

// GetServer retrieves a specific server by name
func (s *AmpStore) GetServer(name string) (types.CommonServer, bool, error) {
	servers, err := s.editor.ListServers()
	if err != nil {
		return types.CommonServer{}, false, err
	}

	server, exists := servers[name]
	return server, exists, nil
}

// AddServer adds or updates a server
func (s *AmpStore) AddServer(server types.CommonServer, overwrite bool) error {
	return s.editor.AddMCPServer(server, overwrite)
}

// RemoveServer removes a server by name
func (s *AmpStore) RemoveServer(name string) error {
	return s.editor.RemoveMCPServer(name)
}

// IsServerDisabled checks if a server is disabled
func (s *AmpStore) IsServerDisabled(name string) (bool, error) {
	return s.editor.IsServerDisabled(name)
}

// EnableServer enables a server
func (s *AmpStore) EnableServer(name string) error {
	return s.editor.EnableMCPServer(name)
}

// DisableServer disables a server
func (s *AmpStore) DisableServer(name string) error {
	return s.editor.DisableMCPServer(name)
}

// ResolveTarget resolves and validates the target for Amp
func (s *AmpStore) ResolveTarget(target string) error {
	supportedTargets := s.GetSupportedTargets()

	if target == "" || target == "default" {
		return nil
	}

	return fmt.Errorf("unsupported target '%s' for Amp. Supported targets: %s",
		target, strings.Join(supportedTargets, ", "))
}

// GetSupportedTargets returns the supported targets for Amp
func (s *AmpStore) GetSupportedTargets() []string {
	return []string{"default", ""}
}

// GetConfigPath returns the path to the configuration file
func (s *AmpStore) GetConfigPath() string {
	return s.editor.GetConfigPath()
}

// AmpCodeStore implements the Store interface for AmpCode configuration
type AmpCodeStore struct {
	editor types.ServerConfigEditor
	path   string
}

// NewAmpCodeStore creates a new AmpCode configuration store
func NewAmpCodeStore() (*AmpCodeStore, error) {
	configPath, err := config.GetAmpCodeConfigPath()
	if err != nil {
		return nil, err
	}

	editor, err := config.NewAmpCodeEditor(configPath)
	if err != nil {
		return nil, err
	}

	return &AmpCodeStore{
		editor: editor,
		path:   configPath,
	}, nil
}

// Load loads the configuration
func (s *AmpCodeStore) Load() error {
	// The editor automatically loads on creation, so this is a no-op
	return nil
}

// Save saves the configuration
func (s *AmpCodeStore) Save() error {
	return s.editor.Save()
}

// ListServers returns all servers in the configuration
func (s *AmpCodeStore) ListServers() (map[string]types.CommonServer, error) {
	return s.editor.ListServers()
}

// GetServer retrieves a specific server by name
func (s *AmpCodeStore) GetServer(name string) (types.CommonServer, bool, error) {
	servers, err := s.editor.ListServers()
	if err != nil {
		return types.CommonServer{}, false, err
	}

	server, exists := servers[name]
	return server, exists, nil
}

// AddServer adds or updates a server
func (s *AmpCodeStore) AddServer(server types.CommonServer, overwrite bool) error {
	return s.editor.AddMCPServer(server, overwrite)
}

// RemoveServer removes a server by name
func (s *AmpCodeStore) RemoveServer(name string) error {
	return s.editor.RemoveMCPServer(name)
}

// IsServerDisabled checks if a server is disabled
func (s *AmpCodeStore) IsServerDisabled(name string) (bool, error) {
	return s.editor.IsServerDisabled(name)
}

// EnableServer enables a server
func (s *AmpCodeStore) EnableServer(name string) error {
	return s.editor.EnableMCPServer(name)
}

// DisableServer disables a server
func (s *AmpCodeStore) DisableServer(name string) error {
	return s.editor.DisableMCPServer(name)
}

// ResolveTarget resolves and validates the target for AmpCode
func (s *AmpCodeStore) ResolveTarget(target string) error {
	supportedTargets := s.GetSupportedTargets()

	if target == "" || target == "default" {
		return nil
	}

	return fmt.Errorf("unsupported target '%s' for AmpCode. Supported targets: %s",
		target, strings.Join(supportedTargets, ", "))
}

// GetSupportedTargets returns the supported targets for AmpCode
func (s *AmpCodeStore) GetSupportedTargets() []string {
	return []string{"default", ""}
}

// GetConfigPath returns the path to the configuration file
func (s *AmpCodeStore) GetConfigPath() string {
	return s.editor.GetConfigPath()
}
