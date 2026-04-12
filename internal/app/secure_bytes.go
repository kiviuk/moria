package app

import (
	"github.com/awnumar/memguard"
)

// SecureBytes holds a mutable byte buffer that can be securely wiped from memory.
// Unlike Go strings which are immutable and cannot be securely erased,
// SecureBytes uses a mutable byte slice that can be zeroized when no longer needed.
type SecureBytes struct {
	data []byte
}

// NewSecureBytes creates a SecureBytes from a byte slice, copying the data.
// The original slice can be safely wiped by the caller after this call.
func NewSecureBytes(data []byte) *SecureBytes {
	sb := &SecureBytes{
		data: make([]byte, len(data)),
	}
	copy(sb.data, data)
	return sb
}

// NewSecureBytesFromString creates a SecureBytes from a string.
// Note: The original string's backing array cannot be wiped - only this copy can be.
func NewSecureBytesFromString(s string) *SecureBytes {
	sb := &SecureBytes{
		data: make([]byte, len(s)),
	}
	copy(sb.data, s)
	return sb
}

// Bytes returns the underlying byte slice.
// WARNING: The returned slice references the internal buffer; do not retain it after Wipe().
func (sb *SecureBytes) Bytes() []byte {
	return sb.data
}

// String returns the data as a string.
// WARNING: This creates a copy - the string cannot be wiped. Use sparingly.
func (sb *SecureBytes) String() string {
	return string(sb.data)
}

// Len returns the length of the data.
func (sb *SecureBytes) Len() int {
	return len(sb.data)
}

// Wipe securely erases the data from memory.
// After calling Wipe, the SecureBytes is empty and should not be used.
func (sb *SecureBytes) Wipe() {
	memguard.WipeBytes(sb.data)
	sb.data = nil
}

// IsWiped returns true if the data has been wiped.
func (sb *SecureBytes) IsWiped() bool {
	return sb.data == nil
}

// TrimSpace removes leading/trailing whitespace in-place, returning the same SecureBytes.
// This avoids creating intermediate strings.
func (sb *SecureBytes) TrimSpace() *SecureBytes {
	if sb.data == nil {
		return sb
	}

	// Find first non-whitespace byte
	start := 0
	for start < len(sb.data) {
		if !isWhitespace(sb.data[start]) {
			break
		}
		start++
	}

	// Find last non-whitespace byte
	end := len(sb.data)
	for end > start {
		if !isWhitespace(sb.data[end-1]) {
			break
		}
		end--
	}

	// Zero out the trimmed portions
	for i := range start {
		sb.data[i] = 0
	}
	for i := end; i < len(sb.data); i++ {
		sb.data[i] = 0
	}

	// Slice to the trimmed portion
	sb.data = sb.data[start:end]
	return sb
}

// isWhitespace returns true if the byte is whitespace (space, tab, newline, carriage return).
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
