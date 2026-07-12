package main

import (
	"fmt"
	"os"

	"pips-solver/backend/board/visualize"
)

func main() {
	if err := visualize.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pips tui failed: %v\n", err)
		os.Exit(1)
	}
}
