package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kiviuk/moria/internal/app"
)

const colWidth = app.CharactersPerMatrixCell + 1

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
	err          string
}

func newLiveModel(matrix app.Matrix, maxLen int) liveModel {
	return liveModel{
		matrix:       matrix,
		queryLetters: make([]app.QueryLetter, 0),
		maxLen:       maxLen,
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
				m.err = ""
			}
		default:
			s := msg.String()
			if len(s) == 1 {
				dirty := app.DirtySpell{Spell: s}
				_, parseErr := dirty.Parse()
				if parseErr != nil {
					m.err = fmt.Sprintf("invalid char: %q", s)
					return m, nil
				}
				if m.maxLen > 0 && len(m.password) >= m.maxLen {
					m.err = "max length reached"
					return m, nil
				}
				m.spell += s
				letter := app.MagicLetter{Letter: s, LetterPosition: len(m.spell) - 1}
				query := letter.Query()
				m.queryLetters = append(m.queryLetters, query)
				cell, cellErr := m.matrix.Cell(query)
				if cellErr != nil {
					m.err = cellErr.Error()
					return m, nil
				}
				m.password += cell
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
	for col := 0; col < app.PasswordMatrixColumns; col++ {
		sb.WriteString(headerStyle.Render(app.ColHeader(col)))
	}
	sb.WriteByte('\n')

	// Separator
	sb.WriteString(strings.Repeat(" ", colWidth))
	for col := 0; col < app.PasswordMatrixColumns; col++ {
		sb.WriteString(strings.Repeat("─", colWidth-1) + " ")
	}
	sb.WriteByte('\n')

	visited := make(map[string]bool)
	for _, q := range m.queryLetters {
		key := fmt.Sprintf("%d-%d", q.MatrixRow, q.LetterGroup)
		visited[key] = true
	}

	for row := 0; row < app.PasswordMatrixRows; row++ {
		sb.WriteString(rowNumStyle.Render(fmt.Sprintf("%d", row)))
		for col := 0; col < app.PasswordMatrixColumns; col++ {
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
	sb.WriteString(fmt.Sprintf("  Spell:    %s%s\n", spellStyle.Render(m.spell), cursor))

	if m.maxLen > 0 {
		sb.WriteString(fmt.Sprintf("  Password: %s (%d/%d)\n", passStyle.Render(m.password), len(m.password), m.maxLen))
		if len(m.password) >= m.maxLen {
			sb.WriteString(fmt.Sprintf("  %s\n", errorStyle.Render("[MAX LENGTH REACHED]")))
		}
	} else {
		sb.WriteString(fmt.Sprintf("  Password: %s\n", passStyle.Render(m.password)))
	}

	if m.err != "" {
		sb.WriteString(fmt.Sprintf("  %s\n", errorStyle.Render(m.err)))
	}

	sb.WriteByte('\n')
	sb.WriteString(hintStyle.Render("  [Backspace] delete  [Enter] finish  [Ctrl+C] quit"))
	sb.WriteByte('\n')

	return sb.String()
}

func LiveMode(matrix app.Matrix, maxLen int) (liveModel, error) {
	m := newLiveModel(matrix, maxLen)
	p := tea.NewProgram(m, tea.WithAltScreen())

	final, err := p.Run()
	if err != nil {
		return liveModel{}, err
	}

	lm, ok := final.(liveModel)
	if !ok {
		return liveModel{}, fmt.Errorf("unexpected model type returned by bubbletea")
	}

	return lm, nil
}
