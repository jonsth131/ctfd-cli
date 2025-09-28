package constants

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsth131/ctfd-cli/api"
)

var (
	P *tea.Program
	C api.CTFdAPI
	// WindowSize tea.WindowSizeMsg
)

const (
	Timeout = 5 * time.Second
)

var (
	DocStyle         = lipgloss.NewStyle().Margin(0, 2)
	HelpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	ErrStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#bd534b")).Render
	AlertStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render
	BaseStyle        = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	TableStyle       = table.DefaultStyles()
	TableHeaderStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	SelectedRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	FocusedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	CursorStyle      = FocusedStyle
	NoStyle          = lipgloss.NewStyle()
	SpinnerStyle     = FocusedStyle
)

type keymap struct {
	Enter  key.Binding
	Back   key.Binding
	Reload key.Binding
	Quit   key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Back, k.Reload, k.Quit}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

var Keymap = keymap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
}

type screensKeymap struct {
	Challenges key.Binding
	Scoreboard key.Binding
}

func (k screensKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Challenges, k.Scoreboard}
}

func (k screensKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

var ScreensKeymap = screensKeymap{
	Challenges: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "challenges"),
	),
	Scoreboard: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "scoreboard"),
	),
}
