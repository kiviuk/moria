//go:build windows

package main

import (
	"os"
)

// openTTY on Windows opens the console device (CONIN$/CONOUT$) for interactive use.
// Falls back to os.Stdin/os.Stdout if the console cannot be opened.
func openTTY() (in, out *os.File, closeFn func(), err error) {
	conin, inErr := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if inErr != nil {
		return os.Stdin, os.Stdout, func() {}, nil
	}
	conout, outErr := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if outErr != nil {
		conin.Close()
		return os.Stdin, os.Stdout, func() {}, nil
	}
	return conin, conout, func() { conin.Close(); conout.Close() }, nil
}