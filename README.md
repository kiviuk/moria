# pwdgen

A deterministic, matrix-based password generator. Generate unique, strong passwords for every service from a single master password and a memorable spell.

## Core Concept

`pwdgen` uses a **password matrix** — a grid of random character fragments — combined with a **spell** (any memorable string like "amazon" or "gmail") to derive unique passwords. The same master password + spell always produces the same password.

```
Master Password (secret) + Spell (service name) → Unique Password
```

## Features

- **Deterministic** — same inputs always produce the same output
- **Case-sensitive spells** — "amazon" and "AMAZON" produce different passwords
- **No password storage** — passwords are derived on-demand, never stored
- **Shell-safe master passwords** — generated passwords avoid shell metacharacters
- **Interactive live mode** — type your spell and watch the password build in real-time
- **Pretty matrix display** — visualize the password matrix for verification
- **Configurable** — all matrix dimensions are compile-time constants

## Installation

```bash
git clone https://github.com/kiviuk/pwdgen.git
cd pwdgen
make build
```

The binary is built to `bin/pwdgen`.

## Quick Start

### 1. Generate a Master Password

```bash
./bin/pwdgen --magic
```

This outputs a 300-character shell-safe random string. **Save this securely** — it's your master key. Store it in KeePass, 1Password, or any password manager.

### 2. Generate a Service Password

```bash
# Interactive (paste master password when prompted)
./bin/pwdgen "amazon"

# Piped (from password manager)
cat master.txt | ./bin/pwdgen "amazon"
```

Output: a unique password derived from your master password + the spell "amazon".

### 3. Display the Matrix

```bash
cat master.txt | ./bin/pwdgen --pretty
```

Shows the full password matrix with column headers:

```
       Non    ABC    DEF    GHI    JKL    MNO    PQR    STU    VWX    YZ
       ───    ───    ───    ───    ───    ───    ───    ───    ───    ───
0      xK9    !mP    2@n    Q7#    rT5    $wY    8aB    cD4    eF6    gH7
1      ...
...
```

### 4. Interactive Live Mode

```bash
cat master.txt | ./bin/pwdgen --live
```

Type your spell character by character. The matrix highlights visited cells and the password builds in real-time. Press Enter to output the final password.

### 5. Limit Password Length

Some sites cap password length. Use `--max-len` to truncate:

```bash
cat master.txt | ./bin/pwdgen --max-len 16 "amazon"
```

## How It Works

### The Algorithm

1. **Master Password → Matrix**: Your 300-character master password is arranged into a 10×10 grid of 3-character cells
2. **Spell → Path**: Each character in your spell determines a cell to read:
   - **Row** = character position in spell, modulo 10 (uppercase letters shift by +5)
   - **Column** = letter group (A-C→1, D-F→2, ..., Y-Z→9, non-letters→0)
3. **Extract Password**: Concatenate the cell contents along the path

### Example

Spell: `"amazon"` (6 characters)

| Char | Position | Row | Group | Column | Cell |
|------|----------|-----|-------|--------|------|
| a | 0 | 0 | 1 | 1 | (0,1) |
| m | 1 | 1 | 5 | 5 | (1,5) |
| a | 2 | 2 | 1 | 1 | (2,1) |
| z | 3 | 3 | 9 | 9 | (3,9) |
| o | 4 | 4 | 5 | 5 | (4,5) |
| n | 5 | 5 | 5 | 5 | (5,5) |

Output: 6 cells × 3 chars = 18-character password.

### Case Sensitivity

Uppercase letters shift the row by `PasswordMatrixRows/2`, making `"amazon"` and `"AMAZON"` produce completely different passwords. This adds entropy without requiring a longer spell.

## Security Model

### What's Secret
- **Master password** — the 300-character random string. This is your only secret.

### What's Public
- **Spell** — the service name (e.g., "amazon"). An attacker knowing this gets nothing without the master password.

### Entropy
- **Matrix**: 300 chars × ~6 bits/char = ~1800 bits of entropy
- **Password**: For a 6-letter spell, 18 chars × ~6 bits = ~108 bits
- **Brute force**: Computationally infeasible

### Key Derivation

Any input (SSH key, passphrase, random string) is deterministically expanded to the matrix size using **HKDF-SHA256** with rejection sampling for zero modulo bias. This means you can use your SSH private key as a master key:

```bash
cat ~/.ssh/id_ed25519 | ./bin/pwdgen "amazon"
```

## Configuration

All matrix dimensions are compile-time constants in `internal/app/config.go`:

| Constant | Default | Description |
|----------|---------|-------------|
| `PasswordMatrixRows` | 10 | Number of rows (position modulus) |
| `CharactersPerMatrixCell` | 3 | Characters per cell (password length multiplier) |
| `AlphabetSize` | 26 | Letters in the alphabet |
| `MasterPasswordChars` | 64 chars | Shell-safe character pool for `--magic` |

To change the matrix size, edit the constants and run `make test && make build`. All tests pass with any value.

## CLI Reference

```
Usage: pwdgen [--magic|--pretty|--live] [--max-len N] <spell>

  --magic    Generate a master password
  --pretty   Display the password matrix from your master password
  --live     Interactive mode: type your spell and see the password build in real-time
  --max-len  Truncate output to N characters (live and batch modes only)
  <spell>    Generate a service password from your spell
```

## Project Structure

```
pwdgen/
├── cmd/pwdgen/
│   ├── main.go             # CLI entry point
│   ├── live.go             # Bubbletea TUI for interactive mode
│   ├── live_test.go        # Tests for live mode
│   └── main_test.go        # Tests for CLI, flag parsing, validation
├── internal/
│   ├── app/
│   │   ├── config.go               # Package-level constants
│   │   ├── spell.go                # Core domain types (MagicLetter, QueryLetter, etc.)
│   │   ├── spell_test.go           # Tests for parsing, grouping, resolution
│   │   ├── password_matrix.go      # Matrix type, generation, Pretty(), Cell access
│   │   └── password_matrix_test.go # Matrix dimension, content, and integration tests
│   └── testutil/
│       └── testutil.go             # Shared test data generator (no import cycles)
├── go.mod
└── Makefile
```

## Testing

```bash
make test                          # Run all tests
go clean -testcache && make test   # Clear cache and re-run
go test ./internal/app/ -v         # Verbose output for app tests
go test ./cmd/pwdgen/ -v           # Verbose output for cmd tests
go test ./... -run TestQuery       # Run single test by name
```

All tests pass with any `CharactersPerMatrixCell` and `PasswordMatrixRows` values — expected values are computed from constants, not hardcoded.

## License

MIT
