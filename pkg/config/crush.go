package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// CrushMCPConfig represents the structure of Crush's MCP configuration
type CrushMCPConfig struct {
	MCP map[string]CrushMCPEntry `json:"mcp"`
}

// CrushMCPEntry represents an individual MCP entry in Crush configuration
type CrushMCPEntry struct {
	Type    string            `json:"type"`
	URL     string            `json:"url,omitempty"`
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
			MCP: make(map[string]CrushMCPEntry),
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

// ListServers returns all MCP entries as CommonServer objects
func (c *CrushEditor) ListServers() (map[string]types.CommonServer, error) {
	servers := make(map[string]types.CommonServer)

	for name, entry := range c.config.MCP {
		servers[name] = types.CommonServer{
			Name: name,
			URL:  entry.URL,
			// Convert headers to env for consistency with other editors
			Env:   entry.Headers,
			IsSSE: entry.Type == "http", // Crush uses "http" type for HTTP servers
		}
	}

	return servers, nil
}

// GetServer retrieves a specific server by name
func (c *CrushEditor) GetServer(name string) (types.CommonServer, bool, error) {
	entry, exists := c.config.MCP[name]
	if !exists {
		return types.CommonServer{}, false, nil
	}

	server := types.CommonServer{
		Name:  name,
		URL:   entry.URL,
		Env:   entry.Headers,
		IsSSE: entry.Type == "http",
	}

	return server, true, nil
}

// AddMCPServer adds a new MCP server to the configuration
func (c *CrushEditor) AddMCPServer(server types.CommonServer, overwrite bool) error {
	if _, exists := c.config.MCP[server.Name]; exists && !overwrite {
		return fmt.Errorf("server '%s' already exists", server.Name)
	}

	entry := CrushMCPEntry{
		Type:    "http", // Default to HTTP type for Crush
		URL:     server.URL,
		Headers: server.Env, // Use env as headers
	}

	c.config.MCP[server.Name] = entry
	return nil
}

// RemoveMCPServer removes an MCP server from the configuration
func (c *CrushEditor) RemoveMCPServer(name string) error {
	if _, exists := c.config.MCP[name]; !exists {
		return fmt.Errorf("server '%s' does not exist", name)
	}

	delete(c.config.MCP, name)
	return nil
}

// IsServerDisabled checks if a server is disabled (Crush doesn't have disabled concept, always false)
func (c *CrushEditor) IsServerDisabled(name string) (bool, error) {
	_, exists := c.config.MCP[name]
	if !exists {
		return false, fmt.Errorf("server '%s' does not exist", name)
	}
	// Crush doesn't have a concept of disabled servers, so always return false
	return false, nil
}

// EnableMCPServer enables a server (no-op for Crush since there's no disabled concept)
func (c *CrushEditor) EnableMCPServer(name string) error {
	if _, exists := c.config.MCP[name]; !exists {
		return fmt.Errorf("server '%s' does not exist", name)
	}
	// No-op since Crush doesn't have enabled/disabled concept
	return nil
}

// DisableMCPServer disables a server (no-op for Crush since there's no disabled concept)
func (c *CrushEditor) DisableMCPServer(name string) error {
	if _, exists := c.config.MCP[name]; !exists {
		return fmt.Errorf("server '%s' does not exist", name)
	}
	// No-op since Crush doesn't have enabled/disabled concept
	return nil
}

// ListDisabledServers returns the names of disabled servers (always empty for Crush)
func (c *CrushEditor) ListDisabledServers() ([]string, error) {
	// Crush doesn't have a concept of disabled servers, so always return empty slice
	return []string{}, nil
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
