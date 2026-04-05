package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// passwordModel is the Bubbletea model for the interactive masked password prompt.
type passwordModel struct {
	input textinput.Model
	err   error
}

// newPasswordModel creates a passwordModel with a focused, masked text input.
func newPasswordModel() passwordModel {
	ti := textinput.New()
	ti.Placeholder = "master password"
	ti.Focus()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return passwordModel{
		input: ti,
	}
}

// Init starts the text input blink animation.
func (m passwordModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles keyboard input for the password prompt.
func (m passwordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.err = fmt.Errorf("password entry cancelled")
			return m, tea.Quit
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the password prompt screen.
func (m passwordModel) View() string {
	return fmt.Sprintf(
		"Enter master password:\n\n  %s\n\n  (press Enter to confirm, Esc to cancel)",
		m.input.View(),
	)
}

// getPassword runs the interactive password prompt and returns the entered value.
func getPassword() (string, error) {
	p := tea.NewProgram(newPasswordModel())

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	pm, ok := finalModel.(passwordModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type returned by bubbletea")
	}
	m := pm
	if m.err != nil {
		return "", m.err
	}

	return m.input.Value(), nil
}
