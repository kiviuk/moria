package app

import (
	"fmt"
	"strings"

	"github.com/awnumar/memguard"
)

// MagicLetter represents a single character from a validated spell, paired with
// its zero-based position. It is the intermediate form between a parsed spell
// and a resolved matrix query.
type MagicLetter struct {
	// Letter is the single character from the spell.
	Letter string
	// LetterPosition is the zero-based index of this character in the spell.
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

// MagicSpell holds a validated spell string that has passed all character checks.
// It is safe to use for password extraction.
type MagicSpell struct {
	// Spell is the validated spell string containing only allowed characters.
	Spell string
}

// DirtySpell holds an untrusted spell string that has not yet been validated.
// Call Parse() to validate and convert it to a MagicSpell.
type DirtySpell struct {
	// Spell is the raw, unvalidated spell string.
	Spell string
}

// ParseError describes a single invalid character found during spell parsing.
type ParseError struct {
	// Char is the invalid character that was rejected.
	Char string
	// Position is the zero-based index of the invalid character in the spell.
	Position int
}

// Errors is a collection of ParseError values accumulated during spell parsing.
// It implements the error interface and reports all invalid characters at once.
type Errors []ParseError

// Error returns a formatted string listing all invalid characters and their positions.
func (e Errors) Error() string {
	parts := make([]string, len(e))
	for i, pe := range e {
		parts[i] = fmt.Sprintf("%q at %d", pe.Char, pe.Position)
	}
	return "invalid chars: " + strings.Join(parts, ", ")
}

// IsAllowedSpellChar returns true if the rune is valid for spell input.
// This includes lowercase/uppercase letters, digits, space, and special characters.
func IsAllowedSpellChar(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case strings.ContainsRune(AllowedSpecialChars, r):
		return true
	default:
		return false
	}
}

// Parse validates the spell string, rejecting any characters outside the allowed
// set (letters, digits, space, and permitted special characters). All errors are
// accumulated and returned together rather than failing on the first invalid character.
// Spells exceeding MaxSpellLength are rejected.
//
// Returns MagicSpell on success, or Errors containing all ParseError values on failure.
func (d DirtySpell) Parse() (MagicSpell, error) {
	if d.Spell == "" {
		return MagicSpell{}, fmt.Errorf("spell cannot be empty")
	}
	if len(d.Spell) > MaxSpellLength {
		return MagicSpell{}, fmt.Errorf("spell exceeds maximum length of %d characters", MaxSpellLength)
	}
	var errs Errors
	for i, r := range d.Spell {
		s := string(r)
		if s == "" {
			continue
		}
		if !IsAllowedSpellChar(r) {
			errs = append(errs, ParseError{Char: s, Position: i})
		}
	}
	if len(errs) > 0 {
		return MagicSpell{}, errs
	}
	return MagicSpell{Spell: d.Spell}, nil
}

// LetterGroup returns the column group number for a given letter.
// Letters A-C map to 1, D-F to 2, and so on through X-Z to 9.
// Non-letter characters return 0 (the non-letter column).
// The function is case-insensitive: both 'a' and 'A' return group 1.
func LetterGroup(letter string) int {
	if letter == "" {
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

// MagicLetters converts the spell into a slice of MagicLetter values,
// one for each character, preserving order and position.
func (m MagicSpell) MagicLetters() []MagicLetter {
	var letters []MagicLetter = make([]MagicLetter, len(m.Spell))
	for i, r := range m.Spell {
		letters[i] = MagicLetter{Letter: string(r), LetterPosition: i}
	}
	return letters
}

// ModN returns value modulo n (negative n not to be expected)
func ModN(value, n int) int {
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

// ExtractPassword generates the final password by reading cells from the matrix
// along the path defined by the spell. Each character in the spell contributes
// CharactersPerMatrixCell characters to the output.
// If maxLen > 0, the password is truncated to at most maxLen characters.
// Returns a SecureBytes that can be securely wiped when no longer needed.
func (m MagicSpell) ExtractPassword(matrix Matrix, maxLen int) (*SecureBytes, error) {
	letters := m.MagicLetters()

	// Pre-calculate capacity to avoid reallocations
	capacity := len(letters) * CharactersPerMatrixCell
	if maxLen > 0 && maxLen < capacity {
		capacity = maxLen
	}

	password := make([]byte, 0, capacity)
	currentLen := 0

	for _, l := range letters {
		query := l.Query()
		cell, err := matrix.Cell(query)
		if err != nil {
			memguard.WipeBytes(password)
			return nil, err
		}

		// Check if we need to truncate this cell
		if maxLen > 0 && currentLen+len(cell) > maxLen {
			// Only take what fits to reach maxLen
			remaining := maxLen - currentLen
			password = append(password, cell[:remaining]...)
			break
		}

		password = append(password, cell...)
		currentLen += len(cell)
	}

	return NewSecureBytes(password), nil
}
