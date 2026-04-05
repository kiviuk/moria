package testutil

import (
	"strings"
)

// NewTestMatrixData returns a deterministic string suitable for app.NewMatrix().
// The returned string has length PasswordMatrixRows * PasswordMatrixColumns * CharactersPerMatrixCell.
// Each CharactersPerMatrixCell-byte segment contains unique characters for traceability.
func NewTestMatrixData(rows, cols, cellSize int) string {
	var sb strings.Builder
	cellChars := "abcdefghijklmnopqrstuvwxyz"
	idx := 0
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			for i := 0; i < cellSize; i++ {
				sb.WriteByte(cellChars[idx%len(cellChars)])
				idx++
			}
		}
	}
	return sb.String()
}
