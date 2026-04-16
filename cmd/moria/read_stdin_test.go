package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/kiviuk/moria/internal/app"
)

func TestReadStdin_TooLarge(t *testing.T) {
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() { os.Stdin = oldStdin; r.Close() }()

	oversized := bytes.Repeat([]byte("x"), int(app.MaxMasterPasswordInputBytes)+1)
	go func() {
		_, _ = w.Write(oversized)
		w.Close()
	}()
	os.Stdin = r
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
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() { os.Stdin = oldStdin; r.Close() }()
	input := []byte("  secret\n")
	go func() {
		_, _ = w.Write(input)
		w.Close()
	}()
	os.Stdin = r
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
