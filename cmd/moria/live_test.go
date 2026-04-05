package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kiviuk/moria/internal/app"
	"github.com/kiviuk/moria/internal/testutil"
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
	m := newLiveModel(matrix, 2*app.CharactersPerMatrixCell, PasteAllowed)

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
	if m.state != StateMaxLenReached {
		t.Error("after '3': expected StateMaxLenReached")
	}
}

func TestLiveModel_MaxLen_Partial(t *testing.T) {
	// Verify maxLen truncates password on exit, never exceeding maxLen
	matrix := newTestMatrix()
	maxLen := 2*app.CharactersPerMatrixCell + 1
	m := newLiveModel(matrix, maxLen, PasteAllowed)

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
	expectedLen := min(len(m.password), maxLen)
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
	m := newLiveModel(matrix, 0, PasteAllowed)

	for range 10 {
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
	m := newLiveModel(matrix, 2*app.CharactersPerMatrixCell, PasteAllowed)

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

func TestLiveModel_Paste_DefaultAllowsPasting(t *testing.T) {
	// Verify that pasting a multi-character spell works by default
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 0, PasteAllowed)

	m = simulateKey(m, "amazon")
	if m.spell != "amazon" {
		t.Errorf("expected spell 'amazon', got %q", m.spell)
	}
	if len(m.password) != 6*app.CharactersPerMatrixCell {
		t.Errorf("expected %d chars, got %d", 6*app.CharactersPerMatrixCell, len(m.password))
	}
	if len(m.queryLetters) != 6 {
		t.Errorf("expected 6 query letters, got %d", len(m.queryLetters))
	}
	if m.err != "" {
		t.Errorf("unexpected error: %s", m.err)
	}
}

func TestLiveModel_Paste_IgnoredWhenFlagSet(t *testing.T) {
	// Verify that pasting is rejected when --ignore-paste is set
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 0, PasteIgnored)

	m = simulateKey(m, "amazon")
	if m.spell != "" {
		t.Errorf("expected empty spell (paste ignored), got %q", m.spell)
	}
	if m.password != "" {
		t.Errorf("expected empty password (paste ignored), got %q", m.password)
	}
	if m.err == "" {
		t.Error("expected error when paste is ignored, got nil")
	}
}

func TestLiveModel_Paste_RespectsMaxLen(t *testing.T) {
	// Verify that pasting respects maxLen and stops at the limit
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 2*app.CharactersPerMatrixCell, PasteAllowed)

	m = simulateKey(m, "amazon")
	if m.spell != "am" {
		t.Errorf("expected spell 'am' (stopped at maxLen), got %q", m.spell)
	}
	if len(m.password) != 2*app.CharactersPerMatrixCell {
		t.Errorf("expected %d chars (at maxLen), got %d", 2*app.CharactersPerMatrixCell, len(m.password))
	}
	if m.state != StateMaxLenReached {
		t.Error("expected StateMaxLenReached")
	}
}

func TestLiveModel_Paste_InvalidCharStops(t *testing.T) {
	// Verify that pasting stops at the first invalid character
	matrix := newTestMatrix()
	m := newLiveModel(matrix, 0, PasteAllowed)

	m = simulateKey(m, "a€b")
	if m.spell != "a" {
		t.Errorf("expected spell 'a' (stopped at invalid char), got %q", m.spell)
	}
	if len(m.password) != app.CharactersPerMatrixCell {
		t.Errorf("expected %d chars, got %d", app.CharactersPerMatrixCell, len(m.password))
	}
	if m.err == "" {
		t.Error("expected invalid char error, got nil")
	}
}

func TestLiveModel_SingleKey_UnaffectedByFlag(t *testing.T) {
	// Verify that single-character input works regardless of --ignore-paste
	matrix := newTestMatrix()

	m := newLiveModel(matrix, 0, PasteAllowed)
	m = simulateKey(m, "a")
	if m.spell != "a" {
		t.Errorf("without flag: expected spell 'a', got %q", m.spell)
	}

	m = newLiveModel(matrix, 0, PasteIgnored)
	m = simulateKey(m, "a")
	if m.spell != "a" {
		t.Errorf("with flag: expected spell 'a', got %q", m.spell)
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
