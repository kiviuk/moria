package app

import "testing"

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
	// Verify MapModN wraps position using MatrixN while preserving letter and group
	tuple := LetterTuple{Letter: "x", LetterPosition: 17, LetterGroup: 3}
	result := tuple.MapModN()

	if result.LetterPosition != ModN(17, MatrixN) {
		t.Errorf("expected position 7, got %d", result.LetterPosition)
	}
	if result.Letter != "x" {
		t.Errorf("expected letter 'x', got %q", result.Letter)
	}
	if result.LetterGroup != 3 {
		t.Errorf("expected group 3, got %d", result.LetterGroup)
	}
}

func TestMagicSpell_LetterTuples_MapModN(t *testing.T) {
	// Verify full pipeline: LetterTuples then MapModN maps 10 letters to positions 0-9
	spell := MagicSpell{Spell: "abcdefghij"}
	tuples := spell.LetterTuples()

	result := make([]LetterTuple, len(tuples))
	for i, tuple := range tuples {
		result[i] = tuple.MapModN()
	}

	expected := []LetterTuple{
		{Letter: "a", LetterPosition: 0, LetterGroup: 1},
		{Letter: "b", LetterPosition: 1, LetterGroup: 1},
		{Letter: "c", LetterPosition: 2, LetterGroup: 1},
		{Letter: "d", LetterPosition: 3, LetterGroup: 2},
		{Letter: "e", LetterPosition: 4, LetterGroup: 2},
		{Letter: "f", LetterPosition: 5, LetterGroup: 2},
		{Letter: "g", LetterPosition: 6, LetterGroup: 3},
		{Letter: "h", LetterPosition: 7, LetterGroup: 3},
		{Letter: "i", LetterPosition: 8, LetterGroup: 3},
		{Letter: "j", LetterPosition: 9, LetterGroup: 4},
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d tuples, got %d", len(expected), len(result))
	}

	for i, exp := range expected {
		if result[i].Letter != exp.Letter || result[i].LetterPosition != exp.LetterPosition || result[i].LetterGroup != exp.LetterGroup {
			t.Errorf("index %d: expected (letter=%q, pos=%d, group=%d), got (letter=%q, pos=%d, group=%d)",
				i, exp.Letter, exp.LetterPosition, exp.LetterGroup, result[i].Letter, result[i].LetterPosition, result[i].LetterGroup)
		}
	}
}

func TestMagicSpell_LetterTuples_MapModN_Wraps(t *testing.T) {
	// Verify positions wrap correctly beyond MatrixN and groups remain unchanged
	spell := MagicSpell{Spell: "abcdefghijklmno"}
	tuples := spell.LetterTuples()

	result := make([]LetterTuple, len(tuples))
	for i, tuple := range tuples {
		result[i] = tuple.MapModN()
	}

	if result[10].LetterPosition != 0 {
		t.Errorf("expected position 0 for 'k', got %d", result[10].LetterPosition)
	}
	if result[14].LetterPosition != 4 {
		t.Errorf("expected position 4 for 'o', got %d", result[14].LetterPosition)
	}

	if result[10].LetterGroup != 4 {
		t.Errorf("expected group 4 for 'k', got %d", result[10].LetterGroup)
	}
	if result[14].LetterGroup != 5 {
		t.Errorf("expected group 5 for 'o', got %d", result[14].LetterGroup)
	}
}

func TestMagicSpell_LetterTuples_MapModN_DigitsWrap(t *testing.T) {
	// Verify digit string wraps: 11th character maps back to position 0
	spell := MagicSpell{Spell: "12345678900"}
	tuples := spell.LetterTuples()

	result := make([]LetterTuple, len(tuples))
	for i, tuple := range tuples {
		result[i] = tuple.MapModN()
	}

	if result[10].LetterPosition != 0 {
		t.Errorf("expected last '0' at position 0, got %d", result[10].LetterPosition)
	}
	if result[9].LetterPosition != 9 {
		t.Errorf("expected first '0' at position 9, got %d", result[9].LetterPosition)
	}
}

func TestMagicSpell_LetterTuples_WithGroup(t *testing.T) {
	// Verify alphabet-based grouping: A-C→1, D-F→2, G-I→3
	spell := MagicSpell{Spell: "ABCDEFGHI"}
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
