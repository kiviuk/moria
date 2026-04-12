package main

import (
	"errors"
	"fmt"
	"strings"

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
type liveModel struct {
	matrix            app.Matrix
	masterPasswordRaw string
	spell             string
	queryLetters      []app.QueryLetter
	password          string
	maxLen            int
	pasteMode         PasteMode
	state             LiveState
	err               string
}

// newLiveModel creates a liveModel initialized with the given matrix and settings.
func newLiveModel(matrix app.Matrix, masterPasswordRaw string, maxLen int, pasteMode PasteMode) liveModel {
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyBackspace:
			if m.spell != "" {
				m.spell = m.spell[:len(m.spell)-1]
				m.queryLetters = m.queryLetters[:len(m.queryLetters)-1]
				m.password = m.password[:len(m.password)-app.CharactersPerMatrixCell]
				m.state = StateNormal
				m.err = ""
			}
		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight,
			tea.KeyHome, tea.KeyEnd, tea.KeyPgUp, tea.KeyPgDown,
			tea.KeyInsert, tea.KeyDelete:
			return m, nil
		case tea.KeyRunes, tea.KeySpace:
			runes := msg.Runes
			if msg.Type == tea.KeySpace {
				// Makes TestLiveModel_Space_SingleKey work without relying on the terminal's handling of space input.
				runes = []rune{' '}
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
				var charStr string = string(ch)
				m.spell += charStr
				letter := app.MagicLetter{Letter: charStr, LetterPosition: len(m.spell) - 1}
				query := letter.Query()
				m.queryLetters = append(m.queryLetters, query)
				cell, cellErr := m.matrix.Cell(query)
				if cellErr != nil {
					m.err = cellErr.Error()
					return m, nil
				}
				m.password += cell
				m.state = StateNormal
				m.err = ""
			}
		}
	}
	return m, nil
}

// wrapWithIndent breaks text into lines of max width, indenting continuation lines.
// Returns each visual line as a separate string for independent per-chunk rendering.
// Purely visual — never modifies the underlying model data.
func wrapWithIndent(text string, width int, indent string) []string {
	if len(text) <= width {
		return []string{text}
	}
	var lines []string
	lines = append(lines, text[:width])
	remaining := text[width:]
	for len(remaining) > width {
		lines = append(lines, indent+remaining[:width])
		remaining = remaining[width:]
	}
	if remaining != "" {
		lines = append(lines, indent+remaining)
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
			cell := m.matrix[row][col]
			key := fmt.Sprintf("%d-%d", row, col)
			if visited[key] {
				sb.WriteString(highlightStyle.Render(cell))
			} else {
				sb.WriteString(cellStyle.Render(cell))
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
	spellChunks := wrapWithIndent(m.spell, app.LiveModeWrapWidth, "            ")
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
// Each chunk is rendered with the appropriate format based on its position (first/middle/last).
// The length counter appears only on the final line to align with the visual wrap point.
func (m liveModel) renderPasswordChunks(sb *strings.Builder, withMaxLen bool) {
	// Split password into chunks that fit within LiveModeWrapWidth (80 chars), padded with indent for continuation lines
	wrappedPass := wrapWithIndent(m.password, app.LiveModeWrapWidth, "            ")

	// Iterate through each wrapped chunk of the password
	for i, chunk := range wrappedPass {
		// Check if this is the last chunk - determines where length counter goes (only on last line)
		isLast := i == len(wrappedPass)-1
		// Check if this is the first chunk - determines if "Password:" label is shown
		isFirst := i == 0

		// switch evaluates cases in order - first matching case executes
		switch {
		// First chunk + maxLen enabled + is last = single-line password with max length info
		case isFirst && withMaxLen && isLast:
			fmt.Fprintf(sb, MsgPasswordWithMaxLen, passStyle.Render(chunk), len(m.password), m.maxLen)
		// First chunk + maxLen enabled = first line of wrapped, needs "Password:" label
		case isFirst && withMaxLen:
			fmt.Fprintf(sb, "  Password: %s\n", passStyle.Render(chunk))
		// First chunk + is last = single-line password without max length
		case isFirst && isLast:
			fmt.Fprintf(sb, MsgPasswordNoMaxLen, passStyle.Render(chunk), len(m.password))
		// First chunk = first line of wrapped password, needs "Password:" label
		case isFirst:
			fmt.Fprintf(sb, "  Password: %s\n", passStyle.Render(chunk))
		// Middle chunk + maxLen enabled + is last = last line of wrapped with max length info
		case withMaxLen && isLast:
			fmt.Fprintf(sb, "%s%s (%d/%d)\n", passStyle.Render(chunk), "", len(m.password), m.maxLen)
		// Middle chunk + maxLen enabled = continuation line without length info
		case withMaxLen:
			fmt.Fprintf(sb, "%s\n", passStyle.Render(chunk))
		// Middle chunk + is last = last line of wrapped without max length info
		case isLast:
			fmt.Fprintf(sb, "%s%s (%d)\n", passStyle.Render(chunk), "", len(m.password))
		// Middle chunk = continuation line without length info
		default:
			fmt.Fprintf(sb, "%s\n", passStyle.Render(chunk))
		}
	}
}

// LiveMode starts the interactive live mode TUI and returns the final model state.
// It runs the Bubbletea program with an alternate screen buffer.
func LiveMode(matrix app.Matrix, maxLen int, pasteMode PasteMode, masterPasswordRaw string) (liveModel, error) {
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

	lm.matrix.Wipe()

	return lm, nil
}
