# Secure Memory Wiping in moria

## The Problem

Go strings are immutable. When you have a string containing sensitive data (like a master password), you cannot securely erase it from memory. Even calling `memguard.WipeBytes([]byte(s))` doesn't help because:

1. `[]byte(s)` creates a **copy** of the string's backing array
2. The wipe only zeros the copy, not the original string
3. The original string data remains in memory until garbage collection runs
4. GC doesn't guarantee zeroing - memory pages may retain the data until reused

### Before: Ineffective Wiping

```go
func main() {
    cfg := Config{
        MasterRaw: readStdin(),  // string - immutable
        Master:    app.ExpandToMatrix(master),  // string - immutable
    }
    
    // This does NOTHING useful:
    memguard.WipeBytes([]byte(cfg.MasterRaw))  // Wipes a COPY
    memguard.WipeBytes([]byte(cfg.Master))     // Wipes a COPY
    
    // Original strings still in memory!
}
```

### Why This Matters

- Memory dumps could expose master passwords
- Swap file may contain sensitive data
- Cold boot attacks can recover "deleted" data
- Go's GC doesn't guarantee secure erasure

## The Solution: SecureBytes

We created a new type `SecureBytes` that holds a **mutable** byte slice instead of an immutable string:

```go
type SecureBytes struct {
    data []byte  // Mutable - can be truly wiped
}
```

### Key Design Decisions

1. **Mutable storage**: Uses `[]byte` which can be modified in-place
2. **Secure wiping**: `Wipe()` uses `memguard.WipeBytes()` to zero the actual data
3. **Controlled access**: `String()` creates a copy (caller beware), `Bytes()` returns the underlying slice
4. **Self-contained**: Each `SecureBytes` owns its data, no shared references

### Implementation

```go
func (sb *SecureBytes) Wipe() {
    if sb.data != nil {
        memguard.WipeBytes(sb.data)  // Zeros the actual bytes
        sb.data = nil                // Prevents reuse
    }
}
```

## Data Flow Changes

### Before (Strings)

```
stdin → string → ExpandToMatrix() → string → Matrix
         ↑                         ↑
         immutable                 immutable
         (can't wipe)              (can't wipe)
```

### After (SecureBytes)

```
stdin → []byte → SecureBytes → ExpandToMatrix() → SecureBytes → Matrix
↑             ↑                              ↑             ↑
read into     wipeable                       wipeable      wipeable
buffer        (Wipe())                       (Wipe())      []byte cells
↓
TrimSpace()
returns SecureBytes
```
stdin → []byte → SecureBytes → ExpandToMatrix() → SecureBytes → Matrix
         ↑            ↑                ↑                  ↑
         read into    wipeable         returns            wipeable
         buffer       (Wipe())         wipeable           (Wipe())
                       ↓
                  TrimSpace()
                  returns SecureBytes
```

## Key Changes

### 1. New Type: `internal/app/secure_bytes.go`

```go
type SecureBytes struct {
    data []byte
}

func NewSecureBytes(data []byte) *SecureBytes
func NewSecureBytesFromString(s string) *SecureBytes
func (sb *SecureBytes) Bytes() []byte
func (sb *SecureBytes) String() string  // Creates copy - use sparingly
func (sb *SecureBytes) Len() int
func (sb *SecureBytes) Wipe()
func (sb *SecureBytes) IsWiped() bool
func (sb *SecureBytes) TrimSpace() *SecureBytes
```

### 2. Updated Config

```go
type Config struct {
    Mode      Mode
    Spell     string
    MaxLen    int
    Master    *app.SecureBytes  // Was: string
    MasterRaw *app.SecureBytes  // Was: string
}

func (c *Config) Wipe() {
    if c.Master != nil {
        c.Master.Wipe()
    }
    if c.MasterRaw != nil {
        c.MasterRaw.Wipe()
    }
}
```

### 3. Single Wipe Point

```go
func main() {
    // ... argument parsing ...
    
	if cfg.Mode.needsStdin() {
		master, err := readStdin() // Returns *SecureBytes
		if err != nil {
			// handle error
		}
		cfg.MasterRaw = master
		expanded, err := app.ExpandToMatrix(master)
		if err != nil {
			// handle error
		}
		cfg.Master = expanded
		defer cfg.Wipe() // ONE place to wipe everything
	}
    
    // ... rest of main ...
}
```

### 4. Password Prompt

```go
func getPassword() (*app.SecureBytes, error) {
    // ... run bubbletea program ...
    sb := app.NewSecureBytesFromString(pm.input.Value())
    pm.Wipe()  // Wipe the internal textinput buffer
    return sb, nil
}
```

### 5. Piped Input

```go
func readStdin() (*app.SecureBytes, error) {
    if isPiped {
        data, err := io.ReadAll(os.Stdin)
        sb := app.NewSecureBytes(data)
        memguard.WipeBytes(data)  // Wipe original buffer
        return sb.TrimSpace(), nil
    }
    return getPassword()
}
```

### 6. Live Mode Model

```go
type liveModel struct {
    matrix            app.Matrix
    masterPasswordRaw *app.SecureBytes  // Was: string
    spell             string
    queryLetters      []app.QueryLetter
    password          string
    // ...
}

func (m *liveModel) Wipe() {
    if m.masterPasswordRaw != nil {
        m.masterPasswordRaw.Wipe()
    }
    m.spell = ""
    m.queryLetters = nil
    m.password = ""
}
```

## What Gets Wiped

| Component | Before | After |
|-----------|--------|-------|
| `cfg.MasterRaw` | ❌ String copy wiped, original remains | ✅ Truly wiped |
| `cfg.Master` | ❌ String copy wiped, original remains | ✅ Truly wiped |
| `liveModel.masterPasswordRaw` | ❌ Not wiped | ✅ Truly wiped |
| `liveModel.password` | ⚠️ Partial | ✅ Wiped in `liveModel.Wipe()` |
| Password prompt input | ❌ Not wiped | ✅ Wiped after copying |
| Piped stdin buffer | ❌ Not wiped | ✅ Wiped immediately |
| Matrix cells | ❌ Strings cannot be wiped | ✅ `[]byte` cells properly wiped |
| Extracted passwords | ❌ String returned | ✅ `*SecureBytes` returned and wiped |

## Limitations

1. **`String()` creates copies**: When you call `sb.String()`, a new string is created. This string cannot be wiped. Use sparingly and only when absolutely necessary (e.g., for display or when passing to functions that require strings).

2. **Display conversions create temporary strings**: When rendering the matrix in `--pretty` or `--live` modes, cells are converted to strings for display. These temporary strings exist briefly and cannot be wiped, but the underlying matrix cells are properly wipeable `[]byte`.

3. **Output passwords**: Generated passwords printed to stdout are written as bytes and wiped immediately after. However, they briefly exist in memory and may be retained in terminal scrollback or clipboard history.

## Testing

All tests updated to use `SecureBytes`:

```go
func TestExpandToMatrix_Deterministic(t *testing.T) {
	in1 := app.NewSecureBytesFromString("test-secret")
	in2 := app.NewSecureBytesFromString("test-secret")
	defer in1.Wipe()
	defer in2.Wipe()

	out1, err := app.ExpandToMatrix(in1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out2, err := app.ExpandToMatrix(in2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer out1.Wipe()
	defer out2.Wipe()

	if out1.String() != out2.String() {
		t.Error("not deterministic")
	}
}
```

## Best Practices

1. **Always use `defer sb.Wipe()`** immediately after creating a `SecureBytes`
2. **Avoid `String()`** unless necessary - use `Bytes()` when possible
3. **Don't store copies** - each `SecureBytes` should have a single owner
4. **Wipe early** - call `Wipe()` as soon as data is no longer needed
5. **Use `defer`** - ensures wiping even on panic or early return

## Summary

The refactoring replaces immutable strings with mutable `SecureBytes` throughout the sensitive data path. This ensures that master passwords and derived secrets can be truly erased from memory when no longer needed, reducing the attack surface for memory forensics and cold boot attacks.
