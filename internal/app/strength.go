package app

import (
	"fmt"
	"math"
	"unicode"
)

// Attack scenario speeds in guesses per second.
const (
	OnlineRateLimited = uint64(1_000)
	OfflineSlowHash   = uint64(10_000)
	OfflineFastHash   = uint64(10_000_000_000)
	GPUSupercluster   = uint64(100_000_000_000)
)

// TimeToGuess returns the average seconds to exhaust half the keyspace
// at the given guessing speed.
func TimeToGuess(entropyBits int, guessesPerSec uint64) float64 {
	if guessesPerSec == 0 {
		return math.Inf(1)
	}
	combinations := math.Exp2(float64(entropyBits) - 1)
	return combinations / float64(guessesPerSec)
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
	return fmt.Sprintf("%.1f billion times the age of the universe", universeAges/1_000_000_000)
}

// EstimateMasterEntropy estimates the entropy of the master password based on
// its length and character diversity. It counts the character classes present
// (lowercase, uppercase, digits, special) and returns length * log2(charsetSize).
func EstimateMasterEntropy(input string) int {
	if input == "" {
		return 0
	}
	charsetSize := 0
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, r := range input {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if hasLower {
		charsetSize += 26
	}
	if hasUpper {
		charsetSize += 26
	}
	if hasDigit {
		charsetSize += 10
	}
	if hasSpecial {
		charsetSize += 32
	}

	if charsetSize == 0 {
		charsetSize = 1
	}

	return int(float64(len(input)) * math.Log2(float64(charsetSize)))
}
