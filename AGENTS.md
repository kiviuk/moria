# AGENTS.md — pwdgen Development Guide

## Build / Lint / Test Commands

```bash
make build          # Compile to bin/pwdgen
make run            # Build and execute
make test           # Run all tests: go test ./...
make clean          # Remove bin/
go clean -testcache && make test   # Clear cache and re-run all tests
go test ./internal/app/ -v         # Verbose test output
go test ./internal/app/ -run TestMagicSpell_LetterTuples  # Run single test
go test ./internal/app/ -run TestDirtySpell_Parse_Valid   # Run single test by name
```

**Running a single test:** `go test ./internal/app/ -run <TestName>` is the standard way.
Use `-v` for verbose output showing each test's PASS/FAIL status.

## Project Structure

```
pwdgen/
├── cmd/pwdgen/main.go          # CLI entry point
├── internal/app/
│   ├── config.go               # Package-level constants
│   ├── letter_matrix.go        # Core domain types & logic
│   ├── app_test.go             # Placeholder test
│   └── spell_test.go           # Comprehensive test suite
├── go.mod                      # Module: github.com/kiviuk/pwdgen (go 1.26.1)
└── Makefile
```

## Code Style Guidelines

### Imports
- Standard library only — no third-party dependencies.
- Single parenthesized block, alphabetically ordered:
  ```go
  import (
      "fmt"
      "strings"
  )
  ```

### Formatting
- Use `gofmt` — tabs for indentation, no trailing whitespace.
- No `.golangci.yml` or `.editorconfig` — follow `gofmt` defaults.

### Naming Conventions
- **Exported types:** `PascalCase` — `LetterTuple`, `MagicSpell`, `DirtySpell`, `ParseError`, `Errors`
- **Exported constants:** `PascalCase` — `MatrixN`, `GroupSize`, `AllowedLetters`
- **Exported functions:** `PascalCase` — `LetterGroup()`, `ModN()`
- **Unexported:** `camelCase`
- **Test functions:** `Test<TypeName>_<Method>_<Scenario>` — e.g., `TestDirtySpell_Parse_Valid`
- **Receiver names:** Short, single-letter abbreviations — `d` for `DirtySpell`, `m` for `MagicSpell`/`LetterTuple`, `e` for `Errors`

### Types
- Value types preferred over pointers for domain structs.
- Custom error types: `ParseError` struct + `Errors []ParseError` implementing `error` interface.
- Struct field order: `Letter`, `LetterPosition`, `LetterGroup` (consistent across all usages).

### Error Handling
- **Accumulate all errors** — do not fail on first invalid input. `DirtySpell.Parse()` collects all `ParseError`s and returns them at once.
- Use `fmt.Errorf` for simple single-error cases (e.g., `"spell cannot be empty"`).
- Use type assertion in tests to access accumulated errors: `errs := err.(Errors)`
- Classic Go pattern: `if err != nil { t.Fatalf(...) }` — no `testify` or assertion libraries.

### Testing
- Tests are in the same package (`package app`), not black-box (`package app_test`).
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

### Architecture Patterns
- **Parse pattern:** `DirtySpell` (untrusted) → `.Parse()` → `MagicSpell` (validated)
- **Transformation pipeline:** `MagicSpell.LetterTuples()` → `[]LetterTuple` → `.MapModN()` → wrapped positions
- **Constants-driven:** All magic numbers and character sets live in `config.go`
- **Matrix navigation:** `LetterPosition` = row (0-9), `LetterGroup` = col (0 = non-letters, 1+ = letter groups)
- **Case-insensitive grouping:** `a` and `A` both map to group 1
