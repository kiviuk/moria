package app

import (
	"math"
	"testing"
)

func TestCrackTime_Basic(t *testing.T) {
	// Verify time-to-guess calculation for known entropy and speed
	tests := []struct {
		entropyBits     int
		guessesPerSec   uint64
		expectedSeconds float64
	}{
		{36, 1_000, math.Exp2(35) / 1_000},
		{36, 10_000, math.Exp2(35) / 10_000},
		{108, 100_000_000_000, math.Exp2(107) / 100_000_000_000},
		{0, 1_000, 0.5 / 1_000},
	}

	for _, tt := range tests {
		got := TimeToGuess(tt.entropyBits, tt.guessesPerSec)
		if math.Abs(got-tt.expectedSeconds) > tt.expectedSeconds*0.001 {
			t.Errorf("CrackTime(%d, %d) = %f, expected %f", tt.entropyBits, tt.guessesPerSec, got, tt.expectedSeconds)
		}
	}
}

func TestCrackTime_ZeroSpeed(t *testing.T) {
	// Verify zero speed returns infinity
	got := TimeToGuess(64, 0)
	if !math.IsInf(got, 1) {
		t.Errorf("CrackTime(64, 0) = %f, expected +Inf", got)
	}
}

func TestFormatSeconds_Instant(t *testing.T) {
	// Verify sub-second durations return "instant"
	tests := []float64{0, 0.001, 0.5, 0.99}
	for _, s := range tests {
		got := FormatSeconds(s)
		if got != "instant" {
			t.Errorf("FormatSeconds(%f) = %q, expected %q", s, got, "instant")
		}
	}
}

func TestFormatSeconds_Seconds(t *testing.T) {
	// Verify seconds range formats correctly
	got := FormatSeconds(30)
	if got != "30 seconds" {
		t.Errorf("FormatSeconds(30) = %q, expected %q", got, "30 seconds")
	}
}

func TestFormatSeconds_Minutes(t *testing.T) {
	// Verify minutes range formats correctly
	got := FormatSeconds(120)
	if got != "2 minutes" {
		t.Errorf("FormatSeconds(120) = %q, expected %q", got, "2 minutes")
	}
}

func TestFormatSeconds_Hours(t *testing.T) {
	// Verify hours range formats correctly
	got := FormatSeconds(7200)
	if got != "2 hours" {
		t.Errorf("FormatSeconds(7200) = %q, expected %q", got, "2 hours")
	}
}

func TestFormatSeconds_Days(t *testing.T) {
	// Verify days range formats correctly
	got := FormatSeconds(86400 * 5)
	if got != "5 days" {
		t.Errorf("FormatSeconds(432000) = %q, expected %q", got, "5 days")
	}
}

func TestFormatSeconds_Years(t *testing.T) {
	// Verify years range formats correctly
	got := FormatSeconds(365.25 * 86400 * 10)
	if got != "10 years" {
		t.Errorf("FormatSeconds(10 years) = %q, expected %q", got, "10 years")
	}
}

func TestFormatSeconds_LargeNumbers(t *testing.T) {
	// Verify large year values use magnitude suffixes or universe age multiples
	const year = 365.25 * 86400.0
	tests := []struct {
		seconds  float64
		expected string
	}{
		{year * 500_000, "500.0 thousand years"},
		{year * 5_000_000, "5.0 million years"},
		{year * 5_000_000_000, "5.0 billion years"},
		{year * 13_800_000_000, "1 times the age of the universe"},
		{year * 138_000_000_000, "10 times the age of the universe"},
		{year * 13_800_000_000_000, "1.0 thousand times the age of the universe"},
		{year * 13_800_000_000_000_000, "1.0 million times the age of the universe"},
		{year * 13_800_000_000_000_000_000, "1.0 billion times the age of the universe"},
	}

	for _, tt := range tests {
		got := FormatSeconds(tt.seconds)
		if got != tt.expected {
			t.Errorf("FormatSeconds(%.0f) = %q, expected %q", tt.seconds, got, tt.expected)
		}
	}
}

func TestMagicSpell_Entropy(t *testing.T) {
	// Verify entropy calculation matches spell length × cell size × bits per char
	tests := []struct {
		spell    string
		expected int
	}{
		{"a", CharactersPerMatrixCell * CharsetBits},
		{"ab", 2 * CharactersPerMatrixCell * CharsetBits},
		{"amazon", 6 * CharactersPerMatrixCell * CharsetBits},
		{"", 0},
	}

	for _, tt := range tests {
		m := MagicSpell{Spell: tt.spell}
		got := m.Entropy()
		if got != tt.expected {
			t.Errorf("MagicSpell(%q).Entropy() = %d, expected %d", tt.spell, got, tt.expected)
		}
	}
}

func TestEstimateMasterEntropy(t *testing.T) {
	// Verify master password entropy estimation based on length and character diversity
	tests := []struct {
		input   string
		minBits int
		maxBits int
	}{
		{"", 0, 0},
		{"a", 4, 5},
		{"password", 37, 38},
		{"Password1", 53, 54},
		{"P@ssw0rd!", 58, 59},
		{MasterPasswordChars[:64], 418, 420},
	}

	for _, tt := range tests {
		got := EstimateMasterEntropy(tt.input)
		if got < tt.minBits || got > tt.maxBits {
			t.Errorf("EstimateMasterEntropy(%q) = %d, expected %d-%d", tt.input, got, tt.minBits, tt.maxBits)
		}
	}
}
