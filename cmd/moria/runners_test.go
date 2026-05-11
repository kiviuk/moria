package main

import (
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/app"
	"github.com/kiviuk/moria/internal/testutil"
)

// newTestConfig builds a Config with a valid matrix and raw master for use in runner tests.
func newTestConfig(spell string) *Config {
	matrixData := testutil.NewTestMatrixData(app.PasswordMatrixRows, app.PasswordMatrixColumns, app.CharactersPerMatrixCell)
	master := app.NewSecureBytes(matrixData)
	return &Config{
		Mode:      ModeBatch,
		Spell:     spell,
		Master:    master,
		MasterRaw: app.NewSecureBytesFromString("test-master"),
	}
}

func TestRunMagicMode_OutputNonEmpty(t *testing.T) {
	// Verify runMagicMode writes a non-empty master password to stdout and exits cleanly.
	out := captureStdout(t, func() {
		code := runMagicMode()
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty master password output, got empty")
	}
}

func TestRunMagicMode_OutputCharsFromPool(t *testing.T) {
	// Verify every character in the generated master password belongs to MasterPasswordChars.
	out := captureStdout(t, func() { runMagicMode() })
	for _, ch := range strings.TrimSpace(out) {
		if !strings.ContainsRune(app.MasterPasswordChars, ch) {
			t.Errorf("character %q not in MasterPasswordChars pool", ch)
		}
	}
}

func TestRunPrettyMode_OutputFormat(t *testing.T) {
	// Verify runPrettyMode prints a matrix with column headers and a separator row.
	cfg := newTestConfig("")

	out := captureStdout(t, func() {
		code := runPrettyMode(cfg)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "Non") {
		t.Errorf("expected matrix header 'Non' in output, got %q", out)
	}
	if !strings.Contains(out, "─") {
		t.Errorf("expected separator '─' in output, got %q", out)
	}
}

func TestRunBatchMode_ValidSpell(t *testing.T) {
	// Verify runBatchMode writes a non-empty password to stdout for a valid spell.
	cfg := newTestConfig("amazon")

	out := captureStdout(t, func() {
		code := runBatchMode(cfg)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})
	if len(out) == 0 {
		t.Error("expected non-empty password output, got empty")
	}
}

func TestRunBatchMode_InvalidSpell(t *testing.T) {
	// Verify runBatchMode returns exit code 1 and writes to stderr for an invalid spell.
	cfg := newTestConfig("inv€lid")

	errOut := captureStderr(t, func() {
		code := runBatchMode(cfg)
		if code != 1 {
			t.Errorf("expected exit code 1 for invalid spell, got %d", code)
		}
	})
	if !strings.Contains(errOut, "Invalid spell") {
		t.Errorf("expected 'Invalid spell' in stderr, got %q", errOut)
	}
}

func TestRunPasswordStrengthMode_WritesToStderr(t *testing.T) {
	// Verify runPasswordStrengthMode writes strength output to stderr for a non-trivial password.
	// Use a long random-looking value so zxcvbn returns non-zero entropy.
	masterRaw := app.NewSecureBytesFromString("xK9@mQ3#vL7!nP2$wR5&")
	errOut := captureStderr(t, func() { runPasswordStrengthMode(masterRaw) })
	if strings.TrimSpace(errOut) == "" {
		t.Error("expected non-empty strength output on stderr")
	}
}

func TestRunMode_MagicDispatch(t *testing.T) {
	// Verify runMode dispatches to runMagicMode and exits cleanly.
	cfg := &Config{Mode: ModeMagic}
	captureStdout(t, func() {
		code := runMode(cfg, map[string]bool{"--magic": true})
		if code != 0 {
			t.Errorf("expected exit code 0 from runMode(ModeMagic), got %d", code)
		}
	})
}

func TestRunMode_BatchDispatch(t *testing.T) {
	// Verify runMode dispatches to runBatchMode and exits cleanly for a valid spell.
	cfg := newTestConfig("amazon")
	captureStdout(t, func() {
		code := runMode(cfg, map[string]bool{})
		if code != 0 {
			t.Errorf("expected exit code 0 from runMode(ModeBatch), got %d", code)
		}
	})
}

func TestInputStateName_AllValues(t *testing.T) {
	// Verify inputStateName returns a non-empty, human-readable name for every known state.
	states := []InputState{
		InputStateUnknown,
		InputStatePipedMasterWithSpellArg,
		InputStatePipedMasterNoSpell,
		InputStateInteractiveMasterWithSpellArg,
		InputStateInteractiveMasterNoSpell,
	}
	for _, s := range states {
		name := inputStateName(s)
		if name == "" {
			t.Errorf("inputStateName(%d) returned empty string", s)
		}
	}
}
