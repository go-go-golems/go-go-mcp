package tui

import (
	"fmt"
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
)

// Define the different UI modes
type mode int

const (
	modeMenu mode = iota
	modeList
	modeAddEdit
	modeConfirm
)

// Define configuration types
type ConfigType string

const (
	ConfigTypeCursor ConfigType = "cursor"
	ConfigTypeClaude ConfigType = "claude"
	ConfigTypeNone   ConfigType = "" // Represents no config loaded
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

// Main application model
type Model struct {
	keys       keyMap
	help       help.Model
	mode       mode
	menuList   list.Model
	activeList *list.Model // Pointer to the currently active server list
	width      int
	height     int

	// Configuration editor interface
	currentEditor types.ServerConfigEditor

	// Track which config type is being edited (used for display/loading)
	configType ConfigType // Use the enum type

	// Error message to display
	errorMsg string

	// Form state for add/edit
	formState FormModel

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

// NewModel initializes and returns a new Model
func NewModel() Model {
	keys := defaultKeyMap
	h := help.New()

	// Create menu items
	menuItems := []list.Item{
		listItem{title: "Manage Cursor Configuration", description: "Configure Cursor MCP servers"},
		listItem{title: "Manage Claude Desktop Configuration", description: "Configure Claude desktop MCP servers"},
		listItem{title: "Exit", description: "Exit the application"},
	}

	menuDelegate := list.NewDefaultDelegate()
	menu := list.New(menuItems, menuDelegate, 0, 0)
	menu.Title = "MCP Server Manager"
	menu.SetShowHelp(false)

	return Model{
		keys:     keys,
		help:     h,
		mode:     modeMenu,
		menuList: menu,
		// activeList will be set when a config type is chosen
		formState: NewFormModel(),
	}
}

// Init initializes the model and returns an initial command
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is called when a message is received
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.menuList.SetWidth(msg.Width)
		m.menuList.SetHeight(msg.Height - 4) // leave room for title and help

		// Update active list size if it exists
		if m.activeList != nil {
			m.activeList.SetWidth(msg.Width)
			m.activeList.SetHeight(msg.Height - 4)
		}

		// Also update the help model
		m.help.Width = msg.Width

	case loadedServersMsg: // Unified message for loading servers
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error loading %s servers: %s", msg.configType, msg.err.Error())
			return m, nil
		}

		// Update the editor and config type
		m.currentEditor = msg.editor
		m.configType = msg.configType // Assign the enum value

		// Create a new list model for the active list
		delegate := list.NewDefaultDelegate()
		newList := list.New([]list.Item{}, delegate, m.width, m.height-4)
		// Set title based on the enum
		switch msg.configType {
		case ConfigTypeCursor:
			newList.Title = "Cursor MCP Servers"
		case ConfigTypeClaude:
			newList.Title = "Claude Desktop MCP Servers"
		case ConfigTypeNone:
			newList.Title = "Unknown Server List"
		default:
			// Handle ConfigTypeNone or other unexpected values
			log.Warn().Str("configType", string(msg.configType)).Msg("Unexpected config type in loadedServersMsg")
			newList.Title = "Unknown Server List"
		}
		newList.SetShowHelp(false)
		m.activeList = &newList // Assign the pointer to the new list

		// Get disabled servers from the editor
		disabledServersMap := make(map[string]struct{})
		if m.currentEditor != nil {
			disabled, err := m.currentEditor.ListDisabledServers()
			if err != nil {
				log.Warn().Err(err).Msg("Could not list disabled servers")
				// Proceed without disabled info if there's an error
			} else {
				for _, name := range disabled {
					disabledServersMap[name] = struct{}{}
				}
			}
		}

		// Convert servers from map to sorted slice of list items
		serversList := make([]serverItem, 0, len(msg.servers))
		for name, server := range msg.servers {
			_, isDisabled := disabledServersMap[name]
			serversList = append(serversList, serverItem{
				name:    name,
				command: server.Command,
				args:    server.Args,
				env:     server.Env,
				url:     server.URL,
				isSSE:   server.IsSSE,
				enabled: !isDisabled,
			})
		}

		// Sort servers by name for consistent display
		sort.Slice(serversList, func(i, j int) bool {
			return serversList[i].name < serversList[j].name
		})

		// Convert serverItems to list.Item
		listItems := make([]list.Item, len(serversList))
		for i, item := range serversList {
			listItems[i] = item
		}

		// Set items in the active list
		cmd := m.activeList.SetItems(listItems)
		cmds = append(cmds, cmd)
		m.mode = modeList // Ensure we are in list mode after loading

	case serverDeletedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error deleting server '%s': %s", msg.serverName, msg.err.Error())
			return m, nil
		}
		// Reload server list for the current config type
		return m, m.loadServers(m.configType)

	case serverToggleEnabledMsg:
		log.Debug().
			Str("serverName", msg.serverName).
			Str("configType", string(m.configType)). // Convert enum to string for logging
			Bool("enabled", msg.enabled).
			Bool("success", msg.success).
			Err(msg.err).
			Msg("Received serverToggleEnabledMsg")

		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error toggling server state: %s", msg.err.Error())
			log.Error().Err(msg.err).Msg("Error toggling server state")
			return m, nil
		}

		// Update the item directly in the active list
		if m.activeList == nil {
			log.Error().Msg("activeList is nil during server toggle update")
			return m, m.loadServers(m.configType) // Fallback to reload
		}

		items := m.activeList.Items()
		for i, item := range items {
			if sItem, ok := item.(serverItem); ok && sItem.name == msg.serverName {
				sItem.enabled = msg.enabled
				cmd := m.activeList.SetItem(i, sItem)
				log.Debug().Str("serverName", msg.serverName).Bool("enabled", msg.enabled).Int("index", i).Msg("Updated server item in list")
				return m, cmd // Return the command from SetItem
			}
		}

		log.Warn().Str("serverName", msg.serverName).Msg("Server not found in list after toggle, performing full reload")
		// Fallback to reload if item not found
		return m, m.loadServers(m.configType)

	case serverSavedMsg: // Handle message after saving a server via form
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving server '%s': %s", msg.serverName, msg.err.Error())
			return m, nil
		}
		// Reload the list and switch back to list mode
		m.mode = modeList
		return m, m.loadServers(m.configType)

	case ConfirmMsg:
		// Return to list mode
		m.mode = modeList
		// Process the confirmation result
		if msg.Confirmed {
			switch m.confirmAction {
			case "delete":
				return m, m.deleteServer(m.actionServerName)
			case "save": // Confirmation for overwriting in form
				// Retrieve data from form and save
				commonServer, err := m.formState.ToServer()
				if err != nil {
					m.errorMsg = fmt.Sprintf("Error saving server: %s", err.Error())
					return m, nil
				}
				return m, m.saveServer(commonServer, true) // Pass overwrite=true
			}
		}
		return m, nil

	case errorMsg:
		m.errorMsg = msg.err.Error()

	case tea.KeyMsg:
		// Clear any previous error message on key press
		m.errorMsg = ""

		log.Debug().Str("key", msg.String()).Msg("Key pressed")

		// For text inputs in form mode, prioritize sending the key to the form model
		if m.mode == modeAddEdit {
			// Let the form handle all keys when in form mode
			formModel, cmd := m.formState.Update(msg)
			m.formState = formModel
			cmds = append(cmds, cmd)

			// Check if the form was submitted or cancelled
			if m.formState.submitted {
				log.Debug().Msg("Form submitted, saving server")
				// Form submitted, save the server
				commonServer, err := m.formState.ToServer()
				if err != nil {
					m.errorMsg = fmt.Sprintf("Invalid form data: %s", err.Error())
					// Reset submitted flag so user can fix and resubmit
					m.formState.submitted = false
					return m, nil // Stay in form mode with error
				}
				overwrite := !m.formState.isAddMode // Overwrite only if in edit mode
				cmds = append(cmds, m.saveServer(commonServer, overwrite))
				// Reset submitted flag after handling submission command
				m.formState.submitted = false
				m.mode = modeList
			} else if m.formState.cancelled {
				log.Debug().Msg("Form cancelled, going back to list")
				// Form cancelled, go back to list
				m.mode = modeList
				// Reset cancelled flag
				m.formState.cancelled = false
			}

			return m, tea.Batch(cmds...)
		}

		// Global keys (only processed if not in form mode or special cases)
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Back):
			// Handle going back based on current mode
			switch m.mode {
			case modeMenu:
				return m, nil
			case modeList:
				m.mode = modeMenu
				m.activeList = nil // Clear active list when going back to menu
				m.currentEditor = nil
				return m, nil
			case modeAddEdit:
				m.mode = modeList
				return m, nil
			case modeConfirm:
				m.mode = modeList
				return m, nil
			}

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil // Consume the key press
		}

		// Mode-specific keys
		switch m.mode {
		case modeMenu:
			switch {
			case key.Matches(msg, m.keys.Enter):
				var cmd tea.Cmd
				item := m.menuList.SelectedItem()
				log.Debug().Str("item", item.FilterValue()).Msg("Menu item selected")
				switch item.FilterValue() {
				case "Manage Cursor Configuration":
					log.Debug().Msg("Loading Cursor servers")
					cmd = m.loadServers(ConfigTypeCursor) // Use enum
				case "Manage Claude Desktop Configuration":
					log.Debug().Msg("Loading Claude servers")
					cmd = m.loadServers(ConfigTypeClaude) // Use enum
				case "Exit":
					return m, tea.Quit
				}
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			default:
				var cmd tea.Cmd
				m.menuList, cmd = m.menuList.Update(msg)
				cmds = append(cmds, cmd)
			}

		case modeAddEdit:
		case modeList:
			// Ensure activeList is not nil
			if m.activeList == nil {
				log.Error().Msg("activeList is nil in modeList key handling")
				m.mode = modeMenu // Go back to menu if list isn't loaded
				return m, nil
			}

			// Handle list view key commands
			switch {
			case key.Matches(msg, m.keys.Add):
				// Reset form and switch to add mode
				cmd := m.formState.Reset()
				m.formState.isAddMode = true
				m.mode = modeAddEdit
				// Set IsSSE to false by default for both types
				m.formState.isSSE = false
				return m, cmd

			case key.Matches(msg, m.keys.Edit), key.Matches(msg, m.keys.Enter):
				// Load selected server into form and switch to edit mode
				selectedItem := m.activeList.SelectedItem()
				if selectedItem == nil {
					m.errorMsg = "No server selected"
					return m, nil
				}

				if sItem, ok := selectedItem.(serverItem); ok {
					// Load server data into form
					form, err := m.loadServerToForm(sItem.name) // Pass name to helper
					if err != nil {
						m.errorMsg = err.Error()
						return m, nil
					}
					m.formState = form
					m.formState.isAddMode = false // Ensure it's edit mode
					m.mode = modeAddEdit
				} else {
					m.errorMsg = "Invalid item type selected"
				}
				return m, nil

			case key.Matches(msg, m.keys.Delete):
				// Show confirmation dialog for deletion
				selectedItem := m.activeList.SelectedItem()
				if selectedItem == nil {
					m.errorMsg = "No server selected"
					return m, nil
				}
				if sItem, ok := selectedItem.(serverItem); ok {
					m.confirmAction = "delete"
					m.actionServerName = sItem.name
					m.confirmDialog = NewConfirmModel(
						"Delete Server",
						fmt.Sprintf("Are you sure you want to delete server '%s'?", sItem.name),
					)
					m.mode = modeConfirm
				} else {
					m.errorMsg = "Invalid item type selected"
				}
				return m, nil

			case key.Matches(msg, m.keys.Duplicate):
				// Duplicate server and go to edit mode
				selectedItem := m.activeList.SelectedItem()
				if selectedItem == nil {
					m.errorMsg = "No server selected"
					return m, nil
				}
				if sItem, ok := selectedItem.(serverItem); ok {
					// Load server data into form
					form, err := m.loadServerToForm(sItem.name)
					if err != nil {
						m.errorMsg = err.Error()
						return m, nil
					}
					// Change name to indicate it's a duplicate
					form.nameInput.SetValue(sItem.name + "-copy")
					form.isAddMode = true // Set to add mode since we're creating a new server
					m.formState = form
					m.mode = modeAddEdit
				} else {
					m.errorMsg = "Invalid item type selected"
				}
				return m, nil

			case key.Matches(msg, m.keys.Enable):
				log.Debug().Msg("Enable key pressed in modeList")
				selectedItem := m.activeList.SelectedItem()
				if selectedItem == nil {
					m.errorMsg = "No server selected"
					return m, nil
				}
				if sItem, ok := selectedItem.(serverItem); ok {
					log.Debug().Str("serverName", sItem.name).Msg("Dispatching toggleServerEnabled command")
					return m, m.toggleServerEnabled(sItem.name)
				} else {
					m.errorMsg = "Invalid item type selected"
				}
				return m, nil

			case key.Matches(msg, m.keys.Enter): // Enter in list mode - maybe view details later?
				// Currently does nothing, could be used for edit or view details
				return m, nil

			default:
				// Update the active list
				var cmd tea.Cmd
				*m.activeList, cmd = m.activeList.Update(msg)
				cmds = append(cmds, cmd)
			}

		case modeConfirm:
			// Handle confirmation dialog
			var cmd tea.Cmd
			m.confirmDialog, cmd = m.confirmDialog.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current UI based on the model state
func (m Model) View() string {
	// If there's an error message, display it at the bottom
	errorView := ""
	if m.errorMsg != "" {
		errorView = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: "+m.errorMsg)
	}

	switch m.mode {
	case modeMenu:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.menuList.View(),
			m.help.View(m),
			errorView,
		)

	case modeList:
		if m.activeList == nil {
			// Show a loading or placeholder message if the list isn't ready
			loadingTitle := "Loading servers..."
			if m.configType != ConfigTypeNone {
				loadingTitle = fmt.Sprintf("Loading %s servers...", m.configType) // Enum implicitly converts to string
			}
			return lipgloss.JoinVertical(
				lipgloss.Left,
				loadingTitle,
				m.help.View(m),
				errorView,
			)
		}
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.activeList.View(), // Use the activeList pointer to view
			m.help.View(m),
			errorView,
		)

	case modeAddEdit:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.formState.View(),
			m.help.View(m),
			errorView,
		)

	case modeConfirm:
		// Render the confirmation dialog
		dialogView := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			Render(m.confirmDialog.View())
		// Center the dialog
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			dialogView,
		) + errorView

	default:
		return "Unknown mode" + errorView
	}
}

// Add ShortHelp method to Model
func (m Model) ShortHelp() []key.Binding {
	switch m.mode {
	case modeMenu:
		return []key.Binding{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Help, m.keys.Quit}
	case modeList:
		// Action focused short help for the list view
		return []key.Binding{m.keys.Add, m.keys.Edit, m.keys.Delete, m.keys.Duplicate, m.keys.Enable, m.keys.Back, m.keys.Help, m.keys.Quit}
	case modeAddEdit:
		// TODO: Add specific form keys (like Tab) later if needed
		return []key.Binding{m.formState.keyMap.Submit, m.formState.keyMap.Cancel, m.keys.Back, m.keys.Help, m.keys.Quit}
	case modeConfirm:
		// Use the confirmation dialog's specific keys
		return []key.Binding{m.confirmDialog.keyMap.Left, m.confirmDialog.keyMap.Right, m.confirmDialog.keyMap.Confirm, m.confirmDialog.keyMap.Cancel}
	default:
		return []key.Binding{m.keys.Help, m.keys.Quit}
	}
}

// Add FullHelp method to Model
func (m Model) FullHelp() [][]key.Binding {
	switch m.mode {
	case modeMenu:
		return [][]key.Binding{
			{m.keys.Up, m.keys.Down, m.keys.Enter}, // Navigation
			{m.keys.Help, m.keys.Quit},             // General
		}
	case modeList:
		return [][]key.Binding{
			{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back},                       // Navigation
			{m.keys.Add, m.keys.Edit, m.keys.Delete, m.keys.Duplicate, m.keys.Enable}, // Server Actions
			{m.keys.Help, m.keys.Quit},                                                // General
		}
	case modeAddEdit:
		// TODO: Add specific form keys later if needed
		return [][]key.Binding{
			{m.formState.keyMap.Submit, m.formState.keyMap.Cancel}, // Form Actions
			{m.keys.Back, m.keys.Help, m.keys.Quit},                // General Actions
		}
	case modeConfirm:
		return [][]key.Binding{
			{m.confirmDialog.keyMap.Left, m.confirmDialog.keyMap.Right},     // Selection
			{m.confirmDialog.keyMap.Confirm, m.confirmDialog.keyMap.Cancel}, // Actions
		}
	default:
		return [][]key.Binding{
			{m.keys.Help, m.keys.Quit},
		}
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
	form.nameInput.SetValue(server.Name)
	form.commandInput.SetValue(server.Command)
	form.argsInput.SetValue(strings.Join(server.Args, " "))

	// Format environment variables as KEY=VALUE pairs, one per line
	envText := ""
	keys := make([]string, 0, len(server.Env))
	for k := range server.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Sort keys for consistent display

	for _, k := range keys {
		if envText != "" {
			envText += "\n"
		}
		envText += fmt.Sprintf("%s=%s", k, server.Env[k])
	}

	form.envInput.SetValue(envText)
	form.urlInput.SetValue(server.URL)
	form.isSSE = server.IsSSE
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
		default: // Handles ConfigTypeNone and any other unexpected values
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
