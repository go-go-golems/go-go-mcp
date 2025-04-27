# Continuing the Bubbletea UI Implementation for MCP Server Management

## Project Overview

The go-go-mcp project is implementing a terminal user interface (TUI) using the Bubble Tea framework (github.com/charmbracelet/bubbletea) to provide an interactive way to manage MCP server configurations for both Cursor and Claude desktop applications. This document provides the necessary context and guidance to continue the implementation.

## Current Implementation Status

We have created the basic structure for a Bubble Tea application that allows users to:
1. Navigate a main menu to select which configuration type to manage (Cursor or Claude)
2. View a list of configured servers (although the display is still in progress)
3. Navigation between different views (menu, list, form, etc.)

The following files have been created:

### 1. Command Entry Point

**File:** `cmd/go-go-mcp/cmds/ui_cmd.go`
This file defines the Cobra command for launching the UI:

```go
package cmds

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-go-golems/go-go-mcp/pkg/ui/tui"
    "github.com/spf13/cobra"
)

func NewUICommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "ui",
        Short: "Interactive terminal UI for managing MCP servers",
        Long:  `A terminal user interface for managing MCP server configurations for Cursor and Claude desktop.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
            _, err := p.Run()
            return err
        },
    }

    return cmd
}
```

The UI command is registered in `cmd/go-go-mcp/main.go` within the `initRootCmd()` function:

```go
// Add UI command
rootCmd.AddCommand(mcp_cmds.NewUICommand())
```

### 2. Main Model

**File:** `pkg/ui/tui/model.go`
This file defines the main application model and implements the Bubble Tea Model interface (Init, Update, View). The model contains:

- A keyMap for keyboard shortcuts
- Multiple list models for different views (menu, cursor servers, claude servers)
- Mode tracking for different UI states (menu, list, add/edit, confirm)
- Configuration editors for both Cursor and Claude desktop
- Error message handling

The file defines several important types:
- `mode` enum (modeMenu, modeList, modeAddEdit, modeConfirm)
- `keyMap` struct for key bindings
- `serverItem` for the server list items
- `listItem` for menu items
- `Model` as the main application model

### 3. Form Handler

**File:** `pkg/ui/tui/form.go`
This file implements the form interface for adding and editing servers:

- `FormModel` struct with text inputs for server details
- Methods to handle form navigation and input
- Basic view rendering (not fully implemented)
- Helper methods for handling form data

### 4. Data Handlers

**File:** `pkg/ui/tui/handlers.go`
This file contains functions for loading and manipulating server configurations:

- Message types for loading servers
- Functions to load server configurations from both types of config files
- Helper functions for converting between data formats
- Utility functions for parsing arguments and environment variables

## Configuration Interfaces

The UI interacts with two configuration interfaces:

1. **Cursor MCP Editor**: `pkg/config/cursor.go`
   - `CursorMCPEditor` struct for editing Cursor MCP configurations
   - Functions for adding, removing, enabling, and disabling servers

2. **Claude Desktop Editor**: `pkg/config/claude_desktop.go`
   - `ClaudeDesktopEditor` struct for editing Claude desktop configurations
   - Similar functions to the Cursor editor

Both editors provide a similar interface but work with slightly different configuration formats.

## Tasks to Complete

### 1. Complete the Form View

- [ ] Implement the `View()` method to properly render form fields
- [ ] Add toggle functionality for the Cursor SSE/Command type
- [ ] Improve form validation and submission handling
- [ ] Add styling with lipgloss (github.com/charmbracelet/lipgloss)

### 2. Implement Server Operations

- [x] Server deletion with confirmation dialog
- [x] Server enabling/disabling (Note: User reported potential issue with spacebar, needs verification)
- [x] Loading selected server data into the form for editing
- [x] Saving changes back to the configuration files (partially done via enable/disable/delete, needs add/edit save)

### 3. Confirmation Dialog

- [x] Implement a simple dialog model
- [x] Add yes/no navigation
- [x] Handle confirmation actions

### 4. Style and Polish

- [ ] Add consistent styling using lipgloss
- [x] Improve the help text and keyboard shortcut display (now context-aware)
- [ ] Add status messages for operations
- [ ] Implement proper error handling and display

## Technical Details

### Bubble Tea Basics

Bubble Tea follows the Elm architecture with three main components:

1. **Model**: The application state
2. **Update**: A function that updates the state based on messages
3. **View**: A function that renders the current state

Messages are passed to the Update function, which returns an updated model and commands to run.

### Refactoring: Abstracting Configuration Editing

To simplify the `Model` logic and make it easier to potentially support other configuration types in the future, we should introduce an interface to abstract the configuration editing operations.

**Proposed Interface:**

```go
// ServerConfigEditor defines the interface for managing server configurations.
type ServerConfigEditor interface {
	// ListServers retrieves all configured servers (both enabled and disabled).
	// It might be beneficial to return a map[string]CommonServer where CommonServer is a struct
	// that reconciles the fields of CursorMCPServer and MCPServer.
	ListServers() map[string]interface{} 

	// ListDisabledServers returns the names of disabled servers.
	ListDisabledServers() []string

	// EnableMCPServer enables a specific server.
	EnableMCPServer(name string) error

	// DisableMCPServer disables a specific server.
	DisableMCPServer(name string) error

	// AddMCPServer adds or updates a server.
	// This will likely need a common Server representation passed in.
	AddMCPServer(name string, server interface{}, overwrite bool) error

	// RemoveMCPServer removes a specific server.
	RemoveMCPServer(name string) error

	// Save persists the configuration changes.
	Save() error

	// GetConfigPath returns the path of the configuration file.
	GetConfigPath() string

    // IsServerDisabled checks if a server is disabled.
    IsServerDisabled(name string) bool

    // A common `Server` struct could be defined to unify CursorMCPServer and MCPServer:
    // type CommonServer struct {
    //     Name    string // Added for convenience in the UI
    //     Command string
    //     Args    []string
    //     Env     map[string]string
    //     URL     string // Specific to Cursor SSE
    //     Enabled bool   // Can be derived when loading
    // }
}
```

**Implementation:**

- Both `config.CursorMCPEditor` and `config.ClaudeDesktopEditor` will be updated to implement this `ServerConfigEditor` interface. This might require adjustments to method signatures or return types, potentially introducing adapter methods or a common `Server` struct to handle the differing fields between `CursorMCPServer` and `MCPServer`.

**Model Changes (`pkg/ui/tui/model.go`):**

- The `Model` struct will no longer hold separate `cursorEditor` and `claudeEditor` fields. Instead, it will have a single field:
  ```go
  currentEditor ServerConfigEditor
  ```
- The `configType` string field might still be kept for display logic, but the core operations will go through `currentEditor`.
- When loading a configuration type (e.g., in `loadCursorServers` or `loadClaudeServers`), the appropriate editor (`CursorMCPEditor` or `ClaudeDesktopEditor`) will be instantiated and assigned to `currentEditor` (cast to the interface type).
- All calls within `Model` that previously accessed `m.cursorEditor` or `m.claudeEditor` (like in `toggleServerEnabled`, `deleteServer`, `loadServerToForm`) will now use methods on `m.currentEditor`.

**Benefits:**

- **Reduced Complexity:** The `Model`'s `Update` function becomes significantly simpler as it doesn't need to constantly check `m.configType` before calling editor methods.
- **Improved Maintainability:** Easier to modify editor logic without touching the UI model extensively.
- **Extensibility:** Adding support for new configuration types would primarily involve creating a new struct that implements `ServerConfigEditor` and updating the loading logic in `Model`, minimizing changes to the core UI operation handlers.

### Key Files and Functions to Modify

1. **form.go**: 
   - `FormModel.View()` - Implement proper form rendering
   - Add form submission handling

2. **model.go**:
   - Add handlers for server operations in the `Update()` method
   - Improve the list view rendering

3. **handlers.go**:
   - Add functions for saving configuration changes
   - Implement server enable/disable operations

### Creating a Confirmation Dialog

Create a new file `pkg/ui/tui/confirm.go` with:

```go
package tui

import (
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// ConfirmModel represents a confirmation dialog
type ConfirmModel struct {
    title    string
    message  string
    selected bool // true = yes, false = no
}

// NewConfirmModel creates a new confirmation dialog
func NewConfirmModel(title, message string) ConfirmModel {
    return ConfirmModel{
        title:    title,
        message:  message,
        selected: false,
    }
}

// Update handles confirmation dialog input
func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
    // Handle key presses for yes/no selection and confirmation
    // ...
}

// View renders the confirmation dialog
func (m ConfirmModel) View() string {
    // Render the dialog with yes/no options
    // ...
}
```

### Context-Aware Help

The help view now dynamically changes based on the current mode (Menu, List, Add/Edit). The `Model` struct implements the `key.Map` interface, providing the relevant key bindings to the `help.Model` in the `View` function.

## Useful Resources

1. **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
2. **Bubble Tea Examples**: https://github.com/charmbracelet/bubbletea/tree/master/examples
3. **Lipgloss Documentation**: https://github.com/charmbracelet/lipgloss
4. **Bubbles Components**: https://github.com/charmbracelet/bubbles

## Development Approach

1. Start by completing the form view to allow adding new servers
2. Implement server listing with proper styling
3. Add server operations (enable/disable, edit, delete)
4. Create the confirmation dialog
5. Polish the UI and error handling

Use the `go run cmd/go-go-mcp/main.go ui` command to test your changes.

## Design Guidelines

- Keep the UI simple and intuitive
- Use consistent styling for all elements
- Provide clear feedback for all operations
- Handle errors gracefully
- Follow the keyboard navigation patterns used in other Bubble Tea applications

## Testing

Test each component individually:
1. Test form input and navigation
2. Test server listing and selection
3. Test configuration loading and saving
4. Test error handling

Use manual testing to verify the UI works as expected. There are no automated tests for the UI components yet.

Good luck with continuing the implementation! 