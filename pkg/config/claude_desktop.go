package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/pkg/errors"
	"maps"
)

// ClaudeDesktopConfig represents the configuration for the Claude desktop application
type ClaudeDesktopConfig struct {
	MCPServers      map[string]MCPServer `json:"mcpServers"`
	DisabledServers map[string]MCPServer `json:"disabledServersConfig,omitempty"`
}

// MCPServer represents a server configuration
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// ClaudeDesktopEditor manages the Claude desktop configuration
type ClaudeDesktopEditor struct {
	config *ClaudeDesktopConfig
	path   string
}

// GetDefaultClaudeDesktopConfigPath returns the default path for the Claude desktop configuration file
func GetDefaultClaudeDesktopConfigPath() (string, error) {
	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get user config directory: %w", err)
	}
	return filepath.Join(xdgConfigPath, "Claude", "claude_desktop_config.json"), nil
}

// NewClaudeDesktopEditor creates a new editor for the Claude desktop configuration
func NewClaudeDesktopEditor(path string) (*ClaudeDesktopEditor, error) {
	if path == "" {
		var err error
		path, err = GetDefaultClaudeDesktopConfigPath()
		if err != nil {
			return nil, fmt.Errorf("could not get default config path: %w", err)
		}
	}

	editor := &ClaudeDesktopEditor{
		path: path,
	}

	// Try to load existing config
	if err := editor.load(); err != nil {
		// If file doesn't exist, create a new config
		if os.IsNotExist(err) {
			editor.config = &ClaudeDesktopConfig{
				MCPServers:      make(map[string]MCPServer),
				DisabledServers: make(map[string]MCPServer),
			}
		} else {
			return nil, fmt.Errorf("could not load config: %w", err)
		}
	}

	return editor, nil
}

// load reads the configuration from disk
func (e *ClaudeDesktopEditor) load() error {
	data, err := os.ReadFile(e.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to read config file %s", e.path)
		}
		e.config = &ClaudeDesktopConfig{
			MCPServers:      make(map[string]MCPServer),
			DisabledServers: make(map[string]MCPServer),
		}
		return nil
	}

	e.config = &ClaudeDesktopConfig{}
	if err := json.Unmarshal(data, e.config); err != nil {
		return errors.Wrapf(err, "could not parse config file %s", e.path)
	}

	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]MCPServer)
	}
	if e.config.DisabledServers == nil {
		e.config.DisabledServers = make(map[string]MCPServer)
	}

	return nil
}

// Save writes the configuration to disk
func (e *ClaudeDesktopEditor) Save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(e.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrapf(err, "could not create config directory %s", dir)
	}

	data, err := json.MarshalIndent(e.config, "", "  ")
	if err != nil {
		return errors.Wrap(err, "could not marshal config")
	}

	if err := os.WriteFile(e.path, data, 0644); err != nil {
		return errors.Wrapf(err, "could not write config to %s", e.path)
	}

	return nil
}

// AddMCPServer adds or updates a server configuration using the CommonServer struct.
func (e *ClaudeDesktopEditor) AddMCPServer(server types.CommonServer, overwrite bool) error {
	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]MCPServer)
	}
	if e.config.DisabledServers == nil {
		e.config.DisabledServers = make(map[string]MCPServer)
	}

	name := server.Name
	_, existsEnabled := e.config.MCPServers[name]
	_, existsDisabled := e.config.DisabledServers[name]

	if (existsEnabled || existsDisabled) && !overwrite {
		return fmt.Errorf("MCP server '%s' already exists. Use overwrite option to replace it", name)
	}

	claudeServer := MCPServer{
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Env,
	}

	if existsDisabled {
		delete(e.config.DisabledServers, name)
	}

	e.config.MCPServers[name] = claudeServer

	return nil
}

// RemoveMCPServer removes an MCP server configuration from both enabled and disabled lists.
func (e *ClaudeDesktopEditor) RemoveMCPServer(name string) error {
	_, existsEnabled := e.config.MCPServers[name]
	_, existsDisabled := e.config.DisabledServers[name]

	if !existsEnabled && !existsDisabled {
		return fmt.Errorf("MCP server '%s' not found", name)
	}

	if existsEnabled {
		delete(e.config.MCPServers, name)
	}
	if existsDisabled {
		delete(e.config.DisabledServers, name)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file
func (e *ClaudeDesktopEditor) GetConfigPath() string {
	return e.path
}

// ListServers returns a map of all configured servers (enabled and disabled) as CommonServer.
func (e *ClaudeDesktopEditor) ListServers() (map[string]types.CommonServer, error) {
	servers := make(map[string]types.CommonServer)

	// Add enabled MCP servers
	for name, server := range e.config.MCPServers {
		servers[name] = types.CommonServer{
			Name:    name,
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			IsSSE:   false,
		}
	}

	// Add disabled MCP servers
	if e.config.DisabledServers != nil {
		for name, server := range e.config.DisabledServers {
			if _, exists := servers[name]; !exists {
				servers[name] = types.CommonServer{
					Name:    name,
					Command: server.Command,
					Args:    server.Args,
					Env:     server.Env,
					IsSSE:   false,
				}
			}
		}
	}

	return servers, nil
}

// EnableMCPServer enables a previously disabled MCP server
func (e *ClaudeDesktopEditor) EnableMCPServer(name string) error {
	if len(e.config.DisabledServers) == 0 {
		if _, exists := e.config.MCPServers[name]; exists {
			return fmt.Errorf("server '%s' is already enabled", name)
		}
		return fmt.Errorf("server '%s' not found in disabled servers", name)
	}

	server, exists := e.config.DisabledServers[name]
	if !exists {
		if _, enabledExists := e.config.MCPServers[name]; enabledExists {
			return fmt.Errorf("server '%s' is already enabled", name)
		}
		return fmt.Errorf("server '%s' not found in disabled servers", name)
	}

	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]MCPServer)
	}
	e.config.MCPServers[name] = server
	delete(e.config.DisabledServers, name)

	return nil
}

// DisableMCPServer disables an MCP server without removing its configuration
func (e *ClaudeDesktopEditor) DisableMCPServer(name string) error {
	server, exists := e.config.MCPServers[name]
	if !exists {
		if _, disabledExists := e.config.DisabledServers[name]; disabledExists {
			return fmt.Errorf("server '%s' is already disabled", name)
		}
		return fmt.Errorf("enabled MCP server '%s' not found", name)
	}

	if e.config.DisabledServers == nil {
		e.config.DisabledServers = make(map[string]MCPServer)
	}
	e.config.DisabledServers[name] = server
	delete(e.config.MCPServers, name)

	return nil
}

// IsServerDisabled checks if a server is in the disabled list.
func (e *ClaudeDesktopEditor) IsServerDisabled(name string) (bool, error) {
	if e.config.DisabledServers == nil {
		if _, exists := e.config.MCPServers[name]; exists {
			return false, nil
		}
		return false, fmt.Errorf("server '%s' not found", name)
	}
	_, exists := e.config.DisabledServers[name]
	if !exists {
		if _, enabledExists := e.config.MCPServers[name]; !enabledExists {
			return false, fmt.Errorf("server '%s' not found", name)
		}
	}
	return exists, nil
}

// ListDisabledServers returns a list of disabled server names
func (e *ClaudeDesktopEditor) ListDisabledServers() ([]string, error) {
	if e.config.DisabledServers == nil {
		return []string{}, nil
	}
	disabledServers := make([]string, 0, len(e.config.DisabledServers))
	for name := range e.config.DisabledServers {
		disabledServers = append(disabledServers, name)
	}
	return disabledServers, nil
}

// GetServer retrieves a specific server configuration by name as CommonServer.
func (e *ClaudeDesktopEditor) GetServer(name string) (types.CommonServer, bool, error) {
	server, exists := e.config.MCPServers[name]
	isDisabled := false
	if !exists {
		if e.config.DisabledServers != nil {
			server, exists = e.config.DisabledServers[name]
			if exists {
				isDisabled = true
			}
		}
	}

	if !exists {
		return types.CommonServer{}, false, nil
	}

	common := types.CommonServer{
		Name:    name,
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Env,
		IsSSE:   false,
	}
	_ = isDisabled

	return common, true, nil
}

// --- Deprecated Methods (kept temporarily for compatibility, remove later) ---

// AddMCPServer adds or updates an MCP server configuration
// DEPRECATED: Use AddMCPServer with CommonServer instead.
func (e *ClaudeDesktopEditor) AddMCPServerRaw(name string, command string, args []string, env map[string]string, overwrite bool) error {
	common := types.CommonServer{
		Name:    name,
		Command: command,
		Args:    args,
		Env:     env,
		IsSSE:   false,
	}
	return e.AddMCPServer(common, overwrite)
}

// ListServers returns a list of configured MCP servers
// DEPRECATED: Use the new ListServers which returns map[string]CommonServer.
func (e *ClaudeDesktopEditor) ListServersRaw() map[string]MCPServer {
	servers := make(map[string]MCPServer)

	// Add enabled MCP servers
	maps.Copy(servers, e.config.MCPServers)

	// Add disabled MCP servers
	if e.config.DisabledServers != nil {
		maps.Copy(servers, e.config.DisabledServers)
	}

	return servers
}
