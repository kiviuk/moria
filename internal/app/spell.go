package app

import (
	"fmt"
	"strings"
)

type MagicLetter struct {
	Letter         string
	LetterPosition int
}

// QueryLetter is a MagicLetter whose position has been resolved to a valid matrix row.
// MatrixRow is the wrapped row index (0-9), not the original spell position.
// This type is used to query the matrix for its password fragment.
type QueryLetter struct {
	Letter      string
	MatrixRow   int
	LetterGroup int
}

type MagicSpell struct {
	Spell string
}

type DirtySpell struct {
	Spell string
}

type ParseError struct {
	Char     string
	Position int
}

type Errors []ParseError

func (e Errors) Error() string {
	parts := make([]string, len(e))
	for i, pe := range e {
		parts[i] = fmt.Sprintf("%q at %d", pe.Char, pe.Position)
	}
	return "invalid chars: " + strings.Join(parts, ", ")
}

var allowedPattern = "[" + AllowedLetters + AllowedNumbers + AllowedSpecialChars + AllowedSpace + "]"

func (d DirtySpell) Parse() (MagicSpell, error) {
	if len(d.Spell) == 0 {
		return MagicSpell{}, fmt.Errorf("spell cannot be empty")
	}
	var errs Errors
	for i, r := range d.Spell {
		s := string(r)
		if s == "" {
			continue
		}
		matched := false
		if r >= 'a' && r <= 'z' {
			matched = true
		} else if r >= 'A' && r <= 'Z' {
			matched = true
		} else if r >= '0' && r <= '9' {
			matched = true
		} else if r == ' ' {
			matched = true
		} else if strings.ContainsRune(AllowedSpecialChars, r) {
			matched = true
		}
		if !matched {
			errs = append(errs, ParseError{Char: s, Position: i})
		}
	}
	if len(errs) > 0 {
		return MagicSpell{}, errs
	}
	return MagicSpell{Spell: d.Spell}, nil
}

func LetterGroup(letter string) int {
	if len(letter) == 0 {
		return 0
	}
	r := rune(letter[0])
	selected := rune(0)
	if r >= 'A' && r <= 'Z' {
		selected = 'A'
	} else if r >= 'a' && r <= 'z' {
		selected = 'a'
	}
	if selected == 0 {
		return 0
	}
	return int(r-selected)/CharactersPerMatrixCell + 1
}

func (m MagicSpell) MagicLetters() []MagicLetter {
	letters := make([]MagicLetter, len(m.Spell))
	for i, r := range m.Spell {
		letters[i] = MagicLetter{Letter: string(r), LetterPosition: i}
	}
	return letters
}

func ModN(value int, n int) int {
	return value % n
}

// Query transforms a MagicLetter into a QueryLetter with resolved matrix coordinates.
// Each character in the spell acts as a pointer into the password matrix.
// The spell position determines the matrix row (wrapped via modulo to fit PasswordMatrixRows rows).
// Uppercase letters shift the row by PasswordMatrixRows/2, making case significant.
// Dividing by 2 ensures zero overlap: every uppercase letter lands on a row
// that no lowercase letter can reach at the same position.
// The letter determines the column via LetterGroup.
// This creates a deterministic path through the matrix: the same spell always
// reads the same cells, producing the same password from the same matrix.
func (m MagicLetter) Query() QueryLetter {
	row := ModN(m.LetterPosition, PasswordMatrixRows)
	if m.Letter >= "A" && m.Letter <= "Z" {
		row = ModN(m.LetterPosition+PasswordMatrixRows/2, PasswordMatrixRows)
	}
	return QueryLetter{
		Letter:      m.Letter,
		MatrixRow:   row,
		LetterGroup: LetterGroup(m.Letter),
	}
}

func (m MagicSpell) ExtractPassword(matrix Matrix) (string, error) {
	letters := m.MagicLetters()
	var password strings.Builder
	for _, l := range letters {
		query := l.Query()
		cell, err := matrix.Cell(query)
		if err != nil {
			return "", err
		}
		password.WriteString(cell)
	}
	return password.String(), nil
}
