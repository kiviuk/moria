package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kiviuk/pwdgen/internal/app"
)

var (
	cellStyle      = lipgloss.NewStyle().Width(4).Align(lipgloss.Center)
	highlightStyle = lipgloss.NewStyle().Width(4).Align(lipgloss.Center).Foreground(lipgloss.Color("10"))
	headerStyle    = lipgloss.NewStyle().Width(4).Align(lipgloss.Center).Bold(true)
	rowNumStyle    = lipgloss.NewStyle().Width(4).Align(lipgloss.Center).Foreground(lipgloss.Color("241"))
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
	err          string
}

func newLiveModel(matrix app.Matrix) liveModel {
	return liveModel{
		matrix:       matrix,
		queryLetters: make([]app.QueryLetter, 0),
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

	// Matrix header
	sb.WriteString(strings.Repeat(" ", 4))
	for col := 0; col < app.PasswordMatrixColumns; col++ {
		sb.WriteString(headerStyle.Render(app.ColHeader(col)))
	}
	sb.WriteByte('\n')

	// Separator
	sb.WriteString(strings.Repeat(" ", 4))
	for col := 0; col < app.PasswordMatrixColumns; col++ {
		sb.WriteString("─── ")
	}
	sb.WriteByte('\n')

	// Build visited set for highlighting
	visited := make(map[string]bool)
	for _, q := range m.queryLetters {
		key := fmt.Sprintf("%d-%d", q.MatrixRow, q.LetterGroup)
		visited[key] = true
	}

	// Data rows
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

	// Spell display
	cursor := ""
	if len(m.spell)%2 == 0 {
		cursor = "█"
	} else {
		cursor = " "
	}
	sb.WriteString(fmt.Sprintf("  Spell:    %s%s\n", spellStyle.Render(m.spell), cursor))

	// Password display
	sb.WriteString(fmt.Sprintf("  Password: %s\n", passStyle.Render(m.password)))

	// Error message
	if m.err != "" {
		sb.WriteString(fmt.Sprintf("  %s\n", errorStyle.Render(m.err)))
	}

	sb.WriteByte('\n')
	sb.WriteString(hintStyle.Render("  [Backspace] delete  [Enter] finish  [Ctrl+C] quit"))
	sb.WriteByte('\n')

	return sb.String()
}

func LiveMode(matrix app.Matrix) (string, error) {
	m := newLiveModel(matrix)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	lm := finalModel.(liveModel)
	return lm.password, nil
}
