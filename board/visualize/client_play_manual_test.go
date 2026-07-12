package visualize

import (
	"strings"
	"testing"

	"pips-solver/backend/board/types"

	tea "github.com/charmbracelet/bubbletea"
)

func clientTestCell(row, col int) types.Cell {
	return types.Cell{R: row, C: col}
}

func newClientManualPlayGame() *GameState {
	existing := map[types.Cell]bool{
		clientTestCell(0, 0): true,
		clientTestCell(0, 1): true,
		clientTestCell(1, 0): true,
		clientTestCell(1, 1): true,
	}

	cells := make(map[types.Cell]CellState)
	for cell := range existing {
		cells[cell] = CellState{
			Exists:   true,
			RegionID: 0,
		}
	}

	return &GameState{
		Date:       "2026-07-08",
		Difficulty: "test",

		Puzzle: types.Puzzle{
			Solution: [][][]int{
				{{0, 0}, {0, 1}},
				{{1, 0}, {1, 1}},
			},
		},

		Geometry: BoardGeometry{
			MinRow:        0,
			MaxRow:        1,
			MinCol:        0,
			MaxCol:        1,
			ExistingCells: existing,
		},

		Cells: cells,

		Dominoes: []DominoState{
			{
				ID: 0,
				A:  1,
				B:  2,
				V1: 1,
				V2: 2,
			},
			{
				ID: 1,
				A:  3,
				B:  4,
				V1: 3,
				V2: 4,
			},
		},

		Cursor:           clientTestCell(0, 0),
		SelectedDominoID: 0,
		Orientation:      HorizontalRight,
	}
}

func newClientManualPlayModel() Model {
	return Model{
		screen: screenGame,
		game:   newClientManualPlayGame(),
	}
}

func sendClientKey(t *testing.T, m Model, key string) Model {
	t.Helper()

	updated, cmd := m.Update(clientKeyMsg(key))
	if cmd != nil {
		t.Fatalf("expected no command for key %q, got non-nil command", key)
	}

	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model has type %T, want visualize.Model", updated)
	}

	return next
}

func clientKeyMsg(key string) tea.KeyMsg {
	switch key {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune(key),
		}
	}
}

func TestClientManualPlayMoveCursorWithArrowKeys(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "right")

	if m.game.Cursor != clientTestCell(0, 1) {
		t.Fatalf("cursor = %+v, want r0c1", m.game.Cursor)
	}

	m = sendClientKey(t, m, "down")

	if m.game.Cursor != clientTestCell(1, 1) {
		t.Fatalf("cursor = %+v, want r1c1", m.game.Cursor)
	}

	m = sendClientKey(t, m, "left")

	if m.game.Cursor != clientTestCell(1, 0) {
		t.Fatalf("cursor = %+v, want r1c0", m.game.Cursor)
	}

	m = sendClientKey(t, m, "up")

	if m.game.Cursor != clientTestCell(0, 0) {
		t.Fatalf("cursor = %+v, want r0c0", m.game.Cursor)
	}
}

func TestClientManualPlayMoveCursorWithVimKeys(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "l")

	if m.game.Cursor != clientTestCell(0, 1) {
		t.Fatalf("cursor = %+v, want r0c1", m.game.Cursor)
	}

	m = sendClientKey(t, m, "j")

	if m.game.Cursor != clientTestCell(1, 1) {
		t.Fatalf("cursor = %+v, want r1c1", m.game.Cursor)
	}

	m = sendClientKey(t, m, "h")

	if m.game.Cursor != clientTestCell(1, 0) {
		t.Fatalf("cursor = %+v, want r1c0", m.game.Cursor)
	}

	m = sendClientKey(t, m, "k")

	if m.game.Cursor != clientTestCell(0, 0) {
		t.Fatalf("cursor = %+v, want r0c0", m.game.Cursor)
	}
}

func TestClientManualPlayMoveCursorRejectsMissingCell(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "left")

	if m.game.Cursor != clientTestCell(0, 0) {
		t.Fatalf("cursor = %+v, want it to remain r0c0", m.game.Cursor)
	}

	if !strings.Contains(m.game.Message, "outside") {
		t.Fatalf("message = %q, want outside-board message", m.game.Message)
	}
}

func TestClientManualPlayRotateWithR(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "r")

	if m.game.Orientation != VerticalDown {
		t.Fatalf("orientation = %v, want VerticalDown", m.game.Orientation)
	}

	m = sendClientKey(t, m, "r")

	if m.game.Orientation != HorizontalLeft {
		t.Fatalf("orientation = %v, want HorizontalLeft", m.game.Orientation)
	}
}

func TestClientManualPlayFlipWithF(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "f")

	d := m.game.Dominoes[0]
	if d.V1 != 2 || d.V2 != 1 {
		t.Fatalf("selected domino values = %d,%d, want 2,1", d.V1, d.V2)
	}
}

func TestClientManualPlayPlaceWithEnter(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "enter")

	d := m.game.Dominoes[0]
	if !d.Placed {
		t.Fatalf("domino 0 should be placed")
	}

	if d.C1 == nil || *d.C1 != clientTestCell(0, 0) {
		t.Fatalf("domino C1 = %+v, want r0c0", d.C1)
	}

	if d.C2 == nil || *d.C2 != clientTestCell(0, 1) {
		t.Fatalf("domino C2 = %+v, want r0c1", d.C2)
	}

	c1 := m.game.Cells[clientTestCell(0, 0)]
	if c1.Value == nil || *c1.Value != 1 {
		t.Fatalf("cell r0c0 value = %v, want 1", c1.Value)
	}

	c2 := m.game.Cells[clientTestCell(0, 1)]
	if c2.Value == nil || *c2.Value != 2 {
		t.Fatalf("cell r0c1 value = %v, want 2", c2.Value)
	}
}

func TestClientManualPlayPlaceWithSpace(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, " ")

	if !m.game.Dominoes[0].Placed {
		t.Fatalf("domino 0 should be placed")
	}
}

func TestClientManualPlayPlaceReportsInvalidPlacement(t *testing.T) {
	m := newClientManualPlayModel()

	m.game.Cursor = clientTestCell(0, 1)
	m.game.Orientation = HorizontalRight

	m = sendClientKey(t, m, "enter")

	if m.game.Dominoes[0].Placed {
		t.Fatalf("domino should not be placed outside board")
	}

	if m.game.Message == "" {
		t.Fatalf("expected error message after invalid placement")
	}
}

func TestClientManualPlayRemoveWithX(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "enter")
	m = sendClientKey(t, m, "x")

	if m.game.Dominoes[0].Placed {
		t.Fatalf("domino 0 should be removed")
	}

	c1 := m.game.Cells[clientTestCell(0, 0)]
	if c1.Value != nil || c1.DominoID != nil || c1.IsHead {
		t.Fatalf("cell r0c0 should be cleared, got %+v", c1)
	}

	c2 := m.game.Cells[clientTestCell(0, 1)]
	if c2.Value != nil || c2.DominoID != nil || c2.IsHead {
		t.Fatalf("cell r0c1 should be cleared, got %+v", c2)
	}
}

func TestClientManualPlayRemoveWithBackspace(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "enter")
	m = sendClientKey(t, m, "backspace")

	if m.game.Dominoes[0].Placed {
		t.Fatalf("domino 0 should be removed")
	}
}

func TestClientManualPlayRemoveReportsEmptyCell(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "x")

	if m.game.Message == "" {
		t.Fatalf("expected message after removing from empty cell")
	}
}

func TestClientManualPlayTabSelectsNextUnplacedDomino(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "tab")

	if m.game.SelectedDominoID != 1 {
		t.Fatalf("SelectedDominoID = %d, want 1", m.game.SelectedDominoID)
	}
}

func TestClientManualPlayShiftTabSelectsPreviousUnplacedDomino(t *testing.T) {
	m := newClientManualPlayModel()
	m.game.SelectedDominoID = 1

	m = sendClientKey(t, m, "shift+tab")

	if m.game.SelectedDominoID != 0 {
		t.Fatalf("SelectedDominoID = %d, want 0", m.game.SelectedDominoID)
	}
}

func TestClientManualPlayTabSkipsPlacedDomino(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "enter")
	m = sendClientKey(t, m, "tab")

	if m.game.SelectedDominoID != 1 {
		t.Fatalf("SelectedDominoID = %d, want 1", m.game.SelectedDominoID)
	}
}

func TestClientManualPlayNumberKeySelectsVisibleDomino(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "2")

	if m.game.SelectedDominoID != 1 {
		t.Fatalf("SelectedDominoID = %d, want 1", m.game.SelectedDominoID)
	}
}

func TestClientManualPlayNumberKeyCountsOnlyUnplacedDominoes(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "enter")
	m = sendClientKey(t, m, "1")

	if m.game.SelectedDominoID != 1 {
		t.Fatalf("SelectedDominoID = %d, want 1 because domino 0 is placed", m.game.SelectedDominoID)
	}
}

func TestClientManualPlayVerifyMatchingSolution(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "enter")

	m.game.SelectedDominoID = 1
	m.game.Cursor = clientTestCell(1, 0)
	m.game.Orientation = HorizontalRight

	m = sendClientKey(t, m, "enter")
	m = sendClientKey(t, m, "v")

	if !strings.Contains(m.game.Message, "Verified") {
		t.Fatalf("message = %q, want verified message", m.game.Message)
	}
}

func TestClientManualPlayVerifyNonMatchingSolution(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "enter")
	m = sendClientKey(t, m, "v")

	if !strings.Contains(m.game.Message, "Does not match") {
		t.Fatalf("message = %q, want non-matching message", m.game.Message)
	}
}

func TestClientManualPlayDReturnsToDifficultyScreen(t *testing.T) {
	m := newClientManualPlayModel()

	m = sendClientKey(t, m, "d")

	if m.screen != screenDifficulty {
		t.Fatalf("screen = %v, want screenDifficulty", m.screen)
	}
}
