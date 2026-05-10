// Package main implements the moria CLI — a deterministic, matrix-based password generator.
//
// moria derives unique passwords from a master secret and a memorable "spell"
// (typically a service name). The same inputs always produce the same output.
//
// Usage:
//
//	moria --magic                    # Generate a master password
//	moria "amazon"                   # Prompted for master, spell on argv
//	cat master.txt | moria "amazon"  # Master piped, spell on argv
//	cat master.txt | moria           # Master piped, prompted for spell
//	moria                            # Prompted for master, then for spell
//	cat master.txt | moria --pretty  # Display the matrix
//	cat master.txt | moria --live    # Interactive live mode
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

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

type InputState int

const (
	// InputStateUnknown indicates the stdin/arg state could not be determined.
	InputStateUnknown InputState = iota
	// InputStatePipedMasterWithSpellArg: master provided on stdin (pipe) and spell provided as argv.
	// Non-interactive, suitable for scripting: no prompts required.
	InputStatePipedMasterWithSpellArg
	// InputStatePipedMasterNoSpell: master provided on stdin (pipe) but no spell on argv.
	// Typical flow: read master from stdin, then prompt for spell interactively (TTY) if available.
	InputStatePipedMasterNoSpell
	// InputStateInteractiveMasterWithSpellArg: no piped master (interactive TTY) but spell provided as argv.
	// Typical flow: prompt for masked master password via TTY and use provided spell.
	InputStateInteractiveMasterWithSpellArg
	// InputStateInteractiveMasterNoSpell: neither master nor spell provided (interactive only).
	// Typical flow: prompt for master (masked) and then prompt for spell (unmasked) via TTY.
	InputStateInteractiveMasterNoSpell
)

func determineInputState(cfg Config) (InputState, error) {
	// If MORIA_MASTER_FILE is set, treat input as piped master so tests can override stdin.
	if mf := os.Getenv("MORIA_MASTER_FILE"); mf != "" {
		debugf("determineInputState: MORIA_MASTER_FILE set; treating stdin as piped")
		if cfg.Spell != "" {
			return InputStatePipedMasterWithSpellArg, nil
		}
		return InputStatePipedMasterNoSpell, nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return InputStateUnknown, fmt.Errorf("could not stat stdin: %w", err)
	}
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0
	spellProvided := cfg.Spell != ""
	if isPiped {
		if spellProvided {
			return InputStatePipedMasterWithSpellArg, nil
		}
		return InputStatePipedMasterNoSpell, nil
	}
	if spellProvided {
		return InputStateInteractiveMasterWithSpellArg, nil
	}
	return InputStateInteractiveMasterNoSpell, nil
}

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
	matrix, err := app.NewMatrix(master.Bytes())
	if err != nil {
		return app.Matrix{}, fmt.Errorf(ErrFailedCreateMatrix, err)
	}
	return matrix, nil
}

func readStdin() (*app.SecureBytes, error) {
	// Support an environment override to read the master password from a file.
	// This makes PTY-based tests easier without changing normal CLI behavior.
	if mf := os.Getenv("MORIA_MASTER_FILE"); mf != "" {
		debugf("readStdin: MORIA_MASTER_FILE set: %s", mf)
		data, err := os.ReadFile(mf)
		if err != nil {
			debugf("readStdin: failed to read master file: %v", err)
			return nil, fmt.Errorf(ErrFailedReadMaster, err)
		}
		debugf("readStdin: read %d bytes from file", len(data))
		if len(data) > app.MaxMasterPasswordInputBytes {
			debugf("readStdin: master file too large: %d bytes", len(data))
			memguard.WipeBytes(data)
			return nil, fmt.Errorf(ErrStdinTooLarge, app.MaxMasterPasswordInputBytes/1024)
		}
		sb := app.NewSecureBytes(data)
		memguard.WipeBytes(data)
		return sb.TrimSpace(), nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat stdin: %w", err)
	}
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0
	debugf("readStdin: isPiped=%v", isPiped)

	if isPiped {
		debugf("Limiting stdin to %d bytes (reading limit+1 to detect oversize)", app.MaxMasterPasswordInputBytes)
		// Read one byte beyond the limit so we can detect oversized input.
		limited := io.LimitReader(os.Stdin, app.MaxMasterPasswordInputBytes+1)
		data, err := io.ReadAll(limited)
		if err != nil {
			debugf("readStdin: read error: %v", err)
			return nil, fmt.Errorf(ErrFailedReadMaster, err)
		}
		debugf("readStdin: read %d bytes", len(data))
		if len(data) > app.MaxMasterPasswordInputBytes {
			debugf("readStdin: stdin too large: %d bytes", len(data))
			memguard.WipeBytes(data)
			return nil, fmt.Errorf(ErrStdinTooLarge, app.MaxMasterPasswordInputBytes/1024)
		}
		sb := app.NewSecureBytes(data)
		memguard.WipeBytes(data)
		debugf("readStdin: secure bytes length after trim: %d", sb.Len())
		return sb.TrimSpace(), nil
	}
	return getPassword()
}

func parseArgs(args []string) (Config, map[string]bool, error) {
	cfg := Config{Mode: ModeBatch}
	flags := make(map[string]bool)
	var positional []string
	var positionalAfterFlagEnd []string
	modeSet := false
	flagEnd := false
	i := 0

	for i < len(args) {
		arg := args[i]

		if flagEnd {
			positionalAfterFlagEnd = append(positionalAfterFlagEnd, arg)
			i++
			continue
		}

		handled := handleModeFlag(arg, flags, &cfg, &modeSet)
		if handled {
			i++
			continue
		}

		switch arg {
		case "--":
			flagEnd = true
		case "--max-len":
			flags["--max-len"] = true
			if i+1 >= len(args) {
				return cfg, flags, errors.New(ErrMaxLenRequiresValue)
			}
			val, err := strconv.Atoi(args[i+1])
			if err != nil {
				return cfg, flags, errors.New(ErrMaxLenNotNumber)
			}
			if val <= 0 {
				return cfg, flags, errors.New(ErrMaxLenNotNumber)
			}
			cfg.MaxLen = val
			i++
		case "--ignore-paste":
			flags["--ignore-paste"] = true
		case "--help", "-h":
			flags["--help"] = true
		default:
			positional = append(positional, arg)
		}
		i++
	}

	for _, arg := range positional {
		if strings.HasPrefix(arg, "--") {
			return cfg, flags, fmt.Errorf(ErrUnknownFlag, arg)
		}
	}

	if len(positional) > 0 {
		cfg.Spell = positional[0]
	} else if len(positionalAfterFlagEnd) > 0 {
		cfg.Spell = positionalAfterFlagEnd[0]
	}

	return cfg, flags, nil
}

func handleModeFlag(arg string, flags map[string]bool, cfg *Config, modeSet *bool) bool {
	switch arg {
	case "--magic":
		flags["--magic"] = true
		if !*modeSet {
			cfg.Mode = ModeMagic
			*modeSet = true
		}
		return true
	case "--pretty":
		flags["--pretty"] = true
		if !*modeSet {
			cfg.Mode = ModePretty
			*modeSet = true
		}
		return true
	case "--live":
		flags["--live"] = true
		if !*modeSet {
			cfg.Mode = ModeLive
			*modeSet = true
		}
		return true
	case "--show-strength":
		flags["--show-strength"] = true
		if !*modeSet {
			cfg.Mode = ModeShowPasswordStrength
			*modeSet = true
		}
		return true
	}
	return false
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
		return errors.New(ErrPasswordStrengthNoSpell)
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
	fmt.Println(MsgOptSeparator)
	fmt.Println(MsgOptHelp)
	fmt.Println()
	fmt.Println(MsgUsageExamples)
	fmt.Println(MsgExMagic)
	fmt.Println(MsgExSpell)
	fmt.Println(MsgExInteractive)
	fmt.Println(MsgExPipe)
	fmt.Println(MsgExPipeNoSpell)
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

func main() {
	memguard.SafeExit(run())
}

func run() int {
	cfg, flags, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
		return 1
	}

	if flags["--help"] {
		printUsage()
		return 0
	}

	state, serr := determineInputState(cfg)
	if serr != nil {
		fmt.Fprintf(os.Stderr, MsgErrorPrefix, serr)
		return 1
	}
	debugf("determined input state: %s", inputStateName(state))

	// If master was piped (or we are in an interactive TTY with no spell),
	// read the master first to consume stdin, then prompt for the spell via the TTY.
	// Only applies to modes that both need stdin and require a spell (i.e. batch mode).
	if cfg.Mode.needsStdin() && cfg.Mode.needsSpell() &&
		(state == InputStatePipedMasterNoSpell || state == InputStateInteractiveMasterNoSpell) {
		debugf("input state requires reading master then prompting for spell: %s", inputStateName(state))
		debugf("reading master from stdin (limit=%d bytes)", app.MaxMasterPasswordInputBytes)
		master, err := readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
			return 1
		}
		debugf("read master: len=%d", master.Len())
		cfg.MasterRaw = master
		expanded, err := app.ExpandToMatrix(master)
		if err != nil {
			fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
			return 1
		}
		cfg.Master = expanded
		defer cfg.Wipe()

		debugf("prompting for spell on TTY")
		spell, perr := getSpell()
		if perr != nil {
			fmt.Fprintf(os.Stderr, MsgErrorPrefix, perr)
			return 1
		}
		debugf("received spell: len=%d", len(spell))
		cfg.Spell = spell
	}

	if err := validateConfig(cfg, flags); err != nil {
		fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
		return 1
	}

	// Only read stdin here if we haven't already consumed it above based on input state.
	if cfg.Mode.needsStdin() && cfg.MasterRaw == nil {
		master, err := readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
			return 1
		}
		cfg.MasterRaw = master
		expanded, err := app.ExpandToMatrix(master)
		if err != nil {
			fmt.Fprintf(os.Stderr, MsgErrorPrefix, err)
			return 1
		}
		cfg.Master = expanded
		defer cfg.Wipe()
	}

	return runMode(&cfg, flags)
}

func runMode(cfg *Config, flags map[string]bool) int {
	switch cfg.Mode {
	case ModeMagic:
		return runMagicMode()
	case ModePretty:
		return runPrettyMode(cfg)
	case ModeLive:
		return runLiveMode(cfg, flags)
	case ModeBatch:
		return runBatchMode(cfg)
	case ModeShowPasswordStrength:
		runPasswordStrengthMode(cfg.MasterRaw)
		return 0
	default:
		return 0
	}
}

func runMagicMode() int {
	master, err := app.GenerateMasterPassword(app.MatrixBytes, app.MasterPasswordChars)
	if err != nil {
		fmt.Fprintf(os.Stderr, ErrFailedGenerateMaster+"\n", err)
		return 1
	}
	os.Stdout.Write(master.Bytes())
	master.Wipe()
	return 0
}

func runPrettyMode(cfg *Config) int {
	matrix, err := getMatrix(cfg.Master)
	if err != nil {
		matrix.Wipe()
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer matrix.Wipe()
	fmt.Print(matrix.Pretty())
	return 0
}

func runLiveMode(cfg *Config, flags map[string]bool) int {
	matrix, err := getMatrix(cfg.Master)
	if err != nil {
		matrix.Wipe()
		fmt.Fprintln(os.Stderr, err)
		return 1
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
		return 1
	}
	passwordBytes := finalModel.password
	if cfg.MaxLen > 0 && len(passwordBytes) > cfg.MaxLen {
		passwordBytes = passwordBytes[:cfg.MaxLen]
	}
	if len(passwordBytes) > 0 {
		os.Stdout.Write(passwordBytes)
	}
	finalModel.Wipe()
	return 0
}

func runBatchMode(cfg *Config) int {
	matrix, err := getMatrix(cfg.Master)
	if err != nil {
		matrix.Wipe()
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer matrix.Wipe()
	dirty := app.DirtySpell{Spell: cfg.Spell}
	magic, err := dirty.Parse()
	if err != nil {
		matrix.Wipe()
		fmt.Fprintf(os.Stderr, ErrInvalidSpell+": %v\n", err)
		return 1
	}
	password, err := matrix.ExtractPassword(magic, cfg.MaxLen)
	if err != nil {
		matrix.Wipe()
		fmt.Fprintf(os.Stderr, ErrExtractPassword+": %v\n", err)
		return 1
	}
	defer password.Wipe()
	if password.Len() > 0 {
		os.Stdout.Write(password.Bytes())
	}
	return 0
}

func runPasswordStrengthMode(masterPassword *app.SecureBytes) {
	masterResult := app.CalculateMasterPasswordStrength(masterPassword.Bytes())
	printStrengthTable(masterResult)
	masterPassword.Wipe()
}
