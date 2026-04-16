# moria — Cryptographic Security Model

## Overview

moria derives unique passwords from a single master secret and a memorable "spell" (typically a service name). The same inputs always produce the same output. This document explains the cryptographic design, attack vectors, and why your master password is the linchpin of your entire digital life.

## How It Works

### The Algorithm

1. **Master Password → Matrix**: Your master password is deterministically expanded into a grid of random character fragments
2. **Spell → Path**: Each character in your spell determines a cell to read:
   - **Row** = character position in spell, modulo `PasswordMatrixRows` (uppercase letters shift by `PasswordMatrixRows/2`)
   - **Column** = letter group (A-C→1, D-F→2, ..., Y-Z→9, non-letters→0)
3. **Extract Password**: Concatenate the cell contents along the path

### Example

Spell: `"phrase-i-can-remember"` (21 characters, including hyphens)

| Char | Position | Row | Group | Column | Cell |
|------|----------|-----|-------|--------|------|
| p | 0 | 0 | 6 (PQR) | 6 | (0,6) |
| h | 1 | 1 | 3 (GHI) | 3 | (1,3) |
| r | 2 | 2 | 6 (PQR) | 6 | (2,6) |
| a | 3 | 3 | 1 (ABC) | 1 | (3,1) |
| s | 4 | 4 | 7 (STU) | 7 | (4,7) |
| e | 5 | 5 | 2 (DEF) | 2 | (5,2) |
| - | 6 | 6 | 0 (Non) | 0 | (6,0) |
| i | 7 | 7 | 3 (GHI) | 3 | (7,3) |
| - | 8 | 8 | 0 (Non) | 0 | (8,0) |
| c | 9 | 9 | 1 (ABC) | 1 | (9,1) |
| a | 10 | 10 | 1 (ABC) | 1 | (10,1) |
| n | 11 | 11 | 5 (MNO) | 5 | (11,5) |
| - | 12 | 12 | 0 (Non) | 0 | (12,0) |
| r | 13 | 13 | 6 (PQR) | 6 | (13,6) |
| e | 14 | 14 | 2 (DEF) | 2 | (14,2) |
| m | 15 | 15 | 5 (MNO) | 5 | (15,5) |
| e | 16 | 16 | 2 (DEF) | 2 | (16,2) |
| m | 17 | 17 | 5 (MNO) | 5 | (17,5) |
| b | 18 | 18 | 1 (ABC) | 1 | (18,1) |
| e | 19 | 19 | 2 (DEF) | 2 | (19,2) |
| r | 20 | 0 | 6 (PQR) | 6 | (0,6) |

Note: position 20 wraps to row 0 (20 mod 20 = 0), the same row as 'p'.

Output: 21 cells × 3 chars = 63-character password.

### Case Sensitivity

Uppercase letters shift the row by `PasswordMatrixRows/2`, making `"PHrase-I-can-remember"` and `"phrase-i-can-remember"` produce completely different passwords. This adds entropy without requiring a longer spell.

### Entropy

- **Matrix**: 600 chars × ~6.19 bits/char ≈ ~3,700 bits of entropy
- **Password**: For an 18-character spell, 54 chars × ~6.19 bits ≈ ~334 bits
- **Brute force**: Computationally infeasible

## The Four Pieces of the Puzzle

When you use moria, four things exist in the world:

| Piece | Example | Who Knows It |
|-------|---------|-------------|
| **Spell** | `amazon` | You (but attacker may guess it) |
| **Amazon's Database Hash** | `$2b$12$xQ...` | Attacker (stolen in breach) |
| **Generated Password** | `54Oy^L0mn2JL,S6ETv` | You (what you type to log in — derived by moria from master + spell) |
| **Master Password** | `i'm super hunger today` | You (your only secret) |

The attacker's goal: recover your **Master Password**. Why? Because it unlocks every account you've ever generated — Amazon, Gmail, PayPal, your bank, your crypto wallet.

## Two Attack Vectors

### Attack 1: The Front Door (Brute-Force the Generated Password)

The attacker steals Amazon's database and sees the hash of your login password. They try to brute-force it directly:

```
GPU guesses: a, b, c, ... 54Oy^L0mn2JL,S6ETv
```

**Why this fails:** Your generated password is 18 characters of pure random noise (108 bits of entropy). Even at 25 trillion guesses/sec (GPU cluster against MD5/SHA1), it would take **15 times the age of the universe**.

The front door is bolted shut.

### Attack 2: The Side Door (Guess the Master Password)

The attacker does reconnaissance. They discover you use moria. Now they know:

1. The spell for Amazon is likely `"amazon"` or a variant (guessable from context)
2. Humans pick weak master passwords

The attacker writes a custom script:

```
for each guess in dictionary:
    matrix = Argon2id(guess) → HKDF → 600-char matrix
    fake_password = extract(matrix, spell="amazon")
    if hash(fake_password) == stolen_amazon_hash:
        FOUND IT → guess is the master password
```

**Why this is slower:** The attacker must run Argon2id on **every single guess**. Argon2id with 64MB memory is memory-bandwidth bound, bottlenecking even a GPU cluster to ~100,000 guesses/sec.

**Result:** A 50-bit passphrase like `"i'm super hunger today"` takes **178 years** on a GPU cluster.

## Why Argon2id Is the Shield

Argon2id is a **memory-hard key derivation function**. Unlike fast hashes (MD5, SHA1), it requires significant RAM per guess:

| Parameter | Value | Purpose |
|-----------|-------|---------|
| Time cost | 1 iteration | ~500ms derivation time (OWASP minimum; t=2 would double attacker work at ~2× user latency cost) |
| Memory | 64 MB | Forces RAM bottleneck |
| Parallelism | 4 threads | Balances speed vs. security |
| Key length | 32 bytes | High-entropy output |

**The magic:** Even if your master password is weak (50 bits from dictionary words), Argon2id turns it into a mathematical wall. Without Argon2id, 50 bits would crack in **~1.5 hours** on a GPU cluster. With Argon2id, it takes **178 years**.

## The Catastrophe Scenario

If an attacker cracks your master password, they don't just get your Amazon account. They get **everything**:

```bash
# Attacker has your master password now
$ echo "i'm super hunger today" | moria "paypal"
xK9!mPaB2@cD4eF6    # Your PayPal password — cracked
$ echo "i'm super hunger today" | moria "gmail"
rT5$wY8aBcD4eF6g    # Your Gmail password — cracked
$ echo "i'm super hunger today" | moria "bank"
jK1lM2nO3pQ4rS5t    # Your bank password — cracked
```

They don't need to hack any other database. They just use moria locally with your master password and the service name.

**This is why your master password must be strong.** It's not just one account — it's your entire digital life.

## Key Derivation Pipeline

Your master password goes through a two-stage process:

```
Master Password
    ↓
Argon2id (1 iter, 64MB, 4 threads)
    ↓ 32-byte high-entropy key
HKDF-SHA256 (expand to 600 chars)
    ↓ 600 random characters
Password Matrix (20 rows × 10 cols × 3 chars/cell)
    ↓
Spell → Path through matrix
    ↓
Generated Password
```

### Stage 1: Argon2id

```go
salt := []byte("moria-argon-salt-v1")
key := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
```

Takes any input (random string, passphrase, SSH key) and produces a 32-byte high-entropy key. The 64MB memory requirement makes brute-force attacks computationally expensive.

**Static salt trade-off:** moria uses a fixed, public salt (`"moria-argon-salt-v1"`) rather than a random per-user salt. This is a deliberate design decision: a random salt would need to be stored and synchronised across devices, defeating the "no storage required" property. The primary protection is Argon2id's memory-hardness — the cost of the memory-hard function makes precomputed attacks infeasible at useful scale regardless of salt uniqueness.

### Stage 2: HKDF Expansion

```go
hkdfReader := hkdf.New(sha256.New, key, nil, []byte("moria-matrix-expansion"))
matrix := mapBytesSourceToAlphabet(hkdfReader, MasterPasswordChars, MatrixBytes)
```

Expands the 32-byte key to 600 characters using HKDF (RFC 5869). The output is deterministic — same key always produces the same matrix.

**This means you can use any input as long as it has sufficient entropy:**
- **SSH/GPG keys**: High entropy (generated with crypto/rand) — ideal
- **Strong passphrases**: Sufficient if long enough (see `--show-strength`)
- **Weak passphrases**: Risky — Argon2id slows attacks but doesn't replace missing entropy

### Stage 3: Rejection Sampling

When generating random passwords, a common mistake is to use the modulo operator (`%`) to map random bytes to a character set. This introduces **modulo bias** — some characters become slightly more likely than others, weakening the password.

Moria uses **rejection sampling** instead: if a random byte falls in the "biased" range, it's discarded and a new byte is drawn. This guarantees every character in the pool has exactly equal probability, preserving the full entropy of your passwords.

Imagine you have a 52-card deck and want to randomly pick a number from 1 to 10. If you just divide the card value by 10 and take the remainder, the numbers 1 and 2 would come up more often than the rest — because 52 doesn't divide evenly by 10, leaving 2 "extra" cards that loop back to the beginning.
Rejection sampling fixes this by saying: "If you draw one of those extra cards, put it back and draw again." You keep drawing until you get a card from the fair range. The result is that every number from 1 to 10 has exactly the same chance of being picked.

In moria's case, a random byte can be 0–255 (256 values), but the character pool has 73 characters. Since 256 doesn't always divide evenly into the pool size, some characters would be slightly more likely without rejection sampling. By discarding the "extra" bytes and drawing fresh ones, every character gets a perfectly fair shot.

```go
threshold := 256 - (256 % poolLen)
if b < threshold {
    result[i] = poolBytes[b%poolLen]  // unbiased
}
// else: discard and try next byte
```

## Entropy Analysis

### Generated Password

For a 6-letter spell:
- 6 cells × 3 chars/cell = 18 characters
- Each character from 73-char pool = ~6.19 bits
- Total: **~111 bits** (computationally infeasible to brute-force)

### Master Password

Use `--show-strength` to check your master password:

```bash
$ echo "i'm super hunger today" | moria --show-strength
zxcvbn master password entropy: 50 bits

zxcvbn crack time estimate (generic): centuries

Assuming attacker 100K guesses/sec and 50 bits (from zxcvbn), worst case: 357 years
```

**zxcvbn** detects dictionary words, patterns, and common substitutions — giving a realistic entropy estimate rather than naive `length × charset` math.

### The Nuclear Option: `--magic`

When you run `moria --magic`, the tool bypasses human psychology entirely. It goes straight to `crypto/rand` (true machine randomness) and generates a password that is exactly 600 characters long — the exact size needed to fill your 20×10×3 matrix.

```bash
$ moria --magic > master.txt
$ cat master.txt | moria --show-strength
zxcvbn master password entropy: 3346 bits

zxcvbn crack time estimate (generic): effectively uncrackable

Assuming attacker 100K guesses/sec and 3346 bits (from zxcvbn), worst case: effectively uncrackable
```

**Why did it score 3346 bits?**

When zxcvbn analyzed it, it scanned for dictionary words, patterns, and keyboard walks. It found nothing. Because it was pure, unadulterated random noise, zxcvbn calculated the entropy based on its character frequency analysis:

```
600 characters of random noise ≈ 3346 bits
```

**To put 3346 bits into perspective:**

| Scale | Bits | Comparison |
|-------|------|------------|
| Atoms in the observable universe | ~256 | `2^256` |
| Your `--magic` master password | **3346** | `2^3346` |

If every single atom in the universe was a GPU supercomputer, and they had all been guessing your master password since the Big Bang, they wouldn't have even scratched the surface.

**The Catch (And Why `--magic` Exists)**

You now have a master password that defeats the NSA, time itself, and the heat death of the universe.

The only problem? You cannot memorize a 600-character string of gibberish.

This is exactly why moria was designed to accept piped inputs:

```bash
cat master.txt | moria "amazon"
```

By using `--magic`, you are abandoning the "brain-only" approach. You save that massive 3346-bit string into a text file, put it on an encrypted USB drive (or inside a local password manager), and use it as a **keyfile**.

**You traded human convenience (memorization) for absolute, flawless mathematical invincibility.**

## Security Rules

1. **Your master password is your only secret.** Protect it like your life depends on it — because your digital life does.

2. **A long spell cannot compensate for a weak master.** `"amazon"` vs `"amazonprime"` doesn't matter if your master is `"password123"`.

3. **Use `--show-strength` to validate.** If it shows "instant" or "minutes", pick a stronger master — not a longer spell.

4. **Same master + same spell = same password.** This is deterministic by design. If you forget your master password, there is no recovery.

5. **Never reuse a master password.** If compromised, every derived password is compromised.

## What's Public vs. Secret

| Component | Visibility | Reason |
|-----------|-----------|--------|
| Master password | **Secret** | Your only secret. Compromise = total loss. |
| Spell | Hidden (but guessable) | Service name or variant. Attacker may guess it from context. |
| Matrix | Secret (derived) | Only reproducible with master password. |
| Generated password | Secret (until leaked) | What you type to log in. |
| Algorithm | Public | moria is open-source. Security through math, not obscurity. |

## Comparison to Password Managers

| Feature | moria | Traditional Password Manager |
|---------|-------|----------------------------|
| Storage | None (derived on-demand) | Encrypted database file |
| Single point of failure | Master password | Master password + database file |
| Breach impact | Only if master is cracked | Database can be attacked offline |
| Recovery | None (deterministic) | Backup files |
| Sync | Not needed (same output everywhere) | Requires cloud sync or manual transfer |
| Trust model | Math (open-source) | Vendor's encryption implementation |

## Recommendations

1. **Generate a random master password** with `moria --magic` and store it securely
2. **Check its strength** with `moria --show-strength`
3. **Use memorable but strong passphrases** if you must type it manually (e.g., 5-6 random words)
4. **Never use dictionary phrases** that zxcvbn can detect
5. **Back up your master password** — there is no recovery if lost
