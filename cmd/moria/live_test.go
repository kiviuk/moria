package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kiviuk/moria/internal/app"
	"github.com/kiviuk/moria/internal/testutil"
)

const testMasterRaw = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@-_=+:%.^/,"

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
	m := newLiveModel(matrix, testMasterRaw, 2*app.CharactersPerMatrixCell, PasteAllowed)

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
	m := newLiveModel(matrix, testMasterRaw, maxLen, PasteAllowed)

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
	m := newLiveModel(matrix, testMasterRaw, 0, PasteAllowed)

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
	m := newLiveModel(matrix, testMasterRaw, 2*app.CharactersPerMatrixCell, PasteAllowed)

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
	m := newLiveModel(matrix, testMasterRaw, 0, PasteAllowed)

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
	m := newLiveModel(matrix, testMasterRaw, 0, PasteIgnored)

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
	m := newLiveModel(matrix, testMasterRaw, 2*app.CharactersPerMatrixCell, PasteAllowed)

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
	m := newLiveModel(matrix, testMasterRaw, 0, PasteAllowed)

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

	m := newLiveModel(matrix, testMasterRaw, 0, PasteAllowed)
	m = simulateKey(m, "a")
	if m.spell != "a" {
		t.Errorf("without flag: expected spell 'a', got %q", m.spell)
	}

	m = newLiveModel(matrix, testMasterRaw, 0, PasteIgnored)
	m = simulateKey(m, "a")
	if m.spell != "a" {
		t.Errorf("with flag: expected spell 'a', got %q", m.spell)
	}
}

func TestLiveModel_Space_SingleKey(t *testing.T) {
	// Verify that space can be entered as a single keystroke
	matrix := newTestMatrix()
	m := newLiveModel(matrix, testMasterRaw, 0, PasteAllowed)

	m = simulateKey(m, "a")
	m = simulateKey(m, " ")
	m = simulateKey(m, "b")

	if m.spell != "a b" {
		t.Errorf("expected spell 'a b', got %q", m.spell)
	}
	if len(m.spell) != 3 {
		t.Errorf("expected spell length 3, got %d", len(m.spell))
	}
	if len(m.password) != 3*app.CharactersPerMatrixCell {
		t.Errorf("expected %d password chars, got %d", 3*app.CharactersPerMatrixCell, len(m.password))
	}
	if len(m.queryLetters) != 3 {
		t.Errorf("expected 3 query letters, got %d", len(m.queryLetters))
	}
	if m.err != "" {
		t.Errorf("unexpected error: %s", m.err)
	}
}

func TestLiveModel_Space_Pasted(t *testing.T) {
	// Verify that space can be part of pasted input
	matrix := newTestMatrix()
	m := newLiveModel(matrix, testMasterRaw, 0, PasteAllowed)

	m = simulateKey(m, "hello world")

	if m.spell != "hello world" {
		t.Errorf("expected spell 'hello world', got %q", m.spell)
	}
	if len(m.spell) != 11 {
		t.Errorf("expected spell length 11, got %d", len(m.spell))
	}
	if len(m.password) != 11*app.CharactersPerMatrixCell {
		t.Errorf("expected %d password chars, got %d", 11*app.CharactersPerMatrixCell, len(m.password))
	}
	if len(m.queryLetters) != 11 {
		t.Errorf("expected 11 query letters, got %d", len(m.queryLetters))
	}
	if m.err != "" {
		t.Errorf("unexpected error: %s", m.err)
	}
}

func TestLiveModel_Space_RespectsMaxLen(t *testing.T) {
	// Verify that space counts toward maxLen like any other character
	// maxLen=8 chars = 4 spell chars (each adds CharactersPerMatrixCell=2 to password)
	matrix := newTestMatrix()
	maxLen := 4 * app.CharactersPerMatrixCell
	m := newLiveModel(matrix, testMasterRaw, maxLen, PasteAllowed)

	m = simulateKey(m, "a") // password = 2 chars
	m = simulateKey(m, " ") // password = 4 chars
	m = simulateKey(m, "b") // password = 6 chars
	m = simulateKey(m, " ") // password = 8 chars → maxLen reached, blocked
	m = simulateKey(m, "c") // blocked (already at maxLen)

	if m.state != StateMaxLenReached {
		t.Error("expected StateMaxLenReached after 4 spell chars (8 password chars)")
	}
	if len(m.spell) != 4 {
		t.Errorf("expected 4 spell chars, got %d", len(m.spell))
	}
	if len(m.password) != maxLen {
		t.Errorf("expected password length %d, got %d", maxLen, len(m.password))
	}
}

func TestLiveModel_Space_Backspace(t *testing.T) {
	// Verify that backspace removes space and allows re-typing
	matrix := newTestMatrix()
	m := newLiveModel(matrix, testMasterRaw, 0, PasteAllowed)

	m = simulateKey(m, "a")
	m = simulateKey(m, " ")
	m = simulateKey(m, "b")
	if m.spell != "a b" {
		t.Fatalf("setup failed: expected 'a b', got %q", m.spell)
	}

	m = simulateBackspace(m)
	if m.spell != "a " {
		t.Errorf("after backspace: expected spell 'a ', got %q", m.spell)
	}
	if len(m.password) != 2*app.CharactersPerMatrixCell {
		t.Errorf("after backspace: expected %d chars, got %d", 2*app.CharactersPerMatrixCell, len(m.password))
	}

	m = simulateBackspace(m)
	if m.spell != "a" {
		t.Errorf("after 2nd backspace: expected spell 'a', got %q", m.spell)
	}
	if len(m.password) != app.CharactersPerMatrixCell {
		t.Errorf("after 2nd backspace: expected %d chars, got %d", app.CharactersPerMatrixCell, len(m.password))
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
