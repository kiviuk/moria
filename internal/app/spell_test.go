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
	// Verify end-to-end: DirtySpell.Parse().MagicLetters() works
	dirty := DirtySpell{Spell: "abc"}
	spell, err := dirty.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	letters := spell.MagicLetters()
	if len(letters) != 3 {
		t.Fatalf("expected 3 letters, got %d", len(letters))
	}
	if letters[0].Letter != "a" || letters[1].Letter != "b" || letters[2].Letter != "c" {
		t.Errorf("unexpected letters: %v", letters)
	}
}

func TestMagicSpell_Length(t *testing.T) {
	// Verify MagicSpell.Spell length is computed correctly
	spell := MagicSpell{Spell: "abracadabra"}
	if got := len(spell.Spell); got != 11 {
		t.Errorf("expected length 11, got %d", got)
	}
}

func TestMagicSpell_MagicLetters(t *testing.T) {
	// Verify MagicLetters builds correct position and letter for each character
	spell := MagicSpell{Spell: "hello"}
	letters := spell.MagicLetters()

	expected := []MagicLetter{
		{Letter: "h", LetterPosition: 0},
		{Letter: "e", LetterPosition: 1},
		{Letter: "l", LetterPosition: 2},
		{Letter: "l", LetterPosition: 3},
		{Letter: "o", LetterPosition: 4},
	}

	if len(letters) != len(expected) {
		t.Fatalf("expected %d letters, got %d", len(expected), len(letters))
	}

	for i, exp := range expected {
		if letters[i].LetterPosition != exp.LetterPosition || letters[i].Letter != exp.Letter {
			t.Errorf("index %d: expected (pos=%d, letter=%q), got (pos=%d, letter=%q)",
				i, exp.LetterPosition, exp.Letter, letters[i].LetterPosition, letters[i].Letter)
		}
	}
}

func TestLetterGroup_CaseInsensitive(t *testing.T) {
	// Verify lowercase and uppercase letters map to the same group
	tests := []struct {
		lower, upper string
	}{
		{"a", "A"}, {"b", "B"}, {"c", "C"},
		{"d", "D"}, {"e", "E"}, {"f", "F"},
		{"g", "G"}, {"h", "H"}, {"i", "I"},
		{"j", "J"}, {"k", "K"}, {"l", "L"},
		{"m", "M"}, {"n", "N"}, {"o", "O"},
		{"p", "P"}, {"q", "Q"}, {"r", "R"},
		{"s", "S"}, {"t", "T"}, {"u", "U"},
		{"v", "V"}, {"w", "W"}, {"x", "X"},
		{"y", "Y"}, {"z", "Z"},
	}

	for _, tt := range tests {
		lg := LetterGroup(tt.lower)
		ug := LetterGroup(tt.upper)
		if lg != ug {
			t.Errorf("LetterGroup(%q)=%d, LetterGroup(%q)=%d, expected equal",
				tt.lower, lg, tt.upper, ug)
		}
		if lg == 0 {
			t.Errorf("LetterGroup(%q)=%d, expected non-zero", tt.lower, lg)
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

func TestQuery(t *testing.T) {
	// Verify Query wraps position to MatrixRow while preserving letter and group
	letter := MagicLetter{Letter: "x", LetterPosition: 17}
	result := letter.Query()

	expectedRow := ModN(17, PasswordMatrixRows)

	if result.MatrixRow != expectedRow {
		t.Errorf("expected row 7, got %d", result.MatrixRow)
	}
	if result.Letter != "x" {
		t.Errorf("expected letter 'x', got %q", result.Letter)
	}
	if result.LetterGroup != LetterGroup("x") {
		t.Errorf("expected group %d, got %d", LetterGroup("x"), result.LetterGroup)
	}
}

func TestMagicSpell_MagicLetters_Query(t *testing.T) {
	// Verify full pipeline: MagicLetters then Query maps 10 letters to matrix rows 0-9
	spell := MagicSpell{Spell: "abcdefghij"}
	letters := spell.MagicLetters()

	result := make([]QueryLetter, len(letters))
	for i, l := range letters {
		result[i] = l.Query()
	}

	if len(result) != len(letters) {
		t.Fatalf("expected %d letters, got %d", len(letters), len(result))
	}

	for i, l := range letters {
		if result[i].Letter != l.Letter {
			t.Errorf("index %d: expected letter %q, got %q", i, l.Letter, result[i].Letter)
		}
		if result[i].MatrixRow != ModN(l.LetterPosition, PasswordMatrixRows) {
			t.Errorf("index %d: expected row %d, got %d", i, ModN(l.LetterPosition, PasswordMatrixRows), result[i].MatrixRow)
		}
		if result[i].LetterGroup != LetterGroup(l.Letter) {
			t.Errorf("index %d: expected group %d, got %d", i, LetterGroup(l.Letter), result[i].LetterGroup)
		}
	}
}

func TestMagicSpell_MagicLetters_Query_Wraps(t *testing.T) {
	// Verify positions wrap correctly beyond PasswordMatrixRows and groups remain unchanged
	spell := MagicSpell{Spell: "abcdefghijklmno"}
	letters := spell.MagicLetters()

	result := make([]QueryLetter, len(letters))
	for i, l := range letters {
		result[i] = l.Query()
	}

	if result[10].MatrixRow != 0 {
		t.Errorf("expected row 0 for 'k', got %d", result[10].MatrixRow)
	}
	if result[14].MatrixRow != 4 {
		t.Errorf("expected row 4 for 'o', got %d", result[14].MatrixRow)
	}

	if result[10].LetterGroup != LetterGroup("k") {
		t.Errorf("expected group %d for 'k', got %d", LetterGroup("k"), result[10].LetterGroup)
	}
	if result[14].LetterGroup != LetterGroup("o") {
		t.Errorf("expected group %d for 'o', got %d", LetterGroup("o"), result[14].LetterGroup)
	}
}

func TestMagicSpell_MagicLetters_Query_DigitsWrap(t *testing.T) {
	// Verify digit string wraps: 11th character maps back to row 0
	spell := MagicSpell{Spell: "12345678900"}
	letters := spell.MagicLetters()

	result := make([]QueryLetter, len(letters))
	for i, l := range letters {
		result[i] = l.Query()
	}

	if result[10].MatrixRow != 0 {
		t.Errorf("expected last '0' at row 0, got %d", result[10].MatrixRow)
	}
	if result[9].MatrixRow != 9 {
		t.Errorf("expected first '0' at row 9, got %d", result[9].MatrixRow)
	}
}

func TestQueryLetter_Grouping(t *testing.T) {
	// Verify alphabet-based grouping matches LetterGroup() for each letter
	spell := MagicSpell{Spell: "ABCDEFGHIJKL"}
	letters := spell.MagicLetters()

	result := make([]QueryLetter, len(letters))
	for i, l := range letters {
		result[i] = l.Query()
	}

	for i, l := range letters {
		if result[i].Letter != l.Letter {
			t.Errorf("index %d: expected letter %q, got %q", i, l.Letter, result[i].Letter)
		}
		if result[i].MatrixRow != ModN(l.LetterPosition, PasswordMatrixRows) {
			t.Errorf("index %d: expected row %d, got %d", i, ModN(l.LetterPosition, PasswordMatrixRows), result[i].MatrixRow)
		}
		if result[i].LetterGroup != LetterGroup(l.Letter) {
			t.Errorf("index %d: expected group %d, got %d", i, LetterGroup(l.Letter), result[i].LetterGroup)
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
	expected := matrix[0][0] + matrix[1][0] + matrix[2][0] + matrix[3][0]
	if password != expected {
		t.Errorf("expected %q, got %q", expected, password)
	}
}

func TestMagicSpell_ExtractPassword_OnePerGroup(t *testing.T) {
	// Verify one letter from each group extracts cells across different columns
	matrix := newTestMatrix()
	spell := MagicSpell{Spell: "adgjmpsvy"}
	password, err := spell.ExtractPassword(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Build expected password by computing each letter's group
	letters := spell.MagicLetters()
	var expected strings.Builder
	for _, l := range letters {
		q := l.Query()
		expected.WriteString(matrix[q.MatrixRow][q.LetterGroup])
	}
	if password != expected.String() {
		t.Errorf("expected %q, got %q", expected.String(), password)
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
	expected := matrix[0][0] + matrix[1][0] + matrix[2][0] + matrix[3][0]
	if password != expected {
		t.Errorf("expected %q, got %q", expected, password)
	}
}
