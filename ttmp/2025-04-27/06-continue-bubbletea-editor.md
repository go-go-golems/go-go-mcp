# Continuing the Bubbletea UI Implementation for MCP Server Management

## Project Overview

The go-go-mcp project is implementing a terminal user interface (TUI) using the Bubble Tea framework (github.com/charmbracelet/bubbletea) to provide an interactive way to manage MCP server configurations for both Cursor and Claude desktop applications. This document provides the necessary context and guidance to continue the implementation.

## Current Implementation Status

We have created the basic structure for a Bubble Tea application that allows users to:
1. Navigate a main menu to select which configuration type to manage (Cursor or Claude).
2. View a list of configured servers, showing their name, type (CMD/SSE), primary identifier (Command/URL), and enabled/disabled status.
3. Navigate between different views (menu, list, add/edit form, confirmation dialog).
4. Enable or disable servers using the spacebar in the list view.
5. Delete servers using the 'd' key, with a confirmation dialog.
6. Enter add mode ('a') or edit mode ('e') for the selected server, which displays a form.
7. The form allows editing the server name, type (via a checkbox), command/URL (conditionally shown), arguments (conditionally shown), and environment variables.
8. Navigate the form fields using Tab/Shift+Tab (or Up/Down arrows).
9. Toggle the server type between stdio (command/args) and SSE (url) using the checkbox (Space or Enter when focused).
10. Cancel form edits (Esc) or submit them (Enter).
11. Display context-aware help messages for each view.

The following files have been created and modified:

### 1. Command Entry Point

**File:** `cmd/go-go-mcp/cmds/ui_cmd.go`
- Defines the Cobra command `mcp ui` to launch the TUI.
- Initializes the Bubble Tea program with the main model.

### 2. Main Model

**File:** `pkg/ui/tui/model.go`
- Defines the main application `Model` and implements the Bubble Tea Model interface (Init, Update, View).
- Contains:
    - A `keyMap` for keyboard shortcuts.
    - A `help.Model` for context-aware help display.
    - An enum `ConfigType` (ConfigTypeCursor, ConfigTypeClaude, ConfigTypeNone).
    - A `mode` enum (modeMenu, modeList, modeAddEdit, modeConfirm).
    - A `list.Model` for the main menu (`menuList`).
    - A pointer `*list.Model` for the currently active server list (`activeList`).
    - A `types.ServerConfigEditor` interface (`currentEditor`) to interact with the configuration backend.
    - State for the add/edit form (`formState`).
    - State for the confirmation dialog (`confirmDialog`).
    - Error message handling.
- Defines item types for lists (`listItem`, `serverItem`).
- Defines message types for async operations (`loadedServersMsg`, `serverDeletedMsg`, `serverToggleEnabledMsg`, `serverSavedMsg`, `errorMsg`).
- Implements `Update` logic for handling key presses, mode switching, and messages from commands.
- Implements `View` logic to render different UI states based on the current `mode`.
- Implements `ShortHelp` and `FullHelp` for the `help.Model`.
- Contains command functions (`loadServers`, `deleteServer`, `toggleServerEnabled`, `saveServer`) that interact with the `currentEditor` and return `tea.Msg`.

### 3. Form Handler

**File:** `pkg/ui/tui/form.go`
- Implements the `FormModel` struct for adding and editing servers.
- Uses focus constants (`focusName`, `focusType`, `focusCommand`, etc.) for navigation.
- Defines text inputs for server details (`nameInput`, `commandInput`, `urlInput`, `argsInput`, `envInput`).
- Includes a boolean `isSSE` to track the server type, toggled by a checkbox.
- Implements `Update` logic:
    - Handles navigation (Tab/Shift+Tab) between visible fields (Name -> Type Checkbox -> Command/Args or URL -> Env).
    - Handles checkbox toggling (Enter/Space when focused).
    - Handles form submission (Enter when a text input is focused) and cancellation (Esc).
    - Passes key events to the currently focused text input.
- Implements `View` logic:
    - Renders the form title ("Add Server" or "Edit Server").
    - Renders the Name input.
    - Renders the Type checkbox `[ ] SSE (vs stdio)`, highlighting it when focused.
    - Conditionally renders URL input (if `isSSE` is true) or Command/Args inputs (if `isSSE` is false).
    - Renders the Env input.
    - Renders dynamic help text based on `FormKeyMap`.
- Includes helper methods like `updateFocus` (to manage text input focus) and `ToServer` (to convert form state to a `types.CommonServer`).

### 4. Confirmation Dialog

**File:** `pkg/ui/tui/confirm.go`
- Implements a simple `ConfirmModel` for yes/no dialogs.
- Handles key presses for selection (Left/Right arrows) and confirmation/cancellation (Enter/Esc).
- Renders the dialog with a title, message, and styled Yes/No options.

### 5. Configuration Interfaces & Types

**File:** `pkg/mcp/types/config.go` (and implementations in `pkg/config/`)
- Defines the `ServerConfigEditor` interface used by the TUI model to abstract backend operations.
- Defines the `CommonServer` struct used to pass server data between the form and the editor.

**Interface:** `types.ServerConfigEditor`
```go
// ServerConfigEditor defines the interface for managing server configurations.
type ServerConfigEditor interface {
	ListServers() (map[string]CommonServer, error)
	GetServer(name string) (CommonServer, bool, error)
	ListDisabledServers() ([]string, error)
	IsServerDisabled(name string) (bool, error)
	EnableMCPServer(name string) error
	DisableMCPServer(name string) error
	AddMCPServer(server CommonServer, overwrite bool) error
	RemoveMCPServer(name string) error
	Save() error
	GetConfigPath() string
}
```

**Struct:** `types.CommonServer`
```go
type CommonServer struct {
	Name    string            `json:"-"` // Name is map key in config
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"` // Specific to Cursor SSE
	IsSSE   bool              `json:"is_sse,omitempty"` // Indicates if it's an SSE type
}
```

**Implementations:**
- `config.CursorMCPEditor` and `config.ClaudeDesktopEditor` implement `types.ServerConfigEditor`.

## Tasks to Complete

### 1. Form Functionality

- [x] Implement the `View()` method to properly render form fields
- [x] Add toggle functionality for the Cursor SSE/Command type (via checkbox)
- [x] Conditionally hide/show URL or Command/Args fields based on type
- [x] Improve form validation (basic name/URL/command checks exist in `ToServer`)
- [ ] Enhance Env input (currently basic textinput, consider textarea bubble)
- [ ] Add styling with lipgloss (basic styling exists, can be enhanced)
- [x] Implement form submission handling (calls `saveServer` command)

### 2. Server Operations

- [x] Server deletion with confirmation dialog
- [x] Server enabling/disabling
- [x] Loading selected server data into the form for editing
- [x] Saving changes back to the configuration files (via `saveServer` command triggered by form submission)

### 3. Confirmation Dialog

- [x] Implement a simple dialog model (`confirm.go`)
- [x] Add yes/no navigation
- [x] Handle confirmation actions (triggers `deleteServer` command)

### 4. Style and Polish

- [ ] Add consistent styling using lipgloss (further improvements possible)
- [x] Improve the help text and keyboard shortcut display (context-aware)
- [ ] Add status messages for operations (e.g., "Server saved successfully")
- [x] Implement basic error handling and display (via `errorMsg` at bottom)

### 5. Known Issues/Refinements

- [ ] The `argsInput` and `envInput` parsing helpers (`parseArgsString`, `parseEnvString`) are duplicated in `form.go` and `model.go`. Consolidate them.
- [ ] Env input field could be improved (e.g., using `textarea` bubble).
- [ ] Add more robust error handling (e.g., display errors within the form).

## Technical Details

### Bubble Tea Basics

Bubble Tea follows the Elm architecture with three main components:

1.  **Model**: The application state (`tui.Model`)
2.  **Update**: A function that updates the state based on messages (`Model.Update`)
3.  **View**: A function that renders the current state (`Model.View`)

Messages (`tea.Msg`) are passed to the Update function, which returns an updated model and commands (`tea.Cmd`) to run (e.g., for async operations like loading/saving files).

### Configuration Abstraction

The UI uses the `types.ServerConfigEditor` interface to interact with the configuration backend. This decouples the UI from the specific implementations (`config.CursorMCPEditor`, `config.ClaudeDesktopEditor`). The `currentEditor` field in the `Model` holds the active editor, determined when the user selects "Manage Cursor Configuration" or "Manage Claude Desktop Configuration" from the main menu.

### Key Files and Functions to Modify for Next Steps

1.  **form.go**:
    *   Consider replacing `envInput` with `textarea.Model`.
    *   Improve input validation within `ToServer` or add validation messages to the UI.
2.  **model.go**:
    *   Add user-friendly status messages after save/delete/toggle operations (perhaps clearing `errorMsg` on success and setting a temporary success message).
    *   Refactor parsing helpers (`parseArgsString`, `parseEnvString`) to avoid duplication (e.g., move to a shared utility file or keep them in `model.go` and call from `form.go`).
3.  **General Styling**:
    *   Apply more `lipgloss` styling for a polished look.

## Useful Resources

1.  **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
2.  **Bubble Tea Examples**: https://github.com/charmbracelet/bubbletea/tree/master/examples
3.  **Lipgloss Documentation**: https://github.com/charmbracelet/lipgloss
4.  **Bubbles Components (including textarea)**: https://github.com/charmbracelet/bubbles

## Development Approach

1.  Focus on improving the Env input field in the form.
2.  Add status messages for feedback after operations.
3.  Refactor parsing helpers.
4.  Refine styling and error display.

Use the `go run ./cmd/go-go-mcp ui --log-level debug --log-file /tmp/ui.log` command to test changes and view logs.

## Design Guidelines

- Keep the UI simple and intuitive
- Use consistent styling for all elements
- Provide clear feedback for all operations
- Handle errors gracefully
- Follow the keyboard navigation patterns used in other Bubble Tea applications

Good luck with continuing the implementation! 