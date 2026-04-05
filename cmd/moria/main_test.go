package main

import (
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/app"
	"github.com/kiviuk/moria/internal/testutil"
)

func TestPipeInput_PlainText(t *testing.T) {
	// Verify plain text is returned unchanged when piped
	input := "my-secret"
	got := strings.TrimSpace(input)
	if got != "my-secret" {
		t.Errorf("expected %q, got %q", "my-secret", got)
	}
}

func TestPipeInput_TrailingNewline(t *testing.T) {
	// Verify trailing newline from piped input is stripped
	input := "my-secret\n"
	got := strings.TrimSpace(input)
	if got != "my-secret" {
		t.Errorf("expected %q, got %q", "my-secret", got)
	}
}

func TestPipeInput_CRLF(t *testing.T) {
	// Verify Windows-style line endings are stripped
	input := "my-secret\r\n"
	got := strings.TrimSpace(input)
	if got != "my-secret" {
		t.Errorf("expected %q, got %q", "my-secret", got)
	}
}

func TestPipeInput_LeadingTrailingSpaces(t *testing.T) {
	// Verify leading and trailing whitespace is stripped
	input := "  my-secret  "
	got := strings.TrimSpace(input)
	if got != "my-secret" {
		t.Errorf("expected %q, got %q", "my-secret", got)
	}
}

func TestPipeInput_MultiLine(t *testing.T) {
	// Verify multi-line input (SSH key) preserves internal content
	input := "-----BEGIN KEY-----\nabc123\n-----END KEY-----\n"
	got := strings.TrimSpace(input)
	expected := "-----BEGIN KEY-----\nabc123\n-----END KEY-----"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPipeInput_Empty(t *testing.T) {
	// Verify empty input returns empty string
	input := ""
	got := strings.TrimSpace(input)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestPipeInput_OnlyWhitespace(t *testing.T) {
	// Verify whitespace-only input returns empty string
	input := "   \n\t  "
	got := strings.TrimSpace(input)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestBatchMode_MaxLen(t *testing.T) {
	// Verify batch mode truncates password to maxLen
	matrix := newTestMatrix()

	dirty := app.DirtySpell{Spell: "1111"}
	spell, err := dirty.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	password, err := spell.ExtractPassword(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Full password: 4 cells × CharactersPerMatrixCell
	expectedFull := matrix[0][0] + matrix[1][0] + matrix[2][0] + matrix[3%app.PasswordMatrixRows][0]
	if password != expectedFull {
		t.Fatalf("expected %q, got %q", expectedFull, password)
	}
	fullLen := 4 * app.CharactersPerMatrixCell

	tests := []struct {
		maxLen int
	}{
		{0},
		{fullLen},
		{fullLen - 2},
		{5},
		{1},
		{100},
	}

	for _, tt := range tests {
		result := password
		expectedLen := fullLen
		if tt.maxLen > 0 && len(result) > tt.maxLen {
			result = result[:tt.maxLen]
			expectedLen = tt.maxLen
		}
		if len(result) != expectedLen {
			t.Errorf("maxLen=%d: expected len %d, got %d", tt.maxLen, expectedLen, len(result))
		}
	}
}

func TestParseArgs_MaxLen(t *testing.T) {
	// Verify --max-len flag is parsed correctly
	tests := []struct {
		args        []string
		expectedMax int
		expectedErr bool
	}{
		{[]string{"--max-len", "16", "amazon"}, 16, false},
		{[]string{"--max-len", "5", "test"}, 5, false},
		{[]string{"amazon"}, 0, false},
		{[]string{"--max-len", "abc"}, 0, true},
		{[]string{"--max-len"}, 0, true},
	}

	for _, tt := range tests {
		cfg, _, err := parseArgs(tt.args)
		if tt.expectedErr {
			if err == nil {
				t.Errorf("args %v: expected error, got nil", tt.args)
			}
			continue
		}
		if err != nil {
			t.Errorf("args %v: unexpected error: %v", tt.args, err)
			continue
		}
		if cfg.MaxLen != tt.expectedMax {
			t.Errorf("args %v: expected maxLen %d, got %d", tt.args, tt.expectedMax, cfg.MaxLen)
		}
	}
}

func TestValidateConfig_SpellRequired(t *testing.T) {
	// Verify batch mode requires a spell
	cfg := Config{Mode: ModeBatch, Spell: ""}
	flags := map[string]bool{}

	err := validateConfig(cfg, flags)
	if err == nil {
		t.Error("expected error for missing spell, got nil")
	}
}

func TestValidateConfig_AllowedMods(t *testing.T) {
	// Verify --max-len and --ignore-paste are allowed only in live/batch modes
	tests := []struct {
		cfg         Config
		flags       map[string]bool
		expectedErr bool
	}{
		{Config{Mode: ModeBatch, Spell: "test", MaxLen: 16}, map[string]bool{"--max-len": true}, false},
		{Config{Mode: ModeLive, Spell: "", MaxLen: 16}, map[string]bool{"--max-len": true, "--live": true}, false},
		{Config{Mode: ModeMagic, Spell: "", MaxLen: 16}, map[string]bool{"--max-len": true, "--magic": true}, true},
		{Config{Mode: ModePretty, Spell: "", MaxLen: 16}, map[string]bool{"--max-len": true, "--pretty": true}, true},
		{Config{Mode: ModeLive, Spell: ""}, map[string]bool{"--live": true, "--ignore-paste": true}, false},
		{Config{Mode: ModeBatch, Spell: "test"}, map[string]bool{"--ignore-paste": true}, true},
		{Config{Mode: ModeMagic}, map[string]bool{"--ignore-paste": true, "--magic": true}, true},
		{Config{Mode: ModePretty}, map[string]bool{"--ignore-paste": true, "--pretty": true}, true},
	}

	for _, tt := range tests {
		err := validateConfig(tt.cfg, tt.flags)
		if tt.expectedErr {
			if err == nil {
				t.Errorf("cfg %+v: expected error, got nil", tt.cfg)
			}
		} else {
			if err != nil {
				t.Errorf("cfg %+v: unexpected error: %v", tt.cfg, err)
			}
		}
	}
}

func TestParseArgs_IgnorePaste(t *testing.T) {
	// Verify --ignore-paste flag is parsed correctly
	tests := []struct {
		args          []string
		expectedFlags map[string]bool
		expectedErr   bool
	}{
		{[]string{"--live", "--ignore-paste"}, map[string]bool{"--live": true, "--ignore-paste": true}, false},
		{[]string{"--ignore-paste", "--live"}, map[string]bool{"--live": true, "--ignore-paste": true}, false},
		{[]string{"--live"}, map[string]bool{"--live": true}, false},
		{[]string{"--ignore-paste"}, map[string]bool{"--ignore-paste": true}, false},
	}

	for _, tt := range tests {
		_, flags, err := parseArgs(tt.args)
		if tt.expectedErr {
			if err == nil {
				t.Errorf("args %v: expected error, got nil", tt.args)
			}
			continue
		}
		if err != nil {
			t.Errorf("args %v: unexpected error: %v", tt.args, err)
			continue
		}
		for flag, expected := range tt.expectedFlags {
			if flags[flag] != expected {
				t.Errorf("args %v: flag %s expected %v, got %v", tt.args, flag, expected, flags[flag])
			}
		}
	}
}

func TestBatchMode_OutputNoNewline(t *testing.T) {
	// Verify password output has no trailing newline
	matrix := newTestMatrix()

	dirty := app.DirtySpell{Spell: "test"}
	spell, err := dirty.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	password, err := spell.ExtractPassword(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.HasSuffix(password, "\n") {
		t.Error("password should not have trailing newline")
	}
}

func TestTruncatePassword_Truncates(t *testing.T) {
	password := "abcdefghij"
	result := truncatePassword(password, 5)
	if result != "abcde" {
		t.Errorf("expected %q, got %q", "abcde", result)
	}
}

func TestTruncatePassword_NoTruncateWhenShorter(t *testing.T) {
	password := "abc"
	result := truncatePassword(password, 5)
	if result != "abc" {
		t.Errorf("expected %q, got %q", "abc", result)
	}
}

func TestTruncatePassword_NoTruncateWhenEqual(t *testing.T) {
	password := "abcde"
	result := truncatePassword(password, 5)
	if result != "abcde" {
		t.Errorf("expected %q, got %q", "abcde", result)
	}
}

func TestTruncatePassword_ZeroMaxLen(t *testing.T) {
	password := "abcdef"
	result := truncatePassword(password, 0)
	if result != "abcdef" {
		t.Errorf("expected %q, got %q", "abcdef", result)
	}
}

func TestGetMatrix_ValidInput(t *testing.T) {
	matrixStr := testutil.NewTestMatrixData(app.PasswordMatrixRows, app.PasswordMatrixColumns, app.CharactersPerMatrixCell)
	matrix, err := getMatrix(matrixStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matrix == (app.Matrix{}) {
		t.Error("matrix should not be zero value")
	}
}

func TestGetMatrix_InvalidInput(t *testing.T) {
	_, err := getMatrix("too-short")
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestGetMatrix_WrongLength(t *testing.T) {
	_, err := getMatrix("a")
	if err == nil {
		t.Error("expected error for wrong length input")
	}
}
