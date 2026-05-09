package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type spellModel struct {
	input textinput.Model
	err   error
}

func newSpellModel() spellModel {
	ti := textinput.New()
	ti.Placeholder = "spell"
	ti.Focus()
	ti.EchoMode = textinput.EchoNormal

	return spellModel{
		input: ti,
	}
}

func (m spellModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m spellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.err = fmt.Errorf("%s", MsgSpellCancelled)
			return m, tea.Quit
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m spellModel) View() string {
	return fmt.Sprintf(
		MsgSpellInputPrompt,
		m.input.View(),
	)
}

// getSpell prompts the user for a spell on the TTY and returns the trimmed value.
func getSpell() (string, error) {
	// Test harness override: if MORIA_SPELL is set, print the prompt (so PTY tests
	// that wait for the prompt succeed) and return the value immediately.
	if s := os.Getenv("MORIA_SPELL"); s != "" {
		debugf("getSpell: MORIA_SPELL set, returning it")
		fmt.Fprintln(os.Stdout, "Enter spell:")
		return strings.TrimSpace(s), nil
	}

	in := os.Stdin
	out := os.Stdout
	ttyOpen := false

	if stat, err := os.Stdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
		debugf("getSpell: using os.Stdin/os.Stdout as TTY")
	} else {
		// Try opening /dev/tty when stdin is not a TTY (e.g., master was piped).
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			debugf("getSpell: could not open /dev/tty: %v", err)
			return "", fmt.Errorf("no TTY available for spell prompt")
		}
		in = tty
		out = tty
		ttyOpen = true
		defer tty.Close()
	}

	debugf("getSpell: starting Bubbletea (ttyOpen=%v)", ttyOpen)
	p := tea.NewProgram(newSpellModel(), tea.WithInput(in), tea.WithOutput(out))
	finalModel, err := p.Run()
	debugf("getSpell: Bubbletea.Run returned err=%v finalModelType=%T", err, finalModel)
	if err != nil {
		return "", err
	}

	sm, ok := finalModel.(spellModel)
	if !ok {
		return "", fmt.Errorf("%s", ErrUnexpectedModel)
	}
	if sm.err != nil {
		return "", sm.err
	}

	spell := strings.TrimSpace(sm.input.Value())
	debugf("getSpell: received spell length=%d", len(spell))
	return spell, nil
}
