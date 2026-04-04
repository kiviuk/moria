package main

import (
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

	// Full password: "00|10|20|30|" = 12 chars (4 cells × 3)
	if password != "00|10|20|30|" {
		t.Fatalf("expected %q, got %q", "00|10|20|30|", password)
	}

	tests := []struct {
		maxLen      int
		expectedLen int
		expected    string
	}{
		{0, 12, "00|10|20|30|"},
		{12, 12, "00|10|20|30|"},
		{10, 10, "00|10|20|3"},
		{5, 5, "00|10"},
		{1, 1, "0"},
		{100, 12, "00|10|20|30|"},
	}

	for _, tt := range tests {
		result := password
		if tt.maxLen > 0 && len(result) > tt.maxLen {
			result = result[:tt.maxLen]
		}
		if len(result) != tt.expectedLen {
			t.Errorf("maxLen=%d: expected len %d, got %d", tt.maxLen, tt.expectedLen, len(result))
		}
		if result != tt.expected {
			t.Errorf("maxLen=%d: expected %q, got %q", tt.maxLen, tt.expected, result)
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
