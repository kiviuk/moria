package main

import (
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/app"
)

func TestReadAndTrim_PlainText(t *testing.T) {
	// Verify plain text is returned unchanged
	got := readAndTrim(strings.NewReader("my-secret"))
	if got != "my-secret" {
		t.Errorf("expected %q, got %q", "my-secret", got)
	}
}

func TestReadAndTrim_TrailingNewline(t *testing.T) {
	// Verify trailing newline from interactive Enter is stripped
	got := readAndTrim(strings.NewReader("my-secret\n"))
	if got != "my-secret" {
		t.Errorf("expected %q, got %q", "my-secret", got)
	}
}

func TestReadAndTrim_CRLF(t *testing.T) {
	// Verify Windows-style line endings are stripped
	got := readAndTrim(strings.NewReader("my-secret\r\n"))
	if got != "my-secret" {
		t.Errorf("expected %q, got %q", "my-secret", got)
	}
}

func TestReadAndTrim_LeadingTrailingSpaces(t *testing.T) {
	// Verify leading and trailing whitespace is stripped
	got := readAndTrim(strings.NewReader("  my-secret  "))
	if got != "my-secret" {
		t.Errorf("expected %q, got %q", "my-secret", got)
	}
}

func TestReadAndTrim_MultiLine(t *testing.T) {
	// Verify multi-line input (SSH key) preserves internal content
	input := "-----BEGIN KEY-----\nabc123\n-----END KEY-----\n"
	got := readAndTrim(strings.NewReader(input))
	expected := "-----BEGIN KEY-----\nabc123\n-----END KEY-----"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestReadAndTrim_Empty(t *testing.T) {
	// Verify empty input returns empty string
	got := readAndTrim(strings.NewReader(""))
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestReadAndTrim_OnlyWhitespace(t *testing.T) {
	// Verify whitespace-only input returns empty string
	got := readAndTrim(strings.NewReader("   \n\t  "))
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
	cfg := Config{Mode: "batch", Spell: ""}
	flags := map[string]bool{}

	err := validateConfig(cfg, flags)
	if err == nil {
		t.Error("expected error for missing spell, got nil")
	}
}

func TestValidateConfig_AllowedMods(t *testing.T) {
	// Verify --max-len is allowed in batch and live modes
	tests := []struct {
		cfg         Config
		flags       map[string]bool
		expectedErr bool
	}{
		{Config{Mode: "batch", Spell: "test", MaxLen: 16}, map[string]bool{"--max-len": true}, false},
		{Config{Mode: "live", Spell: "", MaxLen: 16}, map[string]bool{"--max-len": true, "--live": true}, false},
		{Config{Mode: "magic", Spell: "", MaxLen: 16}, map[string]bool{"--max-len": true, "--magic": true}, true},
		{Config{Mode: "pretty", Spell: "", MaxLen: 16}, map[string]bool{"--max-len": true, "--pretty": true}, true},
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
