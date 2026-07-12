package visualize

import (
	"testing"

	"pips-solver/backend/board/types"
)

func vc(row, col int) types.Cell {
	return types.Cell{R: row, C: col}
}

func vp(dominoID int, c1, c2 types.Cell) types.Placement {
	return types.Placement{
		DominoID: dominoID,
		C1:       c1,
		C2:       c2,
		V1:       0,
		V2:       0,
	}
}

func edgeExists(edges []EdgeKey, want EdgeKey) bool {
	for _, edge := range edges {
		if edge == want {
			return true
		}
	}

	return false
}

func TestNormalizeEdgeKeepsAlreadySortedEdge(t *testing.T) {
	got := NormalizeEdge(vc(0, 0), vc(0, 1))

	want := EdgeKey{
		A: vc(0, 0),
		B: vc(0, 1),
	}

	if got != want {
		t.Fatalf("NormalizeEdge returned %+v, want %+v", got, want)
	}
}

func TestNormalizeEdgeSortsByRowFirst(t *testing.T) {
	got := NormalizeEdge(vc(3, 0), vc(1, 9))

	want := EdgeKey{
		A: vc(1, 9),
		B: vc(3, 0),
	}

	if got != want {
		t.Fatalf("NormalizeEdge returned %+v, want %+v", got, want)
	}
}

func TestNormalizeEdgeSortsByColWhenRowsMatch(t *testing.T) {
	got := NormalizeEdge(vc(2, 5), vc(2, 1))

	want := EdgeKey{
		A: vc(2, 1),
		B: vc(2, 5),
	}

	if got != want {
		t.Fatalf("NormalizeEdge returned %+v, want %+v", got, want)
	}
}

func TestNormalizePlacementsNormalizesPlacementDirection(t *testing.T) {
	placements := []types.Placement{
		vp(0, vc(0, 1), vc(0, 0)),
		vp(1, vc(2, 0), vc(1, 0)),
	}

	got := NormalizePlacements(placements)

	wantA := EdgeKey{
		A: vc(0, 0),
		B: vc(0, 1),
	}

	wantB := EdgeKey{
		A: vc(1, 0),
		B: vc(2, 0),
	}

	if len(got) != 2 {
		t.Fatalf("NormalizePlacements returned %d edges, want 2", len(got))
	}

	if !got[wantA] {
		t.Fatalf("NormalizePlacements missing edge %+v", wantA)
	}

	if !got[wantB] {
		t.Fatalf("NormalizePlacements missing edge %+v", wantB)
	}
}

func TestNormalizeSolutionValidSolution(t *testing.T) {
	solution := [][][]int{
		{{0, 1}, {0, 0}},
		{{2, 0}, {1, 0}},
	}

	got, err := NormalizeSolution(solution)
	if err != nil {
		t.Fatalf("NormalizeSolution returned error: %v", err)
	}

	wantA := EdgeKey{
		A: vc(0, 0),
		B: vc(0, 1),
	}

	wantB := EdgeKey{
		A: vc(1, 0),
		B: vc(2, 0),
	}

	if len(got) != 2 {
		t.Fatalf("NormalizeSolution returned %d edges, want 2", len(got))
	}

	if !got[wantA] {
		t.Fatalf("NormalizeSolution missing edge %+v", wantA)
	}

	if !got[wantB] {
		t.Fatalf("NormalizeSolution missing edge %+v", wantB)
	}
}

func TestNormalizeSolutionRejectsEdgeWithOneCell(t *testing.T) {
	solution := [][][]int{
		{{0, 0}},
	}

	_, err := NormalizeSolution(solution)
	if err == nil {
		t.Fatalf("NormalizeSolution succeeded, want error")
	}
}

func TestNormalizeSolutionRejectsEdgeWithThreeCells(t *testing.T) {
	solution := [][][]int{
		{{0, 0}, {0, 1}, {0, 2}},
	}

	_, err := NormalizeSolution(solution)
	if err == nil {
		t.Fatalf("NormalizeSolution succeeded, want error")
	}
}

func TestNormalizeSolutionRejectsMalformedCoordinate(t *testing.T) {
	solution := [][][]int{
		{{0, 0}, {1}},
	}

	_, err := NormalizeSolution(solution)
	if err == nil {
		t.Fatalf("NormalizeSolution succeeded, want error")
	}
}

func TestCompareEdgesExactMatch(t *testing.T) {
	edgeA := EdgeKey{A: vc(0, 0), B: vc(0, 1)}
	edgeB := EdgeKey{A: vc(1, 0), B: vc(1, 1)}

	solution := map[EdgeKey]bool{
		edgeA: true,
		edgeB: true,
	}

	current := map[EdgeKey]bool{
		edgeB: true,
		edgeA: true,
	}

	missing, extra := CompareEdges(solution, current)

	if len(missing) != 0 {
		t.Fatalf("missing = %+v, want empty", missing)
	}

	if len(extra) != 0 {
		t.Fatalf("extra = %+v, want empty", extra)
	}
}

func TestCompareEdgesFindsMissingAndExtraEdges(t *testing.T) {
	shared := EdgeKey{A: vc(0, 0), B: vc(0, 1)}
	missingEdge := EdgeKey{A: vc(1, 0), B: vc(1, 1)}
	extraEdge := EdgeKey{A: vc(2, 0), B: vc(2, 1)}

	solution := map[EdgeKey]bool{
		shared:      true,
		missingEdge: true,
	}

	current := map[EdgeKey]bool{
		shared:    true,
		extraEdge: true,
	}

	missing, extra := CompareEdges(solution, current)

	if len(missing) != 1 {
		t.Fatalf("len(missing) = %d, want 1: %+v", len(missing), missing)
	}

	if missing[0] != missingEdge {
		t.Fatalf("missing[0] = %+v, want %+v", missing[0], missingEdge)
	}

	if len(extra) != 1 {
		t.Fatalf("len(extra) = %d, want 1: %+v", len(extra), extra)
	}

	if extra[0] != extraEdge {
		t.Fatalf("extra[0] = %+v, want %+v", extra[0], extraEdge)
	}
}

func TestVerifyAgainstAPISolutionMatchesExactSolution(t *testing.T) {
	puzzle := types.Puzzle{
		Solution: [][][]int{
			{{0, 0}, {0, 1}},
			{{1, 0}, {1, 1}},
		},
	}

	placements := []types.Placement{
		vp(0, vc(0, 0), vc(0, 1)),
		vp(1, vc(1, 0), vc(1, 1)),
	}

	got := VerifyAgainstAPISolution(puzzle, placements)

	if !got.MatchesAPISolution {
		t.Fatalf("MatchesAPISolution = false, want true; result = %+v", got)
	}

	if len(got.MissingEdges) != 0 {
		t.Fatalf("MissingEdges = %+v, want empty", got.MissingEdges)
	}

	if len(got.ExtraEdges) != 0 {
		t.Fatalf("ExtraEdges = %+v, want empty", got.ExtraEdges)
	}
}

func TestVerifyAgainstAPISolutionIgnoresPlacementDirection(t *testing.T) {
	puzzle := types.Puzzle{
		Solution: [][][]int{
			{{0, 0}, {0, 1}},
			{{1, 0}, {1, 1}},
		},
	}

	placements := []types.Placement{
		vp(0, vc(0, 1), vc(0, 0)),
		vp(1, vc(1, 1), vc(1, 0)),
	}

	got := VerifyAgainstAPISolution(puzzle, placements)

	if !got.MatchesAPISolution {
		t.Fatalf("MatchesAPISolution = false, want true; result = %+v", got)
	}
}

func TestVerifyAgainstAPISolutionDetectsMissingEdge(t *testing.T) {
	expectedMissing := EdgeKey{
		A: vc(1, 0),
		B: vc(1, 1),
	}

	puzzle := types.Puzzle{
		Solution: [][][]int{
			{{0, 0}, {0, 1}},
			{{1, 0}, {1, 1}},
		},
	}

	placements := []types.Placement{
		vp(0, vc(0, 0), vc(0, 1)),
	}

	got := VerifyAgainstAPISolution(puzzle, placements)

	if got.MatchesAPISolution {
		t.Fatalf("MatchesAPISolution = true, want false")
	}

	if len(got.MissingEdges) != 1 {
		t.Fatalf("len(MissingEdges) = %d, want 1: %+v", len(got.MissingEdges), got.MissingEdges)
	}

	if got.MissingEdges[0] != expectedMissing {
		t.Fatalf("MissingEdges[0] = %+v, want %+v", got.MissingEdges[0], expectedMissing)
	}

	if len(got.ExtraEdges) != 0 {
		t.Fatalf("ExtraEdges = %+v, want empty", got.ExtraEdges)
	}
}

func TestVerifyAgainstAPISolutionDetectsExtraEdge(t *testing.T) {
	expectedExtra := EdgeKey{
		A: vc(2, 0),
		B: vc(2, 1),
	}

	puzzle := types.Puzzle{
		Solution: [][][]int{
			{{0, 0}, {0, 1}},
		},
	}

	placements := []types.Placement{
		vp(0, vc(0, 0), vc(0, 1)),
		vp(1, vc(2, 0), vc(2, 1)),
	}

	got := VerifyAgainstAPISolution(puzzle, placements)

	if got.MatchesAPISolution {
		t.Fatalf("MatchesAPISolution = true, want false")
	}

	if len(got.MissingEdges) != 0 {
		t.Fatalf("MissingEdges = %+v, want empty", got.MissingEdges)
	}

	if len(got.ExtraEdges) != 1 {
		t.Fatalf("len(ExtraEdges) = %d, want 1: %+v", len(got.ExtraEdges), got.ExtraEdges)
	}

	if got.ExtraEdges[0] != expectedExtra {
		t.Fatalf("ExtraEdges[0] = %+v, want %+v", got.ExtraEdges[0], expectedExtra)
	}
}

func TestVerifyAgainstAPISolutionDetectsMissingAndExtraEdge(t *testing.T) {
	expectedMissing := EdgeKey{
		A: vc(1, 0),
		B: vc(1, 1),
	}

	expectedExtra := EdgeKey{
		A: vc(2, 0),
		B: vc(2, 1),
	}

	puzzle := types.Puzzle{
		Solution: [][][]int{
			{{0, 0}, {0, 1}},
			{{1, 0}, {1, 1}},
		},
	}

	placements := []types.Placement{
		vp(0, vc(0, 0), vc(0, 1)),
		vp(1, vc(2, 0), vc(2, 1)),
	}

	got := VerifyAgainstAPISolution(puzzle, placements)

	if got.MatchesAPISolution {
		t.Fatalf("MatchesAPISolution = true, want false")
	}

	if len(got.MissingEdges) != 1 {
		t.Fatalf("len(MissingEdges) = %d, want 1: %+v", len(got.MissingEdges), got.MissingEdges)
	}

	if got.MissingEdges[0] != expectedMissing {
		t.Fatalf("MissingEdges[0] = %+v, want %+v", got.MissingEdges[0], expectedMissing)
	}

	if len(got.ExtraEdges) != 1 {
		t.Fatalf("len(ExtraEdges) = %d, want 1: %+v", len(got.ExtraEdges), got.ExtraEdges)
	}

	if got.ExtraEdges[0] != expectedExtra {
		t.Fatalf("ExtraEdges[0] = %+v, want %+v", got.ExtraEdges[0], expectedExtra)
	}
}

func TestVerifyAgainstAPISolutionReturnsErrorForMalformedAPISolution(t *testing.T) {
	puzzle := types.Puzzle{
		Solution: [][][]int{
			{{0, 0}},
		},
	}

	placements := []types.Placement{
		vp(0, vc(0, 0), vc(0, 1)),
	}

	got := VerifyAgainstAPISolution(puzzle, placements)

	if got.MatchesAPISolution {
		t.Fatalf("MatchesAPISolution = true, want false")
	}

	if len(got.RegionErrors) != 1 {
		t.Fatalf("len(RegionErrors) = %d, want 1: %+v", len(got.RegionErrors), got.RegionErrors)
	}

	if len(got.MissingEdges) != 0 {
		t.Fatalf("MissingEdges = %+v, want empty when API solution is malformed", got.MissingEdges)
	}

	if len(got.ExtraEdges) != 0 {
		t.Fatalf("ExtraEdges = %+v, want empty when API solution is malformed", got.ExtraEdges)
	}
}
