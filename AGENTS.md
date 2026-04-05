# AGENTS.md — pwdgen Development Guide

## Build / Lint / Test Commands

```bash
make build          # Compile to bin/pwdgen
make run            # Build and execute
make test           # Run all tests: go test ./...
make clean          # Remove bin/
go clean -testcache && make test   # Clear cache and re-run all tests
go test ./internal/app/ -v         # Verbose test output
go test ./cmd/pwdgen/ -v           # Verbose output for cmd tests
go test ./... -run TestQuery       # Run single test by name
```

**Running a single test:** `go test ./... -run <TestName>` is the standard way.
Use `-v` for verbose output showing each test's PASS/FAIL status.

## Project Structure

```
pwdgen/
├── cmd/pwdgen/
│   ├── main.go             # CLI entry point (--magic, --pretty, --live, --max-len)
│   ├── live.go             # Bubbletea TUI for interactive password generation
│   ├── live_test.go        # Tests for live mode model
│   └── main_test.go        # Tests for batch mode, flag parsing, validation
├── internal/
│   ├── app/
│   │   ├── config.go               # Package-level constants
│   │   ├── spell.go                # Core domain types (MagicLetter, QueryLetter, MagicSpell, DirtySpell)
│   │   ├── spell_test.go           # Tests for parsing, grouping, resolution
│   │   ├── password_matrix.go      # Matrix type, generation, Pretty(), Cell access
│   │   └── password_matrix_test.go # Matrix dimension, content, and integration tests
│   └── testutil/
│       └── testutil.go             # Shared test data generator (no import cycles)
├── go.mod                          # Module: github.com/kiviuk/pwdgen (go 1.26.1)
└── Makefile
```

## CLI Usage

```bash
# Generate a 300-character shell-safe master password
pwdgen --magic

# Display the password matrix from a master password
pwdgen --pretty < master.txt

# Interactive mode: type spell, see password build in real-time
pwdgen --live < master.txt

# Batch mode: generate password from spell
pwdgen "amazon" < master.txt

# With max length (live and batch modes only)
pwdgen --max-len 16 "amazon" < master.txt
```

**Master password:** Must be exactly `PasswordMatrixRows * PasswordMatrixColumns * CharactersPerMatrixCell` characters (300 by default). Read from stdin — supports piping from KeePass/1Password.

**Output:** Password printed to stdout with no trailing newline — safe for `pbcopy`.

## Code Style Guidelines

### Imports
- Standard library + `charmbracelet/bubbletea` and `charmbracelet/lipgloss` for TUI.
- Single parenthesized block, alphabetically ordered:
  ```go
  import (
      "crypto/rand"
      "fmt"
      "strings"
  )
  ```

### Formatting
- Use `gofmt` — tabs for indentation, no trailing whitespace.
- No `.golangci.yml` or `.editorconfig` — follow `gofmt` defaults.

### Naming Conventions
- **Exported types:** `PascalCase` — `MagicLetter`, `QueryLetter`, `MagicSpell`, `DirtySpell`, `Matrix`, `ParseError`, `Errors`
- **Exported constants:** `PascalCase` — `PasswordMatrixRows`, `CharactersPerMatrixCell`, `AlphabetSize`, `MaxLetterGroups`, `PasswordMatrixColumns`, `MasterPasswordChars`
- **Exported functions:** `PascalCase` — `LetterGroup()`, `ModN()`, `GenerateMasterPassword()`, `NewMatrix()`, `ColHeader()`
- **Unexported:** `camelCase` — `cell()`, `newTestMatrix()`, `readMasterPassword()`
- **Test functions:** `Test<TypeName>_<Method>_<Scenario>` — e.g., `TestDirtySpell_Parse_Valid`
- **Receiver names:** Short, single-letter abbreviations — `d` for `DirtySpell`, `m` for `MagicSpell`/`MagicLetter`, `e` for `Errors`

### Types
- Value types preferred over pointers for domain structs.
- Custom error types: `ParseError` struct + `Errors []ParseError` implementing `error` interface.
- **`MagicLetter`** — raw spell character with position and group
- **`QueryLetter`** — resolved matrix coordinates (MatrixRow, LetterGroup), safe for matrix access
- Struct field order: `Letter`, `LetterPosition`/`MatrixRow`, `LetterGroup` (consistent across all usages).

### Error Handling
- **Accumulate all errors** — do not fail on first invalid input. `DirtySpell.Parse()` collects all `ParseError`s and returns them at once.
- Use `fmt.Errorf` for simple single-error cases (e.g., `"spell cannot be empty"`).
- Use type assertion in tests to access accumulated errors: `errs := err.(Errors)`
- Classic Go pattern: `if err != nil { t.Fatalf(...) }` — no `testify` or assertion libraries.

### Testing
- Tests are in the same package (`package app` or `package main`), not black-box.
- **Table-driven tests** are the standard pattern:
  ```go
  tests := []struct {
      input    string
      expected int
  }{
      {"a", 1}, {"b", 1},
  }
  for _, tt := range tests {
      if got := LetterGroup(tt.input); got != tt.expected {
          t.Errorf("LetterGroup(%q) = %d, expected %d", tt.input, got, tt.expected)
      }
  }
  ```
- Every test function must have a brief comment (max 120 chars) explaining its purpose.
- No `errors.Is`/`errors.As` used — project uses direct type assertion for custom errors.
- **Test helpers:** `internal/testutil` provides `NewTestMatrixData()` for shared test data without import cycles. Each package defines its own `newTestMatrix()` wrapper.
- Tests must work with any `CharactersPerMatrixCell` value — no hardcoded lengths or expected values.

### Architecture Patterns
- **Parse pattern:** `DirtySpell` (untrusted) → `.Parse()` → `MagicSpell` (validated)
- **Resolution pipeline:** `MagicSpell.MagicLetters()` → `[]MagicLetter` → `.Resolve()` → `[]QueryLetter` → `Matrix.Cell()` → password
- **Constants-driven:** All magic numbers and character sets live in `config.go`
- **Matrix navigation:** `MatrixRow` = row (0-9, wrapped via modulo), `LetterGroup` = col (0 = non-letters, 1+ = letter groups)
- **Case-insensitive grouping:** `a` and `A` both map to group 1
- **Deterministic:** Same master password + same spell = same password every time
- **No trailing newline** in password output — safe for piping to `pbcopy`
- **Shell-safe master password:** Uses `MasterPasswordChars` (excludes `{}`, `[]`, `~`, `"`, `'`, space, `$`, `!`, `#`, `&`, `*`, `?`, `()`, `|`, `<>`, `;`, `\`, `` ` ``)
- **SRP:** `GenerateMasterPassword(length, pool)` accepts character pool as parameter — no hardcoded pools
- **Defensive access:** `Matrix.Cell(t QueryLetter)` validates column bounds; row is guaranteed valid by `QueryLetter` type

### Flexibility
- Change `CharactersPerMatrixCell` in `config.go` to adjust cell size (1, 2, 3, etc.)
- All tests pass with any value — expected values are computed from constants, not hardcoded
- `PasswordMatrixColumns` is derived: `(AlphabetSize + CharactersPerMatrixCell - 1) / CharactersPerMatrixCell + 1`
