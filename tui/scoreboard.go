package tui

import (
	"context"
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

type scoreboardUpdatedMsg struct {
	scoreboard []api.ScoreboardEntry
}

type scoreboardModel struct {
	scoreboard  table.Model
	help        help.Model
	screensHelp help.Model
	err         error
	width       int
	height      int
}

func fetchScoreboardCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), constants.Timeout)
		defer cancel()
		log.Default().Println("Fetching scoreboard...")
		scoreboard, err := constants.C.GetScoreboard(ctx)
		if err != nil {
			return createErrMsg(fmt.Errorf("Failed to fetch scoreboard: %v", err))
		}
		log.Default().Println("Fetched scoreboard")
		return scoreboardUpdatedMsg{scoreboard}
	}
}

func setScoreboardTableSize(t *table.Model, width, height int) {
	if height != 0 {
		nameLength := width - 35

		columns := []table.Column{
			{Title: "Position", Width: 8},
			{Title: "Name", Width: nameLength},
			{Title: "Score", Width: 8},
			{Title: "Members", Width: 8},
		}

		t.SetColumns(columns)

		top, right, bottom, left := constants.DocStyle.GetMargin()
		t.SetHeight(height - top - bottom - 5)
		t.SetWidth(width - left - right + 1)
	}
}

func createScoreboardRows(scoreboard []api.ScoreboardEntry) []table.Row {
	rows := make([]table.Row, len(scoreboard))

	for i, entry := range scoreboard {
		rows[i] = table.Row{fmt.Sprintf("%d", entry.Position), entry.Name, fmt.Sprintf("%d", entry.Score), fmt.Sprintf("%d", len(entry.Members))}
	}

	return rows
}

func InitScoreboard(width, height int) (scoreboardModel, tea.Cmd) {
	t := table.New(
		table.WithFocused(true),
	)

	s := constants.TableStyle
	s.Header = constants.TableHeaderStyle
	s.Selected = constants.SelectedRowStyle
	t.SetStyles(s)
	setScoreboardTableSize(&t, width, height)

	return scoreboardModel{
		scoreboard:  t,
		help:        help.New(),
		screensHelp: help.New(),
		err:         nil,
		width:       width,
		height:      height,
	}, fetchScoreboardCmd()
}

func (m scoreboardModel) Init() tea.Cmd { return nil }

func (m scoreboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Default().Printf("Scoreboard view received message: %v, %T\n", msg, msg)
	switch msg := msg.(type) {
	case scoreboardUpdatedMsg:
		m.scoreboard.SetRows(createScoreboardRows(msg.scoreboard))
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		setScoreboardTableSize(&m.scoreboard, m.width, m.height)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Reload):
			return m, fetchScoreboardCmd()
		case key.Matches(msg, constants.Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, constants.ScreensKeymap.Challenges):
			cm, initCmd := InitChallenges(m.width, m.height)
			return cm, initCmd
		}
	case errMsg:
		log.Default().Print(msg)
		m.err = msg
	}

	var cmd tea.Cmd
	m.scoreboard, cmd = m.scoreboard.Update(msg)

	return m, cmd
}

func (m scoreboardModel) View() string {
	screensHelpText := lipgloss.JoinHorizontal(lipgloss.Top, constants.HelpStyle(m.screensHelp.View(constants.ScreensKeymap)))
	helpText := lipgloss.JoinHorizontal(lipgloss.Top, constants.HelpStyle(m.scoreboard.HelpView()), constants.HelpStyle(" • "), constants.HelpStyle(m.help.View(ScoreboardKeymap)))

	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Top, constants.BaseStyle.Render(m.scoreboard.View()), screensHelpText, helpText, constants.ErrStyle(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Top, constants.BaseStyle.Render(m.scoreboard.View()), screensHelpText, helpText)
}
