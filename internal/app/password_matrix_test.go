package app

import (
	"strings"
	"testing"
)

func TestGenerateRandomString_Length(t *testing.T) {
	// Verify generated string matches requested length
	s, err := GenerateRandomString(300, MasterPasswordChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) != 300 {
		t.Errorf("expected length 300, got %d", len(s))
	}
}

func TestGenerateRandomString_Charset(t *testing.T) {
	// Verify all characters in generated string are from the allowed pool
	s, err := GenerateRandomString(1000, MasterPasswordChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, r := range s {
		if !strings.ContainsRune(MasterPasswordChars, r) {
			t.Errorf("char %q at %d not in allowed pool", r, i)
		}
	}
}

func TestGenerateRandomString_NonDeterministic(t *testing.T) {
	// Verify two consecutive calls produce different strings
	s1, err := GenerateRandomString(300, MasterPasswordChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s2, err := GenerateRandomString(300, MasterPasswordChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s1 == s2 {
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
	input := "00a01b02c03d04e05f06g07h08i09j10k11l12m13n14o15p16q17r18s19t20u21v22w23x24y25z26a27b28c29d30e31f32g33h34i35j36k37l38m39n40o41p42q43r44s45t46u47v48w49x50y51z52a53b54c55d56e57f58g59h60i61j62k63l64m65n66o67p68q69r70s71t72u73v74w75x76y77z78a79b80c81d82e83f84g85h86i87j88k89l90m91n92o93p94q95r96s97t98u99v"
	m, err := NewMatrix(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	idx := 0
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			start := (row*PasswordMatrixColumns + col) * CharactersPerMatrixCell
			expected := input[start : start+CharactersPerMatrixCell]
			if m[row][col] != expected {
				t.Errorf("m[%d][%d] = %q, expected %q", row, col, m[row][col], expected)
			}
			idx++
		}
	}
	_ = idx
}

func TestMatrix_Cell(t *testing.T) {
	// Verify Cell returns correct values for valid query letters
	m := newTestMatrix()
	tests := []struct {
		query    QueryLetter
		expected string
	}{
		{QueryLetter{MatrixRow: 0, LetterGroup: 0}, "00|"},
		{QueryLetter{MatrixRow: 0, LetterGroup: 9}, "09|"},
		{QueryLetter{MatrixRow: 5, LetterGroup: 3}, "53|"},
		{QueryLetter{MatrixRow: 9, LetterGroup: 9}, "99|"},
	}
	for _, tt := range tests {
		got, err := m.Cell(tt.query)
		if err != nil {
			t.Errorf("Cell(%+v) unexpected error: %v", tt.query, err)
		}
		if got != tt.expected {
			t.Errorf("Cell(%+v) = %q, expected %q", tt.query, got, tt.expected)
		}
	}
}

func TestMatrix_Cell_OutOfRangeRow(t *testing.T) {
	// Verify Cell returns error for out-of-bounds matrix row in query letter
	m := newTestMatrix()
	tests := []int{-1, 10, 99}
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
	tests := []int{-1, 10, 99}
	for _, col := range tests {
		query := QueryLetter{MatrixRow: 0, LetterGroup: col}
		_, err := m.Cell(query)
		if err == nil {
			t.Errorf("Cell with col %d expected error, got nil", col)
		}
	}
}

// newTestMatrix returns a static 10×10 test matrix for use in generator tests.
// Each cell contains a 3-character string in the format "{row}{col}|".
//
// Matrix layout:
//
//	00|   01|   02|   03|   04|   05|   06|   07|   08|   09|
//	10|   11|   12|   13|   14|   15|   16|   17|   18|   19|
//	20|   21|   22|   23|   24|   25|   26|   27|   28|   29|
//	30|   31|   32|   33|   34|   35|   36|   37|   38|   39|
//	40|   41|   42|   43|   44|   45|   46|   47|   48|   49|
//	50|   51|   52|   53|   54|   55|   56|   57|   58|   59|
//	60|   61|   62|   63|   64|   65|   66|   67|   68|   69|
//	70|   71|   72|   73|   74|   75|   76|   77|   78|   79|
//	80|   81|   82|   83|   84|   85|   86|   87|   88|   89|
//	90|   91|   92|   93|   94|   95|   96|   97|   98|   99|
func newTestMatrix() Matrix {
	var m Matrix
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			m[row][col] = string(rune('0'+row)) + string(rune('0'+col)) + "|"
		}
	}
	return m
}

func TestMatrix_Dimensions(t *testing.T) {
	// Verify the static test matrix has the correct dimensions based on constants
	m := newTestMatrix()
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
	// Verify pattern-based cell values follow the "{row}{col}|" format
	m := newTestMatrix()
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			expected := string(rune('0'+row)) + string(rune('0'+col)) + "|"
			if m[row][col] != expected {
				t.Errorf("m[%d][%d] = %q, expected %q", row, col, m[row][col], expected)
			}
		}
	}
}
