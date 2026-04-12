package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPasswordModel_Update_Enter(t *testing.T) {
	m := newPasswordModel()
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	newModel, _ := m.Update(msg)

	pm := newModel.(passwordModel)
	if pm.err != nil {
		t.Errorf("expected no error on Enter, got %v", pm.err)
	}
}

func TestPasswordModel_Update_CancelEscape(t *testing.T) {
	m := newPasswordModel()
	msg := tea.KeyMsg{Type: tea.KeyEsc}

	newModel, _ := m.Update(msg)

	pm := newModel.(passwordModel)
	if pm.err == nil {
		t.Error("expected error on Escape")
	}
}

func TestPasswordModel_Update_CancelCtrlC(t *testing.T) {
	m := newPasswordModel()
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	newModel, _ := m.Update(msg)

	pm := newModel.(passwordModel)
	if pm.err == nil {
		t.Error("expected error on Ctrl+C")
	}
}

func TestPasswordModel_View(t *testing.T) {
	m := newPasswordModel()
	view := m.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestGetPassword_Cancelled(t *testing.T) {
	// This test verifies the error handling path when user cancels
	// In actual usage, this would require interactive input simulation
	// which is not possible in automated tests without additional mocking
	t.Skip("requires interactive TUI mocking")
}
