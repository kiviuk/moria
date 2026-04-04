package app

import (
	"crypto/rand"
	"fmt"
	"strings"
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

// GenerateRandomString produces a cryptographically secure random string of the given length.
// Characters are drawn from the provided pool.
func GenerateRandomString(length int, pool string) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for i := range b {
		b[i] = pool[int(b[i])%len(pool)]
	}
	return string(b), nil
}

// colHeader returns the display name for a matrix column.
// Column 0 is "Non" (non-letters), columns 1-9 are letter groups (ABC, DEF, ..., YZ).
func colHeader(col int) string {
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
	const colWidth = 4

	var sb strings.Builder

	// Header row
	sb.WriteString(strings.Repeat(" ", colWidth))
	for col := 0; col < PasswordMatrixColumns; col++ {
		sb.WriteString(fmt.Sprintf("%-*s", colWidth, colHeader(col)))
	}
	sb.WriteByte('\n')

	// Separator
	sb.WriteString(strings.Repeat(" ", colWidth))
	for col := 0; col < PasswordMatrixColumns; col++ {
		sb.WriteString("─── ")
	}
	sb.WriteByte('\n')

	// Data rows
	for row := 0; row < PasswordMatrixRows; row++ {
		sb.WriteString(fmt.Sprintf("%-*d", colWidth, row))
		for col := 0; col < PasswordMatrixColumns; col++ {
			sb.WriteString(fmt.Sprintf("%-*s", colWidth, m[row][col]))
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}
