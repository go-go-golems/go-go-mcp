package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// Define key bindings
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Back      key.Binding
	Quit      key.Binding
	Add       key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Duplicate key.Binding
	Yank      key.Binding
	Paste     key.Binding
	Enable    key.Binding
	Help      key.Binding
}

// Set up default key bindings
var defaultKeyMap = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/confirm"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back/cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit application"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add new server"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit selected server"),
	),
	Delete: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "delete selected server"),
	),
	Duplicate: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "duplicate selected server"),
	),
	Yank: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yank (copy) selected server"),
	),
	Paste: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "paste yanked server"),
	),
	Enable: key.NewBinding(
		key.WithKeys("space", " "),
		key.WithHelp("space", "toggle server enabled/disabled"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}
