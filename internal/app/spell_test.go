package app

import (
	"strings"
	"testing"
)

func TestDirtySpell_Parse_Valid(t *testing.T) {
	// Verify valid spell with letters, digits, specials, space passes
	dirty := DirtySpell{Spell: "hello World123!@#"}
	spell, err := dirty.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spell.Spell != "hello World123!@#" {
		t.Errorf("expected spell %q, got %q", "hello World123!@#", spell.Spell)
	}
}

func TestDirtySpell_Parse_Empty(t *testing.T) {
	// Verify empty spell is rejected since it cannot produce a password
	dirty := DirtySpell{Spell: ""}
	_, err := dirty.Parse()
	if err == nil {
		t.Fatal("expected error for empty spell, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error about empty spell, got: %v", err)
	}
}

func TestDirtySpell_Parse_RejectsNewline(t *testing.T) {
	// Verify newline character is rejected as invalid
	dirty := DirtySpell{Spell: "he\nllo"}
	_, err := dirty.Parse()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), `\n`) && !strings.Contains(err.Error(), "\n") {
		t.Errorf("expected error about newline, got: %v", err)
	}
}

func TestDirtySpell_Parse_RejectsTab(t *testing.T) {
	// Verify tab character is rejected as invalid
	dirty := DirtySpell{Spell: "he\tllo"}
	_, err := dirty.Parse()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), `\t`) && !strings.Contains(err.Error(), "\t") {
		t.Errorf("expected error about tab, got: %v", err)
	}
}

func TestDirtySpell_Parse_RejectsUnicode(t *testing.T) {
	// Verify unicode characters are rejected as invalid
	dirty := DirtySpell{Spell: "héllo"}
	_, err := dirty.Parse()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDirtySpell_Parse_MultipleErrors(t *testing.T) {
	// Verify all invalid characters are accumulated, not just the first
	dirty := DirtySpell{Spell: "a\nb\tc\rd"}
	_, err := dirty.Parse()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errs := err.(Errors)
	if len(errs) != 3 {
		t.Errorf("expected 3 errors, got %d: %v", len(errs), errs)
	}
}

func TestDirtySpell_Parse_Integration(t *testing.T) {
	// Verify end-to-end: DirtySpell.Parse().LetterTuples() works
	dirty := DirtySpell{Spell: "abc"}
	spell, err := dirty.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tuples := spell.LetterTuples()
	if len(tuples) != 3 {
		t.Fatalf("expected 3 tuples, got %d", len(tuples))
	}
	if tuples[0].Letter != "a" || tuples[1].Letter != "b" || tuples[2].Letter != "c" {
		t.Errorf("unexpected letters: %v", tuples)
	}
}

func TestMagicSpell_Length(t *testing.T) {
	// Verify MagicSpell.Spell length is computed correctly
	spell := MagicSpell{Spell: "abracadabra"}
	if got := len(spell.Spell); got != 11 {
		t.Errorf("expected length 11, got %d", got)
	}
}

func TestMagicSpell_LetterTuples(t *testing.T) {
	// Verify LetterTuples builds correct position, letter, and group for each character
	spell := MagicSpell{Spell: "hello"}
	tuples := spell.LetterTuples()

	expected := []LetterTuple{
		{Letter: "h", LetterPosition: 0, LetterGroup: 3},
		{Letter: "e", LetterPosition: 1, LetterGroup: 2},
		{Letter: "l", LetterPosition: 2, LetterGroup: 4},
		{Letter: "l", LetterPosition: 3, LetterGroup: 4},
		{Letter: "o", LetterPosition: 4, LetterGroup: 5},
	}

	if len(tuples) != len(expected) {
		t.Fatalf("expected %d tuples, got %d", len(expected), len(tuples))
	}

	for i, exp := range expected {
		if tuples[i].LetterPosition != exp.LetterPosition || tuples[i].Letter != exp.Letter || tuples[i].LetterGroup != exp.LetterGroup {
			t.Errorf("index %d: expected (pos=%d, letter=%q, group=%d), got (pos=%d, letter=%q, group=%d)",
				i, exp.LetterPosition, exp.Letter, exp.LetterGroup, tuples[i].LetterPosition, tuples[i].Letter, tuples[i].LetterGroup)
		}
	}
}

func TestLetterGroup_CaseInsensitive(t *testing.T) {
	// Verify lowercase and uppercase letters map to the same group
	tests := []struct {
		lower, upper string
		group        int
	}{
		{"a", "A", 1}, {"b", "B", 1}, {"c", "C", 1},
		{"d", "D", 2}, {"e", "E", 2}, {"f", "F", 2},
		{"g", "G", 3}, {"h", "H", 3}, {"i", "I", 3},
		{"j", "J", 4}, {"k", "K", 4}, {"l", "L", 4},
		{"m", "M", 5}, {"n", "N", 5}, {"o", "O", 5},
		{"p", "P", 6}, {"q", "Q", 6}, {"r", "R", 6},
		{"s", "S", 7}, {"t", "T", 7}, {"u", "U", 7},
		{"v", "V", 8}, {"w", "W", 8}, {"x", "X", 8},
		{"y", "Y", 9}, {"z", "Z", 9},
	}

	for _, tt := range tests {
		lg := LetterGroup(tt.lower)
		ug := LetterGroup(tt.upper)
		if lg != tt.group || ug != tt.group {
			t.Errorf("LetterGroup(%q)=%d, LetterGroup(%q)=%d, expected %d",
				tt.lower, lg, tt.upper, ug, tt.group)
		}
	}
}

func TestLetterGroup_Digits(t *testing.T) {
	// Verify digits 0-9 return group 0 since they are not letters
	tests := []struct {
		letter   string
		expected int
	}{
		{"0", 0}, {"1", 0}, {"2", 0}, {"3", 0}, {"4", 0},
		{"5", 0}, {"6", 0}, {"7", 0}, {"8", 0}, {"9", 0},
	}

	for _, tt := range tests {
		if got := LetterGroup(tt.letter); got != tt.expected {
			t.Errorf("LetterGroup(%q) = %d, expected %d", tt.letter, got, tt.expected)
		}
	}
}

func TestLetterGroup_SpecialChars(t *testing.T) {
	// Verify special characters return group 0 since they are not letters
	tests := []struct {
		letter   string
		expected int
	}{
		{"!", 0}, {"@", 0}, {"#", 0}, {"$", 0}, {"%", 0},
		{"^", 0}, {"&", 0}, {"*", 0}, {"(", 0}, {")", 0},
		{"-", 0}, {"_", 0}, {"=", 0}, {"+", 0}, {"[", 0},
		{"]", 0}, {"{", 0}, {"}", 0}, {"|", 0}, {";", 0},
		{":", 0}, {"'", 0}, {"\"", 0}, {",", 0}, {".", 0},
		{"/", 0}, {"?", 0}, {" ", 0}, {"\t", 0}, {"\n", 0},
	}

	for _, tt := range tests {
		if got := LetterGroup(tt.letter); got != tt.expected {
			t.Errorf("LetterGroup(%q) = %d, expected %d", tt.letter, got, tt.expected)
		}
	}
}

func TestModN(t *testing.T) {
	// Verify modulo operation returns correct remainders for various inputs
	tests := []struct {
		value    int
		n        int
		expected int
	}{
		{10, 3, 1},
		{7, 5, 2},
		{0, 4, 0},
		{12, 12, 0},
	}

	for _, tt := range tests {
		if got := ModN(tt.value, tt.n); got != tt.expected {
			t.Errorf("ModN(%d, %d) = %d, expected %d", tt.value, tt.n, got, tt.expected)
		}
	}
}

func TestMapModN(t *testing.T) {
	// Verify MapModN wraps position to MatrixRow while preserving letter and group
	tuple := LetterTuple{Letter: "x", LetterPosition: 17, LetterGroup: 3}
	result := tuple.MapModN()

	if result.MatrixRow != ModN(17, PasswordMatrixRows) {
		t.Errorf("expected row 7, got %d", result.MatrixRow)
	}
	if result.Letter != "x" {
		t.Errorf("expected letter 'x', got %q", result.Letter)
	}
	if result.LetterGroup != 3 {
		t.Errorf("expected group 3, got %d", result.LetterGroup)
	}
}

func TestMagicSpell_LetterTuples_MapModN(t *testing.T) {
	// Verify full pipeline: LetterTuples then MapModN maps 10 letters to matrix rows 0-9
	spell := MagicSpell{Spell: "abcdefghij"}
	tuples := spell.LetterTuples()

	result := make([]ResolvedTuple, len(tuples))
	for i, tuple := range tuples {
		result[i] = tuple.MapModN()
	}

	expected := []ResolvedTuple{
		{Letter: "a", MatrixRow: 0, LetterGroup: 1},
		{Letter: "b", MatrixRow: 1, LetterGroup: 1},
		{Letter: "c", MatrixRow: 2, LetterGroup: 1},
		{Letter: "d", MatrixRow: 3, LetterGroup: 2},
		{Letter: "e", MatrixRow: 4, LetterGroup: 2},
		{Letter: "f", MatrixRow: 5, LetterGroup: 2},
		{Letter: "g", MatrixRow: 6, LetterGroup: 3},
		{Letter: "h", MatrixRow: 7, LetterGroup: 3},
		{Letter: "i", MatrixRow: 8, LetterGroup: 3},
		{Letter: "j", MatrixRow: 9, LetterGroup: 4},
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d tuples, got %d", len(expected), len(result))
	}

	for i, exp := range expected {
		if result[i].Letter != exp.Letter || result[i].MatrixRow != exp.MatrixRow || result[i].LetterGroup != exp.LetterGroup {
			t.Errorf("index %d: expected (letter=%q, row=%d, group=%d), got (letter=%q, row=%d, group=%d)",
				i, exp.Letter, exp.MatrixRow, exp.LetterGroup, result[i].Letter, result[i].MatrixRow, result[i].LetterGroup)
		}
	}
}

func TestMagicSpell_LetterTuples_MapModN_Wraps(t *testing.T) {
	// Verify positions wrap correctly beyond PasswordMatrixRows and groups remain unchanged
	spell := MagicSpell{Spell: "abcdefghijklmno"}
	tuples := spell.LetterTuples()

	result := make([]ResolvedTuple, len(tuples))
	for i, tuple := range tuples {
		result[i] = tuple.MapModN()
	}

	if result[10].MatrixRow != 0 {
		t.Errorf("expected row 0 for 'k', got %d", result[10].MatrixRow)
	}
	if result[14].MatrixRow != 4 {
		t.Errorf("expected row 4 for 'o', got %d", result[14].MatrixRow)
	}

	if result[10].LetterGroup != 4 {
		t.Errorf("expected group 4 for 'k', got %d", result[10].LetterGroup)
	}
	if result[14].LetterGroup != 5 {
		t.Errorf("expected group 5 for 'o', got %d", result[14].LetterGroup)
	}
}

func TestMagicSpell_LetterTuples_MapModN_DigitsWrap(t *testing.T) {
	// Verify digit string wraps: 11th character maps back to row 0
	spell := MagicSpell{Spell: "12345678900"}
	tuples := spell.LetterTuples()

	result := make([]ResolvedTuple, len(tuples))
	for i, tuple := range tuples {
		result[i] = tuple.MapModN()
	}

	if result[10].MatrixRow != 0 {
		t.Errorf("expected last '0' at row 0, got %d", result[10].MatrixRow)
	}
	if result[9].MatrixRow != 9 {
		t.Errorf("expected first '0' at row 9, got %d", result[9].MatrixRow)
	}
}

func TestMagicSpell_LetterTuples_WithGroup(t *testing.T) {
	// Verify alphabet-based grouping: A-C→1, D-F→2, G-I→3
	spell := MagicSpell{Spell: "ABCDEFGHIJKL"}
	tuples := spell.LetterTuples()

	expected := []LetterTuple{
		{Letter: "A", LetterPosition: 0, LetterGroup: 1},
		{Letter: "B", LetterPosition: 1, LetterGroup: 1},
		{Letter: "C", LetterPosition: 2, LetterGroup: 1},
		{Letter: "D", LetterPosition: 3, LetterGroup: 2},
		{Letter: "E", LetterPosition: 4, LetterGroup: 2},
		{Letter: "F", LetterPosition: 5, LetterGroup: 2},
		{Letter: "G", LetterPosition: 6, LetterGroup: 3},
		{Letter: "H", LetterPosition: 7, LetterGroup: 3},
		{Letter: "I", LetterPosition: 8, LetterGroup: 3},
		{Letter: "J", LetterPosition: 9, LetterGroup: 4},
		{Letter: "K", LetterPosition: 10, LetterGroup: 4},
		{Letter: "L", LetterPosition: 11, LetterGroup: 4},
	}

	if len(tuples) != len(expected) {
		t.Fatalf("expected %d tuples, got %d", len(expected), len(tuples))
	}

	for i, exp := range expected {
		if tuples[i].LetterPosition != exp.LetterPosition || tuples[i].Letter != exp.Letter || tuples[i].LetterGroup != exp.LetterGroup {
			t.Errorf("index %d: expected (pos=%d, letter=%q, group=%d), got (pos=%d, letter=%q, group=%d)",
				i, exp.LetterPosition, exp.Letter, exp.LetterGroup, tuples[i].LetterPosition, tuples[i].Letter, tuples[i].LetterGroup)
		}
	}
}

func TestMagicSpell_ExtractPassword_Digits(t *testing.T) {
	// Verify digits map to group 0 and extract correct cells from the test matrix
	matrix := newTestMatrix()
	spell := MagicSpell{Spell: "1111"}
	password, err := spell.ExtractPassword(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if password != "00|10|20|30|" {
		t.Errorf("expected %q, got %q", "00|10|20|30|", password)
	}
}

func TestMagicSpell_ExtractPassword_OnePerGroup(t *testing.T) {
	// Verify one letter from each group extracts cells across all columns 1-9
	matrix := newTestMatrix()
	spell := MagicSpell{Spell: "adgjmpsvy"}
	password, err := spell.ExtractPassword(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if password != "01|12|23|34|45|56|67|78|89|" {
		t.Errorf("expected %q, got %q", "01|12|23|34|45|56|67|78|89|", password)
	}
}

func TestMagicSpell_ExtractPassword_Spaces(t *testing.T) {
	// Verify spaces map to group 0 same as digits, extracting identical cells
	matrix := newTestMatrix()
	spell := MagicSpell{Spell: "    "}
	password, err := spell.ExtractPassword(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if password != "00|10|20|30|" {
		t.Errorf("expected %q, got %q", "00|10|20|30|", password)
	}
}
