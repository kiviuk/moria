package app

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
)

// Matrix is a grid of password fragments used to generate passwords from a spell.
// Rows (0-9) correspond to character positions in the spell, wrapped by PasswordMatrixRows.
// Columns (0-9) correspond to letter groups: column 0 for non-letters,
// columns 1-9 for letter groups A-C through X-Z (CharactersPerMatrixCell letters per group).
// Each cell holds CharactersPerMatrixCell characters that are concatenated to form the password.
type Matrix [PasswordMatrixRows][PasswordMatrixColumns]string

// NewMatrix distributes a random string into the 2D matrix using arithmetic.
// The random string must be exactly PasswordMatrixRows * PasswordMatrixColumns * CharactersPerMatrixCell bytes.
func NewMatrix(randomString string) (Matrix, error) {
	expectedLen := PasswordMatrixRows * PasswordMatrixColumns * CharactersPerMatrixCell
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

// Cell returns the password fragment for a resolved query letter.
// The row is guaranteed valid by the QueryLetter type, but the column is still validated defensively.
func (m Matrix) Cell(t QueryLetter) (string, error) {
	return m.cell(t.MatrixRow, t.LetterGroup)
}

// GenerateMasterPassword produces a cryptographically secure master password of the given length.
// Characters are drawn from the provided pool using rejection sampling for zero bias.
func GenerateMasterPassword(length int, pool string) (string, error) {
	return mapToCharset(rand.Reader, pool, length), nil
}

// mapToCharset maps bytes from an io.Reader to a character pool using rejection sampling.
// Guarantees zero modulo bias regardless of pool size by discarding bytes that would create bias.
// The io.Reader can be deterministic (hkdf.New for ExpandToMatrix) or random (rand.Reader for GenerateMasterPassword).
// This design preserves determinism: the same io.Reader will always produce the same output.
func mapToCharset(source io.Reader, pool string, length int) string {
	poolBytes := []byte(pool)
	poolLen := len(poolBytes)
	// threshold defines the maximum byte value that can be used without introducing modulo bias.
	// For a pool of size N, we can only use bytes 0 to (256 - (256 % N)) - 1.
	// Bytes >= threshold are discarded to ensure uniform distribution.
	threshold := 256 - (256 % poolLen)

	result := make([]byte, length)
	buf := make([]byte, length*4)
	bytesRead := 0
	j := len(buf)

	for i := 0; i < length; i++ {
		for {
			// If buffer is exhausted, stream more bytes from the source
			if j >= bytesRead {
				n, err := source.Read(buf)
				if err != nil && err != io.EOF {
					panic(fmt.Sprintf("entropy source failed: %v", err))
				}
				bytesRead = n
				j = 0
				if bytesRead == 0 {
					panic("entropy source returned no data")
				}
			}

			b := int(buf[j])
			j++

			// Accept byte only if it falls within the unbiased range
			if b < threshold {
				result[i] = poolBytes[b%poolLen]
				break
			}
			// Otherwise, discard and try next byte (rejection sampling)
		}
	}
	return string(result)
}

// ExpandToMatrix deterministically expands any input string to exactly MatrixBytes characters.
// Uses Argon2id for memory-hard key derivation to resist brute-force attacks on weak passwords,
// followed by HKDF for expansion and rejection sampling for unbiased character mapping.
func ExpandToMatrix(input string) string {
	salt := []byte("moria-salt-v1")
	cpus := uint8(4)
	key := argon2.IDKey([]byte(input), salt, 1, 64*1024, cpus, 32)

	hkdfReader := hkdf.New(sha256.New, key, []byte("moria-salt-v1"), []byte("moria-matrix-expansion"))

	return mapToCharset(hkdfReader, MasterPasswordChars, MatrixBytes)
}

// ColHeader returns the display name for a matrix column.
// Column 0 is "Non" (non-letters), columns 1-9 are letter groups (ABC, DEF, ..., YZ).
func ColHeader(col int) string {
	if col == 0 {
		return "Non"
	}
	start := (col - 1) * CharactersPerMatrixCell
	var sb strings.Builder
	for i := 0; i < CharactersPerMatrixCell; i++ {
		letter := 'A' + rune(start+i)
		if letter > 'Z' {
			sb.WriteByte(' ')
		} else {
			sb.WriteRune(letter)
		}
	}
	return sb.String()
}

// Pretty returns a human-readable string representation of the matrix.
// Column headers are computed dynamically from AlphabetSize and CharactersPerMatrixCell.
func (m Matrix) Pretty() string {
	const colWidth = CharactersPerMatrixCell + 1

	var sb strings.Builder

	// Header row
	sb.WriteString(strings.Repeat(" ", colWidth))
	for col := 0; col < PasswordMatrixColumns; col++ {
		fmt.Fprintf(&sb, "%-*s", colWidth, ColHeader(col))
	}
	sb.WriteByte('\n')

	// Separator
	sb.WriteString(strings.Repeat(" ", colWidth))
	for col := 0; col < PasswordMatrixColumns; col++ {
		sb.WriteString(strings.Repeat("─", colWidth-1) + " ")
	}
	sb.WriteByte('\n')

	// Data rows
	for row := 0; row < PasswordMatrixRows; row++ {
		fmt.Fprintf(&sb, "%-*d", colWidth, row)
		for col := 0; col < PasswordMatrixColumns; col++ {
			fmt.Fprintf(&sb, "%-*s", colWidth, m[row][col])
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}
