// Package main implements the moria CLI — a deterministic, matrix-based password generator.
//
// moria derives unique passwords from a master secret and a memorable "spell"
// (typically a service name). The same inputs always produce the same output.
//
// Usage:
//
//	moria --magic                  # Generate a master password
//	moria "amazon"                 # Generate password for Amazon
//	cat master.txt | moria "amazon"  # Piped from password manager
//	cat master.txt | moria --pretty # Display the matrix
//	cat master.txt | moria --live   # Interactive mode
package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"

	"github.com/awnumar/memguard"

	"github.com/kiviuk/moria/internal/app"
)

type Mode int

const (
	ModeBatch                Mode = iota
	ModeMagic                     // Generate master password
	ModePretty                    // Display matrix
	ModeLive                      // Interactive mode
	ModeShowPasswordStrength      // Analyze password strength
)

func (m Mode) String() string {
	return [...]string{"batch", "magic", "pretty", "live", "show-strength"}[m]
}

func (m Mode) Validate() error {
	if m < ModeBatch || m > ModeShowPasswordStrength {
		return fmt.Errorf("invalid mode: %d", m)
	}
	return nil
}

func (m Mode) needsStdin() bool {
	return m == ModePretty || m == ModeLive || m == ModeBatch || m == ModeShowPasswordStrength
}

func (m Mode) needsSpell() bool {
	return m == ModeBatch
}

func (m Mode) allowedMods() []string {
	switch m {
	case ModeLive:
		return []string{"--live", "--max-len", "--ignore-paste"}
	case ModeBatch:
		return []string{"--max-len"}
	case ModeMagic:
		return []string{"--magic"}
	case ModePretty:
		return []string{"--pretty"}
	case ModeShowPasswordStrength:
		return []string{"--show-strength"}
	default:
		return nil
	}
}

type Config struct {
	Mode      Mode
	Spell     string
	MaxLen    int
	Master    *app.SecureBytes
	MasterRaw *app.SecureBytes
}

func (c *Config) Wipe() {
	if c.Master != nil {
		c.Master.Wipe()
	}
	if c.MasterRaw != nil {
		c.MasterRaw.Wipe()
	}
}

func flagPermittedInMode(allowedFlags []string, flagToCheck string) bool {
	return slices.Contains(allowedFlags, flagToCheck)
}

func getMatrix(master *app.SecureBytes) (app.Matrix, error) {
	matrix, err := app.NewMatrix(master.String())
	if err != nil {
		return app.Matrix{}, fmt.Errorf(ErrFailedCreateMatrix, err)
	}
	return matrix, nil
}

func readStdin() (*app.SecureBytes, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat stdin: %w", err)
	}
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	if isPiped {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf(ErrFailedReadMaster, err)
		}
		sb := app.NewSecureBytes(data)
		memguard.WipeBytes(data)
		return sb.TrimSpace(), nil
	}
	return getPassword()
}

func parseArgs(args []string) (Config, map[string]bool, error) {
	cfg := Config{Mode: ModeBatch}
	flags := make(map[string]bool)
	var positional []string
	modeSet := false
	flagEnd := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if flagEnd {
			positional = append(positional, arg)
			continue
		}

		switch arg {
		case "--":
			flagEnd = true
		case "--magic":
			flags["--magic"] = true
			if !modeSet {
				cfg.Mode = ModeMagic
				modeSet = true
			}
		case "--pretty":
			flags["--pretty"] = true
			if !modeSet {
				cfg.Mode = ModePretty
				modeSet = true
			}
		case "--live":
			flags["--live"] = true
			if !modeSet {
				cfg.Mode = ModeLive
				modeSet = true
			}
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
		case "--show-strength":
			flags["--show-strength"] = true
			if !modeSet {
				cfg.Mode = ModeShowPasswordStrength
				modeSet = true
			}
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

func validateConfig(cfg Config, flags map[string]bool) error {
	for flag, present := range flags {
		if !present {
			continue
		}
		if !flagPermittedInMode(cfg.Mode.allowedMods(), flag) {
			return fmt.Errorf(ErrModNotAllowed, flag, cfg.Mode)
		}
	}

	if cfg.Mode == ModeShowPasswordStrength && cfg.Spell != "" {
		return fmt.Errorf("%s", ErrPasswordStrengthNoSpell)
	}

	if cfg.Mode.needsSpell() && cfg.Spell == "" {
		return fmt.Errorf(ErrSpellRequired, cfg.Mode)
	}

	return nil
}

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

func printStrengthTable(masterResult app.MasterPasswordResult) {
	if masterResult.Entropy == 0 {
		return
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, MsgMasterEntropy, masterResult.Entropy)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, MsgZxcvbnCrackTime, masterResult.CrackTimeDisplay)

	seconds := app.TimeToGuess(masterResult.Entropy, app.MasterPasswordGPUCluster)
	guessSpeed := formatGuessesPerSec(app.MasterPasswordGPUCluster)
	fmt.Fprintf(os.Stderr, MsgTimeToGuessWorstCase, guessSpeed, masterResult.Entropy, app.FormatSeconds(seconds))
}

func formatGuessesPerSec(n uint64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%dM", n/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%dK", n/1_000)
	}
	return fmt.Sprintf("%d", n)
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
		expanded, err := app.ExpandToMatrix(master)
		if err != nil {
			fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
			os.Exit(1)
		}
		cfg.Master = expanded
		defer cfg.Wipe()
	}

	switch cfg.Mode {
	case ModeMagic:
		master, err := app.GenerateMasterPassword(app.MatrixBytes, app.MasterPasswordChars)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrFailedGenerateMaster+"\n", err)
			os.Exit(1)
		}
		fmt.Print(master.String())
		master.Wipe()

	case ModePretty:
		matrix, err := getMatrix(cfg.Master)
		if err != nil {
			matrix.Wipe()
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer matrix.Wipe()
		fmt.Print(matrix.Pretty())

	case ModeLive:
		matrix, err := getMatrix(cfg.Master)
		if err != nil {
			matrix.Wipe()
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer matrix.Wipe()
		pasteMode := PasteAllowed
		if flags["--ignore-paste"] {
			pasteMode = PasteIgnored
		}
		finalModel, err := LiveMode(matrix, cfg.MaxLen, pasteMode, cfg.MasterRaw)
		if err != nil {
			matrix.Wipe()
			fmt.Fprintf(os.Stderr, ErrLiveMode+": %v\n", err)
			os.Exit(1)
		}
		password := finalModel.password
		// Truncate password to maxLen if specified
		passwordBytes := []byte(password)
		if cfg.MaxLen > 0 && len(passwordBytes) > cfg.MaxLen {
			passwordBytes = passwordBytes[:cfg.MaxLen]
		}
		if len(passwordBytes) > 0 {
			os.Stdout.Write(passwordBytes)
		}
		finalModel.Wipe()
		memguard.WipeBytes(passwordBytes)

	case ModeBatch:
		matrix, err := getMatrix(cfg.Master)
		if err != nil {
			matrix.Wipe()
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer matrix.Wipe()
		dirty := app.DirtySpell{Spell: cfg.Spell}
		magic, err := dirty.Parse()
		if err != nil {
			matrix.Wipe()
			fmt.Fprintf(os.Stderr, ErrInvalidSpell+": %v\n", err)
			os.Exit(1)
		}
		password, err := matrix.ExtractPassword(magic, cfg.MaxLen)
		if err != nil {
			matrix.Wipe()
			fmt.Fprintf(os.Stderr, ErrExtractPassword+": %v\n", err)
			os.Exit(1)
		}
		defer password.Wipe()
		if password.Len() > 0 {
			os.Stdout.Write(password.Bytes())
		}

	case ModeShowPasswordStrength:
		runPasswordStrengthMode(cfg.MasterRaw)
	}
}

func runPasswordStrengthMode(masterPassword *app.SecureBytes) {
	masterResult := app.CalculateMasterPasswordStrength(masterPassword.Bytes())
	printStrengthTable(masterResult)
}
