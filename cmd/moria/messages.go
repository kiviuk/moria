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
