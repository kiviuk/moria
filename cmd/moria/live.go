package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kiviuk/moria/internal/app"
)

const colWidth = app.CharactersPerMatrixCell + 1

type PasteMode int

const (
	PasteAllowed PasteMode = iota
	PasteIgnored
)

type LiveState int

const (
	StateNormal LiveState = iota
	StateMaxLenReached
)

var (
	cellStyle      = lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Left)
	highlightStyle = lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Left).Foreground(lipgloss.Color("10"))
	headerStyle    = lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Left).Bold(true)
	rowNumStyle    = lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Left).Foreground(lipgloss.Color("241"))
	spellStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	passStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	hintStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type liveModel struct {
	matrix       app.Matrix
	spell        string
	queryLetters []app.QueryLetter
	password     string
	maxLen       int
	pasteMode    PasteMode
	state        LiveState
	err          string
}

func newLiveModel(matrix app.Matrix, maxLen int, pasteMode PasteMode) liveModel {
	return liveModel{
		matrix:       matrix,
		queryLetters: make([]app.QueryLetter, 0),
		maxLen:       maxLen,
		pasteMode:    pasteMode,
		state:        StateNormal,
	}
}

func (m liveModel) Init() tea.Cmd {
	return nil
}

func (m liveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyBackspace:
			if len(m.spell) > 0 {
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
		case tea.KeyRunes:
			if m.pasteMode == PasteIgnored && len(msg.Runes) > 1 {
				m.err = MsgPasteIgnored
				return m, nil
			}
			for _, ch := range msg.Runes {
				if ch < 32 || ch == 127 {
					continue
				}
				charStr := string(ch)
				dirty := app.DirtySpell{Spell: charStr}
				_, parseErr := dirty.Parse()
				if parseErr != nil {
					m.err = fmt.Sprintf(MsgInvalidChar, charStr)
					return m, nil
				}
				if m.maxLen > 0 && len(m.password) >= m.maxLen {
					m.state = StateMaxLenReached
					return m, nil
				}
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
	fmt.Fprintf(&sb, "  Spell:    %s%s\n", spellStyle.Render(m.spell), cursor)

	if m.maxLen > 0 {
		fmt.Fprintf(&sb, "  Password: %s (%d/%d)\n", passStyle.Render(m.password), len(m.password), m.maxLen)
		if m.state == StateMaxLenReached {
			fmt.Fprintf(&sb, "  %s\n", errorStyle.Render(fmt.Sprintf(MsgMaxPasswordReached, m.maxLen)))
		}
	} else {
		fmt.Fprintf(&sb, "  Password: %s\n", passStyle.Render(m.password))
	}

	if m.err != "" {
		fmt.Fprintf(&sb, "  %s\n", errorStyle.Render(m.err))
	}

	sb.WriteByte('\n')
	sb.WriteString(hintStyle.Render("  [Backspace] delete  [Enter] finish  [Ctrl+C]|[ESC] quit"))
	sb.WriteByte('\n')

	return sb.String()
}

func LiveMode(matrix app.Matrix, maxLen int, pasteMode PasteMode) (liveModel, error) {
	m := newLiveModel(matrix, maxLen, pasteMode)
	p := tea.NewProgram(m, tea.WithAltScreen())

	final, err := p.Run()
	if err != nil {
		return liveModel{}, err
	}

	lm, ok := final.(liveModel)
	if !ok {
		return liveModel{}, fmt.Errorf(ErrUnexpectedModel)
	}

	return lm, nil
}
