package app

import (
"bytes"
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

func TestDirtySpell_Parse_RejectsInvalidChars(t *testing.T) {
// Verify each invalid character class is rejected with an error mentioning the offending character
tests := []struct {
name            string
spell           string
wantErrContains string
}{
{"newline", "he\nllo", `"\n"`},
{"tab", "he\tllo", `"\t"`},
{"unicode", "héllo", `"é"`},
}
for _, tt := range tests {
dirty := DirtySpell{Spell: tt.spell}
_, err := dirty.Parse()
if err == nil {
t.Errorf("%s: expected error, got nil", tt.name)
continue
}
if !strings.Contains(err.Error(), tt.wantErrContains) {
t.Errorf("%s: error %q does not mention invalid char %s", tt.name, err.Error(), tt.wantErrContains)
}
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

func TestLetterGroup_NonLetters(t *testing.T) {
// Verify digits and special characters return group 0
tests := []string{
"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
"!", "@", "#", "$", "%", "^", "&", "*", "(", ")",
"-", "_", "=", "+", "[", "]", "{", "}", "|", ";",
":", "'", "\"", ",", ".", "/", "?", " ", "\t", "\n",
}
for _, s := range tests {
if got := LetterGroup(s); got != 0 {
t.Errorf("LetterGroup(%q) = %d, expected 0", s, got)
}
}
}

func TestModN(t *testing.T) {
// Verify modulo operation returns correct remainders for various inputs
tests := []struct {
value, n, expected int
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
// Verify full pipeline: MagicLetters then Query maps letters to correct rows and groups
spell := MagicSpell{Spell: "abcdefghijkl"}
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

func TestMagicSpell_MagicLetters_Query_Wraps(t *testing.T) {
// Verify positions wrap correctly beyond PasswordMatrixRows for both letters and digits
tests := []struct {
name  string
spell string
checks []struct{ pos int; wantRow int; wantGroup int }
}{
{
name:  "letters wrap at position 10 and 14",
spell: "abcdefghijklmno",
checks: []struct{ pos int; wantRow int; wantGroup int }{
{10, ModN(10, PasswordMatrixRows), LetterGroup("k")},
{14, ModN(14, PasswordMatrixRows), LetterGroup("o")},
},
},
{
name:  "digits wrap at position 10 and 9",
spell: "12345678900",
checks: []struct{ pos int; wantRow int; wantGroup int }{
{9, ModN(9, PasswordMatrixRows), 0},
{10, ModN(10, PasswordMatrixRows), 0},
},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
letters := MagicSpell{Spell: tt.spell}.MagicLetters()
results := make([]QueryLetter, len(letters))
for i, l := range letters {
results[i] = l.Query()
}
for _, c := range tt.checks {
if results[c.pos].MatrixRow != c.wantRow {
t.Errorf("pos %d: want row %d, got %d", c.pos, c.wantRow, results[c.pos].MatrixRow)
}
if results[c.pos].LetterGroup != c.wantGroup {
t.Errorf("pos %d: want group %d, got %d", c.pos, c.wantGroup, results[c.pos].LetterGroup)
}
}
})
}
}

func TestQueryLetter_CaseSensitiveRow(t *testing.T) {
// Verify uppercase letters shift row by PasswordMatrixRows/2
spellLower := MagicSpell{Spell: "abc"}
spellUpper := MagicSpell{Spell: "ABC"}

lowerLetters := spellLower.MagicLetters()
upperLetters := spellUpper.MagicLetters()

for i := 0; i < len(lowerLetters); i++ {
lowerQuery := lowerLetters[i].Query()
upperQuery := upperLetters[i].Query()

expectedShift := ModN(lowerLetters[i].LetterPosition+PasswordMatrixRows/2, PasswordMatrixRows)
if upperQuery.MatrixRow != expectedShift {
t.Errorf("index %d: uppercase row %d, expected %d", i, upperQuery.MatrixRow, expectedShift)
}
if lowerQuery.MatrixRow == upperQuery.MatrixRow {
t.Errorf("index %d: lowercase and uppercase produced same row %d", i, lowerQuery.MatrixRow)
}
if lowerQuery.LetterGroup != upperQuery.LetterGroup {
t.Errorf("index %d: letter group should be same for case, got %d vs %d", i, lowerQuery.LetterGroup, upperQuery.LetterGroup)
}
}
}

func TestMatrix_ExtractPassword_Digits(t *testing.T) {
// Verify digits map to group 0 and extract correct cells from the test matrix
matrix := newTestMatrix()
dirty := DirtySpell{Spell: "1111"}
spell, err := dirty.Parse()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
password, err := matrix.ExtractPassword(spell, 0)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
defer password.Wipe()
expected := append(append(append(
matrix[0][0],
matrix[1][0]...),
matrix[2][0]...),
matrix[3%PasswordMatrixRows][0]...)
if !bytes.Equal(password.Bytes(), expected) {
t.Errorf("expected %q, got %q", expected, password.Bytes())
}
}

func TestMatrix_ExtractPassword_OnePerGroup(t *testing.T) {
// Verify one letter from each group extracts cells across different columns
matrix := newTestMatrix()
dirty := DirtySpell{Spell: "adgjmpsvy"}
spell, err := dirty.Parse()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
password, err := matrix.ExtractPassword(spell, 0)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
defer password.Wipe()

letters := spell.MagicLetters()
var expected []byte
for _, l := range letters {
q := l.Query()
expected = append(expected, matrix[q.MatrixRow][q.LetterGroup]...)
}
if !bytes.Equal(password.Bytes(), expected) {
t.Errorf("expected %q, got %q", expected, password.Bytes())
}
}

func TestMatrix_ExtractPassword_Spaces(t *testing.T) {
// Verify spaces map to group 0 same as digits, extracting identical cells
matrix := newTestMatrix()
dirty := DirtySpell{Spell: " "}
spell, err := dirty.Parse()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
password, err := matrix.ExtractPassword(spell, 0)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
defer password.Wipe()
if !bytes.Equal(password.Bytes(), matrix[0][0]) {
t.Errorf("expected %q, got %q", matrix[0][0], password.Bytes())
}
}

func TestMatrix_ExtractPassword_CaseSensitive(t *testing.T) {
// Verify that changing case of letters produces different passwords
matrix := newTestMatrix()

parse := func(s string) MagicSpell {
sp, err := DirtySpell{Spell: s}.Parse()
if err != nil {
t.Fatalf("parse %q: %v", s, err)
}
return sp
}
extract := func(sp MagicSpell) *SecureBytes {
p, err := matrix.ExtractPassword(sp, 0)
if err != nil {
t.Fatalf("extract: %v", err)
}
return p
}

passLower := extract(parse("amazon"))
defer passLower.Wipe()
passUpper := extract(parse("AMAZON"))
defer passUpper.Wipe()
passMixed := extract(parse("AmAzOn"))
defer passMixed.Wipe()

if bytes.Equal(passLower.Bytes(), passUpper.Bytes()) {
t.Error("lowercase and uppercase spells produced identical passwords")
}
if bytes.Equal(passLower.Bytes(), passMixed.Bytes()) {
t.Error("lowercase and mixed case spells produced identical passwords")
}
if bytes.Equal(passUpper.Bytes(), passMixed.Bytes()) {
t.Error("uppercase and mixed case spells produced identical passwords")
}
}

func TestMatrix_ExtractPassword_Truncation(t *testing.T) {
// Verify truncation works correctly at various lengths including mid-cell boundaries
matrix := newTestMatrix()
spell, err := DirtySpell{Spell: "aaaa"}.Parse()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

tests := []struct {
name     string
maxLen   int
expected int
}{
{"no truncation", 0, 4 * CharactersPerMatrixCell},
{"exact length", 4 * CharactersPerMatrixCell, 4 * CharactersPerMatrixCell},
{"longer than password", 100, 4 * CharactersPerMatrixCell},
{"truncate to 5 chars", 5, 5},
{"truncate to 1 char", 1, 1},
{"exact cell boundary", 3 * CharactersPerMatrixCell, 3 * CharactersPerMatrixCell},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
password, err := matrix.ExtractPassword(spell, tt.maxLen)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
defer password.Wipe()
if password.Len() != tt.expected {
t.Errorf("maxLen=%d: expected len %d, got %d", tt.maxLen, tt.expected, password.Len())
}
})
}
}

func TestIsAllowedSpellChar(t *testing.T) {
// Verify allowed and rejected characters
tests := []struct {
char     rune
expected bool
desc     string
}{
{'a', true, "lowercase letter"},
{'z', true, "lowercase letter"},
{'A', true, "uppercase letter"},
{'Z', true, "uppercase letter"},
{'0', true, "digit"},
{'9', true, "digit"},
{' ', true, "space"},
{'!', true, "special char"},
{'-', true, "special char"},
{'@', true, "special char"},
{'\x00', false, "control char NUL"},
{'\x1f', false, "control char"},
{127, false, "DEL"},
}

for _, tt := range tests {
got := IsAllowedSpellChar(tt.char)
if got != tt.expected {
t.Errorf("IsAllowedSpellChar(%q) = %v, want %v (%s)", tt.char, got, tt.expected, tt.desc)
}
}
}
