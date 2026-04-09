# moria

A deterministic, matrix-based password generator. Generate unique, strong passwords for every service/login/vault from a single master password and a memorable spell.

> *"Speak, friend, and enter."* — Your spell is the password. The matrix is the mine.

![Moria live mode](docs/moria-live.png)

**Primary use case for developers:** If you already have an SSH private key (e.g., for GitHub, servers, or CI/CD), you can reuse it as your master password. Generate passwords tied to that same ecosystem — no new secret to manage.

Inspired by [pwgen](https://www.uni-muenster.de/CERT/pwgen/index.php?lang=en&mode=pwcard)

Based on [zxcvb](https://github.com/ccojocar/zxcvbn-go) and [Argon2id](https://en.wikipedia.org/wiki/Argon2)

Videos: [zxcvbn](https://www.youtube.com/watch?v=vf37jh3dV2I) [Argon2id](https://youtu.be/Sc3aHMCc4h0?t=114)

## Core Concept

`moria` uses a **password matrix** — a grid of random character fragments — combined with a **spell** (any memorable key phrase) to derive unique passwords. The same master password + spell always produces the same password.

```
Master Password (secret) + Spell (pass-phrase) → Unique Password
```

**Primary use case for developers:** If you already have an SSH private key (e.g., for GitHub, servers, or CI/CD), you can reuse it as your master password. Generate passwords tied to that same ecosystem — no new secret to manage.

Example: Your `id_ed25519` key grants access to GitHub. Use the same key to generate your GitHub password, personal access tokens, or other GitHub-related credentials.

## Installation

```bash
git clone https://github.com/kiviuk/moria.git
cd moria
make build
```

The binary is built to `bin/moria`.

## Quick Start

### 1. Generate a Password

**Option A: Use your existing SSH key (recommended)**
If you already have an SSH private key, use it directly — no new secret to manage:
```bash
cat ~/.ssh/id_ed25519 | ./bin/moria "phrase-i-can-remember"
# → xK9!nQ7#5$wYBcD4
```

**Option B: Generate a new master password with `--magic`**
If you don't have an SSH key, generate a cryptographically secure master password:
```bash
./bin/moria --magic
```
This outputs a 600-character random string. **Save this securely.**

You can then pipe it or store it in a password manager:
```bash
# Pipe from file
cat the-key.txt | ./bin/moria "phrase-i-can-remember"
# → xK9!nQ7#5$wYBcD4

# Interactive prompt (password manager)
./bin/moria "phrase-i-can-remember"
# → you'll be prompted to paste your master password (input is masked with •••)
```

### 2. Interactive Live Mode

```bash
cat the-key.txt | ./bin/moria --live
# Or without piping — you'll be prompted to paste your master password
./bin/moria --live
```

Type your spell character by character. The matrix highlights visited cells and the password builds in real-time. Press Enter to output the final password.

```
Spell:    phrase-i-can-remember
Password: xK9!nQ7#5$wYBcD4 (18/18)
```

### 3. Display the Matrix

```bash
cat the-key.txt | ./bin/moria --pretty
# Or without piping — you'll be prompted to paste your master password
./bin/moria --pretty
```

Shows the full password matrix with column headers:

```
       Non    ABC    DEF    GHI    JKL    MNO    PQR    STU    VWX    YZ
       ────   ────   ────   ────   ────   ────   ────   ────   ────   ────
0      xK9!   nQ7#   5$wY   BcD4   6gH7   1lM2   3pQ4   5tU6   7xY8   9bC0
1      ...
...
```

### 4. Limit Password Length

Some sites cap password length. Use `--max-len` to truncate:

```bash
cat the-key.txt | ./bin/moria --max-len 16 "phrase-i-can-remember"
# → xK9!nQ7#5$wYBcD4
```

### 5. Check Password Strength

Analyze the strength of any password using [zxcvbn](https://github.com/ccojocar/zxcvbn-go) pattern detection:

```bash
echo "i'm super hunger today" | ./bin/moria --show-strength
```

Output:
```
zxcvbn master password entropy: 50 bits

zxcvbn crack time estimate (generic): centuries

Assuming attacker 100k guesses/sec and 50 bits (from zxcvbn), worst case: 357 years
```

## Security Model

Your master password and spell are secret. The generated password is what you use to log in.

| Component | Secret? | Notes |
|-----------|---------|-------|
| **Master password** | Yes | Your only secret. Compromise = total loss. |
| **Spell** | Yes | Your memorable phrase per service. Keep private. |
| **Generated password** | Until leaked | What you type to log in. Safe if master is secure. |

## Understanding Your Security

The strength of your derived passwords is limited by your master password. A long spell cannot compensate for a weak master.

`--show-strength` analyzes your master password strength using `zxcvbn` pattern detection:

### Example: A Passphrase Master Password

```bash
echo "i'm super hunger today" | ./bin/moria --show-strength
```

Output:
```
zxcvbn master password entropy: 50 bits

zxcvbn crack time estimate (generic): centuries

Assuming attacker 100k guesses/sec and 50 bits (from zxcvbn), worst case: 357 years
```

[zxcvbn](https://github.com/ccojocar/zxcvbn-go) detects that `"i'm super hunger today"` is four common English words. Instead of multiplying 22 × 6 bits (which assumes random gibberish), it calculates the actual entropy of a dictionary-word passphrase.

The **357 years** estimate is calculated as: `(2^50 guesses) ÷ (100k guesses/sec)`. The 50 bits reflects the effective entropy after accounting for dictionary patterns.

All four words ("i'm", "super", "hungry", "today") are common, but zxcvbn can't detect *semantic combinations*. It sees 4 dictionary words, not a common phrase. The password "i'm super hungry today" is memorable and guessable to humans — but it's not in any attacker's wordlist. **Pattern detection is limited to what attackers precompute**.

**Practical takeaway:** Combine common words in unique, memorable ways. Even simple phrases are safer than you think because attackers can't precompute every possible combination.

## Configuration

All matrix dimensions are compile-time constants in `internal/app/config.go`:

| Constant | Default | Description |
|----------|---------|-------------|
| `PasswordMatrixRows` | 20 | Number of rows (position modulus) |
| `CharactersPerMatrixCell` | 3 | Characters per cell (password length multiplier) |
| `AlphabetSize` | 26 | Letters in the alphabet |
| `MasterPasswordChars` | 73 chars | Bash-friendly characters for `--magic` |

To change the matrix size, edit the constants and run `make test && make build`. All tests pass with any value.

## CLI Reference

```
Usage: moria [--magic|--pretty|--live|--show-strength] [--max-len N] [--ignore-paste] [spell]

Options:
  --magic                Generate a master password
  --pretty               Display the password matrix from your master password
  --live                 Interactive mode: type your spell and see the password build in real-time
  --show-strength    Analyze password strength from stdin (standalone, no spell)
  --max-len N            Truncate output to N > 0 characters (live and batch modes only)
  --ignore-paste         Ignore pasted input in live mode (live mode only)
  -h, --help             Show this help message
```

## Project Structure

```
moria/
├── cmd/moria/
│   ├── main.go                # CLI entry point
│   ├── live.go                # Bubbletea TUI for interactive mode
│   ├── live_test.go           # Tests for live mode
│   ├── password_prompt.go     # Bubbletea password input prompt
│   ├── messages.go            # CLI error messages and live mode UI strings
│   └── main_test.go           # Tests for CLI, flag parsing, validation
├── internal/
│   ├── app/
│   │   ├── config.go               # Package-level constants
│   │   ├── spell.go                # Core domain types (MagicLetter, QueryLetter, etc.)
│   │   ├── spell_test.go           # Tests for parsing, grouping, resolution
│   │   ├── password_matrix.go      # Matrix type, generation, Pretty(), Cell access
│   │   └── password_matrix_test.go # Matrix dimension, content, and integration tests
│   └── testutil/
│       └── testutil.go             # Shared test data generator (no import cycles)
├── .golangci.yml                   # golangci-lint configuration
├── go.mod
└── Makefile
```

## Testing

```bash
make test                          # Run all tests
make lint                          # Run golangci-lint
go clean -testcache && make test   # Clear cache and re-run
go test ./internal/app/ -v         # Verbose output for app tests
go test ./cmd/moria/ -v            # Verbose output for cmd tests
go test ./... -run TestQuery       # Run single test by name
```

All tests pass with any `CharactersPerMatrixCell` and `PasswordMatrixRows` values — expected values are computed from constants, not hardcoded.

## License

MIT
