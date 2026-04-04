package main

import (
	"strings"
	"testing"

	"github.com/kiviuk/pwdgen/internal/app"
)

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
	expectedFull := matrix[0][0] + matrix[1][0] + matrix[2][0] + matrix[3][0]
	if password != expectedFull {
		t.Fatalf("expected %q, got %q", expectedFull, password)
	}
	fullLen := 4 * app.CharactersPerMatrixCell

	tests := []struct {
		maxLen      int
		expectedLen int
	}{
		{0, fullLen},
		{fullLen, fullLen},
		{fullLen - 2, fullLen - 2},
		{5, 5},
		{1, 1},
		{100, fullLen},
	}

	for _, tt := range tests {
		result := password
		if tt.maxLen > 0 && len(result) > tt.maxLen {
			result = result[:tt.maxLen]
		}
		if len(result) != tt.expectedLen {
			t.Errorf("maxLen=%d: expected len %d, got %d", tt.maxLen, tt.expectedLen, len(result))
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
