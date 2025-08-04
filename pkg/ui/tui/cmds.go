package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// --- String Parsing/Formatting Helpers ---

// parseArgsString converts a space-separated string into a slice of args
func parseArgsString(argsStr string) []string {
	argsStr = strings.TrimSpace(argsStr)
	if argsStr == "" {
		return []string{}
	}
	// TODO: Handle quoted arguments properly if needed
	return strings.Fields(argsStr)
}

// parseEnvString converts newline-separated KEY=VALUE pairs into a map
func parseEnvString(envStr string) map[string]string {
	envMap := make(map[string]string)
	envStr = strings.TrimSpace(envStr)
	if envStr == "" {
		return envMap
	}

	// Handle both newline and carriage return + newline
	envStr = strings.ReplaceAll(envStr, "\r\n", "\n")

	lines := strings.Split(envStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") { // Allow comments
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Consider logging a warning for invalid lines
			log.Warn().Str("line", line).Msg("Invalid environment variable format, expected KEY=VALUE")
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" {
			envMap[key] = value
		}
	}

	return envMap
}

// --- Commands ---

// loadServers returns a command that loads the appropriate config and lists servers.
func (m *Model) loadServers(configType ConfigType) tea.Cmd { // Use ConfigType enum
	return func() tea.Msg {
		var editor types.ServerConfigEditor
		var err error
		var configPath string

		// Use switch statement for clarity
		switch configType {
		case ConfigTypeNone:
			err = fmt.Errorf("unknown or unsupported config type: %s", configType)
		case ConfigTypeCursor:
			configPath, err = config.GetGlobalCursorMCPConfigPath()
			if err == nil {
				editor, err = config.NewCursorMCPEditor(configPath)
			}
		case ConfigTypeClaude:
			configPath, err = config.GetDefaultClaudeDesktopConfigPath()
			if err == nil {
				editor, err = config.NewClaudeDesktopEditor(configPath)
			}
		case ConfigTypeAmpCode:
			configPath, err = config.GetAmpCodeConfigPath()
			if err == nil {
				editor, err = config.NewAmpCodeEditor(configPath)
			}
		case ConfigTypeAmp:
			configPath, err = config.GetAmpConfigPath()
			if err == nil {
				editor, err = config.NewAmpCodeEditor(configPath)
			}
		case ConfigTypeCrushLocal:
			configPath = ".crush.json"
			editor, err = config.NewCrushEditor(configPath)
		case ConfigTypeCrushCwd:
			configPath = "crush.json"
			editor, err = config.NewCrushEditor(configPath)
		case ConfigTypeCrushGlobal:
			configPath = viper.GetString("HOME") + "/.config/crush/crush.json"
			if configPath == "/.config/crush/crush.json" {
				// Fallback if HOME is not set
				var homeDir string
				homeDir, err = os.UserHomeDir()
				if err == nil {
					configPath = filepath.Join(homeDir, ".config", "crush", "crush.json")
				}
			}
			if err == nil {
				editor, err = config.NewCrushEditor(configPath)
			}
		case ConfigTypeProfile:
			// Profile config type doesn't use the server config editor
			// so we return an appropriate error
			err = fmt.Errorf("profile config type doesn't use server config editor")
		default: // Handles any other unexpected values
			err = fmt.Errorf("unknown or unsupported config type: %s", configType)
		}

		if err != nil {
			return loadedServersMsg{err: fmt.Errorf("failed to initialize editor: %w", err)}
		}

		servers, err := editor.ListServers()
		if err != nil {
			return loadedServersMsg{err: fmt.Errorf("failed to list servers: %w", err)}
		}

		return loadedServersMsg{
			editor:     editor,
			servers:    servers,
			configType: configType, // Pass the enum value
			err:        nil,
		}
	}
}

// deleteServer returns a command to delete the named server.
func (m *Model) deleteServer(name string) tea.Cmd {
	return func() tea.Msg {
		if m.currentEditor == nil {
			return serverDeletedMsg{serverName: name, err: fmt.Errorf("no editor loaded")}
		}
		err := m.currentEditor.RemoveMCPServer(name)
		if err != nil {
			return serverDeletedMsg{serverName: name, err: err}
		}
		err = m.currentEditor.Save()
		if err != nil {
			// Log the save error, but report deletion success to the user
			log.Error().Err(err).Msg("Failed to save config after deleting server")
			return serverDeletedMsg{serverName: name, err: fmt.Errorf("failed to save config after deletion: %w", err)}
		}
		return serverDeletedMsg{serverName: name, err: nil}
	}
}

// toggleServerEnabled returns a command to toggle the enabled state.
func (m *Model) toggleServerEnabled(name string) tea.Cmd {
	return func() tea.Msg {
		if m.currentEditor == nil {
			return serverToggleEnabledMsg{serverName: name, err: fmt.Errorf("no editor loaded")}
		}

		isDisabled, err := m.currentEditor.IsServerDisabled(name)
		if err != nil {
			return serverToggleEnabledMsg{serverName: name, err: fmt.Errorf("failed to check server status: %w", err)}
		}

		var toggleErr error
		newStateEnabled := false
		if isDisabled {
			toggleErr = m.currentEditor.EnableMCPServer(name)
			newStateEnabled = true
		} else {
			toggleErr = m.currentEditor.DisableMCPServer(name)
			newStateEnabled = false
		}

		if toggleErr != nil {
			return serverToggleEnabledMsg{serverName: name, err: fmt.Errorf("failed to toggle server: %w", toggleErr)}
		}

		saveErr := m.currentEditor.Save()
		if saveErr != nil {
			// Log save error, but report toggle success based on toggleErr
			log.Error().Err(saveErr).Msg("Failed to save config after toggling server")
			// Return toggle success but include save error info
			return serverToggleEnabledMsg{serverName: name, enabled: newStateEnabled, success: true, err: fmt.Errorf("failed to save config: %w", saveErr)}
		}

		return serverToggleEnabledMsg{serverName: name, enabled: newStateEnabled, success: true, err: nil}
	}
}

// saveServer returns a command to add/update a server and save the config.
func (m *Model) saveServer(server types.CommonServer, overwrite bool) tea.Cmd {
	return func() tea.Msg {
		log.Debug().Str("serverName", server.Name).Bool("overwrite", overwrite).Msg("Saving server")
		if m.currentEditor == nil {
			log.Debug().Msg("No editor loaded, returning error")
			return serverSavedMsg{serverName: server.Name, err: fmt.Errorf("no editor loaded")}
		}

		log.Debug().Str("serverName", server.Name).Bool("overwrite", overwrite).Msg("Adding server")
		err := m.currentEditor.AddMCPServer(server, overwrite)
		if err != nil {
			log.Debug().Msg("Error adding server, returning error")
			// If error is about existing server and not overwriting, maybe trigger confirmation dialog?
			// For now, just return the error.
			return serverSavedMsg{serverName: server.Name, err: err}
		}

		err = m.currentEditor.Save()
		if err != nil {
			log.Debug().Msg("Error saving config, returning error")
			return serverSavedMsg{serverName: server.Name, err: fmt.Errorf("failed to save config after adding/updating server: %w", err)}
		}

		return serverSavedMsg{serverName: server.Name, err: nil}
	}
}

// loadProfiles attempts to load profiles from the config file
func (m *Model) loadProfiles() tea.Cmd {
	return func() tea.Msg {
		configFile, err := config.GetProfilesPath(viper.ConfigFileUsed())
		if err != nil {
			return loadedProfilesMsg{err: fmt.Errorf("could not get profiles path: %w", err)}
		}

		editor, err := config.NewConfigEditor(configFile)
		if err != nil {
			return loadedProfilesMsg{err: fmt.Errorf("could not create config editor: %w", err)}
		}

		profiles, err := editor.GetProfiles()
		if err != nil {
			return loadedProfilesMsg{err: fmt.Errorf("could not get profiles: %w", err)}
		}

		defaultProfile, err := editor.GetDefaultProfile()
		// If we can't get the default profile, we'll still return the profiles but with an empty default
		if err != nil {
			defaultProfile = ""
		}

		return loadedProfilesMsg{
			editor:         editor,
			profiles:       profiles,
			defaultProfile: defaultProfile,
			err:            nil,
		}
	}
}

// saveProfile adds or updates a profile in the config
func (m *Model) saveProfile(name, description string, toolDirs, toolFiles, promptDirs, promptFiles []string, isNewProfile bool) tea.Cmd {
	return func() tea.Msg {
		if m.profileEditor == nil {
			return profileSavedMsg{err: fmt.Errorf("no profile editor initialized")}
		}

		var err error
		if isNewProfile {
			// Add new profile
			err = m.profileEditor.AddProfile(name, description)
			if err != nil {
				return profileSavedMsg{err: fmt.Errorf("could not add profile: %w", err)}
			}
		} else {
			// For editing, we need to create a new profile and delete the old one
			// Since there's no direct "edit description" function in the editor
			oldProfiles, err := m.profileEditor.GetProfiles()
			if err != nil {
				return profileSavedMsg{err: fmt.Errorf("could not get profiles: %w", err)}
			}

			// If the profile exists, delete and recreate it
			if _, exists := oldProfiles[name]; exists {
				// Delete the existing profile first
				err = m.profileEditor.DeleteProfile(name)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not delete existing profile: %w", err)}
				}

				// Now create the profile with the proper name and new description
				err = m.profileEditor.AddProfile(name, description)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not recreate profile: %w", err)}
				}
			} else {
				// If it doesn't exist, just create it
				err = m.profileEditor.AddProfile(name, description)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not create profile: %w", err)}
				}
			}
		}

		// Add tool directories if provided
		for _, dir := range toolDirs {
			if dir != "" {
				err = m.profileEditor.AddToolDirectory(name, dir, map[string]interface{}{})
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not add tool directory %s: %w", dir, err)}
				}
			}
		}

		// Add tool files if provided
		for _, file := range toolFiles {
			if file != "" {
				err = m.profileEditor.AddToolFile(name, file)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not add tool file %s: %w", file, err)}
				}
			}
		}

		// Add prompt directories if provided
		for _, dir := range promptDirs {
			if dir != "" {
				err = m.profileEditor.AddPromptDirectory(name, dir, map[string]interface{}{})
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not add prompt directory %s: %w", dir, err)}
				}
			}
		}

		// Add prompt files if provided
		for _, file := range promptFiles {
			if file != "" {
				err = m.profileEditor.AddPromptFile(name, file)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not add prompt file %s: %w", file, err)}
				}
			}
		}

		// Save changes to the config file
		err = m.profileEditor.Save()
		if err != nil {
			return profileSavedMsg{err: fmt.Errorf("could not save config: %w", err)}
		}

		return profileSavedMsg{
			profileName: name,
			err:         nil,
		}
	}
}

// deleteProfile removes a profile from the config
func (m *Model) deleteProfile(name string) tea.Cmd {
	return func() tea.Msg {
		if m.profileEditor == nil {
			return profileDeletedMsg{err: fmt.Errorf("no profile editor initialized")}
		}

		// Delete the profile
		if err := m.profileEditor.DeleteProfile(name); err != nil {
			return profileDeletedMsg{profileName: name, err: err}
		}

		// Save the changes
		if err := m.profileEditor.Save(); err != nil {
			return profileDeletedMsg{profileName: name, err: fmt.Errorf("could not save after deleting: %w", err)}
		}

		return profileDeletedMsg{profileName: name, err: nil}
	}
}

// setDefaultProfile sets the default profile in the config
func (m *Model) setDefaultProfile(name string) tea.Cmd {
	return func() tea.Msg {
		if m.profileEditor == nil {
			return defaultProfileSetMsg{err: fmt.Errorf("no profile editor initialized")}
		}

		err := m.profileEditor.SetDefaultProfile(name)
		if err != nil {
			return defaultProfileSetMsg{err: fmt.Errorf("could not set default profile: %w", err)}
		}

		// Save changes to the config file
		err = m.profileEditor.Save()
		if err != nil {
			return defaultProfileSetMsg{err: fmt.Errorf("could not save config: %w", err)}
		}

		return defaultProfileSetMsg{
			profileName: name,
			err:         nil,
		}
	}
}
