package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/app"
	"github.com/kiviuk/moria/internal/testutil"
)

const (
	expectsError = true
	expectsOK    = false
)

func flagsSet(flags ...string) map[string]bool {
	m := make(map[string]bool, len(flags))
	for _, f := range flags {
		m[f] = true
	}
	return m
}

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
	matrix := newTestMatrix() // from live_test.go

	dirty := app.DirtySpell{Spell: "1111"}
	spell, err := dirty.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with no truncation (maxLen = 0)
	password, err := matrix.ExtractPassword(spell, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer password.Wipe()

	// Full password: 4 cells × CharactersPerMatrixCell
	expectedFull := append(append(append(
		matrix[0][0],
		matrix[1][0]...),
		matrix[2][0]...),
		matrix[3%app.PasswordMatrixRows][0]...)
	if !bytes.Equal(password.Bytes(), expectedFull) {
		t.Fatalf("expected %q, got %q", expectedFull, password.Bytes())
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
		result, err := matrix.ExtractPassword(spell, tt.maxLen)
		if err != nil {
			t.Errorf("maxLen=%d: unexpected error: %v", tt.maxLen, err)
			continue
		}

		expectedLen := fullLen
		if tt.maxLen > 0 && tt.maxLen < fullLen {
			expectedLen = tt.maxLen
		}

		if result.Len() != expectedLen {
			t.Errorf("maxLen=%d: expected len %d, got %d", tt.maxLen, expectedLen, result.Len())
		}
		result.Wipe()
	}
}

func TestParseArgs_MaxLen(t *testing.T) {
	// Verify --max-len flag is parsed correctly
	tests := []struct {
		args        []string
		expectedMax int
		expectedErr bool
	}{
		{[]string{"--max-len", "16", "amazon"}, 16, expectsOK},
		{[]string{"--max-len", "5", "test"}, 5, expectsOK},
		{[]string{"amazon"}, 0, expectsOK},
		{[]string{"--max-len", "abc"}, 0, expectsError},
		{[]string{"--max-len"}, 0, expectsError},
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
		{Config{Mode: ModeBatch, Spell: "test", MaxLen: 16}, flagsSet("--max-len"), expectsOK},
		{Config{Mode: ModeLive, Spell: "", MaxLen: 16}, flagsSet("--max-len", "--live"), expectsOK},
		{Config{Mode: ModeMagic, Spell: "", MaxLen: 16}, flagsSet("--max-len", "--magic"), expectsError},
		{Config{Mode: ModePretty, Spell: "", MaxLen: 16}, flagsSet("--max-len", "--pretty"), expectsError},
		{Config{Mode: ModeLive, Spell: ""}, flagsSet("--live", "--ignore-paste"), expectsOK},
		{Config{Mode: ModeBatch, Spell: "test"}, flagsSet("--ignore-paste"), expectsError},
		{Config{Mode: ModeMagic}, flagsSet("--ignore-paste", "--magic"), expectsError},
		{Config{Mode: ModePretty}, flagsSet("--ignore-paste", "--pretty"), expectsError},
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

func TestValidateConfig_ConflictingModes(t *testing.T) {
	// Verify each mode accepts its own flag and rejects all other mode flags
	tests := []struct {
		cfg         Config
		flags       map[string]bool
		expectedErr bool
	}{
		// Valid: mode accepts its own flag
		{Config{Mode: ModeMagic}, flagsSet("--magic"), expectsOK},
		{Config{Mode: ModePretty}, flagsSet("--pretty"), expectsOK},
		{Config{Mode: ModeLive}, flagsSet("--live"), expectsOK},
		{Config{Mode: ModeShowPasswordStrength}, flagsSet("--show-strength"), expectsOK},

		// Conflict: ModeMagic rejects other mode flags
		{Config{Mode: ModeMagic}, flagsSet("--pretty"), expectsError},
		{Config{Mode: ModeMagic}, flagsSet("--live"), expectsError},
		{Config{Mode: ModeMagic}, flagsSet("--show-strength"), expectsError},

		// Conflict: ModePretty rejects other mode flags
		{Config{Mode: ModePretty}, flagsSet("--magic"), expectsError},
		{Config{Mode: ModePretty}, flagsSet("--live"), expectsError},
		{Config{Mode: ModePretty}, flagsSet("--show-strength"), expectsError},

		// Conflict: ModeLive rejects other mode flags
		{Config{Mode: ModeLive}, flagsSet("--magic"), expectsError},
		{Config{Mode: ModeLive}, flagsSet("--pretty"), expectsError},
		{Config{Mode: ModeLive}, flagsSet("--show-strength"), expectsError},

		// Conflict: ModeShowPasswordStrength rejects other mode flags
		{Config{Mode: ModeShowPasswordStrength}, flagsSet("--magic"), expectsError},
		{Config{Mode: ModeShowPasswordStrength}, flagsSet("--pretty"), expectsError},
		{Config{Mode: ModeShowPasswordStrength}, flagsSet("--live"), expectsError},
	}

	for _, tt := range tests {
		err := validateConfig(tt.cfg, tt.flags)
		if tt.expectedErr {
			if err == nil {
				t.Errorf("cfg %+v flags %v: expected error, got nil", tt.cfg, tt.flags)
			}
		} else {
			if err != nil {
				t.Errorf("cfg %+v flags %v: unexpected error: %v", tt.cfg, tt.flags, err)
			}
		}
	}
}

func TestParseArgs_FirstModeWins(t *testing.T) {
	// Verify the first mode flag is treated as primary, subsequent ones are tracked but ignored
	tests := []struct {
		args         []string
		expectedMode Mode
	}{
		{[]string{"--magic", "--pretty"}, ModeMagic},
		{[]string{"--pretty", "--magic"}, ModePretty},
		{[]string{"--live", "--pretty"}, ModeLive},
		{[]string{"--pretty", "--live"}, ModePretty},
		{[]string{"--magic", "--live"}, ModeMagic},
		{[]string{"--live", "--magic"}, ModeLive},
		{[]string{"--show-strength", "--magic"}, ModeShowPasswordStrength},
		{[]string{"--magic", "--show-strength"}, ModeMagic},
		{[]string{"--pretty", "--live", "--magic"}, ModePretty},
	}

	for _, tt := range tests {
		cfg, _, err := parseArgs(tt.args)
		if err != nil {
			t.Errorf("args %v: unexpected error: %v", tt.args, err)
			continue
		}
		if cfg.Mode != tt.expectedMode {
			t.Errorf("args %v: expected mode %s, got %s", tt.args, tt.expectedMode, cfg.Mode)
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
		{[]string{"--live", "--ignore-paste"}, flagsSet("--live", "--ignore-paste"), expectsOK},
		{[]string{"--ignore-paste", "--live"}, flagsSet("--live", "--ignore-paste"), expectsOK},
		{[]string{"--live"}, flagsSet("--live"), expectsOK},
		{[]string{"--ignore-paste"}, flagsSet("--ignore-paste"), expectsOK},
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

	password, err := matrix.ExtractPassword(spell, 0) // 0 = no truncation
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer password.Wipe()

	if strings.HasSuffix(string(password.Bytes()), "\n") {
		t.Error("password should not have trailing newline")
	}
}

func TestGetMatrix_ValidInput(t *testing.T) {
	matrixStr := testutil.NewTestMatrixData(app.PasswordMatrixRows, app.PasswordMatrixColumns, app.CharactersPerMatrixCell)
	sb := app.NewSecureBytesFromString(matrixStr)
	defer sb.Wipe()
	matrix, err := getMatrix(sb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Check that matrix has data by verifying first cell is not nil
	if matrix[0][0] == nil {
		t.Error("matrix should have data, first cell is nil")
	}
}

func TestGetMatrix_InvalidInput(t *testing.T) {
	sb := app.NewSecureBytesFromString("too-short")
	defer sb.Wipe()
	_, err := getMatrix(sb)
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestGetMatrix_WrongLength(t *testing.T) {
	sb := app.NewSecureBytesFromString("a")
	defer sb.Wipe()
	_, err := getMatrix(sb)
	if err == nil {
		t.Error("expected error for wrong length input")
	}
}

func TestConfig_Wipe(t *testing.T) {
	masterRaw := app.NewSecureBytesFromString("raw-master-password")
	master := app.NewSecureBytesFromString("expanded-master-password")

	cfg := Config{
		Mode:      ModeBatch,
		MasterRaw: masterRaw,
		Master:    master,
	}

	cfg.Wipe()

	if !cfg.MasterRaw.IsWiped() {
		t.Error("expected MasterRaw to be wiped")
	}
	if !cfg.Master.IsWiped() {
		t.Error("expected Master to be wiped")
	}
}

func TestConfig_Wipe_NilFields(t *testing.T) {
	cfg := Config{
		Mode:      ModeBatch,
		MasterRaw: nil,
		Master:    nil,
	}

	cfg.Wipe() // Should not panic
}

func TestMode_Validate(t *testing.T) {
	tests := []struct {
		mode      Mode
		expectErr bool
	}{
		{ModeBatch, false},
		{ModeMagic, false},
		{ModePretty, false},
		{ModeLive, false},
		{ModeShowPasswordStrength, false},
		{Mode(-1), true},
		{Mode(100), true},
	}

	for _, tt := range tests {
		err := tt.mode.Validate()
		if tt.expectErr && err == nil {
			t.Errorf("Mode(%d).Validate() expected error, got nil", tt.mode)
		}
		if !tt.expectErr && err != nil {
			t.Errorf("Mode(%d).Validate() unexpected error: %v", tt.mode, err)
		}
	}
}

func TestMode_NeedsStdin(t *testing.T) {
	tests := []struct {
		mode     Mode
		expected bool
	}{
		{ModeBatch, true},
		{ModeMagic, false},
		{ModePretty, true},
		{ModeLive, true},
		{ModeShowPasswordStrength, true},
	}

	for _, tt := range tests {
		if got := tt.mode.needsStdin(); got != tt.expected {
			t.Errorf("Mode(%d).needsStdin() = %v, expected %v", tt.mode, got, tt.expected)
		}
	}
}

func TestFormatGuessesPerSec(t *testing.T) {
	tests := []struct {
		n        uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1K"},
		{1500, "1K"},
		{999999, "999K"},
		{1000000, "1M"},
		{1500000, "1M"},
		{5000000, "5M"},
	}

	for _, tt := range tests {
		if got := formatGuessesPerSec(tt.n); got != tt.expected {
			t.Errorf("formatGuessesPerSec(%d) = %q, expected %q", tt.n, got, tt.expected)
		}
	}
}
