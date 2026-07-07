package types

type PuzzlePayload struct {
	Data Puzzles `json:"data"`
}

type Puzzles struct {
	Easy   Puzzle `json:"easy"`
	Medium Puzzle `json:"medium"`
	Hard   Puzzle `json:"hard"`
}

type Puzzle struct {
	Id       int       `json:"id"`
	Dominoes [][]int   `json:"dominoes"`
	Regions  []Region  `json:"regions"`
	Solution [][][]int `json:"solution"`
}

type Region struct {
	RegionType string  `json:"type"`
	Target     int     `json:"target"`
	Indices    [][]int `json:"indices"`
}
