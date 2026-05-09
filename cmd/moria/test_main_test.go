package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var testBinPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "moria-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	testBinPath = filepath.Join(tmpDir, "moria_bin")

	// Determine the package directory (this file's directory) and build there.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintf(os.Stderr, "runtime.Caller failed\n")
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	pkgDir := filepath.Dir(filename)

	cmd := exec.Command("go", "build", "-o", testBinPath, ".")
	cmd.Dir = pkgDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "go build failed: %v\n%s", err, string(out))
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

func runCLI(stdin string, args ...string) (string, error) {
	cmd := exec.Command(testBinPath, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}
