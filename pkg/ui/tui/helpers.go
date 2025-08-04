package tui

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
)

// loadServerToForm loads the data for the selected server into the form model.
func (m *Model) loadServerToForm(serverName string) (FormModel, error) {
	if m.currentEditor == nil {
		return FormModel{}, fmt.Errorf("no configuration editor loaded")
	}

	server, found, err := m.currentEditor.GetServer(serverName)
	if err != nil {
		return FormModel{}, fmt.Errorf("error getting server '%s': %w", serverName, err)
	}
	if !found {
		return FormModel{}, fmt.Errorf("server '%s' not found", serverName)
	}

	// Create and populate a new form model
	form := NewFormModel()
	form.LoadFromServer(server)
	form.isAddMode = false // Explicitly set to edit mode

	return form, nil
}

// loadCommonServerToForm loads data from a CommonServer directly into the form model.
func (m *Model) loadCommonServerToForm(server types.CommonServer) (FormModel, error) {
	// Create and populate a new form model
	form := NewFormModel()
	form.LoadFromServer(server)
	form.isAddMode = false // Explicitly set to edit mode

	return form, nil
}
