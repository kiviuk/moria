package app

import (
	"bytes"
	"testing"
)

func TestSecureBytes_Wipe_ZeroizesData(t *testing.T) {
	// Verify that Wipe() zeroizes the underlying data
	original := []byte("sensitive-password-data")
	sb := NewSecureBytes(original)

	sb.Wipe()

	// Verify the slice is nil after wipe
	if sb.data != nil {
		t.Error("expected data to be nil after wipe")
	}

	// Create a new SecureBytes and verify wiping works
	sb2 := NewSecureBytes([]byte("another-secret"))
	_ = sb2.Bytes() // Get reference
	sb2.Wipe()
	if sb2.data != nil {
		t.Error("expected data to be nil after wipe")
	}
}

func TestSecureBytes_IsWiped(t *testing.T) {
	sb := NewSecureBytes([]byte("test"))

	if sb.IsWiped() {
		t.Error("expected IsWiped to be false before wipe")
	}

	sb.Wipe()

	if !sb.IsWiped() {
		t.Error("expected IsWiped to be true after wipe")
	}
}

func TestSecureBytes_TrimSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"\n\ttest\n\t", "test"},
		{"no-trim", "no-trim"},
		{"   ", ""},
		{"", ""},
	}

	for _, tt := range tests {
		sb := NewSecureBytesFromString(tt.input)
		trimmed := sb.TrimSpace()

		if trimmed.String() != tt.expected {
			t.Errorf("TrimSpace(%q) = %q, expected %q", tt.input, trimmed.String(), tt.expected)
		}

		// Verify original is unchanged
		if sb.String() != tt.input {
			t.Errorf("original changed after TrimSpace: got %q, expected %q", sb.String(), tt.input)
		}

		sb.Wipe()
		trimmed.Wipe()
	}
}

func TestSecureBytes_TrimSpace_Newline(t *testing.T) {
	// Common case: piped input with trailing newline
	sb := NewSecureBytes([]byte("master-password\n"))
	trimmed := sb.TrimSpace()

	if trimmed.String() != "master-password" {
		t.Errorf("TrimSpace: got %q, expected %q", trimmed.String(), "master-password")
	}

	sb.Wipe()
	trimmed.Wipe()
}

func TestSecureBytes_Len(t *testing.T) {
	sb := NewSecureBytes([]byte("hello"))
	if sb.Len() != 5 {
		t.Errorf("Len: got %d, expected 5", sb.Len())
	}

	sb.Wipe()
	if sb.Len() != 0 {
		t.Errorf("Len after wipe: got %d, expected 0", sb.Len())
	}
}

func TestSecureBytes_Bytes(t *testing.T) {
	original := []byte("test-data")
	sb := NewSecureBytes(original)

	// Bytes() returns the underlying slice
	bytes := sb.Bytes()
	if !bytesEqual(bytes, original) {
		t.Errorf("Bytes: got %v, expected %v", bytes, original)
	}

	// Verify it's the same underlying data (not a copy)
	if &bytes[0] != &sb.data[0] {
		t.Error("Bytes() should return reference to underlying data")
	}

	sb.Wipe()
}

func TestSecureBytes_String_CreatesCopy(t *testing.T) {
	sb := NewSecureBytes([]byte("secret"))
	s := sb.String()

	// String should be a copy, so wiping shouldn't affect it
	sb.Wipe()

	if s != "secret" {
		t.Errorf("String copy affected by wipe: got %q, expected %q", s, "secret")
	}
}

func TestSecureBytes_NewSecureBytesFromString(t *testing.T) {
	input := "test-string"
	sb := NewSecureBytesFromString(input)

	if sb.String() != input {
		t.Errorf("NewSecureBytesFromString: got %q, expected %q", sb.String(), input)
	}

	// Verify it's a copy
	originalBytes := []byte(input)
	sbBytes := sb.Bytes()
	if &sbBytes[0] == &originalBytes[0] {
		t.Error("NewSecureBytesFromString should create a copy")
	}

	sb.Wipe()
}

func TestSecureBytes_DoubleWipe(t *testing.T) {
	// Verify double wipe doesn't panic
	sb := NewSecureBytes([]byte("test"))
	sb.Wipe()
	sb.Wipe() // Should be safe

	if !sb.IsWiped() {
		t.Error("expected IsWiped after double wipe")
	}
}

func TestSecureBytes_WipeOnEmpty(t *testing.T) {
	// Verify wiping empty/nil data is safe
	sb := &SecureBytes{data: nil}
	sb.Wipe() // Should not panic

	sb2 := NewSecureBytes([]byte{})
	sb2.Wipe() // Should not panic
}

func bytesEqual(a, b []byte) bool {
	return bytes.Equal(a, b)
}
