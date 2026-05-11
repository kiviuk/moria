package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPromptModel_Update_Enter(t *testing.T) {
	// Verify Enter key confirms without error
	m := newPromptModel(passwordPromptCfg)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pm := newModel.(promptModel)
	if pm.err != nil {
		t.Errorf("expected no error on Enter, got %v", pm.err)
	}
}

func TestPromptModel_Update_Cancel(t *testing.T) {
	// Verify Escape and Ctrl+C both set an error
	for _, key := range []tea.KeyType{tea.KeyEsc, tea.KeyCtrlC} {
		m := newPromptModel(passwordPromptCfg)
		newModel, _ := m.Update(tea.KeyMsg{Type: key})
		pm := newModel.(promptModel)
		if pm.err == nil {
			t.Errorf("key %v: expected cancel error, got nil", key)
		}
	}
}

func TestPromptModel_View_NonEmpty(t *testing.T) {
	// Verify View returns a non-empty string
	if v := newPromptModel(passwordPromptCfg).View(); v == "" {
		t.Error("expected non-empty view")
	}
}
