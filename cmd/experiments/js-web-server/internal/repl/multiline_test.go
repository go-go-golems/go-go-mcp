package repl

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestMultilinePasteSupport tests the multiline paste functionality
func TestMultilinePasteSupport(t *testing.T) {
	model := NewModel(false)

	// Simulate pasting multiline JavaScript code
	multilineCode := `app.get('/users/:id', (req, res) => {
    const userId = req.params.id;
    const users = db.query('SELECT * FROM users WHERE id = ?', [userId]);
    res.json({user: users[0], path: req.path});
});`

	// Create a KeyRunes message simulating paste
	pasteMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(multilineCode),
	}

	// Update the model with the paste
	updatedModel, _ := model.Update(pasteMsg)
	m := updatedModel.(Model)

	// Verify that multiline mode was activated
	if !m.multilineMode {
		t.Error("Expected multiline mode to be activated after pasting multiline content")
	}

	// Verify that the multiline text contains the pasted lines
	lines := strings.Split(multilineCode, "\n")
	expectedLines := len(lines) - 1 // Last line goes to current input

	if len(m.multilineText) != expectedLines {
		t.Errorf("Expected %d multiline entries, got %d", expectedLines, len(m.multilineText))
	}

	// Verify the last line is in the text input
	lastLine := lines[len(lines)-1]
	if m.textInput.Value() != lastLine {
		t.Errorf("Expected current input to be '%s', got '%s'", lastLine, m.textInput.Value())
	}

	// Verify first line contains the app.get call
	if !strings.Contains(m.multilineText[0], "app.get") {
		t.Errorf("Expected first line to contain 'app.get', got '%s'", m.multilineText[0])
	}
}

// TestCtrlCBehavior tests that Ctrl+C no longer quits but clears/cancels instead
func TestCtrlCBehavior(t *testing.T) {
	model := NewModel(false)

	// Test Ctrl+C in normal mode (should clear input)
	model.textInput.SetValue("some input")
	ctrlCMsg := tea.KeyMsg{Type: tea.KeyCtrlC}

	updatedModel, cmd := model.Update(ctrlCMsg)
	m := updatedModel.(Model)

	// Should not quit (cmd should be nil since we don't quit)
	if cmd != nil {
		// cmd is a function, we need to execute it to see what it returns
		if result := cmd(); result != nil {
			if _, isQuit := result.(tea.QuitMsg); isQuit {
				t.Error("Ctrl+C should not quit in normal mode")
			}
		}
	}

	// Should clear input
	if m.textInput.Value() != "" {
		t.Errorf("Expected input to be cleared, got '%s'", m.textInput.Value())
	}

	// Test Ctrl+C in multiline mode (should cancel multiline mode)
	m.multilineMode = true
	m.multilineText = []string{"line1", "line2"}
	m.textInput.SetValue("line3")

	updatedModel, cmd = m.Update(ctrlCMsg)
	m2 := updatedModel.(Model)

	// Should not quit (cmd should be nil since we don't quit)
	if cmd != nil {
		// cmd is a function, we need to execute it to see what it returns
		if result := cmd(); result != nil {
			if _, isQuit := result.(tea.QuitMsg); isQuit {
				t.Error("Ctrl+C should not quit in multiline mode")
			}
		}
	}

	// Should exit multiline mode
	if m2.multilineMode {
		t.Error("Expected multiline mode to be disabled after Ctrl+C")
	}

	// Should clear multiline text
	if len(m2.multilineText) != 0 {
		t.Errorf("Expected multiline text to be cleared, got %d lines", len(m2.multilineText))
	}

	// Should clear current input
	if m2.textInput.Value() != "" {
		t.Errorf("Expected input to be cleared, got '%s'", m2.textInput.Value())
	}
}

// TestCtrlAltEEditor tests the new Ctrl+Alt+E key combination for external editor
func TestCtrlAltEEditor(t *testing.T) {
	model := NewModel(false)
	model.textInput.SetValue("console.log('test')")

	// Test Ctrl+E without Alt (should not trigger editor)
	ctrlEMsg := tea.KeyMsg{Type: tea.KeyCtrlE, Alt: false}

	updatedModel, _ := model.Update(ctrlEMsg)
	m := updatedModel.(Model)

	// Input should remain unchanged since editor wasn't triggered
	if m.textInput.Value() != "console.log('test')" {
		t.Error("Ctrl+E without Alt should not trigger external editor")
	}

	// Test Ctrl+Alt+E (should attempt to trigger editor, but will fail in test environment)
	ctrlAltEMsg := tea.KeyMsg{Type: tea.KeyCtrlE, Alt: true}

	updatedModel, _ = model.Update(ctrlAltEMsg)
	m2 := updatedModel.(Model)

	// Since external editor will fail in test environment, we just verify the key combination is handled
	// The actual editor functionality is tested in the existing integration tests
	_ = m2 // Just verify no panic occurred
}

// TestSingleLinePaste tests that single-line paste still works normally
func TestSingleLinePaste(t *testing.T) {
	model := NewModel(false)

	// Simulate pasting single-line code
	singleLineCode := "console.log('Hello World')"

	pasteMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(singleLineCode),
	}

	// Update the model with the paste
	updatedModel, _ := model.Update(pasteMsg)
	m := updatedModel.(Model)

	// Should NOT activate multiline mode for single-line paste
	if m.multilineMode {
		t.Error("Expected multiline mode to remain disabled for single-line paste")
	}

	// The text should be handled by the normal textinput component
	// (We can't easily test this without mocking the textinput, but we verify multiline mode wasn't triggered)
}

// TestMultilineWithExistingContent tests pasting into existing multiline content
func TestMultilineWithExistingContent(t *testing.T) {
	model := NewModel(false)

	// Start with some existing multiline content
	model.multilineMode = true
	model.multilineText = []string{"// Existing code", "let x = 1;"}
	model.textInput.SetValue("// More code")

	// Paste additional multiline content
	additionalCode := `if (x > 0) {
    console.log('positive');
}`

	pasteMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(additionalCode),
	}

	updatedModel, _ := model.Update(pasteMsg)
	m := updatedModel.(Model)

	// Should still be in multiline mode
	if !m.multilineMode {
		t.Error("Expected to remain in multiline mode")
	}

	// Should have combined the existing content with the pasted content
	// Original multilineText: ["// Existing code", "let x = 1;"] + current input "// More code"
	// Paste: "if (x > 0) {\n    console.log('positive');\n}"
	// The paste logic works as follows:
	// - Since we're already in multiline mode, the current input "// More code" is NOT automatically added to multilineText
	// - The first line of paste "if (x > 0) {" gets appended to the LAST line of multilineText (index 1)
	// - The remaining lines of paste (except last) get added as new lines
	// - The last line of paste "}" becomes the new current input
	// Result: ["// Existing code", "let x = 1;if (x > 0) {", "    console.log('positive');"]
	expectedLines := 3 // 2 existing + 1 from paste middle lines
	if len(m.multilineText) != expectedLines {
		t.Errorf("Expected %d lines in multiline text, got %d", expectedLines, len(m.multilineText))
		for i, line := range m.multilineText {
			t.Logf("Line %d: %s", i, line)
		}
	}

	// The second line (index 1) should contain the existing line + first line of paste
	if len(m.multilineText) >= 2 {
		expectedSecondLine := "let x = 1;if (x > 0) {"
		if m.multilineText[1] != expectedSecondLine {
			t.Errorf("Expected second line to be '%s', got '%s'", expectedSecondLine, m.multilineText[1])
		}
	}

	// The current input should be the last line of the paste
	if !strings.Contains(m.textInput.Value(), "}") {
		t.Errorf("Expected current input to be the last line of paste, got '%s'", m.textInput.Value())
	}
}
