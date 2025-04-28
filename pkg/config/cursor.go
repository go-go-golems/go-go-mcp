package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/pkg/errors"
)

// CursorMCPConfig represents the configuration for Cursor MCP
type CursorMCPConfig struct {
	MCPServers      map[string]CursorMCPServer `json:"mcpServers"`
	DisabledServers map[string]CursorMCPServer `json:"disabledServersConfig,omitempty"`
}

// CursorMCPServer represents a server configuration for Cursor
type CursorMCPServer struct {
	// For stdio format
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	// For SSE format
	URL string `json:"url,omitempty"`
}

// CursorMCPEditor manages the Cursor MCP configuration
type CursorMCPEditor struct {
	config *CursorMCPConfig
	path   string
}

// GetGlobalCursorMCPConfigPath returns the path for the global Cursor MCP configuration file
func GetGlobalCursorMCPConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".cursor", "mcp.json"), nil
}

// GetProjectCursorMCPConfigPath returns the path for the project-specific Cursor MCP configuration file
func GetProjectCursorMCPConfigPath(projectDir string) string {
	return filepath.Join(projectDir, ".cursor", "mcp.json")
}

// NewCursorMCPEditor creates a new editor for the Cursor MCP configuration
func NewCursorMCPEditor(path string) (*CursorMCPEditor, error) {
	editor := &CursorMCPEditor{
		path: path,
	}

	// Try to load existing config
	if err := editor.load(); err != nil {
		// If file doesn't exist, create a new config
		if os.IsNotExist(err) {
			editor.config = &CursorMCPConfig{
				MCPServers:      make(map[string]CursorMCPServer),
				DisabledServers: make(map[string]CursorMCPServer),
			}
		} else {
			return nil, fmt.Errorf("could not load config: %w", err)
		}
	}

	return editor, nil
}

// load reads the configuration from disk
func (e *CursorMCPEditor) load() error {
	data, err := os.ReadFile(e.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to read config file %s", e.path)
		}
		e.config = &CursorMCPConfig{
			MCPServers:      make(map[string]CursorMCPServer),
			DisabledServers: make(map[string]CursorMCPServer),
		}
		return nil
	}

	e.config = &CursorMCPConfig{}
	if err := json.Unmarshal(data, e.config); err != nil {
		return errors.Wrapf(err, "could not parse config file %s", e.path)
	}

	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]CursorMCPServer)
	}
	if e.config.DisabledServers == nil {
		e.config.DisabledServers = make(map[string]CursorMCPServer)
	}

	return nil
}

// Save writes the configuration to disk
func (e *CursorMCPEditor) Save() error {
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
func (e *CursorMCPEditor) AddMCPServer(server types.CommonServer, overwrite bool) error {
	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]CursorMCPServer)
	}
	if e.config.DisabledServers == nil {
		e.config.DisabledServers = make(map[string]CursorMCPServer)
	}

	name := server.Name
	_, existsEnabled := e.config.MCPServers[name]
	_, existsDisabled := e.config.DisabledServers[name]

	if (existsEnabled || existsDisabled) && !overwrite {
		return fmt.Errorf("MCP server '%s' already exists. Use overwrite option to replace it", name)
	}

	cursorServer := CursorMCPServer{
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Env,
		URL:     server.URL,
	}

	if existsDisabled {
		delete(e.config.DisabledServers, name)
	}

	e.config.MCPServers[name] = cursorServer

	return nil
}

// RemoveMCPServer removes an MCP server configuration from both enabled and disabled lists.
func (e *CursorMCPEditor) RemoveMCPServer(name string) error {
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
func (e *CursorMCPEditor) GetConfigPath() string {
	return e.path
}

// ListServers returns a map of all configured servers (enabled and disabled) as CommonServer.
func (e *CursorMCPEditor) ListServers() (map[string]types.CommonServer, error) {
	servers := make(map[string]types.CommonServer)

	// Add enabled MCP servers
	for name, server := range e.config.MCPServers {
		servers[name] = types.CommonServer{
			Name:    name,
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			URL:     server.URL,
			IsSSE:   server.URL != "",
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
					URL:     server.URL,
					IsSSE:   server.URL != "",
				}
			}
		}
	}

	return servers, nil
}

// EnableMCPServer enables a previously disabled MCP server
func (e *CursorMCPEditor) EnableMCPServer(name string) error {
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
		e.config.MCPServers = make(map[string]CursorMCPServer)
	}
	e.config.MCPServers[name] = server
	delete(e.config.DisabledServers, name)

	return nil
}

// DisableMCPServer disables an MCP server without removing its configuration
func (e *CursorMCPEditor) DisableMCPServer(name string) error {
	server, exists := e.config.MCPServers[name]
	if !exists {
		if _, disabledExists := e.config.DisabledServers[name]; disabledExists {
			return fmt.Errorf("server '%s' is already disabled", name)
		}
		return fmt.Errorf("enabled MCP server '%s' not found", name)
	}

	if e.config.DisabledServers == nil {
		e.config.DisabledServers = make(map[string]CursorMCPServer)
	}
	e.config.DisabledServers[name] = server
	delete(e.config.MCPServers, name)

	return nil
}

// IsServerDisabled checks if a server is in the disabled list.
func (e *CursorMCPEditor) IsServerDisabled(name string) (bool, error) {
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
func (e *CursorMCPEditor) ListDisabledServers() ([]string, error) {
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
func (e *CursorMCPEditor) GetServer(name string) (types.CommonServer, bool, error) {
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
		URL:     server.URL,
		IsSSE:   server.URL != "",
	}
	_ = isDisabled

	return common, true, nil
}

// --- Deprecated Methods (kept temporarily for compatibility, remove later) ---

// AddMCPServer adds or updates an MCP server configuration (stdio format)
// DEPRECATED: Use AddMCPServer with CommonServer instead.
func (e *CursorMCPEditor) AddMCPServerStdio(name string, command string, args []string, env map[string]string, overwrite bool) error {
	common := types.CommonServer{
		Name:    name,
		Command: command,
		Args:    args,
		Env:     env,
		IsSSE:   false,
	}
	return e.AddMCPServer(common, overwrite)
}

// AddMCPServerSSE adds or updates an MCP server configuration (SSE format)
// DEPRECATED: Use AddMCPServer with CommonServer instead.
func (e *CursorMCPEditor) AddMCPServerSSE(name string, url string, env map[string]string, overwrite bool) error {
	common := types.CommonServer{
		Name:  name,
		URL:   url,
		Env:   env,
		IsSSE: true,
	}
	return e.AddMCPServer(common, overwrite)
}

// GetServer retrieves a server's configuration by name
// DEPRECATED: Use the new GetServer which returns CommonServer.
func (e *CursorMCPEditor) GetServerRaw(name string) (CursorMCPServer, error) {
	server, exists := e.config.MCPServers[name]
	if !exists {
		if e.config.DisabledServers != nil {
			server, exists = e.config.DisabledServers[name]
			if exists {
				return server, nil
			}
		}
		return CursorMCPServer{}, fmt.Errorf("MCP server %s not found", name)
	}
	return server, nil
}

// ListServers returns a list of configured MCP servers
// DEPRECATED: Use the new ListServers which returns map[string]CommonServer.
func (e *CursorMCPEditor) ListServersRaw() map[string]CursorMCPServer {
	servers := make(map[string]CursorMCPServer)

	// Add enabled MCP servers
	for name, server := range e.config.MCPServers {
		servers[name] = server
	}

	// Add disabled MCP servers
	if e.config.DisabledServers != nil {
		for name, server := range e.config.DisabledServers {
			servers[name] = server
		}
	}

	return servers
}
