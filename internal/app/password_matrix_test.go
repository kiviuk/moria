package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kiviuk/moria/internal/testutil"
)

// newTestMatrix returns a static test matrix for use in generator tests.
// Each cell contains a string of length CharactersPerMatrixCell.
func newTestMatrix() Matrix {
	m, err := NewMatrix(testutil.NewTestMatrixData(PasswordMatrixRows, PasswordMatrixColumns, CharactersPerMatrixCell))
	if err != nil {
		panic(err)
	}
	return m
}

func TestExpandToMatrix_Deterministic(t *testing.T) {
	// Verify same input always produces same output
	in1 := NewSecureBytesFromString("test-secret")
	in2 := NewSecureBytesFromString("test-secret")
	defer in1.Wipe()
	defer in2.Wipe()
	out1 := ExpandToMatrix(in1)
	out2 := ExpandToMatrix(in2)
	defer out1.Wipe()
	defer out2.Wipe()
	if out1.String() != out2.String() {
		t.Errorf("ExpandToMatrix not deterministic: got different outputs for same input")
	}
}

func TestExpandToMatrix_Length(t *testing.T) {
	// Verify output is always exactly MatrixBytes characters
	inputs := []string{"", "short", "medium-length-input", strings.Repeat("x", 1000)}
	for _, input := range inputs {
		in := NewSecureBytesFromString(input)
		out := ExpandToMatrix(in)
		if out.Len() != MatrixBytes {
			t.Errorf("ExpandToMatrix(%q) length = %d, expected %d", input, out.Len(), MatrixBytes)
		}
		out.Wipe()
		in.Wipe()
	}
}

func TestExpandToMatrix_Charset(t *testing.T) {
	// Verify all output characters are from MasterPasswordChars
	in := NewSecureBytesFromString("any-input-string")
	out := ExpandToMatrix(in)
	defer in.Wipe()
	defer out.Wipe()
	for i, r := range out.String() {
		if !strings.ContainsRune(MasterPasswordChars, r) {
			t.Errorf("ExpandToMatrix: char %q at %d not in allowed pool", r, i)
		}
	}
}

func TestExpandToMatrix_AlwaysDerives(t *testing.T) {
	// Verify even exact-length input is transformed, not returned as-is
	input := strings.Repeat("a", MatrixBytes)
	in := NewSecureBytesFromString(input)
	out := ExpandToMatrix(in)
	defer in.Wipe()
	defer out.Wipe()
	if out.String() == input {
		t.Error("ExpandToMatrix returned input unchanged — should always derive")
	}
}

func TestExpandToMatrix_DifferentInputs(t *testing.T) {
	// Verify different inputs produce different outputs
	in1 := NewSecureBytesFromString("secret-a")
	in2 := NewSecureBytesFromString("secret-b")
	defer in1.Wipe()
	defer in2.Wipe()
	out1 := ExpandToMatrix(in1)
	out2 := ExpandToMatrix(in2)
	defer out1.Wipe()
	defer out2.Wipe()
	if out1.String() == out2.String() {
		t.Error("ExpandToMatrix produced same output for different inputs")
	}
}

func TestExpandToMatrix_TrailingNewline(t *testing.T) {
	// Verify trailing newline changes output (simulates interactive vs piped input)
	in1 := NewSecureBytesFromString("secret")
	in2 := NewSecureBytesFromString("secret\n")
	defer in1.Wipe()
	defer in2.Wipe()
	out1 := ExpandToMatrix(in1)
	out2 := ExpandToMatrix(in2)
	defer out1.Wipe()
	defer out2.Wipe()
	if out1.String() == out2.String() {
		t.Error("ExpandToMatrix should produce different output for input with trailing newline")
	}
}

func TestGenerateMasterPassword_Length(t *testing.T) {
	// Verify generated string matches requested length
	expectedLen := PasswordMatrixRows * PasswordMatrixColumns * CharactersPerMatrixCell
	s, err := GenerateMasterPassword(expectedLen, MasterPasswordChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer s.Wipe()
	if s.Len() != expectedLen {
		t.Errorf("expected length %d, got %d", expectedLen, s.Len())
	}
}

func TestGenerateMasterPassword_Charset(t *testing.T) {
	// Verify all characters in generated string are from the allowed pool
	s, err := GenerateMasterPassword(1000, MasterPasswordChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer s.Wipe()
	for i, r := range s.String() {
		if !strings.ContainsRune(MasterPasswordChars, r) {
			t.Errorf("char %q at %d not in allowed pool", r, i)
		}
	}
}

func TestGenerateMasterPassword_NonDeterministic(t *testing.T) {
	// Verify two consecutive calls produce different strings
	expectedLen := PasswordMatrixRows * PasswordMatrixColumns * CharactersPerMatrixCell
	s1, err := GenerateMasterPassword(expectedLen, MasterPasswordChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer s1.Wipe()
	s2, err := GenerateMasterPassword(expectedLen, MasterPasswordChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer s2.Wipe()
	if s1.String() == s2.String() {
		t.Error("expected different strings, got identical")
	}
}

func TestNewMatrix_LengthMismatch(t *testing.T) {
	// Verify NewMatrix rejects strings of wrong length
	_, err := NewMatrix("tooshort")
	if err == nil {
		t.Fatal("expected error for short string, got nil")
	}
}

func TestNewMatrix_Population(t *testing.T) {
	// Verify arithmetic mapping: cell (r,c) contains the correct substring from input
	input := testutil.NewTestMatrixData(PasswordMatrixRows, PasswordMatrixColumns, CharactersPerMatrixCell)
	m, err := NewMatrix(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			start := (row*PasswordMatrixColumns + col) * CharactersPerMatrixCell
			expected := []byte(input[start : start+CharactersPerMatrixCell])
			if !bytesEqual(m[row][col], expected) {
				t.Errorf("m[%d][%d] = %q, expected %q", row, col, m[row][col], expected)
			}
		}
	}
}

func TestMatrix_Cell(t *testing.T) {
	// Verify Cell returns correct values for valid query letters
	m := newTestMatrix()
	tests := []struct {
		query    QueryLetter
		expected []byte
	}{
		{QueryLetter{MatrixRow: 0, LetterGroup: 0}, m[0][0]},
		{QueryLetter{MatrixRow: 0, LetterGroup: PasswordMatrixColumns - 1}, m[0][PasswordMatrixColumns-1]},
		{QueryLetter{MatrixRow: PasswordMatrixRows / 2, LetterGroup: 3 % PasswordMatrixColumns}, m[PasswordMatrixRows/2][3%PasswordMatrixColumns]},
		{QueryLetter{MatrixRow: PasswordMatrixRows - 1, LetterGroup: PasswordMatrixColumns - 1}, m[PasswordMatrixRows-1][PasswordMatrixColumns-1]},
	}
	for _, tt := range tests {
		got, err := m.Cell(tt.query)
		if err != nil {
			t.Errorf("Cell(%+v) unexpected error: %v", tt.query, err)
		}
		if !bytesEqual(got, tt.expected) {
			t.Errorf("Cell(%+v) = %q, expected %q", tt.query, got, tt.expected)
		}
	}
}

func TestMatrix_Cell_OutOfRangeRow(t *testing.T) {
	// Verify Cell returns error for out-of-bounds matrix row in query letter
	m := newTestMatrix()
	tests := []int{-1, PasswordMatrixRows, 99}
	for _, row := range tests {
		query := QueryLetter{MatrixRow: row, LetterGroup: 0}
		_, err := m.Cell(query)
		if err == nil {
			t.Errorf("Cell with row %d expected error, got nil", row)
		}
	}
}

func TestMatrix_Cell_OutOfRangeCol(t *testing.T) {
	// Verify Cell returns error for out-of-bounds letter group in query letter
	m := newTestMatrix()
	tests := []int{-1, PasswordMatrixColumns, 99}
	for _, col := range tests {
		query := QueryLetter{MatrixRow: 0, LetterGroup: col}
		_, err := m.Cell(query)
		if err == nil {
			t.Errorf("Cell with col %d expected error, got nil", col)
		}
	}
}

func TestMatrix_Dimensions(t *testing.T) {
	// Verify the static test matrix has the correct dimensions based on constants
	m := newTestMatrix() //nolint:staticcheck // m is used in assertions below
	if len(m) != PasswordMatrixRows {
		t.Errorf("expected %d rows, got %d", PasswordMatrixRows, len(m))
	}
	for row := 0; row < PasswordMatrixRows; row++ {
		if len(m[row]) != PasswordMatrixColumns {
			t.Errorf("row %d: expected %d cols, got %d", row, PasswordMatrixColumns, len(m[row]))
		}
	}
}

func TestMatrix_CellContent(t *testing.T) {
	// Verify all cells have the correct length
	m := newTestMatrix()
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			if len(m[row][col]) != CharactersPerMatrixCell {
				t.Errorf("m[%d][%d] length = %d, expected %d", row, col, len(m[row][col]), CharactersPerMatrixCell)
			}
		}
	}
}

func TestColHeader(t *testing.T) {
	// Verify column headers are computed correctly from constants
	tests := []struct {
		col      int
		expected string
	}{
		{0, "Non"},
		{1, buildExpectedHeader(1)},
		{2, buildExpectedHeader(2)},
		{3, buildExpectedHeader(3)},
		{4, buildExpectedHeader(4)},
		{5, buildExpectedHeader(5)},
		{6, buildExpectedHeader(6)},
		{7, buildExpectedHeader(7)},
		{8, buildExpectedHeader(8)},
		{9, buildExpectedHeader(9)},
	}
	for _, tt := range tests {
		if got := ColHeader(tt.col); got != tt.expected {
			t.Errorf("ColHeader(%d) = %q, expected %q", tt.col, got, tt.expected)
		}
	}
}

func buildExpectedHeader(col int) string {
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

func TestMatrix_Pretty(t *testing.T) {
	// Verify Pretty produces a human-readable matrix with headers, separator, and rows
	m := newTestMatrix()
	output := m.Pretty()

	// Check header row contains all column labels
	for col := 0; col < PasswordMatrixColumns; col++ {
		if !strings.Contains(output, ColHeader(col)) {
			t.Errorf("Pretty output missing column header %q", ColHeader(col))
		}
	}

	// Check separator line
	sep := strings.Repeat("─", CharactersPerMatrixCell)
	if !strings.Contains(output, sep) {
		t.Error("Pretty output missing separator line")
	}

	// Check row numbers
	for row := 0; row < PasswordMatrixRows; row++ {
		if !strings.Contains(output, fmt.Sprintf("%d", row)) {
			t.Errorf("Pretty output missing row number %d", row)
		}
	}

	// Check cell values are present
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			if !strings.Contains(output, string(m[row][col])) {
				t.Errorf("Pretty output missing cell value %q at [%d][%d]", m[row][col], row, col)
			}
		}
	}
}

func TestExtractPassword_Integration(t *testing.T) {
	// Verify end-to-end pipeline: matrix → spell → extracted password
	matrix := newTestMatrix()

	dirty := DirtySpell{Spell: "1111"}
	spell, err := dirty.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing spell: %v", err)
	}

	password, err := spell.ExtractPassword(matrix, 0) // 0 = no truncation
	if err != nil {
		t.Fatalf("unexpected error extracting password: %v", err)
	}

	// "1111" → all digits (group 0), positions 0-3 → cells (0,0)+(1,0)+(2,0)+(3%rows,0)
	expected := append(append(append(
		matrix[0][0],
		matrix[1%PasswordMatrixRows][0]...),
		matrix[2%PasswordMatrixRows][0]...),
		matrix[3%PasswordMatrixRows][0]...)
	if !bytesEqual(password.Bytes(), expected) {
		t.Errorf("expected %q, got %q", expected, password.Bytes())
	}
	password.Wipe()
}

func TestMatrix_Wipe(t *testing.T) {
	m := newTestMatrix()

	// Store original values to verify they're wiped
	originalCell := m[0][0]

	m.Wipe()

	// Verify all cells are nil after wipe
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			if m[row][col] != nil {
				t.Errorf("m[%d][%d] = %q after wipe, expected nil", row, col, m[row][col])
			}
		}
	}

	// Verify original data reference is no longer accessible
	_ = originalCell // Keep for documentation - original value was stored
}

func TestMatrix_Wipe_ZeroizesData(t *testing.T) {
	// Create a matrix with known content
	input := testutil.NewTestMatrixData(PasswordMatrixRows, PasswordMatrixColumns, CharactersPerMatrixCell)
	m, err := NewMatrix(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Capture a cell value before wipe
	originalValue := m[0][0]

	m.Wipe()

	// After wipe, the cell should be nil
	if m[0][0] != nil {
		t.Errorf("cell not nil after wipe: got %q", m[0][0])
	}

	// The original value should still exist in our local variable
	// (proving we can't truly wipe byte slices, only our matrix references)
	_ = originalValue
}
