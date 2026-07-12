package visualize

import (
	"pips-solver/backend/board/types"

	tea "github.com/charmbracelet/bubbletea"
)

func RunManualDemo() error {
	p := tea.NewProgram(NewManualDemoModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func NewManualDemoModel() Model {
	return Model{
		screen:  screenGame,
		message: "Manual demo loaded.",
		game:    newManualDemoGame(),
	}
}

func newManualDemoGame() *GameState {
	cell00 := types.Cell{R: 0, C: 0}
	cell01 := types.Cell{R: 0, C: 1}
	cell10 := types.Cell{R: 1, C: 0}
	cell11 := types.Cell{R: 1, C: 1}

	existing := map[types.Cell]bool{
		cell00: true,
		cell01: true,
		cell10: true,
		cell11: true,
	}

	cells := map[types.Cell]CellState{
		cell00: {Exists: true, RegionID: 0},
		cell01: {Exists: true, RegionID: 0},
		cell10: {Exists: true, RegionID: 1},
		cell11: {Exists: true, RegionID: 1},
	}

	return &GameState{
		Date:       "manual-demo",
		Difficulty: "demo",

		Puzzle: types.Puzzle{
			// Verification target:
			// top row:    r0c0-r0c1
			// bottom row: r1c0-r1c1
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

		Cursor:           cell00,
		SelectedDominoID: 0,
		Orientation:      HorizontalRight,
		Message:          "Manual demo: place both dominoes horizontally, then press v.",
	}
}
