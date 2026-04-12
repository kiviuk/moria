package main

import (
	"fmt"

	"github.com/awnumar/memguard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kiviuk/moria/internal/app"
)

type passwordModel struct {
	input textinput.Model
	err   error
}

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

func (m passwordModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m passwordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.err = fmt.Errorf("%s", MsgPasswordCancelled)
			return m, tea.Quit
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m passwordModel) View() string {
	return fmt.Sprintf(
		MsgPasswordPrompt,
		m.input.View(),
	)
}

func (m *passwordModel) Wipe() {
	value := m.input.Value()
	if value != "" {
		memguard.WipeBytes([]byte(value))
	}
}

func getPassword() (*app.SecureBytes, error) {
	p := tea.NewProgram(newPasswordModel())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	pm, ok := finalModel.(passwordModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type returned by bubbletea")
	}
	if pm.err != nil {
		return nil, pm.err
	}

	sb := app.NewSecureBytesFromString(pm.input.Value())
	pm.Wipe()

	return sb, nil
}
