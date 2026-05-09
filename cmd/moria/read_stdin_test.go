package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiviuk/moria/internal/app"
)

func TestReadStdin_TooLarge(t *testing.T) {
	masterFile := filepath.Join(t.TempDir(), "master_oversized.txt")
	oos := make([]byte, int(app.MaxMasterPasswordInputBytes)+1)
	for i := range oos {
		oos[i] = 'x'
	}
	if err := os.WriteFile(masterFile, oos, 0600); err != nil {
		t.Fatalf("write oversized file: %v", err)
	}

	t.Setenv("MORIA_MASTER_FILE", masterFile)

	sb, err := readStdin()
	if sb != nil {
		t.Fatalf("expected nil SecureBytes on error, got %v", sb)
	}
	expected := fmt.Sprintf(ErrStdinTooLarge, app.MaxMasterPasswordInputBytes/1024)
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error %q, got %v", expected, err)
	}
}

func TestReadStdin_OK(t *testing.T) {
	masterFile := filepath.Join(t.TempDir(), "master.txt")
	if err := os.WriteFile(masterFile, []byte("  secret\n"), 0600); err != nil {
		t.Fatalf("write master file: %v", err)
	}

	t.Setenv("MORIA_MASTER_FILE", masterFile)

	sb, err := readStdin()
	if err != nil {
		t.Fatalf("readStdin error: %v", err)
	}
	if sb == nil {
		t.Fatalf("expected SecureBytes return")
	}
	// TrimSpace is already applied by readStdin; check content
	if got := string(sb.Bytes()); got != "secret" {
		t.Fatalf("expected %q got %q", "secret", got)
	}
	// wipe returned secure bytes
	sb.Wipe()
	if !sb.IsWiped() {
		t.Fatalf("expected wiped SecureBytes")
	}
}
