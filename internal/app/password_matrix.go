package app

import (
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
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
	raw := make([]byte, length*2)
	_, err := rand.Read(raw)
	if err != nil {
		return "", err
	}
	return mapToCharset(raw, pool, length), nil
}

// mapToCharset maps random bytes to a character pool using rejection sampling.
// Guarantees zero modulo bias regardless of pool size.
func mapToCharset(raw []byte, pool string, length int) string {
	poolBytes := []byte(pool)
	poolLen := len(poolBytes)
	threshold := 256 - (256 % poolLen)

	result := make([]byte, length)
	j := 0
	for i := 0; i < length; i++ {
		for {
			b := int(raw[j])
			j++
			if b < threshold {
				result[i] = poolBytes[b%poolLen]
				break
			}
			if j >= len(raw) {
				more := make([]byte, length*2)
				rand.Read(more)
				raw = append(raw, more...)
			}
		}
	}
	return string(result)
}

// ExpandToMatrix deterministically expands any input string to exactly MatrixBytes characters.
// Uses Argon2id for memory-hard key derivation to resist brute-force attacks on weak passwords,
// followed by HKDF for expansion and rejection sampling for unbiased character mapping.
func ExpandToMatrix(input string) string {
	// Argon2id parameters: time=1, memory=64MB, threads=2, keyLength=32
	// Provides ~500ms derivation time, making brute-force attacks infeasible.
	salt := []byte("moria-salt-v1")
	key := argon2.IDKey([]byte(input), salt, 1, 64*1024, 2, 32)

	// Expand the 32-byte high-entropy key to MatrixBytes using HKDF.
	// Safe here because Argon2id output is already high-entropy.
	raw, err := hkdf.Key(sha256.New, key, nil, "moria-matrix-expansion", MatrixBytes*2)
	if err != nil {
		panic(err)
	}

	return mapToCharset(raw, MasterPasswordChars, MatrixBytes)
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
