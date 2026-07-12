package visualize

import (
	"fmt"

	"pips-solver/backend/board/types"
)

type Placement = types.Placement

func orientationOffset(o Orientation) Cell {
	switch o {
	case HorizontalRight:
		return Cell{R: 0, C: 1}
	case VerticalDown:
		return Cell{R: 1, C: 0}
	case HorizontalLeft:
		return Cell{R: 0, C: -1}
	case VerticalUp:
		return Cell{R: -1, C: 0}
	default:
		return Cell{R: 0, C: 1}
	}
}

func addCell(a, b Cell) Cell {
	return Cell{
		R: a.R + b.R,
		C: a.C + b.C,
	}
}

func (g *GameState) Rotate() {
	g.Orientation = Orientation(int(g.Orientation+1) % 4)
}

func (g *GameState) FlipSelected() {
	if g.SelectedDominoID < 0 || g.SelectedDominoID >= len(g.Dominoes) {
		return
	}

	d := &g.Dominoes[g.SelectedDominoID]
	d.V1, d.V2 = d.V2, d.V1
	g.Message = "Flipped domino"
}

func (g *GameState) PlaceSelected() error {
	if g.SelectedDominoID < 0 || g.SelectedDominoID >= len(g.Dominoes) {
		return fmt.Errorf("no selected domain")
	}

	d := &g.Dominoes[g.SelectedDominoID]
	if d.Placed {
		return fmt.Errorf("selected domino is already placed")
	}

	c1 := g.Cursor
	c2 := addCell(c1, orientationOffset(g.Orientation))

	if !g.Geometry.ExistingCells[c1] || !g.Geometry.ExistingCells[c2] {
		return fmt.Errorf("placement outside board")
	}

	if g.Cells[c1].DominoID != nil || g.Cells[c2].DominoID != nil {
		return fmt.Errorf("target cells occupied")
	}

	id := d.ID

	cell1 := g.Cells[c1]
	cell1.Value = &d.V1
	cell1.DominoID = &id
	cell1.IsHead = true
	g.Cells[c1] = cell1

	cell2 := g.Cells[c2]
	cell2.Value = &d.V2
	cell2.DominoID = &id
	cell2.IsHead = false
	g.Cells[c2] = cell2

	d.Placed = true
	d.C1 = &c1
	d.C2 = &c2

	g.Message = "Placed domino"
	return nil
}

func (g *GameState) RemoveAtCursor() error {
	cell := g.Cells[g.Cursor]
	if cell.DominoID == nil {
		return fmt.Errorf("no domino at cursor")
	}

	id := *cell.DominoID
	if id < 0 || id >= len(g.Dominoes) {
		return fmt.Errorf("invalid domino id")
	}

	d := &g.Dominoes[id]

	if d.C1 != nil {
		c1 := *d.C1
		cs := g.Cells[c1]
		cs.Value = nil
		cs.DominoID = nil
		cs.IsHead = false
		g.Cells[c1] = cs
	}

	if d.C2 != nil {
		c2 := *d.C2
		cs := g.Cells[c2]
		cs.Value = nil
		cs.DominoID = nil
		cs.IsHead = false
		g.Cells[c2] = cs
	}

	d.Placed = false
	d.C1 = nil
	d.C2 = nil

	g.SelectedDominoID = id
	g.Message = "Removed domino"

	return nil
}

func (g *GameState) CurrentPlacements() []Placement {
	placements := make([]Placement, 0)

	for _, d := range g.Dominoes {
		if !d.Placed || d.C1 == nil || d.C2 == nil {
			continue
		}

		placements = append(placements, Placement{
			DominoID: d.ID,
			C1:       *d.C1,
			C2:       *d.C2,
			V1:       d.V1,
			V2:       d.V2,
		})
	}

	return placements
}
