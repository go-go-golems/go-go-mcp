package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// FormKeyMap defines keybindings specific to the form view
type FormKeyMap struct {
	Submit    key.Binding
	Cancel    key.Binding
	Next      key.Binding
	Prev      key.Binding
	ToggleSSE key.Binding // New keybinding for SSE toggle
}

var defaultFormKeyMap = FormKeyMap{
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Next: key.NewBinding(
		key.WithKeys("tab", "down"),
		key.WithHelp("tab/↓", "next field"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab", "up"),
		key.WithHelp("shift+tab/↑", "prev field"),
	),
	ToggleSSE: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle stdio/sse"),
	),
}

// FormModel represents the form for adding/editing servers
type FormModel struct {
	keyMap       FormKeyMap
	nameInput    textinput.Model
	commandInput textinput.Model
	urlInput     textinput.Model
	argsInput    textinput.Model
	envInput     textinput.Model
	isSSE        bool // True if the server uses SSE (Cursor), false for command (Claude)
	isAddMode    bool // True if adding a new server, false if editing
	activeInput  int  // Index of the currently focused input
	submitted    bool // Flag indicating form submission was triggered
	cancelled    bool // Flag indicating form cancellation was triggered
	width        int
	height       int
}

// NewFormModel creates a new form model with initialized inputs
func NewFormModel() FormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Server name (required)"
	nameInput.Focus()
	nameInput.CharLimit = 50
	nameInput.Width = 40

	commandInput := textinput.New()
	commandInput.Placeholder = "Command path (if not SSE)"
	commandInput.CharLimit = 200
	commandInput.Width = 40

	urlInput := textinput.New()
	urlInput.Placeholder = "SSE URL (if SSE)"
	urlInput.CharLimit = 200
	urlInput.Width = 40

	argsInput := textinput.New()
	argsInput.Placeholder = "Arguments (space separated, if not SSE)"
	argsInput.CharLimit = 500
	argsInput.Width = 40

	envInput := textinput.New()
	envInput.Placeholder = "Environment variables (KEY=VALUE, one per line)"
	envInput.CharLimit = 1000
	envInput.Width = 40
	// TODO: Make envInput a textarea for better multi-line editing?

	return FormModel{
		keyMap:       defaultFormKeyMap,
		nameInput:    nameInput,
		commandInput: commandInput,
		urlInput:     urlInput,
		argsInput:    argsInput,
		envInput:     envInput,
		activeInput:  0,
		isSSE:        false, // Default, can be overridden when opening form
		isAddMode:    true,  // Default
	}
}

// Update handles form input messages
func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Cancel):
			m.cancelled = true
			return m, nil

		case key.Matches(msg, m.keyMap.Submit):
			// Trigger submission only if Enter is pressed on the *last* active input?
			// Or allow submit anytime?
			// For now, allow submit anytime Enter is pressed (might need refinement)
			m.submitted = true
			return m, nil

		case key.Matches(msg, m.keyMap.ToggleSSE):
			m.isSSE = !m.isSSE
			// Determine which input *should* be focused after toggle
			newFocusIndex := m.activeInput
			currentInputs := m.VisibleInputs()
			if newFocusIndex >= len(currentInputs) {
				newFocusIndex = len(currentInputs) - 1
			}
			m.activeInput = newFocusIndex // Set the potentially adjusted index

			// Reset focus to the correct field after toggling visibility
			cmd := m.FocusField(m.activeInput)
			return m, cmd

		case key.Matches(msg, m.keyMap.Next), key.Matches(msg, m.keyMap.Prev):
			// Handle navigation between inputs
			currentInputs := m.VisibleInputs()

			if key.Matches(msg, m.keyMap.Prev) {
				m.activeInput--
			} else {
				m.activeInput++
			}

			if m.activeInput >= len(currentInputs) {
				m.activeInput = 0
			} else if m.activeInput < 0 {
				m.activeInput = len(currentInputs) - 1
			}

			// Set focus on the active input
			for i, input := range currentInputs {
				if i == m.activeInput {
					cmds = append(cmds, input.Focus())
				} else {
					input.Blur()
				}
			}
			// Update blur/focus state for all inputs (visible or not)
			for _, input := range m.AllInputs() {
				isFocused := false
				for i, visibleInput := range currentInputs {
					if i == m.activeInput && input == visibleInput {
						isFocused = true
						break
					}
				}
				if !isFocused {
					input.Blur()
				}
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle text input for the focused field
	focusedInput := m.VisibleInputs()[m.activeInput]
	var cmd tea.Cmd
	*focusedInput, cmd = focusedInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the form
func (m FormModel) View() string {
	// TODO: Implement a proper form view using lipgloss
	// For now, just show the input values for debugging
	var sb strings.Builder

	title := "Add Server"
	if !m.isAddMode {
		title = "Edit Server"
	}
	sb.WriteString(title + "\n\n")

	sseStatus := "stdio (command/args)"
	if m.isSSE {
		sseStatus = "sse (url)"
	}
	sb.WriteString(fmt.Sprintf("Type: %s ([s] to toggle)\n\n", sseStatus))

	inputs := m.VisibleInputs()
	for i, input := range inputs {
		sb.WriteString(input.View())
		if i < len(inputs)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n\n")
	sb.WriteString(m.helpView())

	return sb.String()
}

// helpView renders the help text for the form
func (m FormModel) helpView() string {
	// You can generate this dynamically based on the keymap
	return lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf("%s | %s | %s | %s | %s",
			m.keyMap.Submit.Help().Key, m.keyMap.Submit.Help().Desc,
			m.keyMap.Cancel.Help().Key, m.keyMap.Cancel.Help().Desc,
			m.keyMap.Next.Help().Key, m.keyMap.Next.Help().Desc,
			m.keyMap.Prev.Help().Key, m.keyMap.Prev.Help().Desc,
			m.keyMap.ToggleSSE.Help().Key, m.keyMap.ToggleSSE.Help().Desc,
		),
	)
}

// VisibleInputs returns the text inputs currently relevant based on isSSE
func (m *FormModel) VisibleInputs() []*textinput.Model {
	inputs := []*textinput.Model{
		&m.nameInput,
	}
	if m.isSSE {
		inputs = append(inputs, &m.urlInput)
	} else {
		inputs = append(inputs, &m.commandInput, &m.argsInput)
	}
	inputs = append(inputs, &m.envInput)
	return inputs
}

// AllInputs returns all text input fields, regardless of visibility
func (m *FormModel) AllInputs() []*textinput.Model {
	return []*textinput.Model{
		&m.nameInput,
		&m.commandInput,
		&m.urlInput,
		&m.argsInput,
		&m.envInput,
	}
}

// Reset clears all form inputs and resets state
func (m *FormModel) Reset() {
	m.nameInput.Reset()
	m.commandInput.Reset()
	m.urlInput.Reset()
	m.argsInput.Reset()
	m.envInput.Reset()
	m.activeInput = 0
	m.isAddMode = true
	m.isSSE = false
	m.submitted = false
	m.cancelled = false
	m.nameInput.Focus()
	m.commandInput.Blur()
	m.urlInput.Blur()
	m.argsInput.Blur()
	m.envInput.Blur()
}

// FocusField sets focus on the input at the given index among visible inputs.
func (m *FormModel) FocusField(index int) tea.Cmd {
	var cmds []tea.Cmd
	currentInputs := m.VisibleInputs()

	if index < 0 || index >= len(currentInputs) {
		// Should not happen, but handle gracefully
		index = 0
	}

	m.activeInput = index

	// Set focus on the target input, blur others
	for i, input := range m.AllInputs() { // Iterate through ALL inputs
		isVisibleAndFocused := false
		for _, visibleInput := range currentInputs { // Use _ to ignore the index j
			if i == m.activeInput && input == visibleInput { // Check if it's the *visible* focused one
				isVisibleAndFocused = true
				break
			}
		}

		if isVisibleAndFocused {
			cmds = append(cmds, input.Focus())
		} else {
			input.Blur()
		}
	}
	return tea.Batch(cmds...)
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
