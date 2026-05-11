//go:build !windows

package main

import (
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
