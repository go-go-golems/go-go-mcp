package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmKeyMap struct {
	Left    key.Binding
	Right   key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

var defaultConfirmKeyMap = confirmKeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

// ConfirmModel represents a confirmation dialog
type ConfirmModel struct {
	title    string
	message  string
	selected bool // true = yes, false = no
	keyMap   confirmKeyMap
}

// NewConfirmModel creates a new confirmation dialog
func NewConfirmModel(title, message string) ConfirmModel {
	return ConfirmModel{
		title:    title,
		message:  message,
		selected: false,
		keyMap:   defaultConfirmKeyMap,
	}
}

// ConfirmMsg is returned when a selection is confirmed
type ConfirmMsg struct {
	Confirmed bool
}

// Update handles confirmation dialog input
func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Left):
			m.selected = false
			return m, nil

		case key.Matches(msg, m.keyMap.Right):
			m.selected = true
			return m, nil

		case key.Matches(msg, m.keyMap.Confirm):
			return m, func() tea.Msg {
				return ConfirmMsg{Confirmed: m.selected}
			}

		case key.Matches(msg, m.keyMap.Cancel):
			return m, func() tea.Msg {
				return ConfirmMsg{Confirmed: false}
			}
		}
	}
	return m, nil
}

// View renders the confirmation dialog
func (m ConfirmModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)
	messageStyle := lipgloss.NewStyle().MarginBottom(1)

	noStyle := lipgloss.NewStyle().
		Padding(0, 3).
		Background(lipgloss.Color("#888888"))

	yesStyle := lipgloss.NewStyle().
		Padding(0, 3).
		Background(lipgloss.Color("#888888"))

	if m.selected {
		yesStyle = yesStyle.
			Background(lipgloss.Color("#008800")).
			Foreground(lipgloss.Color("#ffffff"))
	} else {
		noStyle = noStyle.
			Background(lipgloss.Color("#880000")).
			Foreground(lipgloss.Color("#ffffff"))
	}

	no := noStyle.Render("No")
	yes := yesStyle.Render("Yes")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, no, "  ", yes)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999")).
		MarginTop(1)

	helpText := helpStyle.Render("←/h: Select No • →/l: Select Yes • Enter: Confirm • Esc: Cancel")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(m.title),
		messageStyle.Render(m.message),
		buttons,
		helpText,
	)
}
