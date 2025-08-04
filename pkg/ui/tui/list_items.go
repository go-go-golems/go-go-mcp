package tui

import "fmt"

// Server item model for the list view
type serverItem struct {
	name    string
	command string
	args    []string
	env     map[string]string
	url     string // for Cursor SSE servers
	enabled bool
	isSSE   bool // Added to distinguish server type
}

func (i serverItem) Title() string { return i.name }
func (i serverItem) Description() string {
	status := "enabled"
	if !i.enabled {
		status = "disabled"
	}

	serverType := "CMD"
	if i.isSSE {
		serverType = "SSE"
	}

	if i.url != "" {
		return fmt.Sprintf("%s: %s (%s)", serverType, i.url, status)
	}
	return fmt.Sprintf("%s: %s (%s)", serverType, i.command, status)
}
func (i serverItem) FilterValue() string { return i.name }

// Simple list item for menu
type listItem struct {
	title       string
	description string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }
