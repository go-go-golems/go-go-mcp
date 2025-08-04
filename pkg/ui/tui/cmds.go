package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-mcp/pkg/core/configstore"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/rs/zerolog/log"
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

// --- New Domain Store Commands ---

// LoadServersCmd returns a command that loads servers using configstore
func LoadServersCmd(configType configstore.ConfigType) tea.Cmd {
	return func() tea.Msg {
		if configType == configstore.ConfigTypeProfile {
			return loadedServersMsg{err: fmt.Errorf("profile config type doesn't use server config editor")}
		}

		store, err := configstore.NewStore(configType)
		if err != nil {
			return loadedServersMsg{err: fmt.Errorf("failed to create store: %w", err)}
		}

		servers, err := store.ListServers()
		if err != nil {
			return loadedServersMsg{err: fmt.Errorf("failed to list servers: %w", err)}
		}

		return loadedServersMsg{
			servers:    servers,
			configType: ConfigType(configType), // Convert to TUI config type
			err:        nil,
		}
	}
}

// SaveServerCmd returns a command to save a server using configstore
func SaveServerCmd(store configstore.Store, server types.CommonServer, overwrite bool) tea.Cmd {
	return func() tea.Msg {
		log.Debug().Str("serverName", server.Name).Bool("overwrite", overwrite).Msg("Saving server")

		err := store.AddServer(server, overwrite)
		if err != nil {
			log.Debug().Msg("Error adding server, returning error")
			return serverSavedMsg{serverName: server.Name, err: err}
		}

		err = store.Save()
		if err != nil {
			log.Debug().Msg("Error saving config, returning error")
			return serverSavedMsg{serverName: server.Name, err: fmt.Errorf("failed to save config after adding/updating server: %w", err)}
		}

		return serverSavedMsg{serverName: server.Name, err: nil}
	}
}

// DeleteServerCmd returns a command to delete a server using configstore
func DeleteServerCmd(store configstore.Store, name string) tea.Cmd {
	return func() tea.Msg {
		err := store.RemoveServer(name)
		if err != nil {
			return serverDeletedMsg{serverName: name, err: err}
		}

		err = store.Save()
		if err != nil {
			log.Error().Err(err).Msg("Failed to save config after deleting server")
			return serverDeletedMsg{serverName: name, err: fmt.Errorf("failed to save config after deletion: %w", err)}
		}

		return serverDeletedMsg{serverName: name, err: nil}
	}
}

// ToggleServerEnabledCmd returns a command to toggle server enabled state using configstore
func ToggleServerEnabledCmd(store configstore.Store, name string) tea.Cmd {
	return func() tea.Msg {
		isDisabled, err := store.IsServerDisabled(name)
		if err != nil {
			return serverToggleEnabledMsg{serverName: name, err: fmt.Errorf("failed to check server status: %w", err)}
		}

		var toggleErr error
		newStateEnabled := false
		if isDisabled {
			toggleErr = store.EnableServer(name)
			newStateEnabled = true
		} else {
			toggleErr = store.DisableServer(name)
			newStateEnabled = false
		}

		if toggleErr != nil {
			return serverToggleEnabledMsg{serverName: name, err: fmt.Errorf("failed to toggle server: %w", toggleErr)}
		}

		saveErr := store.Save()
		if saveErr != nil {
			log.Error().Err(saveErr).Msg("Failed to save config after toggling server")
			return serverToggleEnabledMsg{serverName: name, enabled: newStateEnabled, success: true, err: fmt.Errorf("failed to save config: %w", saveErr)}
		}

		return serverToggleEnabledMsg{serverName: name, enabled: newStateEnabled, success: true, err: nil}
	}
}

// LoadProfilesCmd returns a command that loads profiles using configstore
func LoadProfilesCmd() tea.Cmd {
	return func() tea.Msg {
		store, err := configstore.NewProfileStore()
		if err != nil {
			return loadedProfilesMsg{err: fmt.Errorf("could not create profile store: %w", err)}
		}

		profiles, err := store.GetProfiles()
		if err != nil {
			return loadedProfilesMsg{err: fmt.Errorf("could not get profiles: %w", err)}
		}

		defaultProfile, err := store.GetDefaultProfile()
		if err != nil {
			defaultProfile = ""
		}

		return loadedProfilesMsg{
			profiles:       profiles,
			defaultProfile: defaultProfile,
			err:            nil,
		}
	}
}

// SaveProfileCmd returns a command to save a profile using configstore
func SaveProfileCmd(store configstore.ProfileStore, name, description string, toolDirs, toolFiles, promptDirs, promptFiles []string, isNewProfile bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if isNewProfile {
			err = store.AddProfile(name, description)
			if err != nil {
				return profileSavedMsg{err: fmt.Errorf("could not add profile: %w", err)}
			}
		} else {
			profiles, err := store.GetProfiles()
			if err != nil {
				return profileSavedMsg{err: fmt.Errorf("could not get profiles: %w", err)}
			}

			if _, exists := profiles[name]; exists {
				err = store.DeleteProfile(name)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not delete existing profile: %w", err)}
				}

				err = store.AddProfile(name, description)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not recreate profile: %w", err)}
				}
			} else {
				err = store.AddProfile(name, description)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not create profile: %w", err)}
				}
			}
		}

		// Add tool directories
		for _, dir := range toolDirs {
			if dir != "" {
				err = store.AddToolDirectory(name, dir, map[string]interface{}{})
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not add tool directory %s: %w", dir, err)}
				}
			}
		}

		// Add tool files
		for _, file := range toolFiles {
			if file != "" {
				err = store.AddToolFile(name, file)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not add tool file %s: %w", file, err)}
				}
			}
		}

		// Add prompt directories
		for _, dir := range promptDirs {
			if dir != "" {
				err = store.AddPromptDirectory(name, dir, map[string]interface{}{})
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not add prompt directory %s: %w", dir, err)}
				}
			}
		}

		// Add prompt files
		for _, file := range promptFiles {
			if file != "" {
				err = store.AddPromptFile(name, file)
				if err != nil {
					return profileSavedMsg{err: fmt.Errorf("could not add prompt file %s: %w", file, err)}
				}
			}
		}

		err = store.Save()
		if err != nil {
			return profileSavedMsg{err: fmt.Errorf("could not save config: %w", err)}
		}

		return profileSavedMsg{
			profileName: name,
			err:         nil,
		}
	}
}

// DeleteProfileCmd returns a command to delete a profile using configstore
func DeleteProfileCmd(store configstore.ProfileStore, name string) tea.Cmd {
	return func() tea.Msg {
		if err := store.DeleteProfile(name); err != nil {
			return profileDeletedMsg{profileName: name, err: err}
		}

		if err := store.Save(); err != nil {
			return profileDeletedMsg{profileName: name, err: fmt.Errorf("could not save after deleting: %w", err)}
		}

		return profileDeletedMsg{profileName: name, err: nil}
	}
}

// SetDefaultProfileCmd returns a command to set the default profile using configstore
func SetDefaultProfileCmd(store configstore.ProfileStore, name string) tea.Cmd {
	return func() tea.Msg {
		err := store.SetDefaultProfile(name)
		if err != nil {
			return defaultProfileSetMsg{err: fmt.Errorf("could not set default profile: %w", err)}
		}

		err = store.Save()
		if err != nil {
			return defaultProfileSetMsg{err: fmt.Errorf("could not save config: %w", err)}
		}

		return defaultProfileSetMsg{
			profileName: name,
			err:         nil,
		}
	}
}
