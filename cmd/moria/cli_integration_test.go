package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/app"
)

func TestCLI_WithPipe_Succeeds(t *testing.T) {
	t.Setenv("MORIA_MASTER_FILE", "")
	out, err := runCLI("master\n", "amazon")
	if err != nil {
		t.Fatalf("cli failed: %v\noutput: %s", err, out)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		t.Fatalf("expected non-empty output, got empty")
	}
	if strings.Contains(out, "Error") {
		t.Fatalf("unexpected 'Error' in stdout: %s", out)
	}
}

func TestCLI_OversizedStdin_Fails(t *testing.T) {
	t.Setenv("MORIA_MASTER_FILE", "")
	oversized := string(bytes.Repeat([]byte("x"), int(app.MaxMasterPasswordInputBytes)+1))
	out, err := runCLI(oversized, "amazon")
	if err == nil {
		t.Fatalf("expected non-zero exit for oversized stdin; output: %s", out)
	}
	want := fmt.Sprintf(ErrStdinTooLarge, app.MaxMasterPasswordInputBytes/1024)
	if !strings.Contains(out, want) {
		t.Fatalf("expected error containing %q; got output: %s", want, out)
	}
}
