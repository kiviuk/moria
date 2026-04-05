package app

import (
	"fmt"
	"math"

	"github.com/nbutton23/zxcvbn-go"
)

// Attack scenario speeds in guesses per second.

// Generated password attack speeds (attacker has the leaked password hash).
const (
	OnlineRateLimited = uint64(1_000)
	OfflineSlowHash   = uint64(10_000)
	OfflineFastHash   = uint64(100_000_000_000)
	GPUSupercluster   = uint64(25_000_000_000_000)
)

// Master password attack speeds (attacker must run Argon2id + HKDF per guess).
// Argon2id with 64MB memory is memory-bandwidth bound, not compute bound.
const (
	MasterPasswordSingleCPU  = uint64(10)
	MasterPasswordGPUSingle  = uint64(10_000)
	MasterPasswordGPUCluster = uint64(100_000)
)

// Uncrackable labels returned when entropy exceeds displayable range.
const (
	UncrackableCompact = "uncrackable"
	Uncrackable        = "effectively uncrackable"
)

// TimeToGuess returns the average seconds to exhaust half the keyspace
// at the given guessing speed. Uses log-based calculation to avoid float64 overflow.
func TimeToGuess(entropyBits int, guessesPerSec uint64) float64 {
	if guessesPerSec == 0 {
		return math.Inf(1)
	}
	if entropyBits <= 0 {
		return 0
	}
	log2Combinations := float64(entropyBits) - 1
	log2Speed := math.Log2(float64(guessesPerSec))
	log2Seconds := log2Combinations - log2Speed
	if log2Seconds > 1023 {
		return math.Inf(1)
	}
	return math.Exp2(log2Seconds)
}

// FormatSecondsCompact returns a short human-readable duration for TUI display.
func FormatSecondsCompact(seconds float64) string {
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
		return fmt.Sprintf("%.0f sec", seconds)
	}
	if seconds < hour {
		return fmt.Sprintf("%.0f min", seconds/minute)
	}
	if seconds < day {
		return fmt.Sprintf("%.0f hrs", seconds/hour)
	}
	if seconds < year {
		return fmt.Sprintf("%.0f days", seconds/day)
	}

	years := seconds / year
	if years < 1_000 {
		return fmt.Sprintf("%.0f yrs", years)
	}
	if years < 1_000_000 {
		return fmt.Sprintf("%.1fK yrs", years/1_000)
	}
	if years < 1_000_000_000 {
		return fmt.Sprintf("%.1fM yrs", years/1_000_000)
	}
	if years < 1_000_000_000_000 {
		return fmt.Sprintf("%.1fB yrs", years/1_000_000_000)
	}
	if years < 1_000_000_000_000_000 {
		return fmt.Sprintf("%.1fT yrs", years/1_000_000_000_000)
	}

	const ageOfUniverse = 13.8e9
	universeAges := years / ageOfUniverse
	if math.IsInf(universeAges, 1) || universeAges > 1e15 {
		return UncrackableCompact
	}
	if universeAges < 1_000 {
		return fmt.Sprintf("%.0f x universe age", universeAges)
	}
	if universeAges < 1_000_000 {
		return fmt.Sprintf("%.1fK x universe age", universeAges/1_000)
	}
	if universeAges < 1_000_000_000 {
		return fmt.Sprintf("%.1fM x universe age", universeAges/1_000_000)
	}
	if universeAges < 1_000_000_000_000 {
		return fmt.Sprintf("%.1fB x universe age", universeAges/1_000_000_000)
	}
	return fmt.Sprintf("%.1fT x universe age", universeAges/1_000_000_000_000)
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

// EstimateMasterEntropy uses zxcvbn to evaluate human-chosen passwords.
// zxcvbn detects dictionary words, patterns, common substitutions, and keyboard walks,
// providing a realistic entropy estimate rather than naive length × charset math.
func EstimateMasterEntropy(input string) int {
	if input == "" {
		return 0
	}
	match := zxcvbn.PasswordStrength(input, nil)
	entropy := match.Entropy
	if entropy < 0 {
		return 0
	}
	return int(entropy)
}
