package types

type Cell struct {
	R int
	C int
}

type Placement struct {
	DominoID int

	C1 Cell
	C2 Cell

	V1 int
	V2 int
}
