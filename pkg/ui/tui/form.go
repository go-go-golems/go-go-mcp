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

// Transport types
const (
	TransportStdio = "stdio"
	TransportHTTP  = "http"
	TransportSSE   = "sse"
)

// Focus indices for the form fields
const (
	focusName = iota
	focusType // The transport type radio buttons
	focusCommand
	focusArgs
	focusURL
	focusHeaders
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
	keyMap        FormKeyMap
	nameInput     textinput.Model
	commandInput  textinput.Model // Used for stdio
	urlInput      textinput.Model // Used for http/sse
	argsInput     textinput.Model // Used for stdio
	headersInput  textarea.Model  // Used for http/sse (KEY=VALUE format)
	envInput      textarea.Model  // Used for stdio (KEY=VALUE format)
	transportType string          // Transport type: stdio, http, or sse
	radioOption   int             // Currently selected radio option (0=stdio, 1=http, 2=sse)
	isAddMode     bool            // True if adding a new server, false if editing
	activeInput   int             // Index of the currently focused element (using constants)
	submitted     bool            // Flag indicating form submission was triggered
	cancelled     bool            // Flag indicating form cancellation was triggered
}

// NewFormModel creates a new form model with initialized inputs
func NewFormModel() FormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Server name (required)"
	nameInput.Focus() // Start focus on name
	nameInput.CharLimit = 50
	nameInput.Width = 80

	commandInput := textinput.New()
	commandInput.Placeholder = "Command path"
	commandInput.CharLimit = 200
	commandInput.Width = 80

	urlInput := textinput.New()
	urlInput.Placeholder = "URL (http:// or https://)"
	urlInput.CharLimit = 200
	urlInput.Width = 80

	argsInput := textinput.New()
	argsInput.Placeholder = "Arguments (space separated)"
	argsInput.CharLimit = 500
	argsInput.Width = 80

	// Create textarea for headers (http/sse)
	headersInput := textarea.New()
	headersInput.Placeholder = "Headers (KEY=VALUE, one per line)"
	headersInput.CharLimit = 1000
	headersInput.SetWidth(80)
	headersInput.SetHeight(5)

	// Create textarea for environment variables (stdio)
	envInput := textarea.New()
	envInput.Placeholder = "Environment variables (KEY=VALUE, one per line)"
	envInput.CharLimit = 1000
	envInput.SetWidth(80)
	envInput.SetHeight(5)

	return FormModel{
		keyMap:        defaultFormKeyMap,
		nameInput:     nameInput,
		commandInput:  commandInput,
		urlInput:      urlInput,
		argsInput:     argsInput,
		headersInput:  headersInput,
		envInput:      envInput,
		activeInput:   focusName,      // Start focus on name
		transportType: TransportStdio, // Default to stdio
		radioOption:   0,              // 0=stdio, 1=http, 2=sse
		isAddMode:     true,           // Default
	}
}

// Update handles form input messages
func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Special handling for textarea - pass arrow keys through when focused
		if m.activeInput == focusEnv || m.activeInput == focusHeaders {
			// Let special navigation keys like arrows pass through to textarea
			if msg.Type == tea.KeyUp || msg.Type == tea.KeyDown ||
				msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight {
				var cmd tea.Cmd
				if m.activeInput == focusEnv {
					m.envInput, cmd = m.envInput.Update(msg)
				} else {
					m.headersInput, cmd = m.headersInput.Update(msg)
				}
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
			m.headersInput.Blur()
			m.envInput.Blur()
			return m, nil

		case key.Matches(msg, m.keyMap.Submit):
			// For submit, mark as submitted and return
			m.submitted = true
			return m, nil

		case (msg.Type == tea.KeySpace || msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight) && m.activeInput == focusType:
			// Handle radio button navigation with space/left/right arrows
			//nolint:exhaustive // We only handle the keys in the case condition
			switch msg.Type {
			case tea.KeySpace, tea.KeyRight:
				// Move to next option
				m.radioOption = (m.radioOption + 1) % 3
			case tea.KeyLeft:
				// Move to previous option
				m.radioOption = (m.radioOption - 1 + 3) % 3
			}

			// Update transport type based on radio option
			switch m.radioOption {
			case 0:
				m.transportType = TransportStdio
			case 1:
				m.transportType = TransportHTTP
			case 2:
				m.transportType = TransportSSE
			}

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

				// Skip fields based on transport type
				if m.transportType == TransportStdio {
					// Skip URL and Headers for stdio
					if m.activeInput == focusURL || m.activeInput == focusHeaders {
						continue
					}
				} else {
					// Skip Command, Args, and Env for http/sse (use Headers instead)
					if m.activeInput == focusCommand || m.activeInput == focusArgs || m.activeInput == focusEnv {
						continue
					}
				}
				// Found a valid, visible element to focus
				break
			}

			cmds = append(cmds, m.updateFocus())
			return m, tea.Batch(cmds...)

		case msg.Type == tea.KeyDown && m.activeInput != focusEnv && m.activeInput != focusHeaders:
			// Down key for navigation between fields when not in textarea
			// Cycle through focusable elements, skipping hidden ones
			for {
				m.activeInput = (m.activeInput + 1) % focusMax

				// Skip fields based on transport type
				if m.transportType == TransportStdio {
					// Skip URL and Headers for stdio
					if m.activeInput == focusURL || m.activeInput == focusHeaders {
						continue
					}
				} else {
					// Skip Command, Args, and Env for http/sse (use Headers instead)
					if m.activeInput == focusCommand || m.activeInput == focusArgs || m.activeInput == focusEnv {
						continue
					}
				}
				// Found a valid, visible element to focus
				break
			}

			cmds = append(cmds, m.updateFocus())
			return m, tea.Batch(cmds...)

		case msg.Type == tea.KeyUp && m.activeInput != focusEnv && m.activeInput != focusHeaders:
			// Up key for navigation between fields when not in textarea
			// Cycle through focusable elements, skipping hidden ones
			for {
				m.activeInput = (m.activeInput - 1 + focusMax) % focusMax

				// Skip fields based on transport type
				if m.transportType == TransportStdio {
					// Skip URL and Headers for stdio
					if m.activeInput == focusURL || m.activeInput == focusHeaders {
						continue
					}
				} else {
					// Skip Command, Args, and Env for http/sse (use Headers instead)
					if m.activeInput == focusCommand || m.activeInput == focusArgs || m.activeInput == focusEnv {
						continue
					}
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
			if m.transportType == TransportStdio {
				m.commandInput, cmd = m.commandInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case focusArgs:
			if m.transportType == TransportStdio {
				m.argsInput, cmd = m.argsInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case focusURL:
			if m.transportType == TransportHTTP || m.transportType == TransportSSE {
				m.urlInput, cmd = m.urlInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case focusHeaders:
			if m.transportType == TransportHTTP || m.transportType == TransportSSE {
				m.headersInput, cmd = m.headersInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case focusEnv:
			if m.transportType == TransportStdio {
				m.envInput, cmd = m.envInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case focusType:
			// Special case for the type radio buttons focus - handle Enter as cycle
			if msg.Type == tea.KeyEnter {
				m.radioOption = (m.radioOption + 1) % 3
				switch m.radioOption {
				case 0:
					m.transportType = TransportStdio
				case 1:
					m.transportType = TransportHTTP
				case 2:
					m.transportType = TransportSSE
				}
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
	m.headersInput.Blur()
	m.envInput.Blur()

	// Initialize command slice
	var cmds []tea.Cmd

	// Focus the correct input based on activeInput and transport type
	switch m.activeInput {
	case focusName:
		cmds = append(cmds, m.nameInput.Focus())
	case focusCommand:
		if m.transportType == TransportStdio {
			cmds = append(cmds, m.commandInput.Focus())
		}
	case focusArgs:
		if m.transportType == TransportStdio {
			cmds = append(cmds, m.argsInput.Focus())
		}
	case focusURL:
		if m.transportType == TransportHTTP || m.transportType == TransportSSE {
			cmds = append(cmds, m.urlInput.Focus())
		}
	case focusHeaders:
		if m.transportType == TransportHTTP || m.transportType == TransportSSE {
			cmds = append(cmds, m.headersInput.Focus())
		}
	case focusEnv:
		if m.transportType == TransportStdio {
			cmds = append(cmds, m.envInput.Focus())
		}
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
	sb.WriteString(m.nameInput.View() + "\n\n")

	// Transport Type Radio Buttons
	sb.WriteString("Transport Type:\n")
	radioView := m.renderRadioButtons()
	if m.activeInput == focusType {
		radioView = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(radioView)
	}
	sb.WriteString(radioView + "\n\n")

	// Conditional Fields based on transport type
	switch m.transportType {
	case TransportStdio:
		sb.WriteString(m.commandInput.View() + "\n")
		sb.WriteString(m.argsInput.View() + "\n")
		sb.WriteString(m.envInput.View())
	case TransportHTTP, TransportSSE:
		sb.WriteString(m.urlInput.View() + "\n")
		sb.WriteString(m.headersInput.View())
	}

	sb.WriteString("\n\n")
	sb.WriteString(m.helpView())

	return sb.String()
}

// renderRadioButtons renders the 3-way radio button selector for transport type
func (m FormModel) renderRadioButtons() string {
	options := []string{"stdio", "http", "sse"}
	var parts []string

	for i, option := range options {
		var radioButton string
		if i == m.radioOption {
			radioButton = "(•)"
		} else {
			radioButton = "( )"
		}
		parts = append(parts, fmt.Sprintf("%s %s", radioButton, option))
	}

	return strings.Join(parts, "  ")
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
			key.NewBinding(key.WithKeys("space/←/→"), key.WithHelp("space/←/→", "select transport type")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "cycle transport type")),
		)
	case focusEnv, focusHeaders:
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
	m.headersInput.Reset()
	m.envInput.Reset()
	m.activeInput = focusName        // Reset focus to name
	m.transportType = TransportStdio // Reset to stdio
	m.radioOption = 0                // Reset to stdio option
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

	switch m.transportType {
	case TransportStdio:
		cmd := m.commandInput.Value()
		if cmd == "" {
			return server, fmt.Errorf("command is required for stdio server type")
		}
		server.Command = cmd
		server.Args = parseArgsString(m.argsInput.Value())
		server.Env = parseEnvString(m.envInput.Value())
		server.IsSSE = false

	case TransportHTTP:
		url := m.urlInput.Value()
		if url == "" {
			return server, fmt.Errorf("URL is required for HTTP server type")
		}
		server.URL = url
		server.Env = parseEnvString(m.headersInput.Value()) // Headers stored in Env field
		server.IsSSE = false

	case TransportSSE:
		url := m.urlInput.Value()
		if url == "" {
			return server, fmt.Errorf("URL is required for SSE server type")
		}
		server.URL = url
		server.Env = parseEnvString(m.headersInput.Value()) // Headers stored in Env field
		server.IsSSE = true
	}

	return server, nil
}

// LoadFromServer populates the form with data from an existing server
func (m *FormModel) LoadFromServer(server types.CommonServer) {
	m.nameInput.SetValue(server.Name)

	// Determine transport type and set radio option
	if server.Command != "" && server.URL == "" {
		// stdio type
		m.transportType = TransportStdio
		m.radioOption = 0
		m.commandInput.SetValue(server.Command)
		m.argsInput.SetValue(strings.Join(server.Args, " "))
		m.envInput.SetValue(formatEnvMap(server.Env))
	} else if server.URL != "" {
		// http or sse type
		if server.IsSSE {
			m.transportType = TransportSSE
			m.radioOption = 2
		} else {
			m.transportType = TransportHTTP
			m.radioOption = 1
		}
		m.urlInput.SetValue(server.URL)
		m.headersInput.SetValue(formatEnvMap(server.Env))
	}
}

// formatEnvMap converts a map to KEY=VALUE string format
func formatEnvMap(env map[string]string) string {
	if len(env) == 0 {
		return ""
	}
	var lines []string
	for key, value := range env {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(lines, "\n")
}
