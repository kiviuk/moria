package app

import (
	"strings"
	"testing"
)

func TestGenerateRandomString_Length(t *testing.T) {
	// Verify generated string matches requested length
	s, err := GenerateRandomString(300)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) != 300 {
		t.Errorf("expected length 300, got %d", len(s))
	}
}

func TestGenerateRandomString_Charset(t *testing.T) {
	// Verify all characters in generated string are from the allowed pool
	pool := AllowedLetters + AllowedNumbers + AllowedSpecialChars + AllowedSpace
	s, err := GenerateRandomString(1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, r := range s {
		if !strings.ContainsRune(pool, r) {
			t.Errorf("char %q at %d not in allowed pool", r, i)
		}
	}
}

func TestGenerateRandomString_NonDeterministic(t *testing.T) {
	// Verify two consecutive calls produce different strings
	s1, err := GenerateRandomString(300)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s2, err := GenerateRandomString(300)
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
	// Verify Cell returns correct values for valid indices
	m := newTestMatrix()
	tests := []struct {
		row, col int
		expected string
	}{
		{0, 0, "00a"},
		{0, 9, "09j"},
		{5, 3, "53b"},
		{9, 9, "99v"},
	}
	for _, tt := range tests {
		got, err := m.Cell(tt.row, tt.col)
		if err != nil {
			t.Errorf("Cell(%d, %d) unexpected error: %v", tt.row, tt.col, err)
		}
		if got != tt.expected {
			t.Errorf("Cell(%d, %d) = %q, expected %q", tt.row, tt.col, got, tt.expected)
		}
	}
}

func TestMatrix_Cell_OutOfRange(t *testing.T) {
	// Verify Cell returns error for out-of-bounds indices
	m := newTestMatrix()
	tests := []struct {
		row, col int
	}{
		{-1, 0},
		{10, 0},
		{0, -1},
		{0, 10},
		{99, 99},
	}
	for _, tt := range tests {
		_, err := m.Cell(tt.row, tt.col)
		if err == nil {
			t.Errorf("Cell(%d, %d) expected error, got nil", tt.row, tt.col)
		}
	}
}

// newTestMatrix returns a static 10×10 test matrix for use in generator tests.
// Each cell contains a 3-character string in the format "{row}{col}{filler}".
//
// Matrix layout:
//
//	00a   01b   02c   03d   04e   05f   06g   07h   08i   09j
//	10k   11l   12m   13n   14o   15p   16q   17r   18s   19t
//	20u   21v   22w   23x   24y   25z   26a   27b   28c   29d
//	30e   31f   32g   33h   34i   35j   36k   37l   38m   39n
//	40o   41p   42q   43r   44s   45t   46u   47v   48w   49x
//	50y   51z   52a   53b   54c   55d   56e   57f   58g   59h
//	60i   61j   62k   63l   64m   65n   66o   67p   68q   69r
//	70s   71t   72u   73v   74w   75x   76y   77z   78a   79b
//	80c   81d   82e   83f   84g   85h   86i   87j   88k   89l
//	90m   91n   92o   93p   94q   95r   96s   97t   98u   99v
func newTestMatrix() Matrix {
	var m Matrix
	idx := 0
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			filler := rune('a' + rune(idx%26))
			m[row][col] = string(rune('0'+row)) + string(rune('0'+col)) + string(filler)
			idx++
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
	// Verify pattern-based cell values follow the "{row}{col}{filler}" format
	m := newTestMatrix()
	idx := 0
	for row := 0; row < PasswordMatrixRows; row++ {
		for col := 0; col < PasswordMatrixColumns; col++ {
			filler := rune('a' + rune(idx%26))
			expected := string(rune('0'+row)) + string(rune('0'+col)) + string(filler)
			if m[row][col] != expected {
				t.Errorf("m[%d][%d] = %q, expected %q", row, col, m[row][col], expected)
			}
			idx++
		}
	}
}
