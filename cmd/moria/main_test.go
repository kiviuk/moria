package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/app"
	"github.com/kiviuk/moria/internal/testutil"
)

// captureStdout redirects os.Stdout to a pipe for the duration of f, returns captured output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

// captureStderr redirects os.Stderr to a pipe for the duration of f, returns captured output.
func captureStderr(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	f()
	w.Close()
	os.Stderr = old
	out, _ := io.ReadAll(r)
	return string(out)
}

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
	// Verify --max-len flag is parsed correctly and spell is captured
	tests := []struct {
		args          []string
		expectedMax   int
		expectedSpell string
		expectedErr   bool
	}{
		{[]string{"--max-len", "16", "amazon"}, 16, "amazon", expectsOK},
		{[]string{"--max-len", "5", "test"}, 5, "test", expectsOK},
		{[]string{"amazon"}, 0, "amazon", expectsOK},
		{[]string{"--max-len", "abc"}, 0, "", expectsError},
		{[]string{"--max-len"}, 0, "", expectsError},
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
		if cfg.Spell != tt.expectedSpell {
			t.Errorf("args %v: expected spell %q, got %q", tt.args, tt.expectedSpell, cfg.Spell)
		}
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	// Verify unknown flags starting with -- are rejected unless after --
	tests := []struct {
		args        []string
		expectedErr bool
	}{
		{[]string{"--unknown-flag"}, expectsError},
		{[]string{"--s"}, expectsError},
		{[]string{"--show-strength-typo"}, expectsError},
		{[]string{"--", "--s"}, expectsOK},
		{[]string{"--", "--unknown"}, expectsOK},
		{[]string{"amazon"}, expectsOK},
	}

	for _, tt := range tests {
		_, _, err := parseArgs(tt.args)
		if tt.expectedErr {
			if err == nil {
				t.Errorf("args %v: expected error, got nil", tt.args)
			}
		} else {
			if err != nil {
				t.Errorf("args %v: unexpected error: %v", tt.args, err)
			}
		}
	}
}

func TestParseArgs_FlagSeparator(t *testing.T) {
	// Verify -- separator allows spells starting with --
	cfg, _, err := parseArgs([]string{"--", "--my-spell"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Spell != "--my-spell" {
		t.Errorf("expected spell %q, got %q", "--my-spell", cfg.Spell)
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
	matrixData := testutil.NewTestMatrixData(app.PasswordMatrixRows, app.PasswordMatrixColumns, app.CharactersPerMatrixCell)
	sb := app.NewSecureBytes(matrixData)
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

func TestPrintUsage_ContainsOptions(t *testing.T) {
	// Verify printUsage writes the expected flags and sections to stdout.
	out := captureStdout(t, printUsage)
	for _, want := range []string{"--magic", "--pretty", "--live", "--max-len", "--help", "Examples:"} {
		if !strings.Contains(out, want) {
			t.Errorf("printUsage output missing %q", want)
		}
	}
}

func TestPrintStrengthTable_ZeroEntropy(t *testing.T) {
	// Verify printStrengthTable writes nothing when entropy is 0.
	result := app.MasterPasswordResult{Entropy: 0}
	out := captureStderr(t, func() { printStrengthTable(result) })
	if out != "" {
		t.Errorf("expected no output for zero entropy, got %q", out)
	}
}

func TestPrintStrengthTable_NonZeroEntropy(t *testing.T) {
	// Verify printStrengthTable writes entropy and crack time to stderr when entropy > 0.
	result := app.MasterPasswordResult{Entropy: 100, CrackTimeDisplay: "centuries", CrackTimeSeconds: 1e10, Score: 4}
	out := captureStderr(t, func() { printStrengthTable(result) })
	if !strings.Contains(out, "100") {
		t.Errorf("expected entropy 100 in output, got %q", out)
	}
	if !strings.Contains(out, "centuries") {
		t.Errorf("expected crack time display in output, got %q", out)
	}
}

func TestDetermineInputState_PipedMasterWithSpell(t *testing.T) {
	// Verify MORIA_MASTER_FILE env var + spell yields InputStatePipedMasterWithSpellArg.
	t.Setenv("MORIA_MASTER_FILE", "/any/path")
	cfg := Config{Spell: "amazon"}
	state, err := determineInputState(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != InputStatePipedMasterWithSpellArg {
		t.Errorf("expected InputStatePipedMasterWithSpellArg (%d), got %d", InputStatePipedMasterWithSpellArg, state)
	}
}

func TestDetermineInputState_PipedMasterNoSpell(t *testing.T) {
	// Verify MORIA_MASTER_FILE env var + empty spell yields InputStatePipedMasterNoSpell.
	t.Setenv("MORIA_MASTER_FILE", "/any/path")
	cfg := Config{Spell: ""}
	state, err := determineInputState(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != InputStatePipedMasterNoSpell {
		t.Errorf("expected InputStatePipedMasterNoSpell (%d), got %d", InputStatePipedMasterNoSpell, state)
	}
}
