package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kiviuk/moria/internal/app"
)

type promptConfig struct {
	echoMode    textinput.EchoMode
	echoChar    rune
	placeholder string
	charLimit   int
	promptFmt   string
	cancelMsg   string
}

var passwordPromptCfg = promptConfig{
	echoMode:    textinput.EchoPassword,
	echoChar:    '•',
	placeholder: "master password",
	charLimit:   app.MaxMasterPasswordInputBytes,
	promptFmt:   MsgPasswordPrompt,
	cancelMsg:   MsgPasswordCancelled,
}

var spellPromptCfg = promptConfig{
	echoMode:    textinput.EchoNormal,
	placeholder: "spell",
	promptFmt:   MsgSpellInputPrompt,
	cancelMsg:   MsgSpellCancelled,
}

type promptModel struct {
	input textinput.Model
	cfg   promptConfig
	err   error
}

func newPromptModel(cfg promptConfig) promptModel {
	ti := textinput.New()
	ti.Placeholder = cfg.placeholder
	ti.Focus()
	ti.EchoMode = cfg.echoMode
	if cfg.echoChar != 0 {
		ti.EchoCharacter = cfg.echoChar
	}
	if cfg.charLimit > 0 {
		ti.CharLimit = cfg.charLimit
	}
	return promptModel{input: ti, cfg: cfg}
}

func (m promptModel) Init() tea.Cmd { return textinput.Blink }

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.err = fmt.Errorf("%s", m.cfg.cancelMsg)
			return m, tea.Quit
		}
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m promptModel) View() string {
	return fmt.Sprintf(m.cfg.promptFmt, m.input.View())
}

// runPrompt opens a TTY if needed, runs a Bubbletea text-input prompt, and returns the trimmed value.
func runPrompt(cfg promptConfig) (string, error) {
	in, out, closeFn, err := openTTY()
	if err != nil {
		return "", fmt.Errorf("no TTY available")
	}
	defer closeFn()
	p := tea.NewProgram(newPromptModel(cfg), tea.WithInput(in), tea.WithOutput(out))
	final, runErr := p.Run()
	if runErr != nil {
		return "", runErr
	}
	pm, ok := final.(promptModel)
	if !ok {
		return "", fmt.Errorf("%s", ErrUnexpectedModel)
	}
	if pm.err != nil {
		return "", pm.err
	}
	return strings.TrimSpace(pm.input.Value()), nil
}

func getPassword() (*app.SecureBytes, error) {
	debugf("getPassword: starting prompt")
	// Note: textinput.Value() returns a string — the backing array cannot be wiped.
	// This is a known limitation of the Bubbletea textinput component.
	val, err := runPrompt(passwordPromptCfg)
	if err != nil {
		return nil, err
	}
	sb := app.NewSecureBytesFromString(val)
	debugf("getPassword: received password length=%d", sb.Len())
	return sb, nil
}

func getSpell() (string, error) {
	// SECURITY NOTE: MORIA_SPELL is a development/scripting aid.
	// Environment variables are visible to all processes owned by the same user.
	if s := os.Getenv("MORIA_SPELL"); s != "" {
		debugf("getSpell: MORIA_SPELL set, returning it")
		fmt.Fprintln(os.Stdout, "Enter spell:")
		return strings.TrimSpace(s), nil
	}
	debugf("getSpell: starting prompt")
	spell, err := runPrompt(spellPromptCfg)
	if err != nil {
		return "", err
	}
	debugf("getSpell: received spell length=%d", len(spell))
	return spell, nil
}
