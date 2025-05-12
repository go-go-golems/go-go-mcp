package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProfileFormKeyMap defines keybindings specific to the profile form view
type ProfileFormKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
	Next   key.Binding // Tab
	Prev   key.Binding // Shift+Tab
}

var defaultProfileFormKeyMap = ProfileFormKeyMap{
	Submit: key.NewBinding(
		key.WithKeys("alt+s"),
		key.WithHelp("alt+s", "submit"),
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
}

// Focus indices for the profile form fields
const (
	focusProfileName focusProfileField = iota
	focusProfileDescription
	focusToolDirs
	focusToolFiles
	focusPromptDirs
	focusPromptFiles
	focusProfileMax // Keep track of the total number of potential focus points
)

type focusProfileField int

// ProfileFormModel represents the form for adding/editing profiles
type ProfileFormModel struct {
	keyMap           ProfileFormKeyMap
	nameInput        textinput.Model
	descriptionInput textinput.Model
	toolDirsInput    textinput.Model
	toolFilesInput   textinput.Model
	promptDirsInput  textinput.Model
	promptFilesInput textinput.Model
	activeInput      focusProfileField
	isAddMode        bool // True if adding a new profile, false if editing
	submitted        bool // Flag indicating form submission was triggered
	cancelled        bool // Flag indicating form cancellation was triggered
}

// NewProfileFormModel creates a new profile form model with initialized inputs
func NewProfileFormModel() ProfileFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Profile name (required)"
	nameInput.Focus() // Start focus on name
	nameInput.CharLimit = 50
	nameInput.Width = 80

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "Profile description"
	descriptionInput.CharLimit = 200
	descriptionInput.Width = 80

	toolDirsInput := textinput.New()
	toolDirsInput.Placeholder = "Tool directories (comma separated paths)"
	toolDirsInput.CharLimit = 500
	toolDirsInput.Width = 80

	toolFilesInput := textinput.New()
	toolFilesInput.Placeholder = "Tool files (comma separated paths)"
	toolFilesInput.CharLimit = 500
	toolFilesInput.Width = 80

	promptDirsInput := textinput.New()
	promptDirsInput.Placeholder = "Prompt directories (comma separated paths)"
	promptDirsInput.CharLimit = 500
	promptDirsInput.Width = 80

	promptFilesInput := textinput.New()
	promptFilesInput.Placeholder = "Prompt files (comma separated paths)"
	promptFilesInput.CharLimit = 500
	promptFilesInput.Width = 80

	return ProfileFormModel{
		keyMap:           defaultProfileFormKeyMap,
		nameInput:        nameInput,
		descriptionInput: descriptionInput,
		toolDirsInput:    toolDirsInput,
		toolFilesInput:   toolFilesInput,
		promptDirsInput:  promptDirsInput,
		promptFilesInput: promptFilesInput,
		activeInput:      focusProfileName,
		isAddMode:        true, // Default to add mode
	}
}

// Update handles profile form input messages
func (m ProfileFormModel) Update(msg tea.Msg) (ProfileFormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handling - only handle special keys like ESC, Alt+s
		// Let other keys (like 'q') pass through to the active input
		switch {
		case key.Matches(msg, m.keyMap.Cancel):
			m.cancelled = true
			// Ensure all inputs are blurred on cancel
			m.nameInput.Blur()
			m.descriptionInput.Blur()
			m.toolDirsInput.Blur()
			m.toolFilesInput.Blur()
			m.promptDirsInput.Blur()
			m.promptFilesInput.Blur()
			return m, nil

		case key.Matches(msg, m.keyMap.Submit):
			// For submit, mark as submitted and return
			m.submitted = true
			return m, nil

		case key.Matches(msg, m.keyMap.Next), key.Matches(msg, m.keyMap.Prev):
			direction := 1
			if key.Matches(msg, m.keyMap.Prev) {
				direction = -1
			}

			// Cycle through focusable elements
			m.activeInput = (m.activeInput + focusProfileField(direction) + focusProfileMax) % focusProfileMax

			cmds = append(cmds, m.updateFocus())
			return m, tea.Batch(cmds...)
		}

		// Process the event in the active text input
		var cmd tea.Cmd
		switch m.activeInput {
		case focusProfileName:
			m.nameInput, cmd = m.nameInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case focusProfileDescription:
			m.descriptionInput, cmd = m.descriptionInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case focusToolDirs:
			m.toolDirsInput, cmd = m.toolDirsInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case focusToolFiles:
			m.toolFilesInput, cmd = m.toolFilesInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case focusPromptDirs:
			m.promptDirsInput, cmd = m.promptDirsInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case focusPromptFiles:
			m.promptFilesInput, cmd = m.promptFilesInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case focusProfileMax:
			// This should never happen as focusProfileMax is just a sentinel value
			// Do nothing in this case
		}
	}

	return m, tea.Batch(cmds...)
}

// updateFocus manages blurring/focusing the correct text input
func (m *ProfileFormModel) updateFocus() tea.Cmd {
	// Blur all inputs first
	m.nameInput.Blur()
	m.descriptionInput.Blur()
	m.toolDirsInput.Blur()
	m.toolFilesInput.Blur()
	m.promptDirsInput.Blur()
	m.promptFilesInput.Blur()

	// Initialize command slice
	var cmds []tea.Cmd

	// Focus the correct input based on activeInput
	switch m.activeInput {
	case focusProfileName:
		cmds = append(cmds, m.nameInput.Focus())
	case focusProfileDescription:
		cmds = append(cmds, m.descriptionInput.Focus())
	case focusToolDirs:
		cmds = append(cmds, m.toolDirsInput.Focus())
	case focusToolFiles:
		cmds = append(cmds, m.toolFilesInput.Focus())
	case focusPromptDirs:
		cmds = append(cmds, m.promptDirsInput.Focus())
	case focusPromptFiles:
		cmds = append(cmds, m.promptFilesInput.Focus())
	case focusProfileMax:
		// This should never happen as focusProfileMax is just a sentinel value
		// Do nothing in this case
	}

	return tea.Batch(cmds...)
}

// View renders the profile form
func (m ProfileFormModel) View() string {
	var sb strings.Builder

	title := "Add Profile"
	if !m.isAddMode {
		title = "Edit Profile"
	}
	sb.WriteString(title + "\n\n")

	// Name Input
	sb.WriteString(m.nameInput.View() + "\n")

	// Description Input
	sb.WriteString(m.descriptionInput.View() + "\n\n")

	// Tool configuration section
	labelStyle := lipgloss.NewStyle().Bold(true)
	sb.WriteString(labelStyle.Render("Tools Configuration") + "\n")

	// Tool directories
	sb.WriteString(m.toolDirsInput.View() + "\n")

	// Tool files
	sb.WriteString(m.toolFilesInput.View() + "\n\n")

	// Prompt configuration section
	sb.WriteString(labelStyle.Render("Prompts Configuration") + "\n")

	// Prompt directories
	sb.WriteString(m.promptDirsInput.View() + "\n")

	// Prompt files
	sb.WriteString(m.promptFilesInput.View() + "\n")

	sb.WriteString("\n")
	sb.WriteString(m.helpView())

	return sb.String()
}

// helpView renders the help text for the profile form
func (m ProfileFormModel) helpView() string {
	// Build help keys list
	keys := []key.Binding{
		m.keyMap.Submit,
		m.keyMap.Cancel,
		m.keyMap.Next,
		m.keyMap.Prev,
	}

	helpParts := make([]string, len(keys))
	for i, k := range keys {
		helpParts[i] = fmt.Sprintf("%s %s", k.Help().Key, k.Help().Desc)
	}

	return lipgloss.NewStyle().Faint(true).Render(strings.Join(helpParts, " | "))
}

// SetProfileData populates the form with existing profile data for editing
func (m *ProfileFormModel) SetProfileData(name, description string) {
	m.nameInput.SetValue(name)
	m.descriptionInput.SetValue(description)
	m.isAddMode = false // Set to edit mode

	// Note: In a real implementation, we would also set values for tool directories,
	// tool files, prompt directories, and prompt files from the existing profile.
	// This would require extracting this data from the profile structure.
}

// GetProfileData retrieves the profile data from the form
func (m *ProfileFormModel) GetProfileData() (string, string, error) {
	name := m.nameInput.Value()
	if name == "" {
		return "", "", fmt.Errorf("profile name is required")
	}

	description := m.descriptionInput.Value()

	return name, description, nil
}

// ParsePathList parses a comma-separated list of paths into a slice of strings
func (m *ProfileFormModel) ParsePathList(input string) []string {
	if input == "" {
		return []string{}
	}

	paths := strings.Split(input, ",")
	result := make([]string, 0, len(paths))

	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// GetToolsAndPrompts retrieves tool and prompt configuration from the form
func (m *ProfileFormModel) GetToolsAndPrompts() ([]string, []string, []string, []string) {
	toolDirs := m.ParsePathList(m.toolDirsInput.Value())
	toolFiles := m.ParsePathList(m.toolFilesInput.Value())
	promptDirs := m.ParsePathList(m.promptDirsInput.Value())
	promptFiles := m.ParsePathList(m.promptFilesInput.Value())
	return toolDirs, toolFiles, promptDirs, promptFiles
}
