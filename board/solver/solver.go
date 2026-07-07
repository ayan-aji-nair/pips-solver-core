package solver

import (
	"fmt"
	"github.com/draffensperger/golp"
	"pips-solver/backend/board/solver/helpers"
	"pips-solver/backend/board/types"
	"sort"
)

type Puzzle = types.Puzzle
type Cell = helpers.Cell
type Placement = helpers.Placement

type IlpModel struct {
	puzzle    Puzzle
	cells     []Cell
	edges     [][2]Cell
	placement []Placement
}

func NewILPModel(p Puzzle) (*IlpModel, error) {
	cells := helpers.ExtractCells(p)

	if len(cells) != 2*len(p.Dominoes) {
		return nil, fmt.Errorf("ERROR[Solver] cell count mismatch")
	}

	edges := helpers.LegalEdges(cells)
	placements := helpers.BuildPlacements(p, edges)

	if len(placements) == 0 {
		return nil, fmt.Errorf("ERROR[Solver] no legal placements found")
	}

	return &IlpModel{
		puzzle:    p,
		cells:     cells,
		edges:     edges,
		placement: placements,
	}, nil
}

func (m *IlpModel) Solve() ([]Placement, error) {
	nVars := len(m.placement)

	lp := golp.NewLP(0, nVars)

	for i := range m.placement {
		lp.SetBinary(i, true)
	}

	lp.SetObjFn(make([]float64, nVars))

	if err := m.addDominoUsageConstraints(lp); err != nil {
		return nil, err
	}

	if err := m.addCellCoverageConstraints(lp); err != nil {
		return nil, err
	}

	if err := m.addRegionConstraints(lp); err != nil {
		return nil, err
	}

	status := lp.Solve()
	if status != golp.OPTIMAL {
		return nil, fmt.Errorf("ERROR[Solver] no feasible solution")
	}

	var chosen []Placement

	for i, value := range lp.Variables() {
		if value > 0.5 {
			chosen = append(chosen, m.placement[i])
		}
	}

	sort.Slice(chosen, func(i, j int) bool {
		return chosen[i].DominoID < chosen[j].DominoID
	})

	return chosen, nil
}

func (m *IlpModel) addDominoUsageConstraints(lp *golp.LP) error {
	nVars := len(m.placement)

	for dominoID := range m.puzzle.Dominoes {
		row := make([]float64, nVars)
		for i, placement := range m.placement {
			if placement.DominoID == dominoID {
				row[i] = 1
			}
		}

		if err := lp.AddConstraint(row, golp.EQ, 1); err != nil {
			return fmt.Errorf("ERROR[Solver] could not add constraint: %w", err)
		}
	}

	return nil
}

func (m *IlpModel) addCellCoverageConstraints(lp *golp.LP) error {
	nVars := len(m.placement)

	for _, cell := range m.cells {
		row := make([]float64, nVars)

		for i, placement := range m.placement {
			if placement.C1 == cell || placement.C2 == cell {
				row[i] = 1
			}
		}

		if err := lp.AddConstraint(row, golp.EQ, 1); err != nil {
			return fmt.Errorf("ERROR[Solver] could not add constraint: %w", err)
		}
	}

	return nil
}

func (m *IlpModel) addRegionConstraints(lp *golp.LP) error {
	for _, region := range m.puzzle.Regions {
		switch region.RegionType {
		case "empty":
			continue

		case "sum":
			row := helpers.RegionValueExpression(region, m.placement)
			if err := lp.AddConstraint(row, golp.EQ, float64(region.Target)); err != nil {
				return fmt.Errorf("ERROR[Solver] could not add constraint: %w", err)
			}

		case "less":
			row := helpers.RegionValueExpression(region, m.placement)
			if err := lp.AddConstraint(row, golp.LE, float64(region.Target-1)); err != nil {
				return fmt.Errorf("ERROR[Solver] could not add constraint: %w", err)
			}

		case "greater":
			row := helpers.RegionValueExpression(region, m.placement)
			if err := lp.AddConstraint(row, golp.GE, float64(region.Target+1)); err != nil {
				return fmt.Errorf("ERROR[Solver] could not add constraint: %w", err)
			}

		case "equals":
			if err := helpers.AddEqualsRegionConstrants(lp, region, m.placement); err != nil {
				return fmt.Errorf("ERROR[Solver] could not add constraint: %w", err)
			}

		case "unequal":
			if err := helpers.AddUnequalRegionConstraints(lp, region, m.placement); err != nil {
				return fmt.Errorf("ERROR[Solver] could not add constraint: %w", err)
			}

		default:
			return fmt.Errorf("ERROR[Solver] unknown case")
		}
	}

	return nil
}
