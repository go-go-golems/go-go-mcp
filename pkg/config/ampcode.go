package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/pkg/errors"
	"github.com/tailscale/hujson"
)

// AmpCodeConfig represents the Amp configuration which contains MCP servers settings
type AmpCodeConfig struct {
	AmpMCPServers   map[string]AmpCodeMCPServer `json:"amp.mcpServers"`
	DisabledServers map[string]AmpCodeMCPServer `json:"amp.disabledServersConfig,omitempty"`
	DisabledTools   []string                    `json:"amp.tools.disable,omitempty"`
	// Other fields could be added as needed
}

// AmpCodeMCPServer represents a server configuration for Amp
type AmpCodeMCPServer struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	// Note: Disabled field is kept for backward compatibility during migration
	// but the new approach uses separate maps for enabled/disabled servers
	Disabled bool `json:"disabled,omitempty"`
}

// AmpCodeEditor manages the Amp user settings configuration
type AmpCodeEditor struct {
	config *AmpCodeConfig
	path   string
	// Store the full settings file contents to preserve all other settings
	settingsJSON map[string]interface{}
	// Store the original parsed hujson Value to preserve comments
	originalValue *hujson.Value
}

// GetAmpCodeConfigPath returns the path for the Amp settings file in Cursor
func GetAmpCodeConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}
	// Path for Cursor settings
	return filepath.Join(homeDir, ".config", "Cursor", "User", "settings.json"), nil
}

// GetAmpConfigPath returns the path for the standalone Amp settings file
func GetAmpConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}
	// Path for standalone Amp settings using XDG config directory
	return filepath.Join(homeDir, ".config", "amp", "settings.json"), nil
}

// NewAmpCodeEditor creates a new editor for the AmpCode configuration
func NewAmpCodeEditor(path string) (*AmpCodeEditor, error) {
	editor := &AmpCodeEditor{
		path: path,
	}

	// Try to load existing config
	if err := editor.load(); err != nil {
		// If file doesn't exist, create a new config
		if os.IsNotExist(err) {
			editor.config = &AmpCodeConfig{
				AmpMCPServers:   make(map[string]AmpCodeMCPServer),
				DisabledServers: make(map[string]AmpCodeMCPServer),
				DisabledTools:   []string{},
			}
			editor.settingsJSON = make(map[string]interface{})
			editor.settingsJSON["amp.mcpServers"] = editor.config.AmpMCPServers
			editor.settingsJSON["amp.tools.disable"] = editor.config.DisabledTools
		} else {
			return nil, fmt.Errorf("could not load config: %w", err)
		}
	}

	return editor, nil
}

// load reads the configuration from disk and handles migration from old format
func (e *AmpCodeEditor) load() error {
	data, err := os.ReadFile(e.path)
	if err != nil {
		return errors.Wrapf(err, "failed to read config file %s", e.path)
	}

	// Process VS Code's JSONC format with the tailscale/hujson library
	// which preserves comments and structure
	value, err := hujson.Parse(data)
	if err != nil {
		return errors.Wrapf(err, "could not parse JSONC file %s", e.path)
	}

	// Store the original Value for preserving comments during save
	e.originalValue = &value

	// Make a copy for standardized processing
	standardized := value.Clone()
	standardized.Standardize() // Removes comments and trailing commas for standard JSON

	// Parse the entire settings.json file
	settingsJSON := make(map[string]interface{})
	if err := json.Unmarshal(standardized.Pack(), &settingsJSON); err != nil {
		return errors.Wrapf(err, "could not parse config file %s", e.path)
	}

	// Store the full settings JSON
	e.settingsJSON = settingsJSON

	// Initialize config struct
	e.config = &AmpCodeConfig{
		AmpMCPServers:   make(map[string]AmpCodeMCPServer),
		DisabledServers: make(map[string]AmpCodeMCPServer),
	}

	// Extract amp.mcpServers section
	if mcpServers, ok := settingsJSON["amp.mcpServers"]; ok {
		// Re-marshal and unmarshal to get the correct structure
		data, err := json.Marshal(mcpServers)
		if err != nil {
			return errors.Wrap(err, "could not re-marshal mcpServers configuration")
		}

		mpServers := make(map[string]AmpCodeMCPServer)
		if err := json.Unmarshal(data, &mpServers); err != nil {
			return errors.Wrap(err, "could not parse mcpServers configuration")
		}

		// Check for servers with disabled flag and migrate them
		for name, server := range mpServers {
			if server.Disabled {
				// Remove the disabled flag for the new format
				server.Disabled = false
				e.config.DisabledServers[name] = server
			} else {
				e.config.AmpMCPServers[name] = server
			}
		}
	}

	// Extract amp.disabledServersConfig section if it exists (new format)
	if disabledServers, ok := settingsJSON["amp.disabledServersConfig"]; ok {
		// Re-marshal and unmarshal to get the correct structure
		data, err := json.Marshal(disabledServers)
		if err != nil {
			return errors.Wrap(err, "could not re-marshal disabled servers configuration")
		}

		disabledServersMap := make(map[string]AmpCodeMCPServer)
		if err := json.Unmarshal(data, &disabledServersMap); err != nil {
			return errors.Wrap(err, "could not parse disabled servers configuration")
		}

		// Merge with any migrated servers
		for name, server := range disabledServersMap {
			e.config.DisabledServers[name] = server
		}
	}

	// Extract amp.tools.disable section
	if disabledTools, ok := settingsJSON["amp.tools.disable"]; ok {
		// Re-marshal and unmarshal to get the correct structure
		data, err := json.Marshal(disabledTools)
		if err != nil {
			return errors.Wrap(err, "could not re-marshal disabled tools configuration")
		}

		var tools []string
		if err := json.Unmarshal(data, &tools); err != nil {
			return errors.Wrap(err, "could not parse disabled tools configuration")
		}

		e.config.DisabledTools = tools
	} else {
		// No disabled tools section found, initialize empty
		e.config.DisabledTools = []string{}
	}

	return nil
}

// Save writes the configuration to disk using the separate maps approach
func (e *AmpCodeEditor) Save() error {
	// Update the settings JSON with the current config
	e.settingsJSON["amp.mcpServers"] = e.config.AmpMCPServers
	e.settingsJSON["amp.disabledServersConfig"] = e.config.DisabledServers
	e.settingsJSON["amp.tools.disable"] = e.config.DisabledTools

	if e.originalValue == nil {
		// No original value with comments, just marshal the JSON
		data, err := json.MarshalIndent(e.settingsJSON, "", "    ")
		if err != nil {
			return errors.Wrap(err, "could not marshal config")
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(e.path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "could not create config directory %s", dir)
		}

		// Write the file
		if err := os.WriteFile(e.path, data, 0644); err != nil {
			return errors.Wrapf(err, "could not write config to %s", e.path)
		}

		return nil
	}

	// Instead of using Patch, we'll create a new hujson.Value and format it
	newData, err := json.MarshalIndent(e.settingsJSON, "", "    ")
	if err != nil {
		return errors.Wrap(err, "could not marshal config")
	}

	// Parse the new data
	newValue, err := hujson.Parse(newData)
	if err != nil {
		return errors.Wrap(err, "could not parse new config")
	}

	// Format it and replace the original value
	newValue.Format()
	e.originalValue = &newValue

	// Create directory if it doesn't exist
	dir := filepath.Dir(e.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrapf(err, "could not create config directory %s", dir)
	}

	// Write the formatted JSON with comments preserved
	if err := os.WriteFile(e.path, e.originalValue.Pack(), 0644); err != nil {
		return errors.Wrapf(err, "could not write config to %s", e.path)
	}

	return nil
}

// AddMCPServer adds or updates a server configuration using the CommonServer struct.
// If the server exists in disabled servers, it will be moved to enabled servers.
func (e *AmpCodeEditor) AddMCPServer(server types.CommonServer, overwrite bool) error {
	if e.config.AmpMCPServers == nil {
		e.config.AmpMCPServers = make(map[string]AmpCodeMCPServer)
	}
	if e.config.DisabledServers == nil {
		e.config.DisabledServers = make(map[string]AmpCodeMCPServer)
	}

	name := server.Name
	_, existsInEnabled := e.config.AmpMCPServers[name]
	_, existsInDisabled := e.config.DisabledServers[name]
	exists := existsInEnabled || existsInDisabled

	if exists && !overwrite {
		return fmt.Errorf("MCP server '%s' already exists. Use overwrite option to replace it", name)
	}

	ampServer := AmpCodeMCPServer{
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Env,
		// Note: Disabled field is kept for backward compatibility but not used in new approach
		Disabled: false,
	}

	// Remove from disabled servers if it exists there
	delete(e.config.DisabledServers, name)
	// Add to enabled servers
	e.config.AmpMCPServers[name] = ampServer

	return nil
}

// RemoveMCPServer removes an MCP server configuration
func (e *AmpCodeEditor) RemoveMCPServer(name string) error {
	if e.config.AmpMCPServers == nil {
		return fmt.Errorf("MCP server '%s' not found", name)
	}

	_, exists := e.config.AmpMCPServers[name]
	if !exists {
		return fmt.Errorf("MCP server '%s' not found", name)
	}

	delete(e.config.AmpMCPServers, name)
	return nil
}

// GetConfigPath returns the path to the configuration file
func (e *AmpCodeEditor) GetConfigPath() string {
	return e.path
}

// ListServers returns a map of all configured servers (both enabled and disabled) as CommonServer
func (e *AmpCodeEditor) ListServers() (map[string]types.CommonServer, error) {
	servers := make(map[string]types.CommonServer)

	// Add enabled servers
	for name, server := range e.config.AmpMCPServers {
		servers[name] = types.CommonServer{
			Name:    name,
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		}
	}

	// Add disabled servers
	for name, server := range e.config.DisabledServers {
		servers[name] = types.CommonServer{
			Name:    name,
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		}
	}

	return servers, nil
}

// EnableMCPServer enables a previously disabled MCP server by moving it from disabled to enabled servers
func (e *AmpCodeEditor) EnableMCPServer(name string) error {
	if e.config.DisabledServers == nil {
		return fmt.Errorf("MCP server '%s' not found", name)
	}

	server, exists := e.config.DisabledServers[name]
	if !exists {
		// Check if it's already enabled
		if _, enabledExists := e.config.AmpMCPServers[name]; enabledExists {
			return fmt.Errorf("server '%s' is already enabled", name)
		}
		return fmt.Errorf("MCP server '%s' not found", name)
	}

	// Remove from disabled servers
	delete(e.config.DisabledServers, name)
	// Add to enabled servers
	e.config.AmpMCPServers[name] = server

	return nil
}

// DisableMCPServer disables an MCP server by moving it from enabled to disabled servers
func (e *AmpCodeEditor) DisableMCPServer(name string) error {
	if e.config.AmpMCPServers == nil {
		return fmt.Errorf("MCP server '%s' not found", name)
	}

	server, exists := e.config.AmpMCPServers[name]
	if !exists {
		// Check if it's already disabled
		if _, disabledExists := e.config.DisabledServers[name]; disabledExists {
			return fmt.Errorf("server '%s' is already disabled", name)
		}
		return fmt.Errorf("MCP server '%s' not found", name)
	}

	// Remove from enabled servers
	delete(e.config.AmpMCPServers, name)
	// Add to disabled servers
	e.config.DisabledServers[name] = server

	return nil
}

// IsServerDisabled checks if a server is disabled by looking in the disabled servers map
func (e *AmpCodeEditor) IsServerDisabled(name string) (bool, error) {
	// Check if server exists in either map
	_, existsInEnabled := e.config.AmpMCPServers[name]
	_, existsInDisabled := e.config.DisabledServers[name]

	if !existsInEnabled && !existsInDisabled {
		return false, fmt.Errorf("server '%s' not found", name)
	}

	// Server is disabled if it exists in the disabled servers map
	return existsInDisabled, nil
}

// ListDisabledServers returns a list of disabled server names from the disabled servers map
func (e *AmpCodeEditor) ListDisabledServers() ([]string, error) {
	disabledServers := []string{}

	for name := range e.config.DisabledServers {
		disabledServers = append(disabledServers, name)
	}

	return disabledServers, nil
}

// GetServer retrieves a specific server configuration by name as CommonServer
// It checks both enabled and disabled servers
func (e *AmpCodeEditor) GetServer(name string) (types.CommonServer, bool, error) {
	if e.config.AmpMCPServers == nil && e.config.DisabledServers == nil {
		return types.CommonServer{}, false, nil
	}

	// Check enabled servers first
	if server, exists := e.config.AmpMCPServers[name]; exists {
		common := types.CommonServer{
			Name:    name,
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		}
		return common, true, nil
	}

	// Check disabled servers
	if server, exists := e.config.DisabledServers[name]; exists {
		common := types.CommonServer{
			Name:    name,
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		}
		return common, true, nil
	}

	return types.CommonServer{}, false, nil
}
