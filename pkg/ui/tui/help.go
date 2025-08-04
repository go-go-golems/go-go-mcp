package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// contextualHelp returns help keys based on the current mode
func (m Model) contextualHelp() []key.Binding {
	switch m.mode {
	case modeMenu:
		return []key.Binding{
			m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Help, m.keys.Quit,
		}
	case modeSubmenu:
		return []key.Binding{
			m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back, m.keys.Help, m.keys.Quit,
		}
	case modeList:
		if m.configType == ConfigTypeProfile {
			return []key.Binding{
				m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back,
				m.keys.Add, m.keys.Edit, m.keys.Delete, m.keys.Duplicate,
				m.keys.Enable, m.keys.Help, m.keys.Quit,
			}
		} else {
			return []key.Binding{
				m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back,
				m.keys.Add, m.keys.Edit, m.keys.Delete, m.keys.Duplicate,
				m.keys.Yank, m.keys.Paste, m.keys.Enable,
				m.keys.Help, m.keys.Quit,
			}
		}
	case modeAddEdit:
		if m.configType == ConfigTypeProfile {
			// Return profile form help
			return []key.Binding{
				m.profileFormState.keyMap.Submit,
				m.profileFormState.keyMap.Cancel,
				m.profileFormState.keyMap.Next,
				m.profileFormState.keyMap.Prev,
			}
		} else {
			// Return server form help
			return []key.Binding{
				m.formState.keyMap.Submit,
				m.formState.keyMap.Cancel,
				m.formState.keyMap.Next,
				m.formState.keyMap.Prev,
			}
		}
	case modeConfirm:
		return []key.Binding{
			m.confirmDialog.keyMap.Confirm,
			m.confirmDialog.keyMap.Cancel,
		}
	default:
		return []key.Binding{m.keys.Quit}
	}
}

// ShortHelp implements help.KeyMap for compatibility
func (m Model) ShortHelp() []key.Binding {
	return m.contextualHelp()
}

// FullHelp implements help.KeyMap for compatibility
func (m Model) FullHelp() [][]key.Binding {
	// Return help based on current mode
	switch m.mode {
	case modeMenu:
		return [][]key.Binding{
			{m.keys.Up, m.keys.Down, m.keys.Enter},
			{m.keys.Help, m.keys.Quit},
		}
	case modeSubmenu:
		return [][]key.Binding{
			{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back},
			{m.keys.Help, m.keys.Quit},
		}
	case modeList:
		if m.configType == ConfigTypeProfile {
			return [][]key.Binding{
				{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back},
				{m.keys.Add, m.keys.Edit, m.keys.Delete, m.keys.Duplicate},
				{m.keys.Enable, m.keys.Help, m.keys.Quit},
			}
		}
		return [][]key.Binding{
			{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Back},
			{m.keys.Add, m.keys.Edit, m.keys.Delete, m.keys.Duplicate},
			{m.keys.Yank, m.keys.Paste, m.keys.Enable, m.keys.Help, m.keys.Quit},
		}
	case modeAddEdit:
		if m.configType == ConfigTypeProfile {
			return [][]key.Binding{
				{m.profileFormState.keyMap.Submit, m.profileFormState.keyMap.Cancel},
				{m.profileFormState.keyMap.Next, m.profileFormState.keyMap.Prev},
			}
		}
		return [][]key.Binding{
			{m.formState.keyMap.Submit, m.formState.keyMap.Cancel},
			{m.formState.keyMap.Next, m.formState.keyMap.Prev},
		}
	case modeConfirm:
		return [][]key.Binding{
			{m.confirmDialog.keyMap.Confirm, m.confirmDialog.keyMap.Cancel},
		}
	default:
		return [][]key.Binding{{m.keys.Quit}}
	}
}
