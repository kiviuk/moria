# moria — Cryptographic Security Model

## Overview

moria derives unique passwords from a single master secret and a memorable "spell" (typically a service name). The same inputs always produce the same output. This document explains the cryptographic design, attack vectors, and why your master password is the linchpin of your entire digital life.

## The Four Pieces of the Puzzle

When you use moria, four things exist in the world:

| Piece | Example | Who Knows It |
|-------|---------|-------------|
| **Spell** | `amazon` | Public (it's the website name) |
| **Amazon's Database Hash** | `$2b$12$xQ...` | Attacker (stolen in breach) |
| **Generated Password** | `54Oy^L0mn2JL,S6ETv` | You (what you type to log in) |
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

1. The spell for Amazon is almost certainly `"amazon"` (public knowledge)
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
| Time cost | 1 iteration | ~500ms derivation time |
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
salt := []byte("moria-salt-v1")
key := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
```

Takes any input (random string, passphrase, SSH key) and produces a 32-byte high-entropy key. The 64MB memory requirement makes brute-force attacks computationally expensive.

### Stage 2: HKDF Expansion

```go
hkdfReader := hkdf.New(sha256.New, key, salt, []byte("moria-matrix-expansion"))
matrix := mapToCharset(hkdfReader, MasterPasswordChars, MatrixBytes)
```

Expands the 32-byte key to 600 characters using HKDF (RFC 5869). The output is deterministic — same key always produces the same matrix.

### Stage 3: Rejection Sampling

```go
threshold := 256 - (256 % poolLen)
if b < threshold {
    result[i] = poolBytes[b%poolLen]  // unbiased
}
// else: discard and try next byte
```

Maps random bytes to the character pool without modulo bias. Every character has exactly equal probability.

## Entropy Analysis

### Generated Password

For a 6-letter spell:
- 6 cells × 3 chars/cell = 18 characters
- Each character from 70-char pool = ~6.1 bits
- Total: **~108 bits** (computationally infeasible to brute-force)

### Master Password

Use `--master-password-strength` to check your master password:

```bash
$ echo "i'm super hunger today" | moria --master-password-strength
Master password entropy: 50 bits
zxcvbn crack time estimate (generic): centuries
Time to guess (master password, via Argon2id):
  Single CPU               178 years
  Single GPU               65 days
  GPU cluster              178 years
```

**zxcvbn** detects dictionary words, patterns, and common substitutions — giving a realistic entropy estimate rather than naive `length × charset` math.

## Security Rules

1. **Your master password is your only secret.** Protect it like your life depends on it — because your digital life does.

2. **A long spell cannot compensate for a weak master.** `"amazon"` vs `"amazonprime"` doesn't matter if your master is `"password123"`.

3. **Use `--master-password-strength` to validate.** If it shows "instant" or "minutes", pick a stronger master — not a longer spell.

4. **Same master + same spell = same password.** This is deterministic by design. If you forget your master password, there is no recovery.

5. **Never reuse a master password.** If compromised, every derived password is compromised.

## What's Public vs. Secret

| Component | Visibility | Reason |
|-----------|-----------|--------|
| Master password | **Secret** | Your only secret. Compromise = total loss. |
| Spell | Public | Service name (e.g., "amazon"). Known to attacker. |
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
2. **Check its strength** with `moria --master-password-strength`
3. **Use memorable but strong passphrases** if you must type it manually (e.g., 5-6 random words)
4. **Never use dictionary phrases** that zxcvbn can detect
5. **Back up your master password** — there is no recovery if lost
