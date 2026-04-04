package app

import "testing"

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
