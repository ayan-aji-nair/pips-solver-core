package main

import (
	"log"

	"pips-solver/backend/board/visualize"
)

func main() {
	if err := visualize.RunManualDemo(); err != nil {
		log.Fatal(err)
	}
}
