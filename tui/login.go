package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsth131/ctfd-cli/tui/constants"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type (
	loginMsg struct{}
	errMsg   struct{ error }
)

type loginModel struct {
	focusIndex int
	inputs     []textinput.Model
	spinner    spinner.Model
	loading    bool
	errorMsg   string
}

func InitLogin() (tea.Model, tea.Cmd) {
	m := loginModel{
		inputs:  make([]textinput.Model, 2),
		loading: false,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 255

		switch i {
		case 0:
			t.Placeholder = "Username"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	m.spinner = spinner.New()
	m.spinner.Style = focusedStyle
	m.spinner.Spinner = spinner.Dot

	return m, nil
}

func login(username, password string) tea.Cmd {
	return func() tea.Msg {
		log.Default().Print("Logging in...")
		err := constants.C.Login(username, password)
		if err != nil {
			return errMsg{fmt.Errorf("Failed to login: %v", err)}
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
		constants.WindowSize = msg
		return m, nil
	case errMsg:
		log.Default().Print(msg)
		m.errorMsg = msg.Error()
		m.loading = false
		return m, tea.Batch(cmds...)
	case loginMsg:
		challenges, initCmd := InitChallenges()
		m, updateCmd := challenges.Update(constants.WindowSize)
		return m, tea.Batch(updateCmd, initCmd)
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
				m.errorMsg = ""
				cmds = append(cmds, login(m.inputs[0].Value(), m.inputs[1].Value()))
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
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
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
	fmt.Fprintf(&b, "\n\n%s\n\n%s", *button, constants.ErrStyle(m.errorMsg))

	return b.String()
}
