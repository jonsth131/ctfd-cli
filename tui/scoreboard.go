package tui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsth131/ctfd-cli/api"
	"github.com/jonsth131/ctfd-cli/tui/constants"
)

type scoreboardKeymap struct {
	Reload key.Binding
	Quit   key.Binding
}

func (k scoreboardKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Reload, k.Quit}
}

func (k scoreboardKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

var ScoreboardKeymap = scoreboardKeymap{
	Reload: constants.Keymap.Reload,
	Quit:   constants.Keymap.Quit,
}

type updateScoreboardCmd struct {
	scoreboard []api.ScoreboardEntry
}

type scoreboardModel struct {
	scoreboard  table.Model
	help        help.Model
	screensHelp help.Model
	err         string
}

func fetchScoreboard() tea.Cmd {
	return func() tea.Msg {
		log.Default().Println("Fetching scoreboard...")
		scoreboard, err := constants.C.GetScoreboard()
		if err != nil {
			return errMsg{fmt.Errorf("Failed to fetch scoreboard: %v", err)}
		}
		log.Default().Println("Fetched scoreboard")
		return updateScoreboardCmd{scoreboard}
	}
}

func setScoreboardTableSize(t *table.Model) {
	if constants.WindowSize.Height != 0 {
		nameLength := constants.WindowSize.Width - 35

		columns := []table.Column{
			{Title: "Position", Width: 8},
			{Title: "Name", Width: nameLength},
			{Title: "Score", Width: 8},
			{Title: "Members", Width: 8},
		}

		t.SetColumns(columns)

		top, right, bottom, left := constants.DocStyle.GetMargin()
		t.SetHeight(constants.WindowSize.Height - top - bottom - 4)
		t.SetWidth(constants.WindowSize.Width - left - right + 1)
	}
}

func createScoreboardRows(scoreboard []api.ScoreboardEntry) []table.Row {
	rows := make([]table.Row, len(scoreboard))

	for i, entry := range scoreboard {
		rows[i] = table.Row{fmt.Sprintf("%d", entry.Position), entry.Name, fmt.Sprintf("%d", entry.Score), fmt.Sprintf("%d", len(entry.Members))}
	}

	return rows
}

func InitScoreboard() (scoreboardModel, tea.Cmd) {
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
	setScoreboardTableSize(&t)

	return scoreboardModel{
		scoreboard:  t,
		help:        help.New(),
		screensHelp: help.New(),
		err:         "",
	}, fetchScoreboard()
}

func (m scoreboardModel) Init() tea.Cmd { return nil }

func (m scoreboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Default().Printf("Scoreboard view received message: %v, %T\n", msg, msg)
	switch msg := msg.(type) {
	case updateScoreboardCmd:
		var cmd tea.Cmd
		m.scoreboard.SetRows(createScoreboardRows(msg.scoreboard))
		m.scoreboard, cmd = m.scoreboard.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		setScoreboardTableSize(&m.scoreboard)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Reload):
			return m, fetchScoreboard()
		case key.Matches(msg, constants.Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, constants.ScreensKeymap.Challenges):
			view, initCmd := InitChallenges()
			m, updateCmd := view.Update(constants.WindowSize)
			return m, tea.Batch(updateCmd, initCmd)
		}
	}

	var cmd tea.Cmd
	m.scoreboard, cmd = m.scoreboard.Update(msg)

	return m, cmd
}

func (m scoreboardModel) View() string {
	if len(m.scoreboard.Rows()) == 0 {
		return "Loading scoreboard..."
	}

	screensHelpText := lipgloss.JoinHorizontal(lipgloss.Top, constants.HelpStyle(m.screensHelp.View(constants.ScreensKeymap)))
	helpText := lipgloss.JoinHorizontal(lipgloss.Top, constants.HelpStyle(m.scoreboard.HelpView()), constants.HelpStyle(" • "), constants.HelpStyle(m.help.View(ScoreboardKeymap)))
	return constants.BaseStyle.Render(m.scoreboard.View()) + "\n" + screensHelpText + "\n" + helpText
}
