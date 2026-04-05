package main

// Live mode UI messages
const (
	MsgMaxPasswordReached = "[MAX PASSWORD LENGTH %d REACHED]"
	MsgPasteIgnored       = "paste ignored, use --live without --ignore-paste to allow pasting"
	MsgInvalidChar        = "invalid char: %q"
	MsgMaxLenReached      = "max length reached"
)

// CLI error messages
const (
	ErrMaxLenRequiresValue  = "--max-len requires a value"
	ErrMaxLenNotNumber      = "--max-len value must be a number"
	ErrUnknownMode          = "unknown mode: %s"
	ErrModNotAllowed        = "%s not allowed in %s mode"
	ErrSpellRequired        = "%s mode requires a spell"
	ErrFailedReadMaster     = "Failed to read master password from pipe: %v"
	ErrFailedGenerateMaster = "Failed to generate master password: %v"
	ErrFailedCreateMatrix   = "Failed to create matrix: %v"
	ErrLiveMode             = "Live mode error: %v"
	ErrInvalidSpell         = "Invalid spell: %v"
	ErrExtractPassword      = "Failed to extract password: %v"
	ErrUnexpectedModel      = "unexpected model type returned by bubbletea"
)
