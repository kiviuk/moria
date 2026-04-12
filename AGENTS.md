# AGENTS.md — moria Development Guide

## Build / Lint / Test Commands

```bash
make build          # Compile to bin/moria
make run            # Build and execute
make test           # Run all tests: go test ./...
make lint           # Run golangci-lint (auto-installs if missing via make deps)
make clean          # Remove bin/
go clean -testcache && make test   # Clear cache and re-run all tests
go test ./internal/app/ -v         # Verbose test output
go test ./cmd/moria/ -v            # Verbose output for cmd tests
go test ./... -run TestQuery       # Run single test by name
```

**Linting:** `make lint` runs `golangci-lint` against all packages. The binary is auto-installed to `$GOPATH/bin` if not found. Config lives in `.golangci.yml`.

**Running a single test:** `go test ./... -run <TestName>` is the standard way.
Use `-v` for verbose output showing each test's PASS/FAIL status.

## Project Structure

```
moria/
├── cmd/moria/
│   ├── main.go                # CLI entry point (--magic, --pretty, --live, --max-len, --show-strength, --ignore-paste, -h/--help)
│   ├── live.go                # Bubbletea TUI for interactive password generation
│   ├── live_test.go           # Tests for live mode model
│   ├── password_prompt.go     # Bubbletea password input prompt (masked with •)
│   ├── messages.go            # CLI error messages and live mode UI strings
│   └── main_test.go           # Tests for batch mode, flag parsing, validation, pipe input
├── internal/
│   ├── app/
│   │   ├── config.go               # Package-level constants
│   │   ├── spell.go                # Core domain types (MagicLetter, QueryLetter, MagicSpell, DirtySpell)
│   │   ├── spell_test.go           # Tests for parsing, grouping, resolution, case sensitivity, IsAllowedSpellChar
│   │   ├── password_matrix.go      # Matrix type, generation, Pretty(), Cell access, ExpandToMatrix()
│   │   ├── password_matrix_test.go # Matrix dimension, content, and integration tests
│   │   ├── strength.go             # Time-to-guess calculation and human-readable formatting
│   │   └── strength_test.go        # Tests for CrackTime, FormatSeconds, Entropy
│   └── testutil/
│       └── testutil.go             # Shared test data generator (no import cycles)
├── .golangci.yml                   # golangci-lint configuration
├── go.mod                          # Module: github.com/kiviuk/moria (go 1.26.1)
└── Makefile
```

## CLI Usage

```bash
# Generate a master password (shell-safe, length based on constants)
moria --magic

# Display the password matrix from a master password
moria --pretty < master.txt

# Interactive mode: type spell, see password build in real-time
moria --live < master.txt

# Batch mode: generate password from spell
moria "amazon" < master.txt

# With max length (live and batch modes only)
moria --max-len 16 "amazon" < master.txt

# Show time-to-guess estimates (batch mode only)
moria --show-strength "amazon" < master.txt

# Help
moria --help
moria -h
```

**Master password:** Any input (random string, SSH key, passphrase) is deterministically expanded to the matrix size using Argon2id + HKDF. Read from stdin — supports piping from KeePass/1Password.

**Output:** Password printed to stdout with no trailing newline — safe for `pbcopy`.

## Code Style Guidelines

### Imports
- Standard library + `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `charmbracelet/bubbles/textinput` for TUI.
- `golang.org/x/crypto/argon2` for memory-hard key derivation.
- Single parenthesized block, alphabetically ordered:
  ```go
  import (
      "crypto/hkdf"
      "crypto/rand"
      "crypto/sha256"
      "fmt"
      "strings"

      "github.com/kiviuk/moria/internal/app"
  )
  ```

### Formatting
- Use `gofmt` — tabs for indentation, no trailing whitespace.
- Linting is enforced via `golangci-lint`. Config lives in `.golangci.yml` at the project root.
- Run `make lint` before committing. The `make deps` target auto-installs `golangci-lint` if missing.
- Enabled linters: `govet`, `staticcheck`, `errcheck`, `revive`, `gocyclo` (max 20), `gosec`, `goimports`, `misspell`, `unconvert`, `predeclared`, `ineffassign`, `unused`, `gocritic`, `noctx`.
- Test files (`*_test.go`) are excluded from `errcheck`, `gosec`, `gocyclo`, and `revive`.

### Naming Conventions
- **Exported types:** `PascalCase` — `MagicLetter`, `QueryLetter`, `MagicSpell`, `DirtySpell`, `Matrix`, `ParseError`, `Errors`
- **Exported constants:** `PascalCase` — `PasswordMatrixRows`, `CharactersPerMatrixCell`, `AlphabetSize`, `MaxLetterGroups`, `PasswordMatrixColumns`, `MasterPasswordChars`, `MatrixBytes`, `LiveModeWrapWidth`
- **Exported functions:** `PascalCase` — `IsAllowedSpellChar()`, `LetterGroup()`, `ModN()`, `GenerateMasterPassword()`, `NewMatrix()`, `ColHeader()`, `ExpandToMatrix()`, `ExtractPassword()`, `FormatSecondsCompact()`, `CalculateMasterPasswordEntropy()`, `CalculateMasterPasswordStrength()`
- **Unexported:** `camelCase` — `cell()`, `mapStringSourceToAlphabet()`, `newTestMatrix()`, `getPassword()`, `wrapWithIndent()`
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
- Tests must work with any `CharactersPerMatrixCell` and `PasswordMatrixRows` values — expected values are computed from constants, not hardcoded.

### Architecture Patterns
- **Parse pattern:** `DirtySpell` (untrusted) → `.Parse()` → `MagicSpell` (validated)
- **Resolution pipeline:** `MagicSpell.MagicLetters()` → `[]MagicLetter` → `.Query()` → `[]QueryLetter` → `Matrix.Cell()` → password
- **Key derivation:** Any input → `ExpandToMatrix()` (Argon2id + HKDF) → `MatrixBytes` string → `NewMatrix()` → `Matrix`
- **Constants-driven:** All magic numbers and character sets live in `config.go`
- **Matrix navigation:** `MatrixRow` = row (0-19, wrapped via modulo), `LetterGroup` = col (0 = non-letters, 1+ = letter groups)
- **Case-sensitive rows:** Uppercase letters shift row by `PasswordMatrixRows/2`, making case significant
- **Case-insensitive grouping:** `a` and `A` both map to group 1
- **Deterministic:** Same master password + same spell = same password every time
- **No trailing newline** in password output — safe for piping to `pbcopy`
- **Bash-friendly master password:** Uses `MasterPasswordChars` (excludes `{}`, `[]`, `~`, `"`, `'`, space, `$`, `!`, `#`, `&`, `*`, `?`, `()`, `|`, `<>`, `;`, `\`, `` ` ``)
- **SRP:** `GenerateMasterPassword(length, pool)` accepts character pool as parameter — no hardcoded pools
- **Defensive access:** `Matrix.Cell(t QueryLetter)` validates column bounds; row is guaranteed valid by `QueryLetter` type
- **Rejection sampling:** `mapStringSourceToAlphabet()` ensures zero modulo bias when mapping random bytes to character pool

### Flexibility
- Change `CharactersPerMatrixCell` in `config.go` to adjust cell size (1, 2, 3, etc.)
- Change `PasswordMatrixRows` to adjust matrix row count (any positive integer)
- All tests pass with any value — expected values are computed from constants, not hardcoded
- `PasswordMatrixColumns` is derived: `(AlphabetSize + CharactersPerMatrixCell - 1) / CharactersPerMatrixCell + 1`
- `MatrixBytes` is derived: `PasswordMatrixRows * PasswordMatrixColumns * CharactersPerMatrixCell`
