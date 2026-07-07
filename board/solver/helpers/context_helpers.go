package helpers

import (
	"pips-solver/backend/board/types"
	"sort"
)

type Cell = types.Cell
type Placement = types.Placement
type Puzzle = types.Puzzle

func ExtractCells(p Puzzle) []Cell {
	seen := map[Cell]bool{}

	for _, region := range p.Regions {
		for _, coord := range region.Indices {
			seen[Cell{R: coord[0], C: coord[1]}] = true
		}
	}

	cells := make([]Cell, 0, len(seen))
	for cell := range seen {
		cells = append(cells, cell)
	}

	sort.Slice(cells, func(i, j int) bool {
		if cells[i].R == cells[j].R {
			return cells[i].C < cells[j].C
		}

		return cells[i].R < cells[j].R
	})

	return cells
}

func LegalEdges(cells []Cell) [][2]Cell {
	var edges [][2]Cell

	for i := 0; i < len(cells); i++ {
		for j := i + 1; j < len(cells); j++ {
			if adjacent(cells[i], cells[j]) {
				edges = append(edges, [2]Cell{cells[i], cells[j]})
			}
		}
	}

	return edges
}

func BuildPlacements(p types.Puzzle, edges [][2]Cell) []Placement {
	var placements []Placement

	for dominoId, domino := range p.Dominoes {
		a := domino[0]
		b := domino[1]

		for _, edge := range edges {
			c1 := edge[0]
			c2 := edge[1]

			placements = append(placements, Placement{
				DominoID: dominoId,
				C1:       c1,
				C2:       c2,
				V1:       a,
				V2:       b,
			})

			if a != b {
				placements = append(placements, Placement{
					DominoID: dominoId,
					C1:       c1,
					C2:       c2,
					V1:       b,
					V2:       a,
				})
			}
		}
	}

	return placements
}

func adjacent(a, b Cell) bool {
	dr := abs(a.R - b.R)
	dc := abs(a.C - b.C)

	return (dr + dc) == 1
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
