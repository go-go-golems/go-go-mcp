# Product Requirements Document: Bubbletea UI for MCP Server Management

## Overview
This document outlines the requirements for implementing a terminal-based user interface using the charmbracelet/bubbletea library to manage MCP server configurations for both Cursor and Claude desktop applications. The UI will provide an intuitive interface for viewing, adding, removing, enabling, and disabling server configurations without requiring users to edit JSON configuration files directly.

## Background
The go-go-mcp project currently manages server configurations through command-line tools that modify JSON configuration files. While functional, this approach requires users to remember specific commands and syntax. A terminal UI would improve usability by providing visual feedback and reducing cognitive load.

## Goals
- Create an intuitive TUI for managing MCP server configurations
- Support both Cursor and Claude desktop configurations
- Allow users to view, add, edit, remove, enable, and disable server configurations
- Maintain compatibility with existing configuration file formats
- Provide immediate visual feedback for user actions

## Non-Goals
- Replacing the existing CLI commands
- Modifying the configuration file format
- Providing WYSIWYG editing of raw JSON
- Supporting configuration options beyond server management

## User Experience

### Navigation and Layout
The application will use a multi-panel interface with:
1. A header displaying the current mode and configuration file being edited
2. A main content area showing server configurations
3. A footer displaying available commands and keyboard shortcuts

### Main Menu
The initial screen presents options for:
- Manage Cursor Configuration
- Manage Claude Desktop Configuration
- Exit

### Server List View
When a configuration type is selected, the screen will display:
- List of configured servers (with enabled/disabled status indicators)
- Navigation controls
- Action shortcuts (add, edit, remove, enable/disable)

### Add/Edit Server View
A form interface with fields for:
- Server name
- Command or URL
- Arguments (for command-based servers)
- Environment variables
- Server type (command or SSE for Cursor)

## Technical Requirements

### Configuration Handling
The application needs to:
1. Read configuration files using existing editor classes
2. Modify configurations through the same editor interfaces
3. Save changes back to disk
4. Handle errors gracefully

### UI Component Architecture
The TUI will be built with Bubbletea following the Elm architecture:
- **Model**: Stores application state including configuration data
- **Update**: Processes messages to update application state
- **View**: Renders the current state to the terminal

### Pseudocode Structure

```go
// Main application model
type model struct {
    configType       string // "cursor" or "claude"
    cursorEditor     *config.CursorMCPEditor
    claudeEditor     *config.ClaudeDesktopEditor
    servers          []serverConfig
    selectedServer   int
    mode             string // "list", "add", "edit", "confirm"
    formState        serverFormState
    errorMessage     string
}

// Server form state for add/edit operations
type serverFormState struct {
    name        string
    command     string
    url         string
    args        []string
    env         map[string]string
    cursorType  string // "command" or "sse"
}

// Initialize application
func initialModel() model {
    // Load configurations if files exist
    // Return initial model
}

// Update function processes messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle keyboard input
        // Navigation, selection, form input

    case saveConfigMsg:
        // Save changes to config file
        
    case loadConfigMsg:
        // Load config data
    }
    
    return m, nil
}

// View function renders UI
func (m model) View() string {
    switch m.mode {
    case "menu":
        // Render main menu
    case "list":
        // Render server list
    case "add", "edit":
        // Render form
    case "confirm":
        // Render confirmation dialog
    }
}
```

### Server Operations Pseudocode

```go
// Add server to configuration
func (m *model) addServer() error {
    if m.configType == "cursor" {
        if m.formState.cursorType == "command" {
            return m.cursorEditor.AddMCPServer(
                m.formState.name,
                m.formState.command,
                m.formState.args,
                m.formState.env,
                false, // Don't overwrite
            )
        } else {
            return m.cursorEditor.AddMCPServerSSE(
                m.formState.name,
                m.formState.url,
                m.formState.env,
                false, // Don't overwrite
            )
        }
    } else {
        return m.claudeEditor.AddMCPServer(
            m.formState.name,
            m.formState.command,
            m.formState.args,
            m.formState.env,
            false, // Don't overwrite
        )
    }
}

// Enable server
func (m *model) enableServer(name string) error {
    if m.configType == "cursor" {
        return m.cursorEditor.EnableMCPServer(name)
    } else {
        return m.claudeEditor.EnableMCPServer(name)
    }
}

// Disable server
func (m *model) disableServer(name string) error {
    if m.configType == "cursor" {
        return m.cursorEditor.DisableMCPServer(name)
    } else {
        return m.claudeEditor.DisableMCPServer(name)
    }
}

// Remove server
func (m *model) removeServer(name string) error {
    if m.configType == "cursor" {
        return m.cursorEditor.RemoveMCPServer(name)
    } else {
        return m.claudeEditor.RemoveMCPServer(name)
    }
}
```

## Implementation Plan

### Phase 1: Basic Structure
1. Set up bubbletea application skeleton
2. Implement main menu and navigation
3. Create configuration loading/saving logic

### Phase 2: Server List View
1. Implement server list display with status indicators
2. Add server selection and basic operations (enable/disable)
3. Implement confirmation dialogs for destructive actions

### Phase 3: Add/Edit Functionality
1. Create form component for server configuration
2. Implement validation for form inputs
3. Connect form submission to configuration editors

### Phase 4: Polish and Integration
1. Add error handling and user feedback
2. Implement help screens and documentation
3. Refine UI styling using Lipgloss
4. Add keyboard shortcut overlay

## Integration with Existing Code

The TUI will be added as a new subcommand in the CLI structure:

```go
func NewServerManagerCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "ui",
        Short: "Interactive terminal UI for managing MCP servers",
        Long:  `A terminal user interface for managing MCP server configurations for Cursor and Claude desktop.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            p := tea.NewProgram(initialModel())
            _, err := p.Run()
            return err
        },
    }
    
    return cmd
}

// Add to main.go or relevant command file
func init() {
    rootCmd.AddCommand(NewServerManagerCommand())
}
```

## UI Mockups

```
┌─────────────────────────────────────────────────────┐
│ MCP Server Manager - Cursor Configuration           │
├─────────────────────────────────────────────────────┤
│ > claude-3-opus (enabled)                           │
│   anthropic-api (enabled)                           │
│   openai-gpt4 (disabled)                            │
│   groq-mixtral (enabled)                            │
│                                                     │
│                                                     │
│                                                     │
├─────────────────────────────────────────────────────┤
│ j/k: navigate  e: enable/disable  a: add  d: delete │
│ Enter: edit  q: back  ?: help                       │
└─────────────────────────────────────────────────────┘
```

```
┌─────────────────────────────────────────────────────┐
│ Add New Server - Cursor Configuration               │
├─────────────────────────────────────────────────────┤
│ Name: my-custom-server                              │
│ Type: [x] Command  [ ] SSE                          │
│                                                     │
│ Command: /usr/local/bin/mcp                         │
│ Arguments: server start --profile custom            │
│                                                     │
│ Environment Variables:                              │
│ ANTHROPIC_API_KEY=*********                         │
│ + Add variable                                      │
├─────────────────────────────────────────────────────┤
│ Tab: next field  Enter: save  Esc: cancel           │
└─────────────────────────────────────────────────────┘
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