package testutil

import (
	"strings"
)

// NewTestMatrixData returns a deterministic string suitable for creating a test matrix.
// The returned string has length rows * cols * cellSize characters.
// Each cellSize-byte segment contains unique characters for traceability.
func NewTestMatrixData(rows, cols, cellSize int) string {
	var sb strings.Builder
	cellChars := "abcdefghijklmnopqrstuvwxyz"
	idx := 0
	for range rows {
		for range cols {
			for range cellSize {
				sb.WriteByte(cellChars[idx%len(cellChars)])
				idx++
			}
		}
	}
	return sb.String()
}
