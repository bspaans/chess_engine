package main

import (
	"bufio"
	"os"

	"github.com/bspaans/chess_engine"
)

func main() {
	uci := chess_engine.NewUCI("bs-engine", "Bart Spaans", &chess_engine.BSEngine{})
	reader := bufio.NewReader(os.Stdin)
	uci.Start(reader)
}
