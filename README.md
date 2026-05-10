# moria

A deterministic, matrix-based password generator. Derive unique, reproducible passwords from a master secret and a memorable spell.

```
Master Secret + Spell → Password
```

> *"Speak, friend, and enter."*

![Moria live mode](docs/moria.gif)

![Go 1.26+](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&logoColor=white)
![MIT License](https://img.shields.io/badge/license-MIT-blue)

Inspired by [pwgen](https://www.uni-muenster.de/CERT/pwgen/index.php?lang=en&mode=pwcard) · Based on [zxcvbn](https://github.com/ccojocar/zxcvbn-go) and [Argon2id](https://en.wikipedia.org/wiki/Argon2) · Videos: [zxcvbn](https://www.youtube.com/watch?v=vf37jh3dV2I) [Argon2id](https://youtu.be/Sc3aHMCc4h0?t=114)

---

## Why moria?

**No vault to store, no cloud to trust, no sync to break.** moria derives passwords deterministically — the same master secret + spell always produces the same output. There is nothing to back up, nothing to lose, and nothing to hack.

| | moria | Password managers | `pass` / `gopass` | `lesspass` |
|---|---|---|---|---|
| **State** | None | Encrypted vault | GPG-encrypted files | None |
| **Master secret** | Any input (SSH key, passphrase, random string) | Master password | GPG key | Master password |
| **Key derivation** | Argon2id (64 MB, memory-hard) | N/A | N/A | PBKDF2 (100 rounds) |
| **Per-service passwords** | Spell → matrix path | Randomly generated | Randomly generated | Spell + counter |
| **Memory safety** | SecureBytes wiping, memguard | N/A | N/A | N/A |
| **Offline** | Yes | Varies (some need sync) | Yes | Yes |

**Key differentiators:**

- **Use your SSH key as master** — no new secret to create or remember. Your existing `id_ed25519` is your master.
- **Memory-hard key derivation** — Argon2id with 64 MB memory makes GPU/ASIC attacks impractical even on weak passphrases.
- **Secure memory wiping** — master passwords and derived secrets are zeroized in memory via `memguard`. No lingering secrets after exit.
- **Matrix-based derivation** — each character of your spell navigates a 20×10 grid, making the relationship between spell and password non-obvious.

---

## How It Works

moria expands your master secret into a **password matrix** — a 20-row × 10-column grid of random character fragments. Each cell holds 3 characters from a 73-character bash-friendly alphabet.

Your **spell** (typically a service name like `"amazon"`) navigates this grid, one cell per character:

| Spell character | Determines… | How |
|---|---|---|
| **Which row** | Position in the spell, mod 20 | `a`(0) → row 0, `m`(1) → row 1, `a`(2) → row 2, … |
| **Which column** | Letter group (case-insensitive) | A-C → col 1, D-F → col 2, G-I → col 3, …, X-Z → col 9 |
| **Uppercase shift** | Row offset by +10 | `"A"` at position 0 → row 10 instead of row 0 |
| **Non-letters** | Column 0 | Digits, spaces, and special characters always read from column 0 |

### Walkthrough: spell `"amazon"`

```
     Non  ABC  DEF  GHI  JKL  MNO  PQR  STU  VWX  YZ
     ───  ───  ───  ───  ───  ───  ───  ───  ───  ───
 0   xK9  nQ7  5$w  BcD  6gH  1lM  3pQ  5tU  7xY  9bC
 1   aR2  oP5  rT8  uW1  yX4  zA6  bE7  dF9  gH0  jK3
 2   ...  ...  ...  ...  ...  ...  ...  ...  ...  ...
```

```
Spell:  a      m      a      z      o      n
Row:    0      1      2      3      4      5
Col:    1(ABC) 6(MNO) 1(ABC) 9(YZ)  6(MNO) 6(MNO)
Cell:   nQ7    zA6    oP5    9bC    1lM    3pQ
        ───────────────────────────────────────
Password: nQ7zA6oP59bC1lM3pQ
```

Each character picks a cell; the cell contents concatenate to form the password. The same master + `"amazon"` will always trace the same path and produce the same result.

**Key properties:**

- **Deterministic** — same inputs, same output, every time. No state stored anywhere.
- **Case-sensitive rows** — `"Amazon"` (capital A) reads from row 10 at position 0, producing a different password than `"amazon"`.
- **Case-insensitive grouping** — `a` and `A` both map to column 1 (ABC). Only the row differs.
- **Length = spell length × 3** — each character contributes `CharactersPerMatrixCell` (default: 3) characters to the password.

---

## Installation

**Go install** (requires Go 1.26+):

```bash
go install github.com/kiviuk/moria@latest
```

**From source:**

```bash
git clone https://github.com/kiviuk/moria.git
cd moria
make build
```

The binary is built to `bin/moria`.

**Cross-compile:**

| Target | Command | Output |
|---|---|---|
| macOS (default) | `make build` | `bin/moria` |
| Windows x86-64 | `make win64` | `bin/win64/moria.exe` |
| Linux x86-64 | `make linux64` | `bin/linux64/moria` |

All cross-compiled binaries use `CGO_ENABLED=0` and are stripped of debug info (`-ldflags="-s -w"`).

---

## Quick Start

```bash
# Try it now — generate a master password, then derive a password for "demo":
moria --magic > /tmp/moria-key.txt && cat /tmp/moria-key.txt | moria "demo"
```

moria needs two inputs: a **master secret** (your only secret — never echoed to the terminal) and a **spell** (one per service).

### Input Flows

```bash
# ① Pipe master, type spell when prompted (most private)
cat ~/.ssh/id_ed25519 | moria | pbcopy
# Enter spell: _

# ② Pipe master, spell on argv (convenient for scripting)
cat ~/.ssh/id_ed25519 | moria "amazon" | pbcopy

# ③ Fully interactive — prompted for both (no file, no args)
moria
# Enter master password: •••••••••••••••
# Enter spell: _
```

### Choosing a Master Secret

**Option A — Use your SSH key (recommended):** no new secret to manage; your existing key is your master.

```bash
cat ~/.ssh/id_ed25519 | moria | pbcopy
```

**Option B — Generate a dedicated master password:**

```bash
moria --magic                          # prints a 600-char cryptographically secure string
moria --magic > key.txt && chmod 600 key.txt  # save it, then use it:
cat key.txt | moria | pbcopy
```

---

## Usage

### Modes

| Command | Description |
|---|---|
| `cat key \| moria "spell"` | **Batch** — output password to stdout |
| `cat key \| moria` | **Batch** — master piped, spell prompted interactively |
| `moria` | **Batch** — both master and spell prompted interactively |
| `cat key \| moria --live` | **Live** — type spell character by character, see password build in real-time |
| `cat key \| moria --pretty` | **Pretty** — display the full password matrix |
| `cat key \| moria --max-len 16 "spell"` | Truncate output to 16 characters |
| `echo "passphrase" \| moria --show-strength` | Analyze master password strength |
| `moria --magic` | Generate a new master password |

### Live Mode

Type your spell character by character. The matrix highlights visited cells and the password builds in real-time. Press Enter to copy the final password.

```bash
cat ~/.ssh/id_ed25519 | moria --live
# or fully interactive:
moria --live
```

```
Spell: amazon
Password: xK9!nQ7# (8/8)
```

> Live mode keeps the entered spell in Bubbletea's memory until GC. For maximum security, prefer batch mode.

### Display the Matrix

```bash
cat the-key.txt | moria --pretty
```

```
     Non  ABC  DEF  GHI  JKL  MNO  PQR  STU  VWX  YZ
     ───  ───  ───  ───  ───  ───  ───  ───  ───  ───
 0   xK9  nQ7  5$w  BcD  6gH  1lM  3pQ  5tU  7xY  9bC
 1   ...
```

### Limit Password Length

Some sites cap password length. Use `--max-len` to truncate:

```bash
cat the-key.txt | moria --max-len 16 "amazon"
# → xK9!nQ7#5$wYBcD4
```

### Check Master Password Strength

```bash
echo "i'm super hunger today" | moria --show-strength
```

```
zxcvbn master password entropy: 50 bits

zxcvbn crack time estimate (generic): centuries

Assuming attacker 100K guesses/sec and 50 bits (from zxcvbn), worst case: 357 years
```

---

## Security Model

### What's Secret and What's Not

| Component | Secret? | Notes |
|---|---|---|
| **Master password** | Yes | Your only secret. Compromise = total loss. |
| **Spell** | Yes | Your memorable phrase per service. Keep private. |
| **Generated password** | Until leaked | What you type to log in. Safe if master is secure. |

### How Strength Works

The strength of your derived passwords is limited by your master password. A long spell cannot compensate for a weak master.

`--show-strength` uses [zxcvbn](https://github.com/ccojocar/zxcvbn-go) for pattern detection — it recognizes dictionary words, keyboard walks, dates, and common sequences, giving a realistic entropy estimate instead of naively assuming random characters.

The **worst-case time** is calculated as `(2^entropy guesses) ÷ (guesses per second)`. moria uses **100K guesses/sec** as the attacker model because each guess requires running Argon2id with 64 MB of memory — this is memory-bandwidth bound, not compute bound, making GPU/ASIC attacks far less effective than against PBKDF2 or scrypt.

**Practical advice:** Combine common words in unique, memorable ways. Even simple phrases are safer than you think because attackers can't precompute every possible combination. zxcvbn detects dictionary words but can't detect semantic relationships between them.

### Spell Visibility

When the spell is passed as a command-line argument (`moria "amazon"`), it is visible in `ps aux`, shell history, and OS audit logs. For most use cases (service names like `"amazon"`, `"github"`) this is low risk. If your spell is sensitive, run `moria` without arguments so the spell is entered interactively — it will never appear in the process list or shell history.

### Memory Safety

moria uses [`SecureBytes`](docs/SECURE_MEMORY_WIPE.md) backed by `memguard` to ensure master passwords and derived secrets are zeroized in memory after use. The matrix cells are mutable `[]byte` (not immutable Go strings), so `Matrix.Wipe()` truly erases all data. On exit, `memguard.SafeExit()` performs final cleanup.

**Known limitation:** Displaying the matrix (`--pretty`, `--live`) temporarily creates string copies that cannot be wiped. The underlying cell data is still properly zeroized.

---

## Configuration

All matrix dimensions are compile-time constants in `internal/app/config.go`:

| Constant | Default | Description |
|---|---|---|
| `PasswordMatrixRows` | 20 | Number of rows (position modulus) |
| `CharactersPerMatrixCell` | 3 | Characters per cell (password length multiplier) |
| `AlphabetSize` | 26 | Letters in the alphabet |
| `MasterPasswordChars` | 73 chars | Bash-friendly characters for `--magic` |
| `MaxSpellLength` | 1000 | Maximum allowed spell length |
| `Argon2Salt` | `moria-argon-salt-v1` | Salt for Argon2id (must be consistent across devices) |

To change the matrix size, edit the constants and run `make test && make build`. All tests pass with any value — expected values are computed from constants, not hardcoded.

---

## CLI Reference

```
Usage: moria [--magic|--pretty|--live|--show-strength] [--max-len N] [--ignore-paste] [--] [<spell>]

Options:
  --magic             Generate a master password
  --pretty            Display the password matrix from your master password
  --live              Interactive mode: type your spell and see the password build in real-time
  --show-strength     Show strength of password from stdin (standalone mode)
  --max-len N         Truncate generated output to N characters (live and batch modes only)
  --ignore-paste      Ignore pasted input in live mode (single characters only, live mode only)
  --                  Spell separator (use before spells starting with --)
  -h, --help          Show this help message

Examples:
  moria --magic                       Generate a new master password
  moria "amazon"                      Prompted for master, spell on argv
  moria                               Prompted for master, then for spell
  cat master.txt | moria "amazon"     Master piped, spell on argv
  cat master.txt | moria              Master piped, prompted for spell
  cat master.txt | moria --pretty     Display the matrix
  cat master.txt | moria --live       Interactive live mode (paste allowed)
  cat master.txt | moria --live --ignore-paste  Live mode (paste blocked)
  cat master.txt | moria --max-len 16 "amazon"  Limited length
```

---

## Development

### Project Structure

```
moria/
├── cmd/moria/
│   ├── main.go                  # CLI entry point, input state machine, run loop
│   ├── live.go                  # Bubbletea TUI for interactive live mode
│   ├── live_test.go             # Tests for live mode model
│   ├── password_prompt.go       # Bubbletea masked password input prompt (•••)
│   ├── spell_prompt.go          # Bubbletea spell input prompt (visible text)
│   ├── debug_helpers.go         # debugf() logging (MORIA_DEBUG env var)
│   ├── messages.go              # All CLI error messages and UI strings
│   ├── main_test.go             # Tests for batch mode, flag parsing, validation
│   ├── cli_integration_test.go  # Integration tests (pipe + runCLI helper)
│   ├── cli_pty_integration_test.go  # Integration tests using MORIA_MASTER_FILE/SPELL
│   ├── read_stdin_test.go       # Tests for readStdin / size limits
│   └── test_main_test.go        # TestMain: builds binary once for all CLI tests
├── internal/
│   ├── app/
│   │   ├── config.go            # Package-level constants
│   │   ├── spell.go             # Core domain types (MagicLetter, QueryLetter, etc.)
│   │   ├── spell_test.go        # Tests for parsing, grouping, resolution
│   │   ├── password_matrix.go   # Matrix type, generation, Pretty(), Cell access
│   │   ├── password_matrix_test.go  # Matrix dimension, content, and integration tests
│   │   ├── secure_bytes.go      # SecureBytes type for in-memory secret management
│   │   ├── secure_bytes_test.go # Tests for SecureBytes
│   │   ├── strength.go          # Time-to-guess calculation and human-readable formatting
│   │   └── strength_test.go     # Tests for CrackTime, FormatSeconds, Entropy
│   └── testutil/
│       └── testutil.go          # Shared test data generator (no import cycles)
├── docs/
│   ├── moria.gif                # Live mode demo animation
│   ├── moria-live.png           # Live mode screenshot
│   └── SECURE_MEMORY_WIPE.md    # SecureBytes design and implementation details
├── .golangci.yml                # golangci-lint configuration
├── go.mod
└── Makefile
```

### Build & Test

```bash
make build                        # Compile to bin/moria
make test                         # Run all tests
make lint                         # Run golangci-lint
make clean                        # Remove bin/
go clean -testcache && make test  # Clear cache and re-run all tests
go test ./internal/app/ -v       # Verbose output for app tests
go test ./cmd/moria/ -v          # Verbose output for cmd tests
go test ./... -run TestQuery     # Run single test by name
```

All tests pass with any `CharactersPerMatrixCell` and `PasswordMatrixRows` values — expected values are computed from constants, not hardcoded.

---

## License

MIT
