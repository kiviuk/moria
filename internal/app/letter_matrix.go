package app

import (
	"fmt"
	"strings"
)

type LetterTuple struct {
	Letter         string
	LetterPosition int
	LetterGroup    int
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
	return int(r-selected)/GroupSize + 1
}

func (m MagicSpell) LetterTuples() []LetterTuple {
	tuples := make([]LetterTuple, len(m.Spell))
	for i, r := range m.Spell {
		tuples[i] = LetterTuple{Letter: string(r), LetterPosition: i, LetterGroup: LetterGroup(string(r))}
	}
	return tuples
}

func ModN(value int, n int) int {
	return value % n
}

func (m LetterTuple) MapModN() LetterTuple {
	return LetterTuple{
		Letter:         m.Letter,
		LetterPosition: ModN(m.LetterPosition, MatrixN),
		LetterGroup:    m.LetterGroup,
	}
}
