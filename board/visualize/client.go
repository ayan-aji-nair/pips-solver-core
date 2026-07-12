package visualize

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"pips-solver/backend/board/requests"
	"pips-solver/backend/board/solver"
	"pips-solver/backend/board/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
)

type Puzzle = types.Puzzle
type Cell = types.Cell
type PuzzlePayload = types.PuzzlePayload

type screen int

const (
	screenDateSearch screen = iota
	screenLoading
	screenDifficulty
	screenGame
	screenSolving
)

type Model struct {
	screen screen

	dateInput string
	message   string
	err       error

	payload *types.PuzzlePayload
	game    *GameState
}

type puzzleLoadedMsg struct {
	Payload *types.PuzzlePayload
}

type puzzleLoadFailedMsg struct {
	Err error
}

type solverFinishedMsg struct {
	Placements []types.Placement
	Duration   time.Duration
}

type solverFailedMsg struct {
	Err      error
	Duration time.Duration
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				Padding(1, 2)

	mutedStyle = lipgloss.NewStyle().
			Faint(true)

	errorStyle = lipgloss.NewStyle().
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Faint(true)
)

func Run() error {
	_ = godotenv.Load()

	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func NewModel() Model {
	return Model{
		screen:    screenDateSearch,
		dateInput: "",
		message:   "Enter a date as YYYY-MM-DD, or press enter for today.",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenDateSearch:
		return m.updateDateSearch(msg)
	case screenLoading:
		return m.updateLoading(msg)
	case screenDifficulty:
		return m.updateDifficulty(msg)
	case screenGame:
		return m.updateGame(msg)
	case screenSolving:
		return m.updateSolving(msg)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.screen {
	case screenDateSearch:
		return m.viewDateSearch()
	case screenLoading:
		return m.viewLoading()
	case screenDifficulty:
		return m.viewDifficulty()
	case screenGame:
		return m.viewGame(false)
	case screenSolving:
		return m.viewGame(true)
	default:
		return "unknown screen"
	}
}

func (m Model) updateDateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			date := strings.TrimSpace(m.dateInput)
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}

			m.dateInput = date
			m.message = "Loading puzzle..."
			m.err = nil
			m.screen = screenLoading

			return m, loadPuzzleCmd(date)

		case tea.KeyBackspace:
			if len(m.dateInput) > 0 {
				m.dateInput = m.dateInput[:len(m.dateInput)-1]
			}

		case tea.KeyRunes:
			m.dateInput += string(msg.Runes)
		}
	}

	return m, nil
}

func (m Model) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}

	case puzzleLoadedMsg:
		m.payload = msg.Payload
		m.err = nil
		m.message = "Choose difficulty: e = easy, m = medium, h = hard."
		m.screen = screenDifficulty
		return m, nil

	case puzzleLoadFailedMsg:
		m.err = msg.Err
		m.message = "Failed to load puzzle. Edit the date and try again."
		m.screen = screenDateSearch
		return m, nil
	}

	return m, nil
}

func (m Model) updateDifficulty(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "d", "backspace":
			m.screen = screenDateSearch
			m.message = "Enter a date as YYYY-MM-DD, or press enter for today."
			return m, nil

		case "e", "E":
			return m.buildGameForDifficulty("easy")

		case "m", "M":
			return m.buildGameForDifficulty("medium")

		case "h", "H", "enter":
			return m.buildGameForDifficulty("hard")
		}
	}

	return m, nil
}

func (m Model) buildGameForDifficulty(difficulty string) (tea.Model, tea.Cmd) {
	if m.payload == nil {
		m.err = fmt.Errorf("no puzzle payload loaded")
		m.message = "No puzzle loaded. Return to date search and try again."
		m.screen = screenDateSearch
		return m, nil
	}

	var puzzle types.Puzzle

	switch difficulty {
	case "easy":
		puzzle = m.payload.Data.Easy
	case "medium":
		puzzle = m.payload.Data.Medium
	case "hard":
		puzzle = m.payload.Data.Hard
	default:
		m.err = fmt.Errorf("unknown difficulty %q", difficulty)
		m.message = "Unknown difficulty."
		return m, nil
	}

	game, err := NewGameState(m.dateInput, difficulty, puzzle)
	if err != nil {
		m.err = err
		m.message = "Failed to build game state."
		return m, nil
	}

	m.game = game
	m.err = nil
	m.message = ""
	m.screen = screenGame

	return m, nil
}

func (m Model) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.game == nil {
		m.screen = screenDateSearch
		m.message = "No game loaded."
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "d":
			m.screen = screenDifficulty
			m.message = "Choose difficulty: e = easy, m = medium, h = hard."
			return m, nil

		case "left", "h":
			moveCursor(m.game, 0, -1)

		case "right", "l":
			moveCursor(m.game, 0, 1)

		case "up", "k":
			moveCursor(m.game, -1, 0)

		case "down", "j":
			moveCursor(m.game, 1, 0)

		case "tab":
			selectNextUnplacedDomino(m.game)

		case "shift+tab":
			selectPrevUnplacedDomino(m.game)

		case "r":
			m.game.Rotate()

		case "f":
			m.game.FlipSelected()

		case " ", "enter":
			if err := m.game.PlaceSelected(); err != nil {
				m.game.Message = err.Error()
			}

		case "x", "backspace":
			if err := m.game.RemoveAtCursor(); err != nil {
				m.game.Message = err.Error()
			}

		case "v":
			result := VerifyAgainstAPISolution(
				m.game.Puzzle,
				m.game.CurrentPlacements(),
			)

			if result.MatchesAPISolution {
				m.game.Message = "Verified: matches API solution."
			} else {
				m.game.Message = fmt.Sprintf(
					"Does not match API solution: missing %d edge(s), extra %d edge(s).",
					len(result.MissingEdges),
					len(result.ExtraEdges),
				)
			}

		case "s":
			m.game.Message = "Solving..."
			m.screen = screenSolving
			return m, runSolverCmd(m.game.Puzzle)
		}

		if n, ok := numberKey(msg.String()); ok {
			selectVisibleDomino(m.game, n)
		}
	}

	return m, nil
}

func (m Model) updateSolving(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.game == nil {
		m.screen = screenDateSearch
		m.message = "No game loaded."
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case solverFinishedMsg:
		m.screen = screenGame

		if err := applyPlacementsToGame(m.game, msg.Placements); err != nil {
			m.game.Message = fmt.Sprintf("Solver returned invalid placements: %v", err)
			return m, nil
		}

		result := VerifyAgainstAPISolution(m.game.Puzzle, msg.Placements)

		if result.MatchesAPISolution {
			m.game.Message = fmt.Sprintf("Solved in %s. Matches API solution.", msg.Duration.Round(time.Millisecond))
		} else {
			m.game.Message = fmt.Sprintf(
				"Solved in %s, but differs from API layout: missing %d, extra %d.",
				msg.Duration.Round(time.Millisecond),
				len(result.MissingEdges),
				len(result.ExtraEdges),
			)
		}

		return m, nil

	case solverFailedMsg:
		m.screen = screenGame
		m.game.Message = fmt.Sprintf("Solver failed after %s: %v", msg.Duration.Round(time.Millisecond), msg.Err)
		return m, nil
	}

	return m, nil
}

func loadPuzzleCmd(date string) tea.Cmd {
	return func() tea.Msg {
		baseURL := os.Getenv("BASE_URL")
		token := os.Getenv("TOKEN")

		if baseURL == "" {
			return puzzleLoadFailedMsg{Err: fmt.Errorf("BASE_URL is not set")}
		}

		if token == "" {
			return puzzleLoadFailedMsg{Err: fmt.Errorf("TOKEN is not set")}
		}

		http_client := &http.Client{
			Timeout: 10 * time.Second,
		}

		client, _ := requests.NewClient(baseURL, token, http_client)

		ctx := context.Background()
		payload, err := client.GetPuzzles(ctx, date)
		if err != nil {
			return puzzleLoadFailedMsg{Err: err}
		}

		return puzzleLoadedMsg{Payload: payload}
	}
}

func runSolverCmd(p types.Puzzle) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()

		model, err := solver.NewILPModel(p)
		if err != nil {
			return solverFailedMsg{
				Err:      err,
				Duration: time.Since(start),
			}
		}

		placements, err := model.Solve()
		if err != nil {
			return solverFailedMsg{
				Err:      err,
				Duration: time.Since(start),
			}
		}

		return solverFinishedMsg{
			Placements: placements,
			Duration:   time.Since(start),
		}
	}
}

func (m Model) viewDateSearch() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Pips TUI"))
	b.WriteString("\n\n")
	b.WriteString("Date: ")
	b.WriteString(activePanelStyle.Render(m.dateInput))
	b.WriteString("\n\n")
	b.WriteString(m.message)

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter: load today/date • backspace: edit • esc/ctrl+c: quit"))

	return b.String()
}

func (m Model) viewLoading() string {
	return titleStyle.Render("Pips TUI") +
		"\n\n" +
		panelStyle.Render(fmt.Sprintf("Loading puzzle for %s...", m.dateInput)) +
		"\n\n" +
		helpStyle.Render("esc/ctrl+c: quit")
}

func (m Model) viewDifficulty() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Pips TUI"))
	b.WriteString("\n\n")
	b.WriteString(panelStyle.Render(
		fmt.Sprintf(
			"Loaded puzzle for %s\n\nChoose difficulty:\n\n  e  Easy\n  m  Medium\n  h  Hard\n\nPress enter for hard.",
			m.dateInput,
		),
	))

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("d/backspace: date search • esc/ctrl+c: quit"))

	return b.String()
}

func (m Model) viewGame(solving bool) string {
	if m.game == nil {
		return "No game loaded."
	}

	header := titleStyle.Render(
		fmt.Sprintf(
			"Pips TUI — %s — %s",
			m.game.Date,
			strings.Title(m.game.Difficulty),
		),
	)

	board := renderBoard(m.game)
	side := renderSidePanel(m.game)

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		board,
		"  ",
		side,
	)

	status := m.game.Message
	if status == "" {
		status = m.message
	}
	if solving {
		status = "Solving..."
	}

	footer := helpStyle.Render(
		"h/j/k/l or arrows: move • tab: next domino • shift+tab: prev • 1-9: select • r: rotate • f: flip • space/enter: place • x: remove • v: verify • s: solve • d: difficulty • ctrl+c: quit",
	)

	return header + "\n\n" + body + "\n\n" + panelStyle.Render(status) + "\n\n" + footer
}

func renderBoard(g *GameState) string {
	var rowViews []string

	for r := g.Geometry.MinRow; r <= g.Geometry.MaxRow; r++ {
		var cellViews []string

		for c := g.Geometry.MinCol; c <= g.Geometry.MaxCol; c++ {
			cell := types.Cell{R: r, C: c}

			if !g.Geometry.ExistingCells[cell] {
				cellViews = append(cellViews, lipgloss.NewStyle().
					Width(8).
					Height(4).
					Render(""))
				continue
			}

			cellViews = append(cellViews, renderCell(g, cell))
		}

		rowViews = append(rowViews, lipgloss.JoinHorizontal(lipgloss.Top, cellViews...))
	}

	return strings.Join(rowViews, "\n")
}

func renderCell(g *GameState, cell types.Cell) string {
	cs := g.Cells[cell]

	value := "."
	if cs.Value != nil {
		value = strconv.Itoa(*cs.Value)
	}

	dominoText := ""
	if cs.DominoID != nil {
		dominoText = fmt.Sprintf("D%d", *cs.DominoID+1)
	}

	content := fmt.Sprintf(
		"  %s  \nR%d %s",
		value,
		cs.RegionID+1,
		dominoText,
	)

	style := lipgloss.NewStyle().
		Width(7).
		Height(3).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder())

	if cell == g.Cursor {
		style = style.Border(lipgloss.ThickBorder())
	}

	return style.Render(content)
}

func renderSidePanel(g *GameState) string {
	var b strings.Builder

	b.WriteString("Dominoes\n")
	b.WriteString(renderDominoTray(g))
	b.WriteString("\n\n")
	b.WriteString("Regions\n")
	b.WriteString(renderRegionList(g))
	b.WriteString("\n\n")
	b.WriteString("Cursor\n")
	b.WriteString(fmt.Sprintf("r%dc%d\n", g.Cursor.R, g.Cursor.C))
	b.WriteString("\n")
	b.WriteString("Orientation\n")
	b.WriteString(orientationName(g.Orientation))

	return panelStyle.Width(38).Render(b.String())
}

func renderDominoTray(g *GameState) string {
	var b strings.Builder

	visible := 0

	for i, d := range g.Dominoes {
		status := " "
		if d.Placed {
			status = "x"
		}

		prefix := " "
		if i == g.SelectedDominoID {
			prefix = ">"
		}

		displayNumber := "-"
		if !d.Placed {
			visible++
			if visible <= 9 {
				displayNumber = strconv.Itoa(visible)
			}
		}

		b.WriteString(fmt.Sprintf(
			"%s %s [%s] %d|%d  %s\n",
			prefix,
			displayNumber,
			status,
			d.V1,
			d.V2,
			dominoLocationText(d),
		))
	}

	if len(g.Dominoes) == 0 {
		b.WriteString(mutedStyle.Render("No dominoes"))
	}

	return b.String()
}

func dominoLocationText(d DominoState) string {
	if !d.Placed || d.C1 == nil || d.C2 == nil {
		return ""
	}

	return fmt.Sprintf(
		"r%dc%d-r%dc%d",
		d.C1.R,
		d.C1.C,
		d.C2.R,
		d.C2.C,
	)
}

func renderRegionList(g *GameState) string {
	var b strings.Builder

	for i, region := range g.Puzzle.Regions {
		b.WriteString(fmt.Sprintf("%02d  ", i+1))

		for j, coord := range region.Indices {
			if len(coord) != 2 {
				continue
			}

			if j > 0 {
				b.WriteString(" ")
			}

			b.WriteString(fmt.Sprintf("r%dc%d", coord[0], coord[1]))
		}

		b.WriteString("\n")
	}

	if len(g.Puzzle.Regions) == 0 {
		b.WriteString(mutedStyle.Render("No regions"))
	}

	return b.String()
}
