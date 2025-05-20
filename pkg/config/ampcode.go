package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/pkg/errors"
)

// AmpMCPServer represents an MCP server configuration for Ampcode
// stored inside the Cursor settings.json file.
type AmpMCPServer struct {
	Command  string            `json:"command"`
	Args     []string          `json:"args,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Disabled bool              `json:"disabled"`
}

// AmpSettings represents the subset of Cursor settings used by Ampcode.
type AmpSettings struct {
	MCPServers map[string]AmpMCPServer `json:"amp.mcpServers"`
}

// AmpSettingsEditor manages the Ampcode MCP configuration.
type AmpSettingsEditor struct {
	config *AmpSettings
	path   string
}

// GetDefaultAmpSettingsPath returns the default path to the Cursor settings.json
// used by Ampcode.
func GetDefaultAmpSettingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "Cursor", "User", "settings.json"), nil
}

// NewAmpSettingsEditor creates a new editor for the Ampcode MCP configuration.
func NewAmpSettingsEditor(path string) (*AmpSettingsEditor, error) {
	editor := &AmpSettingsEditor{path: path}
	if err := editor.load(); err != nil {
		if os.IsNotExist(err) {
			editor.config = &AmpSettings{MCPServers: make(map[string]AmpMCPServer)}
		} else {
			return nil, errors.Wrap(err, "could not load amp settings")
		}
	}
	if editor.config == nil {
		editor.config = &AmpSettings{MCPServers: make(map[string]AmpMCPServer)}
	}
	return editor, nil
}

func (e *AmpSettingsEditor) load() error {
	data, err := os.ReadFile(e.path)
	if err != nil {
		return err
	}
	cfg := &AmpSettings{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return errors.Wrap(err, "failed to unmarshal settings")
	}
	if cfg.MCPServers == nil {
		cfg.MCPServers = make(map[string]AmpMCPServer)
	}
	e.config = cfg
	return nil
}

// Save writes the configuration back to disk.
func (e *AmpSettingsEditor) Save() error {
	if e.config == nil {
		e.config = &AmpSettings{MCPServers: make(map[string]AmpMCPServer)}
	}
	dir := filepath.Dir(e.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrapf(err, "could not create config directory %s", dir)
	}
	data, err := json.MarshalIndent(e.config, "", "  ")
	if err != nil {
		return errors.Wrap(err, "could not marshal settings")
	}
	return os.WriteFile(e.path, data, 0644)
}

// AddMCPServer adds or updates a server configuration.
func (e *AmpSettingsEditor) AddMCPServer(server types.CommonServer, overwrite bool) error {
	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]AmpMCPServer)
	}
	if _, exists := e.config.MCPServers[server.Name]; exists && !overwrite {
		return fmt.Errorf("MCP server '%s' already exists. Use overwrite to replace it", server.Name)
	}
	e.config.MCPServers[server.Name] = AmpMCPServer{
		Command:  server.Command,
		Args:     server.Args,
		Env:      server.Env,
		Disabled: false,
	}
	return nil
}

// RemoveMCPServer removes an MCP server configuration.
func (e *AmpSettingsEditor) RemoveMCPServer(name string) error {
	if e.config.MCPServers == nil {
		return fmt.Errorf("server '%s' not found", name)
	}
	if _, exists := e.config.MCPServers[name]; !exists {
		return fmt.Errorf("server '%s' not found", name)
	}
	delete(e.config.MCPServers, name)
	return nil
}

// ListServers returns all MCP servers, enabled and disabled.
func (e *AmpSettingsEditor) ListServers() (map[string]types.CommonServer, error) {
	servers := make(map[string]types.CommonServer)
	for name, srv := range e.config.MCPServers {
		servers[name] = types.CommonServer{
			Name:    name,
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
			IsSSE:   false,
		}
	}
	return servers, nil
}

// ListDisabledServers returns the names of disabled servers.
func (e *AmpSettingsEditor) ListDisabledServers() ([]string, error) {
	disabled := []string{}
	for name, srv := range e.config.MCPServers {
		if srv.Disabled {
			disabled = append(disabled, name)
		}
	}
	return disabled, nil
}

// EnableMCPServer marks the server as enabled.
func (e *AmpSettingsEditor) EnableMCPServer(name string) error {
	srv, ok := e.config.MCPServers[name]
	if !ok {
		return fmt.Errorf("server '%s' not found", name)
	}
	if !srv.Disabled {
		return fmt.Errorf("server '%s' is already enabled", name)
	}
	srv.Disabled = false
	e.config.MCPServers[name] = srv
	return nil
}

// DisableMCPServer marks the server as disabled.
func (e *AmpSettingsEditor) DisableMCPServer(name string) error {
	srv, ok := e.config.MCPServers[name]
	if !ok {
		return fmt.Errorf("server '%s' not found", name)
	}
	if srv.Disabled {
		return fmt.Errorf("server '%s' is already disabled", name)
	}
	srv.Disabled = true
	e.config.MCPServers[name] = srv
	return nil
}

// GetConfigPath returns the underlying configuration path.
func (e *AmpSettingsEditor) GetConfigPath() string { return e.path }

// IsServerDisabled checks if a server is disabled.
func (e *AmpSettingsEditor) IsServerDisabled(name string) (bool, error) {
	srv, ok := e.config.MCPServers[name]
	if !ok {
		return false, fmt.Errorf("server '%s' not found", name)
	}
	return srv.Disabled, nil
}

// GetServer retrieves a specific server configuration.
func (e *AmpSettingsEditor) GetServer(name string) (types.CommonServer, bool, error) {
	srv, ok := e.config.MCPServers[name]
	if !ok {
		return types.CommonServer{}, false, nil
	}
	return types.CommonServer{
		Name:    name,
		Command: srv.Command,
		Args:    srv.Args,
		Env:     srv.Env,
		IsSSE:   false,
	}, true, nil
}

var _ types.ServerConfigEditor = &AmpSettingsEditor{}
