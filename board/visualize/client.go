package visualize

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"pips-solver/backend/board/types"
)

type Puzzle = types.Puzzle
type Cell = types.Cell

type screen int

const (
	screenDateSearch screen = iota
	screenGame
)

type Board struct {
	Grid [][]string
}

func NewBoard(p Puzzle) *Board {
	maxRow := 0
	maxCol := 0

	for _, region := range p.Regions {
		for _, coord := range region.Indices {
			if len(coord) < 2 {
				continue
			}

			if coord[0] > maxRow {
				maxRow = coord[0]
			}

			if coord[1] > maxCol {
				maxCol = coord[1]
			}
		}
	}

	grid := make([][]string, maxRow+1)
	for r := range grid {
		grid[r] = make([]string, maxCol+1)
	}

	return &Board{Grid: grid}
}

type Model struct {
	screen screen

	dateInput textinput.Model
	help      help.Model
	keys      keyMap

	selectedDate string
	cursor       Cell

	width  int
	height int
}

func NewModel() Model {
	input := textinput.New()
	input.Placeholder = "YYYY-MM-DD"
	input.Focus()
	input.CharLimit = 10
	input.Width = 12

	return Model{
		screen:    screenDateSearch,
		dateInput: input,
		help:      help.New(),
		keys:      newKeyMap(),
		cursor:    Cell{R: 0, C: 0},
	}
}

func Run() error {
	_, err := tea.NewProgram(NewModel(), tea.WithAltScreen()).Run()
	return err
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	switch m.screen {
	case screenDateSearch:
		return m.updateDateSearch(msg)
	case screenGame:
		return m.updateGame(msg)
	default:
		return m, nil
	}
}

func (m Model) updateDateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Submit) {
			m.selectedDate = strings.TrimSpace(m.dateInput.Value())
			if m.selectedDate == "" {
				m.selectedDate = "today"
			}

			m.screen = screenGame
			m.cursor = Cell{R: 0, C: 0}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.dateInput, cmd = m.dateInput.Update(msg)
	return m, cmd
}

func (m Model) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			m.screen = screenDateSearch
			m.dateInput.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Left):
			if m.cursor.C > 0 {
				m.cursor.C--
			}
		case key.Matches(msg, m.keys.Right):
			if m.cursor.C < 3 {
				m.cursor.C++
			}
		case key.Matches(msg, m.keys.Up):
			if m.cursor.R > 0 {
				m.cursor.R--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor.R < 3 {
				m.cursor.R++
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenDateSearch:
		return m.dateSearchView()
	case screenGame:
		return m.gameView()
	default:
		return ""
	}
}

func (m Model) dateSearchView() string {
	panel := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Pips"),
		mutedStyle.Render("Choose a puzzle date."),
		"",
		labelStyle.Render("Date"),
		m.dateInput.View(),
	)

	return pageStyle.
		Width(contentWidth(m.width)).
		Render(lipgloss.JoinVertical(lipgloss.Left, panel, "", footerStyle.Render(m.help.View(m.keys))))
}

func (m Model) gameView() string {
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		titleStyle.Render("Pips"),
		mutedStyle.MarginLeft(2).Render(fmt.Sprintf("Date: %s", m.selectedDate)),
	)

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderBoard(),
		m.renderSidePanel(),
	)

	return pageStyle.
		Width(contentWidth(m.width)).
		Render(lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footerStyle.Render(m.help.View(m.keys))))
}

func (m Model) renderBoard() string {
	rows := make([]string, 0, 4)

	for r := 0; r < 4; r++ {
		cells := make([]string, 0, 4)

		for c := 0; c < 4; c++ {
			label := " "
			if r == 0 && c == 0 {
				label = "S8"
			}
			if r == 1 && c == 2 {
				label = "<6"
			}
			if r == 2 && c == 1 {
				label = "!="
			}

			value := " "
			if r == m.cursor.R && c == m.cursor.C {
				value = "•"
			}

			style := cellStyle
			if r == m.cursor.R && c == m.cursor.C {
				style = cursorCellStyle
			}

			cells = append(cells, style.Render(centerCell(value), centerCell(label)))
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderSidePanel() string {
	lines := []string{
		sideTitleStyle.Render("Dominoes"),
		mutedStyle.Render("[0|0] [1|2] [3|4]"),
		"",
		sideTitleStyle.Render("Regions"),
		mutedStyle.Render("S8  sum equals 8"),
		mutedStyle.Render("<6  sum less than 6"),
		mutedStyle.Render("!=  all values differ"),
	}

	return sidePanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func centerCell(s string) string {
	width := runewidth.StringWidth(s)
	if width >= 5 {
		return s
	}

	left := (5 - width) / 2
	right := 5 - width - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func contentWidth(width int) int {
	if width <= 0 {
		return 80
	}

	if width < 48 {
		return width
	}

	return 80
}

type keyMap struct {
	Submit key.Binding
	Back   key.Binding
	Quit   key.Binding
	Left   key.Binding
	Right  key.Binding
	Up     key.Binding
	Down   key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Back:   key.NewBinding(key.WithKeys("d", "esc"), key.WithHelp("d/esc", "date")),
		Quit:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Left:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:  key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Back, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.Submit, k.Back, k.Quit},
	}
}

var (
	pageStyle = lipgloss.NewStyle().
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("29")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	cellStyle = lipgloss.NewStyle().
			Width(7).
			Height(2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")).
			Align(lipgloss.Center)

	cursorCellStyle = cellStyle.Copy().
			BorderForeground(lipgloss.Color("214")).
			Background(lipgloss.Color("236"))

	sidePanelStyle = lipgloss.NewStyle().
			MarginLeft(3).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Width(26)

	sideTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230"))
)
