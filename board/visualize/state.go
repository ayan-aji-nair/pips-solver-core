package visualize

type Orientation int

const (
	HorizontalRight Orientation = iota
	VerticalDown
	HorizontalLeft
	VerticalUp
)

type BoardGeometry struct {
	MinRow int
	MaxRow int
	MinCol int
	MaxCol int

	ExistingCells map[Cell]bool
}

type CellState struct {
	Exists   bool
	RegionID int
	Value    *int
	DominoID *int
	IsHead   bool
}

type DominoState struct {
	ID     int
	A      int
	B      int
	Placed bool

	C1 *Cell
	C2 *Cell

	V1 int
	V2 int
}

type GameState struct {
	Date       string
	Difficulty string
	Puzzle     Puzzle

	Geometry      BoardGeometry
	RegionsByCell map[Cell]int
	Cells         map[Cell]CellState
	Dominoes      []DominoState

	Cursor           Cell
	SelectedDominoID int
	Orientation      Orientation
	Message          string
	SolvedBy         string
}
