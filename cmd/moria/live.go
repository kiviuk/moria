package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awnumar/memguard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kiviuk/moria/internal/app"
)

const colWidth = app.CharactersPerMatrixCell + 1

const (
	colorRed   = lipgloss.Color("1")
	colorGreen = lipgloss.Color("10")
	colorCyan  = lipgloss.Color("14")
	colorGray  = lipgloss.Color("241")
)

// PasteMode controls whether pasted (multi-character) input is accepted in live mode.
type PasteMode int

const (
	// PasteAllowed allows both single-key and pasted multi-character input.
	PasteAllowed PasteMode = iota
	// PasteIgnored rejects pasted input, accepting only single keystrokes.
	PasteIgnored
)

// LiveState represents the current input state of the live model.
type LiveState int

const (
	// StateNormal indicates normal typing is allowed.
	StateNormal LiveState = iota
	// StateMaxLenReached indicates the password has reached the configured max length.
	StateMaxLenReached
)

var (
	cellStyle      = lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Left)
	highlightStyle = lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Left).Foreground(colorGreen)
	headerStyle    = lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Left).Bold(true)
	rowNumStyle    = lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Left).Foreground(colorGray)
	spellStyle     = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	passStyle      = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	hintStyle      = lipgloss.NewStyle().Foreground(colorGray)
)

// liveModel holds the state for the interactive live mode TUI.
// spell and password are stored as []byte for secure memory wiping.
type liveModel struct {
	matrix            app.Matrix
	masterPasswordRaw *app.SecureBytes
	spell             []byte
	queryLetters      []app.QueryLetter
	password          []byte
	maxLen            int
	pasteMode         PasteMode
	state             LiveState
	err               string
}

// Wipe clears all sensitive data from the model.
func (m *liveModel) Wipe() {
	if m.masterPasswordRaw != nil {
		m.masterPasswordRaw.Wipe()
	}
	m.matrix.Wipe()
	memguard.WipeBytes(m.spell)
	memguard.WipeBytes(m.password)
	m.spell = nil
	m.password = nil
	for i := range m.queryLetters {
		m.queryLetters[i] = app.QueryLetter{}
	}
	m.queryLetters = nil
	m.err = ""
}

// newLiveModel creates a liveModel initialized with the given matrix and settings.
func newLiveModel(matrix app.Matrix, masterPasswordRaw *app.SecureBytes, maxLen int, pasteMode PasteMode) liveModel {
	return liveModel{
		matrix:            matrix,
		masterPasswordRaw: masterPasswordRaw,
		queryLetters:      make([]app.QueryLetter, 0),
		maxLen:            maxLen,
		pasteMode:         pasteMode,
		state:             StateNormal,
	}
}

// Init is the Bubbletea model initialization. Returns nil for no initial command.
func (m liveModel) Init() tea.Cmd {
	return nil
}

// Update handles keyboard input and updates the model state.
func (m liveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
	case tea.KeyEnter:
		return m, tea.Quit
	case tea.KeyBackspace:
		return m.doBackspace(), nil
	case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight,
		tea.KeyHome, tea.KeyEnd, tea.KeyPgUp, tea.KeyPgDown,
		tea.KeyInsert, tea.KeyDelete:
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		return m.doRunes(keyMsg)
	}
	return m, nil
}

// The function handles backspace in live mode
func (m liveModel) doBackspace() liveModel {
	// Nothing to delete if spell is empty
	if len(m.spell) == 0 {
		return m
	}

	// Remove the last character from the spell
	m.spell = m.spell[:len(m.spell)-1]

	// Remove the last query letter from the slice
	if len(m.queryLetters) > 0 {
		m.queryLetters = m.queryLetters[:len(m.queryLetters)-1]
	}

	// Calculate what the password length should be based on remaining query letters
	// Each query letter contributes CharactersPerMatrixCell characters to the password
	expectedLen := len(m.queryLetters) * app.CharactersPerMatrixCell

	// Truncate password to match expected length
	if len(m.password) >= expectedLen {
		m.password = m.password[:expectedLen]
	} else {
		// State is corrupted - clear everything to maintain consistency
		m.spell = nil
		m.queryLetters = nil
		m.password = nil
	}

	m.state = StateNormal
	m.err = ""
	return m
}

// doRunes processes rune input (regular characters and space).
func (m liveModel) doRunes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var runes []rune
	if msg.Type == tea.KeySpace {
		runes = []rune{' '}
	} else {
		runes = msg.Runes
	}
	if m.pasteMode == PasteIgnored && len(msg.Runes) > 1 {
		m.err = MsgPasteIgnored
		return m, nil
	}
	for _, ch := range runes {
		if ch < 32 || ch == 127 {
			continue
		}
		if !app.IsAllowedSpellChar(ch) {
			m.err = fmt.Sprintf(MsgInvalidChar, string(ch))
			return m, nil
		}
		if m.maxLen > 0 && len(m.password) >= m.maxLen {
			m.state = StateMaxLenReached
			return m, nil
		}
		charStr := string(ch)
		letter := app.MagicLetter{Letter: charStr, LetterPosition: len(m.spell)}
		query := letter.Query()
		passwordFragmentCell, err := m.matrix.PasswordFragment(query)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		passwordFragmentToAdd := passwordFragmentCell
		if m.maxLen > 0 && len(m.password)+len(passwordFragmentToAdd) > m.maxLen {
			remaining := m.maxLen - len(m.password)
			passwordFragmentToAdd = passwordFragmentToAdd[:remaining]
		}
		m.spell = append(m.spell, charStr...)
		m.queryLetters = append(m.queryLetters, query)
		m.password = append(m.password, passwordFragmentToAdd...)
		m.state = StateNormal
		m.err = ""
	}
	return m, nil
}

// wrapWithIndent breaks text into lines of max width, indenting continuation lines.
// Returns each visual line as a separate string for independent per-chunk rendering.
// Purely visual — never modifies the underlying model data.
// Uses strings.Builder for efficient string construction.
func wrapWithIndent(text string, width int, indent string) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	var sb strings.Builder
	sb.Grow(width + len(indent))

	sb.WriteString(text[:width])
	lines = append(lines, sb.String())
	sb.Reset()

	remaining := text[width:]
	for len(remaining) > width {
		sb.WriteString(indent)
		sb.WriteString(remaining[:width])
		lines = append(lines, sb.String())
		sb.Reset()
		remaining = remaining[width:]
	}
	if remaining != "" {
		sb.WriteString(indent)
		sb.WriteString(remaining)
		lines = append(lines, sb.String())
	}
	return lines
}

// View renders the live mode TUI screen.
func (m liveModel) View() string {
	var sb strings.Builder

	// Header row
	sb.WriteString(strings.Repeat(" ", colWidth))
	for col := range app.PasswordMatrixColumns {
		sb.WriteString(headerStyle.Render(app.ColHeader(col)))
	}
	sb.WriteByte('\n')

	// Separator
	sb.WriteString(strings.Repeat(" ", colWidth))
	for range app.PasswordMatrixColumns {
		sb.WriteString(strings.Repeat("─", colWidth-1) + " ")
	}
	sb.WriteByte('\n')

	var visited map[string]bool = make(map[string]bool)
	for _, q := range m.queryLetters {
		key := fmt.Sprintf("%d-%d", q.MatrixRow, q.LetterGroup)
		visited[key] = true
	}

	for row := range app.PasswordMatrixRows {
		fmt.Fprintf(&sb, "%s", rowNumStyle.Render(fmt.Sprintf("%d", row)))
		for col := range app.PasswordMatrixColumns {
			visualCell := string(m.matrix[row][col])
			key := fmt.Sprintf("%d-%d", row, col)
			if visited[key] {
				sb.WriteString(highlightStyle.Render(visualCell))
			} else {
				sb.WriteString(cellStyle.Render(visualCell))
			}
		}
		sb.WriteByte('\n')
	}

	sb.WriteByte('\n')

	cursor := ""
	if len(m.spell)%2 == 0 {
		cursor = "█"
	} else {
		cursor = " "
	}
	spellIndent := strings.Repeat(" ", len(SpellPromptLabel))
	spellChunks := wrapWithIndent(string(m.spell), app.LiveModeWrapWidth-len(SpellPromptLabel), spellIndent)
	for i, chunk := range spellChunks {
		isLast := i == len(spellChunks)-1
		if i == 0 {
			if isLast {
				fmt.Fprintf(&sb, MsgSpellPrompt, spellStyle.Render(chunk), cursor)
			} else {
				fmt.Fprintf(&sb, MsgSpellPrompt, spellStyle.Render(chunk), "")
			}
		} else {
			if isLast {
				fmt.Fprintf(&sb, "%s%s\n", spellStyle.Render(chunk), cursor)
			} else {
				fmt.Fprintf(&sb, "%s\n", spellStyle.Render(chunk))
			}
		}
	}

	if m.maxLen > 0 {
		m.renderPasswordChunks(&sb, true)
		if m.state == StateMaxLenReached {
			fmt.Fprintf(&sb, MsgLiveError, errorStyle.Render(fmt.Sprintf(MsgMaxPasswordReached, m.maxLen)))
		}
	} else {
		m.renderPasswordChunks(&sb, false)
	}

	if m.err != "" {
		fmt.Fprintf(&sb, MsgLiveError, errorStyle.Render(m.err))
	}

	sb.WriteByte('\n')
	sb.WriteString(hintStyle.Render(MsgLiveHint))
	sb.WriteByte('\n')

	return sb.String()
}

// renderPasswordChunks renders the password with proper line wrapping and alignment.
// Single-line: label + length counter on one line.
// Multi-line: first line gets "Password:" label, last line gets length counter, middle are just chunks.
func (m liveModel) renderPasswordChunks(sb *strings.Builder, withMaxLen bool) {
	passwordIndent := strings.Repeat(" ", len(PasswordPromptLabel))
	// Calculate available width for content (excluding label)
	contentWidth := app.LiveModeWrapWidth - len(PasswordPromptLabel)
	wrappedPass := wrapWithIndent(string(m.password), contentWidth, passwordIndent)

	// Single-line: one line with both label and length counter
	if len(wrappedPass) == 1 {
		if withMaxLen {
			fmt.Fprintf(sb, MsgPasswordWithMaxLen, passStyle.Render(wrappedPass[0]), len(m.password), m.maxLen)
		} else {
			fmt.Fprintf(sb, MsgPasswordNoMaxLen, passStyle.Render(wrappedPass[0]), len(m.password))
		}
		return
	}

	// Multi-line: iterate through chunks
	for i, chunk := range wrappedPass {
		isLast := i == len(wrappedPass)-1

		// First line: prepend "Password:" label
		if i == 0 {
			fmt.Fprintf(sb, "%s%s\n", PasswordPromptLabel, passStyle.Render(chunk))
			continue
		}

		// Last line: append length counter (chunk already has indent from wrapWithIndent)
		if isLast {
			if withMaxLen {
				fmt.Fprintf(sb, "%s (%d/%d)\n", passStyle.Render(chunk), len(m.password), m.maxLen)
			} else {
				fmt.Fprintf(sb, "%s (%d)\n", passStyle.Render(chunk), len(m.password))
			}
			continue
		}

		// Middle lines: just the chunk (already has indent from wrapWithIndent)
		fmt.Fprintf(sb, "%s\n", passStyle.Render(chunk))
	}
}

// LiveMode starts the interactive live mode TUI and returns the final model state.
// It runs the Bubbletea program with an alternate screen buffer.
// Note: The caller is responsible for wiping the original matrix
func LiveMode(matrix app.Matrix, maxLen int, pasteMode PasteMode, masterPasswordRaw *app.SecureBytes) (liveModel, error) {
	m := newLiveModel(matrix, masterPasswordRaw, maxLen, pasteMode)
	p := tea.NewProgram(m, tea.WithAltScreen())

	final, err := p.Run()
	if err != nil {
		return liveModel{}, err
	}

	lm, ok := final.(liveModel)
	if !ok {
		return liveModel{}, errors.New(ErrUnexpectedModel)
	}

	// Wipe the matrix in the liveModel to prevent sensitive data from lingering
	lm.matrix.Wipe()

	return lm, nil
}
