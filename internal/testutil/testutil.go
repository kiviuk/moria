package testutil

// NewTestMatrixData returns a deterministic byte slice suitable for creating a test matrix.
// The returned slice has length rows * cols * cellSize bytes.
// Each cellSize-byte segment contains unique characters for traceability.
func NewTestMatrixData(rows, cols, cellSize int) []byte {
	cellChars := "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, rows*cols*cellSize)
	idx := 0
	for range rows {
		for range cols {
			for range cellSize {
				result[idx] = cellChars[idx%len(cellChars)]
				idx++
			}
		}
	}
	return result
}
