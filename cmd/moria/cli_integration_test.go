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

func TestCLI_Magic_NoPrompt(t *testing.T) {
	// --magic generates a master password and must exit immediately without prompting.
	// If it blocks waiting for input, runCLI will hang and the test will time out.
	t.Setenv("MORIA_MASTER_FILE", "")
	out, err := runCLI("", "--magic")
	if err != nil {
		t.Fatalf("--magic failed: %v\noutput: %s", err, out)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		t.Fatalf("expected non-empty master password output, got empty")
	}
	if strings.Contains(out, "Error") {
		t.Fatalf("unexpected 'Error' in output: %s", out)
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

func TestCLI_Determinism(t *testing.T) {
	// Verify same master + spell always produces identical output (core security property).
	t.Setenv("MORIA_MASTER_FILE", "")
	out1, err := runCLI("my-master-password\n", "amazon")
	if err != nil {
		t.Fatalf("first run failed: %v; output: %s", err, out1)
	}
	out2, err := runCLI("my-master-password\n", "amazon")
	if err != nil {
		t.Fatalf("second run failed: %v; output: %s", err, out2)
	}
	if out1 != out2 {
		t.Errorf("determinism violation: first=%q second=%q", out1, out2)
	}
	if strings.TrimSpace(out1) == "" {
		t.Error("expected non-empty password output")
	}
}
