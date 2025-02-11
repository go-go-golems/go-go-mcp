package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ClaudeDesktopConfig represents the configuration for the Claude desktop application
type ClaudeDesktopConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
	GoGoMCP    MCPServer            `json:"go-go-mcp"`
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
				MCPServers: make(map[string]MCPServer),
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
		return err
	}

	e.config = &ClaudeDesktopConfig{}
	if err := json.Unmarshal(data, e.config); err != nil {
		return fmt.Errorf("could not parse config: %w", err)
	}

	return nil
}

// Save writes the configuration to disk
func (e *ClaudeDesktopEditor) Save() error {
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

// AddMCPServer adds or updates an MCP server configuration
func (e *ClaudeDesktopEditor) AddMCPServer(name string, command string, args []string, env map[string]string, overwrite bool) error {
	if e.config.MCPServers == nil {
		e.config.MCPServers = make(map[string]MCPServer)
	}

	// Check if server already exists
	if _, exists := e.config.MCPServers[name]; exists && !overwrite {
		return fmt.Errorf("MCP server '%s' already exists. Use --overwrite to replace it", name)
	}

	e.config.MCPServers[name] = MCPServer{
		Command: command,
		Args:    args,
		Env:     env,
	}

	return nil
}

// RemoveMCPServer removes an MCP server configuration
func (e *ClaudeDesktopEditor) RemoveMCPServer(name string) error {
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
func (e *ClaudeDesktopEditor) GetConfigPath() string {
	return e.path
}

// ListServers returns a list of configured MCP servers
func (e *ClaudeDesktopEditor) ListServers() map[string]MCPServer {
	servers := make(map[string]MCPServer)

	// Add go-go-mcp if configured
	if e.config.GoGoMCP.Command != "" {
		servers["go-go-mcp"] = e.config.GoGoMCP
	}

	// Add other MCP servers
	for name, server := range e.config.MCPServers {
		servers[name] = server
	}

	return servers
}
