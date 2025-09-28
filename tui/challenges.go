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

type challengesFetchedMsg struct {
	challenges []api.ListChallenge
}

type challengesModel struct {
	table       table.Model
	help        help.Model
	screensHelp help.Model
	err         error
	width       int
	height      int
}

func fetchChallengesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), constants.Timeout)
		defer cancel()
		log.Default().Print("Fetching challenges...")
		challenges, err := constants.C.GetChallenges(ctx)
		if err != nil {
			return createErrMsg(fmt.Errorf("Failed to fetch challenges: %v", err))
		}
		log.Default().Print("Fetched challenges")
		return challengesFetchedMsg{challenges}
	}
}

func setTableSize(t *table.Model, width, height int) {
	if height != 0 {
		nameLength := ((width - 17) / 4) * 2
		categoryLength := width - 30 - nameLength

		columns := []table.Column{
			{Title: "ID", Width: 5},
			{Title: "Name", Width: nameLength},
			{Title: "Category", Width: categoryLength},
			{Title: "Value", Width: 5},
			{Title: "Solved", Width: 7},
		}

		t.SetColumns(columns)

		top, right, bottom, left := constants.DocStyle.GetMargin()

		t.SetHeight(height - top - bottom - 5)
		t.SetWidth(width - left - right + 1)
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

func InitChallenges(width, height int) (tea.Model, tea.Cmd) {
	t := table.New(
		table.WithFocused(true),
	)

	s := constants.TableStyle
	s.Header = constants.TableHeaderStyle
	s.Selected = constants.SelectedRowStyle
	t.SetStyles(s)
	setTableSize(&t, width, height)

	return challengesModel{
		help:        help.New(),
		screensHelp: help.New(),
		table:       t,
		width:       width,
		height:      height,
	}, tea.Batch(fetchChallengesCmd())
}

func (m challengesModel) Init() tea.Cmd { return nil }

func (m challengesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Default().Printf("Challenges view received message: %v, %T\n", msg, msg)
	switch msg := msg.(type) {
	case challengesFetchedMsg:
		m.table.SetRows(createRows(msg.challenges))
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Reload):
			m.err = nil
			return m, fetchChallengesCmd()
		case key.Matches(msg, constants.Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, constants.Keymap.Enter):
			curr := m.table.SelectedRow()
			if curr == nil {
				return m, nil
			}
			id, _ := strconv.Atoi(curr[0])
			challenge, initCmd := InitChallenge(id, m.width, m.height)
			return challenge, initCmd
		case key.Matches(msg, constants.ScreensKeymap.Scoreboard):
			sbm, initCmd := InitScoreboard(m.width, m.height)
			return sbm, initCmd
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		setTableSize(&m.table, m.width, m.height)
		return m, nil
	case errMsg:
		log.Default().Print(msg)
		m.err = msg
	}
	cmds := make([]tea.Cmd, 2)
	m.table, cmds[0] = m.table.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m challengesModel) View() string {
	helpText := lipgloss.JoinHorizontal(lipgloss.Top, constants.HelpStyle(m.table.HelpView()), constants.HelpStyle(" • "), constants.HelpStyle(m.help.View(ChallengesKeymap)))
	screensHelpText := lipgloss.JoinHorizontal(lipgloss.Top, constants.HelpStyle(m.screensHelp.View(constants.ScreensKeymap)))
	errStr := renderError(m.err)

	return lipgloss.JoinVertical(lipgloss.Top, constants.BaseStyle.Render(m.table.View()),
		screensHelpText, helpText, errStr)
}
