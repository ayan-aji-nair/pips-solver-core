package visualize

import (
	"testing"

	"pips-solver/backend/board/types"
)

func testCell(row, col int) types.Cell {
	return types.Cell{R: row, C: col}
}

func newInputTestGame() *GameState {
	existing := map[types.Cell]bool{
		testCell(0, 0): true,
		testCell(0, 1): true,
		testCell(1, 0): true,
		testCell(1, 1): true,
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

		Cursor:           testCell(0, 0),
		SelectedDominoID: 0,
		Orientation:      HorizontalRight,
	}
}

func TestOrientationOffset(t *testing.T) {
	tests := []struct {
		name        string
		orientation Orientation
		want        types.Cell
	}{
		{
			name:        "horizontal right",
			orientation: HorizontalRight,
			want:        testCell(0, 1),
		},
		{
			name:        "vertical down",
			orientation: VerticalDown,
			want:        testCell(1, 0),
		},
		{
			name:        "horizontal left",
			orientation: HorizontalLeft,
			want:        testCell(0, -1),
		},
		{
			name:        "vertical up",
			orientation: VerticalUp,
			want:        testCell(-1, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orientationOffset(tt.orientation)
			if got != tt.want {
				t.Fatalf("orientationOffset(%v) = %+v, want %+v", tt.orientation, got, tt.want)
			}
		})
	}
}

func TestAddCell(t *testing.T) {
	got := addCell(testCell(2, 3), testCell(-1, 4))
	want := testCell(1, 7)

	if got != want {
		t.Fatalf("addCell returned %+v, want %+v", got, want)
	}
}

func TestRotateCyclesThroughOrientations(t *testing.T) {
	g := newInputTestGame()

	if g.Orientation != HorizontalRight {
		t.Fatalf("initial orientation = %v, want HorizontalRight", g.Orientation)
	}

	g.Rotate()
	if g.Orientation != VerticalDown {
		t.Fatalf("after 1 rotate = %v, want VerticalDown", g.Orientation)
	}

	g.Rotate()
	if g.Orientation != HorizontalLeft {
		t.Fatalf("after 2 rotates = %v, want HorizontalLeft", g.Orientation)
	}

	g.Rotate()
	if g.Orientation != VerticalUp {
		t.Fatalf("after 3 rotates = %v, want VerticalUp", g.Orientation)
	}

	g.Rotate()
	if g.Orientation != HorizontalRight {
		t.Fatalf("after 4 rotates = %v, want HorizontalRight", g.Orientation)
	}
}

func TestFlipSelectedSwapsSelectedDominoValues(t *testing.T) {
	g := newInputTestGame()

	g.FlipSelected()

	d := g.Dominoes[0]
	if d.V1 != 2 || d.V2 != 1 {
		t.Fatalf("after flip got V1=%d V2=%d, want V1=2 V2=1", d.V1, d.V2)
	}

	g.FlipSelected()

	d = g.Dominoes[0]
	if d.V1 != 1 || d.V2 != 2 {
		t.Fatalf("after second flip got V1=%d V2=%d, want V1=1 V2=2", d.V1, d.V2)
	}
}

func TestFlipSelectedIgnoresInvalidSelectedDominoID(t *testing.T) {
	g := newInputTestGame()
	g.SelectedDominoID = 99

	g.FlipSelected()

	if g.Dominoes[0].V1 != 1 || g.Dominoes[0].V2 != 2 {
		t.Fatalf("domino 0 changed unexpectedly: %+v", g.Dominoes[0])
	}

	if g.Dominoes[1].V1 != 3 || g.Dominoes[1].V2 != 4 {
		t.Fatalf("domino 1 changed unexpectedly: %+v", g.Dominoes[1])
	}
}

func TestPlaceSelectedHorizontalRight(t *testing.T) {
	g := newInputTestGame()

	err := g.PlaceSelected()
	if err != nil {
		t.Fatalf("PlaceSelected returned error: %v", err)
	}

	d := g.Dominoes[0]
	if !d.Placed {
		t.Fatalf("domino was not marked placed")
	}

	if d.C1 == nil || *d.C1 != testCell(0, 0) {
		t.Fatalf("domino C1 = %+v, want r0c0", d.C1)
	}

	if d.C2 == nil || *d.C2 != testCell(0, 1) {
		t.Fatalf("domino C2 = %+v, want r0c1", d.C2)
	}

	c1 := g.Cells[testCell(0, 0)]
	if c1.Value == nil || *c1.Value != 1 {
		t.Fatalf("cell r0c0 value = %v, want 1", c1.Value)
	}

	if c1.DominoID == nil || *c1.DominoID != 0 {
		t.Fatalf("cell r0c0 domino id = %v, want 0", c1.DominoID)
	}

	if !c1.IsHead {
		t.Fatalf("cell r0c0 should be marked as head")
	}

	c2 := g.Cells[testCell(0, 1)]
	if c2.Value == nil || *c2.Value != 2 {
		t.Fatalf("cell r0c1 value = %v, want 2", c2.Value)
	}

	if c2.DominoID == nil || *c2.DominoID != 0 {
		t.Fatalf("cell r0c1 domino id = %v, want 0", c2.DominoID)
	}

	if c2.IsHead {
		t.Fatalf("cell r0c1 should not be marked as head")
	}
}

func TestPlaceSelectedVerticalDown(t *testing.T) {
	g := newInputTestGame()
	g.Orientation = VerticalDown

	err := g.PlaceSelected()
	if err != nil {
		t.Fatalf("PlaceSelected returned error: %v", err)
	}

	d := g.Dominoes[0]

	if d.C1 == nil || *d.C1 != testCell(0, 0) {
		t.Fatalf("domino C1 = %+v, want r0c0", d.C1)
	}

	if d.C2 == nil || *d.C2 != testCell(1, 0) {
		t.Fatalf("domino C2 = %+v, want r1c0", d.C2)
	}
}

func TestPlaceSelectedUsesFlippedValues(t *testing.T) {
	g := newInputTestGame()

	g.FlipSelected()

	err := g.PlaceSelected()
	if err != nil {
		t.Fatalf("PlaceSelected returned error: %v", err)
	}

	c1 := g.Cells[testCell(0, 0)]
	c2 := g.Cells[testCell(0, 1)]

	if c1.Value == nil || *c1.Value != 2 {
		t.Fatalf("cell r0c0 value = %v, want 2", c1.Value)
	}

	if c2.Value == nil || *c2.Value != 1 {
		t.Fatalf("cell r0c1 value = %v, want 1", c2.Value)
	}
}

func TestPlaceSelectedRejectsInvalidSelectedDominoID(t *testing.T) {
	g := newInputTestGame()
	g.SelectedDominoID = 99

	err := g.PlaceSelected()
	if err == nil {
		t.Fatalf("PlaceSelected succeeded with invalid selected domino id")
	}
}

func TestPlaceSelectedRejectsOutsideBoard(t *testing.T) {
	g := newInputTestGame()
	g.Cursor = testCell(0, 1)
	g.Orientation = HorizontalRight

	err := g.PlaceSelected()
	if err == nil {
		t.Fatalf("PlaceSelected succeeded, want outside-board error")
	}

	if g.Dominoes[0].Placed {
		t.Fatalf("domino should not be marked placed after failed placement")
	}

	c := g.Cells[testCell(0, 1)]
	if c.Value != nil || c.DominoID != nil {
		t.Fatalf("cell should remain empty after failed placement: %+v", c)
	}
}

func TestPlaceSelectedRejectsOccupiedCell(t *testing.T) {
	g := newInputTestGame()

	err := g.PlaceSelected()
	if err != nil {
		t.Fatalf("first PlaceSelected returned error: %v", err)
	}

	g.SelectedDominoID = 1
	g.Cursor = testCell(0, 1)
	g.Orientation = VerticalDown

	err = g.PlaceSelected()
	if err == nil {
		t.Fatalf("second PlaceSelected succeeded, want occupied-cell error")
	}

	if g.Dominoes[1].Placed {
		t.Fatalf("second domino should not be marked placed")
	}
}

func TestPlaceSelectedRejectsAlreadyPlacedSelectedDomino(t *testing.T) {
	g := newInputTestGame()

	err := g.PlaceSelected()
	if err != nil {
		t.Fatalf("first PlaceSelected returned error: %v", err)
	}

	g.Cursor = testCell(1, 0)
	g.Orientation = HorizontalRight

	err = g.PlaceSelected()
	if err == nil {
		t.Fatalf("second PlaceSelected succeeded, want already-placed error")
	}
}

func TestRemoveAtCursorClearsBothCellsFromHeadCell(t *testing.T) {
	g := newInputTestGame()

	err := g.PlaceSelected()
	if err != nil {
		t.Fatalf("PlaceSelected returned error: %v", err)
	}

	g.Cursor = testCell(0, 0)

	err = g.RemoveAtCursor()
	if err != nil {
		t.Fatalf("RemoveAtCursor returned error: %v", err)
	}

	d := g.Dominoes[0]
	if d.Placed {
		t.Fatalf("domino should be marked unplaced")
	}

	if d.C1 != nil || d.C2 != nil {
		t.Fatalf("domino cells should be nil after removal, got C1=%+v C2=%+v", d.C1, d.C2)
	}

	c1 := g.Cells[testCell(0, 0)]
	if c1.Value != nil || c1.DominoID != nil || c1.IsHead {
		t.Fatalf("cell r0c0 was not fully cleared: %+v", c1)
	}

	c2 := g.Cells[testCell(0, 1)]
	if c2.Value != nil || c2.DominoID != nil || c2.IsHead {
		t.Fatalf("cell r0c1 was not fully cleared: %+v", c2)
	}

	if g.SelectedDominoID != 0 {
		t.Fatalf("selected domino id = %d, want 0", g.SelectedDominoID)
	}
}

func TestRemoveAtCursorClearsBothCellsFromTailCell(t *testing.T) {
	g := newInputTestGame()

	err := g.PlaceSelected()
	if err != nil {
		t.Fatalf("PlaceSelected returned error: %v", err)
	}

	g.Cursor = testCell(0, 1)

	err = g.RemoveAtCursor()
	if err != nil {
		t.Fatalf("RemoveAtCursor returned error: %v", err)
	}

	if g.Dominoes[0].Placed {
		t.Fatalf("domino should be marked unplaced")
	}

	if g.Cells[testCell(0, 0)].DominoID != nil {
		t.Fatalf("head cell should be cleared")
	}

	if g.Cells[testCell(0, 1)].DominoID != nil {
		t.Fatalf("tail cell should be cleared")
	}
}

func TestRemoveAtCursorRejectsEmptyCell(t *testing.T) {
	g := newInputTestGame()

	err := g.RemoveAtCursor()
	if err == nil {
		t.Fatalf("RemoveAtCursor succeeded, want empty-cell error")
	}
}
