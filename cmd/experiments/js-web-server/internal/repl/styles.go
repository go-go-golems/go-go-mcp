package repl

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles defines the visual styling for the REPL
type Styles struct {
	Title    lipgloss.Style
	Prompt   lipgloss.Style
	Result   lipgloss.Style
	Error    lipgloss.Style
	Info     lipgloss.Style
	HelpText lipgloss.Style
}

// DefaultStyles returns the default styling configuration
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("32")).
			Background(lipgloss.Color("240")).
			Padding(0, 1),

		Prompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true),

		Result: lipgloss.NewStyle().
			Foreground(lipgloss.Color("36")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true),

		HelpText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),
	}
}
