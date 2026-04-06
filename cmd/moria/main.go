// Package main implements the moria CLI — a deterministic, matrix-based password generator.
//
// moria derives unique passwords from a master secret and a memorable "spell"
// (typically a service name). The same inputs always produce the same output.
//
// Usage:
//
//	moria --magic                          # Generate a master password
//	moria "amazon"                         # Generate password for Amazon
//	cat master.txt | moria "amazon"        # Piped from password manager
//	cat master.txt | moria --pretty        # Display the matrix
//	cat master.txt | moria --live          # Interactive mode
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kiviuk/moria/internal/app"
)

// Mode represents the CLI operation mode.
type Mode int

const (
	// ModeBatch generates a password from a spell and master password.
	ModeBatch Mode = iota
	// ModeMagic generates a new random master password.
	ModeMagic
	// ModePretty displays the password matrix in human-readable form.
	ModePretty
	// ModeLive runs an interactive TUI for building passwords character by character.
	ModeLive
	// ModePasswordStrength analyzes the strength of a password from stdin.
	ModePasswordStrength
)

// String returns the human-readable name of the mode.
func (m Mode) String() string {
	return [...]string{"batch", "magic", "pretty", "live", "password-strength"}[m]
}

// needsStdin reports whether this mode requires reading a master password from stdin.
func (m Mode) needsStdin() bool {
	return m == ModePretty || m == ModeLive || m == ModeBatch || m == ModePasswordStrength
}

// needsSpell reports whether this mode requires a spell argument.
func (m Mode) needsSpell() bool {
	return m == ModeBatch
}

// allowedMods returns the list of additional flags permitted in this mode.
func (m Mode) allowedMods() []string {
	switch m {
	case ModeLive:
		return []string{"--max-len", "--ignore-paste"}
	case ModeBatch:
		return []string{"--max-len"}
	case ModePasswordStrength:
		return []string{}
	default:
		return nil
	}
}

// Config holds the parsed CLI configuration after argument parsing and validation.
type Config struct {
	// Mode is the selected operation mode.
	Mode Mode
	// Spell is the service name or memorable string for password derivation.
	Spell string
	// MaxLen limits the output password length (0 means no limit).
	MaxLen int
	// Master is the expanded master password string used to build the matrix.
	Master string
	// MasterRaw is the original unexpanded master input for entropy estimation.
	MasterRaw string
}

// contains reports whether slice contains item.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getMatrix builds a Matrix from the expanded master password string.
func getMatrix(master string) (app.Matrix, error) {
	matrix, err := app.NewMatrix(master)
	if err != nil {
		return app.Matrix{}, fmt.Errorf(ErrFailedCreateMatrix, err)
	}
	return matrix, nil
}

// truncatePassword returns password truncated to maxLen characters.
// If maxLen is 0 or password is shorter than maxLen, the original is returned.
func truncatePassword(password string, maxLen int) string {
	if maxLen > 0 && len(password) > maxLen {
		return password[:maxLen]
	}
	return password
}

// readStdin reads the master password from stdin. If stdin is a terminal (not piped),
// it falls back to an interactive masked password prompt.
func readStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("could not stat stdin: %w", err)
	}
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	if isPiped {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf(ErrFailedReadMaster, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return getPassword()
}

// parseArgs converts raw CLI arguments into a Config and a set of recognized flags.
func parseArgs(args []string) (Config, map[string]bool, error) {
	cfg := Config{Mode: ModeBatch}
	flags := make(map[string]bool)
	var positional []string

	for i, arg := range args {
		switch arg {
		case "--magic":
			flags["--magic"] = true
			cfg.Mode = ModeMagic
		case "--pretty":
			flags["--pretty"] = true
			cfg.Mode = ModePretty
		case "--live":
			flags["--live"] = true
			cfg.Mode = ModeLive
		case "--max-len":
			flags["--max-len"] = true
			if i+1 >= len(args) {
				return cfg, flags, fmt.Errorf("%s", ErrMaxLenRequiresValue)
			}
			i++
			val, err := strconv.Atoi(args[i])
			if err != nil {
				return cfg, flags, fmt.Errorf("%s", ErrMaxLenNotNumber)
			}
			if val <= 0 {
				return cfg, flags, fmt.Errorf("%s", ErrMaxLenNotNumber)
			}
			cfg.MaxLen = val
		case "--ignore-paste":
			flags["--ignore-paste"] = true
		case "--password-strength":
			flags["--password-strength"] = true
			cfg.Mode = ModePasswordStrength
		case "--help", "-h":
			flags["--help"] = true
		default:
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		cfg.Spell = positional[0]
	}

	return cfg, flags, nil
}

// validateConfig checks that the parsed configuration is consistent with the selected mode.
func validateConfig(cfg Config, flags map[string]bool) error {
	for flag, present := range flags {
		if !present {
			continue
		}
		if flag == "--magic" || flag == "--pretty" || flag == "--live" || flag == "--help" || flag == "--password-strength" {
			continue
		}
		if !contains(cfg.Mode.allowedMods(), flag) {
			return fmt.Errorf(ErrModNotAllowed, flag, cfg.Mode)
		}
	}

	if cfg.Mode == ModePasswordStrength && cfg.Spell != "" {
		return fmt.Errorf("%s", ErrPasswordStrengthNoSpell)
	}

	if cfg.Mode.needsSpell() && cfg.Spell == "" {
		return fmt.Errorf(ErrSpellRequired, cfg.Mode)
	}

	return nil
}

// printUsage displays the CLI help message.
func printUsage() {
	fmt.Println(MsgUsageTitle)
	fmt.Println()
	fmt.Println(MsgUsageHeader)
	fmt.Println()
	fmt.Println(MsgUsageOptions)
	fmt.Println(MsgOptMagic)
	fmt.Println(MsgOptPretty)
	fmt.Println(MsgOptLive)
	fmt.Println(MsgOptMaxLen)
	fmt.Println(MsgOptIgnorePaste)
	fmt.Println(MsgOptPasswordStrength)
	fmt.Println(MsgOptHelp)
	fmt.Println()
	fmt.Println(MsgUsageExamples)
	fmt.Println(MsgExMagic)
	fmt.Println(MsgExSpell)
	fmt.Println(MsgExPipe)
	fmt.Println(MsgExPretty)
	fmt.Println(MsgExLive)
	fmt.Println(MsgExLiveIgnorePaste)
	fmt.Println(MsgExMaxLen)
	fmt.Println(MsgExPasswordStrength)
}

// printStrengthTable outputs time-to-guess estimates to stderr.
func printStrengthTable(masterResult app.MasterPasswordResult) {
	if masterResult.Entropy == 0 {
		return
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, MsgMasterEntropy, masterResult.Entropy)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, MsgZxcvbnCrackTime, masterResult.CrackTimeDisplay)

	fmt.Fprint(os.Stderr, MsgTimeToGuessMaster)
	masterScenarios := []struct {
		label string
		speed uint64
	}{
		{"Single CPU", app.MasterPasswordSingleCPU},
		{"Single GPU", app.MasterPasswordGPUSingle},
		{"GPU cluster", app.MasterPasswordGPUCluster},
	}
	for _, s := range masterScenarios {
		seconds := app.TimeToGuess(masterResult.Entropy, s.speed)
		fmt.Fprintf(os.Stderr, MsgStrengthTableRow, s.label, app.FormatSeconds(seconds))
	}
}

func main() { //nolint:gocyclo // main has high complexity due to mode switching
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cfg, flags, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
		os.Exit(1)
	}

	if flags["--help"] {
		printUsage()
		os.Exit(0)
	}

	if err := validateConfig(cfg, flags); err != nil {
		fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
		os.Exit(1)
	}

	if cfg.Mode.needsStdin() {
		master, err := readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
			os.Exit(1)
		}
		cfg.MasterRaw = master
		cfg.Master = app.ExpandToMatrix(master)
	}

	switch cfg.Mode {
	case ModeMagic:
		master, err := app.GenerateMasterPassword(app.MatrixBytes, app.MasterPasswordChars)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrFailedGenerateMaster+"\n", err)
			os.Exit(1)
		}
		fmt.Print(master)

	case ModePretty:
		matrix, err := getMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(matrix.Pretty())

	case ModeLive:
		matrix, err := getMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		pasteMode := PasteAllowed
		if flags["--ignore-paste"] {
			pasteMode = PasteIgnored
		}
		finalModel, err := LiveMode(matrix, cfg.MaxLen, pasteMode, cfg.MasterRaw)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrLiveMode+"\n", err)
			os.Exit(1)
		}
		password := finalModel.password
		password = truncatePassword(password, cfg.MaxLen)
		if password != "" {
			fmt.Print(password)
		}

	case ModeBatch:
		matrix, err := getMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		dirty := app.DirtySpell{Spell: cfg.Spell}
		magic, err := dirty.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrInvalidSpell+"\n", err)
			os.Exit(1)
		}
		password, err := magic.ExtractPassword(matrix)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrExtractPassword+"\n", err)
			os.Exit(1)
		}
		password = truncatePassword(password, cfg.MaxLen)
		fmt.Print(password)

	case ModePasswordStrength:
		runPasswordStrengthMode(cfg)
	}
}

// runPasswordStrengthMode calculates and displays the strength of a password from stdin.
func runPasswordStrengthMode(cfg Config) {
	masterResult := app.CalculateMasterPasswordStrength(cfg.MasterRaw)
	printStrengthTable(masterResult)
}
