package main

import (
	"fmt"
	"os"

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
	ti.CharLimit = app.MaxMasterPasswordInputBytes

	return passwordModel{
		input: ti,
	}
}

func (m passwordModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m passwordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
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

func getPassword() (*app.SecureBytes, error) {
	// Prefer os.Stdin/os.Stdout when they are TTYs (common when running under a PTY).
	in := os.Stdin
	out := os.Stdout
	ttyOpen := false

	if stat, err := os.Stdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
		debugf("getPassword: using os.Stdin/os.Stdout as TTY")
	} else {
		// Try opening /dev/tty when stdin is not a TTY (e.g., master was piped).
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			debugf("getPassword: could not open /dev/tty: %v", err)
			return nil, fmt.Errorf("no TTY available for password prompt")
		}
		in = tty
		out = tty
		ttyOpen = true
		defer tty.Close()
	}

	debugf("getPassword: starting Bubbletea (ttyOpen=%v)", ttyOpen)
	p := tea.NewProgram(newPasswordModel(), tea.WithInput(in), tea.WithOutput(out))
	finalModel, err := p.Run()
	debugf("getPassword: Bubbletea.Run returned err=%v finalModelType=%T", err, finalModel)
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

	// Note: pm.input.Value() returns a string from textinput.
	// Strings are immutable and cannot be securely wiped.
	// This is a known limitation of the Bubbletea textinput component.
	// The master password entered here will remain in memory until GC.
	sb := app.NewSecureBytesFromString(pm.input.Value())
	debugf("getPassword: received password length=%d", sb.Len())
	return sb, nil
}
