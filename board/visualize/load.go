package visualize

import (
	"fmt"
	"sort"

	"pips-solver/backend/board/types"
)

func NewGameState(date, difficulty string, p Puzzle) (*GameState, error) {
	geometry, regionsByCell, err := BuildGeometry(p)
	if err != nil {
		return nil, err
	}

	cells := make(map[Cell]CellState)
	for cell := range geometry.ExistingCells {
		regionID := regionsByCell[cell]
		cells[cell] = CellState{
			Exists:   true,
			RegionID: regionID,
		}
	}

	dominoes := make([]DominoState, 0, len(p.Dominoes))
	for i, d := range p.Dominoes {
		dominoes = append(dominoes, DominoState{
			ID: i,
			A:  d[0],
			B:  d[1],
			V1: d[0],
			V2: d[1],
		})
	}

	cursor, err := firstCell(geometry.ExistingCells)
	if err != nil {
		return nil, err
	}

	return &GameState{
		Date:             date,
		Difficulty:       difficulty,
		Puzzle:           p,
		Geometry:         geometry,
		RegionsByCell:    regionsByCell,
		Cells:            cells,
		Dominoes:         dominoes,
		Cursor:           cursor,
		SelectedDominoID: 0,
		Orientation:      HorizontalRight,
		Message:          "Puzzle loaded",
	}, nil
}

func BuildGeometry(p Puzzle) (BoardGeometry, map[Cell]int, error) {
	existing := make(map[Cell]bool)
	regionsByCell := make(map[Cell]int)

	minRow, minCol := int(^uint(0)>>1), int(^uint(0)>>1)
	maxRow, maxCol := -1, -1

	for regionID, region := range p.Regions {
		for _, coord := range region.Indices {
			cell := Cell{
				R: coord[0],
				C: coord[1],
			}

			if existing[cell] {
				return BoardGeometry{}, nil, fmt.Errorf("cell %+v appears in multiple regions", cell)
			}

			existing[cell] = true
			regionsByCell[cell] = regionID

		}
	}

	if len(existing) == 0 {
		return BoardGeometry{}, nil, fmt.Errorf()
	}

	if len(existing) != 2*len(p.Dominoes) {
		return BoardGeometry{}, nil, fmt.Errorf("cell/domino mismatch: got %d cells and %d dominoes", len(existing), len(p.Dominoes))
	}

	return BoardGeometry{
		MinRow:        minRow,
		MaxRow:        maxRow,
		MinCol:        minCol,
		MaxCol:        maxCol,
		ExistingCells: existing,
	}, regionsByCell, nil
}

func firstCell(cells map[Cell]bool) (Cell, error) {
	if len(cells) == 0 {
		return Cell{}, fmt.Errorf("no cells")
	}

	all := make([]Cell, 0, len(cells))
	for c := range cells {
		all = append(all, c)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].R == all[j].R {
			return all[i].C < all[j].C
		}
		return all[i].R < all[j].R
	})

	return all[0], nil
}
