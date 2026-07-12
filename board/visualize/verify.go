package visualize

import (
	"fmt"
	"sort"

	"pips-solver/backend/board/types"
)

type EdgeKey struct {
	A types.Cell
	B types.Cell
}

type VerificationResult struct {
	MatchesAPISolution bool
	SatisfiesRegions   bool
	UsesAllDominoes    bool
	CoversAllCells     bool

	MissingEdges []EdgeKey
	ExtraEdges   []EdgeKey

	RegionErrors []string
}

func NormalizeEdge(a, b types.Cell) EdgeKey {
	if lessCell(b, a) {
		a, b = b, a
	}
	return EdgeKey{A: a, B: b}
}

func lessCell(a, b types.Cell) bool {
	if a.R == b.R {
		return a.C < b.C
	}
	return a.R < b.R
}

func NormalizePlacements(placements []types.Placement) map[EdgeKey]bool {
	out := make(map[EdgeKey]bool)

	for _, p := range placements {
		out[NormalizeEdge(p.C1, p.C2)] = true
	}

	return out
}

func NormalizeSolution(solution [][][]int) (map[EdgeKey]bool, error) {
	out := make(map[EdgeKey]bool)

	for i, pair := range solution {
		if len(pair) != 2 || len(pair[0]) != 2 || len(pair[1]) != 2 {
			return nil, fmt.Errorf("invalid solution edge at index %d", i)
		}

		a := types.Cell{R: pair[0][0], C: pair[0][1]}
		b := types.Cell{R: pair[1][0], C: pair[1][1]}

		out[NormalizeEdge(a, b)] = true
	}

	return out, nil
}

func CompareEdges(solution, current map[EdgeKey]bool) (missing, extra []EdgeKey) {
	for edge := range solution {
		if !current[edge] {
			missing = append(missing, edge)
		}
	}

	for edge := range current {
		if !solution[edge] {
			extra = append(extra, edge)
		}
	}

	sortEdges(missing)
	sortEdges(extra)

	return missing, extra
}

func sortEdges(edges []EdgeKey) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].A == edges[j].A {
			return lessCell(edges[i].B, edges[j].B)
		}
		return lessCell(edges[i].A, edges[j].A)
	})
}

func VerifyAgainstAPISolution(p types.Puzzle, placements []types.Placement) VerificationResult {
	solutionEdges, err := NormalizeSolution(p.Solution)
	if err != nil {
		return VerificationResult{
			MatchesAPISolution: false,
			RegionErrors:       []string{err.Error()},
		}
	}

	currentEdges := NormalizePlacements(placements)
	missing, extra := CompareEdges(solutionEdges, currentEdges)

	return VerificationResult{
		MatchesAPISolution: len(missing) == 0 && len(extra) == 0,
		MissingEdges:       missing,
		ExtraEdges:         extra,
	}
}
