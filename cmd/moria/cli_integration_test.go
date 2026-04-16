package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/app"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "moria_bin")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\noutput: %s", err, string(out))
	}
	return bin
}

func TestCLI_WithPipe_Succeeds(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "amazon")
	cmd.Stdin = bytes.NewBufferString("master\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("cli failed: %v\nstderr: %s", err, stderr.String())
	}

	out := strings.TrimSpace(stdout.String())
	if out == "" {
		t.Fatalf("expected non-empty output, got empty; stderr: %s", stderr.String())
	}
	if strings.Contains(out, "Error") {
		t.Fatalf("unexpected 'Error' in stdout: %s", out)
	}
}

func TestCLI_OversizedStdin_Fails(t *testing.T) {
	bin := buildBinary(t)
	oversized := bytes.Repeat([]byte("x"), int(app.MaxMasterPasswordInputBytes)+1)
	cmd := exec.Command(bin, "amazon")
	cmd.Stdin = bytes.NewReader(oversized)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero exit for oversized stdin; stdout: %s", stdout.String())
	}
	want := fmt.Sprintf(ErrStdinTooLarge, app.MaxMasterPasswordInputBytes/1024)
	if !strings.Contains(stderr.String(), want) && !strings.Contains(stdout.String(), want) {
		t.Fatalf("expected error containing %q in stderr or stdout; got stderr: %s stdout: %s", want, stderr.String(), stdout.String())
	}
}
