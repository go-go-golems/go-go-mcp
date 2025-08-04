package tui

// Define the different UI modes
type mode int

const (
	modeMenu mode = iota
	modeSubmenu
	modeList
	modeAddEdit
	modeConfirm
)

// Define configuration types
type ConfigType string

const (
	ConfigTypeCursor      ConfigType = "cursor"
	ConfigTypeClaude      ConfigType = "claude"
	ConfigTypeAmpCode     ConfigType = "ampcode"      // Configuration for Amp (Cursor)
	ConfigTypeAmp         ConfigType = "amp"          // Configuration for standalone Amp
	ConfigTypeProfile     ConfigType = "profile"      // New config type for profiles
	ConfigTypeCrushLocal  ConfigType = "crush-local"  // .crush.json
	ConfigTypeCrushCwd    ConfigType = "crush-cwd"    // crush.json
	ConfigTypeCrushGlobal ConfigType = "crush-global" // ~/.config/crush/crush.json
	ConfigTypeNone        ConfigType = ""             // Represents no config loaded
)

// Menu hierarchy types
type MenuType string

const (
	MenuTypeClaude   MenuType = "claude"
	MenuTypeCursor   MenuType = "cursor"
	MenuTypeAmpCode  MenuType = "ampcode"
	MenuTypeCrush    MenuType = "crush"
	MenuTypeProfiles MenuType = "profiles"
)
