package app

import (
	"github.com/awnumar/memguard"
)

// SecureBytes holds a mutable byte buffer that is locked into physical RAM,
// preventing the OS from writing it to swap (disk). The buffer is securely
// zeroed when Wipe is called.
//
// WARNING: Bytes() returns a direct reference to the locked buffer.
// Do not retain the slice after calling Wipe().
type SecureBytes struct {
	buf *memguard.LockedBuffer
}

// NewSecureBytes creates a SecureBytes from a byte slice, copying the data into
// a locked memory region. The original slice can be safely wiped by the caller.
func NewSecureBytes(data []byte) *SecureBytes {
	if len(data) == 0 {
		return &SecureBytes{buf: memguard.NewBuffer(0)}
	}
	return &SecureBytes{buf: memguard.NewBufferFromBytes(data)}
}

// NewSecureBytesFromString creates a SecureBytes from a string, copying the data
// into a locked memory region.
// Note: The original string's backing array cannot be wiped — only this copy can be.
func NewSecureBytesFromString(s string) *SecureBytes {
	return NewSecureBytes([]byte(s))
}

// Bytes returns the underlying byte slice from the locked buffer.
// WARNING: The returned slice references locked memory; do not retain it after Wipe().
func (sb *SecureBytes) Bytes() []byte {
	if sb.buf == nil || !sb.buf.IsAlive() {
		return nil
	}
	return sb.buf.Bytes()
}

// String returns the data as a string copy.
// WARNING: The created string cannot be wiped from memory. Use sparingly.
func (sb *SecureBytes) String() string {
	if sb.buf == nil || !sb.buf.IsAlive() {
		return ""
	}
	// Use string([]byte) conversion which copies, NOT buf.String() which
	// returns an unsafe pointer into the locked buffer and dangling after Destroy.
	return string(sb.buf.Bytes())
}

// Len returns the length of the data.
func (sb *SecureBytes) Len() int {
	if sb.buf == nil || !sb.buf.IsAlive() {
		return 0
	}
	return sb.buf.Size()
}

// Wipe securely zeroes and releases the locked memory region.
// After calling Wipe, the SecureBytes is empty and must not be used.
func (sb *SecureBytes) Wipe() {
	if sb.buf != nil && sb.buf.IsAlive() {
		sb.buf.Destroy()
	}
	sb.buf = nil
}

// IsWiped returns true if the data has been wiped.
func (sb *SecureBytes) IsWiped() bool {
	return sb.buf == nil || !sb.buf.IsAlive()
}

// TrimSpace removes leading/trailing whitespace, returning a new SecureBytes
// with the trimmed content. The original is wiped.
func (sb *SecureBytes) TrimSpace() *SecureBytes {
	if sb.buf == nil || !sb.buf.IsAlive() {
		return sb
	}
	data := sb.buf.Bytes()

	start := 0
	for start < len(data) && isWhitespace(data[start]) {
		start++
	}
	end := len(data)
	for end > start && isWhitespace(data[end-1]) {
		end--
	}

	// Copy trimmed region to a plain []byte before destroying the locked buffer.
	// We must NOT pass a sub-slice of a locked buffer to NewBufferFromBytes:
	// memguard wipes the source after copying, which would fault on the guard pages.
	tmp := make([]byte, end-start)
	copy(tmp, data[start:end])
	sb.Wipe()

	return NewSecureBytes(tmp)
}

// isWhitespace returns true if the byte is whitespace (space, tab, newline, carriage return).
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
