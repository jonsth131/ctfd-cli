package tui

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsth131/ctfd-cli/api"
	"github.com/jonsth131/ctfd-cli/tui/constants"
)

type challengesKeymap struct {
	Enter  key.Binding
	Reload key.Binding
	Quit   key.Binding
}

func (k challengesKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Reload, k.Quit}
}

func (k challengesKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

var ChallengesKeymap = challengesKeymap{
	Enter:  constants.Keymap.Enter,
	Reload: constants.Keymap.Reload,
	Quit:   constants.Keymap.Quit,
}

type updateChallengesCmd struct {
	challenges []api.ListChallenge
}

type challengesModel struct {
	table       table.Model
	help        help.Model
	screensHelp help.Model
	msg         string
	err         string
}

func fetchChallenges() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), constants.Timeout)
		defer cancel()
		log.Default().Print("Fetching challenges...")
		challenges, err := constants.C.GetChallenges(ctx)
		if err != nil {
			return createErrMsg(fmt.Errorf("Failed to fetch challenges: %v", err))
		}
		log.Default().Print("Fetched challenges")
		return updateChallengesCmd{challenges}
	}
}

func setTableSize(t *table.Model) {
	if constants.WindowSize.Height != 0 {
		nameLength := ((constants.WindowSize.Width - 17) / 4) * 2
		categoryLength := constants.WindowSize.Width - 30 - nameLength

		columns := []table.Column{
			{Title: "ID", Width: 5},
			{Title: "Name", Width: nameLength},
			{Title: "Category", Width: categoryLength},
			{Title: "Value", Width: 5},
			{Title: "Solved", Width: 7},
		}

		t.SetColumns(columns)

		top, right, bottom, left := constants.DocStyle.GetMargin()

		t.SetHeight(constants.WindowSize.Height - top - bottom - 5)
		t.SetWidth(constants.WindowSize.Width - left - right + 1)
	}
}

func createRows(challenges []api.ListChallenge) []table.Row {
	rows := make([]table.Row, len(challenges))

	for i, challenge := range challenges {
		solved := ""
		if challenge.SolvedByMe {
			solved = "✓"
		}
		rows[i] = table.Row{fmt.Sprintf("%d", challenge.Id), challenge.Name, challenge.Category, fmt.Sprintf("%d", challenge.Value), solved}
	}

	return rows
}

func InitChallenges() (tea.Model, tea.Cmd) {
	t := table.New(
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	setTableSize(&t)

	return challengesModel{
		help:        help.New(),
		screensHelp: help.New(),
		table:       t,
	}, tea.Batch(fetchChallenges())
}

func (m challengesModel) Init() tea.Cmd { return nil }

func (m challengesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Default().Printf("Challenges view received message: %v, %T\n", msg, msg)
	switch msg := msg.(type) {
	case updateChallengesCmd:
		var cmd tea.Cmd
		m.table.SetRows(createRows(msg.challenges))
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Reload):
			m.err = ""
			return m, fetchChallenges()
		case key.Matches(msg, constants.Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, constants.Keymap.Enter):
			curr := m.table.SelectedRow()
			id, _ := strconv.Atoi(curr[0])
			challenge, initCmd := InitChallenge(id)
			return challenge, initCmd
		case key.Matches(msg, constants.ScreensKeymap.Scoreboard):
			view, initCmd := InitScoreboard()
			m, updateCmd := view.Update(constants.WindowSize)
			return m, tea.Batch(updateCmd, initCmd)
		}
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		setTableSize(&m.table)
		return m, nil
	case errMsg:
		log.Default().Print(msg)
		m.err = msg.Error()
	}
	cmds := make([]tea.Cmd, 2)
	m.table, cmds[0] = m.table.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m challengesModel) View() string {
	helpText := lipgloss.JoinHorizontal(lipgloss.Top, constants.HelpStyle(m.table.HelpView()), constants.HelpStyle(" • "), constants.HelpStyle(m.help.View(ChallengesKeymap)))
	screensHelpText := lipgloss.JoinHorizontal(lipgloss.Top, constants.HelpStyle(m.screensHelp.View(constants.ScreensKeymap)))

	return lipgloss.JoinVertical(lipgloss.Top, constants.BaseStyle.Render(m.table.View()),
		screensHelpText, helpText, constants.ErrStyle(m.err))
}
