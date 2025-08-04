package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// CrushMCPConfig represents the structure of Crush's MCP configuration
// using separate maps for enabled and disabled servers
type CrushMCPConfig struct {
	MCP             map[string]CrushMCPEntry `json:"mcp"`
	DisabledServers map[string]CrushMCPEntry `json:"disabledServers,omitempty"`
}

// CrushMCPEntry represents an individual MCP entry in Crush configuration
type CrushMCPEntry struct {
	Type    string            `json:"type"`
	URL     string            `json:"url,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// CrushEditor implements ServerConfigEditor for Crush JSON configuration
type CrushEditor struct {
	filePath string
	config   *CrushMCPConfig
}

// Ensure CrushEditor implements the ServerConfigEditor interface
var _ types.ServerConfigEditor = &CrushEditor{}

// NewCrushEditor creates a new CrushEditor for the given file path
func NewCrushEditor(filePath string) (*CrushEditor, error) {
	editor := &CrushEditor{
		filePath: filePath,
		config: &CrushMCPConfig{
			MCP:             make(map[string]CrushMCPEntry),
			DisabledServers: make(map[string]CrushMCPEntry),
		},
	}

	// Try to load existing config
	if _, err := os.Stat(filePath); err == nil {
		if err := editor.load(); err != nil {
			return nil, fmt.Errorf("failed to load existing config: %w", err)
		}
	}

	return editor, nil
}

// load reads and parses the JSON configuration file
func (c *CrushEditor) load() error {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, c.config); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Initialize maps if they're nil
	if c.config.MCP == nil {
		c.config.MCP = make(map[string]CrushMCPEntry)
	}
	if c.config.DisabledServers == nil {
		c.config.DisabledServers = make(map[string]CrushMCPEntry)
	}

	return nil
}

// Save writes the configuration to the file
func (c *CrushEditor) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal with pretty printing
	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ListServers returns all MCP entries (both enabled and disabled) as CommonServer objects
func (c *CrushEditor) ListServers() (map[string]types.CommonServer, error) {
	servers := make(map[string]types.CommonServer)

	// Add enabled servers
	for name, entry := range c.config.MCP {
		server := types.CommonServer{
			Name:    name,
			Command: entry.Command,
			Args:    entry.Args,
			URL:     entry.URL,
		}

		// Map fields based on transport type
		switch entry.Type {
		case "stdio":
			server.Env = entry.Env
			server.IsSSE = false
		case "http":
			server.Env = entry.Headers // Headers map to Env for http
			server.IsSSE = false
		case "sse":
			server.Env = entry.Headers // Headers map to Env for sse
			server.IsSSE = true
		default:
			// Legacy fallback - assume http if no type specified
			server.Env = entry.Headers
			server.IsSSE = false
		}

		servers[name] = server
	}

	// Add disabled servers
	for name, entry := range c.config.DisabledServers {
		if _, exists := servers[name]; !exists {
			server := types.CommonServer{
				Name:    name,
				Command: entry.Command,
				Args:    entry.Args,
				URL:     entry.URL,
			}

			// Map fields based on transport type
			switch entry.Type {
			case "stdio":
				server.Env = entry.Env
				server.IsSSE = false
			case "http":
				server.Env = entry.Headers // Headers map to Env for http
				server.IsSSE = false
			case "sse":
				server.Env = entry.Headers // Headers map to Env for sse
				server.IsSSE = true
			default:
				// Legacy fallback - assume http if no type specified
				server.Env = entry.Headers
				server.IsSSE = false
			}

			servers[name] = server
		}
	}

	return servers, nil
}

// GetServer retrieves a specific server by name from either enabled or disabled servers
func (c *CrushEditor) GetServer(name string) (types.CommonServer, bool, error) {
	// Check enabled servers first
	entry, exists := c.config.MCP[name]
	if !exists {
		// Check disabled servers
		entry, exists = c.config.DisabledServers[name]
		if !exists {
			return types.CommonServer{}, false, nil
		}
	}

	server := types.CommonServer{
		Name:    name,
		Command: entry.Command,
		Args:    entry.Args,
		URL:     entry.URL,
	}

	// Map fields based on transport type
	switch entry.Type {
	case "stdio":
		server.Env = entry.Env
		server.IsSSE = false
	case "http":
		server.Env = entry.Headers
		server.IsSSE = false
	case "sse":
		server.Env = entry.Headers
		server.IsSSE = true
	default:
		// Legacy fallback
		server.Env = entry.Headers
		server.IsSSE = false
	}

	return server, true, nil
}

// AddMCPServer adds a new MCP server to the configuration (enabled by default)
func (c *CrushEditor) AddMCPServer(server types.CommonServer, overwrite bool) error {
	_, existsInEnabled := c.config.MCP[server.Name]
	_, existsInDisabled := c.config.DisabledServers[server.Name]
	
	if (existsInEnabled || existsInDisabled) && !overwrite {
		return fmt.Errorf("server '%s' already exists", server.Name)
	}

	var entry CrushMCPEntry

	// Determine the type based on CommonServer fields
	if server.Command != "" && server.URL == "" {
		// Stdio server
		entry = CrushMCPEntry{
			Type:    "stdio",
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		}
	} else if server.URL != "" && server.IsSSE {
		// SSE server
		entry = CrushMCPEntry{
			Type:    "sse",
			URL:     server.URL,
			Headers: server.Env,
		}
	} else if server.URL != "" && !server.IsSSE {
		// HTTP server
		entry = CrushMCPEntry{
			Type:    "http",
			URL:     server.URL,
			Headers: server.Env,
		}
	} else {
		return fmt.Errorf("invalid server configuration: must have either command (stdio) or URL (http/sse)")
	}

	// Remove from disabled servers if it exists there
	delete(c.config.DisabledServers, server.Name)
	// Add to enabled servers
	c.config.MCP[server.Name] = entry
	return nil
}

// RemoveMCPServer removes an MCP server from the configuration (both enabled and disabled lists)
func (c *CrushEditor) RemoveMCPServer(name string) error {
	_, existsInEnabled := c.config.MCP[name]
	_, existsInDisabled := c.config.DisabledServers[name]
	
	if !existsInEnabled && !existsInDisabled {
		return fmt.Errorf("server '%s' does not exist", name)
	}

	// Remove from both maps (in case it exists in either)
	delete(c.config.MCP, name)
	delete(c.config.DisabledServers, name)
	return nil
}

// IsServerDisabled checks if a server is disabled by looking in the disabled servers map
func (c *CrushEditor) IsServerDisabled(name string) (bool, error) {
	_, existsInEnabled := c.config.MCP[name]
	_, existsInDisabled := c.config.DisabledServers[name]
	
	if !existsInEnabled && !existsInDisabled {
		return false, fmt.Errorf("server '%s' does not exist", name)
	}
	
	return existsInDisabled, nil
}

// EnableMCPServer enables a previously disabled server by moving it from disabled to enabled servers
func (c *CrushEditor) EnableMCPServer(name string) error {
	server, exists := c.config.DisabledServers[name]
	if !exists {
		// Check if it's already enabled
		if _, enabledExists := c.config.MCP[name]; enabledExists {
			return fmt.Errorf("server '%s' is already enabled", name)
		}
		return fmt.Errorf("server '%s' not found in disabled servers", name)
	}

	// Move from disabled to enabled
	delete(c.config.DisabledServers, name)
	c.config.MCP[name] = server
	return nil
}

// DisableMCPServer disables a server by moving it from enabled to disabled servers
func (c *CrushEditor) DisableMCPServer(name string) error {
	server, exists := c.config.MCP[name]
	if !exists {
		// Check if it's already disabled
		if _, disabledExists := c.config.DisabledServers[name]; disabledExists {
			return fmt.Errorf("server '%s' is already disabled", name)
		}
		return fmt.Errorf("enabled server '%s' not found", name)
	}

	// Move from enabled to disabled
	delete(c.config.MCP, name)
	c.config.DisabledServers[name] = server
	return nil
}

// ListDisabledServers returns the names of disabled servers
func (c *CrushEditor) ListDisabledServers() ([]string, error) {
	disabledServers := make([]string, 0, len(c.config.DisabledServers))
	for name := range c.config.DisabledServers {
		disabledServers = append(disabledServers, name)
	}
	return disabledServers, nil
}

// GetConfigPath returns the path of the configuration file being managed
func (c *CrushEditor) GetConfigPath() string {
	return c.filePath
}

// GetCrushConfigPaths returns the priority-ordered list of Crush config file paths
func GetCrushConfigPaths() []string {
	homeDir, _ := os.UserHomeDir()
	return []string{
		".crush.json",
		"crush.json",
		filepath.Join(homeDir, ".config", "crush", "crush.json"),
	}
}

// GetCrushConfigPath returns the first existing Crush config file, or the default if none exist
func GetCrushConfigPath() (string, error) {
	paths := GetCrushConfigPaths()

	// Check for existing files in priority order
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
			}
			return absPath, nil
		}
	}

	// Return the first (highest priority) path if none exist
	absPath, err := filepath.Abs(paths[0])
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", paths[0], err)
	}
	return absPath, nil
}
