# moria

A deterministic, matrix-based password generator. Generate unique, strong passwords for every service/login/vault from a single master password and a memorable spell.

> *"Speak, friend, and enter."* — Your spell is the password. The matrix is the mine.

![Moria live mode](docs/moria-live.png)

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
```
Output: `xK9!nQ7#5$wYBcD4` (unique password derived from your SSH key + the spell)

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

Analyze the strength of any password using [zxcvbn](https://github.com/ccojocar/zxcvbn-go) pattern detection. The password goes to `stdout`, the analysis goes to `stderr`:

```bash
echo "i'm super hunger today" | ./bin/moria --password-strength
```

Output:
```
Master password entropy: 50 bits

zxcvbn crack time estimate (generic): centuries

Time to guess (master password, via Argon2id):
  Single CPU               1.8 million years
  Single GPU               1.8 thousand years
  GPU cluster              178 years
```

## How It Works

### The Algorithm

1. **Master Password → Matrix**: Your master password is deterministically expanded into a grid of random character fragments
2. **Spell → Path**: Each character in your spell determines a cell to read:
   - **Row** = character position in spell, modulo `PasswordMatrixRows` (uppercase letters shift by `PasswordMatrixRows/2`)
   - **Column** = letter group (A-C→1, D-F→2, ..., Y-Z→9, non-letters→0)
3. **Extract Password**: Concatenate the cell contents along the path

### Example

Spell: `"phrase-i-can-remember"` (18 characters, including hyphens)

| Char | Position | Row | Group | Column | Cell |
|------|----------|-----|-------|--------|------|
| p | 0 | 0 | 6 (PQR) | 6 | (0,6) |
| w | 1 | 1 | 8 (VWX) | 8 | (1,8) |
| d | 2 | 2 | 2 (DEF) | 2 | (2,2) |
| - | 3 | 3 | 0 (Non) | 0 | (3,0) |
| i | 4 | 4 | 3 (GHI) | 3 | (4,3) |
| - | 5 | 5 | 0 (Non) | 0 | (5,0) |
| c | 6 | 6 | 1 (ABC) | 1 | (6,1) |
| a | 7 | 7 | 1 (ABC) | 1 | (7,1) |
| n | 8 | 8 | 5 (MNO) | 5 | (8,5) |
| - | 9 | 9 | 0 (Non) | 0 | (9,0) |
| r | 10 | 10 | 6 (PQR) | 6 | (10,6) |
| e | 11 | 11 | 2 (DEF) | 2 | (11,2) |
| m | 12 | 12 | 5 (MNO) | 5 | (12,5) |
| e | 13 | 13 | 2 (DEF) | 2 | (13,2) |
| m | 14 | 14 | 5 (MNO) | 5 | (14,5) |
| b | 15 | 15 | 1 (ABC) | 1 | (15,1) |
| e | 16 | 16 | 2 (DEF) | 2 | (16,2) |
| r | 17 | 17 | 6 (PQR) | 6 | (17,6) |

Output: 18 cells × 3 chars = 54-character password.

### Case Sensitivity

Uppercase letters shift the row by `PasswordMatrixRows/2`, making `"PHrase-I-can-remember"` and `"phrase-i-can-remember"` produce completely different passwords. This adds entropy without requiring a longer spell.

## Security Model

### What's Secret
- **Master password** — the random string (or any input like an SSH key). This is your master secret (the ring to rule them all).
- **Spell** - a private rememberable password. One for every login.

### What's Public
- **The generated password** from the `spell` e.g., "phrase-i-can-remember". An attacker knowing this gets nothing without the master password.

### Entropy
- **Matrix**: 600 chars × ~6.19 bits/char ≈ ~3,700 bits of entropy
- **Password**: For a 6-letter spell, 18 chars × ~6.19 bits ≈ ~111 bits
- **Brute force**: Computationally infeasible

### Key Derivation

Your master password goes through a two-stage process to become the matrix:

1. **[Argon2id](https://datatracker.ietf.org/doc/html/rfc9106)** — Your input is first passed through Argon2id (1 iteration, 64MB memory, 2 threads), a memory-hard key derivation function. This slows down the derivation to ~500ms, making brute-force attacks computationally expensive even if your master password is weak.
2. **[HKDF-SHA256](https://tools.ietf.org/html/rfc5869)** — The 32-byte high-entropy output from Argon2id is then expanded to the full matrix size (600 characters) using HKDF.

This means you can use any input as long as it has sufficient entropy:

- **SSH/GPG keys**: High entropy (generated with crypto/rand) — ideal
- **Strong passphrases**: Sufficient if long enough (see `--password-strength`)
- **Weak passphrases**: Risky — Argon2id slows attacks but doesn't replace missing entropy

```bash
cat ~/.ssh/id_ed25519 | ./bin/moria "phrase-i-can-remember"
```

### Rejection Sampling

When generating random passwords, a common mistake is to use the modulo operator (`%`) to map random bytes to a character set. This introduces **modulo bias** — some characters become slightly more likely than others, weakening the password.

Moria uses **rejection sampling** instead: if a random byte falls in the "biased" range, it's discarded and a new byte is drawn. This guarantees every character in the pool has exactly equal probability, preserving the full entropy of your passwords.

Imagine you have a 52-card deck and want to randomly pick a number from 1 to 10. If you just divide the card value by 10 and take the remainder, the numbers 1 and 2 would come up more often than the rest — because 52 doesn't divide evenly by 10, leaving 2 "extra" cards that loop back to the beginning.
Rejection sampling fixes this by saying: "If you draw one of those extra cards, put it back and draw again." You keep drawing until you get a card from the fair range. The result is that every number from 1 to 10 has exactly the same chance of being picked.
In moria's case, a random byte can be 0–255 (256 values), but the character pool has 73 characters. Since 256 doesn't always divide evenly into the pool size, some characters would be slightly more likely without rejection sampling. By discarding the "extra" bytes and drawing fresh ones, every character gets a perfectly fair shot.

## Understanding Your Security

The strength of your derived passwords is limited by your master password. A long spell cannot compensate for a weak master.

`--password-strength` analyzes your master password strength using `zxcvbn` pattern detection:

### Example: A Passphrase Master Password

```bash
echo "i'm super hunger today" | ./bin/moria --password-strength
```

**Why does the master password show 50 bits for 22 characters?**

zxcvbn detects that `"i'm super hunger today"` is four common English words. Instead of multiplying 22 × 6 bits (which assumes random gibberish like `X9q!pP2`), it calculates the actual entropy of a dictionary-word passphrase: ~50 bits. That means an attacker needs ~2⁴⁹ guesses (562 trillion) to crack it.

**The magic of Argon2id:**

If a hacker stole your master password hash from a normal website using MD5/SHA1, 50 bits would take them ~1.5 hours on a GPU cluster. Trivial.

But moria forces the attacker through Argon2id (64MB RAM per guess), bottlenecking a GPU cluster to ~100,000 guesses/sec. 562 trillion / 100,000 = **178 years**. Argon2id turned a weak human phrase into a 178-year mathematical wall.

**What this tells you:**

- If a website gets hacked: the attackers get the hash of your generated password. With ~111 bits of entropy, it will never be cracked.
- If someone targets **you**: they know you use moria and know your spell. They'll try to guess your master password. Because it's made of dictionary words, a GPU cluster could crack it in 178 years.

### The Rule

Your effective security is limited by your master password strength. If `--password-strength` shows "instant" or "minutes", pick a stronger master — not a longer spell.

## Configuration

All matrix dimensions are compile-time constants in `internal/app/config.go`:

| Constant | Default | Description |
|----------|---------|-------------|
| `PasswordMatrixRows` | 20 | Number of rows (position modulus) |
| `CharactersPerMatrixCell` | 3 | Characters per cell (password length multiplier) |
| `AlphabetSize` | 26 | Letters in the alphabet |
| `MasterPasswordChars` | 73 chars | Shell-safe character pool for `--magic` |

To change the matrix size, edit the constants and run `make test && make build`. All tests pass with any value.

## CLI Reference

```
Usage: moria [--magic|--pretty|--live|--password-strength] [--max-len N] [--ignore-paste] [spell]

Options:
  --magic                Generate a master password
  --pretty               Display the password matrix from your master password
  --live                 Interactive mode: type your spell and see the password build in real-time
  --password-strength    Analyze password strength from stdin (standalone, no spell)
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
