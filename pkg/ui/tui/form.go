package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// Focus indices for the form fields
const (
	focusName = iota
	focusType // The new checkbox
	focusCommand
	focusArgs
	focusURL
	focusEnv
	focusMax // Keep track of the total number of potential focus points
)

// FormKeyMap defines keybindings specific to the form view
type FormKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
	Next   key.Binding // Tab
	Prev   key.Binding // Shift+Tab
	// ToggleSSE key.Binding // Removed
}

var defaultFormKeyMap = FormKeyMap{
	Submit: key.NewBinding(
		key.WithKeys("alt+s"),
		key.WithHelp("alt+s", "submit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Next: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	// ToggleSSE: key.NewBinding( ... ) // Removed
}

// FormModel represents the form for adding/editing servers
type FormModel struct {
	keyMap       FormKeyMap
	nameInput    textinput.Model
	commandInput textinput.Model // Used for stdio
	urlInput     textinput.Model // Used for SSE
	argsInput    textinput.Model // Used for stdio
	envInput     textarea.Model  // Changed from textinput to textarea
	isSSE        bool            // True if the server uses SSE, false for stdio
	isAddMode    bool            // True if adding a new server, false if editing
	activeInput  int             // Index of the currently focused element (using constants)
	submitted    bool            // Flag indicating form submission was triggered
	cancelled    bool            // Flag indicating form cancellation was triggered
}

// NewFormModel creates a new form model with initialized inputs
func NewFormModel() FormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Server name (required)"
	nameInput.Focus() // Start focus on name
	nameInput.CharLimit = 50
	nameInput.Width = 80

	commandInput := textinput.New()
	commandInput.Placeholder = "Command path (stdio)"
	commandInput.CharLimit = 200
	commandInput.Width = 80

	urlInput := textinput.New()
	urlInput.Placeholder = "SSE URL"
	urlInput.CharLimit = 200
	urlInput.Width = 80

	argsInput := textinput.New()
	argsInput.Placeholder = "Arguments (stdio, space separated)"
	argsInput.CharLimit = 500
	argsInput.Width = 80

	// Create textarea for envInput instead of textinput
	envInput := textarea.New()
	envInput.Placeholder = "Environment variables (KEY=VALUE, one per line)"
	envInput.CharLimit = 1000
	envInput.SetWidth(80)
	envInput.SetHeight(5) // Show 5 lines by default

	return FormModel{
		keyMap:       defaultFormKeyMap,
		nameInput:    nameInput,
		commandInput: commandInput,
		urlInput:     urlInput,
		argsInput:    argsInput,
		envInput:     envInput,
		activeInput:  focusName, // Start focus on name
		isSSE:        false,     // Default to stdio
		isAddMode:    true,      // Default
	}
}

// Update handles form input messages
func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Special handling for textarea - pass arrow keys through when focused
		if m.activeInput == focusEnv {
			// Let special navigation keys like arrows pass through to textarea
			if msg.Type == tea.KeyUp || msg.Type == tea.KeyDown ||
				msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight {
				var cmd tea.Cmd
				m.envInput, cmd = m.envInput.Update(msg)
				return m, cmd
			}
		}

		// Global key handling - only handle special keys like ESC, Alt+s
		// Let other keys (like 'q') pass through to the active input
		switch {
		case key.Matches(msg, m.keyMap.Cancel):
			m.cancelled = true
			// Ensure all inputs are blurred on cancel
			m.nameInput.Blur()
			m.commandInput.Blur()
			m.urlInput.Blur()
			m.argsInput.Blur()
			m.envInput.Blur()
			return m, nil

		case key.Matches(msg, m.keyMap.Submit):
			// For submit, mark as submitted and return
			m.submitted = true
			return m, nil

		case msg.Type == tea.KeySpace && m.activeInput == focusType:
			// Handle checkbox toggle with space
			m.isSSE = !m.isSSE
			cmds = append(cmds, m.updateFocus())
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keyMap.Next), key.Matches(msg, m.keyMap.Prev):
			direction := 1
			if key.Matches(msg, m.keyMap.Prev) {
				direction = -1
			}

			// Cycle through focusable elements, skipping hidden ones
			for {
				m.activeInput = (m.activeInput + direction + focusMax) % focusMax

				// Skip Command/Args if SSE is enabled
				if m.isSSE && (m.activeInput == focusCommand || m.activeInput == focusArgs) {
					continue
				}
				// Skip URL if SSE is disabled (stdio)
				if !m.isSSE && m.activeInput == focusURL {
					continue
				}
				// Found a valid, visible element to focus
				break
			}

			cmds = append(cmds, m.updateFocus())
			return m, tea.Batch(cmds...)

		case msg.Type == tea.KeyDown && m.activeInput != focusEnv:
			// Down key for navigation between fields when not in textarea
			// Cycle through focusable elements, skipping hidden ones
			for {
				m.activeInput = (m.activeInput + 1) % focusMax

				// Skip Command/Args if SSE is enabled
				if m.isSSE && (m.activeInput == focusCommand || m.activeInput == focusArgs) {
					continue
				}
				// Skip URL if SSE is disabled (stdio)
				if !m.isSSE && m.activeInput == focusURL {
					continue
				}
				// Found a valid, visible element to focus
				break
			}

			cmds = append(cmds, m.updateFocus())
			return m, tea.Batch(cmds...)

		case msg.Type == tea.KeyUp && m.activeInput != focusEnv:
			// Up key for navigation between fields when not in textarea
			// Cycle through focusable elements, skipping hidden ones
			for {
				m.activeInput = (m.activeInput - 1 + focusMax) % focusMax

				// Skip Command/Args if SSE is enabled
				if m.isSSE && (m.activeInput == focusCommand || m.activeInput == focusArgs) {
					continue
				}
				// Skip URL if SSE is disabled (stdio)
				if !m.isSSE && m.activeInput == focusURL {
					continue
				}
				// Found a valid, visible element to focus
				break
			}

			cmds = append(cmds, m.updateFocus())
			return m, tea.Batch(cmds...)
		}

		// --- Process the event in the active text input --- //
		// This allows keys like 'q' to go to the text input if focused
		var cmd tea.Cmd
		switch m.activeInput {
		case focusName:
			m.nameInput, cmd = m.nameInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case focusCommand:
			if !m.isSSE {
				m.commandInput, cmd = m.commandInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case focusArgs:
			if !m.isSSE {
				m.argsInput, cmd = m.argsInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case focusURL:
			if m.isSSE {
				m.urlInput, cmd = m.urlInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case focusEnv:
			// Update for textarea
			var cmd tea.Cmd
			m.envInput, cmd = m.envInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case focusType:
			// Special case for the type checkbox focus - handle Enter as toggle
			if msg.Type == tea.KeyEnter {
				m.isSSE = !m.isSSE
				return m, nil
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// updateFocus manages blurring/focusing the correct text input.
func (m *FormModel) updateFocus() tea.Cmd {
	// Blur all inputs first
	m.nameInput.Blur()
	m.commandInput.Blur()
	m.argsInput.Blur()
	m.urlInput.Blur()
	m.envInput.Blur()

	// Initialize command slice
	var cmds []tea.Cmd

	// Focus the correct input based on activeInput and isSSE
	switch m.activeInput {
	case focusName:
		cmds = append(cmds, m.nameInput.Focus())
	case focusCommand:
		if !m.isSSE {
			cmds = append(cmds, m.commandInput.Focus())
		}
	case focusArgs:
		if !m.isSSE {
			cmds = append(cmds, m.argsInput.Focus())
		}
	case focusURL:
		if m.isSSE {
			cmds = append(cmds, m.urlInput.Focus())
		}
	case focusEnv:
		cmds = append(cmds, m.envInput.Focus())
		// case focusType: no text input to focus
	}

	return tea.Batch(cmds...)
}

// View renders the form
func (m FormModel) View() string {
	var sb strings.Builder

	title := "Add Server"
	if !m.isAddMode {
		title = "Edit Server"
	}
	sb.WriteString(title + "\n\n")

	// Name Input (Always visible)
	sb.WriteString(m.nameInput.View() + "\n")

	// Type Checkbox
	checkbox := "[ ]"
	if m.isSSE {
		checkbox = "[x]"
	}
	label := " SSE (vs stdio)"
	checkboxView := lipgloss.JoinHorizontal(lipgloss.Left, checkbox, label)
	if m.activeInput == focusType {
		checkboxView = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(checkboxView)
	}
	sb.WriteString(checkboxView + "\n\n")

	// Conditional Fields
	if m.isSSE {
		// SSE Type: URL
		sb.WriteString(m.urlInput.View() + "\n")
	} else {
		// Stdio Type: Command, Args
		sb.WriteString(m.commandInput.View() + "\n")
		sb.WriteString(m.argsInput.View() + "\n")
	}

	// Env Input (Always visible) - now a textarea
	sb.WriteString(m.envInput.View())

	sb.WriteString("\n\n")
	sb.WriteString(m.helpView())

	return sb.String()
}

// helpView renders the help text for the form
func (m FormModel) helpView() string {
	// Build help keys list
	var keys []key.Binding

	// Common keys
	keys = append(keys, m.keyMap.Submit, m.keyMap.Cancel, m.keyMap.Next, m.keyMap.Prev)

	// Conditional help text based on focus
	switch m.activeInput {
	case focusType:
		keys = append(keys,
			key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle checkbox")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle checkbox (when focused)")),
		)
	case focusEnv:
		keys = append(keys,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "navigate in textarea")),
		)
	default:
		keys = append(keys,
			key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "previous/next field")),
		)
	}

	helpParts := make([]string, len(keys))
	for i, k := range keys {
		helpParts[i] = fmt.Sprintf("%s %s", k.Help().Key, k.Help().Desc)
	}

	return lipgloss.NewStyle().Faint(true).Render(strings.Join(helpParts, " | "))
}

// Reset clears all form inputs and resets state
func (m *FormModel) Reset() tea.Cmd {
	m.nameInput.Reset()
	m.commandInput.Reset()
	m.urlInput.Reset()
	m.argsInput.Reset()
	m.envInput.Reset()        // Textarea also has Reset() method
	m.activeInput = focusName // Reset focus to name
	m.isSSE = false           // Reset to stdio
	m.isAddMode = true
	m.submitted = false
	m.cancelled = false
	return m.updateFocus() // Return the command from updateFocus
}

// ToServer converts the current form state into a CommonServer object.
// It performs basic validation (e.g., name is required).
func (m *FormModel) ToServer() (types.CommonServer, error) {
	server := types.CommonServer{}

	name := m.nameInput.Value()
	if name == "" {
		return server, fmt.Errorf("server name is required")
	}
	server.Name = name

	server.IsSSE = m.isSSE
	if m.isSSE {
		url := m.urlInput.Value()
		if url == "" {
			return server, fmt.Errorf("URL is required for SSE server type")
		}
		server.URL = url
		// Command/Args are ignored for SSE type in the backend, but we parse env
		server.Env = parseEnvString(m.envInput.Value())
	} else {
		cmd := m.commandInput.Value()
		if cmd == "" {
			return server, fmt.Errorf("command is required for non-SSE server type")
		}
		server.Command = cmd
		server.Args = parseArgsString(m.argsInput.Value())
		server.Env = parseEnvString(m.envInput.Value())
	}

	return server, nil
}
