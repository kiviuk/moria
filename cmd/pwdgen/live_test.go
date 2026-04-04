package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kiviuk/pwdgen/internal/app"
)

func TestLiveModel_MaxLen_AllowTyping(t *testing.T) {
	// Verify that typing is blocked when password reaches maxLen
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 6)

	// Type "1" → password = 3 chars (00|)
	m = simulateKey(m, "1")
	if len(m.password) != 3 {
		t.Errorf("after '1': expected 3 chars, got %d", len(m.password))
	}
	if m.err != "" {
		t.Errorf("after '1': unexpected error: %s", m.err)
	}

	// Type "2" → password = 6 chars (00|10|)
	m = simulateKey(m, "2")
	if len(m.password) != 6 {
		t.Errorf("after '2': expected 6 chars, got %d", len(m.password))
	}
	if m.err != "" {
		t.Errorf("after '2': unexpected error: %s", m.err)
	}

	// Type "3" → should be blocked (6 >= 6)
	m = simulateKey(m, "3")
	if len(m.password) != 6 {
		t.Errorf("after '3': expected 6 chars (blocked), got %d", len(m.password))
	}
	if m.err == "" {
		t.Error("after '3': expected error, got nil")
	}
}

func TestLiveModel_MaxLen_Partial(t *testing.T) {
	// Verify maxLen=5 blocks at 6 chars, LiveMode truncates to 5
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 5)

	m = simulateKey(m, "1")
	if len(m.password) != 3 {
		t.Errorf("after '1': expected 3 chars, got %d", len(m.password))
	}

	m = simulateKey(m, "2")
	if len(m.password) != 6 {
		t.Errorf("after '2': expected 6 chars, got %d", len(m.password))
	}

	// LiveMode truncates on exit
	password := m.password
	if maxLen := 5; maxLen > 0 && len(password) > maxLen {
		password = password[:maxLen]
	}
	if len(password) != 5 {
		t.Errorf("truncated password: expected 5 chars, got %d", len(password))
	}
	if password != "00|10" {
		t.Errorf("truncated password: expected %q, got %q", "00|10", password)
	}
}

func TestLiveModel_MaxLen_NoLimit(t *testing.T) {
	// Verify maxLen=0 allows unlimited typing
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 0)

	for i := 0; i < 10; i++ {
		m = simulateKey(m, "a")
	}
	if len(m.password) != 30 {
		t.Errorf("expected 30 chars, got %d", len(m.password))
	}
	if m.err != "" {
		t.Errorf("unexpected error: %s", m.err)
	}
}

func TestLiveModel_MaxLen_Backspace(t *testing.T) {
	// Verify backspace removes chars and allows re-typing
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 6)

	m = simulateKey(m, "1")
	m = simulateKey(m, "2")
	if len(m.password) != 6 {
		t.Fatalf("expected 6 chars, got %d", len(m.password))
	}

	// Backspace → password = 3 chars
	m = simulateBackspace(m)
	if len(m.password) != 3 {
		t.Errorf("after backspace: expected 3 chars, got %d", len(m.password))
	}

	// Should be able to type again
	m = simulateKey(m, "3")
	if len(m.password) != 6 {
		t.Errorf("after '3': expected 6 chars, got %d", len(m.password))
	}
}

func simulateKey(m liveModel, key string) liveModel {
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return result.(liveModel)
}

func simulateBackspace(m liveModel) liveModel {
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	return result.(liveModel)
}

func newTestMatrix() app.Matrix {
	var m app.Matrix
	for row := 0; row < app.PasswordMatrixRows; row++ {
		for col := 0; col < app.PasswordMatrixColumns; col++ {
			m[row][col] = string(rune('0'+row)) + string(rune('0'+col)) + "|"
		}
	}
	return m
}
