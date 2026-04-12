package app

import (
	"strings"

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
	if sb.data != nil {
		memguard.WipeBytes(sb.data)
		sb.data = nil
	}
}

// IsWiped returns true if the data has been wiped.
func (sb *SecureBytes) IsWiped() bool {
	return sb.data == nil
}

// TrimSpace returns a new SecureBytes with leading/trailing whitespace removed.
func (sb *SecureBytes) TrimSpace() *SecureBytes {
	trimmed := strings.TrimSpace(string(sb.data))
	return NewSecureBytesFromString(trimmed)
}
