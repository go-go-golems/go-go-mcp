package tui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-mcp/pkg/core/configstore"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// convertConfigType converts TUI ConfigType to configstore ConfigType
func convertConfigType(tType ConfigType) configstore.ConfigType {
	switch tType {
	case ConfigTypeCursor:
		return configstore.ConfigTypeCursor
	case ConfigTypeClaude:
		return configstore.ConfigTypeClaude
	case ConfigTypeAmpCode:
		return configstore.ConfigTypeAmpCode
	case ConfigTypeAmp:
		return configstore.ConfigTypeAmp
	case ConfigTypeProfile:
		return configstore.ConfigTypeProfile
	case ConfigTypeCrushLocal:
		return configstore.ConfigTypeCrushLocal
	case ConfigTypeCrushCwd:
		return configstore.ConfigTypeCrushCwd
	case ConfigTypeCrushGlobal:
		return configstore.ConfigTypeCrushGlobal
	case ConfigTypeNone:
		return configstore.ConfigTypeNone
	default:
		return configstore.ConfigTypeNone
	}
}

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

	// Domain stores for configuration management
	currentStore configstore.Store
	profileStore configstore.ProfileStore

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

	// Yank/paste clipboard
	yankedServer *types.CommonServer
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
					return m, LoadProfilesCmd()
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
						return m, LoadServersCmd(configstore.ConfigTypeClaude)
					}
				case MenuTypeCursor:
					switch selectedItem.title {
					case "Global Cursor Config":
						m.configType = ConfigTypeCursor
						m.breadcrumb = "Cursor > Global Config"
						return m, LoadServersCmd(configstore.ConfigTypeCursor)
					}
				case MenuTypeAmpCode:
					switch selectedItem.title {
					case "Cursor settings.json":
						m.configType = ConfigTypeAmpCode
						m.breadcrumb = "Amp Code > Cursor settings.json"
						return m, LoadServersCmd(configstore.ConfigTypeAmpCode)
					case "~/.config/amp/settings.json":
						m.configType = ConfigTypeAmp
						m.breadcrumb = "Amp Code > ~/.config/amp/settings.json"
						return m, LoadServersCmd(configstore.ConfigTypeAmp)
					}
				case MenuTypeCrush:
					switch selectedItem.title {
					case ".crush.json (local)":
						m.configType = ConfigTypeCrushLocal
						m.breadcrumb = "Crush > .crush.json (local)"
						return m, LoadServersCmd(configstore.ConfigTypeCrushLocal)
					case "crush.json (cwd)":
						m.configType = ConfigTypeCrushCwd
						m.breadcrumb = "Crush > crush.json (cwd)"
						return m, LoadServersCmd(configstore.ConfigTypeCrushCwd)
					case "~/.config/crush/crush.json (global)":
						m.configType = ConfigTypeCrushGlobal
						m.breadcrumb = "Crush > ~/.config/crush/crush.json (global)"
						return m, LoadServersCmd(configstore.ConfigTypeCrushGlobal)
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
					m.mode = modeMenu
					m.breadcrumb = ""
					return m, nil
				}

			case key.Matches(msg, m.keys.Add):
				// Switch to add mode
				m.mode = modeAddEdit
				if m.configType == ConfigTypeProfile {
					// Reset profile form for adding
					m.profileFormState = NewProfileFormModel()
					m.profileFormState.isAddMode = true
				} else {
					// Reset form for adding
					cmd = m.formState.Reset()
					m.formState.isAddMode = true
				}
				return m, cmd

			case key.Matches(msg, m.keys.Edit):
				// Switch to edit mode for the selected item
				if m.activeList != nil && m.activeList.SelectedItem() != nil {
					if m.configType == ConfigTypeProfile {
						// Edit profile
						selectedItem := m.activeList.SelectedItem().(listItem)
						profileName := selectedItem.title
						profileDescription := selectedItem.description

						m.profileFormState = NewProfileFormModel()
						m.profileFormState.SetProfileData(profileName, profileDescription)
						m.mode = modeAddEdit
					} else {
						// Edit server
						selectedItem := m.activeList.SelectedItem().(serverItem)
						serverName := selectedItem.name

						form, err := m.loadServerToForm(serverName)
						if err != nil {
							m.errorMsg = fmt.Sprintf("Error loading server for editing: %v", err)
							return m, nil
						}

						m.formState = form
						m.mode = modeAddEdit
					}
				}
				return m, nil

			case key.Matches(msg, m.keys.Delete):
				// Trigger confirmation dialog for deletion
				if m.activeList != nil && m.activeList.SelectedItem() != nil {
					if m.configType == ConfigTypeProfile {
						selectedItem := m.activeList.SelectedItem().(listItem)
						profileName := selectedItem.title
						m.confirmDialog = NewConfirmModel("Delete Profile", fmt.Sprintf("Are you sure you want to delete profile '%s'?", profileName))
						m.confirmAction = "delete_profile"
						m.actionServerName = profileName
					} else {
						selectedItem := m.activeList.SelectedItem().(serverItem)
						serverName := selectedItem.name
						m.confirmDialog = NewConfirmModel("Delete Server", fmt.Sprintf("Are you sure you want to delete server '%s'?", serverName))
						m.confirmAction = "delete"
						m.actionServerName = serverName
					}
					m.mode = modeConfirm
				}
				return m, nil

			case key.Matches(msg, m.keys.Duplicate):
				// Duplicate the selected server (only for servers, not profiles)
				if m.configType != ConfigTypeProfile && m.activeList != nil && m.activeList.SelectedItem() != nil {
					selectedItem := m.activeList.SelectedItem().(serverItem)
					serverName := selectedItem.name

					// Load the server into the form and change to add mode
					form, err := m.loadServerToForm(serverName)
					if err != nil {
						m.errorMsg = fmt.Sprintf("Error loading server for duplication: %v", err)
						return m, nil
					}

					// Clear the name to indicate this is a duplicate
					form.nameInput.SetValue(serverName + "_copy")
					form.isAddMode = true

					m.formState = form
					m.mode = modeAddEdit
				}
				return m, nil

			case key.Matches(msg, m.keys.Yank):
				// Yank (copy) the selected server for later pasting
				if m.configType != ConfigTypeProfile && m.activeList != nil && m.activeList.SelectedItem() != nil {
					selectedItem := m.activeList.SelectedItem().(serverItem)
					serverName := selectedItem.name

					if m.currentStore != nil {
						server, found, err := m.currentStore.GetServer(serverName)
						if err != nil {
							m.errorMsg = fmt.Sprintf("Error yanking server: %v", err)
						} else if !found {
							m.errorMsg = fmt.Sprintf("Server '%s' not found", serverName)
						} else {
							m.yankedServer = &server
							// You could display a brief success message here
						}
					}
				}
				return m, nil

			case key.Matches(msg, m.keys.Paste):
				// Paste the yanked server (only for servers, not profiles)
				if m.configType != ConfigTypeProfile && m.yankedServer != nil {
					// Load the yanked server into the form
					form, err := m.loadCommonServerToForm(*m.yankedServer)
					if err != nil {
						m.errorMsg = fmt.Sprintf("Error loading yanked server: %v", err)
						return m, nil
					}

					// Clear the name to indicate this is a paste operation
					form.nameInput.SetValue(m.yankedServer.Name + "_pasted")
					form.isAddMode = true

					m.formState = form
					m.mode = modeAddEdit
				}
				return m, nil

			case key.Matches(msg, m.keys.Enable):
				// Toggle enabled/disabled for the selected server (only for servers, not profiles)
				if m.configType != ConfigTypeProfile && m.activeList != nil && m.activeList.SelectedItem() != nil {
					selectedItem := m.activeList.SelectedItem().(serverItem)
					serverName := selectedItem.name
					return m, ToggleServerEnabledCmd(m.currentStore, serverName)
				} else if m.configType == ConfigTypeProfile && m.activeList != nil && m.activeList.SelectedItem() != nil {
					// For profiles, set as default
					selectedItem := m.activeList.SelectedItem().(listItem)
					profileName := selectedItem.title
					// Need to create/get profile store for setting default
					store, err := configstore.NewProfileStore()
					if err != nil {
						m.errorMsg = fmt.Sprintf("Error creating profile store: %v", err)
						return m, nil
					}
					return m, SetDefaultProfileCmd(store, profileName)
				}
				return m, nil
			}

			// Pass the message to the active list if available
			if m.activeList != nil {
				*m.activeList, cmd = m.activeList.Update(msg)
			}
			return m, cmd

		case modeAddEdit:
			// Handle form input
			if m.configType == ConfigTypeProfile {
				// Handle profile form
				m.profileFormState, cmd = m.profileFormState.Update(msg)

				if m.profileFormState.submitted {
					// Extract form data
					name, description, err := m.profileFormState.GetProfileData()
					if err != nil {
						m.errorMsg = fmt.Sprintf("Form validation error: %v", err)
						m.profileFormState.submitted = false // Reset the flag
						return m, nil
					}

					// Get tools and prompts data
					toolDirs, toolFiles, promptDirs, promptFiles := m.profileFormState.GetToolsAndPrompts()

					// Save the profile
					// Need to create/get profile store for saving
					store, err := configstore.NewProfileStore()
					if err != nil {
						m.errorMsg = fmt.Sprintf("Error creating profile store: %v", err)
						return m, nil
					}
					return m, SaveProfileCmd(store, name, description, toolDirs, toolFiles, promptDirs, promptFiles, m.profileFormState.isAddMode)

				} else if m.profileFormState.cancelled {
					// Go back to list view
					m.mode = modeList
					return m, nil
				}
			} else {
				// Handle server form
				m.formState, cmd = m.formState.Update(msg)

				if m.formState.submitted {
					// Extract form data and save
					server, err := m.formState.ToServer()
					if err != nil {
						m.errorMsg = fmt.Sprintf("Form validation error: %v", err)
						m.formState.submitted = false // Reset the flag
						return m, nil
					}

					// Save the server (overwrite if editing)
					overwrite := !m.formState.isAddMode
					return m, SaveServerCmd(m.currentStore, server, overwrite)

				} else if m.formState.cancelled {
					// Go back to list view
					m.mode = modeList
					return m, nil
				}
			}
			return m, cmd

		case modeConfirm:
			// Handle confirmation dialog
			m.confirmDialog, cmd = m.confirmDialog.Update(msg)

			if m.confirmDialog.Confirmed() {
				// Execute the confirmed action
				switch m.confirmAction {
				case "delete":
					m.mode = modeList
					return m, DeleteServerCmd(m.currentStore, m.actionServerName)
				case "delete_profile":
					m.mode = modeList
					// Need to create/get profile store for deleting
					store, err := configstore.NewProfileStore()
					if err != nil {
						m.errorMsg = fmt.Sprintf("Error creating profile store: %v", err)
						return m, nil
					}
					return m, DeleteProfileCmd(store, m.actionServerName)
				}
				m.mode = modeList
			} else if m.confirmDialog.Cancelled() {
				// Cancel the action and go back to list
				m.mode = modeList
			}
			return m, cmd
		}

	case loadedServersMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error loading servers: %v", msg.err)
			return m, nil
		}

		// Create and store the appropriate store
		store, err := configstore.NewStore(convertConfigType(msg.configType))
		if err != nil {
			m.errorMsg = fmt.Sprintf("Error creating store: %v", err)
			return m, nil
		}
		m.currentStore = store

		// Convert servers to list items
		items := make([]list.Item, 0, len(msg.servers))
		for name, server := range msg.servers {
			// Check if server is disabled
			isDisabled, err := store.IsServerDisabled(name)
			if err != nil {
				// If we can't determine status, assume enabled
				isDisabled = false
			}

			items = append(items, serverItem{
				name:    name,
				command: server.Command,
				args:    server.Args,
				env:     server.Env,
				url:     server.URL,
				enabled: !isDisabled, // Set actual enabled status
				isSSE:   server.IsSSE,
			})
		}

		// Sort items by name for consistent display
		sort.Slice(items, func(i, j int) bool {
			return items[i].(serverItem).name < items[j].(serverItem).name
		})

		// Create the list with servers
		delegate := list.NewDefaultDelegate()
		serverList := list.New(items, delegate, m.width, m.height-3)
		serverList.Title = fmt.Sprintf("%s Servers", msg.configType)
		serverList.SetShowHelp(false)

		m.activeList = &serverList
		m.mode = modeList

		// Clear any previous error
		m.errorMsg = ""

		return m, nil

	case loadedProfilesMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error loading profiles: %v", msg.err)
			return m, nil
		}

		// Create and store the profile store
		store, err := configstore.NewProfileStore()
		if err != nil {
			m.errorMsg = fmt.Sprintf("Error creating profile store: %v", err)
			return m, nil
		}
		m.profileStore = store

		// Convert profiles to list items
		items := make([]list.Item, 0, len(msg.profiles))
		for name, description := range msg.profiles {
			// Mark default profile in description
			displayDesc := description
			if name == msg.defaultProfile {
				displayDesc = fmt.Sprintf("%s (default)", description)
			}
			items = append(items, listItem{
				title:       name,
				description: displayDesc,
			})
		}

		// Sort items by name for consistent display
		sort.Slice(items, func(i, j int) bool {
			return items[i].(listItem).title < items[j].(listItem).title
		})

		// Create the list with profiles
		delegate := list.NewDefaultDelegate()
		profileList := list.New(items, delegate, m.width, m.height-3)
		profileList.Title = "Profiles"
		profileList.SetShowHelp(false)

		m.activeList = &profileList
		m.mode = modeList

		// Clear any previous error
		m.errorMsg = ""

		return m, nil

	case serverDeletedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error deleting server: %v", msg.err)
		} else {
			// Successfully deleted, reload the server list
			return m, LoadServersCmd(convertConfigType(m.configType))
		}
		return m, nil

	case profileDeletedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error deleting profile: %v", msg.err)
		} else {
			// Successfully deleted, reload the profile list
			return m, LoadProfilesCmd()
		}
		return m, nil

	case serverToggleEnabledMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error toggling server: %v", msg.err)
		} else {
			// Successfully toggled, reload the server list
			return m, LoadServersCmd(convertConfigType(m.configType))
		}
		return m, nil

	case serverSavedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving server: %v", msg.err)
			return m, nil
		} else {
			// Successfully saved, go back to list and reload
			m.mode = modeList
			return m, LoadServersCmd(convertConfigType(m.configType))
		}

	case profileSavedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving profile: %v", msg.err)
			return m, nil
		} else {
			// Successfully saved, go back to list and reload
			m.mode = modeList
			return m, LoadProfilesCmd()
		}

	case defaultProfileSetMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error setting default profile: %v", msg.err)
		} else {
			// Successfully set default, reload the profile list
			return m, LoadProfilesCmd()
		}
		return m, nil

	case errorMsg:
		m.errorMsg = msg.Error()
		return m, nil
	}

	return m, nil
}

// View renders the current state of the model
