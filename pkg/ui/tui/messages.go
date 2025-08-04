package tui

import (
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// --- Messages ---

// Message indicating servers have been loaded
type loadedServersMsg struct {
	servers    map[string]types.CommonServer
	configType ConfigType // Use the enum type
	err        error
}

// Message indicating a server was deleted
type serverDeletedMsg struct {
	serverName string
	err        error
}

// Message indicating a server's enabled state was toggled
type serverToggleEnabledMsg struct {
	serverName string
	enabled    bool
	success    bool // Indicate if the operation itself succeeded
	err        error
}

// Message indicating a server save operation completed
type serverSavedMsg struct {
	serverName string
	err        error
}

// Message for generic errors
type errorMsg struct{ err error }

// Helper for creating error messages
func (e errorMsg) Error() string { return e.err.Error() }

// Message indicating profiles have been loaded
type loadedProfilesMsg struct {
	profiles       map[string]string
	defaultProfile string
	err            error
}

// Message indicating a profile was saved
type profileSavedMsg struct {
	profileName string
	err         error
}

// Message indicating a profile was deleted
type profileDeletedMsg struct {
	profileName string
	err         error
}

// Message indicating the default profile was set
type defaultProfileSetMsg struct {
	profileName string
	err         error
}
