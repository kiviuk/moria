package main

// Live mode UI messages displayed to the user during interactive input.
const (
	// MsgMaxPasswordReached is shown when the user tries to type beyond the configured max length.
	MsgMaxPasswordReached = "max password length %d reached"
	// MsgPasteIgnored is shown when the user attempts to paste while --ignore-paste is active.
	MsgPasteIgnored = "paste ignored, use --live without --ignore-paste to allow pasting"
	// MsgInvalidChar is a format string shown when an invalid character is typed.
	MsgInvalidChar = "invalid char: %q"
	// MsgMaxLenReached indicates the maximum length has been reached.
	MsgMaxLenReached = "max length reached"
)

// CLI error messages used for stderr output across all modes.
const (
	// ErrMaxLenRequiresValue is returned when --max-len is provided without a value.
	ErrMaxLenRequiresValue = "--max-len requires a value"
	// ErrMaxLenNotNumber is returned when --max-len value is not a valid integer.
	ErrMaxLenNotNumber = "--max-len value must be a number greater than 0"
	// ErrUnknownMode is returned when an unrecognized mode is detected.
	ErrUnknownMode = "unknown mode: %s"
	// ErrModNotAllowed is a format string returned when a flag is not permitted in the current mode.
	ErrModNotAllowed = "%s not allowed in %s mode"
	// ErrPasswordStrengthNoSpell is returned when --show-strength is used with a spell.
	ErrPasswordStrengthNoSpell = "--show-strength is standalone, spell not allowed"
	// ErrSpellRequired is a format string returned when a mode requires a spell but none was provided.
	ErrSpellRequired = "%s mode requires a spell"
	// ErrUnknownFlag is returned when an argument looks like a flag but isn't recognized.
	ErrUnknownFlag = "unknown flag: %s (use '--' before spell if intentional)"
	// ErrFailedReadMaster is a format string returned when reading the master password from stdin fails.
	ErrFailedReadMaster = "Failed to read master password from pipe: %v"
	// ErrFailedGenerateMaster is a format string returned when master password generation fails.
	ErrFailedGenerateMaster = "Failed to generate master password: %v"
	// ErrFailedCreateMatrix is a format string returned when matrix creation fails.
	ErrFailedCreateMatrix = "Failed to create matrix: %v"
	// ErrLiveMode is returned when live mode encounters an error.
	ErrLiveMode = "Live mode error"
	// ErrInvalidSpell is returned when spell validation fails.
	ErrInvalidSpell = "Invalid spell"
	// ErrExtractPassword is returned when password extraction from the matrix fails.
	ErrExtractPassword = "Failed to extract password"
	// ErrUnexpectedModel is returned when the Bubbletea program returns an unexpected model type.
	ErrUnexpectedModel = "unexpected model type returned by bubbletea"
)

// Strength display messages for master password strength mode.
const (
	// MsgMasterEntropy is the label for master password entropy.
	MsgMasterEntropy = "zxcvbn master password entropy: %d bits\n"
	// MsgZxcvbnCrackTime is the header for zxcvbn's generic crack time estimate.
	MsgZxcvbnCrackTime = "zxcvbn crack time estimate (generic): %s\n"
	// MsgTimeToGuessWorstCase is the worst case time estimate using GPU cluster speed.
	MsgTimeToGuessWorstCase = "\nAssuming attacker %s guesses/sec and %d bits (from zxcvbn), worst case: %s\n"
)

// Live mode UI strings.
const (
	// MsgSpellPrompt is the format string for the spell display line.
	MsgSpellPrompt = "  Spell:    %s%s\n"
	// MsgPasswordWithMaxLen is the format string for password display with max length.
	MsgPasswordWithMaxLen = "  Password: %s (%d/%d)\n" //nolint:gosec // "Password" refers to the generated password, not a credential
	// MsgPasswordNoMaxLen is the format string for password display without max length.
	MsgPasswordNoMaxLen = "  Password: %s (%d)\n" //nolint:gosec // "Password" refers to the generated password, not a credential
	// MsgLiveHint is the hint line shown at the bottom of live mode.
	MsgLiveHint = "  [Backspace] delete  [Enter] finish  [Ctrl+C]|[ESC] quit"
	// MsgLiveError is the format string for error display in live mode.
	MsgLiveError = "  %s\n"
)

// Password prompt messages.
const (
	// MsgPasswordPrompt is the format string for the interactive password prompt.
	MsgPasswordPrompt = "Enter master password:\n\n  %s\n\n  (press Enter to confirm, Esc to cancel)"
	// MsgPasswordCancelled is returned when the user cancels password entry.
	MsgPasswordCancelled = "password entry cancelled"
)

// Help and usage messages.
const (
	// MsgUsageTitle is the title line of the help output.
	MsgUsageTitle = "moria — deterministic password generator"
	// MsgUsageHeader is the usage format line.
	MsgUsageHeader = "Usage: moria [--magic|--pretty|--live|--show-strength] [--max-len N] [--ignore-paste] [--] <spell>"
	// MsgUsageOptions is the options header.
	MsgUsageOptions = "Options:"
	// MsgOptMagic is the description for --magic.
	MsgOptMagic = " --magic Generate a master password"
	// MsgOptPretty is the description for --pretty.
	MsgOptPretty = " --pretty Display the password matrix from your master password"
	// MsgOptLive is the description for --live.
	MsgOptLive = " --live Interactive mode: type your spell and see the password build in real-time"
	// MsgOptMaxLen is the description for --max-len.
	MsgOptMaxLen = " --max-len Truncate generated output to N characters (live and batch modes only)"
	// MsgOptIgnorePaste is the description for --ignore-paste.
	MsgOptIgnorePaste = " --ignore-paste Ignore pasted input in live mode (single characters only, live mode only)"
	// MsgOptPasswordStrength is the description for --show-strength.
	MsgOptPasswordStrength = " --show-strength Show strength of password from stdin (standalone mode)"
	// MsgOptSeparator is the description for --.
	MsgOptSeparator = " -- Spell separator (use before spells starting with --)"
	// MsgOptHelp is the description for --help.
	MsgOptHelp = " -h, --help Show this help message"
	// MsgUsageExamples is the examples header.
	MsgUsageExamples = "Examples:"
	// MsgExMagic is the example for --magic.
	MsgExMagic = "  moria --magic                              # Generate a new master password"
	// MsgExSpell is the example for spell usage.
	MsgExSpell = "  moria \"amazon\"                             # Generate password for Amazon"
	// MsgExPipe is the example for piped usage.
	MsgExPipe = "  cat master.txt | moria \"amazon\"             # Piped from password manager"
	// MsgExPretty is the example for --pretty.
	MsgExPretty = "  cat master.txt | moria --pretty             # Display the matrix"
	// MsgExLive is the example for --live.
	MsgExLive = "  cat master.txt | moria --live               # Interactive mode (paste allowed)"
	// MsgExLiveIgnorePaste is the example for --live --ignore-paste.
	MsgExLiveIgnorePaste = "  cat master.txt | moria --live --ignore-paste # Interactive mode (paste blocked)"
	// MsgExMaxLen is the example for --max-len.
	MsgExMaxLen = "  cat master.txt | moria --max-len 16 \"amazon\"  # Limited length"
	// MsgExPasswordStrength is the example for --show-strength.
	MsgExPasswordStrength = `  echo "x" | moria --show-strength  # Show password strength` //nolint:gosec // example only, not a real password
)

// Generic CLI messages.
const (
	// MsgErrorPrefix is the prefix for error messages.
	MsgErrorPrefix = "Error: %v\n"
)
