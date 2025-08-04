package configstore

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// Store provides a common interface for managing configuration stores
type Store interface {
	// Load loads the current state from the configuration
	Load() error

	// Save saves the current state to the configuration
	Save() error

	// ListServers returns all servers in the configuration
	ListServers() (map[string]types.CommonServer, error)

	// GetServer retrieves a specific server by name
	GetServer(name string) (types.CommonServer, bool, error)

	// AddServer adds or updates a server
	AddServer(server types.CommonServer, overwrite bool) error

	// RemoveServer removes a server by name
	RemoveServer(name string) error

	// IsServerDisabled checks if a server is disabled
	IsServerDisabled(name string) (bool, error)

	// EnableServer enables a server
	EnableServer(name string) error

	// DisableServer disables a server
	DisableServer(name string) error
}

// ProfileStore provides interface for managing profile configurations
type ProfileStore interface {
	// Load loads the current profiles from configuration
	Load() error

	// Save saves the current profiles to configuration
	Save() error

	// GetProfiles returns all profiles
	GetProfiles() (map[string]string, error)

	// GetDefaultProfile returns the default profile name
	GetDefaultProfile() (string, error)

	// AddProfile adds a new profile
	AddProfile(name, description string) error

	// DeleteProfile removes a profile
	DeleteProfile(name string) error

	// SetDefaultProfile sets the default profile
	SetDefaultProfile(name string) error

	// AddToolDirectory adds a tool directory to a profile
	AddToolDirectory(profileName, dir string, config map[string]interface{}) error

	// AddToolFile adds a tool file to a profile
	AddToolFile(profileName, file string) error

	// AddPromptDirectory adds a prompt directory to a profile
	AddPromptDirectory(profileName, dir string, config map[string]interface{}) error

	// AddPromptFile adds a prompt file to a profile
	AddPromptFile(profileName, file string) error
}

// ConfigType represents the type of configuration
type ConfigType string

const (
	ConfigTypeCursor      ConfigType = "cursor"
	ConfigTypeClaude      ConfigType = "claude"
	ConfigTypeAmpCode     ConfigType = "ampcode"
	ConfigTypeAmp         ConfigType = "amp"
	ConfigTypeProfile     ConfigType = "profile"
	ConfigTypeCrushLocal  ConfigType = "crush-local"
	ConfigTypeCrushCwd    ConfigType = "crush-cwd"
	ConfigTypeCrushGlobal ConfigType = "crush-global"
	ConfigTypeNone        ConfigType = ""
)

// NewStore creates a new Store based on the config type
func NewStore(configType ConfigType) (Store, error) {
	switch configType {
	case ConfigTypeCursor:
		return NewCursorStore()
	case ConfigTypeClaude:
		return NewClaudeStore()
	case ConfigTypeAmpCode:
		return NewAmpCodeStore()
	case ConfigTypeAmp:
		return NewAmpStore()
	case ConfigTypeCrushLocal:
		return NewCrushLocalStore()
	case ConfigTypeCrushCwd:
		return NewCrushCwdStore()
	case ConfigTypeCrushGlobal:
		return NewCrushGlobalStore()
	case ConfigTypeProfile:
		return nil, fmt.Errorf("profile config type doesn't use server config editor")
	case ConfigTypeNone:
		return nil, fmt.Errorf("no config type specified")
	default:
		return nil, fmt.Errorf("unsupported config type: %s", configType)
	}
}

// NewProfileStore creates a new ProfileStore
func NewProfileStore() (ProfileStore, error) {
	return NewProfilesStore()
}
