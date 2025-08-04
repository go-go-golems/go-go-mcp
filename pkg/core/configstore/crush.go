package configstore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/spf13/viper"
)

// CrushStore implements the Store interface for Crush configuration
type CrushStore struct {
	editor    types.ServerConfigEditor
	path      string
	storeType string
}

// NewCrushLocalStore creates a new Crush configuration store for .crush.json
func NewCrushLocalStore() (*CrushStore, error) {
	configPath := ".crush.json"
	editor, err := config.NewCrushEditor(configPath)
	if err != nil {
		return nil, err
	}

	return &CrushStore{
		editor:    editor,
		path:      configPath,
		storeType: "local",
	}, nil
}

// NewCrushCwdStore creates a new Crush configuration store for crush.json
func NewCrushCwdStore() (*CrushStore, error) {
	configPath := "crush.json"
	editor, err := config.NewCrushEditor(configPath)
	if err != nil {
		return nil, err
	}

	return &CrushStore{
		editor:    editor,
		path:      configPath,
		storeType: "cwd",
	}, nil
}

// NewCrushGlobalStore creates a new Crush configuration store for global crush.json
func NewCrushGlobalStore() (*CrushStore, error) {
	configPath := viper.GetString("HOME") + "/.config/crush/crush.json"
	if configPath == "/.config/crush/crush.json" {
		// Fallback if HOME is not set
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configPath = filepath.Join(homeDir, ".config", "crush", "crush.json")
	}

	editor, err := config.NewCrushEditor(configPath)
	if err != nil {
		return nil, err
	}

	return &CrushStore{
		editor:    editor,
		path:      configPath,
		storeType: "global",
	}, nil
}

// Load loads the configuration
func (s *CrushStore) Load() error {
	// The editor automatically loads on creation, so this is a no-op
	return nil
}

// Save saves the configuration
func (s *CrushStore) Save() error {
	return s.editor.Save()
}

// ListServers returns all servers in the configuration
func (s *CrushStore) ListServers() (map[string]types.CommonServer, error) {
	return s.editor.ListServers()
}

// GetServer retrieves a specific server by name
func (s *CrushStore) GetServer(name string) (types.CommonServer, bool, error) {
	servers, err := s.editor.ListServers()
	if err != nil {
		return types.CommonServer{}, false, err
	}

	server, exists := servers[name]
	return server, exists, nil
}

// AddServer adds or updates a server
func (s *CrushStore) AddServer(server types.CommonServer, overwrite bool) error {
	return s.editor.AddMCPServer(server, overwrite)
}

// RemoveServer removes a server by name
func (s *CrushStore) RemoveServer(name string) error {
	return s.editor.RemoveMCPServer(name)
}

// IsServerDisabled checks if a server is disabled
func (s *CrushStore) IsServerDisabled(name string) (bool, error) {
	return s.editor.IsServerDisabled(name)
}

// EnableServer enables a server
func (s *CrushStore) EnableServer(name string) error {
	return s.editor.EnableMCPServer(name)
}

// DisableServer disables a server
func (s *CrushStore) DisableServer(name string) error {
	return s.editor.DisableMCPServer(name)
}

// ResolveTarget resolves and validates the target for Crush
func (s *CrushStore) ResolveTarget(target string) error {
	supportedTargets := s.GetSupportedTargets()

	// Normalize target
	if target == "" || target == "default" {
		target = s.storeType
	}

	// Check if this store supports the target
	if target != s.storeType {
		return fmt.Errorf("target '%s' not supported by %s Crush store. Supported targets: %s",
			target, s.storeType, strings.Join(supportedTargets, ", "))
	}

	return nil
}

// GetSupportedTargets returns the supported targets for this Crush store
func (s *CrushStore) GetSupportedTargets() []string {
	return []string{s.storeType}
}

// GetConfigPath returns the path to the configuration file
func (s *CrushStore) GetConfigPath() string {
	return s.editor.GetConfigPath()
}
