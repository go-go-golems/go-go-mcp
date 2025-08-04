package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var content string

	switch m.mode {
	case modeMenu:
		content = m.menuList.View()
	case modeSubmenu:
		content = m.submenuList.View()
	case modeList:
		if m.activeList != nil {
			content = m.activeList.View()
		} else {
			content = "Loading..."
		}
	case modeAddEdit:
		if m.configType == ConfigTypeProfile {
			content = m.profileFormState.View()
		} else {
			content = m.formState.View()
		}
	case modeConfirm:
		content = m.confirmDialog.View()
	default:
		content = "Unknown mode"
	}

	// Add breadcrumb if available
	var header string
	if m.breadcrumb != "" {
		breadcrumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		header = breadcrumbStyle.Render(m.breadcrumb) + "\n"
	}

	// Add error message if there is one
	var errorDisplay string
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		errorDisplay = errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg)) + "\n"
	}

	// Help view
	helpView := m.help.View(m)

	return header + errorDisplay + content + "\n" + helpView
}
