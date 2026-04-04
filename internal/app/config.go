package app

// PasswordMatrixRows is the number of rows in the password fragment matrix.
// Also used as the modulus for wrapping character positions.
const PasswordMatrixRows = 10

// CharactersPerMatrixCell is the number of characters stored in each matrix cell.
const CharactersPerMatrixCell = 1

// AlphabetSize is the total number of letters in the English alphabet.
const AlphabetSize = 26

// MaxLetterGroups is the total number of letter groups (A-C=1, D-F=2, ..., X-Z=9).
const MaxLetterGroups = (AlphabetSize + CharactersPerMatrixCell - 1) / CharactersPerMatrixCell

// PasswordMatrixColumns is the total number of columns in the matrix.
// Equals MaxLetterGroups + 1, where column 0 holds non-letter characters.
const PasswordMatrixColumns = MaxLetterGroups + 1

// AllowedLetters is the regex character range for alphabetic characters.
const AllowedLetters = "a-zA-Z"

// AllowedNumbers is the regex character range for numeric characters.
const AllowedNumbers = "0-9"

// AllowedSpecialChars is the set of permitted special characters.
const AllowedSpecialChars = `!@#$%^&*()-_=+[]{}|;:,.<>?/~` + "`\"'"

// AllowedSpace represents the only permitted whitespace character.
const AllowedSpace = " "

// MasterPasswordChars contains shell-safe characters for master password generation.
// Excludes shell metacharacters: {} [] ~ " ' space $ ! # & * ? ( ) | < > ; \ `
const MasterPasswordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@-_=+:%.^/,"
