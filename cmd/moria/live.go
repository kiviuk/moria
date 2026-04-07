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
				charStr := string(ch)
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

	visited := make(map[string]bool)
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
	fmt.Fprintf(&sb, MsgSpellPrompt, spellStyle.Render(m.spell), cursor)

	if m.maxLen > 0 {
		fmt.Fprintf(&sb, MsgPasswordWithMaxLen, passStyle.Render(m.password), len(m.password), m.maxLen)
		if m.state == StateMaxLenReached {
			fmt.Fprintf(&sb, MsgLiveError, errorStyle.Render(fmt.Sprintf(MsgMaxPasswordReached, m.maxLen)))
		}
	} else {
		fmt.Fprintf(&sb, MsgPasswordNoMaxLen, passStyle.Render(m.password), len(m.password))
	}

	if m.err != "" {
		fmt.Fprintf(&sb, MsgLiveError, errorStyle.Render(m.err))
	}

	sb.WriteByte('\n')
	sb.WriteString(hintStyle.Render(MsgLiveHint))
	sb.WriteByte('\n')

	return sb.String()
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

	return lm, nil
}
