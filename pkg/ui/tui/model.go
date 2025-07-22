package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

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

// Define key bindings
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Back      key.Binding
	Quit      key.Binding
	Add       key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Duplicate key.Binding
	Enable    key.Binding
	Help      key.Binding
}

// Set up default key bindings
var defaultKeyMap = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/confirm"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back/cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit application"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add new server"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit selected server"),
	),
	Delete: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "delete selected server"),
	),
	Duplicate: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "duplicate selected server"),
	),
	Enable: key.NewBinding(
		key.WithKeys("space", " "),
		key.WithHelp("space", "toggle server enabled/disabled"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

// Server item model for the list view
type serverItem struct {
	name    string
	command string
	args    []string
	env     map[string]string
	url     string // for Cursor SSE servers
	enabled bool
	isSSE   bool // Added to distinguish server type
}

func (i serverItem) Title() string { return i.name }
func (i serverItem) Description() string {
	status := "enabled"
	if !i.enabled {
		status = "disabled"
	}

	serverType := "CMD"
	if i.isSSE {
		serverType = "SSE"
	}

	if i.url != "" {
		return fmt.Sprintf("%s: %s (%s)", serverType, i.url, status)
	}
	return fmt.Sprintf("%s: %s (%s)", serverType, i.command, status)
}
func (i serverItem) FilterValue() string { return i.name }

// Menu hierarchy types
type MenuType string

const (
	MenuTypeClaude   MenuType = "claude"
	MenuTypeCursor   MenuType = "cursor"
	MenuTypeAmpCode  MenuType = "ampcode"
	MenuTypeCrush    MenuType = "crush"
	MenuTypeProfiles MenuType = "profiles"
)

// Main application model
type Model struct {
	keys        keyMap
	help        help.Model
	mode        mode
	menuList    list.Model
	submenuList list.Model
	activeList  *list.Model // Pointer to the currently active server list
	width       int
	height      int

	// Current menu hierarchy
	currentMenuType MenuType
	breadcrumb      string

	// Configuration editor interface
	currentEditor types.ServerConfigEditor

	// Profile configuration editor
	profileEditor *config.ConfigEditor

	// Track which config type is being edited (used for display/loading)
	configType ConfigType // Use the enum type

	// Error message to display
	errorMsg string

	// Form state for add/edit
	formState FormModel

	// Profile form state
	profileFormState ProfileFormModel

	// Confirmation dialog
	confirmDialog    ConfirmModel
	confirmAction    string
	actionServerName string
}

// Simple list item for menu
type listItem struct {
	title       string
	description string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }

// --- Messages ---

// Message indicating servers have been loaded
type loadedServersMsg struct {
	editor     types.ServerConfigEditor
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
	editor         *config.ConfigEditor
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

// NewModel initializes and returns a new Model
func NewModel() Model {
	// Initialize help
	h := help.New()
	h.ShowAll = false // Start with short help

	// Create main menu items (hierarchical)
	items := []list.Item{
		listItem{title: "Claude Desktop", description: "Configure Claude Desktop MCP servers"},
		listItem{title: "Cursor", description: "Configure Cursor MCP servers"},
		listItem{title: "Amp Code", description: "Configure Amp Code MCP servers"},
		listItem{title: "Crush", description: "Configure Crush MCP servers"},
		listItem{title: "Profiles", description: "Configure MCP profiles"},
	}

	// Initialize the menu list
	menuDelegate := list.NewDefaultDelegate()
	menuList := list.New(items, menuDelegate, 0, 0)
	menuList.Title = "Go Go MCP Configuration"
	menuList.SetShowHelp(false) // We'll show our own help

	return Model{
		keys:             defaultKeyMap,
		help:             h,
		mode:             modeMenu,
		menuList:         menuList,
		formState:        NewFormModel(),
		profileFormState: NewProfileFormModel(), // Initialize profile form
		confirmDialog:    NewConfirmModel("Confirm Action", "Are you sure?"),
	}
}

// createSubmenu creates a submenu list for the given menu type
func (m *Model) createSubmenu(menuType MenuType) {
	var items []list.Item
	var title string

	switch menuType {
	case MenuTypeClaude:
		items = []list.Item{
			listItem{title: "Claude Desktop Config", description: "Configure Claude Desktop MCP servers"},
		}
		title = "Claude Desktop"
		m.breadcrumb = "Claude Desktop"

	case MenuTypeCursor:
		items = []list.Item{
			listItem{title: "Global Cursor Config", description: "Configure global Cursor MCP servers"},
		}
		title = "Cursor"
		m.breadcrumb = "Cursor"

	case MenuTypeAmpCode:
		items = []list.Item{
			listItem{title: "Cursor settings.json", description: "Configure Amp MCP servers in Cursor settings.json"},
			listItem{title: "~/.config/amp/settings.json", description: "Configure standalone Amp MCP servers"},
		}
		title = "Amp Code"
		m.breadcrumb = "Amp Code"

	case MenuTypeCrush:
		items = []list.Item{
			listItem{title: ".crush.json (local)", description: "Configure Crush MCP servers in .crush.json"},
			listItem{title: "crush.json (cwd)", description: "Configure Crush MCP servers in crush.json"},
			listItem{title: "~/.config/crush/crush.json (global)", description: "Configure Crush MCP servers in global config"},
		}
		title = "Crush"
		m.breadcrumb = "Crush"

	case MenuTypeProfiles:
		// Profiles don't need a submenu, go directly to list
		m.configType = ConfigTypeProfile
		m.breadcrumb = "Profiles"
		return
	}

	// Create the submenu list
	submenuDelegate := list.NewDefaultDelegate()
	m.submenuList = list.New(items, submenuDelegate, m.width, m.height-3)
	m.submenuList.Title = title
	m.submenuList.SetShowHelp(false)
	m.currentMenuType = menuType
	m.mode = modeSubmenu
}

// Init initializes the model and returns an initial command
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is called when a message is received
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		// Update the menu list dimensions
		headerHeight := 1 // For the title
		footerHeight := 2 // For help view
		verticalMarginHeight := headerHeight + footerHeight

		m.menuList.SetSize(msg.Width, msg.Height-verticalMarginHeight)

		// Update submenu dimensions if it exists
		if m.mode == modeSubmenu {
			m.submenuList.SetSize(msg.Width, msg.Height-verticalMarginHeight)
		}

		// If we have an active list, update its dimensions too
		if m.activeList != nil {
			m.activeList.SetSize(msg.Width, msg.Height-verticalMarginHeight)
		}

		return m, nil

	case tea.KeyMsg:
		// Global key handlers - only apply when not in form editing modes
		switch {
		case key.Matches(msg, m.keys.Quit) && m.mode != modeAddEdit:
			// Only allow quit when not editing in a form
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		// Mode-specific key handlers
		switch m.mode {
		case modeMenu:
			switch {
			case key.Matches(msg, m.keys.Enter):
				selectedItem := m.menuList.SelectedItem().(listItem)
				switch selectedItem.title {
				case "Claude Desktop":
					m.createSubmenu(MenuTypeClaude)
					return m, nil
				case "Cursor":
					m.createSubmenu(MenuTypeCursor)
					return m, nil
				case "Amp Code":
					m.createSubmenu(MenuTypeAmpCode)
					return m, nil
				case "Crush":
					m.createSubmenu(MenuTypeCrush)
					return m, nil
				case "Profiles":
					m.createSubmenu(MenuTypeProfiles)
					return m, m.loadProfiles()
				}
			}

			// Pass the message to the list
			m.menuList, cmd = m.menuList.Update(msg)
			return m, cmd

		case modeSubmenu:
			switch {
			case key.Matches(msg, m.keys.Back):
				m.mode = modeMenu
				m.breadcrumb = ""
				return m, nil
			case key.Matches(msg, m.keys.Enter):
				selectedItem := m.submenuList.SelectedItem().(listItem)

				// Determine which config to load based on submenu type and selection
				switch m.currentMenuType {
				case MenuTypeProfiles:
					// Profiles don't have submenu items, should not reach here
					return m, nil
				case MenuTypeClaude:
					switch selectedItem.title {
					case "Claude Desktop Config":
						m.configType = ConfigTypeClaude
						m.breadcrumb = "Claude Desktop > Config"
						return m, m.loadServers(ConfigTypeClaude)
					}
				case MenuTypeCursor:
					switch selectedItem.title {
					case "Global Cursor Config":
						m.configType = ConfigTypeCursor
						m.breadcrumb = "Cursor > Global Config"
						return m, m.loadServers(ConfigTypeCursor)
					}
				case MenuTypeAmpCode:
					switch selectedItem.title {
					case "Cursor settings.json":
						m.configType = ConfigTypeAmpCode
						m.breadcrumb = "Amp Code > Cursor settings.json"
						return m, m.loadServers(ConfigTypeAmpCode)
					case "~/.config/amp/settings.json":
						m.configType = ConfigTypeAmp
						m.breadcrumb = "Amp Code > ~/.config/amp/settings.json"
						return m, m.loadServers(ConfigTypeAmp)
					}
				case MenuTypeCrush:
					switch selectedItem.title {
					case ".crush.json (local)":
						m.configType = ConfigTypeCrushLocal
						m.breadcrumb = "Crush > .crush.json (local)"
						return m, m.loadServers(ConfigTypeCrushLocal)
					case "crush.json (cwd)":
						m.configType = ConfigTypeCrushCwd
						m.breadcrumb = "Crush > crush.json (cwd)"
						return m, m.loadServers(ConfigTypeCrushCwd)
					case "~/.config/crush/crush.json (global)":
						m.configType = ConfigTypeCrushGlobal
						m.breadcrumb = "Crush > ~/.config/crush/crush.json (global)"
						return m, m.loadServers(ConfigTypeCrushGlobal)
					}
				}
			}

			// Pass the message to the submenu list
			m.submenuList, cmd = m.submenuList.Update(msg)
			return m, cmd

		case modeList:
			// Handle specific actions in list view
			switch {
			case key.Matches(msg, m.keys.Back):
				// Go back to submenu if we came from one, otherwise main menu
				if m.currentMenuType != "" {
					m.createSubmenu(m.currentMenuType)
					return m, nil
				} else {
					m.configType = ConfigTypeNone // Clear the config type
					m.mode = modeMenu             // Go back to menu
					m.breadcrumb = ""
					return m, nil
				}

			case key.Matches(msg, m.keys.Add):
				if m.configType == ConfigTypeProfile {
					// Switch to add profile form
					m.profileFormState = NewProfileFormModel()
					m.mode = modeAddEdit
					return m, nil
				} else {
					// For server configs
					m.formState = NewFormModel()
					m.mode = modeAddEdit
					return m, m.formState.updateFocus()
				}

			case key.Matches(msg, m.keys.Edit):
				if m.activeList != nil && m.activeList.SelectedItem() != nil {
					if m.configType == ConfigTypeProfile {
						// Edit profile
						selectedItem := m.activeList.SelectedItem().(listItem)
						m.profileFormState = NewProfileFormModel()
						m.profileFormState.SetProfileData(selectedItem.title, selectedItem.description)
						m.mode = modeAddEdit
						return m, nil
					} else {
						// Edit server
						selectedItem := m.activeList.SelectedItem().(serverItem)
						var err error
						m.formState, err = m.loadServerToForm(selectedItem.name)
						if err != nil {
							m.errorMsg = fmt.Sprintf("Error loading server: %s", err)
							return m, nil
						}
						m.mode = modeAddEdit
						return m, m.formState.updateFocus()
					}
				}

			case key.Matches(msg, m.keys.Duplicate):
				if m.activeList != nil && m.activeList.SelectedItem() != nil {
					if m.configType == ConfigTypeProfile {
						// Duplicate profile
						selectedItem := m.activeList.SelectedItem().(listItem)
						m.profileFormState = NewProfileFormModel()
						m.profileFormState.SetProfileData(selectedItem.title+"-copy", selectedItem.description)
						m.profileFormState.isAddMode = true
						m.mode = modeAddEdit
						return m, nil
					} else {
						// Duplicate server
						selectedItem := m.activeList.SelectedItem().(serverItem)
						var err error
						m.formState, err = m.loadServerToForm(selectedItem.name)
						if err != nil {
							m.errorMsg = fmt.Sprintf("Error loading server: %s", err)
							return m, nil
						}
						// Set a new name for the duplicate with "-copy" suffix
						m.formState.nameInput.SetValue(selectedItem.name + "-copy")
						m.formState.isAddMode = true
						m.mode = modeAddEdit
						return m, m.formState.updateFocus()
					}
				}

			case key.Matches(msg, m.keys.Delete):
				if m.activeList != nil && m.activeList.SelectedItem() != nil {
					if m.configType == ConfigTypeProfile {
						// Delete profile confirmation
						selectedItem := m.activeList.SelectedItem().(listItem)
						m.confirmDialog = NewConfirmModel(fmt.Sprintf("Delete profile '%s'?", selectedItem.title), "Are you sure you want to delete this profile? This cannot be undone.")
						m.confirmAction = "delete-profile"
						m.actionServerName = selectedItem.title
						m.mode = modeConfirm
						return m, nil
					} else {
						// Delete server confirmation
						selectedItem := m.activeList.SelectedItem().(serverItem)
						m.confirmDialog = NewConfirmModel(fmt.Sprintf("Delete server '%s'?", selectedItem.name), "Are you sure you want to delete this server? This cannot be undone.")
						m.confirmAction = "delete-server"
						m.actionServerName = selectedItem.name
						m.mode = modeConfirm
						return m, nil
					}
				}

			case key.Matches(msg, m.keys.Enable):
				if m.activeList != nil && m.activeList.SelectedItem() != nil && m.configType != ConfigTypeProfile {
					// Only applies to servers, not profiles
					selectedItem := m.activeList.SelectedItem().(serverItem)
					return m, m.toggleServerEnabled(selectedItem.name)
				} else if m.activeList != nil && m.activeList.SelectedItem() != nil && m.configType == ConfigTypeProfile {
					// For profiles, this sets the default profile
					selectedItem := m.activeList.SelectedItem().(listItem)
					return m, m.setDefaultProfile(selectedItem.title)
				}
			}

			// Pass the message to the list
			if m.activeList != nil {
				*m.activeList, cmd = m.activeList.Update(msg)
				return m, cmd
			}

		case modeAddEdit:
			// In add/edit mode, determine if we're editing a profile or server
			if m.configType == ConfigTypeProfile {
				// Handle profile form
				m.profileFormState, cmd = m.profileFormState.Update(msg)

				if m.profileFormState.submitted {
					// Process form submission
					name, description, err := m.profileFormState.GetProfileData()
					if err != nil {
						m.errorMsg = fmt.Sprintf("Error: %s", err)
						m.profileFormState.submitted = false
						return m, nil
					}

					// Get tool and prompt paths
					toolDirs, toolFiles, promptDirs, promptFiles := m.profileFormState.GetToolsAndPrompts()

					isNewProfile := m.profileFormState.isAddMode
					m.mode = modeList // Return to list view
					return m, m.saveProfile(name, description, toolDirs, toolFiles, promptDirs, promptFiles, isNewProfile)
				}

				if m.profileFormState.cancelled {
					m.mode = modeList // Return to list view
					return m, nil
				}
			} else {
				// Handle server form
				m.formState, cmd = m.formState.Update(msg)

				if m.formState.submitted {
					// Form submitted, process data
					server, err := m.formState.ToServer()
					if err != nil {
						m.errorMsg = fmt.Sprintf("Error: %s", err)
						m.formState.submitted = false
						return m, nil
					}

					m.mode = modeList // Return to list view
					return m, m.saveServer(server, !m.formState.isAddMode)
				}

				if m.formState.cancelled {
					m.mode = modeList // Return to list view
					return m, nil
				}
			}

			return m, cmd

		case modeConfirm:
			// Handle confirmation dialog
			m.confirmDialog, cmd = m.confirmDialog.Update(msg)

			if m.confirmDialog.Confirmed() {
				switch m.confirmAction {
				case "delete-server":
					m.mode = modeList // Return to list view
					return m, m.deleteServer(m.actionServerName)
				case "delete-profile":
					m.mode = modeList // Return to list view
					return m, m.deleteProfile(m.actionServerName)
				}
			}

			if m.confirmDialog.Cancelled() {
				m.mode = modeList // Return to list view
				return m, nil
			}

			return m, cmd
		}

	// Handle profile-related messages
	case loadedProfilesMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error loading profiles: %s", msg.err)
			m.mode = modeMenu
			return m, nil
		}

		m.profileEditor = msg.editor

		// Convert profiles to list items
		items := make([]list.Item, 0, len(msg.profiles))
		for name, desc := range msg.profiles {
			title := name
			if name == msg.defaultProfile {
				title = "✓ " + name + " (Default)"
			}
			items = append(items, listItem{
				title:       title,
				description: desc,
			})
		}

		// Sort items by name
		sort.Slice(items, func(i, j int) bool {
			return items[i].(listItem).title < items[j].(listItem).title
		})

		// Create and configure the profile list
		delegate := list.NewDefaultDelegate()
		profileList := list.New(items, delegate, m.width, m.height-3)
		profileList.Title = "Profiles"
		profileList.SetShowHelp(false)

		m.activeList = &profileList
		m.mode = modeList

		return m, nil

	case profileSavedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving profile: %s", msg.err)
		} else {
			m.errorMsg = fmt.Sprintf("Profile '%s' saved", msg.profileName)
		}
		return m, m.loadProfiles()

	case profileDeletedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error deleting profile: %s", msg.err)
		} else {
			m.errorMsg = fmt.Sprintf("Profile '%s' deleted", msg.profileName)
		}
		return m, m.loadProfiles()

	case defaultProfileSetMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error setting default profile: %s", msg.err)
		} else {
			m.errorMsg = fmt.Sprintf("Default profile set to '%s'", msg.profileName)
		}
		return m, m.loadProfiles()

	// Handle server-related messages
	case loadedServersMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error loading servers: %s", msg.err)
			m.mode = modeMenu
			return m, nil
		}

		m.configType = msg.configType
		m.currentEditor = msg.editor

		// Convert servers to list items
		items := make([]list.Item, 0, len(msg.servers))
		for name, server := range msg.servers {
			// Determine if server is enabled based on disabled status (inverse)
			isEnabled := true

			// Check if server is disabled via editor
			disabled, err := m.currentEditor.IsServerDisabled(name)
			if err == nil && disabled {
				isEnabled = false
			}

			items = append(items, serverItem{
				name:    name,
				command: server.Command,
				args:    server.Args,
				env:     server.Env,
				url:     server.URL,
				enabled: isEnabled,
				isSSE:   server.IsSSE,
			})
		}

		// Sort items alphabetically by name
		sort.Slice(items, func(i, j int) bool {
			return items[i].(serverItem).name < items[j].(serverItem).name
		})

		// Create and configure the list
		delegate := list.NewDefaultDelegate()
		serverList := list.New(items, delegate, m.width, m.height-3)

		switch m.configType {
		case ConfigTypeCursor:
			serverList.Title = "Cursor MCP Servers"
		case ConfigTypeClaude:
			serverList.Title = "Claude MCP Servers"
		case ConfigTypeAmpCode:
			serverList.Title = "Amp MCP Servers (Cursor)"
		case ConfigTypeAmp:
			serverList.Title = "Amp MCP Servers"
		case ConfigTypeProfile:
			serverList.Title = "Profiles"
		case ConfigTypeCrushLocal:
			serverList.Title = "Crush MCP Servers (.crush.json)"
		case ConfigTypeCrushCwd:
			serverList.Title = "Crush MCP Servers (crush.json)"
		case ConfigTypeCrushGlobal:
			serverList.Title = "Crush MCP Servers (global)"
		case ConfigTypeNone:
			serverList.Title = "Servers"
		}

		serverList.SetShowHelp(false)

		m.activeList = &serverList
		m.mode = modeList

		return m, nil

	case serverDeletedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error deleting server: %s", msg.err)
		} else {
			m.errorMsg = fmt.Sprintf("Server '%s' deleted", msg.serverName)
		}
		return m, m.loadServers(m.configType)

	case serverToggleEnabledMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error toggling server: %s", msg.err)
		} else {
			status := "enabled"
			if !msg.enabled {
				status = "disabled"
			}
			m.errorMsg = fmt.Sprintf("Server '%s' %s", msg.serverName, status)
		}
		return m, m.loadServers(m.configType)

	case serverSavedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving server: %s", msg.err)
		} else {
			m.errorMsg = fmt.Sprintf("Server '%s' saved", msg.serverName)
		}
		return m, m.loadServers(m.configType)

	case errorMsg:
		m.errorMsg = msg.Error()
		return m, nil
	}

	return m, nil
}

// View renders the current state of the model
func (m Model) View() string {
	var sb strings.Builder

	// Display error message if present
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		sb.WriteString(errorStyle.Render(m.errorMsg) + "\n\n")
	}

	// Display breadcrumb if present
	if m.breadcrumb != "" {
		breadcrumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
		sb.WriteString(breadcrumbStyle.Render(m.breadcrumb) + "\n")
	}

	// Display content based on the current mode
	switch m.mode {
	case modeMenu:
		sb.WriteString(m.menuList.View())
	case modeSubmenu:
		sb.WriteString(m.submenuList.View())
	case modeList:
		if m.activeList != nil {
			sb.WriteString(m.activeList.View())
		}
	case modeAddEdit:
		if m.configType == ConfigTypeProfile {
			// Display profile form
			sb.WriteString(m.profileFormState.View())
		} else {
			// Display server form
			sb.WriteString(m.formState.View())
		}
	case modeConfirm:
		sb.WriteString(m.confirmDialog.View())
	}

	// Display help
	helpView := m.help.View(m)
	sb.WriteString("\n" + helpView)

	return sb.String()
}

// contextualHelp returns the appropriate help bindings for the current mode
func (m Model) contextualHelp() []key.Binding {
	switch m.mode {
	case modeMenu:
		return []key.Binding{
			m.keys.Enter,
			m.keys.Quit,
			m.keys.Help,
		}
	case modeSubmenu:
		return []key.Binding{
			m.keys.Enter,
			m.keys.Back,
			m.keys.Quit,
			m.keys.Help,
		}
	case modeList:
		// Help for list mode depends on the config type
		if m.configType == ConfigTypeProfile {
			return []key.Binding{
				m.keys.Back,
				m.keys.Enter,
				m.keys.Add,
				m.keys.Edit,
				m.keys.Delete,
				m.keys.Duplicate, // Allow duplicating profiles
				m.keys.Enable,    // This is used for setting default profile
				m.keys.Help,
				m.keys.Quit,
			}
		} else {
			return []key.Binding{
				m.keys.Back,
				m.keys.Enter,
				m.keys.Add,
				m.keys.Edit,
				m.keys.Delete,
				m.keys.Enable,
				m.keys.Duplicate,
				m.keys.Help,
				m.keys.Quit,
			}
		}
	case modeAddEdit:
		if m.configType == ConfigTypeProfile {
			// Return profile form help
			return []key.Binding{
				m.profileFormState.keyMap.Submit,
				m.profileFormState.keyMap.Cancel,
				m.profileFormState.keyMap.Next,
				m.profileFormState.keyMap.Prev,
			}
		} else {
			// Return server form help
			return []key.Binding{
				m.formState.keyMap.Submit,
				m.formState.keyMap.Cancel,
				m.formState.keyMap.Next,
				m.formState.keyMap.Prev,
			}
		}
	case modeConfirm:
		return []key.Binding{
			m.confirmDialog.keyMap.Confirm,
			m.confirmDialog.keyMap.Cancel,
		}
	default:
		return []key.Binding{m.keys.Quit}
	}
}

// ShortHelp implements help.KeyMap for compatibility
func (m Model) ShortHelp() []key.Binding {
	return m.contextualHelp()
}

// FullHelp implements help.KeyMap for compatibility
func (m Model) FullHelp() [][]key.Binding {
	// Return help based on current mode
	switch m.mode {
	case modeMenu:
		return [][]key.Binding{
			{m.keys.Up, m.keys.Down, m.keys.Enter},
			{m.keys.Help, m.keys.Quit},
		}
	case modeSubmenu:
		return [][]key.Binding{
			{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back},
			{m.keys.Help, m.keys.Quit},
		}
	case modeList:
		if m.configType == ConfigTypeProfile {
			return [][]key.Binding{
				{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back},
				{m.keys.Add, m.keys.Edit, m.keys.Delete, m.keys.Duplicate},
				{m.keys.Enable, m.keys.Help, m.keys.Quit},
			}
		}
		return [][]key.Binding{
			{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back},
			{m.keys.Add, m.keys.Edit, m.keys.Delete, m.keys.Duplicate},
			{m.keys.Enable, m.keys.Help, m.keys.Quit},
		}
	case modeAddEdit:
		if m.configType == ConfigTypeProfile {
			return [][]key.Binding{
				{m.profileFormState.keyMap.Submit, m.profileFormState.keyMap.Cancel},
				{m.profileFormState.keyMap.Next, m.profileFormState.keyMap.Prev},
			}
		}
		return [][]key.Binding{
			{m.formState.keyMap.Submit, m.formState.keyMap.Cancel},
			{m.formState.keyMap.Next, m.formState.keyMap.Prev},
		}
	case modeConfirm:
		return [][]key.Binding{
			{m.confirmDialog.keyMap.Confirm, m.confirmDialog.keyMap.Cancel},
		}
	default:
		return [][]key.Binding{{m.keys.Quit}}
	}
}

// loadServerToForm loads the data for the selected server into the form model.
func (m *Model) loadServerToForm(serverName string) (FormModel, error) {
	if m.currentEditor == nil {
		return FormModel{}, fmt.Errorf("no configuration editor loaded")
	}

	server, found, err := m.currentEditor.GetServer(serverName)
	if err != nil {
		return FormModel{}, fmt.Errorf("error getting server '%s': %w", serverName, err)
	}
	if !found {
		return FormModel{}, fmt.Errorf("server '%s' not found", serverName)
	}

	// Create and populate a new form model
	form := NewFormModel()
	form.LoadFromServer(server)
	form.isAddMode = false // Explicitly set to edit mode

	return form, nil
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
