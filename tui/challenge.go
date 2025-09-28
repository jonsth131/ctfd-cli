package tui

import (
	"context"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsth131/ctfd-cli/api"
	"github.com/jonsth131/ctfd-cli/tui/constants"
)

type challengeKeymap struct {
	Back   key.Binding
	Reload key.Binding
	Submit key.Binding
	Quit   key.Binding
}

func (k challengeKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back, k.Reload, k.Submit, k.Quit}
}

func (k challengeKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

var ChallengeKeymap = challengeKeymap{
	Back:   constants.Keymap.Back,
	Reload: constants.Keymap.Reload,
	Submit: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "submit flag"),
	),
	Quit: constants.Keymap.Quit,
}

type updateChallengeCmd struct {
	challenge *api.Challenge
}

type setMessageCmd struct {
	message string
}

type mode int

const (
	view mode = iota
	submit
)

type challengeModel struct {
	mode      mode
	viewport  viewport.Model
	challenge *api.Challenge
	help      help.Model
	input     textinput.Model
	err       string
	message   string
}

func fetchChallenge(id int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), constants.Timeout)
		defer cancel()
		log.Default().Printf("Fetching challenge %d...", id)
		challenge, err := constants.C.GetChallenge(ctx, uint16(id))
		if err != nil {
			return createErrMsg(fmt.Errorf("Failed to fetch challenge %d: %v", id, err))
		}

		log.Default().Printf("Fetched challenge %d", id)
		return updateChallengeCmd{challenge}
	}
}

func submitFlagCmd(id int, flag string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), constants.Timeout)
		defer cancel()
		log.Default().Printf("Submitting flag: %s for challenge: %d", flag, id)
		result, err := constants.C.SubmitFlag(ctx, id, flag)
		if err != nil {
			return createErrMsg(fmt.Errorf("Failed to submit flag for challenge %d: %v", id, err))
		}
		return setMessageCmd{result.Message}
	}
}

func InitChallenge(id int) (challengeModel, tea.Cmd) {
	input := textinput.New()
	input.Prompt = "$ "
	input.Placeholder = "Flag"
	input.CharLimit = 250
	input.Width = 50

	m := challengeModel{
		challenge: nil,
		help:      help.New(),
		err:       "",
		message:   "",
		mode:      view,
		input:     input,
	}

	top, right, bottom, left := constants.DocStyle.GetMargin()
	m.viewport = viewport.New(constants.WindowSize.Width-left-right, constants.WindowSize.Height-top-bottom-5)
	m.viewport.Style = lipgloss.NewStyle().Align(lipgloss.Bottom)

	return m, fetchChallenge(id)
}

func formatChallenge(challenge api.Challenge) string {
	files := ""
	if len(challenge.Files) != 0 {
		var formatted []string

		for _, fullURL := range challenge.Files {
			cleanPath := strings.Split(fullURL, "?")[0]
			filename := path.Base(cleanPath)

			formatted = append(formatted, fmt.Sprintf("- %s", filename))
		}

		files = fmt.Sprintf("\n\n## Files:\n\n%s", strings.Join(formatted, "\n"))
	}

	hints := ""
	if len(challenge.Hints) != 0 {
		hints = fmt.Sprintf("\n\n**Hints**: %d\n\n", len(challenge.Hints))
	}

	tags := ""
	if len(challenge.Tags) > 0 {
		tags = fmt.Sprintf("\n**Tags**: %s\n\n", strings.Join(challenge.Tags, ", "))
	}

	attempts := strconv.Itoa(challenge.Attempts)

	if challenge.MaxAttempts > 0 {
		attempts += fmt.Sprintf(" / %d", challenge.MaxAttempts)
	}

	return fmt.Sprintf(`# %s - %d pts

**Category**: %s
%s

**Solves**: %d

**Solved by me**: %t

**Attempts**: %s

## Description

%s

%s

%s%s`, challenge.Name, challenge.Value, challenge.Category, tags, challenge.Solves, challenge.SolvedByMe, attempts, challenge.Description, challenge.ConnectionInfo, files, hints)
}

func (m *challengeModel) setViewportContent() {
	var content string
	if m.challenge == nil {
		content = "Loading challenge..."
	} else {
		content = formatChallenge(*m.challenge)
	}
	str, err := glamour.Render(content, "dark")
	if err != nil {
		m.err = "could not render content with glamour"
	}
	m.viewport.SetContent(str)
}

func (m challengeModel) Init() tea.Cmd { return nil }

func (m challengeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Default().Printf("Challenge view received message: %v, %T\n", msg, msg)
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case updateChallengeCmd:
		m.challenge = msg.challenge
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
	case setMessageCmd:
		m.message = msg.message
	case errMsg:
		log.Default().Print(msg)
		m.err = msg.Error()
	case tea.KeyMsg:
		if m.input.Focused() {
			if key.Matches(msg, constants.Keymap.Enter) {
				if m.mode == submit {
					cmds = append(cmds, submitFlagCmd(int(m.challenge.Id), m.input.Value()))
				}
				m.input.SetValue("")
				m.mode = view
				m.input.Blur()
			}
			if key.Matches(msg, constants.Keymap.Back) {
				m.input.SetValue("")
				m.mode = view
				m.input.Blur()
			}
			// only log keypresses for the input field when it's focused
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			switch {
			case key.Matches(msg, ChallengeKeymap.Submit):
				m.mode = submit
				m.input.Focus()
				cmd = textinput.Blink
			case key.Matches(msg, constants.Keymap.Reload):
				return m, fetchChallenge(int(m.challenge.Id))
			case key.Matches(msg, constants.Keymap.Quit):
				return m, tea.Quit
			case key.Matches(msg, constants.Keymap.Back):
				challenges, initCmd := InitChallenges()
				m, updateCmd := challenges.Update(constants.WindowSize)
				return m, tea.Batch(updateCmd, initCmd)
			}
		}
	}

	m.setViewportContent()

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m challengeModel) View() string {
	if m.challenge == nil {
		return "Loading challenge..."
	}

	alert := lipgloss.JoinHorizontal(lipgloss.Left, constants.ErrStyle(m.err), constants.AlertStyle(m.message))

	if m.input.Focused() {
		formatted := lipgloss.JoinVertical(lipgloss.Top, "\n", m.viewport.View(), m.help.View(ChallengeKeymap), alert, m.input.View())
		return constants.DocStyle.Render(formatted)
	} else {
		formatted := lipgloss.JoinVertical(lipgloss.Top, "\n", m.viewport.View(), m.help.View(ChallengeKeymap), alert)
		return constants.DocStyle.Render(formatted)
	}
}
