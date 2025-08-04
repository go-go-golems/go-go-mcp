package configstore

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// CursorStore implements the Store interface for Cursor configuration
type CursorStore struct {
	editor types.ServerConfigEditor
	path   string
	target string
}

// NewCursorStore creates a new Cursor configuration store
func NewCursorStore() (*CursorStore, error) {
	configPath, err := config.GetGlobalCursorMCPConfigPath()
	if err != nil {
		return nil, err
	}

	editor, err := config.NewCursorMCPEditor(configPath)
	if err != nil {
		return nil, err
	}

	return &CursorStore{
		editor: editor,
		path:   configPath,
		target: "global", // default to global
	}, nil
}

// Load loads the configuration
func (s *CursorStore) Load() error {
	// The editor automatically loads on creation, so this is a no-op
	return nil
}

// Save saves the configuration
func (s *CursorStore) Save() error {
	return s.editor.Save()
}

// ListServers returns all servers in the configuration
func (s *CursorStore) ListServers() (map[string]types.CommonServer, error) {
	return s.editor.ListServers()
}

// GetServer retrieves a specific server by name
func (s *CursorStore) GetServer(name string) (types.CommonServer, bool, error) {
	servers, err := s.editor.ListServers()
	if err != nil {
		return types.CommonServer{}, false, err
	}

	server, exists := servers[name]
	return server, exists, nil
}

// AddServer adds or updates a server
func (s *CursorStore) AddServer(server types.CommonServer, overwrite bool) error {
	return s.editor.AddMCPServer(server, overwrite)
}

// RemoveServer removes a server by name
func (s *CursorStore) RemoveServer(name string) error {
	return s.editor.RemoveMCPServer(name)
}

// IsServerDisabled checks if a server is disabled
func (s *CursorStore) IsServerDisabled(name string) (bool, error) {
	return s.editor.IsServerDisabled(name)
}

// EnableServer enables a server
func (s *CursorStore) EnableServer(name string) error {
	return s.editor.EnableMCPServer(name)
}

// DisableServer disables a server
func (s *CursorStore) DisableServer(name string) error {
	return s.editor.DisableMCPServer(name)
}

// ResolveTarget resolves and validates the target for Cursor
func (s *CursorStore) ResolveTarget(target string) error {
	supportedTargets := s.GetSupportedTargets()

	// Normalize target
	if target == "" || target == "default" {
		target = "global"
	}

	// Check if target is supported
	supported := false
	for _, t := range supportedTargets {
		if t == target {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported target '%s' for Cursor. Supported targets: %s",
			target, strings.Join(supportedTargets, ", "))
	}

	// Switch to the appropriate config if needed
	if s.target != target {
		var configPath string
		var err error

		switch target {
		case "global":
			configPath, err = config.GetGlobalCursorMCPConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get config path for target '%s': %w", target, err)
			}
		case "cwd":
			configPath = config.GetProjectCursorMCPConfigPath(".")
		default:
			return fmt.Errorf("unsupported target: %s", target)
		}

		editor, err := config.NewCursorMCPEditor(configPath)
		if err != nil {
			return fmt.Errorf("failed to create editor for target '%s': %w", target, err)
		}

		s.editor = editor
		s.path = configPath
		s.target = target
	}

	return nil
}

// GetSupportedTargets returns the supported targets for Cursor
func (s *CursorStore) GetSupportedTargets() []string {
	return []string{"global", "cwd"}
}

// GetConfigPath returns the path to the configuration file
func (s *CursorStore) GetConfigPath() string {
	return s.editor.GetConfigPath()
}
