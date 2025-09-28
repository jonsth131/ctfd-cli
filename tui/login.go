package tui

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonsth131/ctfd-cli/tui/constants"
)

var (
	focusedButton = constants.FocusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", constants.BlurredStyle.Render("Submit"))
)

type (
	loginMsg struct{}
)

type loginModel struct {
	focusIndex int
	inputs     []textinput.Model
	spinner    spinner.Model
	loading    bool
	err        error
	width      int
	height     int
}

func InitLogin() (tea.Model, tea.Cmd) {
	m := loginModel{
		inputs:  make([]textinput.Model, 2),
		loading: false,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = constants.CursorStyle
		t.CharLimit = 255

		switch i {
		case 0:
			t.Placeholder = "Username"
			t.Focus()
			t.PromptStyle = constants.FocusedStyle
			t.TextStyle = constants.FocusedStyle
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	m.spinner = spinner.New()
	m.spinner.Style = constants.SpinnerStyle
	m.spinner.Spinner = spinner.Dot

	return m, nil
}

func loginCmd(username, password string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), constants.Timeout)
		defer cancel()
		log.Default().Print("Logging in...")
		err := constants.C.Login(ctx, username, password)
		if err != nil {
			return createErrMsg(fmt.Errorf("Failed to login: %v", err))
		}
		log.Default().Print("Logged in successfully")
		return loginMsg{}
	}
}

func (m loginModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	inputCmds := make([]tea.Cmd, len(m.inputs))
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case errMsg:
		log.Default().Print(msg)
		m.err = msg
		m.loading = false
		return m, tea.Batch(cmds...)
	case loginMsg:
		return InitChallenges(m.width, m.height)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.loading = true
				m.err = nil
				cmds = append(cmds, loginCmd(m.inputs[0].Value(), m.inputs[1].Value()))
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					inputCmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = constants.FocusedStyle
					m.inputs[i].TextStyle = constants.FocusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = constants.NoStyle
				m.inputs[i].TextStyle = constants.NoStyle
			}
		}
	}

	cmds = append(cmds, m.updateInputs(msg))
	cmds = append(cmds, inputCmds...)

	return m, tea.Batch(cmds...)
}

func (m *loginModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m loginModel) View() string {
	if len(m.inputs) == 0 {
		return "initializing..."
	}

	if m.loading {
		return fmt.Sprintf("\n %s%s", m.spinner.View(), "Logging in...")
	}

	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}

	errStr := renderError(m.err)

	fmt.Fprintf(&b, "\n\n%s\n\n%s", *button, errStr)

	return b.String()
}
