package configstore

import (
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/spf13/viper"
)

// ProfilesStore implements the ProfileStore interface for profile configuration
type ProfilesStore struct {
	editor *config.ConfigEditor
	path   string
}

// NewProfilesStore creates a new profiles configuration store
func NewProfilesStore() (*ProfilesStore, error) {
	configFile, err := config.GetProfilesPath(viper.ConfigFileUsed())
	if err != nil {
		return nil, err
	}

	editor, err := config.NewConfigEditor(configFile)
	if err != nil {
		return nil, err
	}

	return &ProfilesStore{
		editor: editor,
		path:   configFile,
	}, nil
}

// Load loads the configuration
func (s *ProfilesStore) Load() error {
	// The editor automatically loads on creation, so this is a no-op
	return nil
}

// Save saves the configuration
func (s *ProfilesStore) Save() error {
	return s.editor.Save()
}

// GetProfiles returns all profiles
func (s *ProfilesStore) GetProfiles() (map[string]string, error) {
	return s.editor.GetProfiles()
}

// GetDefaultProfile returns the default profile name
func (s *ProfilesStore) GetDefaultProfile() (string, error) {
	return s.editor.GetDefaultProfile()
}

// AddProfile adds a new profile
func (s *ProfilesStore) AddProfile(name, description string) error {
	return s.editor.AddProfile(name, description)
}

// DeleteProfile removes a profile
func (s *ProfilesStore) DeleteProfile(name string) error {
	return s.editor.DeleteProfile(name)
}

// SetDefaultProfile sets the default profile
func (s *ProfilesStore) SetDefaultProfile(name string) error {
	return s.editor.SetDefaultProfile(name)
}

// AddToolDirectory adds a tool directory to a profile
func (s *ProfilesStore) AddToolDirectory(profileName, dir string, config map[string]interface{}) error {
	return s.editor.AddToolDirectory(profileName, dir, config)
}

// AddToolFile adds a tool file to a profile
func (s *ProfilesStore) AddToolFile(profileName, file string) error {
	return s.editor.AddToolFile(profileName, file)
}

// AddPromptDirectory adds a prompt directory to a profile
func (s *ProfilesStore) AddPromptDirectory(profileName, dir string, config map[string]interface{}) error {
	return s.editor.AddPromptDirectory(profileName, dir, config)
}

// AddPromptFile adds a prompt file to a profile
func (s *ProfilesStore) AddPromptFile(profileName, file string) error {
	return s.editor.AddPromptFile(profileName, file)
}
