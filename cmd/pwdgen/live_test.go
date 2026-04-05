package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kiviuk/pwdgen/internal/app"
	"github.com/kiviuk/pwdgen/internal/testutil"
)

func newTestMatrix() app.Matrix {
	m, err := app.NewMatrix(testutil.NewTestMatrixData(app.PasswordMatrixRows, app.PasswordMatrixColumns, app.CharactersPerMatrixCell))
	if err != nil {
		panic(err)
	}
	return m
}

func TestLiveModel_MaxLen_AllowTyping(t *testing.T) {
	// Verify that typing is blocked when password reaches maxLen
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 2*app.CharactersPerMatrixCell)

	// Type "1" → password = CharactersPerMatrixCell chars
	m = simulateKey(m, "1")
	if len(m.password) != app.CharactersPerMatrixCell {
		t.Errorf("after '1': expected %d chars, got %d", app.CharactersPerMatrixCell, len(m.password))
	}
	if m.err != "" {
		t.Errorf("after '1': unexpected error: %s", m.err)
	}

	// Type "2" → password = 2*CharactersPerMatrixCell chars
	m = simulateKey(m, "2")
	if len(m.password) != 2*app.CharactersPerMatrixCell {
		t.Errorf("after '2': expected %d chars, got %d", 2*app.CharactersPerMatrixCell, len(m.password))
	}
	if m.err != "" {
		t.Errorf("after '2': unexpected error: %s", m.err)
	}

	// Type "3" → should be blocked (at limit)
	m = simulateKey(m, "3")
	if len(m.password) != 2*app.CharactersPerMatrixCell {
		t.Errorf("after '3': expected %d chars (blocked), got %d", 2*app.CharactersPerMatrixCell, len(m.password))
	}
	if m.err == "" {
		t.Error("after '3': expected error, got nil")
	}
}

func TestLiveModel_MaxLen_Partial(t *testing.T) {
	// Verify maxLen truncates password on exit, never exceeding maxLen
	matrix := newTestMatrix()
	maxLen := 5
	m := newLiveModel(matrix, maxLen)

	m = simulateKey(m, "1")
	if len(m.password) != app.CharactersPerMatrixCell {
		t.Errorf("after '1': expected %d chars, got %d", app.CharactersPerMatrixCell, len(m.password))
	}

	m = simulateKey(m, "2")
	if len(m.password) != 2*app.CharactersPerMatrixCell {
		t.Errorf("after '2': expected %d chars, got %d", 2*app.CharactersPerMatrixCell, len(m.password))
	}

	// LiveMode truncates on exit - result is min(password length, maxLen)
	password := m.password
	if maxLen > 0 && len(password) > maxLen {
		password = password[:maxLen]
	}
	expectedLen := len(m.password)
	if expectedLen > maxLen {
		expectedLen = maxLen
	}
	if len(password) != expectedLen {
		t.Errorf("truncated password: expected %d chars, got %d", expectedLen, len(password))
	}
	if len(password) > maxLen {
		t.Errorf("password exceeds maxLen: %d > %d", len(password), maxLen)
	}
}

func TestLiveModel_MaxLen_NoLimit(t *testing.T) {
	// Verify maxLen=0 allows unlimited typing
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 0)

	for i := 0; i < 10; i++ {
		m = simulateKey(m, "a")
	}
	if len(m.password) != 10*app.CharactersPerMatrixCell {
		t.Errorf("expected %d chars, got %d", 10*app.CharactersPerMatrixCell, len(m.password))
	}
	if m.err != "" {
		t.Errorf("unexpected error: %s", m.err)
	}
}

func TestLiveModel_MaxLen_Backspace(t *testing.T) {
	// Verify backspace removes chars and allows re-typing
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 2*app.CharactersPerMatrixCell)

	m = simulateKey(m, "1")
	m = simulateKey(m, "2")
	if len(m.password) != 2*app.CharactersPerMatrixCell {
		t.Fatalf("expected %d chars, got %d", 2*app.CharactersPerMatrixCell, len(m.password))
	}

	// Backspace → password = CharactersPerMatrixCell chars
	m = simulateBackspace(m)
	if len(m.password) != app.CharactersPerMatrixCell {
		t.Errorf("after backspace: expected %d chars, got %d", app.CharactersPerMatrixCell, len(m.password))
	}

	// Should be able to type again
	m = simulateKey(m, "3")
	if len(m.password) != 2*app.CharactersPerMatrixCell {
		t.Errorf("after '3': expected %d chars, got %d", 2*app.CharactersPerMatrixCell, len(m.password))
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
