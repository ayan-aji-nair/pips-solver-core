package solver

import (
	"testing"

	"pips-solver/backend/board/types"
)

func TestSolveSimpleSumPuzzle(t *testing.T) {
	p := Puzzle{
		Dominoes: [][]int{
			{1, 2},
			{3, 4},
		},
		Regions: []types.Region{
			region("sum", 3, Cell{R: 0, C: 0}, Cell{R: 0, C: 1}),
			region("sum", 7, Cell{R: 1, C: 0}, Cell{R: 1, C: 1}),
		},
	}

	chosen := solvePuzzleForTest(t, p)
	assertSolutionSatisfiesPuzzle(t, p, chosen)
}

func TestSolveEqualsAndUnequalPuzzle(t *testing.T) {
	p := Puzzle{
		Dominoes: [][]int{
			{5, 5},
			{1, 2},
		},
		Regions: []types.Region{
			region("equals", 0, Cell{R: 0, C: 0}, Cell{R: 0, C: 1}),
			region("unequal", 0, Cell{R: 1, C: 0}, Cell{R: 1, C: 1}),
		},
	}

	chosen := solvePuzzleForTest(t, p)
	assertSolutionSatisfiesPuzzle(t, p, chosen)
}

func TestSolveLessAndGreaterPuzzle(t *testing.T) {
	p := Puzzle{
		Dominoes: [][]int{
			{1, 2},
			{5, 6},
		},
		Regions: []types.Region{
			region("less", 4, Cell{R: 0, C: 0}, Cell{R: 0, C: 1}),
			region("greater", 10, Cell{R: 1, C: 0}, Cell{R: 1, C: 1}),
		},
	}

	chosen := solvePuzzleForTest(t, p)
	assertSolutionSatisfiesPuzzle(t, p, chosen)
}

func TestSolveReturnsErrorForInfeasiblePuzzle(t *testing.T) {
	p := Puzzle{
		Dominoes: [][]int{
			{1, 2},
			{3, 4},
		},
		Regions: []types.Region{
			region("sum", 100, Cell{R: 0, C: 0}, Cell{R: 0, C: 1}),
			region("empty", 0, Cell{R: 1, C: 0}, Cell{R: 1, C: 1}),
		},
	}

	model, err := NewILPModel(p)
	if err != nil {
		t.Fatalf("NewILPModel failed: %v", err)
	}

	_, err = model.Solve()
	if err == nil {
		t.Fatalf("expected infeasible puzzle to return error")
	}
}

func solvePuzzleForTest(t *testing.T, p Puzzle) []Placement {
	t.Helper()

	model, err := NewILPModel(p)
	if err != nil {
		t.Fatalf("NewILPModel failed: %v", err)
	}

	chosen, err := model.Solve()
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	return chosen
}

func assertSolutionSatisfiesPuzzle(t *testing.T, p Puzzle, chosen []Placement) {
	t.Helper()

	if len(chosen) != len(p.Dominoes) {
		t.Fatalf("expected %d chosen placements, got %d", len(p.Dominoes), len(chosen))
	}

	usedDominoes := map[int]int{}
	coveredCells := map[Cell]int{}
	cellValues := map[Cell]int{}

	for _, placement := range chosen {
		if placement.DominoID < 0 || placement.DominoID >= len(p.Dominoes) {
			t.Fatalf("invalid domino index %d", placement.DominoID)
		}

		domino := p.Dominoes[placement.DominoID]

		if !sameDominoValues(domino, placement.V1, placement.V2) {
			t.Fatalf(
				"placement [%d,%d] does not match domino %d %v",
				placement.V1,
				placement.V2,
				placement.DominoID,
				domino,
			)
		}

		if !adjacentForTest(placement.C1, placement.C2) {
			t.Fatalf("placement cells are not adjacent: %+v %+v", placement.C1, placement.C2)
		}

		usedDominoes[placement.DominoID]++

		coveredCells[placement.C1]++
		coveredCells[placement.C2]++

		cellValues[placement.C1] = placement.V1
		cellValues[placement.C2] = placement.V2
	}

	for dominoID := range p.Dominoes {
		if usedDominoes[dominoID] != 1 {
			t.Fatalf("domino %d used %d times, expected once", dominoID, usedDominoes[dominoID])
		}
	}

	for _, region := range p.Regions {
		for _, coord := range region.Indices {
			cell := Cell{R: coord[0], C: coord[1]}

			if coveredCells[cell] != 1 {
				t.Fatalf("cell %+v covered %d times, expected once", cell, coveredCells[cell])
			}

			if _, ok := cellValues[cell]; !ok {
				t.Fatalf("cell %+v has no assigned value", cell)
			}
		}
	}

	assertRegionsSatisfied(t, p, cellValues)
}

func assertRegionsSatisfied(t *testing.T, p Puzzle, values map[Cell]int) {
	t.Helper()

	for _, region := range p.Regions {
		switch region.RegionType {
		case "empty":
			continue

		case "sum":
			got := regionSum(region, values)
			if got != region.Target {
				t.Fatalf("sum region got %d, expected %d", got, region.Target)
			}

		case "less":
			got := regionSum(region, values)
			if got >= region.Target {
				t.Fatalf("less region got %d, expected < %d", got, region.Target)
			}

		case "greater":
			got := regionSum(region, values)
			if got <= region.Target {
				t.Fatalf("greater region got %d, expected > %d", got, region.Target)
			}

		case "equals":
			if len(region.Indices) <= 1 {
				continue
			}

			base := valueAt(region.Indices[0], values)

			for _, coord := range region.Indices[1:] {
				got := valueAt(coord, values)
				if got != base {
					t.Fatalf("equals region has values %d and %d", base, got)
				}
			}

		case "unequal":
			seen := map[int]bool{}

			for _, coord := range region.Indices {
				got := valueAt(coord, values)

				if seen[got] {
					t.Fatalf("unequal region repeated value %d", got)
				}

				seen[got] = true
			}

		default:
			t.Fatalf("unknown region type %q", region.RegionType)
		}
	}
}

func region(regionType string, target int, cells ...Cell) types.Region {
	indices := make([][]int, 0, len(cells))

	for _, cell := range cells {
		indices = append(indices, []int{cell.R, cell.C})
	}

	return types.Region{
		RegionType: regionType,
		Target:     target,
		Indices:    indices,
	}
}

func regionSum(region types.Region, values map[Cell]int) int {
	sum := 0

	for _, coord := range region.Indices {
		sum += valueAt(coord, values)
	}

	return sum
}

func valueAt(coord []int, values map[Cell]int) int {
	return values[Cell{R: coord[0], C: coord[1]}]
}

func sameDominoValues(domino []int, a int, b int) bool {
	return domino[0] == a && domino[1] == b ||
		domino[0] == b && domino[1] == a
}

func adjacentForTest(a Cell, b Cell) bool {
	return absForTest(a.R-b.R)+absForTest(a.C-b.C) == 1
}

func absForTest(x int) int {
	if x < 0 {
		return -x
	}

	return x
}
