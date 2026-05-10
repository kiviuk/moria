package main

import (
	"fmt"
	"os"
)

// openTTY returns the (in, out) files to use for a Bubbletea program plus a no-arg close function.
// If stdin is already a TTY, returns (os.Stdin, os.Stdout, noop, nil).
// Otherwise opens /dev/tty and returns it as both input and output.
func openTTY() (in, out *os.File, closeFn func(), err error) {
	if stat, sErr := os.Stdin.Stat(); sErr == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
		return os.Stdin, os.Stdout, func() {}, nil
	}
	tty, oErr := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if oErr != nil {
		return nil, nil, nil, oErr
	}
	return tty, tty, func() { tty.Close() }, nil
}

// debugEnabled checks whether debug logging is active via MORIA_DEBUG env var.
func debugEnabled() bool {
	return os.Getenv("MORIA_DEBUG") != ""
}

// debugf prints debug messages to stderr when MORIA_DEBUG is set.
func debugf(format string, args ...interface{}) {
	if !debugEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "DEBUG: "+format+"\n", args...)
}

// inputStateName returns a human-friendly name for an InputState value.
func inputStateName(s InputState) string {
	switch s {
	case InputStateUnknown:
		return "Unknown"
	case InputStatePipedMasterWithSpellArg:
		return "PipedMasterWithSpellArg"
	case InputStatePipedMasterNoSpell:
		return "PipedMasterNoSpell"
	case InputStateInteractiveMasterWithSpellArg:
		return "InteractiveMasterWithSpellArg"
	case InputStateInteractiveMasterNoSpell:
		return "InteractiveMasterNoSpell"
	default:
		return fmt.Sprintf("InputState(%d)", s)
	}
}
