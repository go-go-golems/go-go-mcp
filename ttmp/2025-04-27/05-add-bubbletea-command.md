# Product Requirements Document: Bubbletea UI for MCP Server Management

## Overview
This document outlines the requirements for implementing a terminal-based user interface using the charmbracelet/bubbletea library to manage MCP server configurations for both Cursor and Claude desktop applications. The UI will provide an intuitive interface for viewing, adding, removing, enabling, and disabling server configurations without requiring users to edit JSON configuration files directly.

## Background
The go-go-mcp project currently manages server configurations through command-line tools that modify JSON configuration files. While functional, this approach requires users to remember specific commands and syntax. A terminal UI improves usability by providing visual feedback and reducing cognitive load.

## Goals
- Create an intuitive TUI for managing MCP server configurations
- Support both Cursor and Claude desktop configurations
- Allow users to view, add, edit, remove, enable, and disable server configurations
- Maintain compatibility with existing configuration file formats
- Provide immediate visual feedback for user actions

## Non-Goals
- Replacing the existing CLI commands entirely (the TUI is an alternative interface)
- Modifying the configuration file format
- Providing WYSIWYG editing of raw JSON
- Supporting configuration options beyond server management

## User Experience

### Navigation and Layout
The application uses a full-screen interface managed by Bubble Tea:
1. A header/title area displaying the current context (e.g., "MCP Server Manager", "Cursor MCP Servers", "Add Server")
2. A main content area showing the menu, server list, form, or confirmation dialog.
3. A footer displaying context-aware keyboard shortcuts via the `help` bubble.

### Main Menu
The initial screen presents options for:
- Manage Cursor Configuration
- Manage Claude Desktop Configuration
- Exit
Navigation uses Up/Down arrows or j/k keys, selection uses Enter.

### Server List View
When a configuration type is selected, the screen displays:
- A list of configured servers showing:
    - Name
    - Type (CMD or SSE)
    - Command/URL
    - Enabled/Disabled status
- Navigation controls (Up/Down or j/k)
- Action shortcuts:
    - `a`: Add new server (opens form)
    - `e`: Edit selected server (opens form)
    - `d`: Delete selected server (shows confirmation)
    - `space`: Toggle enable/disable for selected server
    - `esc`: Back to main menu
    - `?`: Toggle full help
    - `q`: Quit

### Add/Edit Server View
A form interface with fields for:
- Server name (text input)
- Type (checkbox, toggles between SSE and stdio)
- URL (text input, shown only if Type is SSE)
- Command (text input, shown only if Type is stdio)
- Arguments (text input, shown only if Type is stdio)
- Environment variables (text input, multi-line via basic text input for now)

Form Navigation:
- `Tab`/`Down`: Move to the next field (cycles through visible fields)
- `Shift+Tab`/`Up`: Move to the previous field (cycles through visible fields)
- `Enter`: Submit form / Toggle checkbox if focused
- `Space`: Toggle checkbox if focused
- `Esc`: Cancel and return to list view

### Confirmation View
Displays a simple dialog asking for confirmation (e.g., before deleting a server) with Yes/No options.
- `Left`/`Right` arrows: Select Yes/No
- `Enter`: Confirm selection
- `Esc`: Cancel

## Technical Requirements

### Configuration Handling
The application needs to:
1. Read configuration files using the `types.ServerConfigEditor` interface (`config.CursorMCPEditor` or `config.ClaudeDesktopEditor`).
2. Modify configurations through the same editor interface.
3. Save changes back to disk via the `Save()` method on the editor.
4. Handle errors gracefully (displaying messages in the UI).

### UI Component Architecture
The TUI is built with Bubbletea following the Elm architecture:
- **Model**: `tui.Model` stores application state including the current mode, active list, form state, confirmation state, and the `currentEditor` interface.
- **Update**: `Model.Update` processes messages (`tea.Msg`) including key presses and results from async commands (like loading/saving files) to update the application state.
- **View**: `Model.View` renders the current state (menu, list, form, or confirmation dialog) to the terminal using `lipgloss` for styling.

### Pseudocode Structure (Simplified Conceptual)

```go
// Main application model
type Model struct {
	currentEditor types.ServerConfigEditor // Interface for backend
	configType    tui.ConfigType           // cursor, claude, or none
	activeList    *list.Model
	menuList      list.Model
	formState     tui.FormModel
	confirmState  tui.ConfirmModel
	mode          tui.mode                 // menu, list, addEdit, confirm
	errorMessage  string
	help          help.Model
	keys          tui.keyMap
}

// Initialize application
func NewModel() Model {
	// Initialize lists, help, keys...
	return Model{mode: tui.modeMenu, ...}
}

// Update function processes messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keyboard input based on m.mode
		// Switch mode (e.g., list -> addEdit)
		// Trigger commands (e.g., m.loadServers(tui.ConfigTypeCursor))

	case loadedServersMsg:
		// Populate m.activeList with servers from msg
		// Set m.currentEditor = msg.editor
		// Set m.mode = tui.modeList

	case serverSavedMsg, serverDeletedMsg, serverToggleEnabledMsg:
		// Handle success/failure, potentially show error/status msg
		// Reload server list: return m, m.loadServers(m.configType)

	case ConfirmMsg:
		// If msg.Confirmed, trigger the relevant action (e.g., delete)
		// Set m.mode = tui.modeList
	}
	// Update sub-components (lists, form, etc.)
	// ...
	return m, tea.Batch(cmds...)
}

// View function renders UI
func (m Model) View() string {
	switch m.mode {
	case tui.modeMenu:
		// Render m.menuList and help
	case tui.modeList:
		// Render m.activeList and help
	case tui.modeAddEdit:
		// Render m.formState and help
	case tui.modeConfirm:
		// Render m.confirmState centered
	}
	// Render error message if present
}
```

### Server Operations (Conceptual - Handled by Commands in `model.go`)

```go
// Example: deleteServer command function in model.go
func (m *Model) deleteServer(name string) tea.Cmd {
	return func() tea.Msg {
		if m.currentEditor == nil {
			return serverDeletedMsg{..., err: errors.New("no editor")}
		}
		err := m.currentEditor.RemoveMCPServer(name)
		if err != nil {
			return serverDeletedMsg{..., err: err}
		}
		err = m.currentEditor.Save()
		// Handle save error
		return serverDeletedMsg{serverName: name, err: saveErr}
	}
}

// Example: saveServer command function in model.go
func (m *Model) saveServer(server types.CommonServer, overwrite bool) tea.Cmd {
	return func() tea.Msg {
		if m.currentEditor == nil {
			return serverSavedMsg{..., err: errors.New("no editor")}
		}
		err := m.currentEditor.AddMCPServer(server, overwrite)
		if err != nil {
			return serverSavedMsg{..., err: err}
		}
		err = m.currentEditor.Save()
		// Handle save error
		return serverSavedMsg{serverName: server.Name, err: saveErr}
	}
}
```

## Implementation Plan

### Phase 1: Basic Structure (Completed)
1. Set up bubbletea application skeleton
2. Implement main menu and navigation
3. Create configuration loading/saving logic (via editor interface)

### Phase 2: Server List View (Completed)
1. Implement server list display with status indicators
2. Add server selection and basic operations (enable/disable)
3. Implement confirmation dialogs for destructive actions (delete)

### Phase 3: Add/Edit Functionality (Partially Completed)
1. Create form component for server configuration (done)
2. Implement validation for form inputs (basic checks done)
3. Connect form submission to configuration editors (done)
4. Add checkbox for SSE/stdio type and conditional fields (done)
5. *Remaining: Improve Env input, enhance validation.* 

### Phase 4: Polish and Integration (In Progress)
1. Add error handling and user feedback (basic error display done)
2. Implement help screens (context-aware help done)
3. Refine UI styling using Lipgloss (basic styling done)
4. Add keyboard shortcut overlay (done via help bubble)
5. *Remaining: Add status messages, improve styling, consolidate helpers.* 

## Integration with Existing Code

The TUI is added as a subcommand `ui`:

```go
// In cmd/go-go-mcp/cmds/ui_cmd.go
func NewUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Interactive terminal UI for managing MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup logging flags if needed
			p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	}
	// Add flags for logging, etc.
	return cmd
}

// In cmd/go-go-mcp/main.go
func initRootCmd() *cobra.Command {
    // ... other commands
    rootCmd.AddCommand(cmds.NewUICommand())
    // ...
}
```

## UI Mockups (Updated)

**List View:**
```
Cursor MCP Servers

  claude-3-opus (CMD: /path/to/claude-opus) (enabled)
> anthropic-api (SSE: http://localhost:8080/sse) (enabled)
  openai-gpt4 (CMD: /path/to/openai-gpt4) (disabled)


[a add | e edit | d delete | space toggle | esc back | ? help | q quit]
```

**Add/Edit Form (Stdio type selected):**
```
Edit Server

Server name (required)
my-cmd-server
[ ] SSE (vs stdio)

Command path (stdio)
/usr/local/bin/mcp
Arguments (stdio, space separated)
server start --profile custom
Environment variables (KEY=VALUE, one per line)
ANTHROPIC_API_KEY=********|


[enter submit/toggle | esc cancel | tab/↓ next field | shift+tab/↑ prev field | space toggle checkbox]
```

**Add/Edit Form (SSE type selected):**
```
Edit Server

Server name (required)
my-sse-server
[x] SSE (vs stdio)

SSE URL
http://localhost:8080/events|
Environment variables (KEY=VALUE, one per line)
MY_VAR=some_value


[enter submit/toggle | esc cancel | tab/↓ next field | shift+tab/↑ prev field | space toggle checkbox]
```

## Success Metrics
- User feedback reporting improved workflow efficiency
- Reduced errors in configuration files
- Increased usage of advanced server configurations
- Less time spent in CLI command references

## Future Enhancements
- Configuration file validation and repair
- Server status monitoring
- Server log viewing integration
- Use `textarea` bubble for multi-line Env input
