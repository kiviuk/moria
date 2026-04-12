package app

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
)

// Matrix is a grid of password fragments used to generate passwords from a spell.
// Rows (0-9) correspond to character positions in the spell, wrapped by PasswordMatrixRows.
// Columns (0-9) correspond to letter groups: column 0 for non-letters,
// columns 1-9 for letter groups A-C through X-Z (CharactersPerMatrixCell letters per group).
// Each cell holds CharactersPerMatrixCell characters that are concatenated to form the password.
type Matrix [PasswordMatrixRows][PasswordMatrixColumns]string

// Matrix factory distributes a random string into the 2D matrix using arithmetic.
// The random string must be exactly PasswordMatrixRows * PasswordMatrixColumns * CharactersPerMatrixCell bytes.
func NewMatrix(randomString string) (Matrix, error) {
	expectedLen := MatrixBytes
	if len(randomString) != expectedLen {
		return Matrix{}, fmt.Errorf("random string length %d, expected %d", len(randomString), expectedLen)
	}
	var m Matrix
	for row := range PasswordMatrixRows {
		for col := range PasswordMatrixColumns {
			start := (row*PasswordMatrixColumns + col) * CharactersPerMatrixCell
			m[row][col] = randomString[start : start+CharactersPerMatrixCell]
		}
	}
	return m, nil
}

// Cell returns the password fragment for a resolved query letter.
// The row is guaranteed valid by the QueryLetter type, but the column is still validated defensively.
func (m Matrix) Cell(t QueryLetter) (string, error) {
	return m.cell(t.MatrixRow, t.LetterGroup)
}

// cell returns the password fragment at the given row and column.
// Index validation is performed here as a defensive measure, even if input was validated upstream.
func (m Matrix) cell(row, col int) (string, error) {
	if row < 0 || row >= PasswordMatrixRows {
		return "", fmt.Errorf("row %d out of range [0, %d)", row, PasswordMatrixRows)
	}
	if col < 0 || col >= PasswordMatrixColumns {
		return "", fmt.Errorf("col %d out of range [0, %d)", col, PasswordMatrixColumns)
	}
	return m[row][col], nil
}

// Wipe zeroizes all cells in the matrix.
// Should be called when the matrix is no longer needed to prevent sensitive
// data from lingering in memory after the program exits.
func (m *Matrix) Wipe() {
	for row := range PasswordMatrixRows {
		for col := range PasswordMatrixColumns {
			memguard.WipeBytes([]byte(m[row][col]))
			m[row][col] = ""
		}
	}
}

// Pretty returns a human-readable string representation of the matrix.
// Column headers are computed dynamically from AlphabetSize and CharactersPerMatrixCell.
func (m Matrix) Pretty() string {
	const colWidth = CharactersPerMatrixCell + 1

	var sb strings.Builder

	// Header row
	sb.WriteString(strings.Repeat(" ", colWidth))
	for col := range PasswordMatrixColumns {
		fmt.Fprintf(&sb, "%-*s", colWidth, ColHeader(col))
	}
	sb.WriteByte('\n')

	// Separator
	sb.WriteString(strings.Repeat(" ", colWidth))
	for range PasswordMatrixColumns {
		sb.WriteString(strings.Repeat("─", colWidth-1) + " ")
	}
	sb.WriteByte('\n')

	// Data rows
	for row := range PasswordMatrixRows {
		fmt.Fprintf(&sb, "%-*d", colWidth, row)
		for col := range PasswordMatrixColumns {
			fmt.Fprintf(&sb, "%-*s", colWidth, m[row][col])
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}

// GenerateMasterPassword produces a cryptographically secure master password of the given length.
// Characters are drawn from the provided pool using rejection sampling for zero bias.
// Returns a SecureBytes that can be securely wiped when no longer needed.
func GenerateMasterPassword(length int, pool string) (*SecureBytes, error) {
	result, err := mapBytesSourceToAlphabet(rand.Reader, pool, length)
	if err != nil {
		return nil, err
	}
	return NewSecureBytes(result), nil
}

// mapBytesSourceToAlphabet maps bytes from an io.Reader to an alphabet using rejection sampling.
// Guarantees zero modulo bias regardless of alphabet size by discarding bytes that would create bias.
// Returns a []byte that should be wiped by the caller or wrapped in SecureBytes.
func mapBytesSourceToAlphabet(source io.Reader, alphabet string, length int) ([]byte, error) {
	alphabetLen := len(alphabet)
	threshold := 256 - (256 % alphabetLen)

	result := make([]byte, length)
	buf := make([]byte, length*4)
	bytesRead := 0
	j := len(buf)

	for i := range length {
		for {
			if j >= bytesRead {
				n, err := source.Read(buf)
				if err != nil && err != io.EOF {
					memguard.WipeBytes(result)
					memguard.WipeBytes(buf)
					return nil, fmt.Errorf("entropy source failed: %w", err)
				}
				bytesRead = n
				j = 0
				if bytesRead == 0 {
					memguard.WipeBytes(result)
					memguard.WipeBytes(buf)
					return nil, fmt.Errorf("entropy source returned no data")
				}
			}

			b := int(buf[j])
			j++

			if b < threshold {
				result[i] = alphabet[b%alphabetLen]
				break
			}
		}
	}

	memguard.WipeBytes(buf)
	return result, nil
}

// ExpandToMatrix deterministically expands any input to exactly MatrixBytes characters.
// Uses Argon2id for memory-hard key derivation to resist brute-force attacks on weak passwords,
// followed by HKDF for expansion and rejection sampling for unbiased character mapping.
// Returns a SecureBytes that can be securely wiped when no longer needed.
func ExpandToMatrix(input *SecureBytes) *SecureBytes {
	cpus := uint8(4)
	saltBytes := []byte(Argon2Salt)
	key := argon2.IDKey(input.Bytes(), saltBytes, 1, 64*1024, cpus, 32)

	hkdfReader := hkdf.New(sha256.New, key, nil, []byte("moria-matrix-expansion"))

	result, err := mapBytesSourceToAlphabet(hkdfReader, MasterPasswordChars, MatrixBytes)
	if err != nil {
		memguard.WipeBytes(key)
		panic(fmt.Sprintf("deterministic expansion failed: %v", err))
	}

	memguard.WipeBytes(key)

	sb := NewSecureBytes(result)
	memguard.WipeBytes(result)
	return sb
}

// ColHeader returns the display name for a matrix column.
// Column 0 is "Non" (non-letters), columns 1-9 are letter groups (ABC, DEF, ..., YZ).
func ColHeader(col int) string {
	if col == 0 {
		return "Non"
	}
	start := (col - 1) * CharactersPerMatrixCell
	var sb strings.Builder
	for i := range CharactersPerMatrixCell {
		letter := 'A' + rune(start+i)
		if letter > 'Z' {
			sb.WriteByte(' ')
		} else {
			sb.WriteRune(letter)
		}
	}
	return sb.String()
}
