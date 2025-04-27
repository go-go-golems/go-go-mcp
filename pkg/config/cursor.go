package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
				MCPServers: make(map[string]CursorMCPServer),
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
		return err
	}

	e.config = &CursorMCPConfig{}
	if err := json.Unmarshal(data, e.config); err != nil {
		return fmt.Errorf("could not parse config: %w", err)
	}

	return nil
}

// Save writes the configuration to disk
func (e *CursorMCPEditor) Save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(e.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	data, err := json.MarshalIndent(e.config, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if err := os.WriteFile(e.path, data, 0644); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}

	return nil
}

// AddMCPServer adds or updates an MCP server configuration (stdio format)
func (e *CursorMCPEditor) AddMCPServer(name string, command string, args []string, env map[string]string, overwrite bool) error {
	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]CursorMCPServer)
	}

	// Check if server already exists
	if _, exists := e.config.MCPServers[name]; exists && !overwrite {
		return fmt.Errorf("MCP server '%s' already exists. Use --overwrite to replace it", name)
	}

	e.config.MCPServers[name] = CursorMCPServer{
		Command: command,
		Args:    args,
		Env:     env,
	}

	return nil
}

// AddMCPServerSSE adds or updates an MCP server configuration (SSE format)
func (e *CursorMCPEditor) AddMCPServerSSE(name string, url string, env map[string]string, overwrite bool) error {
	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]CursorMCPServer)
	}

	// Check if server already exists
	if _, exists := e.config.MCPServers[name]; exists && !overwrite {
		return fmt.Errorf("MCP server '%s' already exists. Use --overwrite to replace it", name)
	}

	e.config.MCPServers[name] = CursorMCPServer{
		URL: url,
		Env: env,
	}

	return nil
}

// RemoveMCPServer removes an MCP server configuration
func (e *CursorMCPEditor) RemoveMCPServer(name string) error {
	if e.config.MCPServers == nil {
		return fmt.Errorf("no MCP servers configured")
	}

	if _, exists := e.config.MCPServers[name]; !exists {
		return fmt.Errorf("MCP server %s not found", name)
	}

	delete(e.config.MCPServers, name)
	return nil
}

// GetConfigPath returns the path to the configuration file
func (e *CursorMCPEditor) GetConfigPath() string {
	return e.path
}

// ListServers returns a list of configured MCP servers
func (e *CursorMCPEditor) ListServers() map[string]CursorMCPServer {
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

// EnableMCPServer enables a previously disabled MCP server
func (e *CursorMCPEditor) EnableMCPServer(name string) error {
	if len(e.config.DisabledServers) == 0 {
		return fmt.Errorf("no disabled servers found")
	}

	// Check if server exists in disabled servers
	server, exists := e.config.DisabledServers[name]
	if !exists {
		return fmt.Errorf("server '%s' is not disabled", name)
	}

	// Move server from disabled to enabled
	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]CursorMCPServer)
	}
	e.config.MCPServers[name] = server
	delete(e.config.DisabledServers, name)

	return nil
}

// DisableMCPServer disables an MCP server without removing its configuration
func (e *CursorMCPEditor) DisableMCPServer(name string) error {
	// Check if server exists
	server, exists := e.config.MCPServers[name]
	if !exists {
		return fmt.Errorf("MCP server '%s' not found", name)
	}

	// Move server to disabled
	if e.config.DisabledServers == nil {
		e.config.DisabledServers = make(map[string]CursorMCPServer)
	}
	e.config.DisabledServers[name] = server
	delete(e.config.MCPServers, name)

	return nil
}

// IsServerDisabled checks if a server is disabled
func (e *CursorMCPEditor) IsServerDisabled(name string) bool {
	if e.config.DisabledServers == nil {
		return false
	}
	_, exists := e.config.DisabledServers[name]
	return exists
}

// ListDisabledServers returns a list of disabled server names
func (e *CursorMCPEditor) ListDisabledServers() []string {
	if e.config.DisabledServers == nil {
		return []string{}
	}
	disabledServers := make([]string, 0, len(e.config.DisabledServers))
	for name := range e.config.DisabledServers {
		disabledServers = append(disabledServers, name)
	}
	return disabledServers
}

// GetServer retrieves a server's configuration by name
func (e *CursorMCPEditor) GetServer(name string) (CursorMCPServer, error) {
	server, exists := e.config.MCPServers[name]
	if !exists {
		// Check in disabled servers
		if e.config.DisabledServers != nil {
			server, exists = e.config.DisabledServers[name]
			if exists {
				return server, nil
			}
		}
		return CursorMCPServer{}, fmt.Errorf("MCP server '%s' not found", name)
	}
	return server, nil
}
