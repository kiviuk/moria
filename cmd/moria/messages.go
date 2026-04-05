package main

// Live mode UI messages displayed to the user during interactive input.
const (
	// MsgMaxPasswordReached is shown when the user tries to type beyond the configured max length.
	MsgMaxPasswordReached = "[MAX PASSWORD LENGTH %d REACHED]"
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
	ErrMaxLenNotNumber = "--max-len value must be a number"
	// ErrUnknownMode is returned when an unrecognized mode is detected.
	ErrUnknownMode = "unknown mode: %s"
	// ErrModNotAllowed is a format string returned when a flag is not permitted in the current mode.
	ErrModNotAllowed = "%s not allowed in %s mode"
	// ErrSpellRequired is a format string returned when a mode requires a spell but none was provided.
	ErrSpellRequired = "%s mode requires a spell"
	// ErrFailedReadMaster is a format string returned when reading the master password from stdin fails.
	ErrFailedReadMaster = "Failed to read master password from pipe: %v"
	// ErrFailedGenerateMaster is a format string returned when master password generation fails.
	ErrFailedGenerateMaster = "Failed to generate master password: %v"
	// ErrFailedCreateMatrix is a format string returned when matrix creation fails.
	ErrFailedCreateMatrix = "Failed to create matrix: %v"
	// ErrLiveMode is a format string returned when live mode encounters an error.
	ErrLiveMode = "Live mode error: %v"
	// ErrInvalidSpell is a format string returned when spell validation fails.
	ErrInvalidSpell = "Invalid spell: %v"
	// ErrExtractPassword is a format string returned when password extraction from the matrix fails.
	ErrExtractPassword = "Failed to extract password: %v"
	// ErrUnexpectedModel is returned when the Bubbletea program returns an unexpected model type.
	ErrUnexpectedModel = "unexpected model type returned by bubbletea"
)

// Strength display messages for batch mode output.
const (
	// MsgPwdEntropy is the label for generated password entropy.
	MsgPwdEntropy = "\nPassword entropy: %d bits\n"
	// MsgMasterEntropy is the label for master password entropy.
	MsgMasterEntropy = "Master entropy:   %d bits\n"
	// MsgTimeToGuessGenerated is the header for generated password crack times.
	MsgTimeToGuessGenerated = "\nTime to guess (generated password):\n"
	// MsgTimeToGuessMaster is the header for master password crack times.
	MsgTimeToGuessMaster = "\nTime to guess (master password, via Argon2id):\n"
	// MsgStrengthTableRow is the format for each row in the strength table.
	MsgStrengthTableRow = "  %-24s %s\n"
)

// Live mode UI strings.
const (
	// MsgSpellPrompt is the format string for the spell display line.
	MsgSpellPrompt = "  Spell:    %s%s\n"
	// MsgPasswordWithMaxLen is the format string for password display with max length.
	MsgPasswordWithMaxLen = "  Password: %s (%d/%d)\n" //nolint:gosec // "Password" refers to the generated password, not a credential
	// MsgPasswordNoMaxLen is the format string for password display without max length.
	MsgPasswordNoMaxLen = "  Password: %s (%d)\n" //nolint:gosec // "Password" refers to the generated password, not a credential
	// MsgStrengthBar is the format string for the strength bar line.
	MsgStrengthBar = "  Strength: %s\n"
	// MsgTimeToGuessMasterPass is shown when the master password is the bottleneck.
	MsgTimeToGuessMasterPass = "  Time to guess (8x 4090): %s (bottleneck: master pass)\n" //nolint:gosec // "pass" is short for password, not a credential literal
	// MsgTimeToGuessGeneratedPass is shown when the generated password is the bottleneck.
	MsgTimeToGuessGeneratedPass = "  Time to guess (8x 4090): %s (bottleneck: generated pass)\n" //nolint:gosec // "pass" is short for password, not a credential literal
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
	MsgUsageHeader = "Usage: moria [--magic|--pretty|--live] [--max-len N] [--ignore-paste] [--super-strength] <spell>"
	// MsgUsageOptions is the options header.
	MsgUsageOptions = "Options:"
	// MsgOptMagic is the description for --magic.
	MsgOptMagic = "  --magic              Generate a master password"
	// MsgOptPretty is the description for --pretty.
	MsgOptPretty = "  --pretty             Display the password matrix from your master password"
	// MsgOptLive is the description for --live.
	MsgOptLive = "  --live               Interactive mode: type your spell and see the password build in real-time"
	// MsgOptMaxLen is the description for --max-len.
	MsgOptMaxLen = "  --max-len            Truncate generated output to N characters (live and batch modes only)"
	// MsgOptIgnorePaste is the description for --ignore-paste.
	MsgOptIgnorePaste = "  --ignore-paste       Ignore pasted input in live mode (single characters only, live mode only)"
	// MsgOptStrength is the description for --super-strength.
	MsgOptStrength = "  --super-strength     Show time-to-guess estimates (SLOW: ~20s, batch mode only)"
	// MsgOptHelp is the description for --help.
	MsgOptHelp = "  -h, --help           Show this help message"
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
	// MsgExStrength is the example for --super-strength.
	MsgExStrength = "  cat master.txt | moria --super-strength \"amazon\"  # SLOW: ~20s"
)

// Strength display messages for time formatting.
const (
	// MsgUncrackableCompact is the short label for extremely high entropy values (live mode).
	MsgUncrackableCompact = "uncrackable"
	// MsgUncrackable is the full label for extremely high entropy values (batch mode).
	MsgUncrackable = "effectively uncrackable"
)

// Generic CLI messages.
const (
	// MsgErrorPrefix is the prefix for error messages.
	MsgErrorPrefix = "Error: %v\n"
)
