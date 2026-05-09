package main

import (
	"fmt"
	"os"
)

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
