package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/app"
)

// Test that when the master password is provided via MORIA_MASTER_FILE and MORIA_SPELL is set,
// the program reads the master and returns a password (non-PTY run).
func TestCLI_PipedMaster_PromptsSpellUsingPTY(t *testing.T) {
	masterFile := filepath.Join(t.TempDir(), "master.txt")
	if err := os.WriteFile(masterFile, []byte("master-pass\n"), 0600); err != nil {
		t.Fatalf("write master file: %v", err)
	}

	t.Setenv("MORIA_DEBUG", "1")
	t.Setenv("MORIA_MASTER_FILE", masterFile)
	t.Setenv("MORIA_SPELL", "amazon")

	out, err := runCLI("")
	if err != nil {
		t.Fatalf("cli failed: %v; output: %s", err, out)
	}

	if strings.Contains(out, "Error:") {
		t.Fatalf("unexpected Error in output: %s", out)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("expected non-empty password output, got empty; full output: %q", out)
	}
}

// Test that oversized stdin still triggers the safety limit when using MORIA_MASTER_FILE.
func TestCLI_PipedMaster_Oversized_ReportsError(t *testing.T) {
	masterFile := filepath.Join(t.TempDir(), "master_oversized.txt")
	oos := make([]byte, int(app.MaxMasterPasswordInputBytes)+1)
	for i := range oos {
		oos[i] = 'x'
	}
	if err := os.WriteFile(masterFile, oos, 0600); err != nil {
		t.Fatalf("write oversized file: %v", err)
	}

	t.Setenv("MORIA_MASTER_FILE", masterFile)

	out, err := runCLI("")
	if err == nil {
		t.Fatalf("expected non-zero exit for oversized stdin; output: %s", out)
	}
	want := fmt.Sprintf(ErrStdinTooLarge, app.MaxMasterPasswordInputBytes/1024)
	if !strings.Contains(out, want) {
		t.Fatalf("expected error containing %q; got output: %q", want, out)
	}
}
