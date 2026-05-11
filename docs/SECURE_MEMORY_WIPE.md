# Secure Memory in moria

## What's Protected and Why

When moria exits, the OS reclaims all virtual memory pages and zeroes them before handing them to the next process — so nothing leaks between processes regardless of wiping. The real threat is **swap / hibernation**: if a memory page is written to disk before the process has a chance to clean up, it stays on disk until that swap slot is overwritten.

moria addresses this by keeping two sensitive values in **mlock'd memory** via [`memguard`](https://github.com/awnumar/memguard):

| Value | Why mlock'd |
|-------|------------|
| Master password (raw input) | Read from stdin; held until matrix is derived |
| Extracted password (final output) | Built from matrix cells; held until written to stdout |

mlock'd pages are pinned in RAM — the OS will never write them to swap or include them in a hibernation image.

## SecureBytes

`SecureBytes` (in `internal/app/secure_bytes.go`) wraps a `*memguard.LockedBuffer`:

```go
type SecureBytes struct {
    buf *memguard.LockedBuffer
}
```

Key operations:

```go
func NewSecureBytes(data []byte) *SecureBytes        // copies into locked memory
func NewSecureBytesFromString(s string) *SecureBytes // same, from string
func (sb *SecureBytes) Bytes() []byte                // direct reference — do not retain after Wipe
func (sb *SecureBytes) String() string               // returns a plain Go string copy (cannot be wiped)
func (sb *SecureBytes) Wipe()                        // zeroes and deallocates the locked buffer
func (sb *SecureBytes) TrimSpace() *SecureBytes      // returns new SecureBytes, wipes original
```

`memguard.SafeExit()` in `main()` destroys all live `LockedBuffer`s on exit as a final safety net.

## What's NOT Protected

The following live on the plain Go heap and are not mlock'd:

- **Password matrix** (`Matrix` — `[20][10][]byte`) — derived from the master password but not individually locked. The Argon2 output key (32 bytes) is wiped immediately after use with `defer memguard.WipeBytes(key)`.
- **Spell bytes** in live mode (`LiveModel.spell []byte`) — per-keystroke heap allocations.
- **Go strings** — `cfg.Spell`, `MagicLetter.Letter`, rendered TUI strings — immutable by design; cannot be zeroed.

For these, the OS-level page zeroing on reuse provides the protection after exit. The risk window is only during process lifetime, specifically if a hibernation event occurs while moria is running.

## Limitations

1. **Go strings cannot be zeroed.** Any value that passes through a `string` at any point leaves an unzeroable copy. This affects spell input, individual letter values during parsing, and all TUI rendering.
2. **Bubbletea value-copy model.** `LiveModel` is copied by value on every keystroke. Each copy holds the current spell and password bytes; only the final copy is available when the program exits.
3. **`String()` creates an unzeroable copy.** Call it only when a string is unavoidable (e.g., final stdout write). Use `Bytes()` everywhere else.
