package helpers

import (
	"fmt"
	"github.com/draffensperger/golp"
	"pips-solver/backend/board/types"
)

type Region = types.Region

func RegionValueExpression(region Region, placements []Placement) []float64 {
	row := make([]float64, len(placements))
	regionCells := map[Cell]bool{}

	for _, coord := range region.Indices {
		regionCells[Cell{R: coord[0], C: coord[1]}] = true
	}

	for i, placement := range placements {
		if regionCells[placement.C1] {
			row[i] += float64(placement.V1)
		}

		if regionCells[placement.C2] {
			row[i] += float64(placement.V2)
		}
	}

	return row
}

func AddEqualsRegionConstrants(lp *golp.LP, region Region, placements []Placement) error {
	if len(region.Indices) <= 1 {
		return nil
	}

	base := Cell{
		R: region.Indices[0][0],
		C: region.Indices[0][1],
	}

	for i := 1; i < len(region.Indices); i++ {
		other := Cell{
			R: region.Indices[i][0],
			C: region.Indices[i][1],
		}

		row := make([]float64, len(placements))

		for j, placement := range placements {
			if placement.C1 == base {
				row[j] += float64(placement.V1)
			}

			if placement.C2 == base {
				row[j] += float64(placement.V2)
			}

			if placement.C1 == other {
				row[j] -= float64(placement.V1)
			}

			if placement.C2 == other {
				row[j] -= float64(placement.V2)
			}
		}

		if err := lp.AddConstraint(row, golp.EQ, 0); err != nil {
			return fmt.Errorf("ERROR[LP_HELPERS] error adding equals region constraint: %w", err)
		}
	}

	return nil
}

func AddUnequalRegionConstraints(lp *golp.LP, region Region, placements []Placement) error {
	regionCells := map[Cell]bool{}

	for _, coord := range region.Indices {
		regionCells[Cell{R: coord[0], C: coord[1]}] = true
	}

	for pip := 0; pip <= 6; pip++ {
		row := make([]float64, len(placements))

		for i, placement := range placements {
			if regionCells[placement.C1] && placement.V1 == pip {
				row[i] += 1
			}

			if regionCells[placement.C2] && placement.V2 == pip {
				row[i] += 1
			}
		}

		if err := lp.AddConstraint(row, golp.LE, 1); err != nil {
			return fmt.Errorf("ERROR[LP_HELPERS] error adding not equal region constraints: %w", err)
		}
	}

	return nil
}
