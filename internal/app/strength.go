package app

import (
	"fmt"
	"math"

	"github.com/ccojocar/zxcvbn-go"
)

// Master password attack speeds (attacker must run Argon2id + HKDF per guess).
// Argon2id with 64MB memory is memory-bandwidth bound, not compute bound.
const MasterPasswordGPUCluster = uint64(100_000)

// Uncrackable label returned when entropy exceeds displayable range.
const Uncrackable = "effectively uncrackable"

// TimeToGuess returns the average seconds to exhaust half the keyspace
// at the given guessing speed. Uses log-based calculation to avoid float64 overflow.
func TimeToGuess(entropyBits int, guessesPerSec uint64) float64 {
	if guessesPerSec == 0 {
		return math.Inf(1)
	}
	if entropyBits <= 0 {
		return 0
	}
	log2Combinations := float64(entropyBits)
	log2Speed := math.Log2(float64(guessesPerSec))
	log2Seconds := log2Combinations - log2Speed
	if log2Seconds > 1023 {
		return math.Inf(1)
	}
	return math.Exp2(log2Seconds)
}

// FormatSeconds returns a human-readable duration string.
func FormatSeconds(seconds float64) string {
	if seconds < 1 {
		return "instant"
	}

	const (
		minute = 60.0
		hour   = 60.0 * minute
		day    = 24.0 * hour
		year   = 365.25 * day
	)

	if seconds < minute {
		return fmt.Sprintf("%.0f seconds", seconds)
	}
	if seconds < hour {
		return fmt.Sprintf("%.0f minutes", seconds/minute)
	}
	if seconds < day {
		return fmt.Sprintf("%.0f hours", seconds/hour)
	}
	if seconds < year {
		return fmt.Sprintf("%.0f days", seconds/day)
	}

	years := seconds / year
	if years < 1_000 {
		return fmt.Sprintf("%.0f years", years)
	}
	if years < 1_000_000 {
		return fmt.Sprintf("%.1f thousand years", years/1_000)
	}
	if years < 1_000_000_000 {
		return fmt.Sprintf("%.1f million years", years/1_000_000)
	}

	const ageOfUniverse = 13.8e9
	universeAges := years / ageOfUniverse
	if universeAges < 1 {
		return fmt.Sprintf("%.1f billion years", years/1_000_000_000)
	}
	if universeAges < 1_000 {
		return fmt.Sprintf("%.0f times the age of the universe", universeAges)
	}
	if universeAges < 1_000_000 {
		return fmt.Sprintf("%.1f thousand times the age of the universe", universeAges/1_000)
	}
	if universeAges < 1_000_000_000 {
		return fmt.Sprintf("%.1f million times the age of the universe", universeAges/1_000_000)
	}
	if math.IsInf(universeAges, 1) || universeAges > 1e15 {
		return Uncrackable
	}
	return fmt.Sprintf("%.1f billion times the age of the universe", universeAges/1_000_000_000)
}

// MasterPasswordResult contains the strength analysis from zxcvbn.
type MasterPasswordResult struct {
	Entropy          int
	CrackTimeDisplay string
	CrackTimeSeconds float64
	Score            int
}

// CalculateMasterPasswordStrength returns detailed strength analysis from zxcvbn.
// NOTE: This function must convert to string for zxcvbn, which creates an immutable copy.
// The caller should wipe the original []byte after calling this function.
func CalculateMasterPasswordStrength(input []byte) MasterPasswordResult {
	if len(input) == 0 {
		return MasterPasswordResult{}
	}
	// zxcvbn requires a string - this creates an immutable copy that cannot be wiped
	match := zxcvbn.PasswordStrength(string(input), nil)
	entropy := match.Entropy
	if entropy < 0 {
		entropy = 0
	}
	return MasterPasswordResult{
		Entropy:          int(entropy),
		CrackTimeDisplay: match.CrackTimeDisplay,
		CrackTimeSeconds: match.CrackTime,
		Score:            match.Score,
	}
}
