package visualize

import (
	"fmt"
	"strconv"

	"pips-solver/backend/board/types"
)

func moveCursor(g *GameState, dr, dc int) {
	next := types.Cell{
		R: g.Cursor.R + dr,
		C: g.Cursor.C + dc,
	}

	if g.Geometry.ExistingCells[next] {
		g.Cursor = next
		g.Message = fmt.Sprintf("Cursor: r%dc%d", next.R, next.C)
		return
	}

	g.Message = "Cannot move outside puzzle board."
}

func selectNextUnplacedDomino(g *GameState) {
	if len(g.Dominoes) == 0 {
		g.Message = "No dominoes."
		return
	}

	start := g.SelectedDominoID

	for i := 1; i <= len(g.Dominoes); i++ {
		idx := (start + i) % len(g.Dominoes)

		if !g.Dominoes[idx].Placed {
			g.SelectedDominoID = idx
			g.Message = fmt.Sprintf("Selected domino %d.", idx+1)
			return
		}
	}

	g.Message = "No unplaced dominoes left."
}

func selectPrevUnplacedDomino(g *GameState) {
	if len(g.Dominoes) == 0 {
		g.Message = "No dominoes."
		return
	}

	start := g.SelectedDominoID

	for i := 1; i <= len(g.Dominoes); i++ {
		idx := (start - i + len(g.Dominoes)) % len(g.Dominoes)

		if !g.Dominoes[idx].Placed {
			g.SelectedDominoID = idx
			g.Message = fmt.Sprintf("Selected domino %d.", idx+1)
			return
		}
	}

	g.Message = "No unplaced dominoes left."
}

func selectVisibleDomino(g *GameState, visibleNumber int) {
	if visibleNumber <= 0 {
		return
	}

	unplacedSeen := 0

	for i := range g.Dominoes {
		if g.Dominoes[i].Placed {
			continue
		}

		unplacedSeen++

		if unplacedSeen == visibleNumber {
			g.SelectedDominoID = i
			g.Message = fmt.Sprintf("Selected domino %d.", i+1)
			return
		}
	}

	g.Message = fmt.Sprintf("No visible domino %d.", visibleNumber)
}

func numberKey(s string) (int, bool) {
	if len(s) != 1 {
		return 0, false
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	if n < 1 || n > 9 {
		return 0, false
	}

	return n, true
}

func applyPlacementsToGame(g *GameState, placements []types.Placement) error {
	clearGamePlacements(g)

	for _, p := range placements {
		if err := applyPlacementToGame(g, p); err != nil {
			return err
		}
	}

	g.SolvedBy = "solver"
	g.Message = "Solver placements applied."

	return nil
}

func clearGamePlacements(g *GameState) {
	for cell, cs := range g.Cells {
		cs.Value = nil
		cs.DominoID = nil
		cs.IsHead = false
		g.Cells[cell] = cs
	}

	for i := range g.Dominoes {
		g.Dominoes[i].Placed = false
		g.Dominoes[i].C1 = nil
		g.Dominoes[i].C2 = nil
		g.Dominoes[i].V1 = g.Dominoes[i].A
		g.Dominoes[i].V2 = g.Dominoes[i].B
	}
}

func applyPlacementToGame(g *GameState, p types.Placement) error {
	if p.DominoID < 0 || p.DominoID >= len(g.Dominoes) {
		return fmt.Errorf("invalid domino id %d", p.DominoID)
	}

	if !g.Geometry.ExistingCells[p.C1] {
		return fmt.Errorf("placement uses missing cell r%dc%d", p.C1.R, p.C1.C)
	}

	if !g.Geometry.ExistingCells[p.C2] {
		return fmt.Errorf("placement uses missing cell r%dc%d", p.C2.R, p.C2.C)
	}

	if g.Cells[p.C1].DominoID != nil {
		return fmt.Errorf("cell r%dc%d is occupied more than once", p.C1.R, p.C1.C)
	}

	if g.Cells[p.C2].DominoID != nil {
		return fmt.Errorf("cell r%dc%d is occupied more than once", p.C2.R, p.C2.C)
	}

	id := p.DominoID
	c1 := p.C1
	c2 := p.C2

	d := &g.Dominoes[id]
	d.Placed = true
	d.C1 = &c1
	d.C2 = &c2
	d.V1 = p.V1
	d.V2 = p.V2

	cell1 := g.Cells[p.C1]
	cell1.Value = &d.V1
	cell1.DominoID = &id
	cell1.IsHead = true
	g.Cells[p.C1] = cell1

	cell2 := g.Cells[p.C2]
	cell2.Value = &d.V2
	cell2.DominoID = &id
	cell2.IsHead = false
	g.Cells[p.C2] = cell2

	return nil
}

func orientationName(o Orientation) string {
	switch o {
	case HorizontalRight:
		return "right"
	case VerticalDown:
		return "down"
	case HorizontalLeft:
		return "left"
	case VerticalUp:
		return "up"
	default:
		return "unknown"
	}
}
